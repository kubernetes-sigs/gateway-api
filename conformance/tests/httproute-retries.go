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
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteRetries)
}

var HTTPRouteRetries = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteRetries",
	Description: "An HTTPRoute that has a Retry policy configured should retry failed requests according to the specified codes and attempt limits, returning a successful response when the backend recovers within the retry budget and surfacing the original error when it does not.",
	Manifests:   []string{"tests/httproute-retries.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteRetries,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "retries", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		testCases := []http.ExpectedResponse{
			{
				Request: http.Request{
					Path: "/retry/code-500-attempts-3?responseCode=500&succeedAfter=2",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-500-attempts-3?responseCode=500&succeedAfter=4",
				},
				Response:  http.Response{StatusCode: 500},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-500-attempts-3?responseCode=503&succeedAfter=2",
				},
				Response:  http.Response{StatusCode: 503},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=500&succeedAfter=1",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=500&succeedAfter=3",
				},
				Response:  http.Response{StatusCode: 500},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=502&succeedAfter=1",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=502&succeedAfter=3",
				},
				Response:  http.Response{StatusCode: 502},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=503&succeedAfter=1",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=503&succeedAfter=3",
				},
				Response:  http.Response{StatusCode: 503},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=504&succeedAfter=1",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
			{
				Request: http.Request{
					Path: "/retry/code-all-attempts-2?responseCode=504&succeedAfter=3",
				},
				Response:  http.Response{StatusCode: 504},
				Backend:   confsuite.InfraBackendServiceNameV3,
				Namespace: ns,
			},
		}
		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()

				path := tc.Request.Path

				http.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
					tc.Request.Path = path + "&uuid=" + uuid.New().String()
					req := http.MakeRequest(t, &tc, gwAddr, "HTTP", "http")

					cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
						return false
					}

					if err := http.CompareRoundTrip(t, &req, cReq, cRes, tc); err != nil {
						tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
						return false
					}

					return true
				})
				tlog.Logf(t, "Request passed")
			})
		}
	},
}
