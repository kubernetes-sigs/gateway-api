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

// UDPRoute is a resource that specifies how a Gateway should forward UDP traffic.
type UDPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of UDPRoute.
	Spec UDPRouteSpec `json:"spec"`

	// Status defines the current state of UDPRoute.
	Status UDPRouteStatus `json:"status,omitempty"`
}

// UDPRouteSpec defines the desired state of UDPRoute.
type UDPRouteSpec struct {
	// ParentRefs references the resources (usually Gateways) that a Route wants
	// to be attached to. Note that the referenced parent resource needs to
	// allow this for the attachment to be complete. For Gateways, that means
	// the Gateway needs to allow attachment from Routes of this kind and
	// namespace.
	//
	// The only kind of parent resource with "Core" support is Gateway. This API
	// may be extended in the future to support additional kinds of parent
	// resources such as one of the route kinds.
	//
	// It is invalid to reference an identical parent more than once. It is
	// valid to reference multiple distinct sections within the same parent
	// resource, such as 2 Listeners within a Gateway.
	//
	// It is possible to separately reference multiple distinct objects that may
	// be collapsed by an implementation. For example, some implementations may
	// choose to merge compatible Gateway Listeners together. If that is the
	// case, the list of routes attached to those resources should also be
	// merged.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	ParentRefs []ParentRef `json:"parentRefs,omitempty"`

	// Rules are a list of UDP matchers and actions.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Rules []UDPRouteRule `json:"rules"`
}

// UDPRouteStatus defines the observed state of UDPRoute.
type UDPRouteStatus struct {
	RouteStatus `json:",inline"`
}

// UDPRouteRule is the configuration for a given rule.
type UDPRouteRule struct {
	// Matches define conditions used for matching the rule against incoming UDP
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
	Matches []UDPRouteMatch `json:"matches,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent.
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

// UDPRouteMatch defines the predicate used to match packets to a
// given action.
type UDPRouteMatch struct {
	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior.  For example, resource "myudproutematcher" in group
	// "networking.acme.io". If the referent cannot be found, the rule is not
	// included in the route. The controller should raise the "ResolvedRefs"
	// condition on the Gateway with the "DegradedRoutes" reason. The gateway
	// status for this route should be updated with a condition that describes
	// the error more specifically.
	//
	// Support: Custom
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

// +kubebuilder:object:root=true

// UDPRouteList contains a list of UDPRoute
type UDPRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UDPRoute `json:"items"`
}
