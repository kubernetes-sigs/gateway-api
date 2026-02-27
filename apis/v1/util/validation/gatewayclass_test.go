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

package validation

import (
	"strings"
	"testing"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestIsControllerNameValid(t *testing.T) {
	tests := []struct {
		name           string
		controllerName gatewayv1.GatewayController
		want           bool
	}{
		// Valid cases
		{
			name:           "valid simple controller name",
			controllerName: "example.com/gateway-controller",
			want:           true,
		},
		{
			name:           "valid with subdomain",
			controllerName: "gateway.example.com/controller",
			want:           true,
		},
		{
			name:           "valid with multiple path segments",
			controllerName: "example.com/my/nested/controller",
			want:           true,
		},
		{
			name:           "valid with special characters in path",
			controllerName: "example.com/my-controller_v1.0",
			want:           true,
		},
		{
			name:           "valid with numbers in domain",
			controllerName: "ex4mple.com/controller",
			want:           true,
		},
		{
			name:           "valid with hyphen in domain",
			controllerName: "ex-ample.com/controller",
			want:           true,
		},

		// Invalid cases - empty
		{
			name:           "empty string",
			controllerName: "",
			want:           false,
		},

		// Invalid cases - missing slash
		{
			name:           "missing slash separator",
			controllerName: "example.com",
			want:           false,
		},

		// Invalid cases - domain issues
		{
			name:           "uppercase in domain",
			controllerName: "Example.com/controller",
			want:           false,
		},
		{
			name:           "domain starts with hyphen",
			controllerName: "-example.com/controller",
			want:           false,
		},
		{
			name:           "domain ends with hyphen",
			controllerName: "example-.com/controller",
			want:           false,
		},
		{
			name:           "domain starts with dot",
			controllerName: ".example.com/controller",
			want:           false,
		},

		// Invalid cases - path issues
		{
			name:           "empty path after slash",
			controllerName: "example.com/",
			want:           false,
		},

		// Edge case - length (NEW TEST for our improvement)
		{
			name:           "exactly at max length (253 chars)",
			controllerName: gatewayv1.GatewayController("example.com/" + strings.Repeat("a", 241)), // 12 + 241 = 253
			want:           true,
		},
		{
			name:           "exceeds max length (254 chars)",
			controllerName: gatewayv1.GatewayController("example.com/" + strings.Repeat("a", 242)), // 12 + 242 = 254
			want:           false,
		},
		{
			name:           "very long string (potential DoS)",
			controllerName: gatewayv1.GatewayController(strings.Repeat("a", 10000)),
			want:           false,
		},

		// Edge cases - whitespace
		{
			name:           "leading whitespace",
			controllerName: " example.com/controller",
			want:           false,
		},
		{
			name:           "trailing whitespace",
			controllerName: "example.com/controller ",
			want:           false,
		},
		{
			name:           "embedded whitespace in domain",
			controllerName: "exam ple.com/controller",
			want:           false,
		},
		{
			name:           "embedded whitespace in path",
			controllerName: "example.com/my controller",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsControllerNameValid(tt.controllerName)
			if got != tt.want {
				t.Errorf("IsControllerNameValid(%q) = %v, want %v", tt.controllerName, got, tt.want)
			}
		})
	}
}

// BenchmarkIsControllerNameValid ensures the function performs well
// even with edge case inputs
func BenchmarkIsControllerNameValid(b *testing.B) {
	validName := gatewayv1.GatewayController("example.com/gateway-controller")
	longName := gatewayv1.GatewayController(strings.Repeat("a", 10000))

	b.Run("valid short name", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsControllerNameValid(validName)
		}
	})

	b.Run("invalid long name", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsControllerNameValid(longName)
		}
	})
}
