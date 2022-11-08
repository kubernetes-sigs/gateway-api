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

package validation_test

import (
	"testing"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
	validationtutils "sigs.k8s.io/gateway-api/apis/v1beta1/util/validation"
)

func TestContainsInSectionNameSlice(t *testing.T) {
	targetSectionSlice := []gatewayv1b1.SectionName{
		gatewayv1b1.SectionName("Section A"),
		gatewayv1b1.SectionName("Section B"),
	}
	testCases := []struct {
		name        string
		sectionName gatewayv1b1.SectionName
		isvalid     bool
	}{
		{
			name:        "Section found in slice",
			sectionName: "Section A",
			isvalid:     true,
		},
		{
			name:        "SectionName not found in slice",
			sectionName: "Section C",
			isvalid:     false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, isValid := validationtutils.ContainsInSectionNameSlice(targetSectionSlice, &tc.sectionName)
			if isValid != tc.isvalid {
				t.Errorf("Expected validity %t, got %t", tc.isvalid, isValid)
			}
		})
	}
}
