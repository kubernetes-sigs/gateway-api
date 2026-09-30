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
	"fmt"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// TestBackendTLSPolicyValidationTrustSource covers the standard channel rules that
// require exactly one of CACertificateRefs or WellKnownCACertificates. The
// experimental channel replaces them with a rule that also accepts
// ClusterTrustBundleRef, see backendtlspolicy_experimental_test.go.
func TestBackendTLSPolicyValidationTrustSource(t *testing.T) {
	tests := []struct {
		name             string
		wantErrors       []string
		policyValidation gatewayv1.BackendTLSPolicyValidation
	}{
		{
			name:             "invalid BackendTLSPolicyValidation with no trust source",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{Hostname: "foo.example.com"},
			wantErrors:       []string{"must specify either CACertificateRefs or WellKnownCACertificates"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with both CACertificateRefs and WellKnownCACertificates",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				WellKnownCACertificates: new(gatewayv1.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{"must not contain both CACertificateRefs and WellKnownCACertificates"},
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
