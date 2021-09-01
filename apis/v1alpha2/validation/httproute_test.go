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

package validation

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"
	utilpointer "k8s.io/utils/pointer"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	pkgutils "sigs.k8s.io/gateway-api/pkg/util"
)

func TestValidateHTTPRoute(t *testing.T) {
	testService := "test-service"
	specialService := "special-service"
	tests := []struct {
		name        string
		hRouteRules []gatewayv1a2.HTTPRouteRule
		errCount    int
	}{
		{
			name: "valid httpRoute with no filters",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					BackendRefs: []gatewayv1a2.HTTPBackendRef{
						{
							BackendRef: gatewayv1a2.BackendRef{
								BackendObjectReference: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
								Weight: utilpointer.Int32(100),
							},
						},
					},
				},
			},
			errCount: 0,
		},
		{
			name: "valid httpRoute with 1 filter",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8081),
								},
							},
						},
					},
				},
			},
			errCount: 0,
		},
		{
			name: "invalid httpRoute with 2 extended filters",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: specialService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
					},
				},
			},
			errCount: 1,
		},
		{
			name: "invalid httpRoute with mix of filters and one duplicate",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
								Set: []gatewayv1a2.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
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
			errCount: 1,
		},
		{
			name: "invalid httpRoute with multiple duplicate filters",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
								Set: []gatewayv1a2.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
								Add: []gatewayv1a2.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: specialService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
					},
				},
			},
			errCount: 2,
		},
		{
			name: "valid httpRoute with duplicate ExtensionRef filters",
			hRouteRules: []gatewayv1a2.HTTPRouteRule{
				{
					Matches: []gatewayv1a2.HTTPRouteMatch{
						{
							Path: &gatewayv1a2.HTTPPathMatch{
								Type:  pkgutils.PathMatchTypePtr("Prefix"),
								Value: utilpointer.String("/"),
							},
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
								Set: []gatewayv1a2.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: &gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: "ExtensionRef",
						},
						{
							Type: "ExtensionRef",
						},
						{
							Type: "ExtensionRef",
						},
					},
				},
			},
			errCount: 0,
		},
	}
	for _, tt := range tests {
		// copy variable to avoid scope problems with ranges
		t.Run(tt.name, func(t *testing.T) {
			errs := validateHTTPRouteUniqueFilters(tt.hRouteRules, field.NewPath("spec").Child("rules"))
			if len(errs) != tt.errCount {
				t.Errorf("ValidateHTTPRoute() got %v errors, want %v errors", len(errs), tt.errCount)
			}
		})
	}
}

func TestValidateHTTPBackendUniqueFilters(t *testing.T) {
	var testService = "testService"
	var specialService = "specialService"
	tests := []struct {
		name     string
		hRoute   gatewayv1a2.HTTPRoute
		errCount int
	}{
		{
			name: "valid httpRoute Rules backendref filters",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							BackendRefs: []gatewayv1a2.HTTPBackendRef{
								{
									BackendRef: gatewayv1a2.BackendRef{
										BackendObjectReference: gatewayv1a2.BackendObjectReference{
											Name: testService,
											Port: pkgutils.PortNumberPtr(8080),
										},
										Weight: utilpointer.Int32(100),
									},
									Filters: []gatewayv1a2.HTTPRouteFilter{
										{
											Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
											RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
												BackendRef: &gatewayv1a2.BackendObjectReference{
													Name: testService,
													Port: pkgutils.PortNumberPtr(8081),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			errCount: 0,
		},

		{
			name: "invalid httpRoute Rules backendref filters",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							BackendRefs: []gatewayv1a2.HTTPBackendRef{
								{
									Filters: []gatewayv1a2.HTTPRouteFilter{
										{
											Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
											RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
												BackendRef: &gatewayv1a2.BackendObjectReference{
													Name: testService,
													Port: pkgutils.PortNumberPtr(8080),
												},
											},
										},
										{
											Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
											RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
												BackendRef: &gatewayv1a2.BackendObjectReference{
													Name: specialService,
													Port: pkgutils.PortNumberPtr(8080),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			errCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for index, rule := range tt.hRoute.Spec.Rules {
				errs := validateHTTPBackendUniqueFilters(rule.BackendRefs, field.NewPath("spec").Child("rules"), index)
				if len(errs) != tt.errCount {
					t.Errorf("ValidateHTTPRoute() got %v errors, want %v errors", len(errs), tt.errCount)
				}
			}
		})
	}
}
