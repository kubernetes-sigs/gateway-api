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

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestValidateGatewayClassUpdate(t *testing.T) {
	ctx := context.Background()
	baseGatewayClass := gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo",
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: "example.net/gateway-controller",
		},
	}

	testCases := []struct {
		desc           string
		creationMutate func(gw *gatewayv1.GatewayClass)
		updationMutate func(gw *gatewayv1.GatewayClass)
		wantError      string
	}{
		{
			desc: "cannot upgrade controllerName",
			creationMutate: func(gwc *gatewayv1.GatewayClass) {
				gwc.Spec.ControllerName = "example.net/gateway-controller-1"
			},
			updationMutate: func(gwc *gatewayv1.GatewayClass) {
				gwc.Spec.ControllerName = "example.net/gateway-controller-2"
			},
			wantError: "Value is immutable",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gwc := baseGatewayClass.DeepCopy()
			gwc.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())

			tc.creationMutate(gwc)
			if err := k8sClient.Create(ctx, gwc); err != nil {
				t.Fatalf("Failed to create GatewayClass: %v", err)
			}
			tc.updationMutate(gwc)
			err := k8sClient.Update(ctx, gwc)

			if (tc.wantError != "") != (err != nil) {
				t.Fatalf("Unexpected error while updating GatewayClass; got err=\n%v\n;want error=%v", err, tc.wantError != "")
			}
			if tc.wantError != "" && !celErrorStringMatches(err.Error(), tc.wantError) {
				t.Fatalf("Unexpected error while updating GatewayClass; got err=\n%v\n;want substring within error=%q", err, tc.wantError)
			}
		})
	}
}
