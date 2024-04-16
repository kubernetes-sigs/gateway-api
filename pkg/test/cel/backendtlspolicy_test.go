//go:build experimental
// +build experimental

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
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestBackendTLSPolicyConfig(t *testing.T) {
	tests := []struct {
		name        string
		wantErrors  []string
		routeConfig gatewayv1a2.BackendTLSPolicyConfig
	}{
		{
			name: "valid BackendTLSPolicyConfig with WellKnownCACerts",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{
				WellKnownCACerts: ptrTo(gatewayv1a2.WellKnownCACertType("System")),
				Hostname:         "foo.example.com",
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTLSPolicyConfig with CACertRefs",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{
				CACertRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
			},
			wantErrors: []string{},
		},
		{
			name:        "invalid BackendTLSPolicyConfig with missing fields",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{},
			wantErrors:  []string{"spec.tls.hostname in body should be at least 1 chars long", "must specify either CACertRefs or WellKnownCACerts"},
		},
		{
			name: "invalid BackendTLSPolicyConfig with both CACertRefs and WellKnownCACerts",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{
				CACertRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				WellKnownCACerts: ptrTo(gatewayv1a2.WellKnownCACertType("System")),
				Hostname:         "foo.example.com",
			},

			wantErrors: []string{"must not contain both CACertRefs and WellKnownCACerts"},
		},
		{
			name: "invalid BackendTLSPolicyConfig with Unsupported value for WellKnownCACerts",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{
				WellKnownCACerts: ptrTo(gatewayv1a2.WellKnownCACertType("bar")),
				Hostname:         "foo.example.com",
			},
			wantErrors: []string{"supported values: \"System\""},
		},
		{
			name: "invalid BackendTLSPolicyConfig with empty Hostname field",
			routeConfig: gatewayv1a2.BackendTLSPolicyConfig{
				CACertRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "",
			},
			wantErrors: []string{"spec.tls.hostname in body should be at least 1 chars long"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1a2.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1a2.BackendTLSPolicySpec{
					TargetRef: gatewayv1a2.LocalPolicyTargetReferenceWithSectionName{
						LocalPolicyTargetReference: gatewayv1a2.LocalPolicyTargetReference{
							Group: "group",
							Kind:  "kind",
							Name:  "name",
						},
					},
					TLS: tc.routeConfig,
				},
			}
			validateBackendTLSPolicy(t, route, tc.wantErrors)
		})
	}
}

func validateBackendTLSPolicy(t *testing.T, route *gatewayv1a2.BackendTLSPolicy, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
