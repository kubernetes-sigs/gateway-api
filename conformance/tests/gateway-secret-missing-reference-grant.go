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
	ConformanceTests = append(ConformanceTests, GatewaySecretMissingReferenceGrant)
}

var GatewaySecretMissingReferenceGrant = suite.ConformanceTest{
	ShortName:   "GatewaySecretMissingReferenceGrant",
	Description: "A Gateway in the gateway-conformance-infra namespace should fail to become programmed if the Gateway has a certificateRef for a Secret in the gateway-conformance-web-backend namespace and a ReferenceGrant granting permission to the Secret does not exist",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportReferenceGrant,
	},
	Manifests: []string{"tests/gateway-secret-missing-reference-grant.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ns := suite.InfrastructureNamespace
		gwNN := types.NamespacedName{Name: "gateway-secret-missing-reference-grant", Namespace: ns}
		controlRouteNN := types.NamespacedName{Name: "gateway-secret-missing-reference-grant-control", Namespace: ns}
		certNN := types.NamespacedName{Name: "certificate", Namespace: suite.WebBackendNamespace}

		t.Run("Gateway listener should have a false ResolvedRefs condition with reason RefNotPermitted", func(t *testing.T) {
			kubernetes.GatewayListenerMustHaveConditions(t, s.Client, s.TimeoutConfig, gwNN, "https", []metav1.Condition{{
				Type:   string(v1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.ListenerReasonRefNotPermitted),
			}})

			gwAddr, err := kubernetes.WaitForGatewayAddress(t, s.Client, s.TimeoutConfig, kubernetes.NewGatewayRef(gwNN, "http"))
			require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

			t.Run("Control route is fully programmed", func(t *testing.T) {
				kubernetes.HTTPRouteMustHaveRouteAcceptedConditionsTrue(t, s.Client, s.TimeoutConfig, controlRouteNN, gwNN)
				kubernetes.HTTPRouteMustHaveResolvedRefsConditionsTrue(t, s.Client, s.TimeoutConfig, controlRouteNN, gwNN)
				http.MakeRequestAndExpectEventuallyConsistentResponse(t, s.RoundTripper, s.TimeoutConfig, gwAddr, http.ExpectedResponse{
					Request:   http.Request{Host: "missing-reference-grant.example.com", Path: "/"},
					Response:  http.Response{StatusCode: 200},
					Backend:   suite.InfraBackendServiceNameV1,
					Namespace: ns,
				})
			})

			t.Run("Rejected listener does not accept connections for a hostname matched by its wildcard", func(t *testing.T) {
				certificate, _, err := kubernetes.GetTLSSecret(s.Client, certNN)
				require.NoErrorf(t, err, "failed to get referenced TLS certificate")
				gwIP, _, err := net.SplitHostPort(gwAddr)
				require.NoErrorf(t, err, "failed to split Gateway address %q", gwAddr)
				tls.MakeTLSConnectionAndExpectEventuallyConsistentFailure(t, s.TimeoutConfig, net.JoinHostPort(gwIP, "443"), certificate, "test")
			})
		})
	},
}
