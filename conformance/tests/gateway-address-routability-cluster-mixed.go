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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayAddressRoutabilityClusterMixed)
}

var GatewayAddressRoutabilityClusterMixed = suite.ConformanceTest{
	ShortName:   "GatewayAddressRoutabilityClusterMixed",
	Description: "A Gateway claiming Cluster address routability should report one address for each empty and Cluster request.",
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

		ctx, cancel := context.WithTimeout(context.Background(), s.TimeoutConfig.DefaultTestTimeout)
		defer cancel()
		var gateway v1.Gateway
		waitErr := wait.PollUntilContextTimeout(ctx, s.TimeoutConfig.DefaultPollInterval, s.TimeoutConfig.GatewayMustHaveAddress, true, func(ctx context.Context) (bool, error) {
			if err := s.Client.Get(ctx, gwNN, &gateway); err != nil {
				return false, err
			}
			if len(gateway.Status.Addresses) != 2 {
				return false, nil
			}
			values := map[string]struct{}{}
			clusterAddresses := 0
			for _, address := range gateway.Status.Addresses {
				if address.Type == nil || *address.Type != v1.IPAddressType || address.Value == "" || address.Routability == nil {
					return false, nil
				}
				if *address.Routability == v1.GatewayAddressRoutabilityCluster {
					clusterAddresses++
					if !addressInServiceCIDRs(address.Value, serviceCIDRs) {
						return false, nil
					}
				} else if *address.Routability != v1.GatewayAddressRoutabilityDefault {
					return false, nil
				}
				values[address.Value] = struct{}{}
			}
			return len(values) == 2 && clusterAddresses > 0, nil
		})
		require.NoError(t, waitErr, "waiting for mixed Gateway address routability")
	},
}
