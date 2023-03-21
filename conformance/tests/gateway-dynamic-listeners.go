/*
Copyright 2023 The Kubernetes Authors.

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
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayListenerDynamicPorts)
}

var GatewayListenerDynamicPorts = suite.ConformanceTest{
	ShortName:   "GatewayListenerDynamicPorts",
	Features:    []suite.SupportedFeature{suite.SupportGatewayListenerDynamicPorts},
	Description: "A Gateway in the gateway-conformance-infra namespace should handle adding and removing listeners with arbitrary ports",
	Manifests:   []string{"tests/gateway-dynamic-listeners.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

		// Ephemeral port range
		const (
			portStart    = 49152
			portEnd      = 65535
			portCount    = 10
			tlsPortCount = 5
		)
		var (
			gwNN       = types.NamespacedName{Name: "gateway-dynamic-listener", Namespace: "gateway-conformance-infra"}
			namespaces = []string{"gateway-conformance-infra"}
			certNN     = types.NamespacedName{Name: "tls-wildcard-hostname", Namespace: gwNN.Namespace}
			ports      = sets.New[int]()
			listeners  = make([]v1beta1.Listener, 0, portCount)
			same       = v1beta1.NamespacesFromSame

			expectedListeners = []v1beta1.ListenerStatus{{
				Name: "http",
				SupportedKinds: []v1beta1.RouteGroupKind{{
					Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
					Kind:  v1beta1.Kind("HTTPRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1beta1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: "", //any reason
				}},
				AttachedRoutes: 1,
			}}
		)

		for i := 0; i < portCount; i++ {
			port := nextPort(portStart, portEnd, ports)

			listeners = append(listeners, v1beta1.Listener{
				Name:     v1beta1.SectionName(strconv.Itoa(port)),
				Port:     v1beta1.PortNumber(port),
				Protocol: v1beta1.HTTPProtocolType,
				AllowedRoutes: &v1beta1.AllowedRoutes{
					Namespaces: &v1beta1.RouteNamespaces{From: &same},
				},
			})

			expectedListeners = append(expectedListeners, v1beta1.ListenerStatus{
				Name: v1beta1.SectionName(strconv.Itoa(port)),
				SupportedKinds: []v1beta1.RouteGroupKind{{
					Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
					Kind:  v1beta1.Kind("HTTPRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1beta1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: "", //any reason
				}},
				AttachedRoutes: 1,
			})
		}

		for i := 0; i < tlsPortCount; i++ {
			port := nextPort(portStart, portEnd, ports)
			hostname := v1beta1.Hostname(fmt.Sprintf("%v.example.com", port))

			listeners = append(listeners, v1beta1.Listener{
				Name:     v1beta1.SectionName(strconv.Itoa(port)),
				Port:     v1beta1.PortNumber(port),
				Hostname: &hostname,
				Protocol: v1beta1.HTTPSProtocolType,
				AllowedRoutes: &v1beta1.AllowedRoutes{
					Namespaces: &v1beta1.RouteNamespaces{From: &same},
				},
				TLS: &v1beta1.GatewayTLSConfig{
					CertificateRefs: []v1beta1.SecretObjectReference{{
						Name: v1beta1.ObjectName(certNN.Name),
					}},
				},
			})

			expectedListeners = append(expectedListeners, v1beta1.ListenerStatus{
				Name: v1beta1.SectionName(strconv.Itoa(port)),
				SupportedKinds: []v1beta1.RouteGroupKind{{
					Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
					Kind:  v1beta1.Kind("HTTPRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1beta1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: "", //any reason
				}},
				AttachedRoutes: 1,
			})
		}

		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)
		certBytes, keyBytes, err := GetTLSSecret(s.Client, certNN)
		require.NoErrorf(t, err, "error getting certificate: %v", err)

		t.Run("should be able to add multiple HTTP listeners with dynamic ports that then becomes available for routing traffic", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
			defer cancel()

			original := &v1beta1.Gateway{}
			err := s.Client.Get(ctx, gwNN, original)
			require.NoErrorf(t, err, "error getting Gateway: %v", err)

			// verify that the implementation is tracking the most recent resource changes
			kubernetes.GatewayMustHaveLatestConditions(t, s.TimeoutConfig, original)

			mutate := original.DeepCopy()

			mutate.Spec.Listeners = append(mutate.Spec.Listeners, listeners...)

			err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
			require.NoErrorf(t, err, "error patching the Gateway: %v", err)

			kubernetes.GatewayStatusMustHaveListeners(t, s.Client, s.TimeoutConfig, gwNN, expectedListeners)

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, gwNN)
			require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

			for _, listener := range mutate.Spec.Listeners {
				timeoutConfig := s.TimeoutConfig
				if timeoutConfig.MaxTimeToConsistency < 2*time.Minute {
					timeoutConfig.MaxTimeToConsistency = 2 * time.Minute
				}

				host, _, err := net.SplitHostPort(gwAddr)
				require.NoErrorf(t, err, "unable to split gateway address %q", gwAddr)

				expectedResponse := http.ExpectedResponse{
					Namespace: gwNN.Namespace,
					Request:   http.Request{Path: "/"},
					Response:  http.Response{StatusCode: 200},
				}

				addr := net.JoinHostPort(host, strconv.Itoa(int(listener.Port)))

				if listener.TLS != nil {
					host := string(*listener.Hostname)
					expectedResponse.Request.Host = host
					tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, timeoutConfig, addr, certBytes, keyBytes, host, expectedResponse)
				} else {
					http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, timeoutConfig, addr, expectedResponse)
				}

			}
		})
	},
}

func nextPort(start, end int, ports sets.Set[int]) int {
	port := start + rand.Intn(end-start) //nolint:gosec
	// We want a unique port
	for ports.Has(port) {
		port = start + rand.Intn(end-start) //nolint:gosec
	}
	ports.Insert(port)
	return port
}
