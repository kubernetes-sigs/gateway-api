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
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayWithAddresses)
}

var GatewayWithAddresses = suite.ConformanceTest{
	ShortName:   "GatewayWithAddresses",
	Description: "A Gateway with addresses in the gateway-conformance-infra namespace should be populated in its status.",
	Features: []suite.SupportedFeature{
		suite.SupportGateway,
		suite.SupportGatewayWithAddresses,
	},
	Manifests: []string{"tests/gateway-with-addresses.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		t.Run("Gateway should have one valid address in status", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gateway-with-one-address", Namespace: "gateway-conformance-infra"}
			addresses := []v1beta1.GatewayAddress{{
				Type:  kubernetes.PtrTo(v1beta1.IPAddressType),
				Value: "1.2.3.4",
			}}

			kubernetes.GatewayStatusMustHaveAddresses(t, s.Client, s.TimeoutConfig, gwNN, addresses)
		})

		t.Run("Gateway should have two valid addresses in status", func(t *testing.T) {
			gwNN := types.NamespacedName{Name: "gateway-with-two-addresses", Namespace: "gateway-conformance-infra"}
			addresses := []v1beta1.GatewayAddress{{
				Type:  kubernetes.PtrTo(v1beta1.IPAddressType),
				Value: "1.2.3.4",
			}, {
				Type:  kubernetes.PtrTo(v1beta1.HostnameAddressType),
				Value: "foo.bar",
			}}

			kubernetes.GatewayStatusMustHaveAddresses(t, s.Client, s.TimeoutConfig, gwNN, addresses)
		})
	},
}
