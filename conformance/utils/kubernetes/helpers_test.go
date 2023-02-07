package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestSupportedKindsMatch(t *testing.T) {
	tests := []struct {
		name     string
		expected []v1beta1.ListenerStatus
		actual   []v1beta1.ListenerStatus
		want     bool
	}{
		{
			name: "expected empty and actual is non empty",
			expected: []v1beta1.ListenerStatus{
				{
					Name:           v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{{
						Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
						Kind:  v1beta1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			want: false,
		},
		{
			name: "expected and actual are equal",
			expected: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{{
						Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
						Kind:  v1beta1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{{
						Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
						Kind:  v1beta1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			want: true,
		},
		{
			name: "expected kind not found in actual",
			expected: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{{
						Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
						Kind:  v1beta1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{
						{
							Group: (*v1beta1.Group)(&v1alpha2.GroupVersion.Group),
							Kind:  v1beta1.Kind("GRPCRoute"),
						},
					},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			want: false,
		},
		{
			name: "expected is a subset of actual",
			expected: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
					SupportedKinds: []v1beta1.RouteGroupKind{{
						Group: (*v1beta1.Group)(&v1beta1.GroupVersion.Group),
						Kind:  v1beta1.Kind("HTTPRoute"),
					}},
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
				},
			},
			actual: []v1beta1.ListenerStatus{
				{
					Name: v1beta1.SectionName("https"),
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
					Conditions: []metav1.Condition{{
						Type:   string(v1beta1.ListenerConditionResolvedRefs),
						Status: metav1.ConditionFalse,
						Reason: string(v1beta1.ListenerReasonRefNotPermitted),
					}},
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
