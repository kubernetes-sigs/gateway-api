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
	"fmt"
	"net/url"
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
		features.SupportHTTPRouteRetry,
		features.SupportHTTPRouteRetryConnectionError,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "retries-connection-error", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		type args struct {
			path                  string
			retrySimulationConfig url.Values
		}
		testCases := []struct {
			name string
			args args
			want http.Response
		}{
			{
				name: "succeeds after 2 retries on connection errors and max attempts is 3",
				args: args{
					path: "/retry/no-status-code-attempts-3",
					retrySimulationConfig: url.Values{
						"succeedAfter": []string{"2"},
					},
				},
				want: http.Response{StatusCode: 200},
			},
			{
				name: "fails when required retries on connection errors exceed max attempts",
				args: args{
					path: "/retry/no-status-code-attempts-3",
					retrySimulationConfig: url.Values{
						"succeedAfter": []string{"4"},
					},
				},
				want: http.Response{StatusCodes: []int{500, 503}},
			},
		}
		for i := range testCases {
			tc := testCases[i]
			t.Run(fmt.Sprintf("%d request to '%s' %s", i, tc.args.path, tc.name), func(t *testing.T) {
				assertConsistentRetryBehaviour(t, suite, gwAddr, ns, tc.args.path, tc.args.retrySimulationConfig, tc.want)
			})
		}
	},
}
