/*
Copyright 2022 The Kubernetes Authors.

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

package kubernetes

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// -----------------------------------------------------------------------------
// Test - Public Functions
// -----------------------------------------------------------------------------

func TestNewGatewayRef(t *testing.T) {
	tests := []struct {
		name          string
		nsn           types.NamespacedName
		listenerNames []string
	}{
		{
			name: "verifying the contents of a GatewayRef with no provided listeners",
			nsn:  types.NamespacedName{Namespace: corev1.NamespaceDefault, Name: "fake-gateway"},
		},
		{
			name:          "verifying the contents of a GatewayRef listeners with one listener provided",
			nsn:           types.NamespacedName{Namespace: corev1.NamespaceDefault, Name: "fake-gateway"},
			listenerNames: []string{"fake-listener-1"},
		},
		{
			name: "verifying the contents of a GatewayRef listeners with multiple listeners provided",
			nsn:  types.NamespacedName{Namespace: corev1.NamespaceDefault, Name: "fake-gateway"},
			listenerNames: []string{
				"fake-listener-1",
				"fake-listener-2",
				"fake-listener-3",
			},
		},
	}

	for i := 0; i < len(tests); i++ {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			ref := NewGatewayRef(test.nsn, test.listenerNames...)
			require.IsType(t, GatewayRef{}, ref)
			if test.listenerNames == nil {
				require.Len(t, ref.listenerNames, 1)
				assert.Equal(t, "", string(*ref.listenerNames[0]))
			} else {
				require.Len(t, ref.listenerNames, len(test.listenerNames))
				for i := 0; i < len(ref.listenerNames); i++ {
					assert.Equal(t, test.listenerNames[i], string(*ref.listenerNames[i]))
				}
			}
			assert.Equal(t, test.nsn, ref.NamespacedName)
		})
	}
}

func TestVerifyConditionsMatchGeneration(t *testing.T) {
	tests := []struct {
		name       string
		obj        metav1.Object
		conditions []metav1.Condition
		expected   error
	}{
		{},
		{
			name: "if no conditions are provided this technically passes verification",
		},
		{
			name: "conditions where all match the generation pass verification",
			obj:  &v1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 20},
				{Type: "FakeCondition2", ObservedGeneration: 20},
				{Type: "FakeCondition3", ObservedGeneration: 20},
			},
		},
		{
			name: "conditions where one does not match the generation fail verification",
			obj:  &v1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 20},
				{Type: "FakeCondition2", ObservedGeneration: 19},
				{Type: "FakeCondition3", ObservedGeneration: 20},
			},
			expected: fmt.Errorf("expected observedGeneration to be updated to 20 for all conditions, only 2/3 were updated. stale conditions are: FakeCondition2 (generation 19)"),
		},
		{
			name: "conditions where most do not match the generation fail verification",
			obj:  &v1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 18},
				{Type: "FakeCondition2", ObservedGeneration: 18},
				{Type: "FakeCondition3", ObservedGeneration: 14},
				{Type: "FakeCondition4", ObservedGeneration: 20},
				{Type: "FakeCondition5", ObservedGeneration: 16},
				{Type: "FakeCondition6", ObservedGeneration: 15},
			},
			expected: fmt.Errorf("expected observedGeneration to be updated to 20 for all conditions, only 1/6 were updated. stale conditions are: FakeCondition1 (generation 18), FakeCondition2 (generation 18), FakeCondition3 (generation 14), FakeCondition5 (generation 16), FakeCondition6 (generation 15)"),
		},
	}

	for i := 0; i < len(tests); i++ {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := ConditionsHaveLatestObservedGeneration(test.obj, test.conditions)
			assert.Equal(t, test.expected, err)
		})
	}
}

// -----------------------------------------------------------------------------
// Test - Private Functions
// -----------------------------------------------------------------------------

func Test_listenersMatch(t *testing.T) {
	tests := []struct {
		name     string
		expected []v1beta1.ListenerStatus
		actual   []v1beta1.ListenerStatus
		want     bool
	}{
		{
			name: "listeners do not match if a different number of actual and expected listeners are provided",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("GRPCRoute"),
						},
					},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected empty and actual is non empty",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected and actual are equal",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SupportedKinds: expected and actual are equal values, Group pointers are different",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(pointer.String("gateway.networking.k8s.io")),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SupportedKinds: expected kind not found in actual",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1alpha2.GroupVersion.Group),
							Kind:  v1beta1.Kind("GRPCRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected is a subset of actual",
			expected: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1alpha2.GroupVersion.Group),
							Kind:  v1beta1.Kind("GRPCRoute"),
						},
						{
							Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
							Kind:  v1beta1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "expected and actual can be in different orders",
			expected: []v1beta1.ListenerStatus{
				{Name: "listener-2"},
				{Name: "listener-3"},
				{Name: "listener-1"},
			},
			actual: []v1beta1.ListenerStatus{
				{Name: "listener-1"},
				{Name: "listener-2"},
				{Name: "listener-3"},
			},
			want: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, listenersMatch(t, test.expected, test.actual))
		})
	}
}
