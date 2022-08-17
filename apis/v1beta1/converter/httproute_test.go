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

package converter

import (
	"testing"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func Test_PathMatchTypePtr(t *testing.T) {
	pathMatchTypePtrTests := []struct {
		name         string
		pathType     string
		expectedPath gatewayv1b1.PathMatchType
	}{
		{
			name:         "valid path exact match",
			pathType:     "Exact",
			expectedPath: gatewayv1b1.PathMatchExact,
		},
		{
			name:         "valid path exact match using constant",
			pathType:     string(gatewayv1b1.PathMatchExact),
			expectedPath: gatewayv1b1.PathMatchExact,
		},
		{
			name:         "valid path prefix match",
			pathType:     "PathPrefix",
			expectedPath: gatewayv1b1.PathMatchPathPrefix,
		},
		{
			name:         "valid path regular expression match",
			pathType:     "RegularExpression",
			expectedPath: gatewayv1b1.PathMatchRegularExpression,
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
