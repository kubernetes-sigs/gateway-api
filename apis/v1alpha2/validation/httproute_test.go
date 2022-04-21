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

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	pkgutils "sigs.k8s.io/gateway-api/pkg/util"
)

func TestValidateHTTPRoute(t *testing.T) {
	testService := gatewayv1a2.ObjectName("test-service")
	specialService := gatewayv1a2.ObjectName("special-service")
	tests := []struct {
		name     string
		rules    []gatewayv1a2.HTTPRouteRule
		errCount int
	}{{
		name:     "valid httpRoute with no filters",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
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
	}, {
		name:     "valid httpRoute with 1 filter",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
							Value: utilpointer.String("/"),
						},
					},
				},
				Filters: []gatewayv1a2.HTTPRouteFilter{
					{
						Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: testService,
								Port: pkgutils.PortNumberPtr(8081),
							},
						},
					},
				},
			},
		},
	}, {
		name:     "invalid httpRoute with 2 extended filters",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
							Value: utilpointer.String("/"),
						},
					},
				},
				Filters: []gatewayv1a2.HTTPRouteFilter{
					{
						Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: testService,
								Port: pkgutils.PortNumberPtr(8080),
							},
						},
					},
					{
						Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: specialService,
								Port: pkgutils.PortNumberPtr(8080),
							},
						},
					},
				},
			},
		},
	}, {
		name:     "invalid httpRoute with mix of filters and one duplicate",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
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
							BackendRef: gatewayv1a2.BackendObjectReference{
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
	}, {
		name:     "invalid httpRoute with multiple duplicate filters",
		errCount: 2,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
							Value: utilpointer.String("/"),
						},
					},
				},
				Filters: []gatewayv1a2.HTTPRouteFilter{
					{
						Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
						RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
							BackendRef: gatewayv1a2.BackendObjectReference{
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
							BackendRef: gatewayv1a2.BackendObjectReference{
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
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: specialService,
								Port: pkgutils.PortNumberPtr(8080),
							},
						},
					},
				},
			},
		},
	}, {
		name:     "valid httpRoute with duplicate ExtensionRef filters",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Matches: []gatewayv1a2.HTTPRouteMatch{
					{
						Path: &gatewayv1a2.HTTPPathMatch{
							Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
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
							BackendRef: gatewayv1a2.BackendObjectReference{
								Name: testService,
								Port: pkgutils.PortNumberPtr(8080),
							},
						},
					},
					{
						Type: "ExtensionRef",
						ExtensionRef: &gatewayv1a2.LocalObjectReference{
							Kind: "Service",
							Name: "test",
						},
					},
					{
						Type: "ExtensionRef",
						ExtensionRef: &gatewayv1a2.LocalObjectReference{
							Kind: "Service",
							Name: "test",
						},
					},
					{
						Type: "ExtensionRef",
						ExtensionRef: &gatewayv1a2.LocalObjectReference{
							Kind: "Service",
							Name: "test",
						},
					},
				},
			},
		},
	}, {
		name:     "valid redirect path modifier",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{
			{
				Filters: []gatewayv1a2.HTTPRouteFilter{
					{
						Type: gatewayv1a2.HTTPRouteFilterRequestRedirect,
						RequestRedirect: &gatewayv1a2.HTTPRequestRedirectFilter{
							Path: &gatewayv1a2.HTTPPathModifier{
								Type:            gatewayv1a2.FullPathHTTPPathModifier,
								ReplaceFullPath: utilpointer.String("foo"),
							},
						},
					},
				},
			},
		},
	}, {
		name:     "redirect path modifier with type mismatch",
		errCount: 2,
		rules: []gatewayv1a2.HTTPRouteRule{{
			Filters: []gatewayv1a2.HTTPRouteFilter{{
				Type: gatewayv1a2.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1a2.HTTPRequestRedirectFilter{
					Path: &gatewayv1a2.HTTPPathModifier{
						Type:            gatewayv1a2.PrefixMatchHTTPPathModifier,
						ReplaceFullPath: utilpointer.String("foo"),
					},
				},
			}},
		}},
	}, {
		name:     "valid rewrite path modifier",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{{
			Filters: []gatewayv1a2.HTTPRouteFilter{{
				Type: gatewayv1a2.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1a2.HTTPURLRewriteFilter{
					Path: &gatewayv1a2.HTTPPathModifier{
						Type:               gatewayv1a2.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: utilpointer.String("foo"),
					},
				},
			}},
		}},
	}, {
		name:     "redirect path modifier with type mismatch",
		errCount: 2,
		rules: []gatewayv1a2.HTTPRouteRule{{
			Filters: []gatewayv1a2.HTTPRouteFilter{{
				Type: gatewayv1a2.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1a2.HTTPURLRewriteFilter{
					Path: &gatewayv1a2.HTTPPathModifier{
						Type:               gatewayv1a2.FullPathHTTPPathModifier,
						ReplacePrefixMatch: utilpointer.String("foo"),
					},
				},
			}},
		}},
	}, {
		name:     "rewrite and redirect filters combined (invalid)",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{{
			Filters: []gatewayv1a2.HTTPRouteFilter{{
				Type: gatewayv1a2.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1a2.HTTPURLRewriteFilter{
					Path: &gatewayv1a2.HTTPPathModifier{
						Type:               gatewayv1a2.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: utilpointer.String("foo"),
					},
				},
			}, {
				Type: gatewayv1a2.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1a2.HTTPRequestRedirectFilter{
					Path: &gatewayv1a2.HTTPPathModifier{
						Type:               gatewayv1a2.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: utilpointer.String("foo"),
					},
				},
			}},
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs field.ErrorList
			route := gatewayv1a2.HTTPRoute{Spec: gatewayv1a2.HTTPRouteSpec{Rules: tc.rules}}
			errs = ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateHTTPBackendUniqueFilters(t *testing.T) {
	var testService v1alpha2.ObjectName = "testService"
	var specialService v1alpha2.ObjectName = "specialService"
	tests := []struct {
		name     string
		rules    []gatewayv1a2.HTTPRouteRule
		errCount int
	}{{
		name:     "valid httpRoute Rules backendref filters",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{{
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
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
					},
				},
			},
		}},
	}, {
		name:     "invalid httpRoute Rules duplicate mirror filter",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{
				{
					BackendRef: gatewayv1a2.BackendRef{
						BackendObjectReference: gatewayv1a2.BackendObjectReference{
							Name: testService,
							Port: pkgutils.PortNumberPtr(8080),
						},
					},
					Filters: []gatewayv1a2.HTTPRouteFilter{
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: testService,
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						},
						{
							Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1a2.BackendObjectReference{
									Name: specialService,
									Port: pkgutils.PortNumberPtr(8080),
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
			route := gatewayv1a2.HTTPRoute{Spec: gatewayv1a2.HTTPRouteSpec{Rules: tc.rules}}
			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateHTTPPathMatch(t *testing.T) {
	tests := []struct {
		name     string
		path     *gatewayv1a2.HTTPPathMatch
		errCount int
	}{{
		name: "invalid httpRoute prefix",
		path: &gatewayv1a2.HTTPPathMatch{
			Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
			Value: utilpointer.String("/."),
		},
		errCount: 1,
	}, {
		name: "invalid httpRoute Exact",
		path: &gatewayv1a2.HTTPPathMatch{
			Type:  pkgutils.PathMatchTypePtr("Exact"),
			Value: utilpointer.String("/foo/./bar"),
		},
		errCount: 1,
	}, {
		name: "invalid httpRoute prefix",
		path: &gatewayv1a2.HTTPPathMatch{
			Type:  pkgutils.PathMatchTypePtr("PathPrefix"),
			Value: utilpointer.String("/"),
		},
		errCount: 0,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1a2.HTTPRoute{Spec: gatewayv1a2.HTTPRouteSpec{
				Rules: []gatewayv1a2.HTTPRouteRule{{
					Matches: []gatewayv1a2.HTTPRouteMatch{{
						Path: tc.path,
					}},
					BackendRefs: []gatewayv1a2.HTTPBackendRef{{
						BackendRef: gatewayv1a2.BackendRef{
							BackendObjectReference: gatewayv1a2.BackendObjectReference{
								Name: gatewayv1a2.ObjectName("test"),
								Port: pkgutils.PortNumberPtr(8080),
							},
						},
					}},
				}},
			}}

			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateServicePort(t *testing.T) {
	portPtr := func(n int) *gatewayv1a2.PortNumber {
		p := gatewayv1a2.PortNumber(n)
		return &p
	}

	groupPtr := func(g string) *gatewayv1a2.Group {
		p := gatewayv1a2.Group(g)
		return &p
	}

	kindPtr := func(k string) *gatewayv1a2.Kind {
		p := gatewayv1a2.Kind(k)
		return &p
	}

	tests := []struct {
		name     string
		rules    []gatewayv1a2.HTTPRouteRule
		errCount int
	}{{
		name:     "default groupkind with port",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{{
				BackendRef: gatewayv1a2.BackendRef{
					BackendObjectReference: gatewayv1a2.BackendObjectReference{
						Name: "backend",
						Port: portPtr(99),
					},
				},
			}},
		}},
	}, {
		name:     "default groupkind with no port",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{{
				BackendRef: gatewayv1a2.BackendRef{
					BackendObjectReference: gatewayv1a2.BackendObjectReference{
						Name: "backend",
					},
				},
			}},
		}},
	}, {
		name:     "explicit service with port",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{{
				BackendRef: gatewayv1a2.BackendRef{
					BackendObjectReference: gatewayv1a2.BackendObjectReference{
						Group: groupPtr(""),
						Kind:  kindPtr("Service"),
						Name:  "backend",
						Port:  portPtr(99),
					},
				},
			}},
		}},
	}, {
		name:     "explicit service with no port",
		errCount: 1,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{{
				BackendRef: gatewayv1a2.BackendRef{
					BackendObjectReference: gatewayv1a2.BackendObjectReference{
						Group: groupPtr(""),
						Kind:  kindPtr("Service"),
						Name:  "backend",
					},
				},
			}},
		}},
	}, {
		name:     "explicit ref with no port",
		errCount: 0,
		rules: []gatewayv1a2.HTTPRouteRule{{
			BackendRefs: []gatewayv1a2.HTTPBackendRef{{
				BackendRef: gatewayv1a2.BackendRef{
					BackendObjectReference: gatewayv1a2.BackendObjectReference{
						Group: groupPtr("foo.example.com"),
						Kind:  kindPtr("Foo"),
						Name:  "backend",
					},
				},
			}},
		}},
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1a2.HTTPRoute{Spec: gatewayv1a2.HTTPRouteSpec{Rules: tc.rules}}
			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateHTTPRouteTypeMatchesField(t *testing.T) {
	tests := []struct {
		name        string
		routeFilter gatewayv1a2.HTTPRouteFilter
		errCount    int
	}{{
		name: "valid HTTPRouteFilterRequestHeaderModifier route filter",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
			RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{
				Set:    []gatewayv1a2.HTTPHeader{{Name: "name"}},
				Add:    []gatewayv1a2.HTTPHeader{{Name: "add"}},
				Remove: []string{"remove"},
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with non-matching field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type:          gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
			RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with empty value field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestHeaderModifier,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterRequestMirror route filter",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
			RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{BackendRef: gatewayv1a2.BackendObjectReference{
				Group:     new(gatewayv1a2.Group),
				Kind:      new(gatewayv1a2.Kind),
				Name:      "name",
				Namespace: new(gatewayv1a2.Namespace),
				Port:      pkgutils.PortNumberPtr(22),
			}},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterRequestMirror type filter with non-matching field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type:                  gatewayv1a2.HTTPRouteFilterRequestMirror,
			RequestHeaderModifier: &gatewayv1a2.HTTPRequestHeaderFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestMirror type filter with empty value field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestMirror,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterRequestRedirect route filter",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestRedirect,
			RequestRedirect: &gatewayv1a2.HTTPRequestRedirectFilter{
				Scheme:     new(string),
				Hostname:   new(gatewayv1a2.PreciseHostname),
				Path:       &gatewayv1a2.HTTPPathModifier{},
				Port:       new(gatewayv1a2.PortNumber),
				StatusCode: new(int),
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterRequestRedirect type filter with non-matching field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type:          gatewayv1a2.HTTPRouteFilterRequestRedirect,
			RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestRedirect type filter with empty value field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterRequestRedirect,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterExtensionRef filter",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterExtensionRef,
			ExtensionRef: &gatewayv1a2.LocalObjectReference{
				Group: "group",
				Kind:  "kind",
				Name:  "name",
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterExtensionRef type filter with non-matching field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type:          gatewayv1a2.HTTPRouteFilterExtensionRef,
			RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterExtensionRef type filter with empty value field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterExtensionRef,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterURLRewrite route filter",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterURLRewrite,
			URLRewrite: &gatewayv1a2.HTTPURLRewriteFilter{
				Hostname: new(gatewayv1a2.PreciseHostname),
				Path:     &gatewayv1a2.HTTPPathModifier{},
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterURLRewrite type filter with non-matching field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type:          gatewayv1a2.HTTPRouteFilterURLRewrite,
			RequestMirror: &gatewayv1a2.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterURLRewrite type filter with empty value field",
		routeFilter: gatewayv1a2.HTTPRouteFilter{
			Type: gatewayv1a2.HTTPRouteFilterURLRewrite,
		},
		errCount: 1,
	}, {
		name:        "empty type filter is valid (caught by CRD validation)",
		routeFilter: gatewayv1a2.HTTPRouteFilter{},
		errCount:    0,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1a2.HTTPRoute{
				Spec: gatewayv1a2.HTTPRouteSpec{
					Rules: []gatewayv1a2.HTTPRouteRule{{
						Filters: []gatewayv1a2.HTTPRouteFilter{tc.routeFilter},
						BackendRefs: []gatewayv1a2.HTTPBackendRef{{
							BackendRef: gatewayv1a2.BackendRef{
								BackendObjectReference: gatewayv1a2.BackendObjectReference{
									Name: gatewayv1a2.ObjectName("test"),
									Port: pkgutils.PortNumberPtr(8080),
								},
							},
						}},
					}},
				},
			}
			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}
