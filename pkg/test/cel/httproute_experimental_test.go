//go:build experimental
// +build experimental

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

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

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

func TestHTTPRouteParentRefExperimental(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		parentRefs []gatewayv1b1.ParentReference
	}{
		{
			name:       "invalid because duplicate parent refs without port or section name",
			wantErrors: []string{"sectionName or port must be unique when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "invalid because duplicate parent refs with only one port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(80)),
			}},
		},
		{
			name:       "invalid because duplicate parent refs with only one sectionName and port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:        ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
				Port:        ptrTo(gatewayv1b1.PortNumber(80)),
			}},
		},
		{
			name:       "invalid because duplicate parent refs with duplicate ports",
			wantErrors: []string{"sectionName or port must be unique when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(80)),
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(80)),
			}},
		},
		{
			name:       "valid single parentRef without sectionName or port",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "valid single parentRef with sectionName and port",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:        ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
				Port:        ptrTo(gatewayv1b1.PortNumber(443)),
			}},
		},
		{
			name:       "valid because duplicate parent refs with different ports",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(80)),
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(443)),
			}},
		},
		{
			name:       "invalid ParentRefs with multiple mixed references to the same parent",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:        ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1b1.PortNumber(443)),
			}},
		},
		{
			name:       "valid ParentRefs with multiple same port references to different section of a parent",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Name:        "example",
				Port:        ptrTo(gatewayv1b1.PortNumber(443)),
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
			}, {
				Name:        "example",
				Port:        ptrTo(gatewayv1b1.PortNumber(443)),
				SectionName: ptrTo(gatewayv1b1.SectionName("bar")),
			}},
		},
		{
			// when referencing the same object, both parentRefs need to specify
			// the same optional fields (both parentRefs must specify port,
			// sectionName, or both)
			name:       "invalid because duplicate parent refs with first having sectionName and second having both sectionName and port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:        ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
			}, {
				Kind:        ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				Port:        ptrTo(gatewayv1b1.PortNumber(443)),
				SectionName: ptrTo(gatewayv1b1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because first parentRef has namespace while second doesn't",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:      ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:     ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:      "example",
				Namespace: ptrTo(gatewayv1b1.Namespace("test")),
			}, {
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "valid because second parentRef has namespace while first doesn't",
			wantErrors: []string{},
			parentRefs: []gatewayv1b1.ParentReference{{
				Kind:  ptrTo(gatewayv1b1.Kind("Gateway")),
				Group: ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:      ptrTo(gatewayv1b1.Kind("Gateway")),
				Group:     ptrTo(gatewayv1b1.Group("gateway.networking.k8s.io")),
				Name:      "example",
				Namespace: ptrTo(gatewayv1b1.Namespace("test")),
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1b1.HTTPRouteSpec{
					CommonRouteSpec: gatewayv1b1.CommonRouteSpec{
						ParentRefs: tc.parentRefs,
					},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func toDuration(durationString string) *gatewayv1b1.Duration {
	return (*gatewayv1b1.Duration)(&durationString)
}

func TestHTTPRouteTimeouts(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1b1.HTTPRouteRule
	}{
		{
			name:       "invalid timeout unit us is not supported",
			wantErrors: []string{"Invalid value: \"100us\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request: toDuration("100us"),
					},
				},
			},
		},
		{
			name:       "invalid timeout unit ns is not supported",
			wantErrors: []string{"Invalid value: \"500ns\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request: toDuration("500ns"),
					},
				},
			},
		},
		{
			name: "valid timeout request and backendRequest",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("4s"),
						BackendRequest: toDuration("2s"),
					},
				},
			},
		},
		{
			name: "valid timeout request",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request: toDuration("0s"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request day unit not supported",
			wantErrors: []string{"Invalid value: \"1d\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request: toDuration("1d"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request decimal not supported ",
			wantErrors: []string{"Invalid value: \"0.5s\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request: toDuration("0.5s"),
					},
				},
			},
		},
		{
			name: "valid timeout request infinite greater than backendRequest 1ms",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("0s"),
						BackendRequest: toDuration("1ms"),
					},
				},
			},
		},
		{
			name: "valid timeout request 1s greater than backendRequest 200ms",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("1s"),
						BackendRequest: toDuration("200ms"),
					},
				},
			},
		},
		{
			name: "valid timeout request 10s equal backendRequest 10s",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("10s"),
						BackendRequest: toDuration("10s"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request 200ms less than backendRequest 1s",
			wantErrors: []string{"Invalid value: \"object\": backendRequest timeout cannot be longer than request timeout"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("200ms"),
						BackendRequest: toDuration("1s"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1b1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1b1.HTTPRouteSpec{Rules: tc.rules},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}
