/*
Copyright 2022 The Kubernetes Authors.

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
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestSelfSignedCertHasCA(t *testing.T) {
	secret := MustCreateSelfSignedCertSecret(t, "test-ns", "test-secret", []string{"example.com"})

	block, _ := pem.Decode(secret.Data[corev1.TLSCertKey])
	require.NotNil(t, block)

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	require.True(t, cert.IsCA, "self-signed cert used as trust root must have IsCA=true")
	require.NotZero(t, cert.KeyUsage&x509.KeyUsageCertSign, "self-signed cert must have KeyUsageCertSign")
}
