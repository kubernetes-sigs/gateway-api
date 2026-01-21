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
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayxv1a1 "sigs.k8s.io/gateway-api/apisx/v1alpha1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetProtocolConflict)
}

var ListenerSetProtocolConflict = suite.ConformanceTest{
	ShortName:   "ListenerSetProtocolConflict",
	Description: "Listener Set listener with protocol conflicts to validate Listener Precedence",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
	},
	Manifests: []string{
		"tests/listenerset-protocol-conflict.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the listeners without conflicts should work
			{
				Request:   http.Request{Host: "gateway-listener.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listener-set-1-listener.com", Path: "/listener-set-1-route"},
				Backend:   "infra-backend-v2",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "listener-set-2-listener.com", Path: "/listener-set-2-route"},
				Backend:   "infra-backend-v3",
				Namespace: ns,
			},
			// Requests to the listener with protocol conflict should work on the first listener (based on listener precedence - gateway listener)
			{
				Request:   http.Request{Host: "protocol-conflict-with-gateway-listener.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "protocol-conflict-with-gateway-listener.com", Path: "/listener-set-1-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "protocol-conflict-with-gateway-listener.com", Path: "/listener-set-2-route"},
				Response: http.Response{StatusCode: 404},
			},
			// Requests to the listener with protocol conflict should work on the first listener (based on listener precedence - alphabetic / creation time)
			{
				Request:   http.Request{Host: "protocol-conflict-with-listener-set-listener.com", Path: "/listener-set-1-route"},
				Backend:   "infra-backend-v2",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "protocol-conflict-with-listener-set-listener.com", Path: "/listener-set-2-route"},
				Response: http.Response{StatusCode: 404},
			},
		}

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
		protocolConflictedListenerConditions := []metav1.Condition{
			{
				Type:   string(gatewayv1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.ListenerConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
			{
				Type:   string(gatewayv1.ListenerConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
			{
				Type:   string(gatewayv1.ListenerConditionConflicted),
				Status: metav1.ConditionTrue,
				Reason: string(gatewayv1.ListenerReasonProtocolConflict),
			},
		}

		// Gateway, route and conditions
		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-protocol-conflict", Namespace: ns}
		gwRoutes := []types.NamespacedName{
			{Name: "gateway-route", Namespace: ns},
		}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.HTTPRoute{}, false, gwRoutes...)
		kubernetes.GatewayListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, acceptedListenerConditions, "gateway-listener")
		// The first conflicted listener is accepted based on Listener precedence
		kubernetes.GatewayListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, acceptedListenerConditions, "protocol-conflict-with-gateway-listener")

		// The following listenerSets are accepted since they have at least one valid listener :
		// - listenerset-with-protocol-conflict-with-gateway-1
		// - listenerset-with-protocol-conflict-with-listener-set-1
		// The following listenerSets are not accepted since they do not have at least one valid listener :
		// - listenerset-with-protocol-conflict-with-gateway-2
		// - listenerset-with-protocol-conflict-with-listener-set-2
		kubernetes.GatewayMustHaveAttachedListeners(t, suite.Client, suite.TimeoutConfig, gwNN, 2)

		// listenerset-with-protocol-conflict-with-gateway-1, route and conditions
		lsNN := types.NamespacedName{Name: "listenerset-with-protocol-conflict-with-gateway-1", Namespace: ns}
		lsRoutes := []types.NamespacedName{
			{Namespace: ns, Name: "listenerset-with-protocol-conflict-with-gateway-1-route"},
		}
		for _, routeNN := range lsRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, lsNN)
		}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionTrue,
			// TODO: Maybe this should be just accepted ????
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, acceptedListenerConditions, "listener-set-1-listener")
		// The conflicted listener should not be accepted
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, protocolConflictedListenerConditions, "protocol-conflict-with-gateway-listener")
		// The first conflicted listener is accepted based on Listener precedence
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, acceptedListenerConditions, "protocol-conflict-with-listener-set-listener")

		// listenerset-with-protocol-conflict-with-gateway-2, route and conditions
		lsNN = types.NamespacedName{Name: "listenerset-with-protocol-conflict-with-gateway-2", Namespace: ns}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		// The conflicted listener should not be accepted
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, protocolConflictedListenerConditions, "protocol-conflict-with-gateway-listener")

		// listenerset-with-protocol-conflict-with-listener-set-1, route and conditions
		lsNN = types.NamespacedName{Name: "listenerset-with-protocol-conflict-with-listener-set-1", Namespace: ns}
		lsRoutes = []types.NamespacedName{
			{Namespace: ns, Name: "listenerset-with-protocol-conflict-with-listener-set-1-route"},
		}
		for _, routeNN := range lsRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, lsNN)
		}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionTrue,
			// TODO: Maybe this should be just accepted ????
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, acceptedListenerConditions, "listener-set-2-listener")
		// The conflicted listener should not be accepted
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, protocolConflictedListenerConditions, "protocol-conflict-with-listener-set-listener")

		// listenerset-with-protocol-conflict-with-listener-set-2, route and conditions
		lsNN = types.NamespacedName{Name: "listenerset-with-protocol-conflict-with-listener-set-2", Namespace: ns}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayxv1a1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayxv1a1.ListenerSetReasonListenersNotValid),
		})
		// The conflicted listener should not be accepted
		kubernetes.ListenerSetListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, lsNN, protocolConflictedListenerConditions, "protocol-conflict-with-listener-set-listener")

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
