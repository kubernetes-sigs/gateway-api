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
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayxv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetAllowedNamespaceSame)
}

var ListenerSetAllowedNamespaceSame = suite.ConformanceTest{
	ShortName:   "ListenerSetAllowedNamespaceSame",
	Description: "ListenerSet in the same namespace as the Gateway",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{
		"tests/listenerset-allowed-namespace-same.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the route defined on the gateway (should only match routes attached to the gateway)
			{
				Request:   http.Request{Host: "gateway-listener-1.com", Path: "/route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the listenerset listeners in the same namespace as the parent gateway should pass
			{
				Request:   http.Request{Host: "listener-set-1-listener-1.com", Path: "/route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listener-set-1-listener-2.com", Path: "/route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the listenerset listeners in a different namespace than the parent gateway should fail
			{
				Request:  http.Request{Host: "listener-set-2-listener-1.com", Path: "/route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "listener-set-2-listener-2.com", Path: "/route"},
				Response: http.Response{StatusCode: 404},
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-http-listener", Namespace: ns}

		// Gateway and conditions
		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.NewGatewayRef(gwNN))
		require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

		// ListenerSet gateway-conformance-infra/listenerset-with-http-listener is accepted since it is in the same ns as the parent gateway
		// ListenerSet gateway-api-example-ns1/enerset-not-allowed is accepted since it is in a different ns than the parent gateway
		kubernetes.GatewayMustHaveAttachedListeners(t, suite.Client, suite.TimeoutConfig, gwNN, 1)

		// Allowed ListenerSet, route and conditions
		lsNN := types.NamespacedName{Name: "listenerset-with-http-listener", Namespace: ns}
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

		routes := []types.NamespacedName{
			{Name: "attaches-to-all-listeners", Namespace: ns},
		}
		routeParentRefs := []gatewayv1.ParentReference{
			gatewayToParentRef(gwNN),
			listenerSetToParentRef(lsNN),
		}
		for _, routeNN := range routes {
			kubernetes.HTTPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, generateAcceptedRouteParentStatus(suite.ControllerName, routeParentRefs...), true)
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, lsNN)
		}

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

func generateAcceptedRouteParentStatus(controllerName string, parentRefs ...gatewayv1.ParentReference) []gatewayv1.RouteParentStatus {
	var routeParentStatus []gatewayv1.RouteParentStatus
	for _, parent := range parentRefs {
		routeParentStatus = append(routeParentStatus, gatewayv1.RouteParentStatus{
			ParentRef:      parent,
			ControllerName: gatewayv1.GatewayController(controllerName),
			Conditions: []metav1.Condition{{
				Type:   string(gatewayv1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gatewayv1.RouteReasonAccepted),
			}},
		})
	}
	return routeParentStatus
}

func gatewayToParentRef(gateway types.NamespacedName) gatewayv1.ParentReference {
	var (
		group     = gatewayv1.Group(gatewayv1.GroupName)
		kind      = gatewayv1.Kind("Gateway")
		namespace = gatewayv1.Namespace(gateway.Namespace)
		name      = gatewayv1.ObjectName(gateway.Name)
	)

	return gatewayv1.ParentReference{
		Group:     &group,
		Kind:      &kind,
		Namespace: &namespace,
		Name:      name,
	}
}

func listenerSetToParentRef(listenerSet types.NamespacedName) gatewayv1.ParentReference {
	var (
		group     = gatewayxv1a1.Group(gatewayxv1a1.GroupName)
		kind      = gatewayxv1a1.Kind("XListenerSet")
		namespace = gatewayxv1a1.Namespace(listenerSet.Namespace)
		name      = gatewayxv1a1.ObjectName(listenerSet.Name)
	)

	return gatewayv1.ParentReference{
		Group:     &group,
		Kind:      &kind,
		Namespace: &namespace,
		Name:      name,
	}
}
