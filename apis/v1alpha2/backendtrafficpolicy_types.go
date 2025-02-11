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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BackendTrafficPolicy struct {
	// BackendTrafficPolicy defines the configuration for how traffic to a
	// target backend should be handled.
	//
	// Support: Extended
	//
	// +optional
	// <gateway:experimental>
	//
	// Note: there is no Override or Default policy configuration.

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// Status defines the current state of BackendTrafficPolicy.
	Status PolicyStatus `json:"status,omitempty"`
}

type BackendTrafficPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	// Currently, Backends (i.e. Service, ServiceImport, or any
	// implementation-specific backendRef) are the only valid API
	// target references.
	// +listType=map
	// +listMapKey=group
	// +listMapKey=kind
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	TargetRefs []LocalPolicyTargetReference `json:"targetRefs"`

	// Retry defines the configuration for when to allow or prevent retries to a
	// target backend.
	//
	// While the static number of retries performed by the client are
	// configured within HTTPRoute Retry stanzas, configuring the
	// CommonRetryPolicy allows you to constrain further retries after a
	// dynamic budget for retries has been exceeded.
	//
	// Support: Extended
	//
	// +optional
	// <gateway:experimental>
	Retry *CommonRetryPolicy `json:"retry,omitempty"`

	// SessionPersistence defines and configures session persistence
	// for the backend.
	//
	// Support: Extended
	//
	// +optional
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
}

// CommonRetryPolicy defines the configuration for when to retry a request.
type CommonRetryPolicy struct {
	// BudgetPercent defines the maximum percentage of active requests that may
	// be made up of retries.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:default=20
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	BudgetPercent *int `json:"budgetPercent,omitempty"`

	// BudgetInterval defines the duration in which requests will be considered
	// for calculating the budget for retries.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:default=10s
	BudgetInterval *Duration `json:"budgetInterval,omitempty"`

	// MinRetryRate defines the minimum rate of retries that will be allowable
	// over a specified duration of time.
	//
	// This ensures that requests can still be retried during periods of low
	// traffic, where the budget for retries may be calculated as a very low
	// value.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:default={count: 10, interval: 1s}
	MinRetryRate *RequestRate `json:"minRetryRate,omitempty"`
}

// RequestRate expresses a rate of requests over a given period of time.
type RequestRate struct {
	// Count specifies the number of requests per time interval.
	//
	// Support: Extended
	Count *int `json:"count,omitempty"`

	// Interval specifies the divisor of the rate of requests, the amount of
	// time during which the given count of requests occur.
	//
	// Support: Extended
	Interval *Duration `json:"interval,omitempty"`
}
