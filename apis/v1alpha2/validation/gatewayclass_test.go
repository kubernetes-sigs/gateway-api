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
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/util/validation/field"

	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestValidateGatewayClassUpdate(t *testing.T) {
	type args struct {
		oldClass *gatewayv1a2.GatewayClass
		newClass *gatewayv1a2.GatewayClass
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "changing parameters reference is allowed",
			args: args{
				oldClass: &gatewayv1a2.GatewayClass{
					Spec: gatewayv1a2.GatewayClassSpec{
						Controller: "foo",
					},
				},
				newClass: &gatewayv1a2.GatewayClass{
					Spec: gatewayv1a2.GatewayClassSpec{
						Controller: "foo",
						ParametersRef: &gatewayv1a2.ParametersReference{
							Group: "example.com",
							Kind:  "GatewayClassConfig",
							Name:  "foo",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "changing controller field results in an error",
			args: args{
				oldClass: &gatewayv1a2.GatewayClass{
					Spec: gatewayv1a2.GatewayClassSpec{
						Controller: "foo.com/gateway",
					},
				},
				newClass: &gatewayv1a2.GatewayClass{
					Spec: gatewayv1a2.GatewayClassSpec{
						Controller: "bar.com/gateway",
					},
				},
			},
			want: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.controller",
					Detail:   "cannot update an immutable field",
					BadValue: gatewayv1a2.GatewayController("bar.com/gateway"),
				},
			},
		},
		{
			name: "nil input result in no errors",
			args: args{
				oldClass: nil,
				newClass: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateGatewayClassUpdate(tt.args.oldClass, tt.args.newClass); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ValidateGatewayClassUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
