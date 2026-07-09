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
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

const gatewayMaximumNameLengthPrefix = "gateway-max-name-length-"

var gatewayMaximumNameLengthName = gatewayMaximumNameLengthPrefix + strings.Repeat("a", 253-len(gatewayMaximumNameLengthPrefix))

func init() {
	ConformanceTests = append(ConformanceTests, GatewayMaximumNameLength)
}

var GatewayMaximumNameLength = suite.ConformanceTest{
	ShortName: "GatewayMaximumNameLength",
	Description: "A Gateway with a 253-character name should be reconciled by the implementation. " +
		"The test only verifies that status conditions have an updated observedGeneration and does not assert specific condition values.",
	Features: []features.FeatureName{
		features.SupportGateway,
	},
	Manifests: []string{"tests/gateway-maximum-name-length.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: gatewayMaximumNameLengthName, Namespace: suite.InfrastructureNamespace}

		kubernetes.GatewayMustHaveLatestConditions(t, s.Client, s.TimeoutConfig, gwNN)
	},
}
