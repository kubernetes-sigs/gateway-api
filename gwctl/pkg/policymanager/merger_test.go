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

package policymanager

import (
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestMergePoliciesOfSimilarKind(t *testing.T) {
	timeSmall := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}.String()
	timeLarge := metav1.Time{Time: time.Now()}.String()
	policies := []Policy{
		{
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "foo.com/v1",
					"kind":       "HealthCheckPolicy",
					"metadata": map[string]interface{}{
						"name":              "health-check-1",
						"creationTimestamp": timeSmall,
					},
					"spec": map[string]interface{}{
						"override": map[string]interface{}{
							"key1": "a",
							"key3": "b",
						},
						"default": map[string]interface{}{
							"key2": "d",
							"key4": "e",
							"key5": "c",
						},
					},
				},
			},
			inherited: true,
		},
		{
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "foo.com/v1",
					"kind":       "HealthCheckPolicy",
					"metadata": map[string]interface{}{
						"name":              "health-check-2",
						"creationTimestamp": timeLarge,
					},
					"spec": map[string]interface{}{
						"override": map[string]interface{}{
							"key1": "f",
						},
						"default": map[string]interface{}{
							"key2": "i",
							"key4": "j",
						},
					},
				},
			},
			inherited: true,
		},
		{
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "bar.com/v1",
					"kind":       "TimeoutPolicy",
					"metadata": map[string]interface{}{
						"name": "timeout-policy-1",
					},
					"spec": map[string]interface{}{
						"condition": "path=/def",
						"seconds":   float64(30),
						"targetRef": map[string]interface{}{
							"kind": "Namespace",
							"name": "default",
						},
					},
				},
			},
		},
		{
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "bar.com/v1",
					"kind":       "TimeoutPolicy",
					"metadata": map[string]interface{}{
						"name": "timeout-policy-2",
					},
					"spec": map[string]interface{}{
						"condition": "path=/abc",
						"seconds":   float64(60),
						"targetRef": map[string]interface{}{
							"kind": "Namespace",
							"name": "default",
						},
					},
				},
			},
		},
	}

	want := map[PolicyCrdID]Policy{
		PolicyCrdID("HealthCheckPolicy.foo.com"): {
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "foo.com/v1",
					"kind":       "HealthCheckPolicy",
					"metadata": map[string]interface{}{
						"name":              "health-check-1",
						"creationTimestamp": timeSmall,
					},
					"spec": map[string]interface{}{
						"override": map[string]interface{}{
							"key1": "f",
							"key3": "b",
						},
						"default": map[string]interface{}{
							"key2": "d",
							"key4": "e",
							"key5": "c",
						},
					},
				},
			},
			inherited: true,
		},
		PolicyCrdID("TimeoutPolicy.bar.com"): {
			u: unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "bar.com/v1",
					"kind":       "TimeoutPolicy",
					"metadata": map[string]interface{}{
						"name": "timeout-policy-1",
					},
					"spec": map[string]interface{}{
						"condition": "path=/def",
						"seconds":   float64(30),
						"targetRef": map[string]interface{}{
							"kind": "Namespace",
							"name": "default",
						},
					},
				},
			},
		},
	}

	got, err := MergePoliciesOfSimilarKind(policies)
	if err != nil {
		t.Fatalf("MergePoliciesOfSimilarKind returned err=%v; want no error", err)
	}
	cmpopts := cmp.Exporter(func(t reflect.Type) bool {
		return t == reflect.TypeOf(Policy{})
	})
	if diff := cmp.Diff(want, got, cmpopts); diff != "" {
		t.Errorf("MergePoliciesOfSimilarKind returned unexpected diff (-want, +got):\n%v", diff)
	}
}

func TestMergePoliciesOfDifferentHierarchy(t *testing.T) {
	testCases := []struct {
		name           string
		parentPolicies []Policy
		childPolicies  []Policy

		wantMergedPolicies []Policy
		wantErr            bool
	}{
		{
			name: "parent and child both have overrides and defaults",
			parentPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-1",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "parentValue1",
								"key2": "parentValue2",
							},
							"default": map[string]interface{}{
								"key4": "parentValue4",
								"key5": "parentValue5",
							},
						},
					},
				},
			}},
			childPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-2",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "childValue1",
								"key3": "childValue3",
							},
							"default": map[string]interface{}{
								"key4": "childValue4",
								"key6": "childValue6",
							},
						},
					},
				},
			}},
			wantMergedPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-2",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "parentValue1",
								"key2": "parentValue2",
								"key3": "childValue3",
							},
							"default": map[string]interface{}{
								"key4": "childValue4",
								"key5": "parentValue5",
								"key6": "childValue6",
							},
						},
					},
				},
			}},
		},
		{
			name: "parent has defaults, child has overrides",
			parentPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-1",
						},
						"spec": map[string]interface{}{
							"default": map[string]interface{}{
								"key4": "parentValue4",
								"key5": "parentValue5",
							},
						},
					},
				},
			}},
			childPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-2",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "childValue1",
								"key3": "childValue3",
							},
						},
					},
				},
			}},
			wantMergedPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-2",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "childValue1",
								"key3": "childValue3",
							},
							"default": map[string]interface{}{
								"key4": "parentValue4",
								"key5": "parentValue5",
							},
						},
					},
				},
			}},
		},
		{
			name: "policies of different kind do not intersect with each other",
			parentPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "foo.com/v1",
						"kind":       "HealthCheckPolicy",
						"metadata": map[string]interface{}{
							"name": "health-check-1",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "a",
								"key3": "b",
							},
							"default": map[string]interface{}{
								"key2": "d",
								"key4": "e",
								"key5": "c",
							},
						},
					},
				},
			}},
			childPolicies: []Policy{{
				inherited: true,
				u: unstructured.Unstructured{
					Object: map[string]interface{}{
						"apiVersion": "bar.com/v1",
						"kind":       "TimeoutPolicy",
						"metadata": map[string]interface{}{
							"name": "timeout-policy-2",
						},
						"spec": map[string]interface{}{
							"override": map[string]interface{}{
								"key1": "childValue1",
								"key3": "childValue3",
							},
							"default": map[string]interface{}{
								"key4": "childValue4",
								"key6": "childValue6",
							},
						},
					},
				},
			}},
			wantMergedPolicies: []Policy{
				{
					inherited: true,
					u: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "bar.com/v1",
							"kind":       "TimeoutPolicy",
							"metadata": map[string]interface{}{
								"name": "timeout-policy-2",
							},
							"spec": map[string]interface{}{
								"override": map[string]interface{}{
									"key1": "childValue1",
									"key3": "childValue3",
								},
								"default": map[string]interface{}{
									"key4": "childValue4",
									"key6": "childValue6",
								},
							},
						},
					},
				},
				{
					inherited: true,
					u: unstructured.Unstructured{
						Object: map[string]interface{}{
							"apiVersion": "foo.com/v1",
							"kind":       "HealthCheckPolicy",
							"metadata": map[string]interface{}{
								"name": "health-check-1",
							},
							"spec": map[string]interface{}{
								"override": map[string]interface{}{
									"key1": "a",
									"key3": "b",
								},
								"default": map[string]interface{}{
									"key2": "d",
									"key4": "e",
									"key5": "c",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotMergedPolicies, err := MergePoliciesOfDifferentHierarchy(
				policySliceToMap(tc.parentPolicies),
				policySliceToMap(tc.childPolicies),
			)

			if (err != nil) != tc.wantErr {
				t.Fatalf("MergePoliciesOfDifferentHierarchy(...) returned err=%v; want err=%v", err, tc.wantErr)
			}

			// Use a custom transformer to only compare specific fields of the Policy
			// that we are interested in testing.
			cmpopts := cmp.Transformer("PolicyTransformer", func(p Policy) map[string]interface{} {
				return map[string]interface{}{
					"u":         p.u,
					"inherited": p.inherited,
				}
			})
			if diff := cmp.Diff(policySliceToMap(tc.wantMergedPolicies), gotMergedPolicies, cmpopts); diff != "" {
				t.Errorf("MergePoliciesOfDifferentHierarchy returned unexpected diff (-want, +got):\n%v", diff)
			}

		})
	}
}

func policySliceToMap(policies []Policy) map[PolicyCrdID]Policy {
	res := make(map[PolicyCrdID]Policy)
	for _, policy := range policies {
		res[policy.PolicyCrdID()] = policy
	}
	return res
}
