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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

func TestValidateGateway(t *testing.T) {
	listeners := []gatewayv1a2.Listener{
		{
			Hostname: nil,
		},
	}
	baseGateway := gatewayv1a2.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1a2.GatewaySpec{
			GatewayClassName: "foo",
			Listeners:        listeners,
		},
	}

	testCases := map[string]struct {
		mutate             func(gw *gatewayv1a2.Gateway)
		expectErrsOnFields []string
	}{
		"nil hostname": {
			mutate:             func(gw *gatewayv1a2.Gateway) {},
			expectErrsOnFields: []string{},
		},
		"empty string hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"wildcard hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"wildcard-prefixed hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.example.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		"valid dns subdomain": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("foo.example.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{},
		},
		// Invalid use cases
		"IPv4 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3.4")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"Invalid IPv4 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3..4")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"IPv4 address with port hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("1.2.3.4:8080")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"IPv6 address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("2001:db8::68")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname", "spec.listeners[0].hostname"},
		},
		"IPv6 link-local address hostname": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("fe80::/10")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with port": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("foo.example.com:8080")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with invalid wildcard label": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with multiple wildcards": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.foo.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
		"dns subdomain with wildcard root label": {
			mutate: func(gw *gatewayv1a2.Gateway) {
				hostname := gatewayv1a2.Hostname("*.foo.*.com")
				gw.Spec.Listeners[0].Hostname = &hostname
			},
			expectErrsOnFields: []string{"spec.listeners[0].hostname"},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			gw := baseGateway.DeepCopy()
			tc.mutate(gw)
			errs := ValidateGateway(gw)
			if len(tc.expectErrsOnFields) != len(errs) {
				t.Fatalf("Expected %d errors, got %d errors: %v", len(tc.expectErrsOnFields), len(errs), errs)
			}
			for i, err := range errs {
				if err.Field != tc.expectErrsOnFields[i] {
					t.Errorf("Expected error on field: %s, got: %s", tc.expectErrsOnFields[i], err.Error())
				}
			}
		})
	}
}

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
