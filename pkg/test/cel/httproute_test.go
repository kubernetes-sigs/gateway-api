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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
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
		path       *gatewayv1.HTTPPathMatch
	}{
		{
			name:       "invalid because path does not start with '/'",
			wantErrors: []string{"value must be an absolute path and start with '/' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
				Value: ptrTo("foo"),
			},
		},
		{
			name:       "invalid httpRoute prefix (/.)",
			wantErrors: []string{"must not end with '/.' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
				Value: ptrTo("/."),
			},
		},
		{
			name:       "invalid exact (/./)",
			wantErrors: []string{"must not contain '/./' when type one of ['Exact', 'PathPrefix']"},
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("Exact")),
				Value: ptrTo("/foo/./bar"),
			},
		},
		{
			name:       "invalid type",
			wantErrors: []string{"must be one of ['Exact', 'PathPrefix', 'RegularExpression']"},
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("FooBar")),
				Value: ptrTo("/path"),
			},
		},
		{
			name: "valid because type is RegularExpression but would not be valid for Exact",
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("RegularExpression")),
				Value: ptrTo("/foo/./bar"),
			},
		},
		{
			name: "valid httpRoute prefix",
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
				Value: ptrTo("/path"),
			},
		},
		{
			name: "valid path with some special characters",
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("Exact")),
				Value: ptrTo("/abc/123'/a-b-c/d@gmail/%0A"),
			},
		},
		{
			name: "invalid prefix path (/[])",
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
				Value: ptrTo("/[]"),
			},
			wantErrors: []string{"must only contain valid characters (matching ^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$) for types ['Exact', 'PathPrefix']"},
		},
		{
			name: "invalid exact path (/^)",
			path: &gatewayv1.HTTPPathMatch{
				Type:  ptrTo(gatewayv1.PathMatchType("Exact")),
				Value: ptrTo("/^"),
			},
			wantErrors: []string{"must only contain valid characters (matching ^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$) for types ['Exact', 'PathPrefix']"},
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
					Rules: []gatewayv1.HTTPRouteRule{{
						Matches: []gatewayv1.HTTPRouteMatch{{
							Path: tc.path,
						}},
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: gatewayv1.ObjectName("test"),
									Port: ptrTo(gatewayv1.PortNumber(8080)),
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
	portPtr := func(n int) *gatewayv1.PortNumber {
		//nolint:gosec
		p := gatewayv1.PortNumber(n)
		return &p
	}

	groupPtr := func(g string) *gatewayv1.Group {
		p := gatewayv1.Group(g)
		return &p
	}

	kindPtr := func(k string) *gatewayv1.Kind {
		p := gatewayv1.Kind(k)
		return &p
	}

	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
		backendRef gatewayv1.BackendObjectReference
	}{
		{
			name: "default groupkind with port",
			backendRef: gatewayv1.BackendObjectReference{
				Name: "backend",
				Port: portPtr(99),
			},
		},
		{
			name:       "default groupkind with no port",
			wantErrors: []string{"Must have port for Service reference"},
			backendRef: gatewayv1.BackendObjectReference{
				Name: "backend",
			},
		},
		{
			name: "explicit service with port",
			backendRef: gatewayv1.BackendObjectReference{
				Group: groupPtr(""),
				Kind:  kindPtr("Service"),
				Name:  "backend",
				Port:  portPtr(99),
			},
		},
		{
			name:       "explicit service with no port",
			wantErrors: []string{"Must have port for Service reference"},
			backendRef: gatewayv1.BackendObjectReference{
				Group: groupPtr(""),
				Kind:  kindPtr("Service"),
				Name:  "backend",
			},
		},
		{
			name: "explicit ref with no port",
			backendRef: gatewayv1.BackendObjectReference{
				Group: groupPtr("foo.example.com"),
				Kind:  kindPtr("Foo"),
				Name:  "backend",
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
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{{
						BackendRefs: []gatewayv1.HTTPBackendRef{{
							BackendRef: gatewayv1.BackendRef{
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
		routeFilter gatewayv1.HTTPRouteFilter
	}{
		{
			name: "valid HTTPRouteFilterRequestHeaderModifier route filter",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set:    []gatewayv1.HTTPHeader{{Name: "name", Value: "foo"}},
					Add:    []gatewayv1.HTTPHeader{{Name: "add", Value: "foo"}},
					Remove: []string{"remove"},
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with non-matching field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type:          gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type", "filter.requestMirror must be nil if the filter.type is not RequestMirror"},
		},
		{
			name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with empty value field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
			},
			wantErrors: []string{"filter.requestHeaderModifier must be specified for RequestHeaderModifier filter.type"},
		},
		{
			name: "valid HTTPRouteFilterRequestMirror route filter",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestMirror,
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
			name: "invalid HTTPRouteFilterRequestMirror type filter with non-matching field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type:                  gatewayv1.HTTPRouteFilterRequestMirror,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{},
			},
			wantErrors: []string{"filter.requestHeaderModifier must be nil if the filter.type is not RequestHeaderModifier", "filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterRequestMirror type filter with empty value field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestMirror,
			},
			wantErrors: []string{"filter.requestMirror must be specified for RequestMirror filter.type"},
		},
		{
			name: "valid HTTPRouteFilterRequestRedirect route filter",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Scheme:   ptrTo("http"),
					Hostname: ptrTo(gatewayv1.PreciseHostname("hostname")),
					Path: &gatewayv1.HTTPPathModifier{
						Type:            gatewayv1.FullPathHTTPPathModifier,
						ReplaceFullPath: ptrTo("path"),
					},
					Port:       ptrTo(gatewayv1.PortNumber(8080)),
					StatusCode: ptrTo(302),
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterRequestRedirect type filter with non-matching field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type:          gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.requestRedirect must be specified for RequestRedirect filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterRequestRedirect type filter with empty value field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
			},
			wantErrors: []string{"filter.requestRedirect must be specified for RequestRedirect filter.type"},
		},
		{
			name: "valid HTTPRouteFilterExtensionRef filter",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gatewayv1.LocalObjectReference{
					Group: "group",
					Kind:  "kind",
					Name:  "name",
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterExtensionRef type filter with non-matching field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type:          gatewayv1.HTTPRouteFilterExtensionRef,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterExtensionRef type filter with empty value field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterExtensionRef,
			},
			wantErrors: []string{"filter.extensionRef must be specified for ExtensionRef filter.type"},
		},
		{
			name: "valid HTTPRouteFilterURLRewrite route filter",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Hostname: ptrTo(gatewayv1.PreciseHostname("hostname")),
					Path: &gatewayv1.HTTPPathModifier{
						Type:            gatewayv1.FullPathHTTPPathModifier,
						ReplaceFullPath: ptrTo("path"),
					},
				},
			},
		},
		{
			name: "invalid HTTPRouteFilterURLRewrite type filter with non-matching field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type:          gatewayv1.HTTPRouteFilterURLRewrite,
				RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
			},
			wantErrors: []string{"filter.requestMirror must be nil if the filter.type is not RequestMirror", "filter.urlRewrite must be specified for URLRewrite filter.type"},
		},
		{
			name: "invalid HTTPRouteFilterURLRewrite type filter with empty value field",
			routeFilter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
			},
			wantErrors: []string{"filter.urlRewrite must be specified for URLRewrite filter.type"},
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
					Rules: []gatewayv1.HTTPRouteRule{{
						Filters: []gatewayv1.HTTPRouteFilter{tc.routeFilter},
					}},
				},
			}
			validateHTTPRoute(t, route, tc.wantErrors)
		})
	}
}

func TestHTTPRouteRule(t *testing.T) {
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name: "valid httpRoute with no filters",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
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
			name: "valid httpRoute with 1 filter",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8081)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid httpRoute with duplicate ExtensionRef filters",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
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
							Type: gatewayv1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
								},
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1.LocalObjectReference{
								Kind: "Service",
								Name: "test",
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1.LocalObjectReference{
								Kind: "Service",
								Name: "test",
							},
						},
						{
							Type: "ExtensionRef",
							ExtensionRef: &gatewayv1.LocalObjectReference{
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
			rules: []gatewayv1.HTTPRouteRule{
				{
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:            gatewayv1.FullPathHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
						Value: ptrTo("/bar"),
					},
				}},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "multiple actions for different request headers",
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
					RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
						Add: []gatewayv1.HTTPHeader{
							{
								Name:  gatewayv1.HTTPHeaderName("x-vegetable"),
								Value: "carrot",
							},
							{
								Name:  gatewayv1.HTTPHeaderName("x-grain"),
								Value: "rye",
							},
						},
						Set: []gatewayv1.HTTPHeader{
							{
								Name:  gatewayv1.HTTPHeaderName("x-fruit"),
								Value: "watermelon",
							},
							{
								Name:  gatewayv1.HTTPHeaderName("x-spice"),
								Value: "coriander",
							},
						},
					},
				}},
			}},
		},
		{
			name: "multiple actions for different response headers",
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
					ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
						Add: []gatewayv1.HTTPHeader{{
							Name:  gatewayv1.HTTPHeaderName("x-example"),
							Value: "blueberry",
						}},
						Set: []gatewayv1.HTTPHeader{{
							Name:  gatewayv1.HTTPHeaderName("x-different"),
							Value: "turnip",
						}},
					},
				}},
			}},
		},
		{
			name:       "backendref with request redirect httpRoute filter",
			wantErrors: []string{"RequestRedirect filter must not be used together with backendRefs"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Scheme:     ptrTo("https"),
								StatusCode: ptrTo(301),
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "request redirect without backendref in httpRoute filter",
			rules: []gatewayv1.HTTPRouteRule{
				{
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
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
			rules: []gatewayv1.HTTPRouteRule{
				{
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Set: []gatewayv1.HTTPHeader{{Name: "name", Value: "foo"}},
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "backendref without any filter",
			rules: []gatewayv1.HTTPRouteRule{
				{
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(80)),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid use of URLRewrite filter",
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid URLRewrite filter because too many path matches",
			wantErrors: []string{"When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid URLRewrite filter because wrong path match type",
			wantErrors: []string{"When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchRegularExpression), // Incorrect Path match Type for URLRewrite filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "valid use of RequestRedirect filter",
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid RequestRedirect filter because too many path matches",
			wantErrors: []string{"When using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid RequestRedirect filter because path match has type RegularExpression",
			wantErrors: []string{"When using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchRegularExpression), // Incorrect Path match Type for RequestRedirect filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name: "valid use of URLRewrite filter (within backendRefs)",
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid URLRewrite filter (within backendRefs) because path match has type RegularExpression",
			wantErrors: []string{"Within backendRefs, When using URLRewrite filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchRegularExpression), // Incorrect Path match Type for URLRewrite filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/foo"),
						},
					},
					{ // Cannot have multiple path matches.
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchPathPrefix),
							Value: ptrTo("/bar"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						}},
					},
				},
			}},
		},
		{
			name:       "invalid RequestRedirect filter (within backendRefs) because path match has type RegularExpression",
			wantErrors: []string{"Within backendRefs, when using RequestRedirect filter with path.replacePrefixMatch, exactly one PathPrefix match must be specified"},
			rules: []gatewayv1.HTTPRouteRule{{
				Matches: []gatewayv1.HTTPRouteMatch{
					{
						Path: &gatewayv1.HTTPPathMatch{
							Type:  ptrTo(gatewayv1.PathMatchRegularExpression), // Incorrect Path match Type for RequestRedirect filter with ReplacePrefixMatch.
							Value: ptrTo("/foo"),
						},
					},
				},
				BackendRefs: []gatewayv1.HTTPBackendRef{
					{
						BackendRef: gatewayv1.BackendRef{
							BackendObjectReference: gatewayv1.BackendObjectReference{
								Name: testService,
								Port: ptrTo(gatewayv1.PortNumber(80)),
							},
						},
						Filters: []gatewayv1.HTTPRouteFilter{{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{{
				Filters: []gatewayv1.HTTPRouteFilter{{
					Type: gatewayv1.HTTPRouteFilterURLRewrite,
					URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}, {
					Type: gatewayv1.HTTPRouteFilterRequestRedirect,
					RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
						Path: &gatewayv1.HTTPPathModifier{
							Type:               gatewayv1.PrefixMatchHTTPPathModifier,
							ReplacePrefixMatch: ptrTo("foo"),
						},
					},
				}},
			}},
		},
		{
			name:       "invalid because repeated URLRewrite filter",
			wantErrors: []string{"URLRewrite filter cannot be repeated"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
									ReplacePrefixMatch: ptrTo("foo"),
								},
							},
						},
						{
							Type: gatewayv1.HTTPRouteFilterURLRewrite,
							URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
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
							Type: gatewayv1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
								},
							},
						},
						{
							Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
							RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
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
		{
			name:       "invalid because multiple filters are repeated",
			wantErrors: []string{"ResponseHeaderModifier filter cannot be repeated", "RequestRedirect filter cannot be repeated"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
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
							Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
							ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
								Add: []gatewayv1.HTTPHeader{
									{
										Name:  "my-header",
										Value: "bar",
									},
								},
							},
						},
						{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:            gatewayv1.FullPathHTTPPathModifier,
									ReplaceFullPath: ptrTo("foo"),
								},
							},
						},
						{
							Type: gatewayv1.HTTPRouteFilterRequestRedirect,
							RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
								Path: &gatewayv1.HTTPPathModifier{
									Type:            gatewayv1.FullPathHTTPPathModifier,
									ReplaceFullPath: ptrTo("bar"),
								},
							},
						},
					},
				},
			},
		},
		{
			name:       "too many matches and rules",
			wantErrors: []string{"total number of matches across all rules in a route must be less than 128"},
			rules: func() []gatewayv1.HTTPRouteRule {
				match := gatewayv1.HTTPRouteMatch{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
						Value: ptrTo("/"),
					},
				}
				var rules []gatewayv1.HTTPRouteRule
				for range 7 { // rules
					rule := gatewayv1.HTTPRouteRule{}
					for range 20 { // matches
						rule.Matches = append(rule.Matches, match)
					}
					rules = append(rules, rule)
				}
				return rules
			}(),
		},
		{
			name:       "many matches and few rules",
			wantErrors: nil,
			rules: func() []gatewayv1.HTTPRouteRule {
				match := gatewayv1.HTTPRouteMatch{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
						Value: ptrTo("/"),
					},
				}
				var rules []gatewayv1.HTTPRouteRule
				for range 2 { // rules
					rule := gatewayv1.HTTPRouteRule{}
					for range 48 { // matches
						rule.Matches = append(rule.Matches, match)
					}
					rules = append(rules, rule)
				}
				return rules
			}(),
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

func TestHTTPBackendRef(t *testing.T) {
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.HTTPRouteRule
	}{
		{
			name:       "invalid because repeated URLRewrite filter within backendRefs",
			wantErrors: []string{"URLRewrite filter cannot be repeated"},
			rules: []gatewayv1.HTTPRouteRule{
				{
					Matches: []gatewayv1.HTTPRouteMatch{
						{
							Path: &gatewayv1.HTTPPathMatch{
								Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
								Value: ptrTo("/"),
							},
						},
					},
					BackendRefs: []gatewayv1.HTTPBackendRef{
						{
							BackendRef: gatewayv1.BackendRef{
								BackendObjectReference: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(80)),
								},
							},
							Filters: []gatewayv1.HTTPRouteFilter{
								{
									Type: gatewayv1.HTTPRouteFilterURLRewrite,
									URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
										Path: &gatewayv1.HTTPPathModifier{
											Type:               gatewayv1.PrefixMatchHTTPPathModifier,
											ReplacePrefixMatch: ptrTo("foo"),
										},
									},
								},
								{
									Type: gatewayv1.HTTPRouteFilterURLRewrite,
									URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
										Path: &gatewayv1.HTTPPathModifier{
											Type:               gatewayv1.PrefixMatchHTTPPathModifier,
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

func TestHTTPPathModifier(t *testing.T) {
	tests := []struct {
		name         string
		wantErrors   []string
		pathModifier gatewayv1.HTTPPathModifier
	}{
		{
			name: "valid ReplaceFullPath",
			pathModifier: gatewayv1.HTTPPathModifier{
				Type:            gatewayv1.FullPathHTTPPathModifier,
				ReplaceFullPath: ptrTo("foo"),
			},
		},
		{
			name:       "replaceFullPath must be specified when type is set to 'ReplaceFullPath'",
			wantErrors: []string{"replaceFullPath must be specified when type is set to 'ReplaceFullPath'"},
			pathModifier: gatewayv1.HTTPPathModifier{
				Type: gatewayv1.FullPathHTTPPathModifier,
			},
		},
		{
			name:       "type must be 'ReplaceFullPath' when replaceFullPath is set",
			wantErrors: []string{"type must be 'ReplaceFullPath' when replaceFullPath is set"},
			pathModifier: gatewayv1.HTTPPathModifier{
				Type:            gatewayv1.PrefixMatchHTTPPathModifier,
				ReplaceFullPath: ptrTo("foo"),
			},
		},
		{
			name: "valid ReplacePrefixMatch",
			pathModifier: gatewayv1.HTTPPathModifier{
				Type:               gatewayv1.PrefixMatchHTTPPathModifier,
				ReplacePrefixMatch: ptrTo("/foo"),
			},
		},
		{
			name:       "replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'",
			wantErrors: []string{"replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'"},
			pathModifier: gatewayv1.HTTPPathModifier{
				Type: gatewayv1.PrefixMatchHTTPPathModifier,
			},
		},
		{
			name:       "type must be 'ReplacePrefixMatch' when replacePrefixMatch is set",
			wantErrors: []string{"type must be 'ReplacePrefixMatch' when replacePrefixMatch is set"},
			pathModifier: gatewayv1.HTTPPathModifier{
				Type:               gatewayv1.FullPathHTTPPathModifier,
				ReplacePrefixMatch: ptrTo("/foo"),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pathModifier := tc.pathModifier
			route := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{
						{
							Filters: []gatewayv1.HTTPRouteFilter{
								{
									Type: gatewayv1.HTTPRouteFilterRequestRedirect,
									RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
										Path: &pathModifier,
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

func validateHTTPRoute(t *testing.T, route *gatewayv1.HTTPRoute, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating HTTPRoute %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating HTTPRoute %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
