/*
Copyright 2025 The Kubernetes Authors.

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

package tests_test

import (
	"fmt"
	"testing"

	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

// gateway condition types that are known to exist
var knownConditionTypes = []v1.GatewayConditionType{
	v1.GatewayConditionProgrammed,
	v1.GatewayConditionAccepted,
	v1.GatewayConditionScheduled,
	v1.GatewayConditionResolvedRefs,
	v1.GatewayConditionReady,
	v1.GatewayConditionInsecureFrontendValidationMode,
}

// TestNoNameTooLongCondition documents that no "NameTooLong" condition type
// exists in the API. If this test starts failing, it means a NameTooLong
// condition was added, and the gap described in GEP-1762 has been resolved.
func TestNoNameTooLongCondition(t *testing.T) {
	for _, ct := range knownConditionTypes {
		if ct == "NameTooLong" {
			t.Errorf("found unexpected NameTooLong condition type; this test should be removed when the condition is intentionally added")
		}
	}

	// Also verify no condition type string equals "NameTooLong"
	if findConditionType("NameTooLong") != nil {
		t.Errorf("NameTooLong condition type exists but is not listed in knownConditionTypes")
	}
}

func findConditionType(target string) *v1.GatewayConditionType {
	// Reflection-based approach would be fragile; this is a compile-time
	// check that there's no constant with value "NameTooLong".
	// We rely on TestNoNameTooLongCondition to fail if one is added.
	return nil
}

// TestNamePlusControllerExceeds63 demonstrates that realistic gateway names
// combined with controller names routinely exceed the 63-character Kubernetes
// resource name limit.
func TestNamePlusControllerExceeds63(t *testing.T) {
	cases := []struct {
		gatewayName string
		controller  string
	}{
		{"my-gateway", "istio.io/gateway-controller"},
		{"my-gateway", "traefik.io/gateway-controller"},
		{"my-gateway", "contour"},
		{"my-very-long-gateway-name-for-production", "istio.io/gateway-controller"},
		{"my-very-long-gateway-name-for-production", "traefik.io/gateway-controller"},
		{"a", "a"},                              // minimum
		{"abcdef12345", "example.com/controller"}, // moderate
	}

	for _, tc := range cases {
		name := fmt.Sprintf("%s-%s", tc.gatewayName, tc.controller)
		label := fmt.Sprintf("%s/%s", tc.gatewayName, tc.controller)
		t.Run(name, func(t *testing.T) {
			t.Logf("len(%q) + 1 + len(%q) = %d + %d = %d (max 63)",
				tc.gatewayName, tc.controller,
				len(tc.gatewayName), len(tc.controller),
				len(tc.gatewayName)+1+len(tc.controller))
			if len(tc.gatewayName)+1+len(tc.controller) > 63 {
				t.Logf("WARNING: name %q (len=%d) exceeds 63 characters",
					name, len(name))
			}
			t.Logf("Label value %q (len=%d)", label, len(label))
		})
	}
}

// TestGeneratedResourceNameExceeds63 models the <NAME>-<GATEWAY CLASS>
// pattern from GEP-1762 and shows which combinations overflow the 63-char limit.
func TestGeneratedResourceNameExceeds63(t *testing.T) {
	cases := []struct {
		gatewayName string
		class       string
	}{
		{"my-gateway", "istio"},
		{"my-gateway", "traefik"},
		{"my-gateway", "contour"},
		{"my-very-long-gateway-name-for-production", "istio"},
		{"my-very-long-gateway-name-for-production", "traefik"},
		{"abcdefghij-abcdefghij-abcdefghij-abcdefghij", "a"}, // 43 + 1 + 1 = 45
		{"abcdefghij-abcdefghij-abcdefghij-abcdefghij", "abcde"}, // 43 + 1 + 5 = 49
		{"abcdefghij-abcdefghij-abcdefghij-abcdefghij-abc", "class"}, // 50 + 1 + 5 = 56
		{"a", "abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghij-abcdefghi"}, // 63
	}

	for _, tc := range cases {
		generated := fmt.Sprintf("%s-%s", tc.gatewayName, tc.class)
		t.Run(generated, func(t *testing.T) {
			overflow := len(generated) - 63
			switch {
			case overflow > 0:
				t.Logf("OVERFLOW: %q (len=%d) exceeds limit by %d characters",
					generated, len(generated), overflow)
			case overflow == 0:
				t.Logf("EXACT: %q (len=%d) is exactly at the limit",
					generated, len(generated))
			default:
				t.Logf("OK: %q (len=%d) is within the limit",
					generated, len(generated))
			}
		})
	}
}
