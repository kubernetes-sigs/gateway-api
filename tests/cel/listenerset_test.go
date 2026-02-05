//go:build experimental
// +build experimental

/*
Copyright 2026 The Kubernetes Authors.

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestValidateListenerSet(t *testing.T) {
	ctx := context.Background()
	baseListenerSet := gatewayv1.ListenerSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: gatewayv1.ListenerSetSpec{
			ParentRef: gatewayv1.ParentGatewayReference{
				Kind: ptr.To(gatewayv1.Kind("Gateway")),
				Name: gatewayv1.ObjectName("example"),
			},
			Listeners: []gatewayv1.ListenerEntry{
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
		mutate       func(ls *gatewayv1.ListenerSet)
		mutateStatus func(ls *gatewayv1.ListenerSet)
		wantErrors   []string
	}{
		{
			desc: "tls config present with tls protocol",
			mutate: func(ls *gatewayv1.ListenerSet) {
				ls.Spec.Listeners = []gatewayv1.ListenerEntry{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS: &gatewayv1.ListenerTLSConfig{
							Mode: ptrTo(gatewayv1.TLSModeType("Passthrough")),
						},
					},
				}
			},
		},
		{
			desc: "tls config not set with tls protocol",
			mutate: func(ls *gatewayv1.ListenerSet) {
				ls.Spec.Listeners = []gatewayv1.ListenerEntry{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
					},
				}
			},
			wantErrors: []string{"tls mode must be set for protocol TLS"},
		},
		{
			desc: "tls config present but empty with tls protocol",
			mutate: func(ls *gatewayv1.ListenerSet) {
				ls.Spec.Listeners = []gatewayv1.ListenerEntry{
					{
						Name:     gatewayv1.SectionName("tls"),
						Protocol: gatewayv1.TLSProtocolType,
						Port:     gatewayv1.PortNumber(8443),
						TLS:      &gatewayv1.ListenerTLSConfig{},
					},
				}
			},
			wantErrors: []string{"certificateRefs or options must be specified when mode is Terminate"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			ls := baseListenerSet.DeepCopy()
			ls.Name = fmt.Sprintf("foo-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(ls)
			}
			err := k8sClient.Create(ctx, ls)

			if tc.mutateStatus != nil {
				tc.mutateStatus(ls)
				err = k8sClient.Status().Update(ctx, ls)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating ListenerSet; got err=\n%v\n;want error=%v", err, tc.wantErrors != nil)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !celErrorStringMatches(err.Error(), wantError) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating ListenerSet; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
