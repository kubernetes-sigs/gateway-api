/*
Copyright 2021 The Kubernetes Authors.

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

package utils

import (
	"testing"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func Test_PortNumberPtr(t *testing.T) {
	var exportedPort65535 gatewayv1b1.PortNumber = 65535
	var exportedPort1 gatewayv1b1.PortNumber = 1
	var exportedPort0 gatewayv1b1.PortNumber
	var exportedPort65536 gatewayv1b1.PortNumber = 65536

	portNumberPtrTests := []struct {
		name         string
		port         int
		expectedPort *gatewayv1b1.PortNumber
	}{
		{
			name:         "invalid port number",
			port:         0,
			expectedPort: &exportedPort0,
		},
		{
			name:         "valid port number",
			port:         65535,
			expectedPort: &exportedPort65535,
		},
		{
			name:         "invalid port number",
			port:         65536,
			expectedPort: &exportedPort65536,
		},
		{
			name:         "valid port number",
			port:         1,
			expectedPort: &exportedPort1,
		},
	}

	for _, tc := range portNumberPtrTests {
		t.Run(tc.name, func(t *testing.T) {
			port := PortNumberPtr(tc.port)
			if port == nil || tc.expectedPort == nil {
				if port != tc.expectedPort {
					t.Errorf("Expected port %d, got %d", tc.expectedPort, port)
				}
			} else if *port != *tc.expectedPort {
				t.Errorf("Expected port %d, got %d", *tc.expectedPort, *port)
			}
		})
	}
}

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
