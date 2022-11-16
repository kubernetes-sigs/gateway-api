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

package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var path = *new(field.Path)

func TestValidateParentRefs(t *testing.T) {
	namespace := gatewayv1b1.Namespace("example-namespace")
	kind := gatewayv1b1.Kind("Gateway")
	sectionA := gatewayv1b1.SectionName("Section A")
	sectionB := gatewayv1b1.SectionName("Section B")
	sectionC := gatewayv1b1.SectionName("Section C")

	tests := []struct {
		name       string
		parentRefs []gatewayv1b1.ParentReference
		errCount   int
	}{{
		name: "valid ParentRefs includes 1 reference",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
		},
		errCount: 0,
	}, {
		name: "valid ParentRefs includes 2 references",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionB,
			},
		},
		errCount: 0,
	}, {
		name: "valid ParentRefs when different references have the same section name",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example A",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
			{
				Name:        "example B",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
		},
		errCount: 0,
	}, {
		name: "valid ParentRefs includes more references to the same parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionB,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionC,
			},
		},
		errCount: 0,
	}, {
		name: "invalid ParentRefs due to the same section names to the same parentRefs",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				Kind:        &kind,
				SectionName: &sectionA,
			},
		},
		errCount: 1,
	}, {
		name: "invalid ParentRefs due to section names not set to the same ParentRefs",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name: "example",
			},
			{
				Name: "example",
			},
		},
		errCount: 1,
	}, {
		name: "invalid ParentRefs due to more same section names to the same ParentRefs",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: nil,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionB,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionA,
			},
		},
		errCount: 1,
	}, {
		name: "invalid ParentRefs when one ParentRef section name not set to the same ParentRefs",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: nil,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionA,
			},
		},
		errCount: 1,
	}, {
		name: "invalid ParentRefs when next ParentRef section name not set to the same ParentRefs",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: nil,
			},
		},
		errCount: 1,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := gatewayv1b1.CommonRouteSpec{
				ParentRefs: tc.parentRefs,
			}
			errs := ValidateParentRefs(spec.ParentRefs, path.Child("spec"))
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}
