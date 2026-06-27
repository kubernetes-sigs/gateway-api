/*
Copyright 2025 The Kubernetes Authors.

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

package controller

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	v1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/util/pool"
	testutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/util/testing"
)

var (
	basePod1  = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}, Status: corev1.PodStatus{PodIP: "address-1"}}
	basePod2  = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}, Status: corev1.PodStatus{PodIP: "address-2"}}
	basePod3  = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod3"}, Status: corev1.PodStatus{PodIP: "address-3"}}
	basePod11 = &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}, Status: corev1.PodStatus{PodIP: "address-11"}}
)

func TestPodReconciler(t *testing.T) {
	tests := []struct {
		name         string
		pool         *v1.InferencePool
		existingPods []*corev1.Pod
		incomingPod  *corev1.Pod
		wantPods     []*corev1.Pod
		req          *ctrl.Request
	}{
		{
			name:         "Add new pod",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod3).
				Labels(map[string]string{"some-key": "some-val"}).
				ReadyCondition().ObjRef(),
			wantPods: []*corev1.Pod{basePod1, basePod2, basePod3},
		},
		{
			name:         "Update pod1 address",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod11).
				Labels(map[string]string{"some-key": "some-val"}).
				ReadyCondition().ObjRef(),
			wantPods: []*corev1.Pod{basePod11, basePod2},
		},
		{
			name:         "Delete pod with DeletionTimestamp",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod1).
				Labels(map[string]string{"some-key": "some-val"}).
				DeletionTimestamp().
				ReadyCondition().ObjRef(),
			wantPods: []*corev1.Pod{basePod2},
		},
		{
			name:         "Delete notfound pod",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			req:      &ctrl.Request{NamespacedName: types.NamespacedName{Name: "pod1"}},
			wantPods: []*corev1.Pod{basePod2},
		},
		{
			name:         "New pod, not ready, valid selector",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod3).
				Labels(map[string]string{"some-key": "some-val"}).ObjRef(),
			wantPods: []*corev1.Pod{basePod1, basePod2},
		},
		{
			name:         "Remove pod that does not match selector",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod1).
				Labels(map[string]string{"some-wrong-key": "some-val"}).
				ReadyCondition().ObjRef(),
			wantPods: []*corev1.Pod{basePod2},
		},
		{
			name:         "Remove pod that is not ready",
			existingPods: []*corev1.Pod{basePod1, basePod2},
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
					Selector: v1.LabelSelector{
						MatchLabels: map[v1.LabelKey]v1.LabelValue{
							"some-key": "some-val",
						},
					},
				},
			},
			incomingPod: testutil.FromBase(basePod1).
				Labels(map[string]string{"some-wrong-key": "some-val"}).
				ReadyCondition().ObjRef(),
			wantPods: []*corev1.Pod{basePod2},
		},
	}
	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {
			// Set up the scheme.
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			initialObjects := []client.Object{}
			if test.incomingPod != nil {
				initialObjects = append(initialObjects, test.incomingPod)
			}
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(initialObjects...).
				Build()

			// Configure the initial state of the datastore.
			store := datastore.NewDatastore(t.Context())
			_ = store.PoolSet(t.Context(), fakeClient, pool.InferencePoolToEndpointPool(test.pool))
			for _, pod := range test.existingPods {
				store.PodUpdateOrAddIfNotExist(t.Context(), pod)
			}

			podReconciler := &PodReconciler{Reader: fakeClient, Datastore: store}
			if test.req == nil {
				namespacedName := types.NamespacedName{Name: test.incomingPod.Name, Namespace: test.incomingPod.Namespace}
				test.req = &ctrl.Request{NamespacedName: namespacedName}
			}
			if _, err := podReconciler.Reconcile(context.Background(), *test.req); err != nil {
				t.Errorf("Unexpected InferencePool reconcile error: %v", err)
			}

			pods := store.PodList(datastore.AllPodsPredicate)
			gotPods := make([]*corev1.Pod, 0, len(pods))
			for _, pm := range pods {
				pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: pm.PodName, Namespace: pm.NamespacedName.Namespace}, Status: corev1.PodStatus{PodIP: pm.Address}}
				gotPods = append(gotPods, pod)
			}
			if !cmp.Equal(gotPods, test.wantPods, cmpopts.SortSlices(func(a, b *corev1.Pod) bool { return a.Name < b.Name })) {
				t.Errorf("got (%v) != want (%v);", gotPods, test.wantPods)
			}
		})
	}
}
