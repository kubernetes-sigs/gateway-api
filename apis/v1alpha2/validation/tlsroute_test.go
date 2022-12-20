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

func TestValidateTLSRoute(t *testing.T) {
	t.Parallel()

	var portNumber int32 = 9080

	tests := []struct {
		name  string
		rules []gatewayv1a2.TLSRouteRule
		errs  field.ErrorList
	}{
		{
			name:  "valid TLSRoute with 1 backendRef",
			rules: makeRouteRules[gatewayv1a2.TLSRouteRule](&portNumber),
		},
		{
			name:  "invalid TLSRoute with 1 backendRef (missing port)",
			rules: makeRouteRules[gatewayv1a2.TLSRouteRule](nil),
			errs: field.ErrorList{
				{
					Type:   field.ErrorTypeRequired,
					Field:  "spec.rules[0].backendRefs[0].port",
					Detail: "missing port for Service reference",
				},
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			route := gatewayv1a2.TLSRoute{Spec: gatewayv1a2.TLSRouteSpec{Rules: tc.rules}}
			errs := ValidateTLSRoute(&route)
			if len(errs) != len(tc.errs) {
				t.Fatalf("got %d errors, want %d errors: %s", len(errs), len(tc.errs), errs)
			}
			for i := 0; i < len(errs); i++ {
				realErr := errs[i].Error()
				expectedErr := tc.errs[i].Error()
				if realErr != expectedErr {
					t.Fatalf("expect error message: %s, but got: %s", expectedErr, realErr)
				}
			}
		})
	}
}
