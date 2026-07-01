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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	logutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/common/observability/logging"
	"sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/datastore"
	podutil "sigs.k8s.io/gateway-api/conformance/epp-basic/pkg/util/pod"
)

type PodReconciler struct {
	client.Reader
	Datastore datastore.Datastore
}

func (c *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	if !c.Datastore.PoolHasSynced() {
		logger.V(logutil.VERBOSE).Info("InferencePool not yet synced, requeueing pod reconcile")
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	logger.V(logutil.VERBOSE).Info("Pod being reconciled")

	pod := &corev1.Pod{}
	if err := c.Get(ctx, req.NamespacedName, pod); err != nil {
		if apierrors.IsNotFound(err) {
			c.Datastore.PodDelete(req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("unable to get pod - %w", err)
	}

	c.updateDatastore(ctx, pod)
	return ctrl.Result{}, nil
}

func (c *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	filter := predicate.Funcs{
		CreateFunc: func(ce event.CreateEvent) bool {
			pod := ce.Object.(*corev1.Pod)
			return c.Datastore.PoolLabelsMatch(pod.GetLabels())
		},
		UpdateFunc: func(ue event.UpdateEvent) bool {
			oldPod := ue.ObjectOld.(*corev1.Pod)
			newPod := ue.ObjectNew.(*corev1.Pod)
			return c.Datastore.PoolLabelsMatch(oldPod.GetLabels()) || c.Datastore.PoolLabelsMatch(newPod.GetLabels())
		},
		DeleteFunc: func(de event.DeleteEvent) bool {
			pod := de.Object.(*corev1.Pod)
			return c.Datastore.PoolLabelsMatch(pod.GetLabels())
		},
		GenericFunc: func(ge event.GenericEvent) bool {
			pod := ge.Object.(*corev1.Pod)
			return c.Datastore.PoolLabelsMatch(pod.GetLabels())
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(filter).
		Complete(c)
}

func (c *PodReconciler) updateDatastore(ctx context.Context, pod *corev1.Pod) {
	logger := log.FromContext(ctx)
	if !podutil.IsPodReady(pod) || !c.Datastore.PoolLabelsMatch(pod.Labels) {
		logger.V(logutil.DEBUG).Info("Pod removed or not added")
		c.Datastore.PodDelete(pod.Name)
	} else {
		if c.Datastore.PodUpdateOrAddIfNotExist(ctx, pod) {
			logger.V(logutil.DEFAULT).Info("Pod added")
		} else {
			logger.V(logutil.DEFAULT).Info("Pod already exists")
		}
	}
}
