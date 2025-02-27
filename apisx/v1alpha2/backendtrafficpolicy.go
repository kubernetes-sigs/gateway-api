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

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=gateway-api,shortName=btrafficpolicy
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
//
// BackendTrafficPolicy is a Direct Attached Policy.
// +kubebuilder:metadata:labels="gateway.networking.k8s.io/policy=Direct"

// BackendTrafficPolicy defines the configuration for how traffic to a
// target backend should be handled.
type BackendTrafficPolicy struct {
	// Support: Extended
	//
	// +optional
	// <gateway:experimental>

	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of BackendTrafficPolicy.
	Spec BackendTrafficPolicySpec `json:"spec"`

	// Status defines the current state of BackendTrafficPolicy.
	Status PolicyStatus `json:"status,omitempty"`
}

// BackendTrafficPolicyList contains a list of BackendTrafficPolicies
// +kubebuilder:object:root=true
type BackendTrafficPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendTrafficPolicy `json:"items"`
}

// BackendTrafficPolicySpec define the desired state of BackendTrafficPolicy
// Note: there is no Override or Default policy configuration.
type BackendTrafficPolicySpec struct {
	// TargetRef identifies an API object to apply policy to.
	// Currently, Backends (i.e. Service, ServiceImport, or any
	// implementation-specific backendRef) are the only valid API
	// target references.
	//
	// +listType=map
	// +listMapKey=group
	// +listMapKey=kind
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	TargetRefs []LocalPolicyTargetReference `json:"targetRefs"`

	// RetryConstraint defines the configuration for when to allow or prevent
	// further retries to a target backend by dynamically calculating a 'retry
	// budget'. This budget is calculated based on the percentage of incoming
	// traffic composed of retries over a given time interval. Once the budget
	// is exceeded, additional retries will be rejected by the backend.
	//
	// For example, if the retry budget interval is 10 seconds, there have been
	// 1000 active requests in the past 10 seconds, and the allowed percentage
	// of requests that can be retried is 20% (the default), then 200 of those
	// requests may be composed of retries. Active requests will only be
	// considered for the duration of the interval when calculating the retry
	// budget.
	//
	// Configuring a RetryConstraint in BackendTrafficPolicy is compatible with
	// HTTPRoute Retry settings for each HTTPRouteRule that targets the same
	// backend. While the HTTPRouteRule Retry stanza can specify whether a
	// request will be retried, and the number of retry attempts each client
	// may perform, RetryConstraint helps prevent cascading failures such as
	// retry storms during periods of consistent failures.
	//
	// After the retry budget has been exceeded, additional retries to the
	// backend MUST return a 503 response to the client.
	//
	// Additional configurations for defining a constraint on retries MAY be
	// defined in the future.
	//
	// Support: Extended
	//
	// +optional
	// <gateway:experimental>
	RetryConstraint *RetryConstraint `json:"retry,omitempty"`

	// SessionPersistence defines and configures session persistence
	// for the backend.
	//
	// Support: Extended
	//
	// +optional
	SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
}

// RetryConstraint defines the configuration for when to retry a request.
type RetryConstraint struct {
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
