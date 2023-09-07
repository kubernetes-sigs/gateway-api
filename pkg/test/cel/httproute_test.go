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
	"strings"
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

func TestHTTPPathMatch(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		path       *gatewayv1b1.HTTPPathMatch
	}{
		{
			name:       "invalid because path does not start with '/'",
			wantErrors: []string{"value must be an absolute path and start with '/' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
				Value: ptrTo("foo"),
			},
		},
		{
			name:       "invalid httpRoute prefix (/.)",
			wantErrors: []string{"must not end with '/.' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
				Value: ptrTo("/."),
			},
		},
		{
			name:       "invalid exact (/./)",
			wantErrors: []string{"must not contain '/./' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("Exact")),
				Value: ptrTo("/foo/./bar"),
			},
		},
		{
			name:       "invalid type",
			wantErrors: []string{"type must be one of ['Exact', 'PathPrefix', 'RegularExpression']"},
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("FooBar")),
				Value: ptrTo("/path"),
			},
		},
		{
			name: "valid because type is RegularExpression but would not be valid for Exact",
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("RegularExpression")),
				Value: ptrTo("/foo/./bar"),
			},
		},
		{
			name: "valid httpRoute prefix",
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
				Value: ptrTo("/path"),
			},
		},
		{
			name: "valid path with some special characters",
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("Exact")),
				Value: ptrTo("/abc/123'/a-b-c/d@gmail/%0A"),
			},
		},
		{
			name: "invalid prefix path (/[])",
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
				Value: ptrTo("/[]"),
			},
			wantErrors: []string{"must only contain valid characters (matching ^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$) for types ['Exact', 'PathPrefix']"},
		},
		{
			name: "invalid exact path (/^)",
			path: &gatewayv1b1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1b1.PathMatchType("Exact")),
				Value: ptrTo("/^"),
			},
			wantErrors: []string{"must only contain valid characters (matching ^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$) for types ['Exact', 'PathPrefix']"},
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
					Rules: []gatewayv1b1.HTTPRouteRule{{
						Matches: []gatewayv1b1.HTTPRouteMatch{{
							Path: tc.path,
						}},
						BackendRefs: []gatewayv1b1.HTTPBackendRef{{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: gatewayv1b1.ObjectName("test"),
									Port: ptrTo(gatewayv1b1.PortNumber(8080)),
								},
							},
						}},
					}},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestBackendObjectReference(t *testing.T) {
	portPtr := func(n int) *gatewayv1b1.PortNumber {
		p := gatewayv1b1.PortNumber(n)
		return &p
	}

	groupPtr := func(g string) *gatewayv1b1.Group {
		p := gatewayv1b1.Group(g)
		return &p
	}

	kindPtr := func(k string) *gatewayv1b1.Kind {
		p := gatewayv1b1.Kind(k)
		return &p
	}

	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1b1.HTTPRouteRule
		backendRef gatewayv1b1.BackendObjectReference
	}{
		{
			name: "default groupkind with port",
			backendRef: gatewayv1b1.BackendObjectReference{
				Name: "backend",
				Port: portPtr(99),
			},
		},
		{
			name:       "default groupkind with no port",
			wantErrors: []string{"Must have port for Service reference"},
			backendRef: gatewayv1b1.BackendObjectReference{
				Name: "backend",
			},
		},
		{
			name: "explicit service with port",
			backendRef: gatewayv1b1.BackendObjectReference{
				Group: groupPtr(""),
				Kind:  kindPtr("Service"),
				Name:  "backend",
				Port:  portPtr(99),
			},
		},
		{
			name:       "explicit service with no port",
			wantErrors: []string{"Must have port for Service reference"},
			backendRef: gatewayv1b1.BackendObjectReference{
				Group: groupPtr(""),
				Kind:  kindPtr("Service"),
				Name:  "backend",
			},
		},
		{
			name: "explicit ref with no port",
			backendRef: gatewayv1b1.BackendObjectReference{
				Group: groupPtr("foo.example.com"),
				Kind:  kindPtr("Foo"),
				Name:  "backend",
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
				Spec: gatewayv1b1.HTTPRouteSpec{
					Rules: []gatewayv1b1.HTTPRouteRule{{
						BackendRefs: []gatewayv1b1.HTTPBackendRef{{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: tc.backendRef,
							},
						}},
					}},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRouteFilter(t *testing.T) {
	tests := []struct {
		name        string
		wantErrors  []string
		routeFilter gatewayv1b1.HTTPRouteFilter
	}{
		{
			name: "valid HTTPRouteFilterRequestHeaderModifier route filter",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
					Set:    []gatewayv1b1.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1b1.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type:          gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
				RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type"},
		},
		{
			name: "valid HTTPRouteFilterRequestMirror route filter",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestMirror,
				RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{BackendRef: gatewayv1b1.BackendObjectReference{
					Group:     ptrTo(gatewayv1b1.Group("group")),
					Kind:      ptrTo(gatewayv1b1.Kind("kind")),
					Name:      "name",
					Namespace: ptrTo(gatewayv1b1.Namespace("ns")),
					Port:      ptrTo(gatewayv1b1.PortNumber(22)),
				}},
			},
		},
		{
			name: "invalid HTTPRouteFilterRequestMirror type filter with non-matching field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type:                  gatewayv1b1.HTTPRouteFilterRequestMirror,
				RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be nil if the filter.type is not RequestHeaderModifier", "filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterRequestMirror type filter with empty value field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestMirror,
			},
			wantErrors: []string{"filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "valid HTTPRouteFilterRequestRedirect route filter",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
					Scheme:   ptrTo("http"),
					Hostname: ptrTo(gatewayv1b1.PreciseHostname("hostname")),
					Path: &gatewayv1b1.HTTPPathModifier{
						Type:            gatewayv1b1.FullPathHTTPPathModifier,
						ReplaceFullPath: ptrTo("path"),
					},
					Port:       ptrTo(gatewayv1b1.PortNumber(8080)),
					StatusCode: ptrTo(302),
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterRequestRedirect type filter with non-matching field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type:          gatewayv1b1.HTTPRouteFilterRequestRedirect,
				RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.requestRedirect must be specified for RequestRedirect filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterRequestRedirect type filter with empty value field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
			},
			wantErrors: []string{"filter.requestRedirect must be specified for RequestRedirect filter.type"},
		},
		{
			name: "valid HTTPRouteFilterExtensionRef filter",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1b1.LocalObjectReference{
					Group: "group",
					Kind:  "kind",
					Name:  "name",
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterExtensionRef type filter with non-matching field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type:          gatewayv1b1.HTTPRouteFilterExtensionRef,
				RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterExtensionRef type filter with empty value field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterExtensionRef,
			},
			wantErrors: []string{"filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "valid HTTPRouteFilterURLRewrite route filter",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
					Hostname: ptrTo(gatewayv1b1.PreciseHostname("hostname")),
					Path: &gatewayv1b1.HTTPPathModifier{
						Type:            gatewayv1b1.FullPathHTTPPathModifier,
						ReplaceFullPath: ptrTo("path"),
					},
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterURLRewrite type filter with non-matching field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type:          gatewayv1b1.HTTPRouteFilterURLRewrite,
				RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.urlRewrite must be specified for URLRewrite filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterURLRewrite type filter with empty value field",
			routeFilter: gatewayv1b1.HTTPRouteFilter{
				Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
			},
			wantErrors: []string{"filter.urlRewrite must be specified for URLRewrite filter.type"},
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
					Rules: []gatewayv1b1.HTTPRouteRule{{
						Filters: []gatewayv1b1.HTTPRouteFilter{tc.routeFilter},
					}},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRouteRule(t *testing.T) {
	testService := gatewayv1b1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1b1.HTTPRouteRule
	}{
		{
			name: "valid httpRoute with no filters",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					BackendRefs: []gatewayv1b1.HTTPBackendRef{
						{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(8080)),
								},
								Weight: ptrTo(int32(100)),
							},
						},
					},
				},
			},
		},
		{
			name: "valid httpRoute with 1 filter",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(8081)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid httpRoute with duplicate ExtensionRef filters",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Set: []gatewayv1b1.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(8080)),
								},
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1b1.LocalObjectReference{
								Kind: "Service",
								Name: "test",
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1b1.LocalObjectReference{
								Kind: "Service",
								Name: "test",
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1b1.LocalObjectReference{
								Kind: "Service",
								Name: "test",
							},
						},
					},
				},
			},
		},
		{
			name: "valid redirect path modifier",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:            gatewayv1b1.FullPathHTTPPathModifier,
									ReplaceFullPath: ptrTo("foo"),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid rewrite path modifier",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{{
					Path: &gatewayv1b1.HTTPPathMatch{
						Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
						Value: ptrTo("/bar"),
					},
				}},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "multiple actions for different request headers",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
					RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
						Add: []gatewayv1b1.HTTPHeader{
							{
								Name:  gatewayv1b1.HTTPHeaderName("x-vegetable"),
								Value: "carrot",
							},
							{
								Name:  gatewayv1b1.HTTPHeaderName("x-grain"),
								Value: "rye",
							},
						},
						Set: []gatewayv1b1.HTTPHeader{
							{
								Name:  gatewayv1b1.HTTPHeaderName("x-fruit"),
								Value: "watermelon",
							},
							{
								Name:  gatewayv1b1.HTTPHeaderName("x-spice"),
								Value: "coriander",
							},
						},
					},
				}},
			}},
		},
		{
			name: "multiple actions for different response headers",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterResponseHeaderModifier,
					ResponseHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
						Add: []gatewayv1b1.HTTPHeader{{
							Name:  gatewayv1b1.HTTPHeaderName("x-example"),
							Value: "blueberry",
						}},
						Set: []gatewayv1b1.HTTPHeader{{
							Name:  gatewayv1b1.HTTPHeaderName("x-different"),
							Value: "turnip",
						}},
					},
				}},
			}},
		},
		{
			name:       "backendref with request redirect httpRoute filter",
			wantErrors: []string{"RequestRedirect filter must not be used together with backendRefs"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Scheme:     ptrTo("https"),
								StatusCode: ptrTo(301),
							},
						},
					},
					BackendRefs: []gatewayv1b1.HTTPBackendRef{
						{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "request redirect without backendref in httpRoute filter",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Scheme:     ptrTo("https"),
								StatusCode: ptrTo(301),
							},
						},
					},
				},
			},
		},
		{
			name: "backendref without request redirect filter",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Set: []gatewayv1b1.HTTPHeader{{Name: "name", Value: "foo"}},
							},
						},
					},
					BackendRefs: []gatewayv1b1.HTTPBackendRef{
						{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "backendref without any filter",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					BackendRefs: []gatewayv1b1.HTTPBackendRef{
						{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid use of URLRewrite filter",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid URLRewrite filter because too many path matches",
			wantErrors: []string{"When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid URLRewrite filter because too many path matches",
			wantErrors: []string{"When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchType(gatewayv1b1.FullPathHTTPPathModifier)), // Incorrect Patch match Type for URLRewrite filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "valid use of RequestRedirect filter",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid RequestRedirect filter because too many path matches",
			wantErrors: []string{"When using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid RequestRedirect filter because path match has type ReplaceFullPath",
			wantErrors: []string{"When using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchType(gatewayv1b1.FullPathHTTPPathModifier)), // Incorrect Patch match Type for RequestRedirect filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "valid use of URLRewrite filter (within backendRefs)",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid URLRewrite filter (within backendRefs) because too many path matches",
			wantErrors: []string{"Within backendRefs, When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid URLRewrite filter (within backendRefs) because path match has type ReplaceFullPath",
			wantErrors: []string{"Within backendRefs, When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchType(gatewayv1b1.FullPathHTTPPathModifier)), // Incorrect Patch match Type for URLRewrite filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name: "valid use of RequestRedirect filter (within backendRefs)",
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid RequestRedirect filter (within backendRefs) because too many path matches",
			wantErrors: []string{"Within backendRefs, when using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid RequestRedirect filter (within backendRefs) because path match has type ReplaceFullPath",
			wantErrors: []string{"Within backendRefs, when using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1b1.HTTPRouteRule{{
				Matches: []gatewayv1b1.HTTPRouteMatch{
					{
						Path: &gatewayv1b1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1b1.PathMatchType(gatewayv1b1.FullPathHTTPPathModifier)), // Incorrect Patch match Type for RequestRedirect filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1b1.HTTPBackendRef{
					{
						BackendRef: gatewayv1b1.BackendRef{
							BackendObjectReference: gatewayv1b1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1b1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1b1.HTTPRouteFilter{{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "rewrite and redirect filters combined (invalid)",
			wantErrors: []string{"May specify either httpRouteFilterRequestRedirect or httpRouteFilterRequestRewrite, but not both"}, // errCount: 3,
			rules: []gatewayv1b1.HTTPRouteRule{{
				Filters: []gatewayv1b1.HTTPRouteFilter{{
					Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}, {
					Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
						Path: &gatewayv1b1.HTTPPathModifier{
							Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid because repeated URLRewrite filter",
			wantErrors: []string{"URLRewrite filter cannot be repeated"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("bar"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "invalid because repeated RequestHeaderModifier filter among mix of filters",
			wantErrors: []string{"RequestHeaderModifier filter cannot be repeated"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Set: []gatewayv1b1.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1b1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(8080)),
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Add: []gatewayv1b1.HTTPHeader{
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
		{
			name:       "invalid because multiple filters are repeated",
			wantErrors: []string{"ResponseHeaderModifier filter cannot be repeated", "RequestRedirect filter cannot be repeated"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1b1.HTTPRouteFilter{
						{
							Type: gatewayv1b1.HTTPRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Set: []gatewayv1b1.HTTPHeader{
									{
										Name:  "special-header",
										Value: "foo",
									},
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1b1.HTTPHeaderFilter{
								Add: []gatewayv1b1.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:            gatewayv1b1.FullPathHTTPPathModifier,
									ReplaceFullPath: ptrTo("foo"),
								},
							},
						},
						{
							Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
								Path: &gatewayv1b1.HTTPPathModifier{
									Type:            gatewayv1b1.FullPathHTTPPathModifier,
									ReplaceFullPath: ptrTo("bar"),
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

func TestHTTPBackendRef(t *testing.T) {
	testService := gatewayv1b1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1b1.HTTPRouteRule
	}{
		{
			name:       "invalid because repeated URLRewrite filter within backendRefs",
			wantErrors: []string{"URLRewrite filter cannot be repeated"},
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Matches: []gatewayv1b1.HTTPRouteMatch{
						{
							Path: &gatewayv1b1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1b1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					BackendRefs: []gatewayv1b1.HTTPBackendRef{
						{
							BackendRef: gatewayv1b1.BackendRef{
								BackendObjectReference: gatewayv1b1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1b1.PortNumber(80)),
								},
							},
							Filters: []gatewayv1b1.HTTPRouteFilter{
								{
									Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
									URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
										Path: &gatewayv1b1.HTTPPathModifier{
											Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
											ReplacePrefixMatch: ptrTo("foo"),
										},
									},
								},
								{
									Type: gatewayv1b1.HTTPRouteFilterURLRewrite,
									URLRewrite: &gatewayv1b1.HTTPURLRewriteFilter{
										Path: &gatewayv1b1.HTTPPathModifier{
											Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
											ReplacePrefixMatch: ptrTo("bar"),
										},
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

func TestHTTPPathModifier(t *testing.T) {
	tests := []struct {
		name         string
		wantErrors   []string
		pathModifier gatewayv1b1.HTTPPathModifier
	}{
		{
			name: "valid ReplaceFullPath",
			pathModifier: gatewayv1b1.HTTPPathModifier{
				Type:            gatewayv1b1.FullPathHTTPPathModifier,
				ReplaceFullPath: ptrTo("foo"),
			},
		},
		{
			name:       "replaceFullPath must be specified when type is set to 'ReplaceFullPath'",
			wantErrors: []string{"replaceFullPath must be specified when type is set to 'ReplaceFullPath'"},
			pathModifier: gatewayv1b1.HTTPPathModifier{
				Type: gatewayv1b1.FullPathHTTPPathModifier,
			},
		},
		{
			name:       "type must be 'ReplaceFullPath' when replaceFullPath is set",
			wantErrors: []string{"type must be 'ReplaceFullPath' when replaceFullPath is set"},
			pathModifier: gatewayv1b1.HTTPPathModifier{
				ReplaceFullPath: ptrTo("foo"),
			},
		},
		{
			name: "valid ReplacePrefixMatch",
			pathModifier: gatewayv1b1.HTTPPathModifier{
				Type:               gatewayv1b1.PrefixMatchHTTPPathModifier,
				ReplacePrefixMatch: ptrTo("/foo"),
			},
		},
		{
			name:       "replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'",
			wantErrors: []string{"replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'"},
			pathModifier: gatewayv1b1.HTTPPathModifier{
				Type: gatewayv1b1.PrefixMatchHTTPPathModifier,
			},
		},
		{
			name:       "type must be 'ReplacePrefixMatch' when replacePrefixMatch is set",
			wantErrors: []string{"type must be 'ReplacePrefixMatch' when replacePrefixMatch is set"},
			pathModifier: gatewayv1b1.HTTPPathModifier{
				ReplacePrefixMatch: ptrTo("/foo"),
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
				Spec: gatewayv1b1.HTTPRouteSpec{
					Rules: []gatewayv1b1.HTTPRouteRule{
						{
							Filters: []gatewayv1b1.HTTPRouteFilter{
								{
									Type: gatewayv1b1.HTTPRouteFilterRequestRedirect,
									RequestRedirect: &gatewayv1b1.HTTPRequestRedirectFilter{
										Path: &tc.pathModifier,
									},
								},
							},
						},
					},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func validateHTTPRoute(t *testing.T, route *gatewayv1b1.HTTPRoute, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating HTTPRoute %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating HTTPRoute %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
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
			name: "valid timeout request infinite greater than backend request 1ms",
			rules: []gatewayv1b1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1b1.HTTPRouteTimeouts{
						Request:        toDuration("0s"),
						BackendRequest: toDuration("1ms"),
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
