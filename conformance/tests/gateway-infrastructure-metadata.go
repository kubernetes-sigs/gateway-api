/*
Copyright 2024 The Kubernetes Authors.

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayInfraMetadata)
}

var GatewayInfraMetadata = suite.ConformanceTest{
	ShortName:   "GatewayInfraMetadata",
	Description: "A Gateway should accept infrastructure metadata",
	Features: []features.SupportedFeature{
		features.SupportGateway,
		features.SupportGatewayInfrastructureMetadata,
	},
	Manifests: []string{
		"tests/gateway-infrastructure-metadata.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{
			Name:      "gateway-infra-metadata",
			Namespace: "gateway-conformance-infra",
		}

		conditions := []metav1.Condition{
			{
				Type:   string(gatewayv1.GatewayConditionAccepted),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
			{
				Type:   string(gatewayv1.GatewayConditionProgrammed),
				Status: metav1.ConditionTrue,
				Reason: "", // any reason
			},
		}

		kubernetes.GatewayMustHaveLatestConditions(t, suite.Client, suite.TimeoutConfig, gwNN)
		for _, condition := range conditions {
			kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, condition)
		}
	},
}
