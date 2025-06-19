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

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"

	h "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicy)
}

var BackendTLSPolicy = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicy",
	Description: "A single service that is targeted by a BackendTLSPolicy must successfully complete TLS termination",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backendtlspolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "gateway-conformance-infra-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-backendtlspolicy", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAcceptedMultipleListeners(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		serverStr := "abc.example.com"

		// Verify that the response to a backend-tls-only call to /backendTLS will return the matching SNI.
		t.Run("Simple HTTP request targeting BackendTLSPolicy should reach infra-backend", func(t *testing.T) {
			h.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				h.ExpectedResponse{
					Namespace: ns,
					Request: h.Request{
						Host: serverStr,
						Path: "/backendTLS",
						SNI:  serverStr,
					},
					Response: h.Response{StatusCode: 200},
				})
		})

		// For the re-encrypt case, we  need to use the cert for the frontend tls listener.
		certNN := types.NamespacedName{Name: "tls-checks-certificate", Namespace: ns}
		cPem, keyPem, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		// Verify that the response to a re-encrypted call to /backendTLS will return the matching SNI.
		t.Run("Re-encrypt HTTPS request targeting BackendTLSPolicy should reach infra-backend", func(t *testing.T) {
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, cPem, keyPem, serverStr,
				h.ExpectedResponse{
					Namespace: ns,
					Request: h.Request{
						Host: serverStr,
						Path: "/backendTLS",
						SNI:  serverStr,
					},
					Response: h.Response{StatusCode: 200},
				})
		})

	},
}
