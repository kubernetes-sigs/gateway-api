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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// The TLSRoute resource is similar to TCPRoute, but can be configured
// to match against TLS-specific metadata. This allows more flexibility
// in matching streams for a given TLS listener.
//
// If you need to forward traffic to a single target for a TLS listener, you
// could choose to use a TCPRoute with a TLS listener.
type TLSRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of TLSRoute.
	Spec TLSRouteSpec `json:"spec"`

	// Status defines the current state of TLSRoute.
	Status TLSRouteStatus `json:"status,omitempty"`
}

// TLSRouteSpec defines the desired state of a TLSRoute resource.
type TLSRouteSpec struct {
	CommonRouteSpec `json:",inline"`

	// Hostnames defines a set of SNI names that should match against the
	// SNI attribute of TLS ClientHello message in TLS handshake.
	//
	// SNI can be "precise" which is a domain name without the terminating
	// dot of a network host (e.g. "foo.example.com") or "wildcard", which is
	// a domain name prefixed with a single wildcard label (e.g. `*.example.com`).
	// The wildcard character `*` must appear by itself as the first DNS label
	// and matches only a single label. You cannot have a wildcard label by
	// itself (e.g. Host == `*`).
	//
	// Requests will be matched against the SNI attribute in the following
	// order:
	//
	// 1. If SNI is precise, the request matches this Route if the SNI in
	//    ClientHello is equal to one of the defined SNIs.
	// 2. If SNI is a wildcard, then the request matches this Route if the
	//    SNI is to equal to the suffix (removing the first label) of the
	//    wildcard.
	// 3. If SNIs are unspecified, all requests associated with the gateway TLS
	//    listener will match. This can be used to define a default backend
	//    for a TLS listener.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Hostnames []Hostname `json:"hostnames,omitempty"`

	// Rules are a list of TLS matchers and actions.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Rules []TLSRouteRule `json:"rules"`
}

// TLSRouteStatus defines the observed state of TLSRoute
type TLSRouteStatus struct {
	RouteStatus `json:",inline"`
}

// TLSRouteRule is the configuration for a given rule.
type TLSRouteRule struct {
	// BackendRefs defines the backend(s) where matching requests should be
	// sent. If unspecified or invalid (refers to a non-existent resource or
	// a Service with no endpoints), the rule performs no forwarding; if no
	// filters are specified that would result in a response being sent, the
	// underlying implementation must actively reject request attempts to this
	// backend, by rejecting the connection or returning a 503 status code.
	// Request rejections must respect weight; if an invalid backend is
	// requested to have 80% of requests, then 80% of requests must be rejected
	// instead.
	//
	// Support: Core for Kubernetes Service
	// Support: Custom for any other resource
	//
	// Support for weight: Extended
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []BackendRef `json:"backendRefs,omitempty"`
}

// +kubebuilder:object:root=true

// TLSRouteList contains a list of TLSRoute
type TLSRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TLSRoute `json:"items"`
}
