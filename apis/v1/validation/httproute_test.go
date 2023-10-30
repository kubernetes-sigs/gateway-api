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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestValidateHTTPRoute(t *testing.T) {
	testService := gatewayv1.ObjectName("test-service")
	pathPrefixMatchType := gatewayv1.PathMatchPathPrefix

	tests := []struct {
		name     string
		rules    []gatewayv1.HTTPRouteRule
		errCount int
	}{{
		name:     "valid httpRoute with no filters",
		errCount: 0,
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
	}, {
		name:     "valid httpRoute with 1 filter",
		errCount: 0,
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
	}, {
		name:     "invalid httpRoute with 2 extended filters",
		errCount: 1,
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
	}, {
		name:     "invalid httpRoute with mix of filters and one duplicate",
		errCount: 1,
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
	}, {
		name:     "invalid httpRoute with multiple duplicate filters",
		errCount: 2,
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
					{
						Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
						ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
							Add: []gatewayv1.HTTPHeader{
								{
									Name:  "extra-header",
									Value: "foo",
								},
							},
						},
					},
					{
						Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
						ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
							Set: []gatewayv1.HTTPHeader{
								{
									Name:  "other-header",
									Value: "bat",
								},
							},
						},
					},
				},
			},
		},
	}, {
		name:     "valid httpRoute with duplicate ExtensionRef filters",
		errCount: 0,
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
	}, {
		name:     "valid redirect path modifier",
		errCount: 0,
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
	}, {
		name:     "redirect path modifier with type mismatch",
		errCount: 2,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type:            gatewayv1.PrefixMatchHTTPPathModifier,
						ReplaceFullPath: ptrTo("foo"),
					},
				},
			}},
		}},
	}, {
		name:     "valid rewrite path modifier",
		errCount: 0,
		rules: []gatewayv1.HTTPRouteRule{{
			Matches: []gatewayv1.HTTPRouteMatch{{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  &pathPrefixMatchType,
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
	}, {
		name:     "rewrite path modifier missing path match",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
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
	}, {
		name:     "rewrite path too many matches",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
			Matches: []gatewayv1.HTTPRouteMatch{{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  &pathPrefixMatchType,
					Value: ptrTo("/foo"),
				},
			}, {
				Path: &gatewayv1.HTTPPathMatch{
					Type:  &pathPrefixMatchType,
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
	}, {
		name:     "redirect path modifier with type mismatch",
		errCount: 2,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type:               gatewayv1.FullPathHTTPPathModifier,
						ReplacePrefixMatch: ptrTo("foo"),
					},
				},
			}},
		}},
	}, {
		name:     "rewrite and redirect filters combined (invalid)",
		errCount: 3,
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
	}, {
		name:     "multiple actions for the same request header (invalid)",
		errCount: 2,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{
							Name:  gatewayv1.HTTPHeaderName("x-fruit"),
							Value: "apple",
						},
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
							Name:  gatewayv1.HTTPHeaderName("x-grain"),
							Value: "wheat",
						},
						{
							Name:  gatewayv1.HTTPHeaderName("x-spice"),
							Value: "coriander",
						},
					},
				},
			}},
		}},
	}, {
		name:     "multiple actions for the same request header with inconsistent case (invalid)",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{
							Name:  gatewayv1.HTTPHeaderName("x-fruit"),
							Value: "apple",
						},
					},
					Set: []gatewayv1.HTTPHeader{
						{
							Name:  gatewayv1.HTTPHeaderName("X-Fruit"),
							Value: "watermelon",
						},
					},
				},
			}},
		}},
	}, {
		name:     "multiple of the same action for the same request header (invalid)",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{
							Name:  gatewayv1.HTTPHeaderName("x-fruit"),
							Value: "apple",
						},
						{
							Name:  gatewayv1.HTTPHeaderName("x-fruit"),
							Value: "plum",
						},
					},
				},
			}},
		}},
	}, {
		name:     "multiple actions for different request headers",
		errCount: 0,
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
	}, {
		name:     "multiple actions for the same response header (invalid)",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
			Filters: []gatewayv1.HTTPRouteFilter{{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{{
						Name:  gatewayv1.HTTPHeaderName("x-example"),
						Value: "blueberry",
					}},
					Set: []gatewayv1.HTTPHeader{{
						Name:  gatewayv1.HTTPHeaderName("x-example"),
						Value: "turnip",
					}},
				},
			}},
		}},
	}, {
		name:     "multiple actions for different response headers",
		errCount: 0,
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
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs field.ErrorList
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules}}
			errs = ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateHTTPBackendUniqueFilters(t *testing.T) {
	var testService gatewayv1.ObjectName = "testService"
	var specialService gatewayv1.ObjectName = "specialService"
	tests := []struct {
		name     string
		rules    []gatewayv1.HTTPRouteRule
		errCount int
	}{{
		name:     "valid httpRoute Rules backendref filters",
		errCount: 0,
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{
				{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8080)),
						},
						Weight: ptrTo(int32(100)),
					},
					Filters: []gatewayv1.HTTPRouteFilter{
						{
							Type: gatewayv1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1.BackendObjectReference{
									Name: testService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
								},
							},
						},
					},
				},
			},
		}},
	}, {
		name:     "valid httpRoute Rules duplicate mirror filter",
		errCount: 0,
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{
				{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8080)),
						},
					},
					Filters: []gatewayv1.HTTPRouteFilter{
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
							Type: gatewayv1.HTTPRouteFilterRequestMirror,
							RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
								BackendRef: gatewayv1.BackendObjectReference{
									Name: specialService,
									Port: ptrTo(gatewayv1.PortNumber(8080)),
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
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules}}
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
		path     *gatewayv1.HTTPPathMatch
		errCount int
	}{{
		name: "invalid httpRoute prefix (/.)",
		path: &gatewayv1.HTTPPathMatch{
			Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
			Value: ptrTo("/."),
		},
		errCount: 1,
	}, {
		name: "invalid exact (/./)",
		path: &gatewayv1.HTTPPathMatch{
			Type:  ptrTo(gatewayv1.PathMatchType("Exact")),
			Value: ptrTo("/foo/./bar"),
		},
		errCount: 1,
	}, {
		name: "valid httpRoute prefix",
		path: &gatewayv1.HTTPPathMatch{
			Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
			Value: ptrTo("/"),
		},
		errCount: 0,
	}, {
		name: "invalid httpRoute prefix (/[])",
		path: &gatewayv1.HTTPPathMatch{
			Type:  ptrTo(gatewayv1.PathMatchType("PathPrefix")),
			Value: ptrTo("/[]"),
		},
		errCount: 1,
	}, {
		name: "invalid httpRoute exact (/^)",
		path: &gatewayv1.HTTPPathMatch{
			Type:  ptrTo(gatewayv1.PathMatchType("Exact")),
			Value: ptrTo("/^"),
		},
		errCount: 1,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{
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
			}}

			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateHTTPHeaderMatches(t *testing.T) {
	tests := []struct {
		name          string
		headerMatches []gatewayv1.HTTPHeaderMatch
		expectErr     string
	}{{
		name:          "no header matches",
		headerMatches: nil,
		expectErr:     "",
	}, {
		name: "no header matched more than once",
		headerMatches: []gatewayv1.HTTPHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "Header-Name-3", Value: "val-3"},
		},
		expectErr: "",
	}, {
		name: "header matched more than once (same case)",
		headerMatches: []gatewayv1.HTTPHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "Header-Name-1", Value: "val-3"},
		},
		expectErr: "spec.rules[0].matches[0].headers: Invalid value: \"Header-Name-1\": cannot match the same header multiple times in the same rule",
	}, {
		name: "header matched more than once (different case)",
		headerMatches: []gatewayv1.HTTPHeaderMatch{
			{Name: "Header-Name-1", Value: "val-1"},
			{Name: "Header-Name-2", Value: "val-2"},
			{Name: "HEADER-NAME-2", Value: "val-3"},
		},
		expectErr: "spec.rules[0].matches[0].headers: Invalid value: \"Header-Name-2\": cannot match the same header multiple times in the same rule",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{
				Rules: []gatewayv1.HTTPRouteRule{{
					Matches: []gatewayv1.HTTPRouteMatch{{
						Headers: tc.headerMatches,
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
			}}

			errs := ValidateHTTPRoute(&route)
			if len(tc.expectErr) == 0 {
				assert.Emptyf(t, errs, "expected no errors, got %d errors: %s", len(errs), errs)
			} else {
				require.Lenf(t, errs, 1, "expected one error, got %d errors: %s", len(errs), errs)
				assert.Equal(t, tc.expectErr, errs[0].Error())
			}
		})
	}
}

func TestValidateHTTPQueryParamMatches(t *testing.T) {
	tests := []struct {
		name              string
		queryParamMatches []gatewayv1.HTTPQueryParamMatch
		expectErr         string
	}{{
		name:              "no query param matches",
		queryParamMatches: nil,
		expectErr:         "",
	}, {
		name: "no query param matched more than once",
		queryParamMatches: []gatewayv1.HTTPQueryParamMatch{
			{Name: "query-param-1", Value: "val-1"},
			{Name: "query-param-2", Value: "val-2"},
			{Name: "query-param-3", Value: "val-3"},
		},
		expectErr: "",
	}, {
		name: "query param matched more than once",
		queryParamMatches: []gatewayv1.HTTPQueryParamMatch{
			{Name: "query-param-1", Value: "val-1"},
			{Name: "query-param-2", Value: "val-2"},
			{Name: "query-param-1", Value: "val-3"},
		},
		expectErr: "spec.rules[0].matches[0].queryParams: Invalid value: \"query-param-1\": cannot match the same query parameter multiple times in the same rule",
	}, {
		name: "query param names with different casing are not considered duplicates",
		queryParamMatches: []gatewayv1.HTTPQueryParamMatch{
			{Name: "query-param-1", Value: "val-1"},
			{Name: "query-param-2", Value: "val-2"},
			{Name: "QUERY-PARAM-1", Value: "val-3"},
		},
		expectErr: "",
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{
				Rules: []gatewayv1.HTTPRouteRule{{
					Matches: []gatewayv1.HTTPRouteMatch{{
						QueryParams: tc.queryParamMatches,
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
			}}

			errs := ValidateHTTPRoute(&route)
			if len(tc.expectErr) == 0 {
				assert.Emptyf(t, errs, "expected no errors, got %d errors: %s", len(errs), errs)
			} else {
				require.Lenf(t, errs, 1, "expected one error, got %d errors: %s", len(errs), errs)
				assert.Equal(t, tc.expectErr, errs[0].Error())
			}
		})
	}
}

func TestValidateServicePort(t *testing.T) {
	portPtr := func(n int) *gatewayv1.PortNumber {
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
		name     string
		rules    []gatewayv1.HTTPRouteRule
		errCount int
	}{{
		name:     "default groupkind with port",
		errCount: 0,
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: "backend",
						Port: portPtr(99),
					},
				},
			}},
		}},
	}, {
		name:     "default groupkind with no port",
		errCount: 1,
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: "backend",
					},
				},
			}},
		}},
	}, {
		name:     "explicit service with port",
		errCount: 0,
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
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
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
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
		rules: []gatewayv1.HTTPRouteRule{{
			BackendRefs: []gatewayv1.HTTPBackendRef{{
				BackendRef: gatewayv1.BackendRef{
					BackendObjectReference: gatewayv1.BackendObjectReference{
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
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules}}
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
		routeFilter gatewayv1.HTTPRouteFilter
		errCount    int
	}{{
		name: "valid HTTPRouteFilterRequestHeaderModifier route filter",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
			RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
				Set:    []gatewayv1.HTTPHeader{{Name: "name"}},
				Add:    []gatewayv1.HTTPHeader{{Name: "add"}},
				Remove: []string{"remove"},
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with non-matching field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type:          gatewayv1.HTTPRouteFilterRequestHeaderModifier,
			RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestHeaderModifier type filter with empty value field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterRequestMirror route filter",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestMirror,
			RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{BackendRef: gatewayv1.BackendObjectReference{
				Group:     new(gatewayv1.Group),
				Kind:      new(gatewayv1.Kind),
				Name:      "name",
				Namespace: new(gatewayv1.Namespace),
				Port:      ptrTo(gatewayv1.PortNumber(22)),
			}},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterRequestMirror type filter with non-matching field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type:                  gatewayv1.HTTPRouteFilterRequestMirror,
			RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestMirror type filter with empty value field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestMirror,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterRequestRedirect route filter",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestRedirect,
			RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
				Scheme:     new(string),
				Hostname:   new(gatewayv1.PreciseHostname),
				Path:       &gatewayv1.HTTPPathModifier{},
				Port:       new(gatewayv1.PortNumber),
				StatusCode: new(int),
			},
		},
		errCount: 1,
	}, {
		name: "invalid HTTPRouteFilterRequestRedirect type filter with non-matching field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type:          gatewayv1.HTTPRouteFilterRequestRedirect,
			RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterRequestRedirect type filter with empty value field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterRequestRedirect,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterExtensionRef filter",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterExtensionRef,
			ExtensionRef: &gatewayv1.LocalObjectReference{
				Group: "group",
				Kind:  "kind",
				Name:  "name",
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterExtensionRef type filter with non-matching field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type:          gatewayv1.HTTPRouteFilterExtensionRef,
			RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterExtensionRef type filter with empty value field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterExtensionRef,
		},
		errCount: 1,
	}, {
		name: "valid HTTPRouteFilterURLRewrite route filter",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterURLRewrite,
			URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
				Hostname: new(gatewayv1.PreciseHostname),
				Path:     &gatewayv1.HTTPPathModifier{},
			},
		},
		errCount: 0,
	}, {
		name: "invalid HTTPRouteFilterURLRewrite type filter with non-matching field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type:          gatewayv1.HTTPRouteFilterURLRewrite,
			RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{},
		},
		errCount: 2,
	}, {
		name: "invalid HTTPRouteFilterURLRewrite type filter with empty value field",
		routeFilter: gatewayv1.HTTPRouteFilter{
			Type: gatewayv1.HTTPRouteFilterURLRewrite,
		},
		errCount: 1,
	}, {
		name:        "empty type filter is valid (caught by CRD validation)",
		routeFilter: gatewayv1.HTTPRouteFilter{},
		errCount:    0,
	}}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.HTTPRoute{
				Spec: gatewayv1.HTTPRouteSpec{
					Rules: []gatewayv1.HTTPRouteRule{{
						Filters: []gatewayv1.HTTPRouteFilter{tc.routeFilter},
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
			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func TestValidateRequestRedirectFiltersWithNoBackendRef(t *testing.T) {
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name     string
		rules    []gatewayv1.HTTPRouteRule
		errCount int
	}{
		{
			name:     "backendref with request redirect httpRoute filter",
			errCount: 1,
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
		}, {
			name:     "request redirect without backendref in httpRoute filter",
			errCount: 0,
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var errs field.ErrorList
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules}}
			errs = ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}

func toDuration(durationString string) *gatewayv1.Duration {
	return (*gatewayv1.Duration)(&durationString)
}

func TestValidateHTTPTimeouts(t *testing.T) {
	tests := []struct {
		name     string
		rules    []gatewayv1.HTTPRouteRule
		errCount int
	}{
		{
			name:     "valid httpRoute Rules timeouts",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("1ms"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules timeout set to 0s (disabled)",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("0s"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules timeout set to 0ms (disabled)",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("0ms"),
					},
				},
			},
		}, {}, {
			name:     "valid httpRoute Rules timeout set to 0h (disabled)",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request: toDuration("0h"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules timeout and backendRequest have the same value",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("1ms"),
						BackendRequest: toDuration("1ms"),
					},
				},
			},
		}, {
			name:     "invalid httpRoute Rules backendRequest timeout cannot be longer than request timeout",
			errCount: 1,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("1ms"),
						BackendRequest: toDuration("2ms"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules request timeout 1s and backendRequest timeout 200ms",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("1s"),
						BackendRequest: toDuration("200ms"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules request timeout 10s and backendRequest timeout 10s",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("10s"),
						BackendRequest: toDuration("10s"),
					},
				},
			},
		}, {
			name:     "invalid httpRoute Rules backendRequest timeout cannot be greater than request timeout",
			errCount: 1,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("200ms"),
						BackendRequest: toDuration("1s"),
					},
				},
			},
		}, {
			name:     "valid httpRoute Rules request 0s (infinite) and backendRequest 100ms",
			errCount: 0,
			rules: []gatewayv1.HTTPRouteRule{
				{
					Timeouts: &gatewayv1.HTTPRouteTimeouts{
						Request:        toDuration("0s"),
						BackendRequest: toDuration("100ms"),
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := gatewayv1.HTTPRoute{Spec: gatewayv1.HTTPRouteSpec{Rules: tc.rules}}
			errs := ValidateHTTPRoute(&route)
			if len(errs) != tc.errCount {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), tc.errCount, errs)
			}
		})
	}
}
