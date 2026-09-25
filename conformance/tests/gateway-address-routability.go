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
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutability)
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityClusterClaimsBase)
}

var GatewayAddressRoutabilityClusterClaimsBase = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterClaimsBase",
	Description: "A GatewayClass claiming Cluster address routability should also claim base address routability.",
	Features: []features.FeatureName{
		features.SupportGatewayAddressRoutabilityCluster,
	},
	Parallel: true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		require.True(t, gatewayClassClaimsFeature(t, s, features.SupportGatewayAddressRoutability), "GatewayClass claims %q without claiming %q", features.SupportGatewayAddressRoutabilityCluster, features.SupportGatewayAddressRoutability)
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
		clusterClaimed := gatewayClassClaimsFeature(t, s, features.SupportGatewayAddressRoutabilityCluster)
		var serviceCIDRs []netip.Prefix
		for _, tc := range []struct {
			name             string
			exactRoutability string
		}{
			{name: "gateway-address-routability", exactRoutability: "testing.gateway.networking.k8s.io/sentinel"},
			{name: "gateway-address-routability-empty"},
			{name: "gateway-address-routability-default"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				gwNN := types.NamespacedName{Name: tc.name, Namespace: suite.InfrastructureNamespace}
				gatewayMustHaveAddressesAssigned(t, s, gwNN, metav1.ConditionTrue, string(v1.GatewayReasonAddressesAssigned))

				ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
				defer cancel()
				var gateway v1.Gateway
				waitErr := wait.PollUntilContextTimeout(ctx, s.TimeoutConfig.DefaultPollInterval, s.TimeoutConfig.GatewayMustHaveAddress, true, func(ctx context.Context) (bool, error) {
					if err := s.Client.Get(ctx, gwNN, &gateway); err != nil {
						return false, err
					}
					if len(gateway.Status.Addresses) == 0 {
						return false, nil
					}
					for _, address := range gateway.Status.Addresses {
						if address.Type == nil || *address.Type != v1.IPAddressType || address.Routability == nil || address.Value == "" {
							return false, nil
						}
					}
					return tc.exactRoutability == "" || len(gateway.Status.Addresses) == 1 && string(*gateway.Status.Addresses[0].Routability) == tc.exactRoutability, nil
				})
				require.NoError(t, waitErr, "waiting for Gateway address routability")
				if tc.exactRoutability != "" {
					require.Len(t, gateway.Status.Addresses, 1)
					require.Equal(t, tc.exactRoutability, string(*gateway.Status.Addresses[0].Routability))
					return
				}

				for _, address := range gateway.Status.Addresses {
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
	},
}

func gatewayClassClaimsFeature(t *testing.T, s *suite.ConformanceTestSuite, feature features.FeatureName) bool {
	t.Helper()
	var gatewayClass v1.GatewayClass
	require.NoError(t, s.Client.Get(context.Background(), types.NamespacedName{Name: s.GatewayClassName}, &gatewayClass))
	for _, supportedFeature := range gatewayClass.Status.SupportedFeatures {
		if supportedFeature.Name == v1.FeatureName(feature) {
			return true
		}
	}
	return false
}

func gatewayServiceCIDRs(t *testing.T, s *suite.ConformanceTestSuite) []netip.Prefix {
	t.Helper()
	cidrStrings := s.ServiceCIDRs
	if len(cidrStrings) == 0 {
		serviceCIDRs, err := s.Clientset.NetworkingV1().ServiceCIDRs().List(context.Background(), metav1.ListOptions{})
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
