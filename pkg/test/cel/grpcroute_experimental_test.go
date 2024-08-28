//go:build experimental
// +build experimental

/*
Copyright 2024 The Kubernetes Authors.

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
// `GRPCRouteFilter` hierarchy (i.e. either at the struct level, or on one of
// the immediate descendent fields), then the test will go in the
// TestGRPCRouteFilter function. If the appropriate test function does not
// exist, please create one.
//
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func TestGRPCRequestMirrorFilterExperimental(t *testing.T) {
	var percent int32 = 42
	var denominator int32 = 1000
	var bad_denominator int32 = 0
	testService := gatewayv1.ObjectName("test-service")
	tests := []struct {
		name       string
		wantErrors []string
		rules      []gatewayv1.GRPCRouteRule
	}{
		{
			name: "GRPCRoute - Invalid because both percent and fraction are specified",
			wantErrors: []string{"Only one of percent or fraction may be specified in HTTPRequestMirrorFilter"},
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Percent: &percent,
						Fraction: &gatewayv1.Fraction{
							Numerator: 83,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name: "GRPCRoute - Invalid fraction - numerator greater than denominator",
			wantErrors: []string{"numerator must be less than or equal to denominator"},
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator: 1001,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name: "GRPCRoute - Invalid fraction - denominator is 0",
			wantErrors: []string{"spec.rules[0].filters[0].requestMirror.fraction.denominator in body should be greater than or equal to 1"},
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator: 0,
							Denominator: &bad_denominator,
						},
					},
				}},
			}},
		},
		{
			name: "GRPCRoute - Invalid fraction - numerator is negative",
			wantErrors: []string{"spec.rules[0].filters[0].requestMirror.fraction.numerator in body should be greater than or equal to 0"},
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator: -1,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
		{
			name: "GRPCRoute - Valid with percent",
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
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
			name: "GRPCRoute - Valid with fraction",
			rules: []gatewayv1.GRPCRouteRule{{
				Filters: []gatewayv1.GRPCRouteFilter{{
					Type: gatewayv1.GRPCRouteFilterRequestMirror,
					RequestMirror: &gatewayv1.HTTPRequestMirrorFilter{
						BackendRef: gatewayv1.BackendObjectReference{
							Name: testService,
							Port: ptrTo(gatewayv1.PortNumber(8081)),
						},
						Fraction: &gatewayv1.Fraction{
							Numerator: 83,
							Denominator: &denominator,
						},
					},
				}},
			}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1.GRPCRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.GRPCRouteSpec{Rules: tc.rules},
			}
			validateGRPCRoute(t, route, tc.wantErrors)
		})
	}
}
