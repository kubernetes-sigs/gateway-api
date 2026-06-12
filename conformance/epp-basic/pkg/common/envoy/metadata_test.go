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

package envoy

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestExtractMetadataValues(t *testing.T) {
	var makeFilterMetadata = func() map[string]*structpb.Struct {
		structVal, _ := structpb.NewStruct(map[string]any{
			"hello":      "world",
			"random-key": []any{"hello", "world"},
		})

		return map[string]*structpb.Struct{
			"key-1": structVal,
		}
	}

	tests := []struct {
		name     string
		metadata map[string]*structpb.Struct
		expected map[string]any
	}{
		{
			name:     "Exact match",
			metadata: makeFilterMetadata(),
			expected: map[string]any{
				"key-1": map[string]any{
					"hello":      "world",
					"random-key": []any{"hello", "world"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &extProcPb.ProcessingRequest{
				MetadataContext: &corev3.Metadata{
					FilterMetadata: tt.metadata,
				},
			}

			result := ExtractMetadataValues(req)
			if diff := cmp.Diff(result, tt.expected); diff != "" {
				t.Errorf("ExtractMetadataValues() unexpected response (-want +got): %v", diff)
			}
		})
	}
}
