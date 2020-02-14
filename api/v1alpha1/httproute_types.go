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
	Hosts []HTTPRouteHost `json:"hosts,omitempty"`

	// Default is the default host to use. Default.Hostnames must
	// be an empty list.
	//
	// +optional
	Default *HTTPRouteHost `json:"default"`
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
	Hostname string `json:"hostname,omitempty"`

	// Rules are a list of HTTP matchers, filters and actions.
	Rules []HTTPRouteRule `json:"rules"`

	// Extension is an optional, implementation-specific extension to the
	// "host" block.  The resource may be "configmap" (use the empty string
	// for the group) or an implementation-defined resource (for example,
	// resource "myroutehost" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteHostExtensionObjectReference `json:"extension"`
}

// HTTPRouteRule is the configuration for a given path.
type HTTPRouteRule struct {
	// Match defines which requests match this path.
	// +optional
	Match *HTTPRouteMatch `json:"match"`
	// Filter defines what filters are applied to the request.
	// +optional
	Filter *HTTPRouteFilter `json:"filter"`
	// Action defines what happens to the request.
	// +optional
	Action *HTTPRouteAction `json:"action"`
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
	PathType string `json:"pathType"`
	// Path is the value of the HTTP path as interpreted via
	// PathType.
	//
	// Default: "/"
	Path *string `json:"path"`

	// HeaderType defines the semantics of the `Header` matcher.
	//
	// +optional
	HeaderType *string `json:"headerType"`
	// Header are the Header matches as interpreted via
	// HeaderType.
	//
	// +optional
	Header map[string]string `json:"header"`

	// Extension is an optional, implementation-specific extension to the
	// "match" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutematcher" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteMatchExtensionObjectReference `json:"extension"`
}

// RouteMatchExtensionObjectReference identifies a route-match extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteMatchExtensionObjectReference = LocalObjectReference

// HTTPRouteFilter defines a filter-like action to be applied to
// requests.
type HTTPRouteFilter struct {
	// Headers related filters.
	//
	// Support: extended
	// +optional
	Headers *HTTPHeaderFilter `json:"headers"`

	// Extension is an optional, implementation-specific extension to the
	// "filter" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutefilter" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteFilterExtensionObjectReference `json:"extension"`
}

// RouteFilterExtensionObjectReference identifies a route-filter extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteFilterExtensionObjectReference = LocalObjectReference

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
	Add map[string]string `json:"add"`

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
	Remove []string `json:"remove"`

	// TODO
}

// HTTPRouteAction is the action taken given a match.
type HTTPRouteAction struct {
	// ForwardTo sends requests to the referenced object.  The resource may
	// be "service" (use the empty string for the group), or an
	// implementation may support other resources (for example, resource
	// "myroutetarget" in group "networking.acme.io").
	ForwardTo *RouteActionTargetObjectReference `json:"forwardTo"`

	// Extension is an optional, implementation-specific extension to the
	// "action" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myrouteaction" in group "networking.acme.io").
	//
	// Support: custom
	//
	// +optional
	Extension *RouteActionExtensionObjectReference `json:"extension"`
}

// RouteActionTargetObjectReference identifies a target object for a route
// action within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteActionTargetObjectReference = LocalObjectReference

// RouteActionExtensionObjectReference identifies a route-action extension
// object within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteActionExtensionObjectReference = LocalObjectReference

// RouteHostExtensionObjectReference identifies a route-host extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteHostExtensionObjectReference = LocalObjectReference

// HTTPRouteStatus defines the observed state of HTTPRoute.
type HTTPRouteStatus struct {
	Gateways []GatewayObjectReference `json:"gateways"`
}

// GatewayObjectReference identifies a Gateway object.
type GatewayObjectReference struct {
	// Namespace is the namespace of the referent.
	// +optional
	Namespace string `json:"namespace,omitempty"`
	// Name is the name of the referent.
	//
	// +kubebuilder:validation:Required
	// +required
	Name string `json:"name"`
}

// +kubebuilder:object:root=true

// HTTPRoute is the Schema for the httproutes API
type HTTPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HTTPRouteSpec   `json:"spec,omitempty"`
	Status HTTPRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HTTPRouteList contains a list of HTTPRoute
type HTTPRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HTTPRoute{}, &HTTPRouteList{})
}
