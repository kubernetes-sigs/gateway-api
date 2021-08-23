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

// TCPRoute is the Schema for the TCPRoute resource.
type TCPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of TCPRoute.
	Spec TCPRouteSpec `json:"spec"`

	// Status defines the current state of TCPRoute.
	Status TCPRouteStatus `json:"status,omitempty"`
}

// TCPRouteSpec defines the desired state of TCPRoute
type TCPRouteSpec struct {
	CommonRouteSpec `json:",inline"`

	// Rules are a list of TCP matchers and actions.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Rules []TCPRouteRule `json:"rules"`
}

// TCPRouteStatus defines the observed state of TCPRoute
type TCPRouteStatus struct {
	RouteStatus `json:",inline"`
}

// TCPRouteRule is the configuration for a given rule.
type TCPRouteRule struct {
	// Matches define conditions used for matching the rule against incoming TCP
	// connections. Each match is independent, i.e. this rule will be matched if
	// **any** one of the matches is satisfied. If unspecified (i.e. empty),
	// this Rule will match all requests for the associated Listener.
	//
	// Each client request MUST map to a maximum of one route rule. If a request
	// matches multiple rules, matching precedence MUST be determined in order
	// of the following criteria, continuing on ties:
	//
	// * The most specific match specified by ExtensionRef. Each implementation
	//   that supports ExtensionRef may have different ways of determining the
	//   specificity of the referenced extension.
	//
	// If ties still exist across multiple Routes, matching precedence MUST be
	// determined in order of the following criteria, continuing on ties:
	//
	// * The oldest Route based on creation timestamp. For example, a Route with
	//   a creation timestamp of "2020-09-08 01:02:03" is given precedence over
	//   a Route with a creation timestamp of "2020-09-08 01:02:04".
	// * The Route appearing first in alphabetical order by
	//   "<namespace>/<name>". For example, foo/bar is given precedence over
	//   foo/baz.
	//
	// If ties still exist within the Route that has been given precedence,
	// matching precedence MUST be granted to the first matching rule meeting
	// the above criteria.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Matches []TCPRouteMatch `json:"matches,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent. If unspecified or invalid (refers to a non-existent resource or a
	// Service with no endpoints), the underlying implementation MUST actively
	// reject connection attempts to this backend. Connection rejections must
	// respect weight; if an invalid backend is requested to have 80% of
	// connections, then 80% of connections must be rejected instead.
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

// TCPRouteMatch defines the predicate used to match connections to a
// given action.
type TCPRouteMatch struct {
	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior. For example, resource "mytcproutematcher" in group
	// "networking.example.net". If the referent cannot be found, the rule MUST
	// not be included in the route. The controller must ensure the
	// "ResolvedRefs" condition on the Route status is set to `status: False`.
	//
	// Support: Custom
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

// +kubebuilder:object:root=true

// TCPRouteList contains a list of TCPRoute
type TCPRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TCPRoute `json:"items"`
}
