/*
Copyright 2020 The Kubernetes Authors.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api,shortName=bp
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// BackendPolicy defines policies associated with backends. For the purpose of
// this API, a backend is defined as any resource that a route can forward
// traffic to. A common example of a backend is a Service. Configuration that is
// implementation specific may be represented with similar implementation
// specific custom resources.
type BackendPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of BackendPolicy.
	Spec BackendPolicySpec `json:"spec,omitempty"`

	// Status defines the current state of BackendPolicy.
	Status BackendPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackendPolicyList contains a list of BackendPolicy.
type BackendPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendPolicy `json:"items"`
}

// BackendPolicySpec defines desired policy for a backend.
type BackendPolicySpec struct {
	// BackendRefs define which backends this policy should be applied to. This
	// policy can only apply to backends within the same namespace. If more than
	// one BackendPolicy targets the same backend, precedence must be given to
	// the oldest BackendPolicy.
	//
	// Support: Core
	//
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []BackendRef `json:"backendRefs"`

	// TLS is the TLS configuration for these backends.
	//
	// Support: Extended
	//
	// +optional
	TLS *BackendTLSConfig `json:"tls,omitempty"`
}

// BackendRef identifies an API object within the same namespace
// as the BackendPolicy.
type BackendRef struct {
	// Group is the group of the referent.
	//
	// +kubebuilder:validation:MaxLength=253
	Group string `json:"group"`

	// Kind is the kind of the referent.
	//
	// +kubebuilder:validation:MaxLength=253
	Kind string `json:"kind"`

	// Name is the name of the referent.
	//
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Port is the port of the referent. If unspecified, this policy applies to
	// all ports on the backend.
	//
	// +optional
	Port *PortNumber `json:"port,omitempty"`
}

// BackendTLSConfig describes TLS configuration for a backend.
type BackendTLSConfig struct {
	// CertificateAuthorityRef is a reference to a Kubernetes object that contains
	// one or more trusted CA certificates. The CA certificates are used to establish
	// a TLS handshake to backends listed in BackendRefs. The referenced object MUST
	// reside in the same namespace as BackendPolicy.
	//
	// CertificateAuthorityRef can reference a standard Kubernetes resource, i.e.
	// ConfigMap, or an implementation-specific custom resource.
	//
	// When stored in a Secret, certificates must be PEM encoded and specified within
	// the "ca.crt" data field of the Secret. When multiple certificates are specified,
	// the certificates MUST be concatenated by new lines.
	//
	// CertificateAuthorityRef can also reference a standard Kubernetes resource, i.e.
	// ConfigMap, or an implementation-specific custom resource.
	//
	// Support: Extended
	//
	// +optional
	CertificateAuthorityRef *LocalObjectReference `json:"certificateAuthorityRef,omitempty"`

	// Options are a list of key/value pairs to give extended options to the
	// provider.
	//
	// Support: Implementation-specific
	//
	// +optional
	Options map[string]string `json:"options,omitempty"`
}

// BackendPolicyStatus defines the observed state of BackendPolicy. Conditions
// that are related to a specific Route or Gateway must be placed on the
// Route(s) using backends configured by this BackendPolicy.
type BackendPolicyStatus struct {
	// Conditions describe the current conditions of the BackendPolicy.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BackendPolicyConditionType is a type of condition used to express the current
// state of a BackendPolicy resource.
type BackendPolicyConditionType string

const (
	// Indicates that one or more of the the specified backend references could not be resolved.
	ConditionNoSuchBackend BackendPolicyConditionType = "NoSuchBackend"
)
