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

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	utilpointer "k8s.io/utils/pointer"
)

func TestValidateHTTPRoute(t *testing.T) {
	testService := "test-service"
	specialService := "special-service"
	tests := []struct {
		name     string
		hRoute   gatewayv1a2.HTTPRoute
		errCount int
	}{
		{
			name: "valid httpRoute with no filters",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							ForwardTo: []gatewayv1a2.HTTPRouteForwardTo{
								{
									ServiceName: &testService,
									Port:        portNumberPtr(8080),
									Weight:      utilpointer.Int32(100),
								},
							},
						},
					},
				},
			},
			errCount: 0,
		},
		{
			name: "valid httpRoute with 1 filter",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a2.HTTPRouteFilter{
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8081),
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
			name: "invalid httpRoute with 2 extended filters",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a2.HTTPRouteFilter{
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &specialService,
										Port:        portNumberPtr(8080),
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
			name: "invalid httpRoute with mix of filters and one duplicate",
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a2.HTTPRouteFilter{
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
										Add: map[string]string{"my-header": "bar"},
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
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a2.HTTPRouteFilter{
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
										Add: map[string]string{"my-header": "bar"},
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &specialService,
										Port:        portNumberPtr(8080),
									},
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
			hRoute: gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{
						{
							Matches: []gatewayv1a2.HTTPRouteMatch{
								{
									Path: &gatewayv1a2.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a2.HTTPRouteFilter{
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
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
				},
			},
			errCount: 0,
		},
	}
	for _, tt := range tests {
		// copy variable to avoid scope problems with ranges
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateHTTPRoute(&tt.hRoute)
			if len(errs) != tt.errCount {
				t.Errorf("ValidateHTTPRoute() got %v errors, want %v errors", len(errs), tt.errCount)
			}
		})
	}
}

func pathMatchTypePtr(s string) *gatewayv1a2.PathMatchType {
	result := gatewayv1a2.PathMatchType(s)
	return &result
}

func portNumberPtr(p int) *gatewayv1a2.PortNumber {
	result := gatewayv1a2.PortNumber(p)
	return &result
}
