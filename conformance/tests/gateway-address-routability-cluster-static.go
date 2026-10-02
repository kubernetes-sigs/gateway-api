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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityClusterStatic)
}

var GatewayAddressRoutabilityClusterStatic = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterStatic",
	Description: "A Gateway claiming Cluster address routability should honor static address requests.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayStaticAddresses,
		features.SupportGatewayAddressRoutability,
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Manifests: []string{
		"tests/gateway-address-routability-cluster-static-partial.yaml",
		"tests/gateway-address-routability-cluster-static-unassigned.yaml",
		"tests/gateway-address-routability-cluster-static.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
		defer cancel()
		require.NotEmpty(t, s.UsableNetworkAddresses, "expected configured usable addresses")
		require.NotEmpty(t, s.UnusableNetworkAddresses, "expected configured unusable addresses")
		serviceCIDRs := gatewayServiceCIDRs(t, s)
		for _, address := range s.UsableNetworkAddresses {
			if address.Type != nil {
				require.Equal(t, v1.IPAddressType, *address.Type, "configured usable address must be an IP address")
			}
			require.True(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "configured usable address %q must be in a ServiceCIDR", address.Value)
		}
		for _, address := range s.UnusableNetworkAddresses {
			if address.Type != nil {
				require.Equal(t, v1.IPAddressType, *address.Type, "configured unusable address must be an IP address")
			}
			require.False(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "configured unusable address %q must be outside every ServiceCIDR", address.Value)
		}

		partialGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static-partial", Namespace: suite.InfrastructureNamespace}
		var gateway v1.Gateway
		require.Eventually(t, func() bool {
			if s.Client.Get(ctx, partialGateway, &gateway) != nil {
				return false
			}
			condition := meta.FindStatusCondition(gateway.Status.Conditions, string(v1.GatewayConditionAddressesAssigned))
			return condition != nil && condition.ObservedGeneration >= gateway.Generation
		}, s.TimeoutConfig.GatewayMustHaveCondition, s.TimeoutConfig.DefaultPollInterval, "waiting for AddressesAssigned condition")
		assignment := meta.FindStatusCondition(gateway.Status.Conditions, string(v1.GatewayConditionAddressesAssigned))
		require.NotNil(t, assignment, "missing AddressesAssigned condition")
		require.Equal(t, metav1.ConditionFalse, assignment.Status)
		switch assignment.Reason {
		case string(v1.GatewayReasonAddressesPartiallyAssigned):
			for _, address := range s.UnusableNetworkAddresses {
				require.Contains(t, assignment.Message, address.Value, "message must identify every unsatisfied static address")
			}
			require.Len(t, gateway.Status.Addresses, 1, "only the satisfiable empty request should be reported")
			require.NotEmpty(t, gateway.Status.Addresses[0].Value)
			if gateway.Status.Addresses[0].Type != nil {
				require.Equal(t, v1.IPAddressType, *gateway.Status.Addresses[0].Type)
			}
			require.NotNil(t, gateway.Status.Addresses[0].Routability)
			switch *gateway.Status.Addresses[0].Routability {
			case v1.GatewayAddressRoutabilityDefault:
			case v1.GatewayAddressRoutabilityCluster:
				require.True(t, addressInServiceCIDRs(gateway.Status.Addresses[0].Value, serviceCIDRs), "reported Cluster address %q is not in a ServiceCIDR", gateway.Status.Addresses[0].Value)
			default:
				require.Failf(t, "unexpected status routability", "Gateway reported %q", *gateway.Status.Addresses[0].Routability)
			}
		case string(v1.GatewayReasonAddressesNotAssigned):
			require.Empty(t, gateway.Status.Addresses, "a rejected partial assignment must not report addresses")
			kubernetes.GatewayMustHaveCondition(t, s.Client, s.TimeoutConfig, partialGateway, metav1.Condition{
				Type:   string(v1.GatewayConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: string(v1.GatewayReasonAddressNotAssigned),
			})
		default:
			require.Failf(t, "unexpected AddressesAssigned reason", "got %q", assignment.Reason)
		}

		unassignedGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static-unassigned", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, unassignedGateway, metav1.ConditionFalse, string(v1.GatewayReasonAddressesNotAssigned))
		kubernetes.GatewayMustHaveCondition(t, s.Client, s.TimeoutConfig, unassignedGateway, metav1.Condition{
			Type:   string(v1.GatewayConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(v1.GatewayReasonAddressNotAssigned),
		})
		gateway = v1.Gateway{}
		require.NoError(t, s.Client.Get(ctx, unassignedGateway, &gateway))
		require.Empty(t, gateway.Status.Addresses)

		assignedGateway := types.NamespacedName{Name: "gateway-address-routability-cluster-static", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, assignedGateway, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))
		gateway = v1.Gateway{}
		require.NoError(t, s.Client.Get(ctx, assignedGateway, &gateway))
		var reportedValues []string
		for _, address := range gateway.Status.Addresses {
			reportedValues = append(reportedValues, address.Value)
			if address.Type != nil {
				require.Equal(t, v1.IPAddressType, *address.Type)
			}
			require.NotNil(t, address.Routability)
			require.Equal(t, v1.GatewayAddressRoutabilityCluster, *address.Routability)
		}
		var expectedValues []string
		for _, address := range s.UsableNetworkAddresses {
			expectedValues = append(expectedValues, address.Value)
		}
		require.ElementsMatch(t, expectedValues, reportedValues, "status must contain every requested usable address")
	},
}
