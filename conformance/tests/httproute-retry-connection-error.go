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

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteRetryConnectionError)
}

var HTTPRouteRetryConnectionError = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteRetryConnectionError",
	Description: "An HTTPRoute configured with a Retry policy should retry requests that fail due to TCP connection resets up to the configured maximum number of attempts, returning a successful response only when the backend recovers within the retry budget, and should disable retries when attempts is 0.",
	Manifests:   []string{"tests/httproute-retry-connection-error.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteRetryConnectionError,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "retries-connection-error", Namespace: confsuite.InfrastructureNamespace}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: confsuite.InfrastructureNamespace}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{confsuite.InfrastructureNamespace})

		t.Run("retry with attempts: 100 retries on connection error until healthy backend is reached", func(t *testing.T) {
			// use dedicated time out config with an increased number of required consecutive successes,
			// so that a test run is very unlikely to miss the unhealthy backend
			dedicatedTimeoutConfig := suite.TimeoutConfig
			dedicatedTimeoutConfig.RequiredConsecutiveSuccesses = 10

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, dedicatedTimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/retry-on-connection-errors",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   "infra-backend-connection-error-healthy",
				Namespace: confsuite.InfrastructureNamespace,
			})
		})

		t.Run("retry with attempts: 0 disables retries on connection error", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/retry-on-connection-errors-attempts-0",
				},
				Response:  http.Response{StatusCode: 200},
				Backend:   "infra-backend-connection-error-healthy",
				Namespace: confsuite.InfrastructureNamespace,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

			awaitTimeoutConfig := suite.TimeoutConfig
			awaitTimeoutConfig.RequiredConsecutiveSuccesses = 1

			http.AwaitConvergence(t, awaitTimeoutConfig, func(elapsed time.Duration) bool {
				const batchSize = 25
				var (
					successCount int
					failureCount int
				)

				for range batchSize {
					cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						// Gateway closed connection or connection error between gateway and client
						// when contacting broken backend.
						failureCount++
						continue
					}

					if cRes.StatusCode == 200 {
						if err := http.CompareRoundTrip(t, &req, cReq, cRes, expectedResponse); err != nil {
							tlog.Logf(t, "Response expectation failed for 200 OK request: %v (after %v)", err, elapsed)
							return false
						}
						successCount++
					} else if cRes.StatusCode >= 500 || roundtripper.IsTimeoutError(cRes.StatusCode) {
						// Gateway returned a failure response (e.g. 502 Bad Gateway or 503 Service Unavailable)
						// because connection to the broken backend failed and retries are disabled.
						failureCount++
					} else {
						// Any other status code (such as 404 Not Found) indicates route is not ready yet.
						tlog.Logf(t, "Unexpected status code %d (after %v)", cRes.StatusCode, elapsed)
						return false
					}
				}

				if successCount == 0 {
					tlog.Logf(t, "No successful responses observed in batch of %d requests (after %v)", batchSize, elapsed)
					return false
				}

				if failureCount == 0 {
					tlog.Logf(t, "No failure responses observed in batch of %d requests; retries may not be disabled (after %v)", batchSize, elapsed)
					return false
				}

				tlog.Logf(t, "Observed %d successes and %d failures in batch of %d requests (retries are disabled)", successCount, failureCount, batchSize)
				return true
			})
		})
	},
}
