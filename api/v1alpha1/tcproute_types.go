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

// TCPRouteSpec defines the desired state of TCPRoute.
type TCPRouteSpec struct {
	// Hosts is a list of TCP host definitions.
	Hosts []TCPRouteHost `json:"hosts,omitempty" protobuf:"bytes,1,rep,name=hosts"`
}

// TCPRouteHost is the configuration for a given TCP host.
type TCPRouteHost struct {
	// Hostname is the fully qualified domain name of a network host,
	// as defined by RFC 3986. Note the following deviations from the
	// "host" part of the URI as defined in the RFC:
	//
	// 1. IPs are not allowed.
	// 2. The `:` delimiter is not respected because ports are not allowed.
	//
	// The server_name in a TLS SNI client hello message is used for matching
	// a TCPRoute Hostname. Incoming requests are matched against Hostname
	// before processing TCPRoute rules. For example, if the SNI server_name
	// is foo.example.com, a TCPRoute with hostname foo.example.com will match,
	// but a TCPRoute with hostname example.com or bar.example.com will not.
	//
	// Support: Core
	//
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,1,opt,name=hostname"`

	// Rules are a list of actions to take against a matched TCP request.
	Rules []TCPRouteRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmaps" (omit or specify the
	// empty string for the group) or an implementation-defined resource
	// (for example, resource "myroutehosts" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the TCPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteHostExtensionObjectReference `json:"extensionRef" protobuf:"bytes,3,opt,name=extensionRef"`
}

// TCPRouteRule is the configuration for a given TCP host.
type TCPRouteRule struct {
	// Action defines what happens to the TCP request.
	//
	// +optional
	Action *TCPRouteAction `json:"action" protobuf:"bytes,2,opt,name=action"`
}

// TCPRouteAction is the action taken given a matched TCP request.
type TCPRouteAction struct {
	// ForwardTo sends requests to the referenced object.  The
	// resource may be "services" (omit or use the empty string for the
	// group), or an implementation may support other resources (for
	// example, resource "myroutetargets" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "services".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the TCPRoute will be true.
	ForwardTo *ForwardToTCPTarget `json:"forwardTo" protobuf:"bytes,1,opt,name=forwardTo"`

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
	ExtensionRef *RouteActionExtensionObjectReference `json:"extensionRef" protobuf:"bytes,2,opt,name=extensionRef"`
}

// ForwardToTCPTarget identifies a TCP target object within a known namespace.
type ForwardToTCPTarget struct {
	// TargetRef is an object reference to forward matched requests to.
	//
	// Support: Core (Kubernetes Services)
	// Support: Implementation-specific (Other resource types)
	//
	TargetRef ForwardToTargetObjectReference `json:"targetRef" protobuf:"bytes,1,opt,name=targetRef"`

	// TargetPort specifies the destination port number to use for the TargetRef.
	// If unspecified and TargetRef is a Service object consisting of a single
	// port definition, that port will be used. If unspecified and TargetRef is
	// a Service object consisting of multiple port definitions, an error is
	// surfaced in status.
	//
	// Support: Core
	//
	// +optional
	TargetPort *TargetPort `json:"targetPort" protobuf:"bytes,2,opt,name=targetPort"`
}

// TCPRouteStatus defines the observed state of TCPRoute.
type TCPRouteStatus struct {
	GatewayRefs []GatewayObjectReference `json:"gatewayRefs" protobuf:"bytes,1,rep,name=gatewayRefs"`
}

// +kubebuilder:object:root=true

// TCPRoute is the Schema for the tcproutes API.
type TCPRoute struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   TCPRouteSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status TCPRouteStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true

// TCPRouteList contains a list of TCPRoute.
type TCPRouteList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []TCPRoute `json:"items" protobuf:"bytes,3,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&TCPRoute{}, &TCPRouteList{})
}
