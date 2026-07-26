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
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthGRPCRequestHeadersToAuth)
}

var HTTPRouteExternalAuthGRPCRequestHeadersToAuth = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthGRPCRequestHeadersToAuth",
	Description: "Headers listed in grpc.allowedHeaders are forwarded to the gRPC auth server",
	Manifests:   []string{"tests/httproute-external-auth-grpc-request-headers-to-auth.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthGRPC,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "external-auth-grpc-request-headers-to-auth", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "request header listed in allowedHeaders should arrive at the auth server and be echoed to the backend",
				Request: http.Request{
					Path: "/grpc/allowed",
					Headers: map[string]string{
						"X-Test-Header": "hello",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/grpc/allowed",
						Headers: map[string]string{
							"X-Auth-Received-Test-Header": "hello",
						},
					},
				},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: ns,
				Response:  http.Response{StatusCode: 200},
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
