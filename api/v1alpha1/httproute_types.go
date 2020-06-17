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

// HTTPRouteSpec defines the desired state of HTTPRoute
type HTTPRouteSpec struct {
	// Hosts is a list of Host definitions.
	Hosts []HTTPRouteHost `json:"hosts,omitempty" protobuf:"bytes,1,rep,name=hosts"`

	// Default is the default host to use. Default.Hostnames must
	// be an empty list.
	//
	// +optional
	Default *HTTPRouteHost `json:"default" protobuf:"bytes,2,opt,name=default"`
}

// HTTPRouteHost is the configuration for a given host.
type HTTPRouteHost struct {
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
	// If Hostname is unspecified, the Gateway routes all traffic based on
	// the specified rules.
	//
	// Support: Core
	//
	// +optional
	Hostname string `json:"hostname,omitempty" protobuf:"bytes,1,opt,name=hostname"`

	// Rules are a list of HTTP matchers, filters and actions.
	Rules []HTTPRouteRule `json:"rules" protobuf:"bytes,2,rep,name=rules"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmaps" (omit or specify the
	// empty string for the group) or an implementation-defined resource
	// (for example, resource "myroutehosts" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteHostExtensionObjectReference `json:"extensionRef" protobuf:"bytes,3,opt,name=extensionRef"`
}

// HTTPRouteRule is the configuration for a given path.
type HTTPRouteRule struct {
	// Match defines which requests match this path.
	// +optional
	Match *HTTPRouteMatch `json:"match" protobuf:"bytes,1,opt,name=match"`
	// Filter defines what filters are applied to the request.
	// +optional
	Filter *HTTPRouteFilter `json:"filter" protobuf:"bytes,2,opt,name=filter"`
	// Action defines what happens to the request.
	// +optional
	Action *HTTPRouteAction `json:"action" protobuf:"bytes,3,opt,name=action"`
}

// PathType constants.
const (
	PathTypeExact                = "Exact"
	PathTypePrefix               = "Prefix"
	PathTypeRegularExpression    = "RegularExpression"
	PathTypeImplementionSpecific = "ImplementationSpecific"
)

// HeaderMatchType constants.
const (
	// HeaderMatchTypeExact matches HTTP request-header fields.
	// Field names matches are case-insensitive while field values matches
	// are case-sensitive.
	HeaderMatchTypeExact                = "Exact"
	HeaderMatchTypeImplementionSpecific = "ImplementationSpecific"
)

// HTTPRouteMatch defines the predicate used to match requests to a
// given action.
type HTTPRouteMatch struct {
	// PathType is defines the semantics of the `Path` matcher.
	//
	// Support: core (Exact, Prefix)
	// Support: extended (RegularExpression)
	// Support: custom (ImplementationSpecific)
	//
	// Default: "Exact"
	//
	// +optional
	PathType string `json:"pathType" protobuf:"bytes,1,opt,name=pathType"`
	// Path is the value of the HTTP path as interpreted via
	// PathType.
	//
	// Default: "/"
	Path *string `json:"path" protobuf:"bytes,2,opt,name=path"`

	// HeaderMatchType defines the semantics of the `Header` matcher.
	//
	// Support: core (Exact)
	// Support: custom (ImplementationSpecific)
	//
	// Default: "Exact"
	//
	// +optional
	HeaderMatchType *string `json:"headerMatchType" protobuf:"bytes,3,opt,name=headerMatchType"`
	// Headers are the HTTP Headers to match as interpreted via
	// HeaderMatchType. Multiple headers are ANDed together, meaning, a request
	// must contain all the headers specified in order to select this route.
	//
	// +optional
	Headers map[string]string `json:"headers" protobuf:"bytes,4,rep,name=headers"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematchers" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteMatchExtensionObjectReference `json:"extensionRef" protobuf:"bytes,5,opt,name=extensionRef"`
}

// RouteMatchExtensionObjectReference identifies a route-match extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteMatchExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// HTTPRouteFilter defines a filter-like action to be applied to
// requests.
type HTTPRouteFilter struct {
	// Headers related filters.
	//
	// Support: extended
	// +optional
	Headers *HTTPHeaderFilter `json:"headers" protobuf:"bytes,1,opt,name=headers"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "filter" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutefilters" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteFilterExtensionObjectReference `json:"extensionRef" protobuf:"bytes,2,opt,name=extensionRef"`
}

// RouteFilterExtensionObjectReference identifies a route-filter extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteFilterExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

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
	Add map[string]string `json:"add" protobuf:"bytes,1,rep,name=add"`

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
	Remove []string `json:"remove" protobuf:"bytes,2,rep,name=remove"`

	// TODO
}

// HTTPRouteAction is the action taken given a match.
type HTTPRouteAction struct {
	// ForwardTo sends requests to the referenced object.  The
	// resource may be "services" (omit or use the empty string for the
	// group), or an implementation may support other resources (for
	// example, resource "myroutetargets" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "services".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	ForwardTo *ForwardToTarget `json:"forwardTo" protobuf:"bytes,1,opt,name=forwardTo"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmaps" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myrouteactions" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".  If the referent
	// cannot be found, the "InvalidRoutes" status condition on any Gateway
	// that includes the HTTPRoute will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteActionExtensionObjectReference `json:"extensionRef" protobuf:"bytes,2,opt,name=extensionRef"`
}

// ForwardToTarget identifies a target object within a known namespace.
type ForwardToTarget struct {
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

// TargetPort specifies the destination port number to use for a TargetRef.
type TargetPort int32

// ForwardToTargetObjectReference identifies a target object of a ForwardTo
// route action within a known namespace.
//
// +k8s:deepcopy-gen=false
type ForwardToTargetObjectReference = ServicesDefaultLocalObjectReference

// RouteActionExtensionObjectReference identifies a route-action extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteActionExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// RouteHostExtensionObjectReference identifies a route-host extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteHostExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// HTTPRouteStatus defines the observed state of HTTPRoute.
type HTTPRouteStatus struct {
	GatewayRefs []GatewayObjectReference `json:"gatewayRefs" protobuf:"bytes,1,rep,name=gatewayRefs"`
}

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

// +genclient
// +kubebuilder:object:root=true

// HTTPRoute is the Schema for the httproutes API
type HTTPRoute struct {
	metav1.TypeMeta   `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`

	Spec   HTTPRouteSpec   `json:"spec,omitempty" protobuf:"bytes,3,opt,name=spec"`
	Status HTTPRouteStatus `json:"status,omitempty" protobuf:"bytes,4,opt,name=status"`
}

// +kubebuilder:object:root=true

// HTTPRouteList contains a list of HTTPRoute
type HTTPRouteList struct {
	metav1.TypeMeta `json:",inline" protobuf:"bytes,1,opt,name=typeMeta"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,2,opt,name=metadata"`
	Items           []HTTPRoute `json:"items" protobuf:"bytes,3,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&HTTPRoute{}, &HTTPRouteList{})
}
