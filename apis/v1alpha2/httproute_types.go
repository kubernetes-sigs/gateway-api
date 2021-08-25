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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Hostnames",type=string,JSONPath=`.spec.hostnames`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// HTTPRoute is the Schema for the HTTPRoute resource.
type HTTPRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HTTPRoute.
	Spec HTTPRouteSpec `json:"spec"`

	// Status defines the current state of HTTPRoute.
	Status HTTPRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HTTPRouteList contains a list of HTTPRoute.
type HTTPRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HTTPRoute `json:"items"`
}

// HTTPRouteSpec defines the desired state of HTTPRoute
type HTTPRouteSpec struct {
	CommonRouteSpec `json:",inline"`

	// Hostnames defines a set of hostname that should match against
	// the HTTP Host header to select a HTTPRoute to process the request.
	// Hostname is the fully qualified domain name of a network host,
	// as defined by RFC 3986. Note the following deviations from the
	// "host" part of the URI as defined in the RFC:
	//
	// 1. IPs are not allowed.
	// 2. The `:` delimiter is not respected because ports are not allowed.
	//
	// If a hostname is specified by both the Listener and HTTPRoute, there
	// must be at least one intersecting hostname for the HTTPRoute to be
	// attached to the Listener. For example:
	//
	// * A Listener with `test.example.com` as the hostname matches HTTPRoutes
	//   that have either not specified any hostnames, or have specified at
	//   least one of `test.example.com` or `*.example.com`.
	// * A Listener with `*.example.com` as as the hostname matches HTTPRoutes
	//   that have either not specified any hostnames or have specified at least
	//   one hostname that matches the Listener hostname. For example,
	//   `test.example.com` and `*.example.com` would both match. On the other
	//   hand, `a.b.example.com`, `example.com`, and `test.example.net` would
	//   not match.
	//
	// If both the Listener and HTTPRoute have specified hostnames, any
	// HTTPRoute hostnames that do not match the Listener hostname MUST be
	// ignored. For example, if a Listener specified `*.example.com`, and the
	// HTTPRoute specified `test.example.com` and `test.example.net`,
	// `test.example.net` must not be considered for a match.
	//
	// If hostnames do not match with the criteria above, then the HTTPRoute is
	// not admitted, and the implementation must raise an 'Admitted' Condition
	// with a status of `False` for the target Listener(s).
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Hostnames []Hostname `json:"hostnames,omitempty"`

	// Rules are a list of HTTP matchers, filters and actions.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:default={{matches: {{path: {type: "Prefix", value: "/"}}}}}
	Rules []HTTPRouteRule `json:"rules,omitempty"`
}

// HTTPRouteRule defines semantics for matching an HTTP request based on
// conditions, optionally executing additional processing steps, and forwarding
// the request to an API object.
type HTTPRouteRule struct {
	// Matches define conditions used for matching the rule against incoming
	// HTTP requests. Each match is independent, i.e. this rule will be matched
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
	//
	// Each client request MUST map to a maximum of one route rule. If a request
	// matches multiple rules, matching precedence MUST be determined in order
	// of the following criteria, continuing on ties:
	//
	// * The longest matching hostname.
	// * The longest matching non-wildcard hostname.
	// * The longest matching path.
	// * The largest number of header matches.
	// * The largest number of query param matches.
	//
	// If ties still exist across multiple Routes, matching precedence MUST be
	// determined in order of the following criteria, continuing on ties:
	//
	// * The oldest Route based on creation timestamp. For example, a Route with
	//   a creation timestamp of "2020-09-08 01:02:03" is given precedence over
	//   a Route with a creation timestamp of "2020-09-08 01:02:04".
	// * The Route appearing first in alphabetical order by
	//   "<namespace>/<name>". For example, foo/bar is given precedence over
	//   foo/baz.
	//
	// If ties still exist within the Route that has been given precedence,
	// matching precedence MUST be granted to the first matching rule meeting
	// the above criteria.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{path:{ type: "Prefix", value: "/"}}}
	Matches []HTTPRouteMatch `json:"matches,omitempty"`

	// Filters define the filters that are applied to requests that match
	// this rule.
	//
	// The effects of ordering of multiple behaviors are currently unspecified.
	// This can change in the future based on feedback during the alpha stage.
	//
	// Conformance-levels at this level are defined based on the type of filter:
	//
	// - ALL core filters MUST be supported by all implementations.
	// - Implementers are encouraged to support extended filters.
	// - Implementation-specific custom filters have no API guarantees across
	//   implementations.
	//
	// Specifying a core filter multiple times has unspecified or custom conformance.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []HTTPRouteFilter `json:"filters,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent.

	// If unspecified or invalid (refers to a non-existent resource or a Service
	// with no endpoints), the rule performs no forwarding. If that are also no
	// filters specified that would result in a response being sent, a HTTP 503
	// status code is returned. 503 responses must be sent so that the overall
	// weight is respected; if an invalid backend is requested to have 80% of
	// requests, then 80% of requests must get a 503 instead.
	//
	// Support: Core for Kubernetes Service
	// Support: Custom for any other resource
	//
	// Support for weight: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []HTTPBackendRef `json:"backendRefs,omitempty"`
}

// PathMatchType specifies the semantics of how HTTP paths should be compared.
// Valid PathMatchType values are:
//
// * "Exact"
// * "Prefix"
// * "RegularExpression"
// * "ImplementationSpecific"
//
// Prefix and Exact paths must be syntactically valid:
//
// - Must begin with the '/' character
// - Must not contain consecutive '/' characters (e.g. /foo///, //).
// - For prefix paths, a trailing '/' character in the Path is ignored,
// e.g. /abc and /abc/ specify the same match.
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

// HTTPPathMatch describes how to select a HTTP route by matching the HTTP request path.
type HTTPPathMatch struct {
	// Type specifies how to match against the path Value.
	//
	// Support: Core (Exact, Prefix)
	//
	// Support: Custom (RegularExpression, ImplementationSpecific)
	//
	// Since RegularExpression PathType has custom conformance, implementations
	// can support POSIX, PCRE or any other dialects of regular expressions.
	// Please read the implementation's documentation to determine the supported
	// dialect.
	//
	// +optional
	// +kubebuilder:default=Prefix
	Type *PathMatchType `json:"type,omitempty"`

	// Value of the HTTP path to match against.
	//
	// +optional
	// +kubebuilder:default="/"
	// +kubebuilder:validation:MaxLength=1024
	Value *string `json:"value,omitempty"`
}

// HeaderMatchType specifies the semantics of how HTTP header values should be
// compared. Valid HeaderMatchType values are:
//
// * "Exact"
// * "RegularExpression"
// * "ImplementationSpecific"
//
// +kubebuilder:validation:Enum=Exact;RegularExpression;ImplementationSpecific
type HeaderMatchType string

// HeaderMatchType constants.
const (
	HeaderMatchExact                  HeaderMatchType = "Exact"
	HeaderMatchRegularExpression      HeaderMatchType = "RegularExpression"
	HeaderMatchImplementationSpecific HeaderMatchType = "ImplementationSpecific"
)

// HTTPHeaderName is the name of an HTTP header.
//
// Valid values include:
//
// * "Authorization"
// * "Set-Cookie"
//
// Invalid values include:
//
// * ":method" - ":" is an invalid character. This means that pseudo headers are
//   not currently supported by this type.
// * "/invalid" - "/" is an invalid character
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HTTPHeaderName string

// HTTPHeaderMatch describes how to select a HTTP route by matching HTTP request
// headers.
type HTTPHeaderMatch struct {
	// Type specifies how to match against the value of the header.
	//
	// Support: Core (Exact)
	//
	// Support: Custom (RegularExpression, ImplementationSpecific)
	//
	// Since RegularExpression PathType has custom conformance, implementations
	// can support POSIX, PCRE or any other dialects of regular expressions.
	// Please read the implementation's documentation to determine the supported
	// dialect.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name is the name of the HTTP Header to be matched. Name matching MUST be
	// case insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).
	//
	// If multiple entries specify equivalent header names, only the first
	// entry with an equivalent name MUST be considered for a match. Subsequent
	// entries with an equivalent header name MUST be ignored. Due to the
	// case-insensitivity of header names, "foo" and "Foo" are considered
	// equivalent.
	//
	// When a header is repeated in an HTTP request, it is
	// implementation-specific behavior as to how this is represented.
	// Generally, proxies should follow the guidance from the RFC:
	// https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.2 regarding
	// processing a repeated header, with special handling for "Set-Cookie".
	Name HTTPHeaderName `json:"name"`

	// Value is the value of HTTP Header to be matched.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

// QueryParamMatchType specifies the semantics of how HTTP query parameter
// values should be compared. Valid QueryParamMatchType values are:
//
// * "Exact"
// * "RegularExpression"
// * "ImplementationSpecific"
//
// +kubebuilder:validation:Enum=Exact;RegularExpression;ImplementationSpecific
type QueryParamMatchType string

// QueryParamMatchType constants.
const (
	QueryParamMatchExact                  QueryParamMatchType = "Exact"
	QueryParamMatchRegularExpression      QueryParamMatchType = "RegularExpression"
	QueryParamMatchImplementationSpecific QueryParamMatchType = "ImplementationSpecific"
)

// HTTPQueryParamMatch describes how to select a HTTP route by matching HTTP
// query parameters.
type HTTPQueryParamMatch struct {
	// Type specifies how to match against the value of the query parameter.
	//
	// Support: Extended (Exact)
	//
	// Support: Custom (RegularExpression, ImplementationSpecific)
	//
	// Since RegularExpression QueryParamMatchType has custom conformance,
	// implementations can support POSIX, PCRE or any other dialects of regular
	// expressions. Please read the implementation's documentation to determine
	// the supported dialect.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *QueryParamMatchType `json:"type,omitempty"`

	// Name is the name of the HTTP query param to be matched. This must be an
	// exact string match. (See
	// https://tools.ietf.org/html/rfc7230#section-2.7.3).
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	Name string `json:"name"`

	// Value is the value of HTTP query param to be matched.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	Value string `json:"value"`
}

// HTTPMethod describes how to select a HTTP route by matching the HTTP
// method as defined by
// [RFC 7231](https://datatracker.ietf.org/doc/html/rfc7231#section-4) and
// [RFC 5789](https://datatracker.ietf.org/doc/html/rfc5789#section-2).
// The value is expected in upper case.
// +kubebuilder:validation:Enum=GET;HEAD;POST;PUT;DELETE;CONNECT;OPTIONS;TRACE;PATCH
type HTTPMethod string

const (
	HTTPMethodGet     HTTPMethod = "GET"
	HTTPMethodHead    HTTPMethod = "HEAD"
	HTTPMethodPost    HTTPMethod = "POST"
	HTTPMethodPut     HTTPMethod = "PUT"
	HTTPMethodDelete  HTTPMethod = "DELETE"
	HTTPMethodConnect HTTPMethod = "CONNECT"
	HTTPMethodOptions HTTPMethod = "OPTIONS"
	HTTPMethodTrace   HTTPMethod = "TRACE"
	HTTPMethodPatch   HTTPMethod = "PATCH"
)

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
	// +optional
	// +kubebuilder:default={type: "Prefix", value: "/"}
	Path *HTTPPathMatch `json:"path,omitempty"`

	// Headers specifies HTTP request header matchers. Multiple match values are
	// ANDed together, meaning, a request must match all the specified headers
	// to select the route.
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []HTTPHeaderMatch `json:"headers,omitempty"`

	// QueryParams specifies HTTP query parameter matchers. Multiple match
	// values are ANDed together, meaning, a request must match all the
	// specified query parameters to select the route.
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	QueryParams []HTTPQueryParamMatch `json:"queryParams,omitempty"`

	// Method specifies HTTP method matcher.
	// When specified, this route will be matched only if the request has the
	// specified method.
	//
	// Support: Extended
	//
	// +optional
	Method *HTTPMethod `json:"method,omitempty"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "match" behavior. For example, resource "myroutematcher" in group
	// "networking.example.net". If the referent cannot be found, the rule is
	// not included in the route. The controller must ensure the "ResolvedRefs"
	// condition on the Route status is set to `status: False`.
	//
	// Support: Custom
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

// HTTPRouteFilter defines additional processing steps that must be completed
// during the request or response lifecycle. HTTPRouteFilters are meant as an
// extension point to express additional processing that may be done in Gateway
// implementations. Some examples include request or response modification,
// implementing authentication strategies, rate-limiting, and traffic shaping.
// API guarantee/conformance is defined based on the type of the filter.
// TODO(hbagdi): re-render CRDs once controller-tools supports union tags:
// - https://github.com/kubernetes-sigs/controller-tools/pull/298
// - https://github.com/kubernetes-sigs/controller-tools/issues/461
// +union
type HTTPRouteFilter struct {
	// Type identifies the type of filter to apply. As with other API fields,
	// types are classified into three conformance levels:
	//
	// - Core: Filter types and their corresponding configuration defined by
	//   "Support: Core" in this package, e.g. "RequestHeaderModifier". All
	//   implementations must support core filters.
	//
	// - Extended: Filter types and their corresponding configuration defined by
	//   "Support: Extended" in this package, e.g. "RequestMirror". Implementers
	//   are encouraged to support extended filters.
	//
	// - Custom: Filters that are defined and supported by specific vendors.
	//   In the future, filters showing convergence in behavior across multiple
	//   implementations will be considered for inclusion in extended or core
	//   conformance levels. Filter-specific configuration for such filters
	//   is specified using the ExtensionRef field. `Type` should be set to
	//   "ExtensionRef" for custom filters.
	//
	// Implementers are encouraged to define custom implementation types to
	// extend the core API with implementation-specific behavior.
	//
	// +unionDiscriminator
	Type HTTPRouteFilterType `json:"type"`

	// RequestHeaderModifier defines a schema for a filter that modifies request
	// headers.
	//
	// Support: Core
	//
	// +optional
	RequestHeaderModifier *HTTPRequestHeaderFilter `json:"requestHeaderModifier,omitempty"`

	// RequestMirror defines a schema for a filter that mirrors requests.
	//
	// Support: Extended
	//
	// +optional
	RequestMirror *HTTPRequestMirrorFilter `json:"requestMirror,omitempty"`

	// RequestRedirect defines a schema for a filter that responds to the
	// request with an HTTP redirection.
	//
	// Support: Core
	//
	// +optional
	RequestRedirect *HTTPRequestRedirect `json:"requestRedirect,omitempty"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "filter" behavior.  For example, resource "myroutefilter" in group
	// "networking.example.net"). ExtensionRef MUST NOT be used for core and
	// extended filters.
	//
	// Support: Implementation-specific
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}

// HTTPRouteFilterType identifies a type of HTTPRoute filter.
// +kubebuilder:validation:Enum=RequestHeaderModifier;RequestMirror;RequestRedirect;ExtensionRef
type HTTPRouteFilterType string

const (
	// HTTPRouteFilterRequestHeaderModifier can be used to add or remove an HTTP
	// header from an HTTP request before it is sent to the upstream target.
	//
	// Support in HTTPRouteRule: Core
	//
	// Support in HTTPBackendRef: Extended
	HTTPRouteFilterRequestHeaderModifier HTTPRouteFilterType = "RequestHeaderModifier"

	// HTTPRouteFilterRequestRedirect can be used to redirect a request to
	// another location. This filter can also be used for HTTP to HTTPS
	// redirects.
	//
	// Support in HTTPRouteRule: Core
	//
	// Support in HTTPBackendRef: Extended
	HTTPRouteFilterRequestRedirect HTTPRouteFilterType = "RequestRedirect"

	// HTTPRouteFilterRequestMirror can be used to mirror HTTP requests to a
	// different backend. The responses from this backend MUST be ignored by
	// the Gateway.
	//
	// Support in HTTPRouteRule: Extended
	//
	// Support in HTTPBackendRef: Extended
	HTTPRouteFilterRequestMirror HTTPRouteFilterType = "RequestMirror"

	// HTTPRouteFilterExtensionRef should be used for configuring custom
	// HTTP filters.
	//
	// Support in HTTPRouteRule: Custom
	//
	// Support in HTTPBackendRef: Custom
	HTTPRouteFilterExtensionRef HTTPRouteFilterType = "ExtensionRef"
)

// HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.
type HTTPHeader struct {
	// Name is the name of the HTTP Header to be matched. Name matching MUST be
	// case insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).
	//
	// If multiple entries specify equivalent header names, only the first entry
	// with an equivalent name MUST be considered for a match. Subsequent
	// entries with an equivalent header name MUST be ignored. Due to the
	// case-insensitivity of header names, "foo" and "Foo" are considered
	// equivalent.
	Name HTTPHeaderName `json:"name"`

	// Value is the value of HTTP Header to be matched.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

// HTTPRequestHeaderFilter defines configuration for the RequestHeaderModifier
// filter.
type HTTPRequestHeaderFilter struct {
	// Set overwrites the request with the given header (name, value)
	// before the action.
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//
	// Config:
	//   set: {"my-header": "bar"}
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header: bar
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Set []HTTPHeader `json:"set,omitempty"`

	// Add adds the given header(s) (name, value) to the request
	// before the action. It appends to any existing values associated
	// with the header name.
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//
	// Config:
	//   add: {"my-header": "bar"}
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header: foo
	//   my-header: bar
	//
	// +optional
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=16
	Add []HTTPHeader `json:"add,omitempty"`

	// Remove the given header(s) from the HTTP request before the action. The
	// value of Remove is a list of HTTP header names. Note that the header
	// names are case-insensitive (see
	// https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).
	//
	// Input:
	//   GET /foo HTTP/1.1
	//   my-header1: foo
	//   my-header2: bar
	//   my-header3: baz
	//
	// Config:
	//   remove: ["my-header1", "my-header3"]
	//
	// Output:
	//   GET /foo HTTP/1.1
	//   my-header2: bar
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Remove []string `json:"remove,omitempty"`
}

// HTTPRequestRedirect defines configuration for the RequestRedirect filter.
type HTTPRequestRedirect struct {
	// Protocol is the protocol to be used in the value of the `Location`
	// header in the response.
	// When empty, the protocol of the request is used.
	//
	// Support: Extended
	//
	// +optional
	// +kubebuilder:validation:Enum=HTTP;HTTPS
	Protocol *string `json:"protocol,omitempty"`

	// Hostname is the hostname to be used in the value of the `Location`
	// header in the response.
	// When empty, the hostname of the request is used.
	//
	// Support: Core
	//
	// +optional
	Hostname *Hostname `json:"hostname,omitempty"`

	// Port is the port to be used in the value of the `Location`
	// header in the response.
	// When empty, port (if specified) of the request is used.
	//
	// Support: Extended
	//
	// +optional
	Port *PortNumber `json:"port,omitempty"`

	// StatusCode is the HTTP status code to be used in response.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:default=302
	// +kubebuilder:validation:Enum=301;302
	StatusCode *int `json:"statusCode,omitempty"`
}

// HTTPRequestMirrorFilter defines configuration for the RequestMirror filter.
type HTTPRequestMirrorFilter struct {
	// BackendRef references a resource where mirrored requests are sent.
	//
	// If the referent cannot be found, this BackendRef is invalid and must be
	// dropped from the Gateway. The controller must ensure the "ResolvedRefs"
	// condition on the Route status is set to `status: False` and not configure
	// this backend in the underlying implementation.
	//
	// If there is a cross-namespace reference to an *existing* object
	// that is not allowed by a ReferencePolicy, the controller must ensure the
	// "ResolvedRefs"  condition on the Route is set to `status: False`,
	// with the "RefNotPermitted" reason and not configure this backend in the
	// underlying implementation.
	//
	// In either error case, the Message of the `ResolvedRefs` Condition
	// should be used to provide more detail about the problem.
	//
	// Support: Extended for Kubernetes Service
	// Support: Custom for any other resource
	//
	// +optional
	BackendRef *BackendObjectReference `json:"backendRef,omitempty"`
}

// HTTPBackendRef defines how a HTTPRoute should forward an HTTP request.
type HTTPBackendRef struct {
	// BackendRef is a reference to a backend to forward matched requests to.
	//
	// If the referent cannot be found, this HTTPBackendRef is invalid and must
	// be dropped from the Gateway. The controller must ensure the
	// "ResolvedRefs" condition on the Route is set to `status: False` and not
	// configure this backend in the underlying implementation.
	//
	// If there is a cross-namespace reference to an *existing* object
	// that is not covered by a ReferencePolicy, the controller must ensure the
	// "ResolvedRefs"  condition on the Route is set to `status: true`,
	// with the "RefNotPermitted" reason and not configure this backend in the
	// underlying implementation.
	//
	// In either error case, the Message of the `ResolvedRefs` Condition
	// should be used to provide more detail about the problem.
	//
	// Support: Custom
	//
	// +optional
	BackendRef `json:",inline"`

	// Filters defined at this level should be executed if and only if the
	// request is being forwarded to the backend defined here.
	//
	// Support: Custom (For broader support of filters, use the Filters field
	// in HTTPRouteRule.)
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []HTTPRouteFilter `json:"filters,omitempty"`
}

// HTTPRouteStatus defines the observed state of HTTPRoute.
type HTTPRouteStatus struct {
	RouteStatus `json:",inline"`
}
