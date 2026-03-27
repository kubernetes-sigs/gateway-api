/*
Copyright 2025 The Kubernetes Authors.

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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteNonDisruptiveConfig)
}

var HTTPRouteNonDisruptiveConfig = suite.ConformanceTest{
	ShortName:   "HTTPRouteNonDisruptiveConfig",
	Description: "Non-disruptive configuration changes for HTTPRoutes should not cause traffic disruption",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportGatewayNonDisruptiveConfig,
	},
	Manifests: []string{
		"tests/httproute-nondisruptive-config.yaml",
	},
	Slow: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{ns})

		gwNN := types.NamespacedName{Name: "nondisruptive-httproute", Namespace: ns}

		t.Run("should change route backend without disrupting traffic", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "nondisruptive-backend-change", Namespace: ns}

			// Verify Gateway and route accepted
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// Verify initial traffic hits infra-backend-v1
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/backend-change"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			})

			// Start continuous traffic
			stop := continuousTraffic(t, s.RoundTripper, gwAddr, "", "/backend-change")

			// Baseline traffic period
			time.Sleep(1 * time.Second)

			// Patch HTTPRoute: change backendRef from infra-backend-v1 to infra-backend-v2
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()
			original := &gatewayv1.HTTPRoute{}
			err := s.Client.Get(ctx, routeNN, original)
			require.NoErrorf(t, err, "error getting HTTPRoute")

			mutate := original.DeepCopy()
			require.GreaterOrEqual(t, len(mutate.Spec.Rules), 1, "expected at least one rule in HTTPRoute %s", routeNN.String())
			require.GreaterOrEqual(t, len(mutate.Spec.Rules[0].BackendRefs), 1, "expected at least one backendRef in first rule of HTTPRoute %s", routeNN.String())
			mutate.Spec.Rules[0].BackendRefs[0].Name = "infra-backend-v2"
			err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
			require.NoErrorf(t, err, "error patching HTTPRoute backendRef")

			// Verify eventual consistency: traffic reaches infra-backend-v2
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/backend-change"},
				Backend:   "infra-backend-v2",
				Namespace: ns,
			})

			// Post-mutation stability period
			time.Sleep(2 * time.Second)

			// Stop traffic and verify zero failures
			result := stop()
			tlog.Logf(t, "traffic results: total=%d failed=%d", result.TotalRequests, result.FailedRequests)
			require.Positive(t, result.TotalRequests, "expected at least some traffic to have been sent")
			require.Equal(t, int64(0), result.FailedRequests, "expected zero failed requests during backend change")
		})

		t.Run("should replace route without disrupting traffic", func(t *testing.T) {
			routeANN := types.NamespacedName{Name: "nondisruptive-route-replace-a", Namespace: ns}

			// Verify Gateway and route-A accepted
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewGatewayRef(gwNN), routeANN)

			// Verify initial traffic hits infra-backend-v1
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/route-replace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			})

			// Start continuous traffic
			stop := continuousTraffic(t, s.RoundTripper, gwAddr, "", "/route-replace")

			// Baseline traffic period
			time.Sleep(1 * time.Second)

			// Create HTTPRoute-B programmatically
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()

			routeB := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nondisruptive-route-replace-b",
					Namespace: ns,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: []gatewayv1.ParentReference{
							{
								Name:      gatewayv1.ObjectName(gwNN.Name),
								Namespace: (*gatewayv1.Namespace)(&ns),
							},
						},
					},
					Rules: []gatewayv1.HTTPRouteRule{
						{
							Matches: []gatewayv1.HTTPRouteMatch{
								{
									Path: &gatewayv1.HTTPPathMatch{
										Type:  ptr.To(gatewayv1.PathMatchPathPrefix),
										Value: ptr.To("/route-replace"),
									},
								},
							},
							BackendRefs: []gatewayv1.HTTPBackendRef{
								{
									BackendRef: gatewayv1.BackendRef{
										BackendObjectReference: gatewayv1.BackendObjectReference{
											Name: "infra-backend-v2",
											Port: ptr.To(gatewayv1.PortNumber(8080)),
										},
									},
								},
							},
						},
					},
				},
			}

			err := s.Client.Create(ctx, routeB)
			require.NoErrorf(t, err, "error creating HTTPRoute-B")
			t.Cleanup(func() {
				if cleanupErr := s.Client.Delete(context.Background(), routeB); cleanupErr != nil {
					tlog.Logf(t, "error cleaning up HTTPRoute-B: %v", cleanupErr)
				}
			})

			// Verify route-B is accepted
			routeBNN := types.NamespacedName{Name: "nondisruptive-route-replace-b", Namespace: ns}
			kubernetes.HTTPRouteMustHaveParents(t, s.Client, s.TimeoutConfig, routeBNN, []gatewayv1.RouteParentStatus{
				{
					ParentRef: gatewayv1.ParentReference{
						Name:      gatewayv1.ObjectName(gwNN.Name),
						Namespace: (*gatewayv1.Namespace)(&ns),
					},
					ControllerName: gatewayv1.GatewayController(s.ControllerName),
					Conditions: []metav1.Condition{
						{
							Type:   string(gatewayv1.RouteConditionAccepted),
							Status: metav1.ConditionTrue,
							Reason: string(gatewayv1.RouteReasonAccepted),
						},
					},
				},
			}, false)

			// Delete route-A
			routeA := &gatewayv1.HTTPRoute{}
			err = s.Client.Get(ctx, routeANN, routeA)
			require.NoErrorf(t, err, "error getting HTTPRoute-A")
			err = s.Client.Delete(ctx, routeA)
			require.NoErrorf(t, err, "error deleting HTTPRoute-A")

			// Verify eventual consistency: traffic reaches infra-backend-v2
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Path: "/route-replace"},
				Backend:   "infra-backend-v2",
				Namespace: ns,
			})

			// Post-mutation stability period
			time.Sleep(2 * time.Second)

			// Stop traffic and verify zero failures
			result := stop()
			tlog.Logf(t, "traffic results: total=%d failed=%d", result.TotalRequests, result.FailedRequests)
			require.Positive(t, result.TotalRequests, "expected at least some traffic to have been sent")
			require.Equal(t, int64(0), result.FailedRequests, "expected zero failed requests during route replacement")
		})
	},
}
