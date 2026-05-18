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

	v1 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/udp"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteParentRefPortAndSectionName)
}

var UDPRouteParentRefPortAndSectionName = confsuite.ConformanceTest{
	ShortName:   "UDPRouteParentRefPortAndSectionName",
	Description: "A UDPRoute attaches to a UDP listener by port, by sectionName, by both, or to every UDP listener on a Gateway when neither is set.",
	Manifests:   []string{"tests/udproute-parentref-port-and-section-name.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportUDPRoute,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "udp-multi-listener-gateway", Namespace: ns}

		// The test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		group := v1.Group(v1.GroupVersion.Group)
		kind := v1.Kind("Gateway")
		gwName := v1.ObjectName(gwNN.Name)
		gwNS := v1.Namespace(ns)
		acceptedParent := func() v1.RouteParentStatus {
			return v1.RouteParentStatus{
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
			}
		}

		t.Run("UDPRoute attaches to a UDP listener by port", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "udp-route-by-port", Namespace: ns}
			kubernetes.UDPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN,
				[]v1.RouteParentStatus{acceptedParent()}, false)

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
				kubernetes.NewGatewayRef(gwNN, "dns"))
			if err != nil {
				t.Fatalf("error getting gateway address: %v", err)
			}
			udp.ExpectEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
		})

		t.Run("UDPRoute attaches to a UDP listener by sectionName and port", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "udp-route-by-section-and-port", Namespace: ns}
			kubernetes.UDPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN,
				[]v1.RouteParentStatus{acceptedParent()}, false)

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
				kubernetes.NewGatewayRef(gwNN, "dns"))
			if err != nil {
				t.Fatalf("error getting gateway address: %v", err)
			}
			udp.ExpectEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
		})

		t.Run("UDPRoute with neither sectionName nor port attaches to every UDP listener on the Gateway", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "udp-route-attach-all", Namespace: ns}
			kubernetes.UDPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, routeNN,
				[]v1.RouteParentStatus{acceptedParent()}, false)

			// Both UDP listeners should forward to the configured backend.
			for _, listener := range []string{"dns", "game"} {
				gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
					kubernetes.NewGatewayRef(gwNN, listener))
				if err != nil {
					t.Fatalf("error getting gateway address for listener %q: %v", listener, err)
				}
				udp.ExpectEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
			}
		})
	},
}
