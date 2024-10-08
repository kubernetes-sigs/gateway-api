/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package kubernetes

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_generateCACert(t *testing.T) {
	tests := []struct {
		name        string
		hosts       []string
		expectedErr string
	}{
		{
			name:  "one host generates cert with no host",
			hosts: []string{},
		},
		{
			name:  "one host generates cert for same host",
			hosts: []string{"abc.example.com"},
		},
		{
			name:  "wildcard generates cert for same host",
			hosts: []string{"*.example.com"},
		},
		{
			name:  "two hosts generates cert for both hosts",
			hosts: []string{"abc.example.com", "def.example.com"},
		},
		{
			name:        "bad host generates cert for no host",
			hosts:       []string{"--abc.example.com"},
			expectedErr: "x509: certificate is not valid for any names, but wanted to match --abc.example.com",
		},
		{
			name:        "one good host and one bad host generates cert for only good host",
			hosts:       []string{"---.example.com", "def.example.com"},
			expectedErr: "x509: certificate is valid for def.example.com, not ---.example.com",
		},
	}

	var serverKey, serverCert bytes.Buffer

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serverCert.Reset()
			serverKey.Reset()
			// Test the function generateCACert.  We can only test normative function
			// and hostnames, everything else is hardcoded.
			err := generateCACert(tc.hosts, &serverKey, &serverCert)
			require.NoError(t, err, "unexpected error generating RSA certificate")

			// Test that the CA certificate is decodable, parseable, and has the configured hostname/s.
			block, _ := pem.Decode(serverCert.Bytes())
			if block == nil {
				require.FailNow(t, "failed to decode PEM block containing cert")
			}
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				require.NoError(t, err, "failed to parse certificate")
				for _, h := range tc.hosts {
					if err = cert.VerifyHostname(h); err != nil {
						require.EqualValues(t, tc.expectedErr, err.Error(), "certificate verification failed")
					} else if len(tc.hosts) < 2 && err == nil && tc.expectedErr != "" {
						require.EqualValues(t, tc.expectedErr, nil, "expected an error but certification verification succeeded")
					}
				}
			}
			// Test that the server key is decodable and parseable.
			block, _ = pem.Decode(serverKey.Bytes())
			if block == nil {
				require.FailNow(t, "failed to decode PEM block containing public key")
			}
			if block.Type == "RSA PRIVATE KEY" {
				_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
				require.NoError(t, err, "failed to parse key")
			}
		})
	}
}
