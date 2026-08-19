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
	ConformanceTests = append(ConformanceTests, HTTPRouteRetryConnectionError)
}

var HTTPRouteRetryConnectionError = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteRetryConnectionError",
	Description: "An HTTPRoute configured with a Retry policy should retry requests that fail due to TCP connection resets up to the configured maximum number of attempts, returning a successful response only when the backend recovers within the retry budget.",
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

		// use dedicated time out config with an increased number of required consecutive successes,
		// so that a test run is very unlikely to miss the unhealthy backend
		dedicatedTimeoutConfig := suite.TimeoutConfig
		dedicatedTimeoutConfig.RequiredConsecutiveSuccesses = 10

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, dedicatedTimeoutConfig, gwAddr, http.ExpectedResponse{
			Request:   http.Request{Path: "/retry-on-connection-errors"},
			Response:  http.Response{StatusCode: 200},
			Backend:   "infra-backend-connection-error-healthy",
			Namespace: confsuite.InfrastructureNamespace,
		})
	},
}
