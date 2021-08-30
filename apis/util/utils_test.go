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
	"fmt"
	"testing"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func Test_Utils(t *testing.T) {

	var exportedPort1 gatewayv1a2.PortNumber = 65535
	var exportedPort3 gatewayv1a2.PortNumber = 1

	var expectedPathMatch1 = gatewayv1a2.PathMatchExact
	var expectedPathMatch2 = gatewayv1a2.PathMatchPrefix
	var expectedPathMatch3 = gatewayv1a2.PathMatchRegularExpression
	var expectedPathMatch4 = gatewayv1a2.PathMatchImplementationSpecific

	table := []struct {
		pathType     string
		expectedPath *gatewayv1a2.PathMatchType
		port         int
		expectedPort *gatewayv1a2.PortNumber
	}{
		{
			pathType:     "Exact",
			expectedPath: &expectedPathMatch1,
			port:         0,
			expectedPort: nil,
		},
		{
			pathType:     "Exact",
			expectedPath: &expectedPathMatch1,
			port:         65535,
			expectedPort: &exportedPort1,
		},
		{
			pathType:     "Exact",
			expectedPath: &expectedPathMatch1,
			port:         65536,
			expectedPort: nil,
		},
		{
			pathType:     "Prefix",
			expectedPath: &expectedPathMatch2,
			port:         0,
			expectedPort: nil,
		},
		{
			pathType:     "RegularExpression",
			expectedPath: &expectedPathMatch3,
			port:         65536,
			expectedPort: nil,
		},
		{
			pathType:     "ImplementationSpecific",
			expectedPath: &expectedPathMatch4,
			port:         1,
			expectedPort: &exportedPort3,
		},
		{
			pathType:     "APrefix",
			expectedPath: nil,
			port:         1,
			expectedPort: &exportedPort3,
		},
	}

	for _, entry := range table {
		path := PathMatchTypePtr(entry.pathType)
		if path == nil && entry.expectedPath != nil {
			t.Error("failed in path match type pointer. not expecting nil, but get nil.")
		}

		if path != nil && entry.expectedPath == nil {
			t.Error("failed in path match type pointer. expecting nil but get non-nil")
		}

		if path != nil && entry.expectedPath != nil && *path != *entry.expectedPath {
			t.Error("failed in path match type pointer. go unexpected.")
		}

		port := PortNumberPtr(entry.port)
		if port == nil && entry.expectedPort != nil {
			t.Error("failed in port number pointer. not expected nil, but got nil.")
		}
		if port != nil && entry.expectedPort == nil {
			t.Error("failed in port number port. expecting nil, but got non-nil.")
		}
		if port != nil && entry.expectedPort != nil && *port != *entry.expectedPort {
			fmt.Printf("failed in port number expecting  %d got port %d", entry.expectedPort, *port)
		}

	}
}
