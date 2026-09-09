// Copyright (c) 2026 Pigsty
// SPDX-License-Identifier: AGPL-3.0-or-later

package env

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestWebEnvTLSKeyExchangeDefaults(t *testing.T) {
	for _, debug := range []string{"tlsmlkem=0", "tlsmlkem=1"} {
		t.Run(debug, func(t *testing.T) {
			t.Setenv("GODEBUG", debug)
			hellos := make(chan []tls.CurveID, 1)
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = io.WriteString(w, "lab-value")
			}))
			server.TLS = &tls.Config{GetConfigForClient: func(hello *tls.ClientHelloInfo) (*tls.Config, error) {
				select {
				case hellos <- slices.Clone(hello.SupportedCurves):
				default:
				}
				return nil, nil
			}}
			server.StartTLS()
			defer server.Close()
			roots := x509.NewCertPool()
			roots.AddCert(server.Certificate())
			previousRoots := globalRootCAs
			RegisterGlobalCAs(roots)
			t.Cleanup(func() { RegisterGlobalCAs(previousRoots) })
			endpoint := "env+tls://local:" + strings.Repeat("x", 64) + "@" + server.Listener.Addr().String()
			value, _, _, err := getEnvValueFromHTTP(endpoint, "lab-key")
			if err != nil || value != "lab-value" {
				t.Fatalf("value %q, error %v", value, err)
			}
			curves := <-hellos
			if got, want := slices.Contains(curves, tls.X25519MLKEM768), debug == "tlsmlkem=1"; got != want {
				t.Errorf("ML-KEM offered = %v, want %v; curves %v", got, want, curves)
			}
		})
	}
}
