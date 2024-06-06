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
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteBackendPathRedirect)
}

var HTTPRouteBackendPathRedirect = suite.ConformanceTest{
	ShortName:   "HTTPRouteBackendPathRedirect",
	Description: "A single HTTPRoute with a redirection filter for different backends",
	Features: []features.SupportedFeature{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteBackendPathRedirect,
	},
	Manifests: []string{"tests/httproute-backend-request-redirect.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "backend-request-redirect", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{{
			Request:   http.Request{Path: "/"},
			Backend:   "infra-backend-v1",
			Namespace: ns,
			Response:  http.Response{StatusCode: 302, Headers: map[string]string{"Location": "https://gateway-api.sigs.k8s.io"}},
		}, {
			Request:   http.Request{Path: "/"},
			Backend:   "infra-backend-v2",
			Namespace: ns,
			Response:  http.Response{StatusCode: 302, Headers: map[string]string{"Location": "http://www.redirect-example.com"}},
		}}

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
