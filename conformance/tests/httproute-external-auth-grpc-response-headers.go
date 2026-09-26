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
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthGRPCResponseHeaders)
}

var HTTPRouteExternalAuthGRPCResponseHeaders = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthGRPCResponseHeaders",
	Description: "Headers in the gRPC auth OkHttpResponse are forwarded to the upstream backend and headers in DeniedHttpResponse are passed through to the client",
	Manifests:   []string{"tests/httproute-external-auth-grpc-response-headers.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthGRPC,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-grpc-response-headers", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "headers in OkHttpResponse should be added to the upstream request on approval",
				Request: http.Request{
					Path: "/grpc/allowed/upstream",
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/grpc/allowed/upstream",
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
				TestCaseName: "headers from the auth server DeniedHttpResponse should be present on the 403 response to the client",
				Request: http.Request{
					Path: "/grpc/denied/response-headers",
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
