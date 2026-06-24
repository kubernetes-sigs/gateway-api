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

	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	confsuite "sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/udp"
	"sigs.k8s.io/gateway-api/pkg/features"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteTest)
}

var UDPRouteTest = confsuite.ConformanceTest{
	ShortName:   "UDPRoute",
	Description: "Make sure UDPRoute is working",
	Manifests:   []string{"tests/udproute-simple.yaml"},
	Features: []features.FeatureName{
		features.SupportUDPRoute,
		features.SupportGateway,
	},
	Provisional: true,
	Test: func(t *testing.T, suite *confsuite.ConformanceTestSuite) {
		t.Run("Simple UDP request matching UDPRoute should reach l4-backend", func(t *testing.T) {
			namespace := confsuite.InfrastructureNamespace
			routeNN := types.NamespacedName{Name: "udp-l4-backend", Namespace: namespace}
			gwNN := types.NamespacedName{Name: "udp-gateway", Namespace: namespace}
			gwAddr := kubernetes.GatewayAndUDPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			udp.ExpectEchoResponseFromBackend(t, suite.TimeoutConfig.MaxTimeToConsistency, gwAddr, udp.ExpectedResponse{
				Service:   "l4-backend",
				Namespace: namespace,
			})
		})
	},
}
