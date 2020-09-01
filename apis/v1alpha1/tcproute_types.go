/*

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

// TCPRouteSpec defines the desired state of TCPRoute
type TCPRouteSpec struct {
	// Rules are a list of TCP matchers and actions.
	Rules []TCPRouteRule `json:"rules"`
}

// TCPRouteStatus defines the observed state of TCPRoute
type TCPRouteStatus struct {
	// Gateways is a list of the Gateways that are associated with the
	// TCPRoute, and the status of the route with respect to each of these
	// Gateways.  When a Gateway selects this route, the controller that
	// manages the Gateway should add an entry to this list when the
	// controller first sees the route and should update the entry as
	// appropriate when the route is modified.
	Gateways []RouteGatewayStatus `json:"gateways"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TCPRoute is the Schema for the tcproutes API
type TCPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TCPRouteSpec   `json:"spec,omitempty"`
	Status TCPRouteStatus `json:"status,omitempty"`
}

// TCPRouteRule is the configuration for a given rule.
type TCPRouteRule struct {
	// Matches define conditions used for matching the rule against
	// incoming TCP connections.
	// Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	//
	// +optional
	Matches []TCPRouteMatch `json:"matches"`

	// Action defines what happens to the connection.
	// +required
	Action TCPRouteAction `json:"action"`
}

// TCPRouteAction is the action for a given rule.
type TCPRouteAction struct {
	// ForwardTo sends requests to the referenced object(s). The
	// resource may be "services" (omit or use the empty string for the
	// group), or an implementation may support other resources (for
	// example, resource "myroutetargets" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "services".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: core
	//
	// +kubebuilder:validation:MinItems=1
	ForwardTo []ForwardToTarget `json:"forwardTo"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmaps" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myrouteactions" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the TCPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteActionExtensionObjectReference `json:"extensionRef,omitempty"`
}

// TCPRouteMatch defines the predicate used to match connections to a
// given action.
type TCPRouteMatch struct {
	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematchers" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the TCPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteMatchExtensionObjectReference `json:"extensionRef"`
}

// +kubebuilder:object:root=true

// TCPRouteList contains a list of TCPRoute
type TCPRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TCPRoute `json:"items"`
}
