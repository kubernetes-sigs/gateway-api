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

package translator

import (
	"testing"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_PathMatchTypePtr(t *testing.T) {
	pathMatchTypePtrTests := []struct {
		name         string
		pathType     string
		expectedPath gatewayv1a2.PathMatchType
	}{
		{
			name:         "valid path exact match",
			pathType:     "Exact",
			expectedPath: gatewayv1a2.PathMatchExact,
		},
		{
			name:         "valid path exact match using constant",
			pathType:     string(gatewayv1a2.PathMatchExact),
			expectedPath: gatewayv1a2.PathMatchExact,
		},
		{
			name:         "valid path prefix match",
			pathType:     "PathPrefix",
			expectedPath: gatewayv1a2.PathMatchPathPrefix,
		},
		{
			name:         "valid path regular expression match",
			pathType:     "RegularExpression",
			expectedPath: gatewayv1a2.PathMatchRegularExpression,
		},
	}

	for _, tc := range pathMatchTypePtrTests {
		t.Run(tc.name, func(t *testing.T) {
			path := PathMatchTypePtr(tc.pathType)
			if *path != tc.expectedPath {
				t.Errorf("Expected path %s, got %s", tc.expectedPath, *path)
			}
		})
	}
}
