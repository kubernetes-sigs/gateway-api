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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
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

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			ref := NewGatewayRef(test.nsn, test.listenerNames...)
			require.IsType(t, GatewayRef{}, ref)
			if test.listenerNames == nil {
				require.Len(t, ref.listenerNames, 1)
				assert.Empty(t, string(*ref.listenerNames[0]))
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
			obj:  &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 20},
				{Type: "FakeCondition2", ObservedGeneration: 20},
				{Type: "FakeCondition3", ObservedGeneration: 20},
			},
		},
		{
			name: "a StaleConditionType condition is exempt from the generation check",
			obj:  &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 20},
				{Type: StaleConditionType, ObservedGeneration: 3},
			},
		},
		{
			name: "the StaleConditionType exemption does not extend to other conditions",
			obj:  &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 19},
				{Type: StaleConditionType, ObservedGeneration: 3},
			},
			expected: fmt.Errorf("expected observedGeneration to be updated to 20 for all conditions, only 1/2 were updated. stale conditions are: FakeCondition1 (generation 19)"),
		},
		{
			name: "conditions where one does not match the generation fail verification",
			obj:  &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
			conditions: []metav1.Condition{
				{Type: "FakeCondition1", ObservedGeneration: 20},
				{Type: "FakeCondition2", ObservedGeneration: 19},
				{Type: "FakeCondition3", ObservedGeneration: 20},
			},
			expected: fmt.Errorf("expected observedGeneration to be updated to 20 for all conditions, only 2/3 were updated. stale conditions are: FakeCondition2 (generation 19)"),
		},
		{
			name: "conditions where most do not match the generation fail verification",
			obj:  &gatewayv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "fake-gateway", Generation: 20}},
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

	for i := range tests {
		test := tests[i]
		t.Run(test.name, func(t *testing.T) {
			err := ConditionsHaveLatestObservedGeneration(test.obj, test.conditions)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestHTTPRouteMustBeAcceptedAndResolved(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, InstallGatewayV1(scheme))

	routeNN := types.NamespacedName{Name: "test-route", Namespace: "default"}
	gatewayNN := types.NamespacedName{Name: "test-gateway", Namespace: "default"}

	gwNamespace := gatewayv1.Namespace(gatewayNN.Namespace)
	route := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      routeNN.Name,
			Namespace: routeNN.Namespace,
		},
		Status: gatewayv1.HTTPRouteStatus{
			RouteStatus: gatewayv1.RouteStatus{
				Parents: []gatewayv1.RouteParentStatus{
					{
						ParentRef: gatewayv1.ParentReference{
							Name:      gatewayv1.ObjectName(gatewayNN.Name),
							Namespace: &gwNamespace,
						},
						Conditions: []metav1.Condition{
							{
								Type:               string(gatewayv1.RouteConditionAccepted),
								Status:             metav1.ConditionTrue,
								Reason:             string(gatewayv1.RouteReasonAccepted),
								LastTransitionTime: metav1.Now(),
							},
							{
								Type:               string(gatewayv1.RouteConditionResolvedRefs),
								Status:             metav1.ConditionTrue,
								Reason:             string(gatewayv1.RouteReasonResolvedRefs),
								LastTransitionTime: metav1.Now(),
							},
						},
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()

	timeoutConfig := config.TimeoutConfig{
		HTTPRouteMustHaveCondition: 5 * time.Second,
		DefaultPollInterval:        100 * time.Millisecond,
	}

	HTTPRouteMustBeAcceptedAndResolved(t, c, timeoutConfig, routeNN, gatewayNN)

	DeleteHTTPRoute(t, c, routeNN)

	err := c.Get(context.TODO(), routeNN, &gatewayv1.HTTPRoute{})
	assert.True(t, apierrors.IsNotFound(err))
}

// -----------------------------------------------------------------------------
// Test - Private Functions
// -----------------------------------------------------------------------------

func Test_listenersMatch(t *testing.T) {
	tests := []struct {
		name     string
		expected []gatewayv1.ListenerStatus
		actual   []gatewayv1.ListenerStatus
		want     bool
	}{
		{
			name: "listeners do not match if a different number of actual and expected listeners are provided",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("GRPCRoute"),
						},
					},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected empty and actual is non empty",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected and actual are equal",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SupportedKinds: expected and actual are equal values, Group pointers are different",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: new(gatewayv1.Group("gateway.networking.k8s.io")),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "SupportedKinds: expected kind not found in actual",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&v1alpha2.GroupVersion.Group),
							Kind:  gatewayv1.Kind("GRPCRoute"),
						},
					},
				},
			},
			want: false,
		},
		{
			name: "SupportedKinds: expected is a subset of actual",
			expected: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			actual: []gatewayv1.ListenerStatus{
				{
					SupportedKinds: []gatewayv1.RouteGroupKind{
						{
							Group: (*gatewayv1.Group)(&v1alpha2.GroupVersion.Group),
							Kind:  gatewayv1.Kind("GRPCRoute"),
						},
						{
							Group: (*gatewayv1.Group)(&gatewayv1.GroupVersion.Group),
							Kind:  gatewayv1.Kind("HTTPRoute"),
						},
					},
				},
			},
			want: true,
		},
		{
			name: "expected and actual can be in different orders",
			expected: []gatewayv1.ListenerStatus{
				{Name: "listener-2"},
				{Name: "listener-3"},
				{Name: "listener-1"},
			},
			actual: []gatewayv1.ListenerStatus{
				{Name: "listener-1"},
				{Name: "listener-2"},
				{Name: "listener-3"},
			},
			want: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.want, gatewayListenersMatch(t, test.expected, test.actual))
		})
	}
}

// TestRouteMustHaveParentsIgnoresStaleControllerEntries seeds an entry that is
// stale on purpose: it sits at generation 1 while the Route is at generation 2,
// so the wait only succeeds if entries owned by StaleControllerName are left out
// of the observedGeneration check.
func TestRouteMustHaveParentsIgnoresStaleControllerEntries(t *testing.T) {
	const ownController = gatewayv1.GatewayController("example.com/gateway-controller")

	scheme := runtime.NewScheme()
	require.NoError(t, InstallGatewayV1(scheme))

	routeNN := types.NamespacedName{Name: "test-route", Namespace: "default"}
	gwNamespace := gatewayv1.Namespace("default")

	acceptedCondition := func(observedGeneration int64) []metav1.Condition {
		return []metav1.Condition{{
			Type:               string(gatewayv1.RouteConditionAccepted),
			Status:             metav1.ConditionTrue,
			Reason:             string(gatewayv1.RouteReasonAccepted),
			ObservedGeneration: observedGeneration,
			LastTransitionTime: metav1.Now(),
		}}
	}

	route := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:       routeNN.Name,
			Namespace:  routeNN.Namespace,
			Generation: 2,
		},
		Status: gatewayv1.HTTPRouteStatus{
			RouteStatus: gatewayv1.RouteStatus{
				Parents: []gatewayv1.RouteParentStatus{
					{
						ParentRef: gatewayv1.ParentReference{
							Name:      "test-gateway",
							Namespace: &gwNamespace,
						},
						ControllerName: ownController,
						Conditions:     acceptedCondition(2),
					},
					{
						ParentRef: gatewayv1.ParentReference{
							Name:      "unmanaged-gateway",
							Namespace: &gwNamespace,
						},
						ControllerName: StaleControllerName,
						Conditions:     acceptedCondition(1),
					},
				},
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()

	timeoutConfig := config.TimeoutConfig{
		RouteMustHaveParents: 2 * time.Second,
		DefaultPollInterval:  100 * time.Millisecond,
	}

	HTTPRouteMustHaveParents(t, c, timeoutConfig, routeNN, []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{
				Name:      "test-gateway",
				Namespace: &gwNamespace,
			},
			ControllerName: ownController,
			Conditions: []metav1.Condition{{
				Type:   string(gatewayv1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
			}},
		},
	}, true)
}

func Test_staleParentStatus(t *testing.T) {
	const (
		ownController   = gatewayv1.GatewayController("example.com/gateway-controller")
		otherController = gatewayv1.GatewayController("example.com/other-controller")
	)

	route := &gatewayv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Generation: 2}}

	parent := func(controller gatewayv1.GatewayController, observedGeneration int64) gatewayv1.RouteParentStatus {
		return gatewayv1.RouteParentStatus{
			ParentRef:      gatewayv1.ParentReference{Name: "test-gateway"},
			ControllerName: controller,
			Conditions: []metav1.Condition{{
				Type:               string(gatewayv1.RouteConditionAccepted),
				Status:             metav1.ConditionTrue,
				Reason:             string(gatewayv1.RouteReasonAccepted),
				ObservedGeneration: observedGeneration,
			}},
		}
	}

	tests := []struct {
		name      string
		parents   []gatewayv1.RouteParentStatus
		wantOwner gatewayv1.GatewayController
	}{
		{
			name:    "every entry has caught up",
			parents: []gatewayv1.RouteParentStatus{parent(ownController, 2), parent(otherController, 2)},
		},
		{
			name:      "own entry is behind",
			parents:   []gatewayv1.RouteParentStatus{parent(ownController, 1)},
			wantOwner: ownController,
		},
		{
			name:    "sentinel entry is exempt while it is behind",
			parents: []gatewayv1.RouteParentStatus{parent(ownController, 2), parent(StaleControllerName, 1)},
		},
		{
			name:      "the exemption does not extend to other foreign entries",
			parents:   []gatewayv1.RouteParentStatus{parent(ownController, 2), parent(otherController, 1)},
			wantOwner: otherController,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stale, err := staleParentStatus(route, tt.parents)

			if tt.wantOwner == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			require.Equal(t, tt.wantOwner, stale.ControllerName)
		})
	}
}

// TestBackendTLSPolicyMustHaveConditionIgnoresStaleAncestors seeds an ancestor
// that is stale on purpose: it sits at generation 1 while the policy is at
// generation 2, so the wait only succeeds if ancestors owned by
// StaleControllerName are left out of the observedGeneration check.
func TestBackendTLSPolicyMustHaveConditionIgnoresStaleAncestors(t *testing.T) {
	const ownController = gatewayv1.GatewayController("example.com/gateway-controller")

	scheme := runtime.NewScheme()
	require.NoError(t, InstallGatewayV1(scheme))

	policyNN := types.NamespacedName{Name: "test-policy", Namespace: "default"}
	gwNN := types.NamespacedName{Name: "test-gateway", Namespace: "default"}
	gwNamespace := gatewayv1.Namespace(gwNN.Namespace)

	ancestor := func(name gatewayv1.ObjectName, controller gatewayv1.GatewayController, observedGeneration int64) gatewayv1.PolicyAncestorStatus {
		return gatewayv1.PolicyAncestorStatus{
			AncestorRef: gatewayv1.ParentReference{
				Name:      name,
				Namespace: &gwNamespace,
			},
			ControllerName: controller,
			Conditions: []metav1.Condition{{
				Type:               string(gatewayv1.PolicyConditionAccepted),
				Status:             metav1.ConditionTrue,
				Reason:             string(gatewayv1.PolicyReasonAccepted),
				ObservedGeneration: observedGeneration,
				LastTransitionTime: metav1.Now(),
			}},
		}
	}

	policy := &gatewayv1.BackendTLSPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:       policyNN.Name,
			Namespace:  policyNN.Namespace,
			Generation: 2,
		},
		Status: gatewayv1.PolicyStatus{
			Ancestors: []gatewayv1.PolicyAncestorStatus{
				ancestor(gatewayv1.ObjectName(gwNN.Name), ownController, 2),
				ancestor("unmanaged-gateway", StaleControllerName, 1),
			},
		},
	}

	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(policy).Build()

	timeoutConfig := config.TimeoutConfig{
		HTTPRouteMustHaveCondition: 2 * time.Second,
		DefaultPollInterval:        100 * time.Millisecond,
	}

	BackendTLSPolicyMustHaveCondition(t, c, timeoutConfig, policyNN, gwNN, metav1.Condition{
		Type:   string(gatewayv1.PolicyConditionAccepted),
		Status: metav1.ConditionTrue,
		Reason: string(gatewayv1.PolicyReasonAccepted),
	})

	BackendTLSPolicyMustHaveLatestConditions(t, policy)
}

// TestRouteMustHaveParentsChecksTheStatusItJustRead pins the order of the two
// steps inside the poll: the status has to be read before the
// observedGeneration check runs against it. When the read came last, the first
// iteration checked an empty slice, so a Route whose conditions still carried a
// stale observedGeneration satisfied the wait right away and the check never
// ran at all on the common fast path.
func TestRouteMustHaveParentsChecksTheStatusItJustRead(t *testing.T) {
	const ownController = gatewayv1.GatewayController("example.com/gateway-controller")

	scheme := runtime.NewScheme()
	require.NoError(t, InstallGatewayV1(scheme))

	routeNN := types.NamespacedName{Name: "test-route", Namespace: "default"}
	gwNamespace := gatewayv1.Namespace("default")

	route := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:       routeNN.Name,
			Namespace:  routeNN.Namespace,
			Generation: 2,
		},
		Status: gatewayv1.HTTPRouteStatus{
			RouteStatus: gatewayv1.RouteStatus{
				Parents: []gatewayv1.RouteParentStatus{{
					ParentRef: gatewayv1.ParentReference{
						Name:      "test-gateway",
						Namespace: &gwNamespace,
					},
					ControllerName: ownController,
					Conditions: []metav1.Condition{{
						Type:               string(gatewayv1.RouteConditionAccepted),
						Status:             metav1.ConditionTrue,
						Reason:             string(gatewayv1.RouteReasonAccepted),
						ObservedGeneration: 1,
						LastTransitionTime: metav1.Now(),
					}},
				}},
			},
		},
	}

	var reads int
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(route).
		WithInterceptorFuncs(interceptor.Funcs{
			Get: func(ctx context.Context, cli client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				if err := cli.Get(ctx, key, obj, opts...); err != nil {
					return err
				}

				// The implementation catches up, but only after the first read.
				reads++
				if reads > 1 {
					obj.(*gatewayv1.HTTPRoute).Status.Parents[0].Conditions[0].ObservedGeneration = 2
				}

				return nil
			},
		}).
		Build()

	timeoutConfig := config.TimeoutConfig{
		RouteMustHaveParents: 2 * time.Second,
		DefaultPollInterval:  100 * time.Millisecond,
	}

	HTTPRouteMustHaveParents(t, c, timeoutConfig, routeNN, []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{
				Name:      "test-gateway",
				Namespace: &gwNamespace,
			},
			ControllerName: ownController,
			Conditions: []metav1.Condition{{
				Type:   string(gatewayv1.RouteConditionAccepted),
				Status: metav1.ConditionTrue,
			}},
		},
	}, true)

	require.Greater(t, reads, 1, "wait was satisfied by the first read, leaving the stale observedGeneration unchecked")
}
