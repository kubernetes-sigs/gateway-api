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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayxv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetSameNamespace)
}

var ListenerSetSameNamespace = suite.ConformanceTest{
	ShortName:   "ListenerSetSameNamespace",
	Description: "ListenerSet in the same namespace as the Gateway",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{
		"tests/listenerset-same-namespace.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the route defined on the gateway (should only match routes attached to the gateway)
			{
				Request:   http.Request{Host: "gateway-listener-1.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "gateway-listener-2.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "listenerset-1-listener-1.com", Path: "/gateway-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "listenerset-1-listener-2.com", Path: "/gateway-route"},
				Response: http.Response{StatusCode: 404},
			},
			// Requests to the route defined on the gateway that targets gateway-listener-2-listener
			{
				Request:  http.Request{Host: "gateway-listener-1.com", Path: "/gateway-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "gateway-listener-2.com", Path: "/gateway-listener-2-listener-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "listenerset-1-listener-1.com", Path: "/gateway-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "listenerset-1-listener-2.com", Path: "/gateway-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			// Requests to the route defined on the listener set (should only match listeners defined on the listenerset)
			{
				Request:  http.Request{Host: "gateway-listener-1.com", Path: "/listenerset-1-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "gateway-listener-2.com", Path: "/listenerset-1-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "listenerset-1-listener-1.com", Path: "/listenerset-1-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listenerset-1-listener-2.com", Path: "/listenerset-1-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the route defined on the listenerset that targets listenerset-1-listener-2-listener
			{
				Request:  http.Request{Host: "gateway-listener-1.com", Path: "/listenerset-1-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "gateway-listener-2.com", Path: "/listenerset-1-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "listenerset-1-listener-1.com", Path: "/listenerset-1-listener-2-listener-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "listenerset-1-listener-2.com", Path: "/listenerset-1-listener-2-listener-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to listenerset-2-listener-1-listener on the listener set in a different namespace should not work
			{
				Request:  http.Request{Host: "listenerset-2-listener-1.com", Path: "/listenerset-2-route"},
				Response: http.Response{StatusCode: 404},
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-http-listener", Namespace: ns}
		gwRoutes := []types.NamespacedName{
			{Namespace: ns, Name: "gateway-route"},
			{Namespace: ns, Name: "gateway-listener-2-route"},
		}

		// Gateway, route and conditions
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), gwRoutes...)
		for _, routeNN := range gwRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAttachedListenerSets),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.GatewayReasonListenerSetsAttached),
		})

		// Allowed ListenerSet, route and conditions
		lsNN := types.NamespacedName{Name: "listenerset-with-http-listener", Namespace: ns}
		lsRoutes := []types.NamespacedName{
			{Namespace: ns, Name: "listenerset-1-route"},
			{Namespace: ns, Name: "listenerset-1-listener-2-listener-route"},
		}
		listenerSetGK := schema.GroupKind{
			Group: gatewayxv1a1.GroupVersion.Group,
			Kind:  "XListenerSet",
		}
		listenerSetRef := kubernetes.NewResourceRef(listenerSetGK, lsNN)
		kubernetes.RoutesAndParentMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, listenerSetRef, &gatewayv1.HTTPRoute{}, lsRoutes...)
		for _, routeNN := range lsRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, lsNN)
		}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonAccepted),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonProgrammed),
		})

		// Disallowed ListenerSet, route and conditions
		disallowedLsNN := types.NamespacedName{Name: "listenerset-not-allowed", Namespace: "gateway-api-example-ns1"}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, disallowedLsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonNotAllowed),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, disallowedLsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonNotAllowed),
		})

		for i := range testCases {
			// Declare tc here to avoid loop variable
			// reuse issues across parallel tests.
			tc := testCases[i]
			t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
				t.Parallel()
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
			})
		}
	},
}
