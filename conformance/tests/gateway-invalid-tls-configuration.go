/*
Copyright 2022 The Kubernetes Authors.

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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tls"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayInvalidTLSConfiguration)
}

var GatewayInvalidTLSConfiguration = suite.ConformanceTest{
	ShortName:   "GatewayInvalidTLSConfiguration",
	Description: "Invalid Gateway TLS listeners must not serve traffic",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
	},
	Manifests: []string{"tests/gateway-invalid-tls-configuration.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		testCases := []struct {
			name         string
			gatewayName  string
			controlRoute string
			controlHost  string
		}{
			{
				name:         "Nonexistent secret referenced as CertificateRef in a Gateway listener",
				gatewayName:  "gateway-certificate-nonexistent-secret",
				controlRoute: "gateway-certificate-nonexistent-secret-control",
				controlHost:  "nonexistent-certificate.example.com",
			},
			{
				name:         "Unsupported group resource referenced as CertificateRef in a Gateway listener",
				gatewayName:  "gateway-certificate-unsupported-group",
				controlRoute: "gateway-certificate-unsupported-group-control",
				controlHost:  "unsupported-group-certificate.example.com",
			},
			{
				name:         "Unsupported kind resource referenced as CertificateRef in a Gateway listener",
				gatewayName:  "gateway-certificate-unsupported-kind",
				controlRoute: "gateway-certificate-unsupported-kind-control",
				controlHost:  "unsupported-kind-certificate.example.com",
			},
			{
				name:         "Malformed secret referenced as CertificateRef in a Gateway listener",
				gatewayName:  "gateway-certificate-malformed-secret",
				controlRoute: "gateway-certificate-malformed-secret-control",
				controlHost:  "malformed-certificate.example.com",
			},
		}

		ns := suite.InfrastructureNamespace
		certNN := types.NamespacedName{Name: "tls-validity-checks-certificate", Namespace: ns}
		serverCertificate, _, err := kubernetes.GetTLSSecret(s.Client, certNN)
		require.NoError(t, err, "failed to get TLS certificate")

		conditions := []metav1.Condition{{
			Type:   string(v1.ListenerConditionResolvedRefs),
			Status: metav1.ConditionFalse,
			Reason: string(v1.ListenerReasonInvalidCertificateRef),
		}}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				gwNN := types.NamespacedName{Name: tc.gatewayName, Namespace: ns}
				controlRouteNN := types.NamespacedName{Name: tc.controlRoute, Namespace: ns}

				t.Run("Certificate is recognized as invalid", func(t *testing.T) {
					kubernetes.GatewayListenerMustHaveConditions(t, s.Client, s.TimeoutConfig, gwNN, "https", conditions)
				})

				gwAddr, err := kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "http"))
				require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

				t.Run("Control route is recognized as accepted", func(t *testing.T) {
					kubernetes.HTTPRouteMustHaveRouteAcceptedConditionsTrue(t, s.Client, s.TimeoutConfig, controlRouteNN, gwNN)
				})

				t.Run("Control listener serves traffic", func(t *testing.T) {
					http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
						Request:   http.Request{Host: tc.controlHost, Path: "/"},
						Response:  http.Response{StatusCode: 200},
						Backend:   suite.InfraBackendServiceNameV1,
						Namespace: ns,
					})
				})

				t.Run("Invalid listener does not accept connections", func(t *testing.T) {
					gwIP, _, err := net.SplitHostPort(gwAddr)
					require.NoErrorf(t, err, "failed to split Gateway address %q", gwAddr)
					httpsAddr := net.JoinHostPort(gwIP, "443")
					tls.MakeTLSConnectionAndExpectEventuallyConsistentFailure(t, s.TimeoutConfig, httpsAddr, serverCertificate, "example.org")
				})
			})
		}
	},
}
