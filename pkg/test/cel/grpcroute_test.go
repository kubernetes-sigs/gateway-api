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
	"context"
	"fmt"
	"testing"
	"strings"
	"time"
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGRPCRouteFilter(t *testing.T) {
	tests := []struct {
		name        string
		wantErrors  []string
		routeFilter gatewayv1a2.GRPCRouteFilter
	}{
		{
			name: "valid GRPCRouteFilterRequestHeaderModifier route filter",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
					Set:    []gatewayv1a2.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1a2.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterRequestHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type:          gatewayv1a2.GRPCRouteFilterRequestHeaderModifier,
				RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid GRPCRouteFilterRequestHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterRequestHeaderModifier,
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type"},
		},
		{
			name: "valid GRPCRouteFilterResponseHeaderModifier route filter",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
					Set:    []gatewayv1a2.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1a2.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterResponseHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type:          gatewayv1a2.GRPCRouteFilterResponseHeaderModifier,
				RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.responseHeaderModifier must be specified for ResponseHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid GRPCRouteFilterResponseHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterResponseHeaderModifier,
			},
			wantErrors: []string{"filter.responseHeaderModifier must be specified for ResponseHeaderModifier filter.type"},
		},
		{
			name: "valid GRPCRouteFilterRequestMirror route filter",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterRequestMirror,
				RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{BackendRef: gatewayv1a2.BackendObjectReference{
					Group:     ptrTo(gatewayv1a2.Group("group")),
					Kind:      ptrTo(gatewayv1a2.Kind("kind")),
					Name:      "name",
					Namespace: ptrTo(gatewayv1a2.Namespace("ns")),
					Port:      ptrTo(gatewayv1a2.PortNumber(22)),
				}},
			},
		},
		{
			name: "invalid GRPCRouteFilterRequestMirror type filter with non-matching field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type:                  gatewayv1a2.GRPCRouteFilterRequestMirror,
				RequestHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be nil if the filter.type is not RequestHeaderModifier", "filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "invalid GRPCRouteFilterRequestMirror type filter with empty value field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterRequestMirror,
			},
			wantErrors: []string{"filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "valid GRPCRouteFilterExtensionRef filter",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1a2.LocalObjectReference{
					Group: "group",
					Kind:  "kind",
					Name:  "name",
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterExtensionRef type filter with non-matching field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type:          gatewayv1a2.GRPCRouteFilterExtensionRef,
				RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "invalid GRPCRouteFilterExtensionRef type filter with empty value field",
			routeFilter: gatewayv1a2.GRPCRouteFilter{
				Type: gatewayv1a2.GRPCRouteFilterExtensionRef,
			},
			wantErrors: []string{"filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1a2.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1a2.GRPCRouteSpec{
					Rules: []gatewayv1a2.GRPCRouteRule{{
						Filters: []gatewayv1a2.GRPCRouteFilter{tc.routeFilter},
					}},
				},
			}
			validateGRPCRoute(t, route, tc.wantErrors)
		})
	}
}

func TestGRPCRouteRule(t *testing.T) {
	testService := gatewayv1a2.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1a2.GRPCRouteRule
	}{
		{
			name: "valid GRPCRoute with no filters",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Type:  ptrTo(gatewayv1a2.GRPCMethodMatchType("Exact")),
								Service: ptrTo("helloworld.Greeter"),
							},
						},
					},
					BackendRefs: []gatewayv1a2.GRPCBackendRef{
						{
							BackendRef: gatewayv1a2.BackendRef{
								BackendObjectReference: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1a2.PortNumber(8080)),
								},
								Weight: ptrTo(int32(100)),
							},
						},
					},
				},
			},
		},
		{
			name: "valid GRPCRoute with only Method specified",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Type:  ptrTo(gatewayv1a2.GRPCMethodMatchType("Exact")),
								Method: ptrTo("SayHello"),
							},
						},
					},
					BackendRefs: []gatewayv1a2.GRPCBackendRef{
						{
							BackendRef: gatewayv1a2.BackendRef{
								BackendObjectReference: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1a2.PortNumber(8080)),
								},
								Weight: ptrTo(int32(100)),
							},
						},
					},
				},
			},
		},
		{
			name:       "invalid because multiple filters are repeated",
			wantErrors: []string{"RequestHeaderModifier filter cannot be repeated", "ResponseHeaderModifier filter cannot be repeated"},
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Type:  ptrTo(gatewayv1a2.GRPCMethodMatchType("Exact")),
								Service: ptrTo("helloworld.Greeter"),
							},
						},
					},
					Filters: []gatewayv1a2.GRPCRouteFilter{
						{
							Type: gatewayv1a2.GRPCRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
								Set: []gatewayv1a2.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.GRPCRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
								Add: []gatewayv1a2.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.GRPCRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
								Set: []gatewayv1a2.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.GRPCRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1a2.HTTPHeaderFilter{
								Add: []gatewayv1a2.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1a2.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1a2.GRPCRouteSpec{Rules: tc.rules},
			}
			validateGRPCRoute(t, route, tc.wantErrors)
		})
	}
}

func validateGRPCRoute(t *testing.T, route *gatewayv1a2.GRPCRoute, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
