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
// +kubebuilder:printcolumn:name="Hostnames",type=string,JSONPath=`.spec.hostnames`

// HTTPRoute is the Schema for the HTTPRoute resource.
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

// HTTPRouteSpec defines the desired state of HTTPRoute
type HTTPRouteSpec struct {
	// Gateways defines which Gateways can use this Route.
	// +kubebuilder:default={allow: "SameNamespace"}
	Gateways RouteGateways `json:"gateways,omitempty"`

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
	// +kubebuilder:validation:MaxItems=16
	Hostnames []HTTPRouteHostname `json:"hostnames,omitempty"`

	// TLS defines the TLS certificate to use for Hostnames defined in this
	// Route. This configuration only takes effect if the AllowRouteOverride
	// field is set to true in the associated Gateway resource.
	//
	// Collisions can happen if multiple HTTPRoutes define a TLS certificate
	// for the same hostname. In such a case, conflict resolution guiding
	// principles apply, specificallly, if hostnames are same and two different
	// certificates are specified then the certificate in the
	// oldest resource wins.
	//
	// Please note that HTTP Route-selection takes place after the
	// TLS Handshake (ClientHello). Due to this, TLS certificate defined
	// here will take precedence even if the request has the potential to
	// match multiple routes (in case multiple HTTPRoutes share the same
	// hostname).
	//
	// Support: Core
	//
	// +optional
	TLS *RouteTLSConfig `json:"tls,omitempty"`

	// Rules are a list of HTTP matchers, filters and actions.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=16
	Rules []HTTPRouteRule `json:"rules"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "host" block. The resource may be "configmaps"  or an implementation-defined
	// resource (for example, resource "myroutehosts" in group "networking.acme.io").
	//
	// If the referent cannot be found,
	// the GatewayClass's "InvalidParameters" status condition
	// will be true.
	//
	// Support: custom
	//
	// +optional
	ExtensionRef *RouteHostExtensionObjectReference `json:"extensionRef,omitempty"`
}

// RouteTLSConfig describes a TLS configuration defined at the Route level.
type RouteTLSConfig struct {
	// CertificateRef refers to a Kubernetes object that
	// contains a TLS certificate and private key.
	// This certificate MUST be used for TLS handshakes for the domain
	// this RouteTLSConfig is associated with.
	// If an entry in this list omits or specifies the empty
	// string for both the group and the resource, the resource defaults to "secrets".
	// An implementation may support other resources (for example, resource
	// "mycertificates" in group "networking.acme.io").
	// Support: Core (Kubernetes Secrets)
	// Support: Implementation-specific (Other resource types)
	//
	// +required
	CertificateRef LocalObjectReference `json:"certificateRef"`
}

// HTTPRouteHostname is used to specify a hostname that should be matched by
// HTTPRoute.
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
type HTTPRouteHostname string

// HTTPRouteRule defines semantics for matching an incoming HTTP request against
// a set of matching rules and executing an action (and optionally filters) on
// the request.
type HTTPRouteRule struct {
	// Matches define conditions used for matching the rule against
	// incoming HTTP requests.
	// Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	//
	// For example, take the following matches configuration:
	//
	// ```
	// matches:
	// - path:
	//     value: "/foo"
	//   headers:
	//     values:
	//       version: "2"
	// - path:
	//     value: "/v2/foo"
	// ```
	//
	// For a request to match against this rule, a request should satisfy
	// EITHER of the two conditions:
	//
	// - path prefixed with `/foo` AND contains the header `version: "2"`
	// - path prefix of `/v2/foo`
	//
	// See the documentation for HTTPRouteMatch on how to specify multiple
	// match conditions that should be ANDed together.
	//
	// If no matches are specified, the default is a prefix
	// path match on "/", which has the effect of matching every
	// HTTP request.
	//
	// +kubebuilder:default={{path:{ type: "Prefix", value: "/"}}}
	// +kubebuilder:validation:MaxItems=8
	Matches []HTTPRouteMatch `json:"matches,omitempty"`

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
	// +kubebuilder:validation:MaxItems=16
	Filters []HTTPRouteFilter `json:"filters,omitempty"`

	// ForwardTo defines the backend(s) where matching requests should be sent.
	// +optional
	// +kubebuilder:validation:MaxItems=4
	ForwardTo []HTTPRouteForwardTo `json:"forwardTo,omitempty"`
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

// HTTPPathMatch describes how to select a HTTP route by matching the HTTP request path.
type HTTPPathMatch struct {
	// Type specifies how to match against the path Value.
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
	// +kubebuilder:default=Prefix
	Type PathMatchType `json:"type,omitempty"`

	// Value of the HTTP path to match against.
	//
	// +kubebuilder:validation:MinLength=1
	Value string `json:"value"`
}

// HTTPHeaderMatch describes how to select a HTTP route by matching HTTP request headers.
type HTTPHeaderMatch struct {
	// HeaderMatchType specifies how to match a HTTP request
	// header against the Values map.
	//
	// Support: core (Exact)
	// Support: custom (ImplementationSpecific)
	//
	// Default: "Exact"
	//
	// +kubebuilder:default=Exact
	Type HeaderMatchType `json:"type,omitempty"`

	// Values is a map of HTTP Headers to be matched.
	// It MUST contain at least one entry.
	//
	// The HTTP header field name to match is the map key, and the
	// value of the HTTP header is the map value. HTTP header field
	// names MUST be matched case-insensitively.
	//
	// Multiple match values are ANDed together, meaning, a request
	// must match all the specified headers to select the route.
	Values map[string]string `json:"values"`
}

// HTTPRouteMatch defines the predicate used to match requests to a given
// action. Multiple match types are ANDed together, i.e. the match will
// evaluate to true only if all conditions are satisfied.
//
// For example, the match below will match a HTTP request only if its path
// starts with `/foo` AND it contains the `version: "1"` header:
//
// ```
// match:
//   path:
//     value: "/foo"
//   headers:
//     values:
//       version: "1"
// ```
type HTTPRouteMatch struct {
	// Path specifies a HTTP request path matcher. If this field is not
	// specified, a default prefix match on the "/" path is provided.
	//
	// +kubebuilder:default={type: "Prefix", value: "/"}
	Path HTTPPathMatch `json:"path,omitempty"`

	// Headers specifies a HTTP request header matcher.
	//
	// +optional
	Headers *HTTPHeaderMatch `json:"headers"`

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
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

const (
	// FilterTypeHTTPRequestHeader can be used to add or remove an HTTP
	// header from an HTTP request before it is sent to the upstream target.
	// Support: core
	FilterTypeHTTPRequestHeader = "RequestHeader"

	// FilterTypeHTTPRequestMirror can be used to mirror requests to a
	// different backend. The responses from this backend MUST be ignored
	// by the Gateway.
	// Support: extended
	FilterTypeHTTPRequestMirror = "RequestMirror"

	// FilterTypeImplementationSpecific should be used for configuring
	// custom filters.
	FilterTypeImplementationSpecific = "ImplementationSpecific"
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
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=100
	Type string `json:"type"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "filter" behavior.  The resource may be "configmap" (use the empty
	// string for the group) or an implementation-defined resource (for
	// example, resource "myroutefilters" in group "networking.acme.io").
	// Omitting or specifying the empty string for both the resource and
	// group indicates that the resource is "configmaps".
	// ExtensionRef MUST NOT be used for core and extended filters.
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`

	// Filter-specific configuration definitions for core and extended filters

	// +optional
	RequestHeader *HTTPRequestHeaderFilter `json:"requestHeader,omitempty"`

	// +optional
	RequestMirror *HTTPRequestMirrorFilter `json:"requestMirror,omitempty"`
}

// HTTPRequestHeaderFilter defines configuration for the
// RequestHeader filter.
type HTTPRequestHeaderFilter struct {
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
	// Support: Extended
	Add map[string]string `json:"add"`

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
	// Support: Extended
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove"`
}

// HTTPRequestMirrorFilter defines configuration for the RequestMirror filter.
type HTTPRequestMirrorFilter struct {
	// ServiceName refers to the name of the Service to mirror matched requests
	// to. When specified, this takes the place of BackendRef. If both
	// BackendRef and ServiceName are specified, ServiceName will be given
	// precedence. If the referent cannot be found, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this
	// Route to true.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	ServiceName *string `json:"serviceName,omitempty"`

	// BackendRef is a local object reference to mirror matched requests to. If
	// both BackendRef and ServiceName are specified, ServiceName will be given
	// precedence. If the referent cannot be found, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this Route
	// to true.
	//
	// Support: Custom
	//
	// +optional
	BackendRef *LocalObjectReference `json:"backendRef,omitempty"`

	// Port specifies the destination port number to use for the backend
	// referenced by the ServiceName or BackendRef field. If unspecified and a
	// Service object consisting of a single port definition is the backend,
	// that port will be used. If unspecified and the backend is a Service
	// object consisting of multiple port definitions, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this Route
	// to true.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`
}

// HTTPRouteForwardTo defines how a HTTPRoute should forward a request.
type HTTPRouteForwardTo struct {
	// ServiceName refers to the name of the Service to forward matched requests
	// to. When specified, this takes the place of BackendRef. If both
	// BackendRef and ServiceName are specified, ServiceName will be given
	// precedence. If the referent cannot be found, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this
	// Route to true.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	ServiceName *string `json:"serviceName,omitempty"`

	// BackendRef is a reference to a backend to forward matched requests to. If
	// both BackendRef and ServiceName are specified, ServiceName will be given
	// precedence. If the referent cannot be found, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this Route
	// to true.
	//
	// Support: Custom
	//
	// +optional
	BackendRef *LocalObjectReference `json:"backendRef,omitempty"`

	// Port specifies the destination port number to use for the backend
	// referenced by the ServiceName or BackendRef field. If unspecified and a
	// Service object consisting of a single port definition is the backend,
	// that port will be used. If unspecified and the backend is a Service
	// object consisting of multiple port definitions, controllers must set the
	// "InvalidRoutes" status condition on any Gateway that includes this Route
	// to true.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// Weight specifies the proportion of traffic forwarded to the backend
	// referenced by the ServiceName or BackendRef field. computed as
	// weight/(sum of all weights in this ForwardTo list). Weight is not a
	// percentage and the sum of weights does not need to equal 100. If only one
	// backend is specified, 100% of the traffic is forwarded to that backend.
	// If unspecified, weight defaults to 1.
	//
	// Support: Core
	//
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10000
	Weight int32 `json:"weight,omitempty"`

	// Filters defined at this-level should be executed if and only if the
	// request is being forwarded to the backend defined here.
	//
	// Conformance: Filtering support, including core filters, is NOT guaranteed
	// at this-level. Use Filters in HTTPRouteRule for portable filters across
	// implementations.
	//
	// Support: Custom
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []HTTPRouteFilter `json:"filters,omitempty"`
}

// RouteHostExtensionObjectReference identifies a route-host extension object
// within a known namespace.
//
// +k8s:deepcopy-gen=false
type RouteHostExtensionObjectReference = LocalObjectReference

// HTTPRouteStatus defines the observed state of HTTPRoute.
type HTTPRouteStatus struct {
	RouteStatus `json:",inline"`
}
