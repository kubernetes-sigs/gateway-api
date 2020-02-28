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

// +kubebuilder:object:root=true

// VirtualHost is the Schema for the virtualhosts API.
type VirtualHost struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   VirtualHostSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status VirtualHostStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// VirtualHostSpec defines the desired state of VirtualHost.
type VirtualHostSpec struct {
	// Hostname is the fully qualified domain name of a network host,
	// as defined by RFC 3986. Note the following deviations from the
	// "host" part of the URI as defined in the RFC:
	//
	// 1. IPs are not allowed.
	// 2. The `:` delimiter is not respected because ports are not allowed.
	//
	// Incoming requests are matched against Hostname before processing HTTPRoute
	// rules. For example, if the request header contains host: foo.example.com,
	// an HTTPRoute with hostname foo.example.com will match. However, an
	// HTTPRoute with hostname example.com or bar.example.com will not match.
	// If Hostname is unspecified, all traffic is routed to the VirtualHost based
	// on the specified rules.
	//
	// Support: Core
	//
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,1,opt,name=hostname"`
	// Rules are rules to match, filter and perform actions on requests of a VirtualHost.
	//
	// Support: Core
	//
	// +required
	Rules []VirtualHostRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`
	// Extension is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmap" (use the empty string
	// for the group) or an implementation-defined resource (for example,
	// resource "myextension" in group "networking.acme.io").
	//
	// Support: Custom
	//
	// +optional
	Extension *VirtualServerExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,4,opt,name=extension"`
}

// VirtualHostRule is a rule to match, filter and perform actions on requests
// to a VirtualHost.
type VirtualHostRule struct {
	// Match defines criteria for matching a request.
	//
	// Support: Core
	//
	// +required
	Match *VirtualHostMatch `json:"match" protobuf:"bytes,1,opt,name=match"`
	// Filter defines what filters are applied to the request.
	//
	// Support: Core
	//
	// +optional
	Filter *VirtualHostFilter `json:"filter,omitempty" protobuf:"bytes,2,opt,name=filter"`
	// Action defines what happens to the request.
	//
	// Support: Core
	//
	// +required
	Action *VirtualHostAction `json:"action" protobuf:"bytes,3,opt,name=action"`
}

// PathType constants.
const (
	PathTypeExact                = "Exact"
	PathTypePrefix               = "Prefix"
	PathTypeRegularExpression    = "RegularExpression"
	PathTypeImplementionSpecific = "ImplementationSpecific"
)

// HeaderType constants.
const (
	HeaderTypeExact = "Exact"
)

// VirtualHostMatch defines the predicate used to match requests to a
// VirtualHost.
type VirtualHostMatch struct {
	// PathType is defines the semantics of the `Path` matcher.
	//
	// Support: core (Exact, Prefix)
	// Support: extended (RegularExpression)
	// Support: custom (ImplementationSpecific)
	//
	// Default: "Exact"
	//
	// +optional
	PathType string `json:"pathType,omitempty" protobuf:"bytes,1,opt,name=pathType"`
	// Path is the value of the HTTP path as interpreted via
	// PathType.
	//
	// Default: "/"
	//
	// Support: ?
	//
	// +optional
	Path *string `json:"path,omitempty" protobuf:"bytes,2,opt,name=path"`
	// HeaderType defines the semantics of the `Header` matcher.
	//
	// Support: ?
	//
	// +optional
	HeaderType *string `json:"headerType,omitempty" protobuf:"bytes,3,opt,name=headerType"`
	// Header are the Header matches as interpreted via
	// HeaderType.
	//
	// Support: ?
	//
	// +optional
	Header map[string]string `json:"header,omitempty" protobuf:"bytes,4,rep,name=header"`
	// Extension is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematcher" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *VirtualHostMatchExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,5,opt,name=extension"`
}

// VirtualHostMatchExtensionObjectReference identifies a VirtualHost match extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type VirtualHostMatchExtensionObjectReference = LocalObjectReference

// VirtualHostFilter defines a filter-like action to be applied to
// requests.
type VirtualHostFilter struct {
	// Headers related filters.
	//
	// Support: extended
	// +optional
	Headers *HTTPHeaderFilter `json:"headers,omitempty" protobuf:"bytes,1,opt,name=headers"`

	// Extension is an optional, implementation-specific extension to the
	// "filter" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutefilter" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *VirtualHostFilterExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,2,opt,name=extension"`
}

// VirtualHostFilterExtensionObjectReference identifies a VirtualHost
// filter extension object within a known namespace.
//
// +k8s:deepcopy-gen=false
type VirtualHostFilterExtensionObjectReference = LocalObjectReference

// HTTPHeaderFilter defines the filter behavior for a request match.
type HTTPHeaderFilter struct {
	// Add adds the given header (name, value) to the request
	// before the action.
	//
	// Input:
	//   GET /foo HTTP/1.1
	//
	// Config:
	//   add: {"my-header": "foo"}
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//
	// Support: extended?
	//
	// +optional
	Add map[string]string `json:"add,omitempty" protobuf:"bytes,1,rep,name=add"`
	// Remove the given header(s) on the HTTP request before the
	// action. The value of RemoveHeader is a list of HTTP header
	// names. Note that the header names are case-insensitive
	// [RFC-2616 4.2].
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   My-Header1: ABC
	//   My-Header2: DEF
	//   My-Header2: GHI
	//
	// Config:
	//   remove: ["my-header1", "my-header3"]
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   My-Header2: DEF
	//
	// Support: extended?
	//
	// +optional
	Remove []string `json:"remove,omitempty" protobuf:"bytes,2,rep,name=remove"`
}

// VirtualHostAction is the action taken given a match.
type VirtualHostAction struct {
	// ForwardTo sends requests to the referenced object.  The resource may
	// be "service" (use the empty string for the group), or an
	// implementation may support other resources (for example, resource
	// "my-virtualserver-target" in group "networking.acme.io").
	//
	// Support: Core
	//
	// +optional
	ForwardTo *VirtualHostActionTargetObjectReference `json:"forwardTo,omitempty" protobuf:"bytes,1,opt,name=forwardTo"`
	// Extension is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "my-virtualserver-action" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *VirtualHostActionExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,2,opt,name=extension"`
}

// VirtualHostActionTargetObjectReference identifies a target object for a
// VirtualHost action within a known namespace.
//
// +k8s:deepcopy-gen=false
type VirtualHostActionTargetObjectReference = LocalObjectReference

// VirtualHostActionExtensionObjectReference identifies a VirtualHost
// action extension object within a known namespace.
//
// +k8s:deepcopy-gen=false
type VirtualHostActionExtensionObjectReference = LocalObjectReference

// VirtualServerExtensionObjectReference identifies an extension object
// for a VirtualHost within a known namespace.
//
// +k8s:deepcopy-gen=false
type VirtualServerExtensionObjectReference = LocalObjectReference

// GatewayObjectReference identifies a Gateway object.
type GatewayObjectReference struct {
	// Namespace is the namespace of the referent.
	// +optional
	Namespace string `json:"namespace,omitempty" protobuf:"bytes,1,opt,name=namespace"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	// +required
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
}

// VirtualHostStatus defines the observed state of VirtualHost.
type VirtualHostStatus struct {
	Gateways []GatewayObjectReference `json:"gateways" protobuf:"bytes,1,rep,name=gateways"`
}

// +kubebuilder:object:root=true

// VirtualHostList contains a list of VirtualHost.
type VirtualHostList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []VirtualHost `json:"items" protobuf:"bytes,3,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&VirtualHost{}, &VirtualHostList{})
}
