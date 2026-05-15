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
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	v1 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
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
			expectUDPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
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
			expectUDPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
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
				expectUDPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
			}
		})
	},
}

// expectUDPEchoResponse polls until a UDP echo round-trip against the given
// gateway address succeeds, or the timeout is exceeded.
func expectUDPEchoResponse(t *testing.T, timeout time.Duration, gwAddr string) {
	t.Helper()

	const probe = "gateway-api-conformance-udp-echo"
	tlog.Logf(t, "performing UDP echo probe on %s", gwAddr)
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "udp", gwAddr)
			if err != nil {
				tlog.Logf(t, "failed to dial UDP %s: %v", gwAddr, err)
				return false, nil
			}
			defer conn.Close()

			if err = conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
				return false, fmt.Errorf("setting UDP deadline: %w", err)
			}
			if _, err = conn.Write([]byte(probe)); err != nil {
				tlog.Logf(t, "failed to write UDP probe: %v", err)
				return false, nil
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				tlog.Logf(t, "failed to read UDP echo response: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got UDP echo response (%d bytes) from %s", n, gwAddr)
			return true, nil
		})
	if err != nil {
		t.Errorf("UDP echo probe never succeeded against %s: %v", gwAddr, err)
	}
}
