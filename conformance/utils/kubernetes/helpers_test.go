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
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"

	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

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
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, listenersMatch(t, test.expected, test.actual))
		})
	}
}
