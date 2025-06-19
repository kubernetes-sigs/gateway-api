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
		expectedErr []string
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
			expectedErr: []string{"x509: certificate is not valid for any names, but wanted to match --abc.example.com"},
		},
		{
			name:        "one good host and one bad host generates cert for only good host",
			hosts:       []string{"---.example.com", "def.example.com"},
			expectedErr: []string{"x509: certificate is valid xxx for def.example.com, not ---.example.com", ""},
		},
	}

	var serverKey, serverCert bytes.Buffer

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			serverCert.Reset()
			serverKey.Reset()
			// Test the function generateCACert.  We can only test normative function
			// and hostnames, everything else is hardcoded.
			_, caBytes, caPrivKey, err := generateCACert(tc.hosts)
			require.NoError(t, err, "unexpected error generating RSA certificate")

			var certData bytes.Buffer
			if err := pem.Encode(&certData, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
				require.NoError(t, err, "failed to create certificater")
			}

			var keyData bytes.Buffer
			if err := pem.Encode(&keyData, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey)}); err != nil {
				require.NoError(t, err, "failed to create key")
			}

			// Test that the CA certificate is decodable, parseable, and has the configured hostname/s.
			block, _ := pem.Decode(certData.Bytes())
			if block == nil {
				require.FailNow(t, "failed to decode PEM block containing cert")
			} else if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				require.NoError(t, err, "failed to parse certificate")
				for idx, h := range tc.hosts {
					err = cert.VerifyHostname(h)
					if err != nil && len(tc.expectedErr) > 0 && tc.expectedErr[idx] == "" {
						require.EqualValues(t, tc.expectedErr[idx], err.Error(), "certificate verification failed")
					} else if err == nil && len(tc.expectedErr) > 0 && tc.expectedErr[idx] != "" {
						require.EqualValues(t, tc.expectedErr[idx], err, "expected an error but certification verification succeeded")
					}
				}
			}
		})
	}
}
