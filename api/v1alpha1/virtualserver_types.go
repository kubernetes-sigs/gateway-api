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

// VirtualServer is the Schema for the virtualservers API.
type VirtualServer struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   VirtualServerSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status VirtualServerStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// VirtualServerSpec defines the desired state of VirtualServer.
type VirtualServerSpec struct {
	// Hostnames are the set of domain name that refers to this
	// HTTPServer. These names must be unique across the Listener.
	//
	// Support: Core
	//
	// +required
	Hostnames []string `json:"hostnames" protobuf:"bytes,1,rep,name=hostnames"`
	// TLS is the TLS configuration used for the VirtualServer. If this host has
	// multiple names, each name should be present in the server certificate as
	// a DNS SAN.
	//
	// The ALPNProtocols field in this TLSConfig must contain only valid
	// HTTP protocol identifiers, i.e. "http/0.9", "http/1.0", "http/1.1",
	// "h2". Implementations may accept only a subset of these values if
	// the underlying proxy implementation does not implement the
	// corresponding HTTP protocol version.
	//
	// Support: Core
	//
	// +optional
	TLS *TLSConfig `json:"tls,omitempty" protobuf:"bytes,2,opt,name=tls"`
	// Rules are rules to match, filter and perform actions on requests.
	//
	// Support: Core
	//
	// +required
	Rules []VirtualServerRule `json:"rules" protobuf:"bytes,3,rep,name=rules"`
	// Extension is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmap" (use the empty string
	// for the group) or an implementation-defined resource (for example,
	// resource "myroutehost" in group "networking.acme.io").
	//
	// Support: Custom
	//
	// +optional
	Extension *VirtualServerExtensionObjectReference `json:"extension" protobuf:"bytes,4,opt,name=extension"`
}

// VirtualServerRule is a rule to match, filter and perform actions on requests
// to a VirtualServer.
type VirtualServerRule struct {
	// Match defines criteria for matching a request.
	//
	// Support: Core
	//
	// +required
	Match *VirtualServerMatch `json:"match" protobuf:"bytes,1,opt,name=match"`
	// Filter defines what filters are applied to the request.
	//
	// Support: Core
	//
	// +optional
	Filter *VirtualServerFilter `json:"filter,omitempty" protobuf:"bytes,2,opt,name=filter"`
	// Action defines what happens to the request.
	//
	// Support: Core
	//
	// +required
	Action *VirtualServerAction `json:"action" protobuf:"bytes,3,opt,name=action"`
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

// VirtualServerMatch defines the predicate used to match requests to a
// VirtualServer.
type VirtualServerMatch struct {
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
	Extension *VirtualServerMatchExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,5,opt,name=extension"`
}

// VirtualServerMatchExtensionObjectReference identifies a route-match extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
type VirtualServerMatchExtensionObjectReference = LocalObjectReference

// VirtualServerFilter defines a filter-like action to be applied to
// requests.
type VirtualServerFilter struct {
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
	Extension *VirtualServerFilterExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,2,opt,name=extension"`
}

// VirtualServerFilterExtensionObjectReference identifies a VirtualServer
// filter extension object within a known namespace.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
type VirtualServerFilterExtensionObjectReference = LocalObjectReference

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

// VirtualServerAction is the action taken given a match.
type VirtualServerAction struct {
	// ForwardTo sends requests to the referenced object.  The resource may
	// be "service" (use the empty string for the group), or an
	// implementation may support other resources (for example, resource
	// "my-virtualserver-target" in group "networking.acme.io").
	//
	// Support: Core
	//
	// +optional
	ForwardTo *VirtualServerActionTargetObjectReference `json:"forwardTo,omitempty" protobuf:"bytes,1,opt,name=forwardTo"`
	// Extension is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "my-virtualserver-action" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *VirtualServerActionExtensionObjectReference `json:"extension,omitempty" protobuf:"bytes,2,opt,name=extension"`
}

// VirtualServerActionTargetObjectReference identifies a target object for a
// VirtualServer action within a known namespace.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
type VirtualServerActionTargetObjectReference = LocalObjectReference

// VirtualServerActionExtensionObjectReference identifies a VirtualServer
// action extension object within a known namespace.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
type VirtualServerActionExtensionObjectReference = LocalObjectReference

// VirtualServerExtensionObjectReference identifies an extension object
// for a VirtualServer within a known namespace.
//
// +k8s:deepcopy-gen=false
// +protobuf=false
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

// VirtualServerStatus defines the observed state of VirtualServer.
type VirtualServerStatus struct {
	Gateways []GatewayObjectReference `json:"gateways" protobuf:"bytes,1,rep,name=gateways"`
}

// +kubebuilder:object:root=true

// VirtualServerList contains a list of VirtualServer.
type VirtualServerList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []VirtualServer `json:"items" protobuf:"bytes,3,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&VirtualServer{}, &VirtualServerList{})
}
