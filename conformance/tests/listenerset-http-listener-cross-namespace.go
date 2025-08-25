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
	ConformanceTests = append(ConformanceTests, ListenerSetCrossNamespace)
}

var ListenerSetCrossNamespace = suite.ConformanceTest{
	ShortName:   "ListenerSetCrossNamespace",
	Description: "Listener Set in a different namespace than the Gateway",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
		features.SupportReferenceGrant,
	},
	Manifests: []string{
		"tests/listenerset-http-listener-cross-namespace.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		testCases := []http.ExpectedResponse{
			// Requests to the route defined on the gateway (should match all allowed listeners)
			{
				Request:   http.Request{Host: "gateway.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "example.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "foo.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "bar.com", Path: "/gateway-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the route defined on the gateway that targets the example-com listener
			{
				Request:  http.Request{Host: "gateway.com", Path: "/example-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "example.com", Path: "/example-com"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:  http.Request{Host: "foo.com", Path: "/example-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "bar.com", Path: "/example-com"},
				Response: http.Response{StatusCode: 404},
			},
			// Requests to the route defined on the listener set (should only match listeners defined on the allowed listenerset)
			{
				Request:  http.Request{Host: "gateway.com", Path: "/listenerset-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "example.com", Path: "/listenerset-route"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "foo.com", Path: "/listenerset-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			{
				Request:   http.Request{Host: "bar.com", Path: "/listenerset-route"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the route defined on the listenerset that targets the bar-com listener
			{
				Request:  http.Request{Host: "gateway.com", Path: "/bar-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "example.com", Path: "/bar-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:  http.Request{Host: "foo.com", Path: "/bar-com"},
				Response: http.Response{StatusCode: 404},
			},
			{
				Request:   http.Request{Host: "bar.com", Path: "/bar-com"},
				Backend:   "infra-backend-v1",
				Namespace: ns,
			},
			// Requests to the listenerset that does not match the allowed labels should not work
			{
				Request:  http.Request{Host: "baz.com", Path: "/gateway-route"},
				Response: http.Response{StatusCode: 404},
			},
		}

		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-http-listener", Namespace: ns}
		gwRoutes := []types.NamespacedName{
			{Namespace: "gateway-api-example-ns2", Name: "attaches-to-all-listeners"},
			{Namespace: "gateway-api-example-ns3", Name: "attaches-to-example-com-on-gateway"},
		}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), gwRoutes...)
		for _, routeNN := range gwRoutes {
			kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)
		}

		lsNN := types.NamespacedName{Name: "listenerset-with-http-listener", Namespace: "gateway-api-example-ns1"}
		lsRoutes := []types.NamespacedName{
			{Namespace: "gateway-api-example-ns4", Name: "attaches-to-all-listeners-on-listenerset"},
			{Namespace: "gateway-api-example-ns5", Name: "attaches-to-bar-com-on-listenerset"},
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

		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAttachedListenerSets),
			Status: metav1.ConditionTrue,
			Reason: string(gatewayv1.GatewayReasonListenerSetsAttached),
		})
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

		disallowedLsNN := types.NamespacedName{Name: "disallowed-listenerset-with-http-listener", Namespace: "gateway-api-example-ns6"}
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
