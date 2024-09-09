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
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestBackendTLSPolicyValidation(t *testing.T) {
	tests := []struct {
		name        string
		wantErrors  []string
		routeConfig gatewayv1a3.BackendTLSPolicyValidation
	}{
		{
			name: "valid BackendTLSPolicyValidation with WellKnownCACertificates",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				WellKnownCACertificates: ptrTo(gatewayv1a3.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTLSPolicyValidation with CACertificateRefs",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
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
			name:        "invalid BackendTLSPolicyValidation with missing fields",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{},
			wantErrors:  []string{"spec.validation.hostname in body should be at least 1 chars long", "must specify either CACertificateRefs or WellKnownCACertificates"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with both CACertificateRefs and WellKnownCACertificates",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				WellKnownCACertificates: ptrTo(gatewayv1a3.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},

			wantErrors: []string{"must not contain both CACertificateRefs and WellKnownCACertificates"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with Unsupported value for WellKnownCACertificates",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				WellKnownCACertificates: ptrTo(gatewayv1a3.WellKnownCACertificatesType("bar")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{"supported values: \"System\""},
		},
		{
			name: "invalid BackendTLSPolicyValidation with empty Hostname field",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "",
			},
			wantErrors: []string{"spec.validation.hostname in body should be at least 1 chars long"},
		},
		{
			name: "valid BackendTLSPolicyValidation with SubjectAltName type Hostname",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type:     "Hostname",
						Hostname: "foo.example.com",
					},
				},
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTLSPolicyValidation with SubjectAltName type URI",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type: "URI",
						URI:  "spiffe://mycluster.example",
					},
				},
			},
			wantErrors: []string{},
		},
		{
			name: "invalid BackendTLSPolicyValidation with SubjectAltName type Hostname and empty Hostname field",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type:     "Hostname",
						Hostname: "",
					},
				},
			},
			wantErrors: []string{"SubjectAltName element must contain Hostname, if Type is set to Hostname"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with SubjectAltName type URI and non-empty Hostname field",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type:     "URI",
						Hostname: "foo.example.com",
					},
				},
			},
			wantErrors: []string{"SubjectAltName element must not contain Hostname, if Type is not set to Hostname"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with SubjectAltName type URI and empty URI field",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type: "URI",
						URI:  "",
					},
				},
			},
			wantErrors: []string{"SubjectAltName element must contain URI, if Type is set to URI"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with SubjectAltName type Hostname and non-empty URI field",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type: "Hostname",
						URI:  "test",
					},
				},
			},
			wantErrors: []string{"SubjectAltName element must not contain URI, if Type is not set to URI"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with SubjectAltName type Hostname and both Hostname and URI specified",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type:     "Hostname",
						Hostname: "foo.example.com",
						URI:      "test",
					},
				},
			},
			wantErrors: []string{"SubjectAltName element must not contain URI, if Type is not set to URI"},
		},
		{
			name: "invalid BackendTLSPolicyValidation incorrect URI SAN",
			routeConfig: gatewayv1a3.BackendTLSPolicyValidation{
				CACertificateRefs: []v1beta1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1a3.SubjectAltName{
					{
						Type: "URI",
						URI:  "foo.example.com",
					},
				},
			},
			wantErrors: []string{"spec.validation.subjectAltNames[0].uri in body should match '^(([^:/?#]+):)(//([^/?#]*))([^?#]*)(\\?([^#]*))?(#(.*))?'"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			route := &gatewayv1a3.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1a3.BackendTLSPolicySpec{
					TargetRefs: []gatewayv1a2.LocalPolicyTargetReferenceWithSectionName{
						{
							gatewayv1a2.LocalPolicyTargetReference{
								Group: "group",
								Kind:  "kind",
								Name:  "name",
							},
							// SectionName cannot contain capital letters.
							ptrTo(gatewayv1a2.SectionName("section")),
						},
					},
					Validation: tc.routeConfig,
				},
			}
			validateBackendTLSPolicy(t, route, tc.wantErrors)
		})
	}
}

func validateBackendTLSPolicy(t *testing.T, route *gatewayv1a3.BackendTLSPolicy, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, route)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", route.Namespace, route.Name), err, missingErrorStrings)
	}
}
