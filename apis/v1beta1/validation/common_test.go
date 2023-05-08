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
	"strings"
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
		err        string
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
		err: "",
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
		err: "",
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
		err: "",
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
		err: "",
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
		err: "must be unique when ParentRefs",
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
		err: "sectionNames or ports must be specified",
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
		err: "sectionNames or ports must be specified",
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
		err: "sectionNames or ports must be specified",
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
		err: "sectionNames or ports must be specified",
	}, {
		name: "valid ParentRefs with multiple port references to the same parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(80)),
			},
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(81)),
			},
		},
		err: "",
	}, {
		name: "valid ParentRefs with multiple mixed references to the same parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(80)),
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				SectionName: &sectionA,
			},
		},
		err: "",
	}, {
		name: "invalid ParentRefs due to same port references to the same parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(80)),
			},
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(80)),
			},
		},
		err: "port: Invalid value: 80: must be unique when ParentRefs",
	}, {
		name: "invalid ParentRefs due to mixed port references to the same parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      ptrTo(gatewayv1b1.PortNumber(80)),
			},
			{
				Name:      "example",
				Namespace: &namespace,
				Port:      nil,
			},
		},
		err: "Required value: sectionNames or ports must be specified",
	}, {
		name: "valid ParentRefs with multiple same port references to different section of a  parent",
		parentRefs: []gatewayv1b1.ParentReference{
			{
				Name:        "example",
				Namespace:   &namespace,
				Port:        ptrTo(gatewayv1b1.PortNumber(80)),
				SectionName: &sectionA,
			},
			{
				Name:        "example",
				Namespace:   &namespace,
				Port:        ptrTo(gatewayv1b1.PortNumber(80)),
				SectionName: &sectionB,
			},
		},
		err: "",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			spec := gatewayv1b1.CommonRouteSpec{
				ParentRefs: tc.parentRefs,
			}
			errs := ValidateParentRefs(spec.ParentRefs, path.Child("spec"))
			if tc.err == "" {
				if len(errs) != 0 {
					t.Errorf("got %d errors, want none: %s", len(errs), errs)
				}
			} else {
				if errs == nil {
					t.Errorf("got no errors, want %q", tc.err)
				} else if !strings.Contains(errs.ToAggregate().Error(), tc.err) {
					t.Errorf("got %d errors, want %q: %s", len(errs), tc.err, errs)
				}
			}
		})
	}
}
