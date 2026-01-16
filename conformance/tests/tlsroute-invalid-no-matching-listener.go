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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, TLSRouteInvalidNoMatchingListener)
}

var TLSRouteInvalidNoMatchingListener = suite.ConformanceTest{
	ShortName:   "TLSRouteInvalidNoMatchingListener",
	Description: "A TLSRoute should set Accepted=False with Reason=NoMatchingParent when attaching to a Gateway with no TLS listener",
	Features: []features.FeatureName{
		features.SupportGateway,
		features.SupportHTTPRoute,
		features.SupportTLSRoute,
	},
	Manifests: []string{"tests/tlsroute-invalid-no-matching-listener.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		routeNN := types.NamespacedName{Name: "tlsroute-no-matching-listener", Namespace: "gateway-conformance-infra"}
		gwNN := types.NamespacedName{Name: "gateway-http-only", Namespace: "gateway-conformance-infra"}

		t.Run("TLSRoute has Accepted Condition with status False and Reason NoMatchingParent", func(t *testing.T) {
			acceptedCond := metav1.Condition{
				Type:   string(v1.RouteConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(v1.RouteReasonNoMatchingParent),
			}
			kubernetes.TLSRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, routeNN, gwNN, acceptedCond)
		})

		t.Run("Route should not have Parents accepted in status", func(t *testing.T) {
			kubernetes.TLSRouteMustHaveNoAcceptedParents(t, suite.Client, suite.TimeoutConfig, routeNN)
		})

		t.Run("Gateway should have 0 Routes attached", func(t *testing.T) {
			kubernetes.GatewayMustHaveZeroRoutes(t, suite.Client, suite.TimeoutConfig, gwNN)
		})
	},
}
