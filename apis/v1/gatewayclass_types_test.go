/*
Copyright 2024 The Kubernetes Authors.

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

package v1

import (
	"encoding/json"
	"testing"
)

func TestSupportedFeature_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    SupportedFeature
		expectError bool
	}{
		{
			name:     "old struct input",
			input:    `"featureA"`,
			expected: SupportedFeature{Name: "featureA"},
		},
		{
			name:     "new struct input",
			input:    `{"name": "featureB"}`,
			expected: SupportedFeature{Name: "featureB"},
		},
		{
			name:        "expected error",
			input:       `["featureA", "featureB"]`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result SupportedFeature
			err := json.Unmarshal([]byte(tt.input), &result)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %+v, got %+v", tt.expected, result)
				}
			}
		})
	}
}
