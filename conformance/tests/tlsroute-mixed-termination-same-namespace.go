/*
Copyright 2026 The Kubernetes Authors.

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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	tlsutils "sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteMixedTerminationSameNamespace)
}

var TLSRouteMixedTerminationSameNamespace = suite.ConformanceTest{
	ShortName:   "TLSRouteMixedTerminationSameNamespace",
	Description: "A Gateway with 2 TLS Listeners on different modes, on the same port must route the traffic correctly",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
		features.SupportTLSRouteModeTerminate,
		features.SupportTLSRouteModeMixed,
	},
	Provisional: true,
	Manifests:   []string{"tests/tlsroute-mixed-termination-same-namespace.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeTerminateNN := types.NamespacedName{Name: "gateway-conformance-mixed-terminateroute", Namespace: ns}
		routePassthroughNN := types.NamespacedName{Name: "gateway-conformance-mixed-passthroughroute", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-tlsroute-mixed", Namespace: ns}
		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}
		certNN := types.NamespacedName{Name: "tls-checks-certificate", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr, hostnamesPassthrough := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routePassthroughNN)

		listeners := []v1.ListenerStatus{
			{
				Name: v1.SectionName("tls-terminate"),
				SupportedKinds: []v1.RouteGroupKind{{
					Group: (*v1.Group)(&v1.GroupVersion.Group),
					Kind:  v1.Kind("TLSRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: string(v1.ListenerReasonAccepted),
				}},
				AttachedRoutes: 1,
			},
			{
				Name: v1.SectionName("tls-passthrough"),
				SupportedKinds: []v1.RouteGroupKind{{
					Group: (*v1.Group)(&v1.GroupVersion.Group),
					Kind:  v1.Kind("TLSRoute"),
				}},
				Conditions: []metav1.Condition{{
					Type:   string(v1.ListenerConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: string(v1.ListenerReasonAccepted),
				}},
				AttachedRoutes: 1,
			},
		}
		kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gwNN, listeners)

		if len(hostnamesPassthrough) != 1 {
			t.Fatalf("unexpected error in test configuration, found %d passthrough hostnames", len(hostnamesPassthrough))
		}
		serverStrPassthrough := string(hostnamesPassthrough[0])

		_, hostnamesTerminate := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeTerminateNN)

		if len(hostnamesTerminate) != 1 {
			t.Fatalf("unexpected error in test configuration, found %d terminate hostnames", len(hostnamesTerminate))
		}
		serverStrTerminate := string(hostnamesTerminate[0])

		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		caString, ok := caConfigMap["ca.crt"]
		if !ok {
			t.Fatalf("ca.crt not found in configmap: %s/%s", caCertNN.Namespace, caCertNN.Name)
		}

		serverCertPem, _, err := GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}

		t.Run("Simple MQTT TLS request matching TLSRoute should reach mqtt-backend", func(t *testing.T) {
			t.Parallel()

			// Using the gwAddrPassthrough as it should be the same for both listeners
			t.Logf("Establishing MQTT connection to host %s via %s", serverStrTerminate, gwAddr)

			certpool := x509.NewCertPool()
			if !certpool.AppendCertsFromPEM([]byte(caString)) {
				t.Fatal("Failed to append CA certificate")
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tls://%s", gwAddr))
			opts.SetTLSConfig(&tls.Config{
				RootCAs:    certpool,
				ServerName: serverStrTerminate,
				MinVersion: tls.VersionTLS13,
			})

			msgChan := make(chan string)

			topic := "test/tlsroute-terminate"
			message := "Hello TLSRoute Terminate MQTT!"

			c := mqtt.NewClient(opts)
			if token := c.Connect(); !token.WaitTimeout(suite.TimeoutConfig.DefaultTestTimeout) || token.Error() != nil {
				t.Fatalf("Connection failed or timed out: %v", token.Error())
			}

			if token := c.Publish(topic, 0, true, message); !token.WaitTimeout(suite.TimeoutConfig.DefaultTestTimeout) || token.Error() != nil {
				t.Fatalf("Failed to publish or timeout: %v", token.Error())
			}

			if token := c.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
				t.Logf("Received message: %s\n", string(msg.Payload()))
				msgChan <- string(msg.Payload())
			}); token.WaitTimeout(suite.TimeoutConfig.DefaultTestTimeout) && token.Error() != nil {
				t.Fatalf("Failed to subscribe or timeout: %v", token.Error())
			}

			select {
			case msg := <-msgChan:
				if msg != message {
					t.Fatalf("Expected message %s does not match the received message %s", msg, message)
				}
				t.Log("Round-trip test succeeded")
			case <-time.After(suite.TimeoutConfig.DefaultTestTimeout):
				t.Fatal("Timed out waiting for message")
			}
		})

		t.Run("Simple TLS request matching TLSRoute Passthrough should reach infra-backend", func(t *testing.T) {
			t.Parallel()
			tlsutils.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, serverCertPem, nil, nil, serverStrPassthrough,
				http.ExpectedResponse{
					Request:   http.Request{Host: serverStrPassthrough, Path: "/"},
					Backend:   "tls-backend",
					Namespace: "gateway-conformance-infra",
				})
		})
	},
}
