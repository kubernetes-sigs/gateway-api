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
// +kubebuilder:subresource:status

// UDPRoute is the Schema for the UDPRoute resource.
type UDPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UDPRouteSpec   `json:"spec,omitempty"`
	Status UDPRouteStatus `json:"status,omitempty"`
}

// UDPRouteRule is the configuration for a given rule.
type UDPRouteRule struct {
	// Matches defines which packets match this rule.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Matches []UDPRouteMatch `json:"matches,omitempty"`

	// ForwardTo defines the backend(s) where matching requests should be sent.
	// +optional
	// +kubebuilder:validation:MaxItems=4
	ForwardTo []RouteForwardTo `json:"forwardTo,omitempty"`
}

// UDPRouteSpec defines the desired state of UDPRoute.
type UDPRouteSpec struct {
	// Rules are a list of UDP matchers and actions.
	Rules []UDPRouteRule `json:"rules"`

	// Gateways defines which Gateways can use this Route.
	// +kubebuilder:default={allow: "SameNamespace"}
	Gateways RouteGateways `json:"gateways,omitempty"`
}

// UDPRouteStatus defines the observed state of UDPRoute.
type UDPRouteStatus struct {
	RouteStatus `json:",inline"`
}

// UDPRouteMatch defines the predicate used to match packets to a
// given action.
type UDPRouteMatch struct {
	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematchers" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the UDPRoute will be true.
	//
	// Support: custom
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
