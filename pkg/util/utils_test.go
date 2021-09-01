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

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_Utils(t *testing.T) {

	var exportedPort1 gatewayv1a2.PortNumber = 65535
	var exportedPort3 gatewayv1a2.PortNumber = 1

	portNumberPtrTests := []struct {
		Name         string
		port         int
		expectedPort *gatewayv1a2.PortNumber
	}{
		{
			Name:         "invalid port number",
			port:         0,
			expectedPort: nil,
		},
		{
			Name:         "valid port number",
			port:         65535,
			expectedPort: &exportedPort1,
		},
		{
			Name:         "invalid port number",
			port:         65536,
			expectedPort: nil,
		},
		{
			Name:         "valid port number",
			port:         1,
			expectedPort: &exportedPort3,
		},
	}

	for _, tt := range portNumberPtrTests {
		t.Run(tt.Name, func(t *testing.T) {
			port := PortNumberPtr(tt.port)
			if port == nil || tt.expectedPort == nil {
				if port != tt.expectedPort {
					t.Errorf("Expected port %d, got %d", tt.expectedPort, port)
				}
			} else if *port != *tt.expectedPort {
				t.Errorf("Expected port %d, got %d", *tt.expectedPort, *port)
			}
		})
	}

	pathmatchtypePtrTests := []struct {
		Name         string
		pathType     string
		expectedPath gatewayv1a2.PathMatchType
	}{
		{
			Name:         "valid path exact match",
			pathType:     "Exact",
			expectedPath: gatewayv1a2.PathMatchExact,
		},

		{
			Name:         "valid path prefix match",
			pathType:     "Prefix",
			expectedPath: gatewayv1a2.PathMatchPrefix,
		},
		{
			Name:         "valid path regular expression match",
			pathType:     "RegularExpression",
			expectedPath: gatewayv1a2.PathMatchRegularExpression,
		},
		{
			Name:         "valid path regular implementation specific match",
			pathType:     "ImplementationSpecific",
			expectedPath: gatewayv1a2.PathMatchImplementationSpecific,
		},
	}

	for _, tt := range pathmatchtypePtrTests {
		t.Run(tt.Name, func(t *testing.T) {
			path := PathMatchTypePtr(tt.pathType)
			if *path != tt.expectedPath {
				t.Errorf("Expected path %s, got %s", tt.expectedPath, *path)
			}
		})
	}
}
