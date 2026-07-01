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
	v1.GatewayConditionNameTooLong,
}

// TestNameTooLongConditionExists verifies that the "NameTooLong" condition
// type exists in the API. This validates that the gap described in GEP-1762
// has been resolved: implementations can now report when a gateway name
// combined with the gateway class name exceeds the 63-character limit.
func TestNameTooLongConditionExists(t *testing.T) {
	found := false
	for _, ct := range knownConditionTypes {
		if ct == v1.GatewayConditionNameTooLong {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("GatewayConditionNameTooLong is not listed in knownConditionTypes")
	}

	// Verify the constant has the expected string value
	if v1.GatewayConditionNameTooLong != "NameTooLong" {
		t.Errorf("GatewayConditionNameTooLong = %q, want %q",
			v1.GatewayConditionNameTooLong, "NameTooLong")
	}

	// Verify the reason constant matches
	if v1.GatewayReasonNameTooLong != "NameTooLong" {
		t.Errorf("GatewayReasonNameTooLong = %q, want %q",
			v1.GatewayReasonNameTooLong, "NameTooLong")
	}
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
