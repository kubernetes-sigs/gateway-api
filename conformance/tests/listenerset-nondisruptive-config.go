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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetNonDisruptiveConfig)
}

var ListenerSetNonDisruptiveConfig = suite.ConformanceTest{
	ShortName:   "ListenerSetNonDisruptiveConfig",
	Description: "Non-disruptive configuration changes for ListenerSets should not cause traffic disruption",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportListenerSet,
		features.SupportGatewayNonDisruptiveConfig,
	},
	Manifests: []string{
		"tests/listenerset-nondisruptive-config.yaml",
	},
	Slow: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, []string{ns})

		listenerSetGK := schema.GroupKind{
			Group: gatewayv1.GroupVersion.Group,
			Kind:  "ListenerSet",
		}

		t.Run("should delete conflicted ListenerSet without disrupting traffic", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "nondisruptive-listenerset-conflict", Namespace: ns}
			lsAlphaNN := types.NamespacedName{Name: "nondisruptive-ls-alpha", Namespace: ns}
			lsBetaNN := types.NamespacedName{Name: "nondisruptive-ls-beta", Namespace: ns}
			routeNN := types.NamespacedName{Name: "nondisruptive-conflict-route", Namespace: ns}

			// Verify the Gateway is accepted
			kubernetes.GatewayMustHaveCondition(t, s.Client, s.TimeoutConfig, gwNN, metav1.Condition{
				Type:   string(gatewayv1.GatewayConditionAccepted),
				Status: metav1.ConditionTrue,
			})

			// Get gateway address (route is attached to ListenerSets, not Gateway directly)
			gwAddr, err := kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
			require.NoErrorf(t, err, "error waiting for Gateway address")

			// Verify nondisruptive-ls-alpha is accepted with listeners accepted
			kubernetes.ListenerSetMustHaveCondition(t, s.Client, s.TimeoutConfig, lsAlphaNN, metav1.Condition{
				Type:   string(gatewayv1.ListenerSetConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gatewayv1.ListenerSetReasonAccepted),
			})
			kubernetes.ListenerSetListenersMustHaveConditions(t, s.Client, s.TimeoutConfig, lsAlphaNN, generateAcceptedListenerConditions(), "shared-listener")

			// Verify nondisruptive-ls-beta listener has Conflicted=True
			kubernetes.ListenerSetListenersMustHaveConditions(t, s.Client, s.TimeoutConfig, lsBetaNN, []metav1.Condition{
				{
					Type:   string(gatewayv1.ListenerConditionConflicted),
					Status: metav1.ConditionTrue,
					Reason: string(gatewayv1.ListenerReasonHostnameConflict),
				},
			}, "shared-listener")

			// Verify route is accepted by nondisruptive-ls-alpha
			kubernetes.RoutesAndParentMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewResourceRef(listenerSetGK, lsAlphaNN), &gatewayv1.HTTPRoute{}, routeNN)

			// Verify initial traffic works
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Host: "nondisruptive-shared.example.com", Path: "/test"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			})

			// Start continuous traffic
			stop := continuousTraffic(t, s.RoundTripper, gwAddr, "nondisruptive-shared.example.com", "/test")

			// Baseline traffic period
			time.Sleep(1 * time.Second)

			// Delete the winning (alpha) ListenerSet
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()
			lsAlpha := &gatewayv1.ListenerSet{}
			err = s.Client.Get(ctx, lsAlphaNN, lsAlpha)
			require.NoErrorf(t, err, "error getting ListenerSet nondisruptive-ls-alpha")
			err = s.Client.Delete(ctx, lsAlpha)
			require.NoErrorf(t, err, "error deleting ListenerSet nondisruptive-ls-alpha")

			// Verify nondisruptive-ls-beta listener becomes accepted (conflict resolved)
			kubernetes.ListenerSetListenersMustHaveConditions(t, s.Client, s.TimeoutConfig, lsBetaNN, generateAcceptedListenerConditions(), "shared-listener")

			// Verify route is now accepted by nondisruptive-ls-beta
			kubernetes.RoutesAndParentMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewResourceRef(listenerSetGK, lsBetaNN), &gatewayv1.HTTPRoute{}, routeNN)

			// Post-mutation stability period
			time.Sleep(2 * time.Second)

			// Stop traffic and verify zero failures
			result := stop()
			tlog.Logf(t, "traffic results: total=%d failed=%d", result.TotalRequests, result.FailedRequests)
			require.Positive(t, result.TotalRequests, "expected at least some traffic to have been sent")
			require.Equal(t, int64(0), result.FailedRequests, "expected zero failed requests during non-disruptive config change")
		})

		t.Run("should migrate route from Gateway to ListenerSet without disrupting traffic", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "nondisruptive-route-migration", Namespace: ns}
			lsNN := types.NamespacedName{Name: "nondisruptive-migration-ls", Namespace: ns}
			routeNN := types.NamespacedName{Name: "nondisruptive-migration-route", Namespace: ns}

			// Verify Gateway and route accepted
			gwRoutes := []types.NamespacedName{routeNN}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewGatewayRef(gwNN), gwRoutes...)

			// Verify ListenerSet is accepted
			kubernetes.ListenerSetMustHaveCondition(t, s.Client, s.TimeoutConfig, lsNN, metav1.Condition{
				Type:   string(gatewayv1.ListenerSetConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gatewayv1.ListenerSetReasonAccepted),
			})

			// Verify initial traffic works
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request:   http.Request{Host: "something.nondisruptive-app.example.com", Path: "/test"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			})

			// Start continuous traffic
			stop := continuousTraffic(t, s.RoundTripper, gwAddr, "something.nondisruptive-app.example.com", "/test")

			// Baseline traffic period
			time.Sleep(1 * time.Second)

			// Patch HTTPRoute parentRef: change from Gateway to ListenerSet
			ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
			defer cancel()
			original := &gatewayv1.HTTPRoute{}
			err := s.Client.Get(ctx, routeNN, original)
			require.NoErrorf(t, err, "error getting HTTPRoute")

			mutate := original.DeepCopy()
			require.GreaterOrEqual(t, len(mutate.Spec.ParentRefs), 1, "expected at least one parentRef in HTTPRoute %s", routeNN.String())
			lsGroup := gatewayv1.Group(gatewayv1.GroupVersion.Group)
			lsKind := gatewayv1.Kind("ListenerSet")
			lsNamespace := gatewayv1.Namespace(ns)
			mutate.Spec.ParentRefs = []gatewayv1.ParentReference{
				{
					Group:     &lsGroup,
					Kind:      &lsKind,
					Name:      gatewayv1.ObjectName(lsNN.Name),
					Namespace: &lsNamespace,
				},
			}
			err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
			require.NoErrorf(t, err, "error patching HTTPRoute parentRef")

			// Verify route is now accepted by the ListenerSet
			kubernetes.RoutesAndParentMustBeAccepted(t, s.Client, s.TimeoutConfig, s.ControllerName, kubernetes.NewResourceRef(listenerSetGK, lsNN), &gatewayv1.HTTPRoute{}, routeNN)

			// Post-mutation stability period
			time.Sleep(2 * time.Second)

			// Stop traffic and verify zero failures
			result := stop()
			tlog.Logf(t, "traffic results: total=%d failed=%d", result.TotalRequests, result.FailedRequests)
			require.Positive(t, result.TotalRequests, "expected at least some traffic to have been sent")
			require.Equal(t, int64(0), result.FailedRequests, "expected zero failed requests during route migration")
		})
	},
}
