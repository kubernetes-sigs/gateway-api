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

func TestHTTPRouteParentRefExperimental(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		parentRefs []gatewayv1.ParentReference
	}{
		{
			name:       "invalid because duplicate parent refs without port or section name",
			wantErrors: []string{"sectionName or port must be unique when parentRefs includes 2 or more references to the same parent"},
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
			name:       "invalid because duplicate parent refs with only one port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(80)),
			}},
		},
		{
			name:       "invalid because duplicate parent refs with only one sectionName and port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
				Port:        ptrTo(gatewayv1.PortNumber(80)),
			}},
		},
		{
			name:       "invalid because duplicate parent refs with duplicate ports",
			wantErrors: []string{"sectionName or port must be unique when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(80)),
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(80)),
			}},
		},
		{
			name:       "valid single parentRef without sectionName or port",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
			}},
		},
		{
			name:       "valid single parentRef with sectionName and port",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
				Port:        ptrTo(gatewayv1.PortNumber(443)),
			}},
		},
		{
			name:       "valid because duplicate parent refs with different ports",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(80)),
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(443)),
			}},
		},
		{
			name:       "invalid ParentRefs with multiple mixed references to the same parent",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Kind:  ptrTo(gatewayv1.Kind("Gateway")),
				Group: ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:  "example",
				Port:  ptrTo(gatewayv1.PortNumber(443)),
			}},
		},
		{
			name:       "valid ParentRefs with multiple same port references to different section of a parent",
			wantErrors: []string{},
			parentRefs: []gatewayv1.ParentReference{{
				Name:        "example",
				Port:        ptrTo(gatewayv1.PortNumber(443)),
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Name:        "example",
				Port:        ptrTo(gatewayv1.PortNumber(443)),
				SectionName: ptrTo(gatewayv1.SectionName("bar")),
			}},
		},
		{
			// when referencing the same object, both parentRefs need to specify
			// the same optional fields (both parentRefs must specify port,
			// sectionName, or both)
			name:       "invalid because duplicate parent refs with first having sectionName and second having both sectionName and port",
			wantErrors: []string{"sectionName or port must be specified when parentRefs includes 2 or more references to the same parent"},
			parentRefs: []gatewayv1.ParentReference{{
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				Kind:        ptrTo(gatewayv1.Kind("Gateway")),
				Group:       ptrTo(gatewayv1.Group("gateway.networking.k8s.io")),
				Name:        "example",
				Port:        ptrTo(gatewayv1.PortNumber(443)),
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

func toDuration(durationString string) *gatewayv1.Duration {
	return (*gatewayv1.Duration)(&durationString)
}

func TestHTTPRouteCORS(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		corsfilter *gatewayv1.HTTPCORSFilter
	}{
		{
			name:       "Valid cors should be accepted",
			wantErrors: nil,
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"https://xpto.com",
					"http://*.abcd.com",
					"http://*.abcd.com:12345",
				},
			},
		},
		{
			name:       "Using wildcard only is accepted",
			wantErrors: nil,
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"*",
				},
			},
		},
		{
			name:       "Wildcard and other hosts on the same origin list should be denied",
			wantErrors: []string{"AllowOrigins cannot contain '*' alongside other origins"},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"*",
					"https://xpto.com",
				},
			},
		},
		{
			name:       "An origin without the format scheme://host should be denied",
			wantErrors: []string{"Invalid value: \"xpto.com\": spec.rules[0].filters[0].cors.allowOrigins[1] in body should match"},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"https://xpto.com",
					"xpto.com",
				},
			},
		},
		{
			name:       "An origin as http://*.com should be accepted",
			wantErrors: nil,
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"https://*.com",
				},
			},
		},
		{
			name:       "An origin with an invalid port should be denied",
			wantErrors: []string{"Invalid value: \"https://xpto.com:notaport\": spec.rules[0].filters[0].cors.allowOrigins[0] in body should match"},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"https://xpto.com:notaport",
				},
			},
		},
		{
			name:       "An origin with an value before the scheme definition should be denied",
			wantErrors: []string{"Invalid value: \"xpto/https://xpto.com\""},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowOrigins: []gatewayv1.CORSOrigin{
					"xpto/https://xpto.com",
				},
			},
		},
		{
			name:       "Using an invalid HTTP method should be denied",
			wantErrors: []string{"Unsupported value: \"BAZINGA\""},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowMethods: []gatewayv1.HTTPMethodWithWildcard{
					"BAZINGA",
				},
			},
		},
		{
			name:       "Using wildcard and a valid method should be denied",
			wantErrors: []string{"AllowMethods cannot contain '*' alongside other methods"},
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowMethods: []gatewayv1.HTTPMethodWithWildcard{
					"GET",
					"*",
					"POST",
				},
			},
		},
		{
			name:       "Using an array of valid methods should be accepted",
			wantErrors: nil,
			corsfilter: &gatewayv1.HTTPCORSFilter{
				AllowMethods: []gatewayv1.HTTPMethodWithWildcard{
					"GET",
					"OPTIONS",
					"POST",
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.HTTPRouteSpec{Rules: []gatewayv1.HTTPRouteRule{
					{
						Filters: []gatewayv1.HTTPRouteFilter{
							{
								Type: gatewayv1.HTTPRouteFilterCORS,
								CORS: tc.corsfilter,
							},
						},
					},
				}},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRouteTimeouts(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name:       "invalid timeout unit us is not supported",
			wantErrors: []string{"Invalid value: \"100us\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("100us"),
					},
				},
			},
		},
		{
			name:       "invalid timeout unit ns is not supported",
			wantErrors: []string{"Invalid value: \"500ns\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("500ns"),
					},
				},
			},
		},
		{
			name: "valid timeout request and backendRequest",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("4s"),
						BackendRequest: toDuration("2s"),
					},
				},
			},
		},
		{
			name: "valid timeout request",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("0s"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request day unit not supported",
			wantErrors: []string{"Invalid value: \"1d\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("1d"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request decimal not supported ",
			wantErrors: []string{"Invalid value: \"0.5s\": spec.rules[0].timeouts.request in body should match '^([0-9]{1,5}(h|m|s|ms)){1,4}$'"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("0.5s"),
					},
				},
			},
		},
		{
			name: "valid timeout request infinite greater than backendRequest 1ms",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("0s"),
						BackendRequest: toDuration("1ms"),
					},
				},
			},
		},
		{
			name: "valid timeout request 1s greater than backendRequest 200ms",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("1s"),
						BackendRequest: toDuration("200ms"),
					},
				},
			},
		},
		{
			name: "valid timeout request 10s equal backendRequest 10s",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("10s"),
						BackendRequest: toDuration("10s"),
					},
				},
			},
		},
		{
			name:       "invalid timeout request 200ms less than backendRequest 1s",
			wantErrors: []string{"Invalid value: \"object\": backendRequest timeout cannot be longer than request timeout"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("200ms"),
						BackendRequest: toDuration("1s"),
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRouteRuleExperimental(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name:       "invalid because multiple names are repeated",
			wantErrors: []string{"Rule name must be unique within the route"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Name: ptrTo(gatewayv1.SectionName("name1")),
				},
				{
					Name: ptrTo(gatewayv1.SectionName("name1")),
				},
			},
		},
		{
			name:       "invalid because multiple names are repeated with others",
			wantErrors: []string{"Rule name must be unique within the route"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Name: ptrTo(gatewayv1.SectionName("name1")),
				},
				{
					Name: ptrTo(gatewayv1.SectionName("not-name1")),
				},
				{
					Name: ptrTo(gatewayv1.SectionName("name1")),
				},
			},
		},
		{
			name:       "valid because names are unique",
			wantErrors: nil,
			rules: []gatewayv1.HTTPRouteRule{
				// Ok to have multiple nil
				{Name: nil},
				{Name: nil},
				{Name: ptrTo(gatewayv1.SectionName("name1"))},
				{Name: ptrTo(gatewayv1.SectionName("name2"))},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRequestMirrorFilterExperimental(t *testing.T) {
	var percent int32 = 42
	var denominator int32 = 1000
	var bad_denominator int32 = 0
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name:       "HTTPRoute - Invalid because both percent and fraction are specified",
			wantErrors: []string{"Only one of percent or fraction may be specified in HTTPRequestMirrorFilter"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Percent: &percent,
						Fraction: &gatewayv1.Fraction{
							Numerator:   83,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name:       "HTTPRoute - Invalid fraction - numerator greater than denominator",
			wantErrors: []string{"numerator must be less than or equal to denominator"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator:   1001,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name:       "HTTPRoute - Invalid fraction - denominator is 0",
			wantErrors: []string{"spec.rules[0].filters[0].requestMirror.fraction.denominator in body should be greater than or equal to 1"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator:   0,
							Denominator: &bad_denominator,
						},
					},
				}},
			}},
		},
		{
			name:       "HTTPRoute - Invalid fraction - numerator is negative",
			wantErrors: []string{"spec.rules[0].filters[0].requestMirror.fraction.numerator in body should be greater than or equal to 0"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator:   -1,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name: "HTTPRoute - Valid with percent",
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Percent: &percent,
					},
				}},
			}},
		},
		{
			name: "HTTPRoute - Valid with fraction",
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator:   83,
							Denominator: &denominator,
						},
					},
				}},
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
				Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPExternalAuthFilterExperimental(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name:       "HTTPRoute - Invalid because protocol is GRPC without GRPC config",
			wantErrors: []string{"grpc must be specified when protocol is set to 'GRPC'"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterExternalAuth,
					ExternalAuth: &gatewayv1.HTTPExternalAuthFilter{
						ExternalAuthProtocol: gatewayv1.HTTPRouteExternalAuthGRPCProtocol,
					},
				}},
			}},
		},
		{
			name:       "HTTPRoute - Invalid because protocol is HTTP without HTTP config",
			wantErrors: []string{"http must be specified when protocol is set to 'HTTP'"},
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterExternalAuth,
					ExternalAuth: &gatewayv1.HTTPExternalAuthFilter{
						ExternalAuthProtocol: gatewayv1.HTTPRouteExternalAuthHTTPProtocol,
					},
				}},
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
				Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}

}
