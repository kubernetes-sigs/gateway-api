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
	ConformanceTests = append(ConformanceTests, GatewayListenerFilterExternalAuthInvalidRef)
}

var GatewayListenerFilterExternalAuthInvalidRef = confsuite.ConformanceTest{
	ShortName:   "GatewayListenerFilterExternalAuthInvalidRef",
	Description: "A Gateway Listener whose filters.requests contains an ExternalAuth filter with an unresolvable backendRef MUST have ResolvedRefs=False with Reason=InvalidListenerFilterRef, and with Reason=RefNotPermitted when a cross-namespace backendRef is not allowed by a ReferenceGrant",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportReferenceGrant,
		features.SupportGatewayListenerFilters,
		features.SupportGatewayListenerFilterExternalAuth,
	},
	Manifests: []string{"tests/gateway-listener-filter-externalauth-invalid-ref.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		gwNN := types.NamespacedName{Name: "gateway-listener-filter-externalauth-invalid-ref", Namespace: confsuite.InfrastructureNamespace}

		t.Run("Listener with ExternalAuth backendRef to a nonexistent Service must have ResolvedRefs=False with Reason=InvalidListenerFilterRef", func(t *testing.T) {
			kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "http-nonexistent-backend", []metav1.Condition{{
				Type:   string(v1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.ListenerReasonInvalidListenerFilterRef),
			}})
		})

		t.Run("Listener with cross-namespace ExternalAuth backendRef and no ReferenceGrant must have ResolvedRefs=False with Reason=RefNotPermitted", func(t *testing.T) {
			kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "http-missing-reference-grant", []metav1.Condition{{
				Type:   string(v1.ListenerConditionResolvedRefs),
				Status: metav1.ConditionFalse,
				Reason: string(v1.ListenerReasonRefNotPermitted),
			}})
		})
	},
}
