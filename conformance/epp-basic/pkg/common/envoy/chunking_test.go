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
	"crypto/rand"
	"testing"
)

func TestBuildChunkedBodyResponses(t *testing.T) {
	tests := []struct {
		name                 string
		count                int
		expectedMessageCount int
	}{
		{
			name:                 "zero case",
			count:                0,
			expectedMessageCount: 1,
		},
		{
			name:                 "below limit",
			count:                BodyByteLimit - 1000,
			expectedMessageCount: 1,
		},
		{
			name:                 "at limit",
			count:                BodyByteLimit,
			expectedMessageCount: 1,
		},
		{
			name:                 "off by one error?",
			count:                BodyByteLimit + 1,
			expectedMessageCount: 2,
		},
		{
			name:                 "above limit",
			count:                BodyByteLimit + 1000,
			expectedMessageCount: 2,
		},
		{
			name:                 "above limit",
			count:                (BodyByteLimit * 2) + 1000,
			expectedMessageCount: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			arr := generateBytes(test.count)
			responses := BuildChunkedBodyResponses(arr, true)
			for i, response := range responses {
				eos := response.BodyMutation.GetStreamedResponse().GetEndOfStream()
				if eos == true && i+1 != len(responses) {
					t.Fatalf("EoS should not be set")
				}
				if eos == false && i+1 == len(responses) {
					t.Fatalf("EoS should be set")
				}
			}
			if len(responses) != test.expectedMessageCount {
				t.Fatalf("Expected: %v, Got %v", test.expectedMessageCount, len(responses))
			}
		})
	}
}

func generateBytes(count int) []byte {
	arr := make([]byte, count)
	_, _ = rand.Read(arr)
	return arr
}
