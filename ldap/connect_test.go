package ldap

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"math/big"
	"net"
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
