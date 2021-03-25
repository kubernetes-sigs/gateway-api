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

package admission

import (
	"testing"

	v1alpha1 "sigs.k8s.io/gateway-api/apis/v1alpha1"
)

func TestSAValidator_ValidateHTTPRoute(t *testing.T) {
	testService := "test-service"
	specialService := "special-service"
	tests := []struct {
		name        string
		hRoute      v1alpha1.HTTPRoute
		wantOK      bool
		wantMessage string
		wantErr     bool
	}{
		{
			name: "valid httpRoute with no filters",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							ForwardTo: []v1alpha1.HTTPRouteForwardTo{
								{
									ServiceName: &testService,
									Port:        8080,
									Weight:      100,
								},
							},
						},
					},
				},
			},
			wantOK:      true,
			wantMessage: "",
			wantErr:     false,
		},
		{
			name: "valid httpRoute with 1 filter",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							Filters: []v1alpha1.HTTPRouteFilter{
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        8080,
									},
								},
							},
						},
					},
				},
			},
			wantOK:      true,
			wantMessage: "",
			wantErr:     false,
		},
		{
			name: "invalid httpRoute with 2 extended filters",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							Filters: []v1alpha1.HTTPRouteFilter{
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        8080,
									},
								},
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &specialService,
										Port:        8081,
									},
								},
							},
						},
					},
				},
			},
			wantOK:      false,
			wantMessage: "HTTPRules cannot contain more than one instance of each core or extended HTTPRouteFilterType",
			wantErr:     false,
		},
		{
			name: "valid httpRoute with 1 core and 1 extended filter",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							Filters: []v1alpha1.HTTPRouteFilter{
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        8080,
									},
								},
								{
									Type: "RequestHeaderModifier",
									RequestHeaderModifier: &v1alpha1.HTTPRequestHeaderFilter{
										Add: map[string]string{"my-header": "bar"},
									},
								},
							},
						},
					},
				},
			},
			wantOK:      true,
			wantMessage: "",
			wantErr:     false,
		},
		{
			name: "invalid httpRoute with mix of filters and one duplicate",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							Filters: []v1alpha1.HTTPRouteFilter{
								{
									Type: "RequestHeaderModifier",
									RequestHeaderModifier: &v1alpha1.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        8080,
									},
								},
								{
									Type: "RequestHeaderModifier",
									RequestHeaderModifier: &v1alpha1.HTTPRequestHeaderFilter{
										Add: map[string]string{"my-header": "bar"},
									},
								},
							},
						},
					},
				},
			},
			wantOK:      false,
			wantMessage: "HTTPRules cannot contain more than one instance of each core or extended HTTPRouteFilterType",
			wantErr:     false,
		},
		{
			name: "valid httpRoute with duplicate ExtensionRef filters",
			hRoute: v1alpha1.HTTPRoute{
				Spec: v1alpha1.HTTPRouteSpec{
					Rules: []v1alpha1.HTTPRouteRule{
						{
							Matches: []v1alpha1.HTTPRouteMatch{
								{
									Path: v1alpha1.HTTPPathMatch{
										Type:  "Prefix",
										Value: "/",
									},
								},
							},
							Filters: []v1alpha1.HTTPRouteFilter{
								{
									Type: "RequestHeaderModifier",
									RequestHeaderModifier: &v1alpha1.HTTPRequestHeaderFilter{
										Set: map[string]string{"special-header": "foo"},
									},
								},
								{
									Type: "RequestMirror",
									RequestMirror: &v1alpha1.HTTPRequestMirrorFilter{
										ServiceName: &testService,
										Port:        8080,
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
			wantOK:      true,
			wantMessage: "",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		// copy variable to avoid scope problems with ranges
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			ok, msg, err := ValidateHTTPRoute(tt.hRoute)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHTTPRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ok != tt.wantOK {
				t.Errorf("ValidateHTTPRoute() got = %v, want %v", ok, tt.wantOK)
			}
			if msg != tt.wantMessage {
				t.Errorf("ValidateHTTPRoute() msg = %v, want %v", msg, tt.wantMessage)
			}
		})
	}
}
