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

	"github.com/stretchr/testify/require"
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
	ConformanceTests = append(ConformanceTests, ListenerSetAllowedRoutesNamespaces)
}

var ListenerSetAllowedRoutesNamespaces = suite.ConformanceTest{
	ShortName:   "ListenerSetAllowedRoutesNamespaces",
	Description: "ListenerSet listeners allow routes from the specified namespace",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
		features.SupportReferenceGrant,
	},
	Manifests: []string{
		"tests/listenerset-allowed-routes-namespaces.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		acceptedListenerConditions := []metav1.Condition{
			{
				Type:   string(gatewayv1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionProgrammed),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionConflicted),
				Status: metav1.ConditionFalse,
				Reason: string(gatewayv1.ListenerReasonNoConflicts),
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-with-listener-sets", Namespace: ns}
		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "gateway-listener"))
		require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAccepted),
			Status: metav1.ConditionTrue,
		})
		kubernetes.GatewayMustHaveAttachedListeners(t, suite.Client, suite.TimeoutConfig, gwNN, 1)

		// listenerset-test-allowed-routes
		routes := []types.NamespacedName{
			{Name: "route-in-same-namespace", Namespace: ns},
			{Name: "route-in-selected-namespace", Namespace: "gateway-api-example-allowed-ns"},
			{Name: "route-not-in-selected-namespace", Namespace: "gateway-api-example-not-allowed-ns"},
		}
		listenerSetGK := schema.GroupKind{
			Group: gatewayxv1a1.GroupVersion.Group,
			Kind:  "XListenerSet",
		}
		lsNN := types.NamespacedName{Name: "listenerset-test-allowed-routes", Namespace: ns}
		listenerSetRef := kubernetes.NewResourceRef(listenerSetGK, lsNN)
		kubernetes.RoutesAndParentMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, listenerSetRef, &gatewayv1.HTTPRoute{}, routes...)
		kubernetes.ListenerSetStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, lsNN, []gatewayxv1a1.ListenerEntryStatus{
			{
				Name:           "listener-set-listener-allowed-routes-all",
				SupportedKinds: generateSupportedRouteKinds(),
				// This attaches to route-in-same-namespace, route-in-selected-namespace, route-not-in-selected-namespace
				AttachedRoutes: 3,
				Conditions:     acceptedListenerConditions,
			},
			{
				Name:           "listener-set-listener-allowed-routes-same",
				SupportedKinds: generateSupportedRouteKinds(),
				// This attaches to route-in-same-namespace
				AttachedRoutes: 1,
				Conditions:     acceptedListenerConditions,
			},
			{
				Name:           "listener-set-listener-allowed-routes-selector",
				SupportedKinds: generateSupportedRouteKinds(),
				// This attaches to route-in-selected-namespace
				AttachedRoutes: 1,
				Conditions:     acceptedListenerConditions,
			},
		})

		testCases := []http.ExpectedResponse{
			// Requests to all the routes on `listener-set-listener-allowed-routes-all` should succeed
			{
				Request:   http.Request{Host: "listener-set-listener-allowed-routes-all.com", Path: "/route-in-same-namespace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listener-set-listener-allowed-routes-all.com", Path: "/route-in-selected-namespace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listener-set-listener-allowed-routes-all.com", Path: "/route-not-in-selected-namespace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests only to the route in the same namespace on `listener-set-listener-allowed-routes-same` should succeed
			{
				Request:   http.Request{Host: "listener-set-listener-allowed-routes-same.com", Path: "/route-in-same-namespace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "listener-set-listener-allowed-routes-same.com", Path: "/route-in-selected-namespace"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "listener-set-listener-allowed-routes-same.com", Path: "/route-not-in-selected-namespace"},
				Response: http.Response{StatusCode: 404},
			},
			// Requests only to the route in the selected namespace on `listener-set-listener-allowed-routes-selector` should succeed
			{
				Request:  http.Request{Host: "listener-set-listener-allowed-routes-selector.com", Path: "/route-in-same-namespace"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "listener-set-listener-allowed-routes-selector.com", Path: "/route-in-selected-namespace"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "listener-set-listener-allowed-routes-selector.com", Path: "/route-not-in-selected-namespace"},
				Response: http.Response{StatusCode: 404},
			},
		}

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

// func gatewayToParentRef(gateway types.NamespacedName) gatewayv1.ParentReference {
// 	var (
// 		group     = gatewayv1.Group(gatewayv1.GroupName)
// 		kind      = gatewayv1.Kind("Gateway")
// 		namespace = gatewayv1.Namespace(gateway.Namespace)
// 		name      = gatewayv1.ObjectName(gateway.Name)
// 	)

// 	return gatewayv1.ParentReference{
// 		Group:     &group,
// 		Kind:      &kind,
// 		Namespace: &namespace,
// 		Name:      name,
// 	}
// }

// func listenerSetToParentRef(listenerSet types.NamespacedName) gatewayv1.ParentReference {
// 	var (
// 		group     = gatewayv1.Group(gatewayxv1a1.GroupName)
// 		kind      = gatewayv1.Kind("XListenerSet")
// 		namespace = gatewayv1.Namespace(listenerSet.Namespace)
// 		name      = gatewayv1.ObjectName(listenerSet.Name)
// 	)

// 	return gatewayv1.ParentReference{
// 		Group:     &group,
// 		Kind:      &kind,
// 		Namespace: &namespace,
// 		Name:      name,
// 	}
// }

// func generateAcceptedRouteParentStatus(controllerName string, parentRefs ...gatewayv1.ParentReference) []gatewayv1.RouteParentStatus {
// 	var routeParentStatus []gatewayv1.RouteParentStatus
// 	for _, parent := range parentRefs {
// 		routeParentStatus = append(routeParentStatus, gatewayv1.RouteParentStatus{
// 			ParentRef:      parent,
// 			ControllerName: gatewayv1.GatewayController(controllerName),
// 			Conditions: []metav1.Condition{{
// 				Type:   string(gatewayv1.RouteConditionAccepted),
// 				Status: metav1.ConditionTrue,
// 				Reason: string(gatewayv1.RouteReasonAccepted),
// 			}},
// 		})
// 	}
// 	return routeParentStatus
// }

func generateSupportedRouteKinds() []gatewayv1.RouteGroupKind {
	return []gatewayv1.RouteGroupKind{{
		Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
		Kind:  gatewayv1.Kind("HTTPRoute"),
	}, {
		Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
		Kind:  gatewayv1.Kind("GRPCRoute"),
	}}
}
