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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayListenerFilterExternalAuthReferenceGrant)
}

var GatewayListenerFilterExternalAuthReferenceGrant = confsuite.ConformanceTest{
	ShortName:   "GatewayListenerFilterExternalAuthReferenceGrant",
	Description: "A Gateway Listener in the gateway-conformance-infra namespace whose filters.requests contains an ExternalAuth filter with a backendRef to a Service in the gateway-conformance-web-backend namespace MUST have ResolvedRefs=True and Programmed=True when a ReferenceGrant granting permission to the Service exists",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportReferenceGrant,
		features.SupportGatewayListenerFilters,
		features.SupportGatewayListenerFilterExternalAuth,
	},
	Manifests: []string{"tests/gateway-listener-filter-externalauth-reference-grant.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "gateway-listener-filter-externalauth-reference-grant", Namespace: confsuite.InfrastructureNamespace}

		t.Run("Listener with cross-namespace ExternalAuth backendRef allowed by a ReferenceGrant must have ResolvedRefs=True and Programmed=True", func(t *testing.T) {
			kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "http", []metav1.Condition{
				{
					Type:   string(v1.ListenerConditionProgrammed),
					Status: metav1.ConditionTrue,
					Reason: string(v1.ListenerReasonProgrammed),
				},
				{
					Type:   string(v1.ListenerConditionResolvedRefs),
					Status: metav1.ConditionTrue,
					Reason: string(v1.ListenerReasonResolvedRefs),
				},
			})
		})
	},
}
