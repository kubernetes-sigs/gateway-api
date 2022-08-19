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

	gatewayv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestValidateGatewayClassUpdate(t *testing.T) {
	type args struct {
		oldClass *gatewayv1b1.GatewayClass
		newClass *gatewayv1b1.GatewayClass
	}
	tests := []struct {
		name string
		args args
		want field.ErrorList
	}{
		{
			name: "changing parameters reference is allowed",
			args: args{
				oldClass: &gatewayv1b1.GatewayClass{
					Spec: gatewayv1b1.GatewayClassSpec{
						ControllerName: "foo",
					},
				},
				newClass: &gatewayv1b1.GatewayClass{
					Spec: gatewayv1b1.GatewayClassSpec{
						ControllerName: "foo",
						ParametersRef: &gatewayv1b1.ParametersReference{
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
				oldClass: &gatewayv1b1.GatewayClass{
					Spec: gatewayv1b1.GatewayClassSpec{
						ControllerName: "example.com/gateway",
					},
				},
				newClass: &gatewayv1b1.GatewayClass{
					Spec: gatewayv1b1.GatewayClassSpec{
						ControllerName: "example.org/gateway",
					},
				},
			},
			want: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.controllerName",
					Detail:   "cannot update an immutable field",
					BadValue: gatewayv1b1.GatewayController("example.org/gateway"),
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
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ValidateGatewayClassUpdate(tc.args.oldClass, tc.args.newClass); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("ValidateGatewayClassUpdate() = %v, want %v", got, tc.want)
			}
		})
	}
}
