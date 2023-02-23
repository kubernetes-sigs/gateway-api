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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestValidateGRPCRoute(t *testing.T) {
	t.Parallel()

	service := "foo.Test.Example"
	method := "Login"
	regex := ".*"

	tests := []struct {
		name  string
		rules []gatewayv1a2.GRPCRouteRule
		errs  field.ErrorList
	}{
		{
			name: "valid GRPCRoute with 1 service in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &service,
							},
						},
					},
				},
			},
		},
		{
			name: "valid GRPCRoute with 1 method in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Method: &method,
							},
						},
					},
				},
			},
		},
		{
			name: "invalid GRPCRoute missing service or method in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: nil,
								Method:  nil,
							},
						},
					},
				},
			},
			errs: field.ErrorList{
				{
					Type:   field.ErrorTypeRequired,
					Field:  "spec.rules[0].matches[0].method",
					Detail: "one or both of `service` or `method` must be specified",
				},
			},
		},
		{
			name: "GRPCRoute use regex in service and method with undefined match type",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &regex,
								Method:  &regex,
							},
						},
					},
				},
			},
			errs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					BadValue: regex,
					Field:    "spec.rules[0].matches[0].method",
					Detail:   `must only contain valid characters (matching ^(?i)\.?[a-z_][a-z_0-9]*(\.[a-z_][a-z_0-9]*)*$)`,
				},
				{
					Type:     field.ErrorTypeInvalid,
					BadValue: regex,
					Field:    "spec.rules[0].matches[0].method",
					Detail:   `must only contain valid characters (matching ^[A-Za-z_][A-Za-z_0-9]*$)`,
				},
			},
		},
		{
			name: "GRPCRoute use regex in service and method with match type Exact",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &regex,
								Method:  &regex,
								Type:    ptrTo(gatewayv1a2.GRPCMethodMatchExact),
							},
						},
					},
				},
			},
			errs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					BadValue: regex,
					Field:    "spec.rules[0].matches[0].method",
					Detail:   `must only contain valid characters (matching ^(?i)\.?[a-z_][a-z_0-9]*(\.[a-z_][a-z_0-9]*)*$)`,
				},
				{
					Type:     field.ErrorTypeInvalid,
					BadValue: regex,
					Field:    "spec.rules[0].matches[0].method",
					Detail:   `must only contain valid characters (matching ^[A-Za-z_][A-Za-z_0-9]*$)`,
				},
			},
		},
		{
			name: "GRPCRoute use regex in service and method with match type RegularExpression",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &regex,
								Method:  &regex,
								Type:    ptrTo(gatewayv1a2.GRPCMethodMatchRegularExpression),
							},
						},
					},
				},
			},
			errs: field.ErrorList{},
		},
		{
			name: "GRPCRoute use valid service and method with undefined match type",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &service,
								Method:  &method,
							},
						},
					},
				},
			},
			errs: field.ErrorList{},
		},
		{
			name: "GRPCRoute use valid service and method with match type Exact",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &service,
								Method:  &method,
								Type:    ptrTo(gatewayv1a2.GRPCMethodMatchExact),
							},
						},
					},
				},
			},
			errs: field.ErrorList{},
		},
		{
			name: "GRPCRoute with duplicate ExtensionRef filters",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Filters: []gatewayv1a2.GRPCRouteFilter{{
						Type: "ExtensionRef",
						ExtensionRef: &gatewayv1a2.LocalObjectReference{
							Kind: "Example1",
						},
					}, {
						Type: "ExtensionRef",
						ExtensionRef: &gatewayv1a2.LocalObjectReference{
							Kind: "Example2",
						},
					}},
				},
			},
		},
		{
			name: "GRPCRoute with duplicate RequestMirror filters",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Filters: []gatewayv1a2.GRPCRouteFilter{{
						Type: "RequestMirror",
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: "Example1",
							},
						},
					}, {
						Type: "RequestMirror",
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: "Example2",
							},
						},
					}},
				},
			},
			errs: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					BadValue: "RequestMirror",
					Field:    "spec.rules[0].filters",
					Detail:   "cannot be used multiple times in the same rule",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			route := gatewayv1a2.GRPCRoute{Spec: gatewayv1a2.GRPCRouteSpec{Rules: tc.rules}}
			errs := ValidateGRPCRoute(&route)
			if len(errs) != len(tc.errs) {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), len(tc.errs), errs)
				t.FailNow()
			}
			for i := 0; i < len(errs); i++ {
				realErr := errs[i].Error()
				expectedErr := tc.errs[i].Error()
				if realErr != expectedErr {
					t.Errorf("expect error message: %s, but got: %s", expectedErr, realErr)
					t.FailNow()
				}
			}
		})
	}
}

func TestValidateGRPCBackendUniqueFilters(t *testing.T) {
	var testService gatewayv1a2.ObjectName = "testService"
	var specialService gatewayv1a2.ObjectName = "specialService"
	tests := []struct {
		name     string
		rules    []gatewayv1a2.GRPCRouteRule
		errCount int
	}{{
		name:     "valid grpcRoute Rules backendref filters",
		errCount: 0,
		rules: []gatewayv1a2.GRPCRouteRule{{
			BackendRefs: []gatewayv1a2.GRPCBackendRef{
				{
					BackendRef: gatewayv1a2.BackendRef{
						BackendObjectReference: gatewayv1a2.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1a2.PortNumber(8080)),
						},
						Weight: ptrTo(int32(100)),
					},
					Filters: []gatewayv1a2.GRPCRouteFilter{
						{
							Type: gatewayv1a2.GRPCRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1a2.PortNumber(8080)),
								},
							},
						},
					},
				},
			},
		}},
	}, {
		name:     "invalid grpcRoute Rules duplicate mirror filter",
		errCount: 1,
		rules: []gatewayv1a2.GRPCRouteRule{{
			BackendRefs: []gatewayv1a2.GRPCBackendRef{
				{
					BackendRef: gatewayv1a2.BackendRef{
						BackendObjectReference: gatewayv1a2.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1a2.PortNumber(8080)),
						},
					},
					Filters: []gatewayv1a2.GRPCRouteFilter{
						{
							Type: gatewayv1a2.GRPCRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1a2.PortNumber(8080)),
								},
							},
						},
						{
							Type: gatewayv1a2.GRPCRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: specialService,
									Port: ptrTo(gatewayv1a2.PortNumber(8080)),
								},
							},
						},
					},
				},
			},
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1a2.GRPCRoute{Spec: gatewayv1a2.GRPCRouteSpec{Rules: tc.rules}}
			errs := ValidateGRPCRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateGRPCHeaderMatches(t *testing.T) {
	tests := []struct {
		name          string
		headerMatches []gatewayv1a2.GRPCHeaderMatch
		expectErr     string
	}{{
		name:          "no header matches",
		headerMatches: nil,
		expectErr:     "",
	}, {
		name: "no header matched more than once",
		headerMatches: []gatewayv1a2.GRPCHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "Header-Name-3", Value: "val-3"},
		},
		expectErr: "",
	}, {
		name: "header matched more than once (same case)",
		headerMatches: []gatewayv1a2.GRPCHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "Header-Name-1", Value: "val-3"},
		},
		expectErr: "spec.rules[0].matches[0].headers: Invalid value: \"Header-Name-1\": cannot match the same header multiple times in the same rule",
	}, {
		name: "header matched more than once (different case)",
		headerMatches: []gatewayv1a2.GRPCHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "HEADER-NAME-2", Value: "val-3"},
		},
		expectErr: "spec.rules[0].matches[0].headers: Invalid value: \"Header-Name-2\": cannot match the same header multiple times in the same rule",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1a2.GRPCRoute{Spec: gatewayv1a2.GRPCRouteSpec{
				Rules: []gatewayv1a2.GRPCRouteRule{{
					Matches: []gatewayv1a2.GRPCRouteMatch{{
						Headers: tc.headerMatches,
					}},
					BackendRefs: []gatewayv1a2.GRPCBackendRef{{
						BackendRef: gatewayv1a2.BackendRef{
							BackendObjectReference: gatewayv1a2.BackendObjectReference{
								Name: gatewayv1a2.ObjectName("test"),
								Port: ptrTo(gatewayv1a2.PortNumber(8080)),
							},
						},
					}},
				}},
			}}

			errs := ValidateGRPCRoute(&route)
			if len(tc.expectErr) == 0 {
				assert.Emptyf(t, errs, "expected no errors, got %d errors: %s", len(errs), errs)
			} else {
				require.Lenf(t, errs, 1, "expected one error, got %d errors: %s", len(errs), errs)
				assert.Equal(t, tc.expectErr, errs[0].Error())
			}
		})
	}
}
