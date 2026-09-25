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

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityCluster)
}

var GatewayAddressRoutabilityCluster = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityCluster",
	Description: "A Gateway requesting Cluster address routability should report a ClusterIP address.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayAddressRoutability,
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Manifests: []string{
		"tests/gateway-address-routability-cluster.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "gateway-address-routability-cluster", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, gwNN, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))

		var gateway v1.Gateway
		require.Eventually(t, func() bool {
			return s.Client.Get(t.Context(), gwNN, &gateway) == nil &&
				addressesMatchEntries(gateway.Status.Addresses, func(address v1.GatewayStatusAddress) bool {
					return address.Routability != nil && *address.Routability == v1.GatewayAddressRoutabilityCluster
				})
		}, s.TimeoutConfig.GatewayMustHaveAddress, s.TimeoutConfig.DefaultPollInterval, "waiting for Cluster Gateway address routability")

		serviceCIDRs := gatewayServiceCIDRs(t, s)
		for _, address := range gateway.Status.Addresses {
			require.True(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "reported Cluster address %q is not in a ServiceCIDR", address.Value)
		}
	},
}
