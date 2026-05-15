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
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteValidInvalidListeners)
}

var UDPRouteValidInvalidListeners = confsuite.ConformanceTest{
	ShortName:   "UDPRouteValidInvalidListeners",
	Description: "A UDPRoute should attach to the UDP listener on a Gateway that also has a TCP listener on a different port. The TCP listener must remain unattached and not routable.",
	Manifests:   []string{"tests/udproute-valid-invalid-listeners.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportUDPRoute,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "udp-tcp-listeners-gateway", Namespace: ns}
		routeNN := types.NamespacedName{Name: "udp-route-on-mixed-listener-gateway", Namespace: ns}

		// The test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		group := v1.Group(v1.GroupVersion.Group)
		kind := v1.Kind("Gateway")
		gwName := v1.ObjectName(gwNN.Name)
		gwNS := v1.Namespace(ns)
		expectedParents := []v1.RouteParentStatus{{
			ParentRef: v1.ParentReference{
				Group:     &group,
				Kind:      &kind,
				Name:      gwName,
				Namespace: &gwNS,
			},
			ControllerName: v1.GatewayController(suite.ControllerName),
			Conditions: []metav1.Condition{{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(v1.RouteReasonAccepted),
			}},
		}}

		t.Run("UDPRoute attaches to the UDP listener on a Gateway with both TCP and UDP listeners", func(t *testing.T) {
			kubernetes.UDPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN, expectedParents, false)
		})

		t.Run("UDP listener has 1 attached route and TCP listener has 0 attached routes", func(t *testing.T) {
			ready := []metav1.Condition{
				{
					Type:   string(gatewayv1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: "", // any reason
				},
				{
					Type:   string(gatewayv1.ListenerConditionResolvedRefs),
					Status: metav1.ConditionTrue,
					Reason: "", // any reason
				},
			}
			expectedListeners := []gatewayv1.ListenerStatus{
				{
					Name: gatewayv1.SectionName("udp"),
					SupportedKinds: []gatewayv1.RouteGroupKind{{
						Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
						Kind:  gatewayv1.Kind("UDPRoute"),
					}},
					AttachedRoutes: 1,
					Conditions:     ready,
				},
				{
					Name: gatewayv1.SectionName("tcp"),
					SupportedKinds: []gatewayv1.RouteGroupKind{{
						Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
						Kind:  gatewayv1.Kind("TCPRoute"),
					}},
					AttachedRoutes: 0,
					Conditions:     ready,
				},
			}
			kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gwNN, expectedListeners)
		})

		t.Run("UDP traffic is forwarded through the UDP listener to the backend", func(t *testing.T) {
			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
				kubernetes.NewGatewayRef(gwNN, "udp"))
			if err != nil {
				t.Fatalf("error getting gateway address: %v", err)
			}
			expectUDPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
		})
	},
}
