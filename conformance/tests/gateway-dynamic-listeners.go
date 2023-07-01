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
	"strings"
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
	ConformanceTests = append(ConformanceTests, GatewayListenerHTTPRouteDynamicPorts)
}

var GatewayListenerHTTPRouteDynamicPorts = suite.ConformanceTest{
	ShortName: "GatewayListenerHTTPRouteDynamicPorts",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportHTTPRoute,
		suite.SupportGatewayListenerHTTPRouteDynamicPorts,
	},
	Description: "A Gateway and an HTTPRoute in the gateway-conformance-infra namespace should support adding and removing listeners with arbitrary ports",
	Manifests:   []string{"tests/gateway-dynamic-listeners.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		// Ephemeral port range
		const (
			portCount    = 10
			tlsPortCount = 5
		)
		var (
			portStart = s.ListenerConfig.DynamicPortRange.Start
			portEnd   = s.ListenerConfig.DynamicPortRange.End

			gwNN       = types.NamespacedName{Name: "gateway-dynamic-listener", Namespace: "gateway-conformance-infra"}
			namespaces = []string{"gateway-conformance-infra"}
			certNN     = types.NamespacedName{Name: "tls-wildcard-hostname", Namespace: gwNN.Namespace}
			ports      = sets.New[int]()
			listeners  = make([]v1beta1.Listener, 0, portCount)
			same       = v1beta1.NamespacesFromSame

			expectedConditions = []metav1.Condition{{
				Type:   string(v1beta1.ListenerConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			}, {
				Type:   string(v1beta1.ListenerConditionProgrammed),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			}}

			expectedListeners = []v1beta1.ListenerStatus{{
				Name: "http",
				SupportedKinds: []v1beta1.RouteGroupKind{{
					Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
					Kind:  v1beta1.Kind("HTTPRoute"),
				}},
				Conditions:     expectedConditions,
				AttachedRoutes: 1,
			}}
		)
		if portEnd < portStart {
			t.Fatal("DynamicPortRange.Start must be less than DynamicPortRange.End")
		}

		if portEnd-portStart < portCount {
			t.Fatal("DynamicPortRange input requires at least 10 ports")
		}

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
				Conditions:     expectedConditions,
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
				Conditions: append([]metav1.Condition{{
					Type:   string(v1beta1.ListenerConditionResolvedRefs),
					Status: metav1.ConditionTrue,
					Reason: "", // any reason
				}}, expectedConditions...),
				AttachedRoutes: 1,
			})
		}

		gwAddr, err := kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, gwNN)
		require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")
		host, _, err := net.SplitHostPort(gwAddr)
		require.NoErrorf(t, err, "unable to split gateway address %q", gwAddr)

		kubernetes.NamespacesMustBeReady(t, s.Client, s.TimeoutConfig, namespaces)
		certBytes, keyBytes, err := GetTLSSecret(s.Client, certNN)
		require.NoErrorf(t, err, "error getting certificate: %v", err)

		sendRequestToEachListener := func(t *testing.T, expectedResponse http.ExpectedResponse, listeners []v1beta1.Listener) {
			for _, listener := range listeners {
				addr := net.JoinHostPort(host, strconv.Itoa(int(listener.Port)))

				if listener.TLS != nil {
					listenerHost := string(*listener.Hostname)
					expectedResponse.Request.Host = listenerHost
					tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, addr, certBytes, keyBytes, listenerHost, expectedResponse)
				} else {
					http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, addr, expectedResponse)
				}
			}
		}

		original := &v1beta1.Gateway{}

		t.Log("should be able to add multiple HTTP listeners with dynamic ports that then become available for routing traffic")
		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		err = s.Client.Get(ctx, gwNN, original)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		// verify that the implementation is tracking the most recent resource changes
		kubernetes.GatewayMustHaveLatestConditions(t, s.TimeoutConfig, original)
		mutate := original.DeepCopy()
		mutate.Spec.Listeners = append(mutate.Spec.Listeners, listeners...)

		err = s.Client.Patch(ctx, mutate, client.MergeFrom(original))
		require.NoErrorf(t, err, "error patching the Gateway: %v", err)
		kubernetes.GatewayStatusMustHaveListeners(t, s.Client, s.TimeoutConfig, gwNN, expectedListeners)

		successResponse := http.ExpectedResponse{
			Namespace: gwNN.Namespace,
			Request:   http.Request{Path: "/"},
			Response:  http.Response{StatusCode: 200},
		}

		sendRequestToEachListener(t, successResponse, mutate.Spec.Listeners)

		t.Log("should be able to remove multiple HTTP listeners with dynamic ports")
		ctx, cancel = context.WithTimeout(context.Background(), s.TimeoutConfig.GetTimeout)
		defer cancel()

		err = s.Client.Get(ctx, gwNN, mutate)
		require.NoErrorf(t, err, "error getting Gateway: %v", err)

		mutate.Spec.Listeners = original.Spec.Listeners
		err = s.Client.Update(ctx, mutate)
		require.NoErrorf(t, err, "error patching the Gateway: %v", err)

		expectedListeners = []v1beta1.ListenerStatus{{
			Name: "http",
			SupportedKinds: []v1beta1.RouteGroupKind{{
				Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
				Kind:  v1beta1.Kind("HTTPRoute"),
			}},
			Conditions:     expectedConditions,
			AttachedRoutes: 1,
		}}

		kubernetes.GatewayStatusMustHaveListeners(t, s.Client, s.TimeoutConfig, gwNN, expectedListeners)

		// Original listener should work
		sendRequestToEachListener(t, successResponse, original.Spec.Listeners)

		for _, listener := range listeners {
			addr := net.JoinHostPort(host, strconv.Itoa(int(listener.Port)))

			// Listeners that were removed should stop working
			dial := func(elapsed time.Duration) bool {
				conn, err := net.DialTimeout("tcp", addr, time.Second)
				if conn != nil {
					conn.Close()
					return false
				}
				if err != nil && strings.Contains(err.Error(), "connection refused") {
					return true
				}
				return false
			}
			http.AwaitConvergence(t, s.TimeoutConfig.RequiredConsecutiveSuccesses, s.TimeoutConfig.MaxTimeToConsistency, dial)
		}
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
