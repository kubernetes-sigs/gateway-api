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
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetAllowedRoutesSupportedKinds)
}

var ListenerSetAllowedRoutesSupportedKinds = suite.ConformanceTest{
	ShortName:   "ListenerSetAllowedRoutesSupportedKinds",
	Description: "ListenerSet listeners allow specific route kinds",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerSet,
		features.SupportHTTPRoute,
		features.SupportTLSRoute,
	},
	Manifests: []string{
		"tests/listenerset-allowed-routes-supported-kinds.yaml",
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
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAccepted),
			Status: metav1.ConditionTrue,
		})
		kubernetes.GatewayMustHaveAttachedListeners(t, suite.Client, suite.TimeoutConfig, gwNN, 1)

		// listenerset-test-allowed-routes
		routes := []types.NamespacedName{
			{Name: "http-route", Namespace: ns},
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
				Name: "listener-set-listener-allowed-routes-http-only",
				SupportedKinds: []gatewayv1.RouteGroupKind{{
					Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
					Kind:  gatewayv1.Kind("HTTPRoute"),
				}},
				// This attaches to route-in-same-namespace, route-in-selected-namespace, route-not-in-selected-namespace
				AttachedRoutes: 1,
				Conditions:     acceptedListenerConditions,
			},
		})
	},
}
