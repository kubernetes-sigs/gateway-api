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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteTerminateSimpleSameNamespace)
}

var TLSRouteTerminateSimpleSameNamespace = suite.ConformanceTest{
	ShortName:   "TLSRouteTerminateSimpleSameNamespace",
	Description: "A single TLSRoute in the gateway-conformance-infra namespace attaches to a Gateway using Terminate mode in the same namespace",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTLSRoute,
		features.SupportTLSRouteModeTerminate,
	},
	Manifests: []string{"tests/tlsroute-terminate-simple-same-namespace.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "gateway-conformance-mqtt-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-tlsroute-terminate", Namespace: ns}
		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr, hostnames := kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		if len(hostnames) != 1 {
			t.Fatalf("unexpected error in test configuration, found %d hostnames", len(hostnames))
		}
		serverStr := string(hostnames[0])

		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, caCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		caString, ok := caConfigMap["ca.crt"]
		if !ok {
			t.Fatalf("ca.crt not found in configmap: %s/%s", caCertNN.Namespace, caCertNN.Name)
		}

		t.Run("Simple MQTT TLS request matching TLSRoute should reach mqtt-backend", func(t *testing.T) {
			t.Logf("Establishing MQTT connection to host %s via %s", serverStr, gwAddr)

			certpool := x509.NewCertPool()
			if !certpool.AppendCertsFromPEM([]byte(caString)) {
				t.Fatal("Failed to append CA certificate")
			}

			opts := mqtt.NewClientOptions()
			opts.AddBroker(fmt.Sprintf("tls://%s", gwAddr))
			opts.SetTLSConfig(&tls.Config{
				RootCAs:    certpool,
				ServerName: serverStr,
				MinVersion: tls.VersionTLS13,
			})
			opts.SetConnectRetry(true)

			var wg sync.WaitGroup
			wg.Add(1)

			topic := "test/tlsroute-terminate"
			opts.OnConnect = func(c mqtt.Client) {
				t.Log("Connected to MQTT broker")

				if token := c.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
					t.Logf("Received message: %s\n", string(msg.Payload()))
					wg.Done()
				}); token.Wait() && token.Error() != nil {
					t.Fatalf("Failed to subscribe: %v", token.Error())
				}

				t.Log("Subscribed, publishing test message...")
				if token := c.Publish(topic, 0, false, "Hello TLSRoute Terminate MQTT!"); token.Wait() && token.Error() != nil {
					t.Fatalf("Failed to publish: %v", token.Error())
				}
			}

			client := mqtt.NewClient(opts)
			if token := client.Connect(); token.Wait() && token.Error() != nil {
				t.Fatalf("Connection failed: %v", token.Error())
			}

			waitCh := make(chan struct{})
			go func() {
				wg.Wait()
				close(waitCh)
			}()

			select {
			case <-waitCh:
				t.Log("Round-trip test succeeded")
			case <-time.After(5 * time.Second):
				t.Fatal("Timed out waiting for message")
			}
		})
	},
}
