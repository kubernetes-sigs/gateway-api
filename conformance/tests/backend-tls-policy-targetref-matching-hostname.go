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
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTLSPolicyTargetRefMatchingHostname)
}

var BackendTLSPolicyTargetRefMatchingHostname = suite.ConformanceTest{
	ShortName:   "BackendTLSPolicyTargetRefMatchingHostname",
	Description: "A valid BackendTLSPolicy with a targetRef/servce using CACertificateRef matching for hostname",
	Features: []features.SupportedFeature{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportBackendTLSPolicy,
	},
	Manifests: []string{"tests/backend-tls-policy-targetref-matching-hostname.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {

		// Setup cluster with:
		// 	 gateway: gateway-conformance-infra/tls-backend-gateway with certificateRef backend-tls-checks-certificate and abc.example.com hostname
		//   httpRoute:  gateway-conformance-infra/tls-backend-route
		// 	 deployment: gateway-conformance-infra/tls-backend that serves a cert with common name "www.abc.example.com"
		//   backend service with a cert: gateway-conformance-infra/tls-backend
		// 	 backendTLSPolicy: gateway-conformance-infra/tls-backend-tls-policy with CACertificateRef gateway-conformance-infra/backend-tls-checks-certificate
		// 	 service account?
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "tls-backend-route", Namespace: ns}
		gwNN := types.NamespacedName{Name: "tls-backend-gateway", Namespace: ns}
		certNN := types.NamespacedName{Name: "backend-tls-checks-certificate", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})
		// wait until the specified Gateway has an IP and the specified HTTPRoute is accepted
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		cPem, keyPem, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		t.Run("Simple TLS request to service should reach tls-backend", func(t *testing.T) {
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, cPem, keyPem, "",
				http.ExpectedResponse{
					Request:   http.Request{Host: "abc.example.com", Path: "/"},
					Backend:   "tls-backend",
					Namespace: "gateway-conformance-infra",
				})

		})

		// Check cert details received from the backend

	},
}
