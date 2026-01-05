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

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBackendTLSPolicyTargetRefs(t *testing.T) {
	tests := []struct {
		name       string
		wantErrors []string
		targetRefs []gatewayv1.LocalPolicyTargetReferenceWithSectionName
	}{
		{
			name:       "invalid because duplicate target refs without section name",
			wantErrors: []string{"sectionName must be unique when targetRefs includes 2 or more references to the same target"},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
			}},
		},
		{
			name:       "invalid because duplicate target refs with only one section name",
			wantErrors: []string{"sectionName must be specified when targetRefs includes 2 or more references to the same target"},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example2",
				},
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "invalid because duplicate target refs with duplicate section names",
			wantErrors: []string{"sectionName must be unique when targetRefs includes 2 or more references to the same target"},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("bar")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid single targetRef without sectionName",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
			}},
		},
		{
			name:       "valid single targetRef with sectionName",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because duplicate target refs with different section names",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("bar")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("jin")),
			}},
		},
		{
			name:       "valid because duplicate target refs with different names",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example2",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example3",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because duplicate target refs with different kinds",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("NotService"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
		},
		{
			name:       "valid because duplicate target refs with different groups",
			wantErrors: []string{},
			targetRefs: []gatewayv1.LocalPolicyTargetReferenceWithSectionName{{
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group(corev1.GroupName),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}, {
				LocalPolicyTargetReference: gatewayv1.LocalPolicyTargetReference{
					Group: gatewayv1.Group("svc.other.io"),
					Kind:  gatewayv1.Kind("Service"),
					Name:  "example",
				},
				SectionName: ptrTo(gatewayv1.SectionName("foo")),
			}},
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
					TargetRefs: tc.targetRefs,
					Validation: gatewayv1.BackendTLSPolicyValidation{
						WellKnownCACertificates: ptrTo(gatewayv1.WellKnownCACertificatesType("System")),
						Hostname:                "foo.example.com",
					},
				},
			}
			validateBackendTLSPolicy(t, policy, tc.wantErrors)
		})
	}
}

func TestBackendTLSPolicyValidation(t *testing.T) {
	tests := []struct {
		name             string
		wantErrors       []string
		policyValidation gatewayv1.BackendTLSPolicyValidation
	}{
		{
			name: "valid BackendTLSPolicyValidation with WellKnownCACertificates",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				WellKnownCACertificates: ptrTo(gatewayv1.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{},
		},
		{
			name: "valid BackendTLSPolicyValidation with CACertificateRefs",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
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
			name:             "invalid BackendTLSPolicyValidation with missing fields",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{},
			wantErrors:       []string{"spec.validation.hostname in body should be at least 1 chars long", "must specify either CACertificateRefs or WellKnownCACertificates"},
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
				WellKnownCACertificates: ptrTo(gatewayv1.WellKnownCACertificatesType("System")),
				Hostname:                "foo.example.com",
			},

			wantErrors: []string{"must not contain both CACertificateRefs and WellKnownCACertificates"},
		},
		{
			name: "invalid BackendTLSPolicyValidation with Unsupported value for WellKnownCACertificates",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				WellKnownCACertificates: ptrTo(gatewayv1.WellKnownCACertificatesType("bar")),
				Hostname:                "foo.example.com",
			},
			wantErrors: []string{"supported values: \"System\""},
		},
		{
			name: "invalid BackendTLSPolicyValidation with empty Hostname field",
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
			policyValidation: gatewayv1.BackendTLSPolicyValidation{
				CACertificateRefs: []gatewayv1.LocalObjectReference{
					{
						Group: "group",
						Kind:  "kind",
						Name:  "name",
					},
				},
				Hostname: "foo.example.com",
				SubjectAltNames: []gatewayv1.SubjectAltName{
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
							// SectionName cannot contain capital letters.
							SectionName: ptrTo(gatewayv1.SectionName("section")),
						},
					},
					Validation: tc.policyValidation,
				},
			}
			validateBackendTLSPolicy(t, policy, tc.wantErrors)
		})
	}
}

func validateBackendTLSPolicy(t *testing.T, policy *gatewayv1.BackendTLSPolicy, wantErrors []string) {
	t.Helper()

	ctx := context.Background()
	err := k8sClient.Create(ctx, policy)

	if (len(wantErrors) != 0) != (err != nil) {
		t.Fatalf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;want error=%v", fmt.Sprintf("%v/%v", policy.Namespace, policy.Name), err, wantErrors)
	}

	var missingErrorStrings []string
	for _, wantError := range wantErrors {
		if !celErrorStringMatches(err.Error(), wantError) {
			missingErrorStrings = append(missingErrorStrings, wantError)
		}
	}
	if len(missingErrorStrings) != 0 {
		t.Errorf("Unexpected response while creating BackendTLSPolicy %q; got err=\n%v\n;missing strings within error=%q", fmt.Sprintf("%v/%v", policy.Namespace, policy.Name), err, missingErrorStrings)
	}
}
