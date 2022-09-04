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

package tls

import (
	"crypto/tls"
	"fmt"
	"testing"
)

type DialInfo struct {
	Host       string
	Port       string
	ServerName string
}

type ExpectedOutput struct {
	SANOrCommonName string
}

type TestCase struct {
	DialInfo
	ExpectedOutput
}

func InitiateTLSHandShakeAndValidateSNIMatch(t *testing.T, tlsDialInfo DialInfo, expectedOutput ExpectedOutput) {
	t.Logf("Initiating TLS Handshake to %s with sever_name(SNI): %s", fmt.Sprintf("%s:%s", tlsDialInfo.Host, tlsDialInfo.Port), tlsDialInfo.ServerName)
	tlsConn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", tlsDialInfo.Host, tlsDialInfo.Port),
		&tls.Config{ServerName: tlsDialInfo.ServerName, MinVersion: tls.VersionTLS12})
	if err != nil {
		t.Logf("TLS connection failed : %v", err)
	}
	defer tlsConn.Close()

	err = tlsConn.Handshake()
	if err != nil {
		t.Logf("TLS handshake failed : %v", err)
	}

	for _, san := range tlsConn.ConnectionState().PeerCertificates[0].DNSNames {
		if san == expectedOutput.SANOrCommonName {
			t.Log("SNI matching passed")
			return
		}
	}
	if tlsConn.ConnectionState().PeerCertificates[0].Subject.CommonName == expectedOutput.SANOrCommonName {
		t.Log("SNI matching passed")
		return
	}
	t.Fail()
}
