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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Hostnames",type=string,JSONPath=`.spec.hostnames`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// CustomGRPCRoute provides a way to route gRPC requests. This includes the capability
// to match requests by hostname, gRPC service, gRPC method, or HTTP/2 header.
// Filters can be used to specify additional processing steps. Backends specify
// where matching requests will be routed.
//
// CustomGRPCRoute falls under extended support within the Gateway API. Within the
// following specification, the word "MUST" indicates that an implementation
// supporting GRPCRoute must conform to the indicated requirement, but an
// implementation not supporting this route type need not follow the requirement
// unless explicitly indicated.
//
// Implementations supporting `CustomGRPCRoute` with the `HTTPS` `ProtocolType` MUST
// accept HTTP/2 connections without an initial upgrade from HTTP/1.1, i.e. via
// ALPN. If the implementation does not support this, then it MUST set the
// "Accepted" condition to "False" for the affected listener with a reason of
// "UnsupportedProtocol".  Implementations MAY also accept HTTP/2 connections
// with an upgrade from HTTP/1.
//
// Implementations supporting `CustomGRPCRoute` with the `HTTP` `ProtocolType` MUST
// support HTTP/2 over cleartext TCP (h2c,
// https://www.rfc-editor.org/rfc/rfc7540#section-3.1) without an initial
// upgrade from HTTP/1.1, i.e. with prior knowledge
// (https://www.rfc-editor.org/rfc/rfc7540#section-3.4). If the implementation
// does not support this, then it MUST set the "Accepted" condition to "False"
// for the affected listener with a reason of "UnsupportedProtocol".
// Implementations MAY also accept HTTP/2 connections with an upgrade from
// HTTP/1, i.e. without prior knowledge.
//
// Support: Extended
type CustomGRPCRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of GRPCRoute.
	Spec GRPCRouteSpec `json:"spec,omitempty"`

	// Status defines the current state of GRPCRoute.
	Status GRPCRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GRPCRouteList contains a list of GRPCRoute.
type CustomGRPCRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomGRPCRoute `json:"items"`
}
