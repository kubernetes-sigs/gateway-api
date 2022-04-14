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

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewaySecretInvalidReferencePolicy)
}

var GatewaySecretInvalidReferencePolicy = suite.ConformanceTest{
	ShortName:   "GatewaySecretInvalidReferencePolicy",
	Description: "A Gateway in the gateway-conformance-infra namespace should fail to become ready if the Gateway has a certificateRef for a Secret in the gateway-conformance-web-backend namespace and a ReferencePolicy exists but does not grant permission to that specific Secret",
	Features:    []suite.SupportedFeature{suite.SupportReferencePolicy},
	Manifests:   []string{"tests/gateway-secret-invalid-reference-policy.yaml"},
	Test: func(t *testing.T, s *suite.ConformanceTestSuite) {

	},
}
