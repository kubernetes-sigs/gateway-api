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

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tcp"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayHTTPSAndTLSPassthroughSamePort)
}

var GatewayHTTPSAndTLSPassthroughSamePort = confsuite.ConformanceTest{
	ShortName: "GatewayHTTPSAndTLSPassthroughSamePort",
	Description: "A Gateway with an HTTPS listener and a TLS Passthrough listener on the same port, whose hostnames overlap " +
		"without being equal, must accept both listeners and serve both",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportTLSRoute,
		features.SupportGatewayHTTPSAndTLSPassthroughSamePort,
	},
	// Provisional while the requirement settles: the spec leaves this
	// combination to the implementation, and the feature is what pins it.
	//
	// Both route kinds are declared because both data paths are exercised. No
	// profile carries HTTPRoute and TLSRoute together, so a pass is reported
	// under succeededProvisionalTests rather than under a profile, and a
	// failure reaches the run but no report. Dropping Provisional before the
	// test is split per profile would leave nothing in the report at all.
	Manifests:   []string{"tests/gateway-https-and-tls-passthrough-same-port.yaml"},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "gateway-https-tls-passthrough-same-port", Namespace: ns}
		httpRouteNN := types.NamespacedName{Name: "gateway-conformance-https-terminate", Namespace: ns}
		tlsRouteNN := types.NamespacedName{Name: "gateway-conformance-tls-passthrough", Namespace: ns}
		certNN := types.NamespacedName{Name: "tls-checks-certificate", Namespace: ns}
		caCertNN := types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		// The address carries the port of the first listener, which is the right
		// one for both requests here because both listeners share it.
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), httpRouteNN)
		kubernetes.GatewayAndTLSRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), tlsRouteNN)

		// Only Accepted is asserted: the Conflicted condition is optional when
		// nothing conflicts, and implementations that leave it out are still
		// conformant.
		accepted := []metav1.Condition{{
			Type:   string(v1.ListenerConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: string(v1.ListenerReasonAccepted),
		}}

		listeners := []v1.ListenerStatus{
			{
				Name: v1.SectionName("https-terminate"),
				SupportedKinds: []v1.RouteGroupKind{{
					Group: (*v1.Group)(&v1.GroupVersion.Group),
					Kind:  v1.Kind("HTTPRoute"),
				}},
				Conditions:     accepted,
				AttachedRoutes: 1,
			},
			{
				Name: v1.SectionName("tls-passthrough"),
				SupportedKinds: []v1.RouteGroupKind{{
					Group: (*v1.Group)(&v1.GroupVersion.Group),
					Kind:  v1.Kind("TLSRoute"),
				}},
				Conditions:     accepted,
				AttachedRoutes: 1,
			},
		}
		kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gwNN, listeners)

		serverCertPem, _, err := kubernetes.GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		if len(serverCertPem) == 0 {
			t.Fatalf("missing certificate pem in secret %s", certNN)
		}

		caConfigMap, err := kubernetes.GetConfigMapData(suite.Client, suite.TimeoutConfig, caCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding CA ConfigMap: %v", err)
		}
		caString, ok := caConfigMap["ca.crt"]
		if !ok {
			t.Fatalf("ca.crt not found in configmap: %s/%s", caCertNN.Namespace, caCertNN.Name)
		}

		t.Run("SNI matching the wildcard hostname is terminated by the Gateway", func(t *testing.T) {
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr,
				serverCertPem, nil, nil, "other.example.com",
				http.ExpectedResponse{
					Request:   http.Request{Host: "other.example.com", Path: "/"},
					Backend:   confsuite.InfraBackendServiceNameV1,
					Namespace: ns,
				})
		})

		t.Run("SNI matching the exact hostname is passed through to the backend", func(t *testing.T) {
			tcp.MakeTCPRequestAndExpectEventuallyValidResponse(t, suite.TimeoutConfig, gwAddr, []byte(caString), "abc.example.com", true,
				tcp.ExpectedResponse{
					BackendIsTLS: true,
					Backend:      "tcp-backend",
					Namespace:    ns,
					Hostname:     "abc.example.com",
				})
		})
	},
}
