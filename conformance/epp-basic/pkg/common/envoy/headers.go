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
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// GetHeaderValue safely extracts the string value from an Envoy HeaderValue field.
func GetHeaderValue(header *corev3.HeaderValue) string {
	if len(header.RawValue) > 0 {
		return string(header.RawValue)
	}
	return header.Value
}

// ExtractHeaderValue searches for a specific header key in the processing request and returns its value.
// The lookup is case-insensitive.
// Returns an empty string if the header is missing or if the request structure is nil.
func ExtractHeaderValue(req *extProcPb.ProcessingRequest_RequestHeaders, headerKey string) string {
	headerKeyInLower := strings.ToLower(headerKey)
	if req != nil && req.RequestHeaders != nil && req.RequestHeaders.Headers != nil {
		for _, headerKv := range req.RequestHeaders.Headers.Headers {
			if strings.ToLower(headerKv.Key) == headerKeyInLower {
				return GetHeaderValue(headerKv)
			}
		}
	}
	return ""
}

func GenerateHeadersMutation(headers map[string]string) []*corev3.HeaderValueOption {
	headersMutation := make([]*corev3.HeaderValueOption, 0, len(headers))
	for key, value := range headers {
		headersMutation = append(headersMutation, &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:      key,
				RawValue: []byte(value),
			},
		})
	}
	return headersMutation
}
