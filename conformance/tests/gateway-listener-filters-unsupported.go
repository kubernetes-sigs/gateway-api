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
	ConformanceTests = append(ConformanceTests, GatewayListenerFiltersUnsupported)
}

var GatewayListenerFiltersUnsupported = confsuite.ConformanceTest{
	ShortName:   "GatewayListenerFiltersUnsupported",
	Description: "A Gateway Listener whose filters.requests contains a filter type the implementation does not support MUST have Accepted=False with Reason=UnsupportedValue",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportGatewayListenerFilters,
	},
	Manifests: []string{"tests/gateway-listener-filters-unsupported.yaml"},
	Parallel:  true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		testCases := []struct {
			filterType v1.ListenerFilterType
			feature    features.FeatureName
			gwName     string
		}{
			{
				filterType: v1.ListenerFilterExternalAuth,
				feature:    features.SupportGatewayListenerFilterExternalAuth,
				gwName:     "gateway-listener-filter-externalauth-unsupported",
			},
			{
				filterType: v1.ListenerFilterExtensionRef,
				feature:    features.SupportGatewayListenerFilterExtensionRef,
				gwName:     "gateway-listener-filter-extensionref-unsupported",
			},
		}

		for _, tc := range testCases {
			t.Run("Listener with unsupported "+string(tc.filterType)+" filter must have Accepted=False with Reason=UnsupportedValue", func(t *testing.T) {
				if suite.SupportedFeatures.Has(tc.feature) {
					t.Skipf("%s is supported", tc.feature)
				}
				t.Parallel()

				gwNN := types.NamespacedName{Name: tc.gwName, Namespace: confsuite.InfrastructureNamespace}
				kubernetes.GatewayListenerMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gwNN, "http", []metav1.Condition{{
					Type:   string(v1.ListenerConditionAccepted),
					Status: metav1.ConditionFalse,
					Reason: string(v1.ListenerReasonUnsupportedValue),
				}})
			})
		}
	},
}
