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
	ConformanceTests = append(ConformanceTests, HTTPRouteExternalAuthPerRouteScope)
}

var HTTPRouteExternalAuthPerRouteScope = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteExternalAuthPerRouteScope",
	Description: "ExternalAuth filter is applied only to the route that declares it; routes on the same listener that have no ExternalAuth filter must be reachable without auth",
	Manifests:   []string{"tests/httproute-external-auth-per-route-scope.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteExternalAuth,
		features.SupportHTTPRouteExternalAuthHTTP,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		withAuthNN := types.NamespacedName{Name: "external-auth-per-route-scope-with-auth", Namespace: ns}
		withoutAuthNN := types.NamespacedName{Name: "external-auth-per-route-scope-without-auth", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), withAuthNN, withoutAuthNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, withAuthNN, gwNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, withoutAuthNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				TestCaseName: "route with ExternalAuth: allowed path reaches backend",
				Request:      http.Request{Path: "/http/allowed"},
				Backend:      confsuite.InfraBackendServiceNameV1,
				Namespace:    ns,
				Response:     http.Response{StatusCode: 200},
			},
			{
				// Route "without-auth" matches /http/denied (more specific than /http),
				// so this request bypasses the ExternalAuth filter entirely.
				// If ExternalAuth were applied at HCM/listener level instead of per-route,
				// the auth server would deny this path and the gateway would return 403.
				TestCaseName: "route without ExternalAuth: request reaches backend even on a path the auth server would deny",
				Request:      http.Request{Path: "/http/denied"},
				Backend:      confsuite.InfraBackendServiceNameV1,
				Namespace:    ns,
				Response:     http.Response{StatusCode: 200},
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
