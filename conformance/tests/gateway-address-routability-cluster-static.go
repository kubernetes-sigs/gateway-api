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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityClusterStatic)
}

var GatewayAddressRoutabilityClusterStatic = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterStatic",
	Description: "A Gateway requesting static addresses with Cluster address routability should be assigned only the addresses in the ServiceCIDR.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayStaticAddresses,
		features.SupportGatewayAddressRoutability,
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Manifests: []string{
		"tests/gateway-address-routability-cluster-static.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		require.NotEmpty(t, s.UsableClusterNetworkAddresses, "expected configured usableClusterNetworkAddresses")
		require.NotEmpty(t, s.UnusableClusterNetworkAddresses, "expected configured unusableClusterNetworkAddresses")
		serviceCIDRs := gatewayServiceCIDRs(t, s)
		for _, address := range s.UsableClusterNetworkAddresses {
			if address.Type != nil {
				require.Equal(t, v1.IPAddressType, *address.Type, "configured usable Cluster address must be an IP address")
			}
			require.True(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "configured usable Cluster address %q must be in a ServiceCIDR", address.Value)
		}
		for _, address := range s.UnusableClusterNetworkAddresses {
			if address.Type != nil {
				require.Equal(t, v1.IPAddressType, *address.Type, "configured unusable Cluster address must be an IP address")
			}
			require.False(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "configured unusable Cluster address %q must be outside every ServiceCIDR", address.Value)
		}

		partialGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static-partial", Namespace: suite.InfrastructureNamespace}
		var gateway v1.Gateway
		require.Eventually(t, func() bool {
			if s.Client.Get(t.Context(), partialGateway, &gateway) != nil {
				return false
			}
			condition := meta.FindStatusCondition(gateway.Status.Conditions, string(v1.GatewayConditionAddressesAssigned))
			return condition != nil && condition.ObservedGeneration >= gateway.Generation
		}, s.TimeoutConfig.GatewayMustHaveCondition, s.TimeoutConfig.DefaultPollInterval, "waiting for AddressesAssigned condition")
		assignment := meta.FindStatusCondition(gateway.Status.Conditions, string(v1.GatewayConditionAddressesAssigned))
		require.Equal(t, metav1.ConditionFalse, assignment.Status)
		switch assignment.Reason {
		case string(v1.GatewayReasonAddressesPartiallyAssigned):
			require.Eventually(t, func() bool {
				return s.Client.Get(t.Context(), partialGateway, &gateway) == nil &&
					addressesMatchEntries(gateway.Status.Addresses, func(address v1.GatewayStatusAddress) bool {
						return address.Routability != nil
					})
			}, s.TimeoutConfig.GatewayMustHaveAddress, s.TimeoutConfig.DefaultPollInterval, "only the satisfiable empty request should be reported, with one address or one per IP family")
			for _, address := range gateway.Status.Addresses {
				switch *address.Routability {
				case v1.GatewayAddressRoutabilityDefault:
				case v1.GatewayAddressRoutabilityCluster:
					require.True(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "reported Cluster address %q is not in a ServiceCIDR", address.Value)
				default:
					require.Failf(t, "unexpected status routability", "Gateway reported %q", *address.Routability)
				}
			}
		case string(v1.GatewayReasonAddressesNotAssigned):
			require.Empty(t, gateway.Status.Addresses, "a rejected partial assignment must not report addresses")
			gatewayMustNotBeProgrammedForAddresses(t, s, partialGateway)
		default:
			require.Failf(t, "unexpected AddressesAssigned reason", "got %q", assignment.Reason)
		}

		unassignedGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static-unassigned", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, unassignedGateway, metav1.ConditionFalse, string(v1.GatewayReasonAddressesNotAssigned))
		gatewayMustNotBeProgrammedForAddresses(t, s, unassignedGateway)
		gateway = v1.Gateway{}
		require.NoError(t, s.Client.Get(t.Context(), unassignedGateway, &gateway))
		require.Empty(t, gateway.Status.Addresses)

		assignedGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, assignedGateway, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))
		var accepts []func(v1.GatewayStatusAddress) bool
		for _, requested := range s.UsableClusterNetworkAddresses {
			accepts = append(accepts, func(address v1.GatewayStatusAddress) bool {
				return address.Value == requested.Value && ptr.Equal(address.Routability, ptr.To(v1.GatewayAddressRoutabilityCluster))
			})
		}
		require.Eventually(t, func() bool {
			return s.Client.Get(t.Context(), assignedGateway, &gateway) == nil && addressesMatchEntries(gateway.Status.Addresses, accepts...)
		}, s.TimeoutConfig.GatewayMustHaveAddress, s.TimeoutConfig.DefaultPollInterval, "status must report every requested usable address with Cluster routability")
	},
}
