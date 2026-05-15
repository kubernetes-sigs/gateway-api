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
	ConformanceTests = append(ConformanceTests, UDPTCPListenerMerge)
}

var UDPTCPListenerMerge = confsuite.ConformanceTest{
	ShortName:   "UDPTCPListenerMerge",
	Description: "A Gateway with a UDP and a TCP listener on the same port should accept both a UDPRoute and a TCPRoute, and route UDP traffic via the UDPRoute and TCP traffic via the TCPRoute.",
	Manifests:   []string{"tests/udproute-tcproute-listener-merge.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTCPRoute,
		features.SupportUDPRoute,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "udp-tcp-listener-merge-gateway", Namespace: ns}
		udpRouteNN := types.NamespacedName{Name: "udp-route-listener-merge", Namespace: ns}
		tcpRouteNN := types.NamespacedName{Name: "tcp-route-listener-merge", Namespace: ns}

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

		t.Run("UDPRoute attaches to the UDP listener and is Accepted", func(t *testing.T) {
			kubernetes.UDPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, udpRouteNN, expectedParents, false)
		})

		t.Run("TCPRoute attaches to the TCP listener and is Accepted", func(t *testing.T) {
			kubernetes.TCPRouteMustHaveParents(t, suite.Client, suite.TimeoutConfig, tcpRouteNN, expectedParents, false)
		})

		// Both listeners share the same port, so each listener-scoped
		// WaitForGatewayAddress resolves to host:5300. UDP traffic to that
		// address must be routed via the UDPRoute and TCP traffic via the
		// TCPRoute.
		t.Run("UDP traffic on the merged port is routed via the UDPRoute", func(t *testing.T) {
			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
				kubernetes.NewGatewayRef(gwNN, "udp"))
			if err != nil {
				t.Fatalf("error getting gateway address for UDP listener: %v", err)
			}
			expectUDPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
		})

		t.Run("TCP traffic on the merged port is routed via the TCPRoute", func(t *testing.T) {
			gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig,
				kubernetes.NewGatewayRef(gwNN, "tcp"))
			if err != nil {
				t.Fatalf("error getting gateway address for TCP listener: %v", err)
			}
			expectTCPEchoResponse(t, suite.TimeoutConfig.DefaultTestTimeout, gwAddr)
		})
	},
}

// expectTCPEchoResponse polls until a TCP echo round-trip against the given
// gateway address succeeds, or the timeout is exceeded. It is paired with the
// UDP/TCP echo backend used by these tests, which replies with a JSON envelope
// after receiving a single line of input.
func expectTCPEchoResponse(t *testing.T, timeout time.Duration, gwAddr string) {
	t.Helper()

	const probe = "gateway-api-conformance-tcp-echo\n"
	tlog.Logf(t, "performing TCP echo probe on %s", gwAddr)
	err := wait.PollUntilContextTimeout(context.TODO(), time.Second, timeout, true,
		func(ctx context.Context) (bool, error) {
			var dialer net.Dialer
			conn, err := dialer.DialContext(ctx, "tcp", gwAddr)
			if err != nil {
				tlog.Logf(t, "failed to dial TCP %s: %v", gwAddr, err)
				return false, nil
			}
			defer conn.Close()

			if err = conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
				return false, fmt.Errorf("setting TCP deadline: %w", err)
			}
			if _, err = conn.Write([]byte(probe)); err != nil {
				tlog.Logf(t, "failed to write TCP probe: %v", err)
				return false, nil
			}

			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				tlog.Logf(t, "failed to read TCP echo response: %v", err)
				return false, nil
			}
			tlog.Logf(t, "got TCP echo response (%d bytes) from %s", n, gwAddr)
			return true, nil
		})
	if err != nil {
		t.Errorf("TCP echo probe never succeeded against %s: %v", gwAddr, err)
	}
}
