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
	ConformanceTests = append(ConformanceTests, HTTPRouteRetryWithTimeouts)
}

var HTTPRouteRetryWithTimeouts = confsuite.ConformanceTest{
	ShortName:   "HTTPRouteRetryWithTimeouts",
	Description: "An HTTPRoute that has both Retry and Timeout policies configured should retry failed requests only while the configured timeouts permit, returning a successful response when the backend recovers within the timeout budget and surfacing a timeout error when retries or backend delays exceed the request or backend request timeout.",
	Manifests:   []string{"tests/httproute-retry-with-timeouts.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportHTTPRouteRetry,
		features.SupportHTTPRouteRetryBackendTimeout,
		features.SupportHTTPRouteRequestTimeout,
		features.SupportHTTPRouteBackendTimeout,
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "retries-with-timeouts", Namespace: ns}
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
				name: "succeeds after 2 retries on backend timeout with max attempts is 3",
				args: args{
					path: "/retry/backend-request-timeout-200ms",
					retrySimulationConfig: url.Values{
						"responseCode": []string{"500"},
						"succeedAfter": []string{"2"},
						"delayRetry":   []string{"300ms"},
					},
				},
				want: http.Response{StatusCode: 200},
			},
			{
				name: "fails when required retries on backend timeout exceed max attempts",
				args: args{
					path: "/retry/backend-request-timeout-200ms",
					retrySimulationConfig: url.Values{
						"responseCode": []string{"500"},
						"succeedAfter": []string{"3"},
						"delayRetry":   []string{"300ms"},
					},
				},
				want: http.Response{StatusCode: 504},
			},
			{
				name: "succeeds when retries complete within the request timeout",
				args: args{
					path: "/retry/request-timeout-200ms",
					retrySimulationConfig: url.Values{
						"responseCode": []string{"500"},
						"succeedAfter": []string{"1"},
					},
				},
				want: http.Response{StatusCode: 200},
			},
			{
				name: "fails with 504 when retry delays exceed the request timeout",
				args: args{
					path: "/retry/request-timeout-200ms",
					retrySimulationConfig: url.Values{
						"responseCode": []string{"500"},
						"succeedAfter": []string{"4"},
						"delayRetry":   []string{"100ms"},
					},
				},
				want: http.Response{StatusCode: 504},
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
