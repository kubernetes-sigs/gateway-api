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
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityClusterMixed)
}

var GatewayAddressRoutabilityClusterMixed = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterMixed",
	Description: "A Gateway with empty and Cluster address routability requests should report a matching address for each.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayAddressRoutability,
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Manifests: []string{
		"tests/gateway-address-routability-cluster-mixed.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "gateway-address-routability-cluster-mixed", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, gwNN, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))
		serviceCIDRs := gatewayServiceCIDRs(t, s)

		isCluster := func(address v1.GatewayStatusAddress) bool {
			return address.Routability != nil && *address.Routability == v1.GatewayAddressRoutabilityCluster &&
				addressInServiceCIDRs(address.Value, serviceCIDRs)
		}
		isEmptyOrCluster := func(address v1.GatewayStatusAddress) bool {
			return address.Routability != nil && *address.Routability == v1.GatewayAddressRoutabilityDefault || isCluster(address)
		}
		var gateway v1.Gateway
		require.Eventually(t, func() bool {
			return s.Client.Get(t.Context(), gwNN, &gateway) == nil &&
				addressesMatchEntries(gateway.Status.Addresses, isEmptyOrCluster, isCluster)
		}, s.TimeoutConfig.GatewayMustHaveAddress, s.TimeoutConfig.DefaultPollInterval, "waiting for mixed Gateway address routability")
	},
}
