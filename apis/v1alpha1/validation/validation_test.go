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
	"reflect"
	"testing"

	gatewayv1a1 "sigs.k8s.io/gateway-api/apis/v1alpha1"

	"k8s.io/apimachinery/pkg/util/validation/field"
	utilpointer "k8s.io/utils/pointer"
)

func TestValidateHTTPRoute(t *testing.T) {
	testService := "test-service"
	specialService := "special-service"
	tests := []struct {
		name     string
		hRoute   gatewayv1a1.HTTPRoute
		errCount int
	}{
		{
			name: "valid httpRoute with no filters",
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							ForwardTo: []gatewayv1a1.HTTPRouteForwardTo{
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
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a1.HTTPRouteFilter{
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
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
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a1.HTTPRouteFilter{
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
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
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a1.HTTPRouteFilter{
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a1.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a1.HTTPRequestHeaderFilter{
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
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a1.HTTPRouteFilter{
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a1.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        portNumberPtr(8080),
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a1.HTTPRequestHeaderFilter{
										Add: map[string]string{"my-header": "bar"},
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
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
			hRoute: gatewayv1a1.HTTPRoute{
				Spec: gatewayv1a1.HTTPRouteSpec{
					Rules: []gatewayv1a1.HTTPRouteRule{
						{
							Matches: []gatewayv1a1.HTTPRouteMatch{
								{
									Path: &gatewayv1a1.HTTPPathMatch{
										Type:  pathMatchTypePtr("Prefix"),
										Value: utilpointer.String("/"),
									},
								},
							},
							Filters: []gatewayv1a1.HTTPRouteFilter{
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestHeaderModifier,
									RequestHeaderModifier: &gatewayv1a1.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: gatewayv1a1.HTTPRouteFilterRequestMirror,
									RequestMirror: &gatewayv1a1.HTTPRequestMirrorFilter{
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

func pathMatchTypePtr(s string) *gatewayv1a1.PathMatchType {
	result := gatewayv1a1.PathMatchType(s)
	return &result
}

func portNumberPtr(p int) *gatewayv1a1.PortNumber {
	result := gatewayv1a1.PortNumber(p)
	return &result
}

func TestValidateGatewayClassUpdate(t *testing.T) {
	type args struct {
		oldClass *gatewayv1a1.GatewayClass
		newClass *gatewayv1a1.GatewayClass
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "changing parameters reference is allowed",
			args: args{
				oldClass: &gatewayv1a1.GatewayClass{
					Spec: gatewayv1a1.GatewayClassSpec{
						Controller: "foo",
					},
				},
				newClass: &gatewayv1a1.GatewayClass{
					Spec: gatewayv1a1.GatewayClassSpec{
						Controller: "foo",
						ParametersRef: &gatewayv1a1.ParametersReference{
							Group: "example.com",
							Kind:  "GatewayClassConfig",
							Name:  "foo",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "changing controller field results in an error",
			args: args{
				oldClass: &gatewayv1a1.GatewayClass{
					Spec: gatewayv1a1.GatewayClassSpec{
						Controller: "foo",
					},
				},
				newClass: &gatewayv1a1.GatewayClass{
					Spec: gatewayv1a1.GatewayClassSpec{
						Controller: "bar",
					},
				},
			},
			want: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.controller",
					Detail:   "cannot update an immutable field",
					BadValue: "bar",
				},
			},
		},
		{
			name: "nil input result in no errors",
			args: args{
				oldClass: nil,
				newClass: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateGatewayClassUpdate(tt.args.oldClass, tt.args.newClass); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateGatewayClassUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
