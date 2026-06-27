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
	"sort"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestExtractHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		headers  []*corev3.HeaderValue
		key      string
		expected string
	}{
		{
			name: "Exact match",
			headers: []*corev3.HeaderValue{
				{Key: "x-request-id", RawValue: []byte("123")},
			},
			key:      "x-request-id",
			expected: "123",
		},
		{
			name: "Case-insensitive match",
			headers: []*corev3.HeaderValue{
				{Key: "X-Request-ID", RawValue: []byte("456")},
			},
			key:      "x-request-id",
			expected: "456",
		},
		{
			name: "Non-existent key",
			headers: []*corev3.HeaderValue{
				{Key: "other-header", RawValue: []byte("abc")},
			},
			key:      "x-request-id",
			expected: "",
		},
		{
			name: "String value fallback",
			headers: []*corev3.HeaderValue{
				{Key: "fallback-test", Value: "only-string"},
			},
			key:      "fallback-test",
			expected: "only-string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &extProcPb.ProcessingRequest_RequestHeaders{
				RequestHeaders: &extProcPb.HttpHeaders{
					Headers: &corev3.HeaderMap{
						Headers: tt.headers,
					},
				},
			}

			result := ExtractHeaderValue(req, tt.key)
			if result != tt.expected {
				t.Errorf("ExtractHeaderValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateHeadersMutation(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    []*corev3.HeaderValueOption
	}{
		{
			name:    "empty map returns empty slice",
			headers: map[string]string{},
			want:    []*corev3.HeaderValueOption{},
		},
		{
			name:    "single header",
			headers: map[string]string{"x-api-key": "secret-123"},
			want: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:      "x-api-key",
						RawValue: []byte("secret-123"),
					},
				},
			},
		},
		{
			name: "multiple headers",
			headers: map[string]string{
				"x-api-key":     "key-val",
				"x-request-id":  "req-456",
				"authorization": "Bearer tok",
			},
			want: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:      "authorization",
						RawValue: []byte("Bearer tok"),
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:      "x-api-key",
						RawValue: []byte("key-val"),
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:      "x-request-id",
						RawValue: []byte("req-456"),
					},
				},
			},
		},
		{
			name:    "header with empty value",
			headers: map[string]string{"x-empty": ""},
			want: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:      "x-empty",
						RawValue: []byte(""),
					},
				},
			},
		},
	}

	sortByKey := func(opts []*corev3.HeaderValueOption) {
		sort.Slice(opts, func(i, j int) bool {
			return opts[i].GetHeader().GetKey() < opts[j].GetHeader().GetKey()
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateHeadersMutation(tt.headers)
			sortByKey(got)
			sortByKey(tt.want)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("GenerateHeadersMutation() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGetHeaderValue(t *testing.T) {
	tests := []struct {
		name     string
		header   *corev3.HeaderValue
		expected string
	}{
		{
			name: "Prefers RawValue when present",
			header: &corev3.HeaderValue{
				Key:      "x-test",
				RawValue: []byte("raw-content"),
				Value:    "string-content", // Should be ignored
			},
			expected: "raw-content",
		},
		{
			name: "Falls back to Value when RawValue is nil",
			header: &corev3.HeaderValue{
				Key:      "x-test",
				RawValue: nil,
				Value:    "string-content",
			},
			expected: "string-content",
		},
		{
			name: "Falls back to Value when RawValue is empty slice",
			header: &corev3.HeaderValue{
				Key:      "x-test",
				RawValue: []byte{},
				Value:    "string-content",
			},
			expected: "string-content",
		},
		{
			name: "Returns empty if both are empty",
			header: &corev3.HeaderValue{
				Key:      "x-test",
				RawValue: []byte{},
				Value:    "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetHeaderValue(tt.header)
			if result != tt.expected {
				t.Errorf("GetHeaderValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}
