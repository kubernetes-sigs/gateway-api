//go:build standard
// +build standard

/*
Copyright 2023 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"testing"
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
//
// How are tests named? Where to add new tests?
//
// Ensure that tests for newly added CEL validations are added in the correctly
// named test function. For example, if you added a test at the
// `HTTPRouteFilter` hierarchy (i.e. either at the struct level, or on one of
// the immediate descendent fields), then the test will go in the
// TestHTTPRouteFilter function. If the appropriate test function does not
// exist, please create one.
//
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestHTTPRouteParentRefStandard(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		parentRefs []gatewayv1.ParentReference
	}{
		{
			name:       "invalid because duplicate parent refs without section name",
			wantErrors: []string{"sectionName must be unique when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "invalid because duplicate parent refs with only one section name",
			wantErrors: []string{"sectionName must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "invalid because duplicate parent refs with duplicate section names",
			wantErrors: []string{"sectionName must be unique when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid single parentRef without sectionName",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "valid single parentRef with sectionName",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because duplicate parent refs with different section names",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("bar")),
			}},
		},
		{
			name:       "valid because duplicate parent refs with different names",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example2",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because first parentRef has namespace while second doesn't",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:      ptrTo(gatewayv1.Kind("Gateway")),
				Group:     ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:      "example",
				Namespace: ptrTo(gatewayv1.Namespace("test")),
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "valid because second parentRef has namespace while first doesn't",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:      ptrTo(gatewayv1.Kind("Gateway")),
				Group:     ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:      "example",
				Namespace: ptrTo(gatewayv1.Namespace("test")),
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1.CommonRouteSpec{
						ParentRefs: tc.parentRefs,
					},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}
