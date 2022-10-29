/*
Copyright 2022 The Kubernetes Authors.

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

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestValidateGRPCRoute(t *testing.T) {
	t.Parallel()

	service := "foo"
	method := "login"

	tests := []struct {
		name  string
		rules []gatewayv1a2.GRPCRouteRule
		errs  field.ErrorList
	}{
		{
			name: "valid GRPCRoute with 1 service in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: &service,
							},
						},
					},
				},
			},
		},
		{
			name: "valid GRPCRoute with 1 method in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Method: &method,
							},
						},
					},
				},
			},
		},
		{
			name: "invalid GRPCRoute missing service or method in GRPCMethodMatch field",
			rules: []gatewayv1a2.GRPCRouteRule{
				{
					Matches: []gatewayv1a2.GRPCRouteMatch{
						{
							Method: &gatewayv1a2.GRPCMethodMatch{
								Service: nil,
								Method: nil,
							},
						},
					},
				},
			},
			errs: field.ErrorList{
				{
					Type:	field.ErrorTypeRequired,
					Field:	"spec.rules[0].matches[0].methods",
					Detail: "should have at least one of fields Service and Method",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			route := gatewayv1a2.GRPCRoute{Spec: gatewayv1a2.GRPCRouteSpec{Rules: tc.rules}}
			errs := ValidateGRPCRoute(&route)
			if len(errs) != len(tc.errs) {
				t.Errorf("got %d errors, want %d errors: %s", len(errs), len(tc.errs), errs)
				t.FailNow()
			}
			for i := 0; i < len(errs); i++ {
				realErr := errs[i].Error()
				expectedErr := tc.errs[i].Error()
				if realErr != expectedErr {
					t.Errorf("expect error message: %s, but got: %s", expectedErr, realErr)
					t.FailNow()
				}
			}
		})
	}
}
