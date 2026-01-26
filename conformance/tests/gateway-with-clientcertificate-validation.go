/*
Copyright 2026 The Kubernetes Authors.

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
	cryptotls "crypto/tls"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayFrontendClientCertificateValidation)
}

var GatewayFrontendClientCertificateValidation = suite.ConformanceTest{
	ShortName:   "GatewayFrontendClientCertificateValidation",
	Description: "Gateway's Client Certificate Validation Config should be used for HTTPS traffic",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportGatewayFrontendClientCertificateValidation,
	},
	Manifests: []string{"tests/gateway-with-clientcertificate-validation.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "client-certificate-validation-https-test", Namespace: ns}

		gwNN := types.NamespacedName{Name: "client-validation-default", Namespace: ns}
		// use gateway Adddress without port because we have 2 HTTPS listeners with different port
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.HTTPRoute{}, false, routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		// Get Server certificate, this certificate is the same for both listeners
		certNN := types.NamespacedName{Name: "tls-validity-checks-certificate", Namespace: ns}
		serverCertPem, _, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		// Get Client Certificate for default configuration
		clientCertNN := types.NamespacedName{Name: "tls-validity-checks-client-certificate", Namespace: ns}
		clientCertPem, clientCertKey, err := GetTLSSecret(suite.Client, clientCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		certificate, err := cryptotls.X509KeyPair(clientCertPem, clientCertKey)
		if err != nil {
			t.Errorf("unexpected error creating client cert: %v", err)
		}

		// getValidClientCert is a hook called when Server asks for Client Certificate during TLS Handshake.
		// This function verifies that Client Certificate has been requested.
		getClientCertID := 0

		// nolint:unparam
		getValidClientCert := func(*cryptotls.CertificateRequestInfo) (*cryptotls.Certificate, error) {
			t.Log("getClientCertertificate was called")
			getClientCertID = 1

			return &certificate, nil
		}

		defaultAddr := gwAddr + ":443"
		perPortAddr := gwAddr + ":8443"

		// Send request with default client certificate and validate that default
		// Client Certificate configuration was applied to the first listener only.
		t.Run("Validate default configuration", func(t *testing.T) {
			expectedSuccess := http.ExpectedResponse{
				Request:                  http.Request{Host: "example.org", Path: "/"},
				Response:                 http.Response{StatusCode: 200},
				Backend:                  "infra-backend-v1",
				Namespace:                "gateway-conformance-infra",
				GetClientCertificateHook: getValidClientCert,
			}
			// send request to the first listener and validate that it is passing
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, defaultAddr, serverCertPem, clientCertPem, clientCertKey, "example.org", expectedSuccess)
			if getClientCertID != 1 {
				t.Errorf("Client Certificate was not presented during the handshake to default")
			}
			// reset this value before TLS handshake
			getClientCertID = 0
			expectedFailure := http.ExpectedResponse{
				Request:                  http.Request{Host: "second-example.org", Path: "/"},
				Namespace:                "gateway-conformance-infra",
				GetClientCertificateHook: getValidClientCert,
			}
			// send request to the second listener and validate that it is failing
			tls.MakeTLSRequestAndExpectFailureResponse(t, suite.RoundTripper, perPortAddr, serverCertPem, clientCertPem, clientCertKey, "second-example.org", expectedFailure)
		})

		// Get Client Certificate for per port configuration
		clientCertPerPortNN := types.NamespacedName{Name: "tls-validity-checks-per-port-client-certificate", Namespace: ns}
		clientCertPerPortPem, clientCertPerPortKey, err := GetTLSSecret(suite.Client, clientCertPerPortNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		perPortCertificate, err := cryptotls.X509KeyPair(clientCertPerPortPem, clientCertPerPortKey)
		if err != nil {
			t.Errorf("unexpected error creating client cert: %v", err)
		}
		// nolint:unparam
		getValidPerPortClientCert := func(*cryptotls.CertificateRequestInfo) (*cryptotls.Certificate, error) {
			t.Log("getValidPerPortClientCert was called")
			getClientCertID = 2
			//
			return &perPortCertificate, nil
		}

		// Send request with per port client certificate and validate that per port
		// Client Certificate configuration was applied to the second listener only.
		t.Run("Validate per port configuration", func(t *testing.T) {
			// reset variable before TLS handshake
			getClientCertID = 0
			expectedSucces := http.ExpectedResponse{
				Request:                  http.Request{Host: "second-example.org", Path: "/"},
				Response:                 http.Response{StatusCode: 200},
				Backend:                  "infra-backend-v2",
				Namespace:                "gateway-conformance-infra",
				GetClientCertificateHook: getValidPerPortClientCert,
			}
			// send request to the second listener and validate that it is passing
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, perPortAddr, serverCertPem, clientCertPerPortPem, clientCertPerPortKey, "second-example.org", expectedSucces)
			if getClientCertID != 2 {
				t.Errorf("Client Certificate was not presented during the handshake to per port listener")
			}

			// reset this value before TLS handshake
			getClientCertID = 0
			expectedFailure := http.ExpectedResponse{
				Request:                  http.Request{Host: "example.org", Path: "/"},
				Namespace:                "gateway-conformance-infra",
				GetClientCertificateHook: getValidPerPortClientCert,
			}
			// send request to the first listener and validate that it is failing
			tls.MakeTLSRequestAndExpectFailureResponse(t, suite.RoundTripper, defaultAddr, serverCertPem, clientCertPerPortPem, clientCertPerPortKey, "example.org", expectedFailure)
		})
	},
}
