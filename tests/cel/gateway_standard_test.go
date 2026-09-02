//go:build standard
// +build standard

/*
Copyright The Kubernetes Authors.

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

func TestValidateGatewayStandard(t *testing.T) {
	ctx := context.Background()
	baseGateway := gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1.GatewaySpec{
			GatewayClassName: "foo",
			Listeners: []gatewayv1.Listener{
				{
					Name:     gatewayv1.SectionName("http"),
					Protocol: gatewayv1.HTTPProtocolType,
					Port:     gatewayv1.PortNumber(80),
				},
			},
		},
	}

	testCases := []struct {
		desc         string
		mutate       func(gw *gatewayv1.Gateway)
		mutateStatus func(gw *gatewayv1.Gateway)
		wantErrors   []string
	}{
		{
			desc: "tls config not set with https protocol",
			mutate: func(gw *gatewayv1.Gateway) {
				gw.Spec.Listeners = []gatewayv1.Listener{
					{
						Name:     gatewayv1.SectionName("https"),
						Protocol: gatewayv1.HTTPSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			gw := baseGateway.DeepCopy()
			gw.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(gw)
			}
			err := k8sClient.Create(ctx, gw)

			if tc.mutateStatus != nil {
				tc.mutateStatus(gw)
				err = k8sClient.Status().Update(ctx, gw)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating Gateway; got err=\n%v\n;want error=%v", err, tc.wantErrors != nil)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !celErrorStringMatches(err.Error(), wantError) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating Gateway; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
