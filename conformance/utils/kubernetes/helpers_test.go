package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
