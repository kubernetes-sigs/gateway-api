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
	"time"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGRPCRouteFilter(t *testing.T) {
	tests := []struct {
		name        string
		wantErrors  []string
		routeFilter gatewayv1.GRPCRouteFilter
	}{
		{
			name: "valid GRPCRouteFilterRequestHeaderModifier route filter",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set:    []gatewayv1.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterRequestHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type:          gatewayv1.GRPCRouteFilterRequestHeaderModifier,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid GRPCRouteFilterRequestHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterRequestHeaderModifier,
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type"},
		},
		{
			name: "valid GRPCRouteFilterResponseHeaderModifier route filter",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set:    []gatewayv1.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterResponseHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type:          gatewayv1.GRPCRouteFilterResponseHeaderModifier,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.responseHeaderModifier must be specified for ResponseHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid GRPCRouteFilterResponseHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterResponseHeaderModifier,
			},
			wantErrors: []string{"filter.responseHeaderModifier must be specified for ResponseHeaderModifier filter.type"},
		},
		{
			name: "valid GRPCRouteFilterRequestMirror route filter",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterRequestMirror,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{BackendRef: gatewayv1.BackendObjectReference{
					Group:     ptrTo(gatewayv1.Group("group")),
					Kind:      ptrTo(gatewayv1.Kind("kind")),
					Name:      "name",
					Namespace: ptrTo(gatewayv1.Namespace("ns")),
					Port:      ptrTo(gatewayv1.PortNumber(22)),
				}},
			},
		},
		{
			name: "invalid GRPCRouteFilterRequestMirror type filter with non-matching field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type:                  gatewayv1.GRPCRouteFilterRequestMirror,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be nil if the filter.type is not RequestHeaderModifier", "filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "invalid GRPCRouteFilterRequestMirror type filter with empty value field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterRequestMirror,
			},
			wantErrors: []string{"filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "valid GRPCRouteFilterExtensionRef filter",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: "group",
					Kind:  "kind",
					Name:  "name",
				},
			},
		},
		{
			name: "invalid GRPCRouteFilterExtensionRef type filter with non-matching field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type:          gatewayv1.GRPCRouteFilterExtensionRef,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "invalid GRPCRouteFilterExtensionRef type filter with empty value field",
			routeFilter: gatewayv1.GRPCRouteFilter{
				Type: gatewayv1.GRPCRouteFilterExtensionRef,
			},
			wantErrors: []string{"filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.GRPCRouteSpec{
					Rules: []gatewayv1.GRPCRouteRule{{
						Filters: []gatewayv1.GRPCRouteFilter{tc.routeFilter},
					}},
				},
			}
			validateGRPCRoute(t, route, tc.wantErrors)
		})
	}
}

func TestGRPCRouteRule(t *testing.T) {
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.GRPCRouteRule
	}{
		{
			name: "valid GRPCRoute with no filters",
			rules: []gatewayv1.GRPCRouteRule{
				{
					Matches: []gatewayv1.GRPCRouteMatch{
						{
							Method: &gatewayv1.GRPCMethodMatch{
								Type:    ptrTo(gatewayv1.GRPCMethodMatchType("Exact")),
								Service: ptrTo("helloworld.Greeter"),
							},
						},
					},
					BackendRefs: []gatewayv1.GRPCBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
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
			rules: []gatewayv1.GRPCRouteRule{
				{
					Matches: []gatewayv1.GRPCRouteMatch{
						{
							Method: &gatewayv1.GRPCMethodMatch{
								Type:   ptrTo(gatewayv1.GRPCMethodMatchType("Exact")),
								Method: ptrTo("SayHello"),
							},
						},
					},
					BackendRefs: []gatewayv1.GRPCBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
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
			rules: []gatewayv1.GRPCRouteRule{
				{
					Matches: []gatewayv1.GRPCRouteMatch{
						{
							Method: &gatewayv1.GRPCMethodMatch{
								Type:    ptrTo(gatewayv1.GRPCMethodMatchType("Exact")),
								Service: ptrTo("helloworld.Greeter"),
							},
						},
					},
					Filters: []gatewayv1.GRPCRouteFilter{
						{
							Type: gatewayv1.GRPCRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Set: []gatewayv1.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1.GRPCRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Add: []gatewayv1.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
						{
							Type: gatewayv1.GRPCRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Set: []gatewayv1.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1.GRPCRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Add: []gatewayv1.HTTPHeader{
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
			route := &gatewayv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.GRPCRouteSpec{Rules: tc.rules},
			}
			validateGRPCRoute(t, route, tc.wantErrors)
		})
	}
}

func TestGRPCMethodMatch(t *testing.T) {
	tests := []struct {
		name       string
		method     gatewayv1.GRPCMethodMatch
		wantErrors []string
	}{
		{
			name: "valid GRPCRoute with 1 service in GRPCMethodMatch field",
			method: gatewayv1.GRPCMethodMatch{
				Service: ptrTo("foo.Test.Example"),
			},
		},
		{
			name: "valid GRPCRoute with 1 method in GRPCMethodMatch field",
			method: gatewayv1.GRPCMethodMatch{
				Method: ptrTo("Login"),
			},
		},
		{
			name: "invalid GRPCRoute missing service or method in GRPCMethodMatch field",
			method: gatewayv1.GRPCMethodMatch{
				Service: nil,
				Method:  nil,
			},
			wantErrors: []string{"One or both of 'service' or 'method"},
		},
		{
			name: "GRPCRoute uses regex in service and method with undefined match type",
			method: gatewayv1.GRPCMethodMatch{
				Service: ptrTo(".*"),
				Method:  ptrTo(".*"),
			},
			wantErrors: []string{"service must only contain valid characters (matching ^(?i)\\.?[a-z_][a-z_0-9]*(\\.[a-z_][a-z_0-9]*)*$)", "method must only contain valid characters (matching ^[A-Za-z_][A-Za-z_0-9]*$)"},
		},
		{
			name: "GRPCRoute uses regex in service and method with match type Exact",
			method: gatewayv1.GRPCMethodMatch{
				Type:    ptrTo(gatewayv1.GRPCMethodMatchExact),
				Service: ptrTo(".*"),
				Method:  ptrTo(".*"),
			},
			wantErrors: []string{"service must only contain valid characters (matching ^(?i)\\.?[a-z_][a-z_0-9]*(\\.[a-z_][a-z_0-9]*)*$)", "method must only contain valid characters (matching ^[A-Za-z_][A-Za-z_0-9]*$)"},
		},
		{
			name: "GRPCRoute uses regex in method with undefined match type",
			method: gatewayv1.GRPCMethodMatch{
				Method: ptrTo(".*"),
			},
			wantErrors: []string{"method must only contain valid characters (matching ^[A-Za-z_][A-Za-z_0-9]*$)"},
		},
		{
			name: "GRPCRoute uses regex in service with match type Exact",
			method: gatewayv1.GRPCMethodMatch{
				Type:    ptrTo(gatewayv1.GRPCMethodMatchExact),
				Service: ptrTo(".*"),
			},
			wantErrors: []string{"service must only contain valid characters (matching ^(?i)\\.?[a-z_][a-z_0-9]*(\\.[a-z_][a-z_0-9]*)*$)"},
		},
		{
			name: "GRPCRoute uses regex in service and method with match type RegularExpression",
			method: gatewayv1.GRPCMethodMatch{
				Type:    ptrTo(gatewayv1.GRPCMethodMatchRegularExpression),
				Service: ptrTo(".*"),
				Method:  ptrTo(".*"),
			},
		},
		{
			name: "GRPCRoute uses valid service and method with undefined match type",
			method: gatewayv1.GRPCMethodMatch{
				Service: ptrTo("foo.Test.Example"),
				Method:  ptrTo("Login"),
			},
		},
		{
			name: "GRPCRoute uses valid service and method with match type Exact",
			method: gatewayv1.GRPCMethodMatch{
				Type:    ptrTo(gatewayv1.GRPCMethodMatchExact),
				Service: ptrTo("foo.Test.Example"),
				Method:  ptrTo("Login"),
			},
		},
		{
			name: "GRPCRoute uses a valid service with a leading dot when match type is Exact",
			method: gatewayv1.GRPCMethodMatch{
				Type:    ptrTo(gatewayv1.GRPCMethodMatchExact),
				Service: ptrTo(".foo.Test.Example"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.GRPCRouteSpec{
					Rules: []gatewayv1.GRPCRouteRule{
						{
							Matches: []gatewayv1.GRPCRouteMatch{
								{
									Method: &tc.method,
								},
							},
						},
					},
				},
			}
			validateGRPCRoute(t, &route, tc.wantErrors)
		})
	}
}

func validateGRPCRoute(t *testing.T, route *gatewayv1.GRPCRoute, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating GRPCRoute %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
