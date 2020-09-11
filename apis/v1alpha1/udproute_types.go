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

// UDPRouteSpec defines the desired state of UDPRoute.
type UDPRouteSpec struct {
	// Rules are a list of UDP matchers and actions.
	Rules []UDPRouteRule `json:"rules" protobuf:"bytes,1,rep,name=rules"`
}

// UDPRouteStatus defines the observed state of UDPRoute.
type UDPRouteStatus struct {
	RouteStatus `json:",inline"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UDPRoute is the Schema for the udproutes API.
type UDPRoute struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   UDPRouteSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status UDPRouteStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// UDPRouteRule is the configuration for a given rule.
type UDPRouteRule struct {
	// Match defines which packets match this rule.
	//
	// +optional
	Match *UDPRouteMatch `json:"match" protobuf:"bytes,1,opt,name=match"`
	// Action defines what happens to the packet.
	//
	// +optional
	Action *UDPRouteAction `json:"action" protobuf:"bytes,2,opt,name=action"`
}

// UDPRouteAction is the action for a given rule.
type UDPRouteAction struct {
	// ForwardTo sends requests to the referenced object.  The
	// resource may be "services" (omit or use the empty string for the
	// group), or an implementation may support other resources (for
	// example, resource "myroutetargets" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "services".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the UDPRoute will be true.
	ForwardTo *ForwardToTarget `json:"forwardTo" protobuf:"bytes,1,opt,name=forwardTo"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmaps" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myrouteactions" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the UDPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteActionExtensionObjectReference `json:"extensionRef" protobuf:"bytes,2,opt,name=extensionRef"`
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
	ExtensionRef *RouteMatchExtensionObjectReference `json:"extensionRef" protobuf:"bytes,1,opt,name=extensionRef"`
}

// +kubebuilder:object:root=true

// UDPRouteList contains a list of UDPRoute
type UDPRouteList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []UDPRoute `json:"items" protobuf:"bytes,3,rep,name=items"`
}
