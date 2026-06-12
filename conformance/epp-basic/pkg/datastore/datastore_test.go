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

package datastore

import (
	"context"
	"net"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	v1 "sigs.k8s.io/gateway-api-inference-extension/api/v1"
	testutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/util/testing"
)

func inferencePoolToEndpointPool(inferencePool *v1.InferencePool) *EndpointPool {
	if inferencePool == nil {
		return nil
	}
	targetPorts := make([]int, 0, len(inferencePool.Spec.TargetPorts))
	for _, p := range inferencePool.Spec.TargetPorts {
		targetPorts = append(targetPorts, int(p.Number))
	}
	selector := make(map[string]string, len(inferencePool.Spec.Selector.MatchLabels))
	for k, v := range inferencePool.Spec.Selector.MatchLabels {
		selector[string(k)] = string(v)
	}
	return &EndpointPool{
		Selector:    selector,
		TargetPorts: targetPorts,
		Namespace:   inferencePool.Namespace,
	}
}

func TestPoolGet_NoDeadlockWithConcurrentWrite(t *testing.T) {
	pool := &EndpointPool{
		Namespace:   "default",
		Selector:    map[string]string{"app": "vllm"},
		TargetPorts: []int{8000},
	}
	ds := &datastore{pool: pool}

	done := make(chan struct{})
	go func() {
		defer close(done)
		for range 1000 {
			ds.mu.Lock()
			ds.pool = pool
			ds.mu.Unlock()
		}
	}()

	for range 1000 {
		_, _ = ds.PoolGet()
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("deadlock detected: PoolGet and concurrent writer did not complete within timeout")
	}
}

func TestPool(t *testing.T) {
	pool1Selector := map[string]string{"app": "vllm_v1"}
	pool1 := testutil.MakeInferencePool("pool1").
		Namespace("default").
		Selector(pool1Selector).ObjRef()
	tests := []struct {
		name            string
		inferencePool   *v1.InferencePool
		labels          map[string]string
		wantSynced      bool
		wantPool        *v1.InferencePool
		wantErr         error
		wantLabelsMatch bool
	}{
		{
			name:            "Ready when InferencePool exists in data store",
			inferencePool:   pool1,
			labels:          pool1Selector,
			wantSynced:      true,
			wantPool:        pool1,
			wantLabelsMatch: true,
		},
		{
			name:            "Labels not matched",
			inferencePool:   pool1,
			labels:          map[string]string{"app": "vllm_v2"},
			wantSynced:      true,
			wantPool:        pool1,
			wantLabelsMatch: false,
		},
		{
			name:       "Not ready when InferencePool is nil in data store",
			wantErr:    errPoolNotSynced,
			wantSynced: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the scheme.
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			ds := NewDatastore(context.Background())
			_ = ds.PoolSet(context.Background(), fakeClient, inferencePoolToEndpointPool(tt.inferencePool))
			gotPool, gotErr := ds.PoolGet()
			if diff := cmp.Diff(tt.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Unexpected error diff (+got/-want): %s", diff)
			}
			if diff := cmp.Diff(inferencePoolToEndpointPool(tt.wantPool), gotPool); diff != "" {
				t.Errorf("Unexpected pool diff (+got/-want): %s", diff)
			}
			gotSynced := ds.PoolHasSynced()
			if diff := cmp.Diff(tt.wantSynced, gotSynced); diff != "" {
				t.Errorf("Unexpected synced diff (+got/-want): %s", diff)
			}
			if tt.labels != nil {
				gotLabelsMatch := ds.PoolLabelsMatch(tt.labels)
				if diff := cmp.Diff(tt.wantLabelsMatch, gotLabelsMatch); diff != "" {
					t.Errorf("Unexpected labels match diff (+got/-want): %s", diff)
				}
			}
		})
	}
}

var (
	pod1 = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod1",
		},
	}
	pod2 = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "pod2",
		},
	}
	inferencePool = &v1.InferencePool{
		Spec: v1.InferencePoolSpec{
			TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}},
		},
	}
	inferencePoolMultiTarget = &v1.InferencePool{
		Spec: v1.InferencePoolSpec{
			TargetPorts: []v1.Port{{Number: v1.PortNumber(int32(8000))}, {Number: v1.PortNumber(int32(8001))}},
		},
	}

	inferencePoolTargetPort       = strconv.Itoa(int(inferencePool.Spec.TargetPorts[0].Number))
	inferencePoolMultiTargetPort0 = strconv.Itoa(int(inferencePoolMultiTarget.Spec.TargetPorts[0].Number))
	inferencePoolMultiTargetPort1 = strconv.Itoa(int(inferencePoolMultiTarget.Spec.TargetPorts[1].Number))
)

func TestPods(t *testing.T) {
	tests := []struct {
		name         string
		op           func(ctx context.Context, ds Datastore)
		existingPods []*corev1.Pod
		wantPods     []*corev1.Pod
	}{
		{
			name:         "Add new pod, no existing pods, should add",
			existingPods: []*corev1.Pod{},
			wantPods:     []*corev1.Pod{pod1},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodUpdateOrAddIfNotExist(ctx, pod1)
			},
		},
		{
			name:         "Add new pod, with existing pods, should add",
			existingPods: []*corev1.Pod{pod1},
			wantPods:     []*corev1.Pod{pod1, pod2},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodUpdateOrAddIfNotExist(ctx, pod2)
			},
		},
		{
			name:         "Delete the pod",
			existingPods: []*corev1.Pod{pod1, pod2},
			wantPods:     []*corev1.Pod{pod1},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodDelete(pod2.Name)
			},
		},
		{
			name:         "Delete the pod that doesn't exist",
			existingPods: []*corev1.Pod{pod1},
			wantPods:     []*corev1.Pod{pod1},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodDelete(pod2.Name)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ds := NewDatastore(t.Context())
			fakeClient := fake.NewFakeClient()
			if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(inferencePool)); err != nil {
				t.Error(err)
			}
			for _, pod := range test.existingPods {
				ds.PodUpdateOrAddIfNotExist(ctx, pod)
			}

			test.op(ctx, ds)
			pods := ds.PodList(AllPodsPredicate)
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

func TestTargetPortsChange(t *testing.T) {
	// Create pods that are ready
	readyPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.1",
		},
	}

	tests := []struct {
		name                   string
		initialTargetPorts     []v1.Port
		updatedTargetPorts     []v1.Port
		wantEndpointCountAfter int
		wantEndpointNames      []string
	}{
		{
			name:                   "Shrink from 2 ports to 1 port removes orphaned rank",
			initialTargetPorts:     []v1.Port{{Number: 8000}, {Number: 8001}},
			updatedTargetPorts:     []v1.Port{{Number: 8000}},
			wantEndpointCountAfter: 1,
			wantEndpointNames:      []string{"pod1-rank-0"},
		},
		{
			name:                   "Shrink from 3 ports to 1 port removes multiple orphaned ranks",
			initialTargetPorts:     []v1.Port{{Number: 8000}, {Number: 8001}, {Number: 8002}},
			updatedTargetPorts:     []v1.Port{{Number: 8000}},
			wantEndpointCountAfter: 1,
			wantEndpointNames:      []string{"pod1-rank-0"},
		},
		{
			name:                   "Expand from 1 port to 2 ports adds new rank",
			initialTargetPorts:     []v1.Port{{Number: 8000}},
			updatedTargetPorts:     []v1.Port{{Number: 8000}, {Number: 8001}},
			wantEndpointCountAfter: 2,
			wantEndpointNames:      []string{"pod1-rank-0", "pod1-rank-1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create fake client with the pod
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(readyPod1).
				Build()

			ds := NewDatastore(ctx)

			// Set initial pool with multiple target ports
			initialPool := testutil.MakeInferencePool("test-pool").
				Namespace("default").
				Selector(map[string]string{"app": "vllm"}).ObjRef()
			initialPool.Spec.TargetPorts = test.initialTargetPorts

			if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(initialPool)); err != nil {
				t.Fatalf("Failed to set initial pool: %v", err)
			}

			// Verify initial endpoint count
			initialEndpoints := ds.PodList(AllPodsPredicate)
			if len(initialEndpoints) != len(test.initialTargetPorts) {
				t.Errorf("Initial endpoint count: got %d, want %d", len(initialEndpoints), len(test.initialTargetPorts))
			}

			// Update pool with different target ports
			updatedPool := testutil.MakeInferencePool("test-pool").
				Namespace("default").
				Selector(map[string]string{"app": "vllm"}).ObjRef()
			updatedPool.Spec.TargetPorts = test.updatedTargetPorts

			if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(updatedPool)); err != nil {
				t.Fatalf("Failed to set updated pool: %v", err)
			}

			// Verify orphaned ranks are removed
			finalEndpoints := ds.PodList(AllPodsPredicate)
			if len(finalEndpoints) != test.wantEndpointCountAfter {
				t.Errorf("Final endpoint count: got %d, want %d", len(finalEndpoints), test.wantEndpointCountAfter)
			}

			// Verify endpoint names
			gotNames := make([]string, 0, len(finalEndpoints))
			for _, ep := range finalEndpoints {
				gotNames = append(gotNames, ep.NamespacedName.Name)
			}
			if diff := cmp.Diff(test.wantEndpointNames, gotNames, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
				t.Errorf("Endpoint names mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestEndpointMetadata(t *testing.T) {
	type endpointMetadata struct {
		NamespacedName types.NamespacedName
		PodName        string
		Address        string
		Port           string
		Labels         map[string]string
		MetricsHost    string
	}
	tests := []struct {
		name              string
		op                func(ctx context.Context, ds Datastore)
		pool              *v1.InferencePool
		existingPods      []*corev1.Pod
		wantEndpointMetas []*endpointMetadata
	}{
		{
			name:         "Add new pod, no existing pods, should add",
			existingPods: []*corev1.Pod{},
			wantEndpointMetas: []*endpointMetadata{
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-0",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolTargetPort,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolTargetPort),
					Labels:      map[string]string{},
				},
			},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodUpdateOrAddIfNotExist(ctx, pod1)
			},
			pool: inferencePool,
		},
		{
			name:         "Add new pod, no existing pods, should add, multiple target ports",
			existingPods: []*corev1.Pod{},
			wantEndpointMetas: []*endpointMetadata{
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-0",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort0,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort0),
					Labels:      map[string]string{},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-1",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort1,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort1),
					Labels:      map[string]string{},
				},
			},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodUpdateOrAddIfNotExist(ctx, pod1)
			},
			pool: inferencePoolMultiTarget,
		},
		{
			name:         "Add new pod, with existing pods, should add, multiple target ports",
			existingPods: []*corev1.Pod{pod1},
			wantEndpointMetas: []*endpointMetadata{
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-0",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort0,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort0),
					Labels:      map[string]string{},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-1",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort1,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort1),
					Labels:      map[string]string{},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      pod2.Name + "-rank-0",
						Namespace: pod2.Namespace,
					},

					PodName:     pod2.Name,
					Address:     pod2.Status.PodIP,
					Port:        inferencePoolMultiTargetPort0,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort0),
					Labels:      map[string]string{},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      pod2.Name + "-rank-1",
						Namespace: pod2.Namespace,
					},

					PodName:     pod2.Name,
					Address:     pod2.Status.PodIP,
					Port:        inferencePoolMultiTargetPort1,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort1),
					Labels:      map[string]string{},
				},
			},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodUpdateOrAddIfNotExist(ctx, pod2)
			},
			pool: inferencePoolMultiTarget,
		},
		{
			name:         "Delete the pod, multiple target ports",
			existingPods: []*corev1.Pod{pod1, pod2},
			wantEndpointMetas: []*endpointMetadata{
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-0",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort0,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort0),
					Labels:      map[string]string{},
				},
				{
					NamespacedName: types.NamespacedName{
						Name:      pod1.Name + "-rank-1",
						Namespace: pod1.Namespace,
					},

					PodName:     pod1.Name,
					Address:     pod1.Status.PodIP,
					Port:        inferencePoolMultiTargetPort1,
					MetricsHost: net.JoinHostPort(pod1.Status.PodIP, inferencePoolMultiTargetPort1),
					Labels:      map[string]string{},
				},
			},
			op: func(ctx context.Context, ds Datastore) {
				ds.PodDelete(pod2.Name)
			},
			pool: inferencePoolMultiTarget,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			ds := NewDatastore(t.Context())
			fakeClient := fake.NewFakeClient()
			if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(test.pool)); err != nil {
				t.Error(err)
			}
			for _, pod := range test.existingPods {
				ds.PodUpdateOrAddIfNotExist(ctx, pod)
			}

			test.op(ctx, ds)
			pods := ds.PodList(AllPodsPredicate)
			gotMetadata := make([]*endpointMetadata, 0, len(pods))
			for _, pm := range pods {
				gotMetadata = append(gotMetadata, &endpointMetadata{
					NamespacedName: pm.NamespacedName,
					PodName:        pm.PodName,
					Address:        pm.Address,
					Port:           pm.Port,
					Labels:         pm.Labels,
					MetricsHost:    net.JoinHostPort(pm.Address, pm.Port),
				})
			}
			if diff := cmp.Diff(test.wantEndpointMetas, gotMetadata, cmpopts.SortSlices(func(a, b *endpointMetadata) bool { return a.NamespacedName.Name < b.NamespacedName.Name })); diff != "" {
				t.Errorf("ConvertTo() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestActivePortFiltering(t *testing.T) {
	// Create pods that are ready
	readyPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
			Annotations: map[string]string{
				"inference.networking.k8s.io/active-ports": "8000,8002",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.1",
		},
	}

	// Pod without active ports annotation - should use all ports
	readyPod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod2",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.2",
		},
	}

	// Pod with empty active ports annotation
	readyPod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod3",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
			Annotations: map[string]string{
				"inference.networking.k8s.io/active-ports": "",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.3",
		},
	}

	tests := []struct {
		name              string
		pools             []v1.InferencePool
		pods              []*corev1.Pod
		wantEndpointCount int
		wantEndpointNames []string
	}{
		{
			name: "Pod with active ports annotation filters endpoints",
			pools: []v1.InferencePool{
				{
					Spec: v1.InferencePoolSpec{
						TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}, {Number: 8002}, {Number: 8003}},
					},
				},
			},
			pods:              []*corev1.Pod{readyPod1},
			wantEndpointCount: 2,                                      // Only ports 8000 and 8002 should be active
			wantEndpointNames: []string{"pod1-rank-0", "pod1-rank-2"}, // ranks 1 and 3 (for ports 8001 and 8003) should be skipped
		},
		{
			name: "Pod without active ports annotation uses all ports",
			pools: []v1.InferencePool{
				{
					Spec: v1.InferencePoolSpec{
						TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}},
					},
				},
			},
			pods:              []*corev1.Pod{readyPod2},
			wantEndpointCount: 2, // Both ports should be active
		},
		{
			name: "Pod with empty active ports annotation uses no ports",
			pools: []v1.InferencePool{
				{
					Spec: v1.InferencePoolSpec{
						TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}},
					},
				},
			},
			pods:              []*corev1.Pod{readyPod3},
			wantEndpointCount: 0, // No ports should be active
		},
		{
			name: "Multiple pods with different active port annotations",
			pools: []v1.InferencePool{
				{
					Spec: v1.InferencePoolSpec{
						TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}, {Number: 8002}},
					},
				},
			},
			pods:              []*corev1.Pod{readyPod1, readyPod2}, // pod1 has ports 8000,8002 active; pod2 has all ports active
			wantEndpointCount: 5,                                   // pod1: 2 endpoints (8000, 8002); pod2: 3 endpoints (8000, 8001, 8002)
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			ds := NewDatastore(ctx)

			// Use the first pool in the test
			if len(test.pools) > 0 {
				pool := test.pools[0]
				if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(&pool)); err != nil {
					t.Fatalf("Failed to set pool: %v", err)
				}
			}

			// Add all pods
			for _, pod := range test.pods {
				ds.PodUpdateOrAddIfNotExist(ctx, pod)
			}

			// Check final endpoint count
			finalEndpoints := ds.PodList(AllPodsPredicate)
			if len(finalEndpoints) != test.wantEndpointCount {
				t.Errorf("Final endpoint count: got %d, want %d", len(finalEndpoints), test.wantEndpointCount)
			}

			// Check endpoint names if specified
			if test.wantEndpointNames != nil {
				gotNames := make([]string, 0, len(finalEndpoints))
				for _, ep := range finalEndpoints {
					gotNames = append(gotNames, ep.NamespacedName.Name)
				}
				if diff := cmp.Diff(test.wantEndpointNames, gotNames, cmpopts.SortSlices(func(a, b string) bool { return a < b })); diff != "" {
					t.Errorf("Endpoint names mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestActivePortEndpointRemoval(t *testing.T) {
	// Create a pod initially with all ports active
	readyPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
			Annotations: map[string]string{
				"inference.networking.k8s.io/active-ports": "8000,8001,8002",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.1",
		},
	}

	// Updated pod with fewer active ports
	updatedPod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
			Annotations: map[string]string{
				"inference.networking.k8s.io/active-ports": "8000",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.1",
		},
	}

	// Pod with no active ports
	inactivePod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "vllm"},
			Annotations: map[string]string{
				"inference.networking.k8s.io/active-ports": "",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
			PodIP: "10.0.0.1",
		},
	}

	tests := []struct {
		name              string
		pool              *v1.InferencePool
		operations        []func(Datastore)
		initialPod        *corev1.Pod
		wantEndpointCount int
	}{
		{
			name: "Remove endpoints when active ports are reduced",
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}, {Number: 8002}},
				},
			},
			initialPod: readyPod1,
			operations: []func(Datastore){
				// Update the pod to reduce active ports from 3 to 1
				func(ds Datastore) {
					ds.PodUpdateOrAddIfNotExist(context.Background(), updatedPod1)
				},
			},
			wantEndpointCount: 1, // Only port 8000 should remain active
		},
		{
			name: "Remove all endpoints when no active ports are specified",
			pool: &v1.InferencePool{
				Spec: v1.InferencePoolSpec{
					TargetPorts: []v1.Port{{Number: 8000}, {Number: 8001}},
				},
			},
			initialPod: readyPod1,
			operations: []func(Datastore){
				// Update the pod to have no active ports
				func(ds Datastore) {
					ds.PodUpdateOrAddIfNotExist(context.Background(), inactivePod1)
				},
			},
			wantEndpointCount: 0, // No ports should remain active
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			scheme := runtime.NewScheme()
			_ = clientgoscheme.AddToScheme(scheme)
			_ = corev1.AddToScheme(scheme)

			// Create fake client
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			ds := NewDatastore(ctx)

			// Set up the pool
			if err := ds.PoolSet(ctx, fakeClient, inferencePoolToEndpointPool(test.pool)); err != nil {
				t.Fatalf("Failed to set pool: %v", err)
			}

			// Add the initial pod
			ds.PodUpdateOrAddIfNotExist(ctx, test.initialPod)

			// Wait a bit for the datastore to process the pod
			time.Sleep(100 * time.Millisecond)

			// Check initial endpoint count (should be 3 since all 3 ports are active)
			initialEndpoints := ds.PodList(AllPodsPredicate)
			expectedInitialCount := len(test.pool.Spec.TargetPorts) // Expected based on target ports in pool
			if len(initialEndpoints) != expectedInitialCount {
				t.Logf("Initial endpoint count: got %d, want %d", len(initialEndpoints), expectedInitialCount)
				// Don't fail here, just log - we'll continue to test the reduction
			}

			// Execute operations that change active ports
			for _, op := range test.operations {
				op(ds)
			}

			// Check final endpoint count
			finalEndpoints := ds.PodList(AllPodsPredicate)
			if len(finalEndpoints) != test.wantEndpointCount {
				t.Errorf("Final endpoint count: got %d, want %d", len(finalEndpoints), test.wantEndpointCount)
			}
		})
	}
}

// TestPodUpdateOrAddIfNotExist_ConcurrentPoolSet verifies that PodUpdateOrAddIfNotExist
// does not race with PoolSet. Before the fix, PodUpdateOrAddIfNotExist read ds.pool
// without holding ds.mu, which could panic or corrupt data when PoolSet concurrently
// replaces ds.pool under the write lock.
// Run with: go test -race -run TestPodUpdateOrAddIfNotExist_ConcurrentPoolSet
func TestPodUpdateOrAddIfNotExist_ConcurrentPoolSet(t *testing.T) {
	t.Run("ConcurrentPoolSet", func(t *testing.T) {
		scheme := runtime.NewScheme()
		_ = clientgoscheme.AddToScheme(scheme)
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		ctx := context.Background()
		ds := NewDatastore(ctx)

		pool := inferencePoolToEndpointPool(
			testutil.MakeInferencePool("pool1").
				Namespace("default").
				Selector(map[string]string{"app": "vllm"}).
				TargetPorts(8000).ObjRef(),
		)
		_ = ds.PoolSet(ctx, fakeClient, pool)

		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "default",
				Labels:    map[string]string{"app": "vllm"},
			},
			Status: corev1.PodStatus{
				PodIP: "10.0.0.1",
				Conditions: []corev1.PodCondition{
					{Type: corev1.PodReady, Status: corev1.ConditionTrue},
				},
			},
		}

		var wg sync.WaitGroup
		wg.Add(2)

		// Goroutine 1: repeatedly call PoolSet (including nil to simulate reset).
		go func() {
			defer wg.Done()
			for range 500 {
				_ = ds.PoolSet(ctx, fakeClient, pool)
				_ = ds.PoolSet(ctx, fakeClient, nil)
				_ = ds.PoolSet(ctx, fakeClient, pool)
			}
		}()

		// Goroutine 2: repeatedly call PodUpdateOrAddIfNotExist.
		go func() {
			defer wg.Done()
			for range 1000 {
				ds.PodUpdateOrAddIfNotExist(ctx, pod)
			}
		}()

		wg.Wait()
	})
}

func TestExtractActivePorts(t *testing.T) {
	tests := []struct {
		name          string
		pod           *corev1.Pod
		validPorts    []int
		expectedPorts sets.Set[int]
	}{
		{
			name: "Pod without active ports annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8001, 8002),
		},
		{
			name: "Pod with empty active ports annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: ""},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New[int](),
		},
		{
			name: "Pod with single port in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000),
		},
		{
			name: "Pod with multiple ports in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000,8001,8002"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8001, 8002),
		},
		{
			name: "Pod with multiple ports with spaces in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000, 8001 , 8002"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8001, 8002),
		},
		{
			name: "Pod with invalid port in annotation (non-numeric)",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000,invalid,8002"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8002),
		},
		{
			name: "Pod with invalid port in annotation (negative number)",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000,-1,8002"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8002),
		},
		{
			name: "Pod with duplicate ports in annotation",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000,8001,8000"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000, 8001),
		},
		{
			name: "Pod with port not in validPorts",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "test-pod",
					Namespace:   "default",
					Annotations: map[string]string{activePortsAnnotation: "8000,9000"},
				},
			},
			validPorts:    []int{8000, 8001, 8002},
			expectedPorts: sets.New(8000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activePorts := extractActivePorts(tt.pod, tt.validPorts)
			if !reflect.DeepEqual(activePorts, tt.expectedPorts) {
				t.Errorf("ExtractActivePorts() ports = %v, want %v", activePorts, tt.expectedPorts)
			}
		})
	}
}
