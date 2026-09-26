/*
Copyright The Kubernetes Authors.

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
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthHTTPResponseHeaders)
}

var HTTPRouteExternalAuthHTTPResponseHeaders = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthHTTPResponseHeaders",
	Description: "Headers in the HTTP auth response are forwarded to the upstream backend (via allowedResponseHeaders) and passed through to the client on denial",
	Manifests:   []string{"tests/httproute-external-auth-http-response-headers.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-http-response-headers", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "headers listed in allowedResponseHeaders should be added to the upstream request on approval",
				Request: http.Request{
					Path: "/http/allowed",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/http/allowed",
						Headers: map[string]string{
							"X-User-Id": "42",
						},
					},
				},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
				Response:  http.Response{StatusCode: 200},
			},
			{
				TestCaseName: "headers from the auth server 403 response should be present on the client response",
				Request: http.Request{
					Path: "/http/denied",
				},
				Response: http.Response{
					StatusCode: 403,
					Headers: map[string]string{
						"X-Auth-Error": "access-denied",
					},
					Body: "Access denied by external auth server",
				},
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
