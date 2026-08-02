package ldap

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// A Config that leaves TLS unset must not panic on the plain ldaps:// path.
// Before the guard, l.TLS.Clone() returned nil and the ServerName assignment
// dereferenced it - reached for every non-insecure connection, not just for
// StartTLS as it was upstream.
func TestConnectNilTLSConfigDoesNotPanic(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close() // nothing is listening, so the dial fails fast

	l := &Config{Enabled: true, ServerInsecure: false, ServerStartTLS: false}
	conn, err := l.connect(addr)
	if conn != nil {
		conn.Close()
	}
	if err == nil {
		t.Fatal("expected a dial error against a closed port")
	}
	if strings.Contains(err.Error(), "nil pointer") {
		t.Fatalf("unexpected: %v", err)
	}
	t.Logf("returned an error instead of panicking: %v", err)
}

// The two switches MinIO exposes are independent and Validate rejects no
// combination of them, so all four have to behave. What is asserted is what the
// server sees on the wire before connect() returns: a TLS ClientHello for
// ldaps://, a StartTLS extended request when the connection is to be upgraded,
// and nothing at all only when the caller asked for plain LDAP and no upgrade.
//
// The case that regressed was insecure+starttls. Taking the insecure branch
// first skipped the upgrade, so the bind that follows went out in the clear.
func TestConnectProtocolSelection(t *testing.T) {
	const (
		wantNothing = iota
		wantStartTLS
		wantClientHello
	)
	for _, tc := range []struct {
		name               string
		insecure, starttls bool
		want               int
	}{
		{"ldaps", false, false, wantClientHello},
		{"starttls", false, true, wantStartTLS},
		{"insecure", true, false, wantNothing},
		{"insecure+starttls", true, true, wantStartTLS},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ln, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer ln.Close()

			first := make(chan []byte, 1)
			go func() {
				c, err := ln.Accept()
				if err != nil {
					first <- nil
					return
				}
				defer c.Close()
				// The upgrade cases end as soon as this read returns and the
				// deferred close unblocks the client, so their deadline only
				// bounds a hang. wantNothing has nothing to wait for: there the
				// deadline is the assertion itself, and a short one keeps the
				// case from costing whole seconds. Anything the client did send
				// would have landed on loopback long before it expires.
				deadline := 2 * time.Second
				if tc.want == wantNothing {
					deadline = 500 * time.Millisecond
				}
				c.SetReadDeadline(time.Now().Add(deadline))
				buf := make([]byte, 4096)
				n, _ := c.Read(buf)
				first <- buf[:n]
			}()

			l := &Config{
				Enabled: true, ServerInsecure: tc.insecure, ServerStartTLS: tc.starttls,
				TLS: &tls.Config{InsecureSkipVerify: true},
			}
			conn, _ := l.connect(ln.Addr().String())
			got := <-first
			if conn != nil {
				conn.Close()
			}

			switch tc.want {
			case wantNothing:
				if len(got) != 0 {
					t.Errorf("expected nothing on the wire, got %d bytes", len(got))
				}
			case wantClientHello:
				// TLS record: handshake (0x16), version 3.x.
				if len(got) < 3 || got[0] != 0x16 || got[1] != 0x03 {
					t.Errorf("expected a TLS ClientHello, got %d bytes %x", len(got), got[:min(8, len(got))])
				}
			case wantStartTLS:
				// LDAP extended request carrying the StartTLS OID.
				if !bytes.Contains(got, []byte("1.3.6.1.4.1.1466.20037")) {
					t.Errorf("expected the StartTLS extended request, got %d bytes %x",
						len(got), got[:min(16, len(got))])
				}
			}
		})
	}
}

// startTLSRefusal is what a server answers with when it will not upgrade: an
// extended response carrying a failure result code and nothing else. Encoded by
// hand so the test needs no BER dependency. Byte 4 is the message ID, which has
// to be copied from the request or go-ldap ignores the response.
//
//	30 0c            SEQUENCE          -- LDAPMessage
//	   02 01 00      INTEGER           -- messageID
//	   78 07         [APPLICATION 24]  -- extendedResp
//	      0a 01 01   ENUMERATED        -- resultCode: operationsError
//	      04 00      OCTET STRING      -- matchedDN
//	      04 00      OCTET STRING      -- diagnosticMessage
var startTLSRefusal = []byte{0x30, 0x0c, 0x02, 0x01, 0x00, 0x78, 0x07, 0x0a, 0x01, 0x01, 0x04, 0x00, 0x04, 0x00}

// A dial failure returns no connection, so a failed StartTLS is the only way
// connect() can hand back one alongside an error. Callers only take ownership
// of a connection that arrives without an error, so whatever is left open here
// is left open for good - once per login attempt.
//
// The server declines at the LDAP layer rather than breaking the TLS handshake,
// and the distinction is what makes the test worth having: go-ldap closes the
// connection itself when the handshake fails, so that case would pass whether
// or not connect() does the right thing.
func TestConnectClosesTheConnectionWhenStartTLSIsRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	serverRead := make(chan error, 1)
	go func() {
		c, err := ln.Accept()
		if err != nil {
			serverRead <- err
			return
		}
		defer c.Close()

		buf := make([]byte, 4096)
		c.SetReadDeadline(time.Now().Add(5 * time.Second))
		n, err := c.Read(buf) // the StartTLS extended request
		if err != nil {
			serverRead <- err
			return
		}
		if n < 5 || buf[2] != 0x02 || buf[3] != 0x01 {
			serverRead <- fmt.Errorf("cannot locate the message ID in %x", buf[:min(8, n)])
			return
		}
		resp := append([]byte(nil), startTLSRefusal...)
		resp[4] = buf[4]
		c.Write(resp)

		// The refusal leaves the socket open, so only the client closing it
		// ends this read. Reaching the deadline means it did not.
		c.SetReadDeadline(time.Now().Add(5 * time.Second))
		for {
			if _, err = c.Read(buf); err != nil {
				serverRead <- err
				return
			}
		}
	}()

	l := &Config{
		Enabled:        true,
		ServerStartTLS: true,
		TLS:            &tls.Config{InsecureSkipVerify: true},
	}
	conn, err := l.connect(ln.Addr().String())
	if conn != nil {
		conn.Close()
		t.Error("connect returned a live connection alongside an error; callers drop it, so it leaks")
	}
	if err == nil {
		t.Fatal("expected the refused upgrade to surface as an error")
	}
	if err := <-serverRead; err == nil || errors.Is(err, os.ErrDeadlineExceeded) {
		t.Fatalf("the client never closed the connection: %v", err)
	}
}

// The ldaps:// path once dialed with no TLS config at all, leaving go-ldap to
// build its own from the URL and ignore Config.TLS - RootCAs and
// InsecureSkipVerify with it. That is how a private CA came to be trusted for
// StartTLS but not for ldaps://. Asserting that a ClientHello reaches the wire
// does not catch it, since go-ldap sends one either way; completing a handshake
// against a CA only Config.TLS knows about does.
func TestConnectLDAPSUsesTheConfiguredTLS(t *testing.T) {
	cert, pool := selfSignedCert(t)

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{Certificates: []tls.Certificate{cert}})
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		io.Copy(io.Discard, c) // drives the server handshake, then waits the client out
	}()

	l := &Config{Enabled: true, TLS: &tls.Config{RootCAs: pool}}
	conn, err := l.connect(ln.Addr().String())
	if err != nil {
		t.Fatalf("Config.TLS did not reach the dialer: %v", err)
	}
	conn.Close()
}

// selfSignedCert returns a certificate valid for 127.0.0.1 and a pool that
// trusts it. Showing that Config.TLS is honored needs a CA the system roots
// cannot supply.
func selfSignedCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "silo-pkg ldap test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1)},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(leaf)
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}, pool
}
