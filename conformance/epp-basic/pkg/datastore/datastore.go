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
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	logutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common/observability/logging"
	podutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/util/pod"
)

type Endpoint struct {
	NamespacedName types.NamespacedName
	PodName        string
	Address        string
	Port           string
	Labels         map[string]string
}

type EndpointPool struct {
	Selector    map[string]string
	TargetPorts []int
	Namespace   string
}

var (
	errPoolNotSynced = errors.New("InferencePool is not initialized in data store")
	AllPodsPredicate = func(_ *Endpoint) bool { return true }
)

const (
	// activePortsAnnotation is used to specify which ports on a pod should be considered
	// as active for inference traffic. The value should be a comma-separated list of port numbers.
	// Example: "8000,8001,8002"
	activePortsAnnotation = "inference.networking.k8s.io/active-ports"
)

// The datastore is a local cache of relevant data for the given InferencePool (currently all pulled from k8s-api)
type Datastore interface {
	// InferencePool operations
	// PoolSet sets the given pool in datastore. If the given pool has different label selector than the previous pool
	// that was stored, the function triggers a resync of the pods to keep the datastore updated. If the given pool
	// is nil, this call triggers the datastore.Clear() function.
	PoolSet(ctx context.Context, reader client.Reader, endpointPool *EndpointPool) error
	PoolGet() (*EndpointPool, error)
	PoolHasSynced() bool
	PoolLabelsMatch(podLabels map[string]string) bool

	// PodList lists pods matching the given predicate.
	PodList(predicate func(*Endpoint) bool) []*Endpoint
	PodUpdateOrAddIfNotExist(ctx context.Context, pod *corev1.Pod) bool
	PodDelete(podName string)

	// Clears the store state, happens when the pool gets deleted.
	Clear()
}

// compile-time type assertion
var _ Datastore = &datastore{}

// NewDatastore creates a new data store.
func NewDatastore(parentCtx context.Context) *datastore {
	return &datastore{
		parentCtx: parentCtx,
		pool:      nil,
		mu:        sync.RWMutex{},
		pods:      &sync.Map{},
	}
}

type datastore struct {
	parentCtx context.Context
	mu        sync.RWMutex
	pool      *EndpointPool
	pods      *sync.Map
}

func (ds *datastore) WithEndpointPool(pool *EndpointPool) *datastore {
	ds.pool = pool
	return ds
}

func (ds *datastore) Clear() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.pool = nil
	ds.pods.Clear()
}

// /// Pool APIs ///
func (ds *datastore) PoolSet(ctx context.Context, reader client.Reader, endpointPool *EndpointPool) error {
	if endpointPool == nil {
		ds.Clear()
		return nil
	}
	logger := log.FromContext(ctx)
	ds.mu.Lock()
	defer ds.mu.Unlock()

	oldEndpointPool := ds.pool
	ds.pool = endpointPool

	selectorChanged := oldEndpointPool == nil || !labels.Equals(oldEndpointPool.Selector, endpointPool.Selector)
	targetPortsChanged := oldEndpointPool != nil && !slices.Equal(oldEndpointPool.TargetPorts, endpointPool.TargetPorts)

	if selectorChanged || targetPortsChanged {
		logger.V(logutil.DEFAULT).Info("Updating endpoints", "selector", endpointPool.Selector, "targetPortsChanged", targetPortsChanged)
		// A full resync is required to address the following cases:
		// 1) At startup, the pod events may get processed before the pool is synced with the datastore,
		//    and hence they will not be added to the store since pool selector is not known yet
		// 2) If the selector on the pool was updated, then we will not get any pod events, and so we need
		//    to resync the whole pool: remove pods in the store that don't match the new selector and add
		//    the ones that may have existed already to the store.
		// 3) If the targetPorts changed, we need to resync to remove orphaned rank endpoints that no longer
		//    exist in the new targetPorts configuration.
		if err := ds.podResyncAll(ctx, reader); err != nil {
			return fmt.Errorf("failed to update pods according to the pool selector - %w", err)
		}
	}

	return nil
}

func (ds *datastore) PoolGet() (*EndpointPool, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.pool == nil {
		return nil, errPoolNotSynced
	}
	return ds.pool, nil
}

func (ds *datastore) PoolHasSynced() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.pool != nil
}

func (ds *datastore) PoolLabelsMatch(podLabels map[string]string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.pool == nil {
		return false
	}
	poolSelector := labels.SelectorFromSet(ds.pool.Selector)
	podSet := labels.Set(podLabels)
	return poolSelector.Matches(podSet)
}

// /// Pods/endpoints APIs ///
// TODO: add a flag for callers to specify the staleness threshold for metrics.
// ref: https://github.com/kubernetes-sigs/gateway-api-inference-extension/pull/1046#discussion_r2246351694
func (ds *datastore) PodList(predicate func(*Endpoint) bool) []*Endpoint {
	res := []*Endpoint{}

	ds.pods.Range(func(k, v any) bool {
		ep := v.(*Endpoint)
		if predicate(ep) {
			res = append(res, ep)
		}
		return true
	})

	return res
}

func (ds *datastore) PodUpdateOrAddIfNotExist(ctx context.Context, pod *corev1.Pod) bool {
	ds.mu.RLock()
	pool := ds.pool
	ds.mu.RUnlock()

	return ds.podUpdateOrAddIfNotExist(ctx, pod, pool)
}

func (ds *datastore) podUpdateOrAddIfNotExist(ctx context.Context, pod *corev1.Pod, pool *EndpointPool) bool {
	if pool == nil {
		return true
	}

	labels := make(map[string]string, len(pod.GetLabels()))
	maps.Copy(labels, pod.GetLabels())

	pods := []*Endpoint{}
	activePorts := extractActivePorts(pod, pool.TargetPorts)
	for idx, port := range pool.TargetPorts {
		if !activePorts.Has(port) {
			continue
		}
		pods = append(pods,
			&Endpoint{
				NamespacedName: createEndpointNamespacedName(pod, idx),
				PodName:        pod.Name,
				Address:        pod.Status.PodIP,
				Port:           strconv.Itoa(port),
				Labels:         labels,
			})
	}

	if len(pods) == 0 {
		logger := log.FromContext(ctx)
		logger.V(logutil.VERBOSE).Info("No container ports match pool targetPorts, pod will not receive traffic",
			"pod", pod.Name, "namespace", pod.Namespace, "targetPorts", pool.TargetPorts)
	}

	result := true
	for _, endpointMetadata := range pods {
		_, ok := ds.pods.Load(endpointMetadata.NamespacedName)
		if !ok {
			result = false
		}
		ds.pods.Store(endpointMetadata.NamespacedName, endpointMetadata)
	}

	// remove endpoints that are no longer active in the pool
	for idx, port := range pool.TargetPorts {
		if activePorts.Has(port) {
			continue
		}

		namespacedName := createEndpointNamespacedName(pod, idx)
		if _, ok := ds.pods.Load(namespacedName); ok {
			ds.pods.Delete(namespacedName)
		}
	}

	return result
}

func (ds *datastore) PodDelete(podName string) {
	ds.pods.Range(func(k, v any) bool {
		ep := v.(*Endpoint)
		if ep.PodName == podName {
			ds.pods.Delete(k)
		}
		return true
	})
}

func (ds *datastore) podResyncAll(ctx context.Context, reader client.Reader) error {
	logger := log.FromContext(ctx)
	podList := &corev1.PodList{}
	if err := reader.List(ctx, podList, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(ds.pool.Selector),
		Namespace:     ds.pool.Namespace,
	}); err != nil {
		return fmt.Errorf("failed to list pods - %w", err)
	}

	activeEndpoints := sets.New[types.NamespacedName]()
	for _, pod := range podList.Items {
		if !podutil.IsPodReady(&pod) {
			continue
		}
		namespacedName := types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}
		for idx := range ds.pool.TargetPorts {
			activeEndpoints.Insert(createEndpointNamespacedName(&pod, idx))
		}
		if !ds.podUpdateOrAddIfNotExist(ctx, &pod, ds.pool) {
			logger.V(logutil.DEFAULT).Info("Pod added", "name", namespacedName)
		} else {
			logger.V(logutil.DEFAULT).Info("Pod already exists", "name", namespacedName)
		}
	}

	ds.pods.Range(func(k, v any) bool {
		ep := v.(*Endpoint)
		endpointName := ep.NamespacedName
		if !activeEndpoints.Has(endpointName) {
			logger.V(logutil.VERBOSE).Info("Removing endpoint", "endpoint", endpointName)
			ds.pods.Delete(k)
		}
		return true
	})

	return nil
}

// extractActivePorts extracts the active ports from a pod's annotations.
func extractActivePorts(pod *corev1.Pod, targetPorts []int) sets.Set[int] {
	allPorts := sets.New(targetPorts...)
	annotations := pod.GetAnnotations()
	portsAnnotation, ok := annotations[activePortsAnnotation]
	if !ok {
		return allPorts
	}

	activePorts := sets.New[int]()
	portStrs := strings.SplitSeq(portsAnnotation, ",")
	for portStr := range portStrs {
		var portNum int
		_, err := fmt.Sscanf(strings.TrimSpace(portStr), "%d", &portNum)
		if err == nil && portNum > 0 && allPorts.Has(portNum) {
			activePorts.Insert(portNum)
		}
	}
	return activePorts
}

// createEndpointNamespacedName creates a namespaced name for an endpoint based on pod and rank index.
// This ensures consistent naming between PodUpdateOrAddIfNotExist and podResyncAll.
func createEndpointNamespacedName(pod *corev1.Pod, idx int) types.NamespacedName {
	return types.NamespacedName{
		Name:      pod.Name + "-rank-" + strconv.Itoa(idx),
		Namespace: pod.Namespace,
	}
}
