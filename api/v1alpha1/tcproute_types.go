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

// TcpRouteSpec defines the desired state of TcpRoute.
type TcpRouteSpec struct {
	// Hosts is a list of Host definitions.
	Hosts []TcpRouteHost `json:"hosts,omitempty" protobuf:"bytes,1,rep,name=hosts"`

	// Default is the default host to use. Default.Hostname must be empty.
	//
	// +optional
	Default *TcpRouteHost `json:"default" protobuf:"bytes,2,opt,name=default"`
}

// TcpRouteHost is the configuration for a given host.
type TcpRouteHost struct {
	// Hostname is the fully qualified domain name of a network host,
	// as defined by RFC 3986. Note the following deviations from the
	// "host" part of the URI as defined in the RFC:
	//
	// 1. The `:` delimiter is not respected because ports are not allowed.
	//
	// This field is only available for TcpRoutes that used with TLS-enabled
	// Listeners. The value of this field is matched against the TLS Server
	// Name Indication (SNI) value.
	//
	// Incoming requests are matched against Hostname before processing TcpRoute
	// rules. For example, if the request header contains host: foo.example.com,
	// an HTTPRoute with hostname foo.example.com will match. However, an
	// HTTPRoute with hostname example.com or bar.example.com will not match.
	// If Hostname is unspecified, the Gateway routes all traffic based on
	// the specified rules.
	//
	// Support: Core
	//
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,1,opt,name=hostname"`

	// Rules are a list of matchers and actions.
	Rules []TcpRouteRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`

	// Extension is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmap" (use the empty string
	// for the group) or an implementation-defined resource (for example,
	// resource "myroutehost" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteHostExtensionObjectReference `json:"extension" protobuf:"bytes,3,opt,name=extension"`
}

// TcpRouteRule is the configuration for a given route.
type TcpRouteRule struct {
	// Match defines which requests match this path.
	// +optional
	Match *TcpRouteMatch `json:"match" protobuf:"bytes,1,opt,name=match"`
	// Action defines what happens to the request.
	// +optional
	Action *TcpRouteAction `json:"action" protobuf:"bytes,2,opt,name=action"`
}

// TcpRouteMatch defines the predicate used to match requests to a given action.
type TcpRouteMatch struct {
	// Listener is the name of the Listener that received the routed traffic.
	//
	// Support: Core
	//
	// +optional
	Listener string `json:"listener,omitempty" protobuf:"bytes,1,opt,name=listener"`

	// Port is the TCP port on a Listener that received the routed traffic.
	//
	// Support: Core
	//
	// +optional
	Port *int32 `json:"port,omitempty" protobuf:"varint,2,opt,name=port"`

	// Extension is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematcher" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteMatchExtensionObjectReference `json:"extension" protobuf:"bytes,3,opt,name=extension"`
}

// TcpRouteAction is the action taken given a match.
type TcpRouteAction struct {
	// ForwardTo sends requests to the referenced object.  The resource may
	// be "service" (use the empty string for the group), or an
	// implementation may support other resources (for example, resource
	// "myroutetarget" in group "networking.acme.io").
	ForwardTo *RouteActionTargetObjectReference `json:"forwardTo" protobuf:"bytes,1,opt,name=forwardTo"`

	// Extension is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myrouteaction" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteActionExtensionObjectReference `json:"extension" protobuf:"bytes,2,opt,name=extension"`
}

// TcpRouteStatus defines the observed state of TcpRoute
type TcpRouteStatus struct {
	Gateways []GatewayObjectReference `json:"gateways" protobuf:"bytes,1,rep,name=gateways"`
}

// +kubebuilder:object:root=true

// TcpRoute is the Schema for the tcproutes API
type TcpRoute struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   TcpRouteSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status TcpRouteStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true

// TcpRouteList contains a list of TcpRoute
type TcpRouteList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []TcpRoute `json:"items" protobuf:"bytes,3,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&TcpRoute{}, &TcpRouteList{})
}
