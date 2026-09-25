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
	"net/netip"
	"slices"
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
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutability, GatewayAddressRoutabilityClusterClaimsBase)
}

var GatewayAddressRoutabilityClusterClaimsBase = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterClaimsBase",
	Description: "A GatewayClass claiming Cluster address routability should also claim base address routability.",
	Features: []features.FeatureName{
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		require.True(t, s.SupportedFeatures.Has(features.SupportGatewayAddressRoutability), "%q is supported without %q", features.SupportGatewayAddressRoutabilityCluster, features.SupportGatewayAddressRoutability)
	},
}

var GatewayAddressRoutability = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutability",
	Description: "A Gateway should report the requested address routability.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayAddressRoutability,
	},
	Manifests: []string{
		"tests/gateway-address-routability.yaml",
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		clusterClaimed := s.SupportedFeatures.Has(features.SupportGatewayAddressRoutabilityCluster)
		var serviceCIDRs []netip.Prefix
		for _, tc := range []struct {
			name             string
			exactRoutability string
			explicitAddress  bool
		}{
			{name: "gateway-address-routability-sentinel", exactRoutability: "testing.x-k8s.io/sentinel", explicitAddress: true},
			{name: "gateway-address-routability-empty", explicitAddress: true},
			{name: "gateway-address-routability-unspecified", explicitAddress: true},
			{name: "gateway-address-routability-default"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				gwNN := types.NamespacedName{Name: tc.name, Namespace: suite.InfrastructureNamespace}
				gatewayMustHaveAddressesAssigned(t, s, gwNN, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))

				var gateway v1.Gateway
				require.Eventually(t, func() bool {
					if s.Client.Get(t.Context(), gwNN, &gateway) != nil {
						return false
					}
					if tc.explicitAddress {
						return addressesMatchEntries(gateway.Status.Addresses, func(address v1.GatewayStatusAddress) bool {
							return address.Routability != nil && (tc.exactRoutability == "" || string(*address.Routability) == tc.exactRoutability)
						})
					}
					for _, address := range gateway.Status.Addresses {
						if address.Routability == nil {
							return false
						}
					}
					return len(gateway.Status.Addresses) > 0
				}, s.TimeoutConfig.GatewayMustHaveAddress, s.TimeoutConfig.DefaultPollInterval, "waiting for Gateway address routability")
				if tc.exactRoutability != "" {
					return
				}

				for _, address := range gateway.Status.Addresses {
					if !tc.explicitAddress && address.Type != nil && *address.Type != v1.IPAddressType {
						continue // Non-IP routability is implementation-specific.
					}
					switch *address.Routability {
					case v1.GatewayAddressRoutabilityDefault:
					case v1.GatewayAddressRoutabilityCluster:
						require.True(t, clusterClaimed, "Gateway reports Cluster routability without claiming %q", features.SupportGatewayAddressRoutabilityCluster)
						if serviceCIDRs == nil {
							serviceCIDRs = gatewayServiceCIDRs(t, s)
						}
						require.True(t, addressInServiceCIDRs(address.Value, serviceCIDRs), "reported Cluster address %q is not in a ServiceCIDR", address.Value)
					default:
						require.Failf(t, "unexpected status routability", "Gateway reported unsupported routability %q", *address.Routability)
					}
				}
			})
		}

		unsupported := types.NamespacedName{Name: "gateway-address-routability-unsupported", Namespace: suite.InfrastructureNamespace}
		gatewayMustHaveAddressesAssigned(t, s, unsupported, metav1.ConditionFalse, string(v1.GatewayReasonAddressesNotAssigned))
		kubernetes.GatewayMustHaveCondition(t, s.Client, s.TimeoutConfig, unsupported, metav1.Condition{
			Type:   string(v1.GatewayConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: string(v1.GatewayReasonAddressNotAssigned),
		})
		var gateway v1.Gateway
		require.NoError(t, s.Client.Get(t.Context(), unsupported, &gateway))
		require.Empty(t, gateway.Status.Addresses, "unsupported routability must not be assigned")
	},
}

func gatewayServiceCIDRs(t *testing.T, s *suite.ConformanceTestSuite) []netip.Prefix {
	t.Helper()
	cidrStrings := s.ServiceCIDRs
	if len(cidrStrings) == 0 {
		serviceCIDRs, err := s.Clientset.NetworkingV1().ServiceCIDRs().List(t.Context(), metav1.ListOptions{})
		require.NoError(t, err, "discovering ServiceCIDRs; configure serviceCIDRs when the ServiceCIDR API is unavailable")
		for _, serviceCIDR := range serviceCIDRs.Items {
			cidrStrings = append(cidrStrings, serviceCIDR.Spec.CIDRs...)
		}
	}
	require.NotEmpty(t, cidrStrings, "no ServiceCIDRs found; configure serviceCIDRs in the conformance options")

	prefixes := make([]netip.Prefix, 0, len(cidrStrings))
	for _, cidr := range cidrStrings {
		prefix, err := netip.ParsePrefix(cidr)
		require.NoError(t, err, "parsing ServiceCIDR %q", cidr)
		prefixes = append(prefixes, prefix)
	}
	return prefixes
}

func addressInServiceCIDRs(value string, cidrs []netip.Prefix) bool {
	address, err := netip.ParseAddr(value)
	if err != nil {
		return false
	}
	for _, cidr := range cidrs {
		if cidr.Contains(address) {
			return true
		}
	}
	return false
}

func gatewayMustHaveAddressesAssigned(t *testing.T, s *suite.ConformanceTestSuite, gateway types.NamespacedName, status metav1.ConditionStatus, reason string) {
	t.Helper()
	kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gateway)
	kubernetes.GatewayMustHaveCondition(t, s.Client, s.TimeoutConfig, gateway, metav1.Condition{
		Type:   string(v1.GatewayConditionAddressesAssigned),
		Status: status,
		Reason: reason,
	})
}

// Either reason is valid when a static address cannot meet its routability.
func gatewayMustNotBeProgrammedForAddresses(t *testing.T, s *suite.ConformanceTestSuite, gwNN types.NamespacedName) {
	t.Helper()
	reasons := []string{string(v1.GatewayReasonAddressNotAssigned), string(v1.GatewayReasonAddressNotUsable)}
	require.Eventually(t, func() bool {
		var gateway v1.Gateway
		if s.Client.Get(t.Context(), gwNN, &gateway) != nil {
			return false
		}
		programmed := meta.FindStatusCondition(gateway.Status.Conditions, string(v1.GatewayConditionProgrammed))
		return programmed != nil && programmed.ObservedGeneration == gateway.Generation &&
			programmed.Status == metav1.ConditionFalse && slices.Contains(reasons, programmed.Reason)
	}, s.TimeoutConfig.GatewayMustHaveCondition, s.TimeoutConfig.DefaultPollInterval,
		"Gateway %s must have Programmed=False with reason %s or %s", gwNN, reasons[0], reasons[1])
}

// addressesMatchEntries reports whether addresses can be assigned to entries,
// accepts[i] matching entry i, with one address per entry or one per IP family.
func addressesMatchEntries(addresses []v1.GatewayStatusAddress, accepts ...func(v1.GatewayStatusAddress) bool) bool {
	if len(addresses) < len(accepts) || len(addresses) > 2*len(accepts) {
		return false
	}
	families := make([]int, len(addresses))
	seen := map[netip.Addr]bool{}
	for i, address := range addresses {
		if address.Type != nil && *address.Type != v1.IPAddressType {
			return false
		}
		ip, err := netip.ParseAddr(address.Value)
		if err != nil || seen[ip] {
			return false
		}
		seen[ip] = true
		if ip.Is4() {
			families[i] = 1
		}
	}
	used := make([][2]bool, len(accepts))
	var assign func(i int) bool
	assign = func(i int) bool {
		if i == len(addresses) {
			for _, u := range used {
				if !u[0] && !u[1] {
					return false
				}
			}
			return true
		}
		for entry, accept := range accepts {
			if used[entry][families[i]] || !accept(addresses[i]) {
				continue
			}
			used[entry][families[i]] = true
			if assign(i + 1) {
				return true
			}
			used[entry][families[i]] = false
		}
		return false
	}
	return assign(0)
}
