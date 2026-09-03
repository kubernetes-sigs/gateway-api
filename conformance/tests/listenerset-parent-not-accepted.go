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

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, ListenerSetParentNotAccepted)
}

var ListenerSetParentNotAccepted = confsuite.ConformanceTest{
	ShortName:   "ListenerSetParentNotAccepted",
	Description: "A ListenerSet attached to a Gateway that is not Accepted should have the Accepted condition set to False with reason ParentNotAccepted.",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportListenerSet,
	},
	Manifests: []string{
		"tests/listenerset-parent-not-accepted.yaml",
	},
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		ns := confsuite.InfrastructureNamespace
		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		gwNN := types.NamespacedName{Name: "gateway-with-listenerset-not-accepted", Namespace: ns}
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gatewayv1.GatewayConditionAccepted),
			Status: metav1.ConditionFalse,
		})
		kubernetes.GatewayMustHaveAttachedListeners(t, suite.Client, suite.TimeoutConfig, gwNN, 0)

		lsNN := types.NamespacedName{Name: "listenerset-parent-not-accepted", Namespace: ns}
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayv1.ListenerSetConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gatewayv1.ListenerSetReasonParentNotAccepted),
		})
		kubernetes.ListenerSetMustHaveCondition(t, suite.Client, suite.TimeoutConfig, lsNN, metav1.Condition{
			Type:   string(gatewayv1.ListenerSetConditionProgrammed),
			Status: metav1.ConditionFalse,
			Reason: "", // any reason
		})
	},
}
