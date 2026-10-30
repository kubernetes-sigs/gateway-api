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
	"net"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	client "sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayClusterTrustBundle)
}

var GatewayClusterTrustBundle = confsuite.ConformanceTest{
	ShortName:   "GatewayClusterTrustBundle",
	Description: "A Gateway with client certificate validation referencing a ClusterTrustBundle should validate client certificates using the trust anchor contained in the bundle",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportGatewayFrontendClientCertificateValidation,
		features.SupportClusterTrustBundle,
	},
	Manifests: []string{"tests/gateway-clustertrustbundle.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		routeNN := types.NamespacedName{Name: "cluster-trust-bundle-test", Namespace: ns}
		gwNN := types.NamespacedName{Name: "cluster-trust-bundle", Namespace: ns}

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		// Build a ClusterTrustBundle from the CA certificate that signed the
		// client certificate and apply it. This exercises the
		// ClusterTrustBundle reference feature rather than a ConfigMap.
		bundle := kubernetes.MustCreateClusterTrustBundleFromCAConfigMap(t, suite.Client, suite.TimeoutConfig, "gateway-cluster-trust-bundle", types.NamespacedName{Name: "tls-checks-ca-certificate", Namespace: ns})
		suite.Applier.MustApplyObjectsWithCleanup(t, suite.Client, suite.TimeoutConfig, []client.Object{bundle}, suite.CleanupTestResources)

		routeNNs := []types.NamespacedName{routeNN}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gatewayv1.HTTPRoute{}, false, routeNNs...)
		kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN)

		// Get Server certificate presented by the Gateway.
		certNN := types.NamespacedName{Name: "tls-validity-checks-certificate", Namespace: ns}
		serverCertPem, _, err := kubernetes.GetTLSSecret(suite.Client, certNN)
		if err != nil {
			t.Fatalf("unexpected error finding TLS secret: %v", err)
		}
		if len(serverCertPem) == 0 {
			t.Fatal("missing required server certificate pem for the test")
		}

		// Get the client certificate signed by the CA referenced by the ClusterTrustBundle.
		validClientCertNN := types.NamespacedName{Name: "tls-checks-client-certificate", Namespace: ns}
		validClientCertPem, validClientCertKey, err := kubernetes.GetTLSSecret(suite.Client, validClientCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding client certificate: %v", err)
		}
		if len(validClientCertPem) == 0 || len(validClientCertKey) == 0 {
			t.Fatal("missing required valid client certificate and private key for the test")
		}

		// Get the client certificate signed by a different CA, which must be rejected.
		invalidClientCertNN := types.NamespacedName{Name: "tls-validity-checks-client-certificate", Namespace: ns}
		invalidClientCertPem, invalidClientCertKey, err := kubernetes.GetTLSSecret(suite.Client, invalidClientCertNN)
		if err != nil {
			t.Fatalf("unexpected error finding invalid client certificate: %v", err)
		}
		if len(invalidClientCertPem) == 0 || len(invalidClientCertKey) == 0 {
			t.Fatal("missing required invalid client certificate and private key for the test")
		}

		addr := net.JoinHostPort(gwAddr, "443")

		t.Run("TLS request with a valid client certificate signed by the ClusterTrustBundle CA should succeed", func(t *testing.T) {
			expectedSuccess := http.ExpectedResponse{
				Request:   http.Request{Host: "example.org", Path: "/"},
				Response:  http.Response{StatusCode: 200},
				Backend:   confsuite.InfraBackendServiceNameV1,
				Namespace: confsuite.InfrastructureNamespace,
			}
			tls.MakeTLSRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, serverCertPem, validClientCertPem, validClientCertKey, "example.org", expectedSuccess)
		})

		t.Run("TLS request with a client certificate signed by a different CA should fail", func(t *testing.T) {
			expectedFailure := http.ExpectedResponse{
				Request:   http.Request{Host: "example.org", Path: "/"},
				Namespace: confsuite.InfrastructureNamespace,
			}
			tls.MakeTLSRequestAndExpectEventuallyConsistentFailureResponse(t, suite.RoundTripper, suite.TimeoutConfig, addr, serverCertPem, invalidClientCertPem, invalidClientCertKey, "example.org", expectedFailure)
		})
	},
}
