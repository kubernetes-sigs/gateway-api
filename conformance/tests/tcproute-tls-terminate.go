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

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tcp"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteOnTLSTerminate)
}

var TCPRouteOnTLSTerminate = confsuite.ConformanceTest{
	ShortName:   "TCPRouteOnTLSTerminate",
	Description: "A Gateway with a TLS listener in mode Terminate should terminate TLS and forward decrypted bytes to a TCPRoute backend.",
	Manifests:   []string{"tests/tcproute-tls-terminate.yaml"},
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportTCPRoute,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "tcproute-tls-terminate", Namespace: ns}
		gwNN := types.NamespacedName{Name: "gateway-tcproute-tls-terminate", Namespace: ns}
		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}

		// The test creates an additional Gateway in the gateway-conformance-infra
		// namespace so we have to wait for it to be ready.
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwAddr := kubernetes.GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN, "tls-terminate"), routeNN)

		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS CA configmap: %v", err)
		}
		caString, ok := caConfigMap["ca.crt"]
		if !ok {
			t.Fatalf("ca.crt not found in configmap: %s/%s", caCertNN.Namespace, caCertNN.Name)
		}

		t.Run("Gateway terminates TLS and forwards plaintext bytes to the TCPRoute backend", func(t *testing.T) {
			tcp.MakeTCPRequestAndExpectEventuallyValidResponse(t, suite.TimeoutConfig, gwAddr, []byte(caString), "tls.example.com", true,
				tcp.ExpectedResponse{
					BackendIsTLS: false, // TLS is terminated on the Gateway
					Backend:      "tcp-backend",
					Namespace:    ns,
					// Hostname intentionally empty: the backend receives plaintext
					// after termination and has no SNI to assert against.
					Hostname: "",
				})
		})
	},
}
