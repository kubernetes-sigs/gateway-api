//go:build experimental
// +build experimental

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBackendTLSPolicyValidationClusterTrustBundle(t *testing.T) {
	tests := []struct {
		name             string
		wantErrors       []string
		policyValidation gatewayv1.BackendTLSPolicyValidation
	}{
		{
			name: "valid BackendTLSPolicyValidation with ClusterTrustBundle via caCertificateRefs",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "certificates.k8s.io",
						Kind:  "ClusterTrustBundle",
						Name:  "example.com:internal-signer:v1",
					},
				},
				Hostname: "foo.example.com",
			},
		},
		{
			name: "invalid BackendTLSPolicyValidation with CACertificateRefs and WellKnownCACertificates",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "certificates.k8s.io",
						Kind:  "ClusterTrustBundle",
						Name:  "example.com:internal-signer:v1",
					},
				},
				WellKnownCACertificates: new(gatewayv1.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{"exactly one of the fields in [caCertificateRefs wellKnownCACertificates] must be set"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with no trust source",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{Hostname: "foo.example.com"},
			wantErrors:       []string{"exactly one of the fields in [caCertificateRefs wellKnownCACertificates] must be set"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := &gatewayv1.BackendTLSPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
					Namespace: metav1.NamespaceDefault,
				},
				Spec: gatewayv1.BackendTLSPolicySpec{
					TargetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{
						{
							LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
								Group: "group",
								Kind:  "kind",
								Name:  "name",
							},
							SectionName: new(gatewayv1.SectionName("section")),
						},
					},
					Validation: tc.policyValidation,
				},
			}
			validateBackendTLSPolicy(t, policy, tc.wantErrors)
		})
	}
}

// TestBackendTLSPolicyValidationEmptyCACertificateRefs ensures an explicitly empty
// caCertificateRefs list is not accepted as a trust source. The ExactlyOneOf rule only
// tests for the presence of the field, so MinItems is what rejects it. An unstructured
// object is used because the typed field is omitempty, so an empty slice never reaches
// the API server.
func TestBackendTLSPolicyValidationEmptyCACertificateRefs(t *testing.T) {
	policy := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": gatewayv1.GroupVersion.String(),
		"kind":       "BackendTLSPolicy",
		"metadata": map[string]any{
			"name":      fmt.Sprintf("foo-%v", time.Now().UnixNano()),
			"namespace": metav1.NamespaceDefault,
		},
		"spec": map[string]any{
			"targetRefs": []any{map[string]any{
				"group":       "group",
				"kind":        "kind",
				"name":        "name",
				"sectionName": "section",
			}},
			"validation": map[string]any{
				"hostname":          "foo.example.com",
				"caCertificateRefs": []any{},
			},
		},
	}}

	err := k8sClient.Create(context.Background(), policy)
	if err == nil {
		t.Fatalf("expected BackendTLSPolicy with an empty caCertificateRefs list to be rejected")
	}
	wantError := "should have at least 1 items"
	if !celErrorStringMatches(err.Error(), wantError) {
		t.Errorf("Unexpected response while creating BackendTLSPolicy; got err=\n%v\n;missing string within error=%q", err, wantError)
	}
}
