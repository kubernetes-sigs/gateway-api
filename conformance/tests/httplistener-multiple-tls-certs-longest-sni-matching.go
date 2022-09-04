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

package tests

import (
	"fmt"
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPSLitenerMultiTLSCertsLongestMatchingSNI)
}

var HTTPSLitenerMultiTLSCertsLongestMatchingSNI = suite.ConformanceTest{
	ShortName:   "HTTPSLitenerMultiTLSCertsLongestMatchingSNI",
	Description: "HTTPS listener references multiple TLS certificates should use longest matching SNI out of all available certificates",
	Manifests:   []string{"tests/httplistener-multiple-tls-certs-longest-sni-matching.yaml"},
	Features:    []suite.SupportedFeature{suite.SupportHTTPListenerMultipleTLSCerts},
	Test: func(t *testing.T, cts *suite.ConformanceTestSuite) {

		testCases := []tls.TestCase{
			{
				DialInfo: tls.DialInfo{
					Host:       "example.com",
					Port:       "443",
					ServerName: "abc.xyz.test.example.com",
				},
				ExpectedOutput: tls.ExpectedOutput{
					SANOrCommonName: "*.xyz.test.example.com",
				},
			},
			{
				DialInfo: tls.DialInfo{
					Host:       "example.com",
					Port:       "443",
					ServerName: "test1.example.com",
				},
				ExpectedOutput: tls.ExpectedOutput{
					SANOrCommonName: "*.example.com",
				},
			},
			{
				DialInfo: tls.DialInfo{
					Host:       "example.com",
					Port:       "443",
					ServerName: "internal.test.example.com",
				},
				ExpectedOutput: tls.ExpectedOutput{
					SANOrCommonName: "*.test.example.com",
				},
			},
		}

		for i, tc := range testCases {
			t.Run(fmt.Sprintf("Test case %d", i+1), func(t *testing.T) {
				tls.InitiateTLSHandShakeAndValidateSNIMatch(t, tc.DialInfo, tc.ExpectedOutput)
			})
		}
	},
}
