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
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HTTPRouteSpec defines the desired state of HTTPRoute
type HTTPRouteSpec struct {
	// Hosts is a list of Host definitions.
	Hosts []HTTPRouteHost `json:"hosts,omitempty" protobuf:"bytes,1,rep,name=hosts"`
}

// HTTPRouteHost is the configuration for a given set of hosts.
type HTTPRouteHost struct {
	// Hostnames defines a set of hostname that should match against
	// the HTTP Host header to select a HTTPRoute to process the request.
	// Hostname is the fully qualified domain name of a network host,
	// as defined by RFC 3986. Note the following deviations from the
	// "host" part of the URI as defined in the RFC:
	//
	// 1. IPs are not allowed.
	// 2. The `:` delimiter is not respected because ports are not allowed.
	//
	// Incoming requests are matched against the hostnames before the
	// HTTPRoute rules. If no hostname is specified, traffic is routed
	// based on the HTTPRouteRules.
	//
	// Hostname can be "precise" which is a domain name without the terminating
	// dot of a network host (e.g. "foo.example.com") or "wildcard", which is
	// a domain name prefixed with a single wildcard label (e.g. "*.example.com").
	// The wildcard character '*' must appear by itself as the first DNS
	// label and matches only a single label.
	// You cannot have a wildcard label by itself (e.g. Host == "*").
	// Requests will be matched against the Host field in the following order:
	// 1. If Host is precise, the request matches this rule if
	//    the http host header is equal to Host.
	// 2. If Host is a wildcard, then the request matches this rule if
	//    the http host header is to equal to the suffix
	//    (removing the first label) of the wildcard rule.
	//
	// Support: Core
	//
	// +optional
	Hostnames []string `json:"hostnames,omitempty" protobuf:"bytes,1,opt,name=hostnames"`

	// Rules are a list of HTTP matchers, filters and actions.
	//
	// +kubebuilder:validation:MinItems=1
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

// HTTPRouteRule defines semantics for matching an incoming HTTP request against
// a set of matching rules and executing an action (and optionally filters) on
// the request.
type HTTPRouteRule struct {
	// Matches define conditions used for matching the rule against
	// incoming HTTP requests.
	// Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	// For example, take the following matches configuration:
	// matches:
	// - path: /foo
	//   headers:
	//     version: "2"
	// - path: /v2/foo
	// For a request to match against this rule, a request should satisfy
	// EITHER of the two conditions:
	// - path prefixed with `/foo` AND contains the header `version: "2"`
	// - path prefix of `/v2/foo`
	// Please see doc for HTTPRouteMatch on how to specify multiple
	// match conditions that should be ANDed together.
	// +optional
	Matches []HTTPRouteMatch `json:"matches" protobuf:"bytes,1,rep,name=matches"`

	// Filters define the filters that are applied to requests that match
	// this rule.
	//
	// The effects of ordering of multiple behaviors are currently undefined.
	// This can change in the future based on feedback during the alpha stage.
	//
	// Conformance-levels at this level are defined based on the type of filter:
	// - ALL core filters MUST be supported by all implementations.
	// - Implementers are encouraged to support extended filters.
	// - Implementation-specific custom filters have no API guarantees across implementations.
	// Specifying a core filter multiple times has undefined or custom conformance.
	//
	// Support: core
	//
	// +optional
	Filters []HTTPRouteFilter `json:"filters" protobuf:"bytes,2,rep,name=filters"`

	// Action defines what happens to the request.
	// +optional
	Action *HTTPRouteAction `json:"action" protobuf:"bytes,3,opt,name=action"`
}

// PathMatchType specifies the semantics of how HTTP paths should be compared.
// Valid PathMatchType values are:
//
// * "Exact"
// * "Prefix"
// * "RegularExpression"
// * "ImplementationSpecific"
//
// +kubebuilder:validation:Enum=Exact;Prefix;RegularExpression;ImplementationSpecific
type PathMatchType string

// PathMatchType constants.
const (
	PathMatchExact                  PathMatchType = "Exact"
	PathMatchPrefix                 PathMatchType = "Prefix"
	PathMatchRegularExpression      PathMatchType = "RegularExpression"
	PathMatchImplementationSpecific PathMatchType = "ImplementationSpecific"
)

// HeaderMatchType specifies the semantics of how HTTP headers should be compared.
// Valid HeaderMatchType values are:
//
// * "Exact"
// * "ImplementationSpecific"
//
// +kubebuilder:validation:Enum=Exact;ImplementationSpecific
type HeaderMatchType string

// HeaderMatchType constants.
const (
	// HeaderMatchTypeExact matches HTTP request-header fields.
	// Field name matches are case-insensitive while field value matches
	// are case-sensitive.
	HeaderMatchExact                  HeaderMatchType = "Exact"
	HeaderMatchImplementationSpecific HeaderMatchType = "ImplementationSpecific"
)

// HTTPRouteMatch defines the predicate used to match requests to a
// given action.
// Multiple match types are ANDed together, i.e. the match will evaluate
// to true only if all conditions are satisfied.
// For example:
//  match:
//    path: /foo
//    headers:
//      version: "1"
// will result in a match only if an HTTP request's path starts with `/foo` AND
// contains the `version: "1"` header.
type HTTPRouteMatch struct {
	// PathType defines the semantics of the `Path` matcher.
	//
	// Support: core (Exact, Prefix)
	// Support: custom (RegularExpression, ImplementationSpecific)
	//
	// Since RegularExpression PathType has custom conformance, implementations
	// can support POSIX, PCRE or any other dialects of regular expressions.
	// Please read the implementation's documentation to determine the supported
	// dialect.
	//
	// Default: "Prefix"
	//
	// +optional
	// +kubebuilder:default=Prefix
	PathMatchType PathMatchType `json:"pathMatchType" protobuf:"bytes,1,opt,name=pathMatchType"`

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
	// +kubebuilder:default=Exact
	HeaderMatchType *HeaderMatchType `json:"headerMatchType" protobuf:"bytes,3,opt,name=headerMatchType"`

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

const (
	// FilterHTTPReqeustHeader can be used to add or remove an HTTP
	// header from an HTTP request before it is sent to the upstream target.
	// Support: core
	// +optional
	FilterHTTPRequesttHeader = "RequestHeader"
	// TODO(hbagdi): add FilterHTTPRequestMirroring filter

	FilterImplementationSpecific = "ImplementationSpecific"
)

// HTTPRouteFilter defines additional processing steps that must be completed
// during the request or response lifecycle.
// HTTPRouteFilters are meant as an extension point to express additional
// processing that may be done in Gateway implementations. Some examples include
// request or response modification, implementing authentication strategies,
// rate-limiting, and traffic shaping.
// API guarantee/conformance is defined based on the type of the filter.
// TODO(hbagdi): re-render CRDs once controller-tools supports union tags:
// - https://github.com/kubernetes-sigs/controller-tools/pull/298
// - https://github.com/kubernetes-sigs/controller-tools/issues/461
// +union
type HTTPRouteFilter struct {
	// Type identifies the filter to execute.
	// Types are classified into three conformance-levels (similar to
	// other locations in this API):
	// - Core and extended: These filter types and their corresponding configuration
	//   is defined in this package. All implementations must implement
	//   the core filters. Implementers are encouraged to support extended filters.
	//   Definitions for filter-specific configuration for these
	//   filters is defined in this package.
	// - Custom: These filters are defined and supported by specific vendors.
	//   In the future, filters showing convergence in behavior across multiple
	//   implementations will be considered for inclusion in extended or core
	//   conformance rings. Filter-specific configuration for such filters
	//   is specified using the ExtensionRef field. `Type` should be set to
	//   "ImplementationSpecific" for custom filters.
	//
	// Implementers are encouraged to define custom implementation
	// types to extend the core API with implementation-specific behavior.
	//
	// +unionDiscriminator
	// +kubebuilder:validation:Required
	// +required
	Type string `json:"type" protobuf:"bytes,1,opt,name=type"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "filter" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutefilters" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".
	// ExtensionRef MUST NOT be used for core and extended filters.
	// +optional
	ExtensionRef *RouteFilterExtensionObjectReference `json:"extensionRef" protobuf:"bytes,2,opt,name=extensionRef"`

	// Filter-specific configuration definitions for core and extended filters

	RequestHeader *HTTPRequestHeaderConfig `json:"requestHeader" protobuf:"bytes,3,opt,name=requestHeader"`
}

// HTTPRequestHeaderConfig defines configuration for the
// RequestHeader filter.
type HTTPRequestHeaderConfig struct {
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

	// Remove the given header(s) from the HTTP request before the
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
	// ForwardTo sends requests to the referenced object(s).  The
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
	ForwardTo []ForwardToTarget `json:"forwardTo" protobuf:"bytes,1,rep,name=forwardTo"`

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

// RouteHostExtensionObjectReference identifies a route-host extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteHostExtensionObjectReference = ConfigMapsDefaultLocalObjectReference

// HTTPRouteStatus defines the observed state of HTTPRoute.
type HTTPRouteStatus struct {
	// Gateways is a list of the Gateways that are associated with the
	// HTTPRoute, and the status of the route with respect to each of these
	// Gateways.  When a Gateway selects this route, the controller that
	// manages the Gateway should add an entry to this list when the
	// controller first sees the route and should update the entry as
	// appropriate when the route is modified.
	Gateways []RouteGatewayStatus `json:"gateways" protobuf:"bytes,1,rep,name=gateways"`
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

// RouteGatewayStatus describes the status of a route with respect to an
// associated Gateway.
type RouteGatewayStatus struct {
	// GatewayRef is a reference to a Gateway object that is associated with
	// the route.
	GatewayRef GatewayObjectReference `json:"gatewayRef" protobuf:"bytes,4,rep,name=gatewayRef"`
	// Conditions describes the status of the route with respect to the
	// Gateway.  For example, the "Admitted" condition indicates whether the
	// route has been admitted or rejected by the Gateway, and why.  Note
	// that the route's availability is also subject to the Gateway's own
	// status conditions and listener status.
	Conditions []RouteCondition `json:"conditions,omitempty" protobuf:"bytes,3,rep,name=conditions"`
}

// RouteConditionType is a type of condition for a route.
type RouteConditionType string

const (
	// ConditionRouteAdmitted indicates whether the route has been admitted
	// or rejected by a Gateway, and why.
	ConditionRouteAdmitted RouteConditionType = "Admitted"
)

// RouteCondition is a status condition for a given route.
type RouteCondition struct {
	// Type indicates the type of condition.
	Type RouteConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=RouteConditionType"`
	// Status describes the current state of this condition.  Can be "True",
	// "False", or "Unknown".
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Message is a human-understandable message describing the condition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,3,opt,name=message"`
	// Reason indicates why the condition is in this state.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// LastTransitionTime indicates the last time this condition changed.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,5,opt,name=lastTransitionTime"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

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
