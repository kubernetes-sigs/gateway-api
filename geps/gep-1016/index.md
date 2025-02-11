# GEP-1016: GRPCRoute

* Issue: [#1016](https://github.com/kubernetes-sigs/gateway-api/issues/1016)
* Status: Standard

> **Note**: This GEP is exempt from the [Probationary Period][expprob] rules of
> our GEP overview as it existed before those rules did, and so it has been
> explicitly grandfathered in.

[expprob]:https://gateway-api.sigs.k8s.io/geps/overview/#probationary-period

## Goal

Add an idiomatic GRPCRoute for routing gRPC traffic.

## Non-Goals

While certain gRPC implementations support multiple transports and multiple
interface definition languages (IDLs), this proposal limits itself to
[HTTP/2](https://developers.google.com/web/fundamentals/performance/http2) as
the transport and [Protocol Buffers](https://developers.google.com/protocol-buffers)
as the IDL, which makes up the vast majority of gRPC traffic in the wild.

## Introduction

While it would be possible to support gRPC via custom, out-of-tree CRDs, in the long run, this would
lead to a fragmented ecosystem.

gRPC is a [popular RPC framework adopted widely across the industry](https://grpc.io/about/#whos-using-grpc-and-why).
The protocol is used pervasively within the Kubernetes project itself as the basis for
many interfaces, including:

- [the CSI](https://github.com/container-storage-interface/spec/blob/5b0d4540158a260cb3347ef1c87ede8600afb9bf/spec.md),
- [the CRI](https://github.com/kubernetes/cri-api/blob/49fe8b135f4556ea603b1b49470f8365b62f808e/README.md),
- [the device plugin framework](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/)

Given gRPC's importance in the application-layer networking space and to
the Kubernetes project in particular, we must ensure that the gRPC control plane
configuration landscape does not Balkanize.

### Encapsulated Network Protocols

It is theoretically possible to route gRPC traffic using only `HTTPRoute`
resources, but there are several serious problems with forcing gRPC users to route traffic at
the level of HTTP. This is why we propose a new resource.

In setting this precedent, we must also introduce a coherent policy for _when_
to introduce a custom `Route` resource for an encapsulated protocol for which a
lower layer protocol already exists. We propose the following criteria for such
an addition.

- Users of the encapsulated protocol would miss out on significant conventional features from their ecosystem if forced to route at a lower layer.
- Users of the encapsulated protocol would experience a degraded user experience if forced to route at a lower layer.
- The encapsulated protocol has a significant user base, particularly in the Kubernetes community.

gRPC meets _all_ of these criteria and is therefore, we contend, a strong
candidate for inclusion in the Gateway API.

#### HTTP/2 Cleartext

gRPC allows HTTP/2 cleartext communication (H2C). This is conventionally deployed for
testing. Many control plane implementations do not support this by default and
would require special configuration to work properly.

#### Content-Based Routing

While not included in the scope of this initial GEP, a common use case cited for
routing gRPC is payload-aware routing. That is, routing rules which determine a
backend based on the contents of the protocol buffer payload.

#### User Experience

The user experience would also degrade significantly if forced to route at the level of HTTP.

- Encoding services and methods as URIs (an implementation detail of gRPC)
- The [Transfer Encoding header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Transfer-Encoding) for trailers
- Many features supported by HTTP/2 but not by gRPC, such as:
    - Query parameters
    - Methods besides `POST`
    - [CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)


#### Proxyless Service Mesh

The gRPC library supports proxyless service mesh, a system by which routing
configuration is received and acted upon not by an in-line proxy or sidecar
proxy but by the client itself. Eventually, `GRPCRoute` in the Gateway API
should support this feature.  However, to date, there are no HTTP client
libraries capable of participating in a proxyless service mesh.

--- 

### Cross Serving

Occasionally, gRPC users will place gRPC services on the same hostname/port
combination as HTTP services. For example, `foo.com:443/v1` might serve
REST+JSON while `foo.com:443/com.foo.WidgetService/` serves gRPC. Such an
arrangement in the Gateway API poses complex technical challenges. How are
GRPCRoutes to be reconciled with HTTPRoutes? And how should individual
implementations accomplisht this?

After a long look at the implementations with which the author is familiar, it
was deemed technically infeasible. Furthermore, after surveying the gRPC
community, this was found to be a niche use case to begin with.

In any case, users wishing to accomplish this always have the option of using
`HTTPRoute` resources to achieve this use case, at the cost of a degraded user
experience.

If at some point in the future, demand for this use case increases and we have
reason to believe that the feasibility of implementation has improved, this
would be a backward compatible change.

As such, implementations that support GRPCRoute must enforce uniqueness of
hostnames between `GRPCRoute`s and `HTTPRoute`s. If a route (A) of type `HTTPRoute` or
`GRPCRoute` is attached to a Listener and that listener already has another Route (B) of
the other type attached and the intersection of the hostnames of A and B is
non-empty, then the implementation must reject Route A. That is, the
implementation must raise an 'Accepted' condition with a status of 'False' in
the corresponding RouteParentStatus.


## API

The API deviates from `HTTPRoute` where it results in a better UX for gRPC
users, while mirroring it in all other cases.

### Example `GRPCRoute`

```yaml
kind: GRPCRoute
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: foo-grpcroute
spec:
  parentRefs:
  - name: my-gateway
  hostnames:
  - foo.com
  - bar.com
  rules:
  - matches:
      method:
        service: helloworld.Greeter
        method:  SayHello
      headers:
      - type: Exact
        name: magic
        value: foo

    filters:
    - type: RequestHeaderModifierFilter
      add:
        - name: my-header
          value: foo

    - type: RequestMirrorPolicyFilter
      destination:
        backendRef:
          name: mirror-svc

    backendRefs:
    - name: foo-v1
      weight: 90
    - name: foo-v2
      weight: 10
```

#### Method Matchers

It's been pointed out that the `method` field above stutters. That is, in order
to specify a method matcher, one must type the string `method` twice in a row.
This is an artifact of less-than-clear nomenclature within gRPC. There
_are_ alternatives for the naming here, but none of them would actually be an
improvement on the stutter. Consider the following URI:

`/foo.bar.v1.WidgetService/GetWidget`

- `/foo.bar.v1.WidgetService/GetWidget` is called the method or, less commonly, the _full_ method. 
- `foo.bar.v1.WidgetService` is called the service or, less commonly, the _full_ service (since `WidgetService` can reasonably be called the service)]
- `GetWidget` is called the method.

These terms _could_ be added in, but these names are found almost exclusively
within the various gRPC implementations. And inconsistently across those
implementations.

Therefore, we opt for the stutter over any of the longer names outlined above.

#### Matcher Types

`GRPCRoute` method matchers admits two types: `Exact` and `RegularExpression`.
If not specified, the match will be treated as type `Exact`. Method matchers
will act _as if_ a URI match had been used. A full matrix of equivalent behavior
is provided below:

##### Type Exact

|Service|Method|URI Matcher|
|----------|----------|-----------|
|Specified|Specified|Exact `/${SERVICE}/${METHOD}`|
|Specified|Unspecified|Prefix `/${SERVICE}/`|
|Unspecified|Specified|Suffix `/${METHOD}/` or Regex `/.+/${METHOD}`|
|Unspecified|Unspecified|Not allowed|

##### Type RegularExpression

|Service|Method|URI Matcher|
|----------|----------|-----------|
|Specified|Specified|Regex `/${SERVICE}/${METHOD}`|
|Specified|Unspecified|Regex `/${SERVICE}/.+`|
|Unspecified|Specified|Regex `/.+/${METHOD}`|
|Unspecified|Unspecified|Prefix `/`|

##### Method specified but not Service

In the table above, `Service` unspecified and `Method` specified with type Exact
is listed as being equivalent to a path matcher with type suffix or type regex.
We imagine that many GRPCRoute implementations will be done using translation to
`HTTPRoute`s. `HTTPRoute` does not support a Suffix matcher and its Regex
matcher is specified as "Implementation-specific" support. In order to accommodate
`GRPCRoute` implementations built on top of `HTTPRoute` implementations without
regex support, we list this particular case as having implementation-specific 
support within the context of `GRPCRoute`.

#### Transport

No new `ProtocolType` will be added. While gRPC _does_ have some special
HTTP usage (HTTP/2 cleartext and HTTP/2 without an upgrade from HTTP/1.1),
`GRPCRoute` will be used in conjunction with the existing `HTTP` and `HTTPS`
ProtocolTypes.

Implementations supporting `GRPCRoute` with the `HTTPS` `ProtocolType` must
accept HTTP/2 connections without an [initial upgrade from HTTP/1.1](https://datatracker.ietf.org/doc/html/rfc7230#section-6.7). If the
implementation does not support this, then it should raise a "Detached"
condition for the affected listener with a reason of "UnsupportedProtocol"

Implementations supporting `GRPCRoute` with the `HTTP` `ProtocolType` must
support cleartext HTTP/2 connections without an [initial upgrade from HTTP/1.1](https://datatracker.ietf.org/doc/html/rfc7230#section-6.7). If the implementation does not support this, then it
should raise a "Detached" condition for the affected listener with a reason of
"UnsupportedProtocol"


### Structs

{% raw%}
```go
// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:categories=gateway-api
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
// +kubebuilder:printcolumn:name="Hostnames",type=string,JSONPath=`.spec.hostnames`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// GRPCRoute provides a way to route gRPC requests. This includes the capability
// to match requests by hostname, gRPC service, gRPC method, or HTTP/2 header. Filters can be
// used to specify additional processing steps. Backends specify where matching
// requests should be routed.
//
// Implementations supporting `GRPCRoute` with the `HTTPS` `ProtocolType` must
// accept HTTP/2 connections without an initial upgrade from HTTP/1.1. If the
// implementation does not support this, then it should raise a "Detached"
// condition for the affected listener with a reason of "UnsupportedProtocol"
//
// Implementations supporting `GRPCRoute` with the `HTTP` `ProtocolType` must
// support cleartext HTTP/2 without an initial upgrade from HTTP/1.1. If the
// implementation does not support this, then it should raise a "Detached"
// condition for the affected listener with a reason of "UnsupportedProtocol"
//
// Support: Extended
type GRPCRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of GRPCRoute.
	Spec GRPCRouteSpec `json:"spec,omitempty"`

	// Status defines the current state of GRPCRoute.
	Status GRPCRouteStatus `json:"status,omitempty"`
}

// GRPCRouteStatus defines the observed state of GRPCRoute.
type GRPCRouteStatus struct {
	RouteStatus `json:",inline"`
}

// GRPCRouteSpec defines the desired state of GRPCRoute
type GRPCRouteSpec struct {
	CommonRouteSpec `json:",inline"`

	// Hostnames defines a set of hostname that should match against the GRPC
	// Host header to select a GRPCRoute to process the request. This matches
	// the RFC 1123 definition of a hostname with 2 notable exceptions:
	//
	// 1. IPs are not allowed.
	// 2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard
	//    label must appear by itself as the first label.
	//
	// If a hostname is specified by both the Listener and GRPCRoute, there
	// must be at least one intersecting hostname for the GRPCRoute to be
	// attached to the Listener. For example:
	//
	// * A Listener with `test.example.com` as the hostname matches GRPCRoutes
	//   that have either not specified any hostnames, or have specified at
	//   least one of `test.example.com` or `*.example.com`.
	// * A Listener with `*.example.com` as the hostname matches GRPCRoutes
	//   that have either not specified any hostnames or have specified at least
	//   one hostname that matches the Listener hostname. For example,
	//   `test.example.com` and `*.example.com` would both match. On the other
	//   hand, `example.com` and `test.example.net` would not match.
	//
	// If both the Listener and GRPCRoute have specified hostnames, any
	// GRPCRoute hostnames that do not match the Listener hostname MUST be
	// ignored. For example, if a Listener specified `*.example.com`, and the
	// GRPCRoute specified `test.example.com` and `test.example.net`,
	// `test.example.net` must not be considered for a match.
	//
	// If both the Listener and GRPCRoute have specified hostnames, and none
	// match with the criteria above, then the GRPCRoute is not accepted. The
	// implementation must raise an 'Accepted' Condition with a status of
	// `False` in the corresponding RouteParentStatus.
	//
	// If a Route (A) of type HTTPRoute or GRPCRoute is attached to a
	// Listener and that listener already has another Route (B) of the other
	// type attached and the intersection of the hostnames of A and B is
	// non-empty, then the implementation must reject Route A. That is, the
	// implementation must raise an 'Accepted' condition with a status of
	// 'False' in the corresponding RouteParentStatus.
	//
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Hostnames []Hostname `json:"hostnames,omitempty"`

	// Rules are a list of GRPC matchers, filters and actions.
        // 
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:default={{matches: {{method: {type: "Exact"}}}}}
	Rules []GRPCRouteRule `json:"rules,omitempty"`
}

// GRPCRouteRule defines semantics for matching an gRPC request based on
// conditions (matches), processing it (filters), and forwarding the request to
// an API object (backendRefs).
type GRPCRouteRule struct {
	// Matches define conditions used for matching the rule against incoming
	// gRPC requests. Each match is independent, i.e. this rule will be matched
	// if **any** one of the matches is satisfied.
	//
	// For example, take the following matches configuration:
	//
	// ```
	// matches:
	// - method:
	//     service: foo.bar
	//   headers:
	//     values:
	//       version: 2
	// - method:
	//     service: foo.bar.v2
	// ```
	//
	// For a request to match against this rule, a request should satisfy
	// EITHER of the two conditions:
	//
	// - service of foo.bar AND contains the header `version: 2`
	// - service of foo.bar.v2
	//
	// See the documentation for GRPCRouteMatch on how to specify multiple
	// match conditions that should be ANDed together.
	//
	// If no matches are specified, the implementation must match every gRPC request.
	//
	// Proxy or Load Balancer routing configuration generated from GRPCRoutes
	// MUST prioritize rules based on the following criteria, continuing on
	// ties. Merging must not be done between GRPCRoutes and HTTPRoutes.
	// Precedence must be given to the rule with the largest number of:
	//
	// * Characters in a matching non-wildcard hostname.
	// * Characters in a matching hostname.
        // * Characters in a matching service.
        // * Characters in a matching method.
	// * Header matches.
	//
	// If ties still exist across multiple Routes, matching precedence MUST be
	// determined in order of the following criteria, continuing on ties:
	//
	// * The oldest Route based on creation timestamp.
	// * The Route appearing first in alphabetical order by
	//   "{namespace}/{name}".
	//
	// If ties still exist within the Route that has been given precedence,
	// matching precedence MUST be granted to the first matching rule meeting
	// the above criteria.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:default={{method: {type: "Exact"}}}
	Matches []GRPCRouteMatch `json:"matches,omitempty"`

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
	// Specifying a core filter multiple times has unspecified or 
	// implementation-specific conformance.
	// Support: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []GRPCRouteFilter `json:"filters,omitempty"`

	// BackendRefs defines the backend(s) where matching requests should be
	// sent.

	// If unspecified or invalid (refers to a nonexistent resource or a Service
	// with no endpoints), the rule performs no forwarding. If there are also no
	// filters specified that would result in a response being sent, a gRPC `UNAVAILABLE`
	// status is returned. `UNAVAILABLE` responses must be sent so that the overall
	// weight is respected; if an invalid backend is requested to have 80% of
	// requests, then 80% of requests must get a `UNAVAILABLE` instead.
	// Support: Core for Kubernetes Service
	// Support: Implementation-specific for any other resource
	//
	// Support for weight: Core
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	BackendRefs []GRPCBackendRef `json:"backendRefs,omitempty"`
}

// GRPCRouteMatch defines the predicate used to match requests to a given
// action. Multiple match types are ANDed together, i.e. the match will
// evaluate to true only if all conditions are satisfied.
//
// For example, the match below will match a gRPC request only if its service
// is `foo` AND it contains the `version: v1` header:
//
// ```
// match:
//   method:
//     type: Exact
//     service: "foo"
//   headers:
//   - name: "version"
//     value "v1"
// ```
type GRPCRouteMatch struct {
	// Path specifies a gRPC request service/method matcher. If this field is not
	// specified, all services and methods will match.
	//
	// +optional
	// +kubebuilder:default={type: "Exact"}
	Method *GRPCMethodMatch `json:"path,omitempty"`

	// Headers specifies gRPC request header matchers. Multiple match values are
	// ANDed together, meaning, a request must match all the specified headers
	// to select the route.
	//
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []GRPCHeaderMatch `json:"headers,omitempty"`
}

// GRPCPathMatch describes how to select a gRPC route by matching the gRPC
// request service and/or method..
//
// At least one of Service and Method must be a non-empty string.
type GRPCMethodMatch struct {
	// Type specifies how to match against the service and/or method.
	// Support: Core (Exact with service and method specified)
	//
	// Support Implementation-specific (Exact with method specified but no 
	// service specified)
	//
	// Support: Implementation-specific (RegularExpression)
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *GRPCMethodMatchType `json:"type,omitempty"`


	// Value of the service to match against. If left empty or omitted, will
	// match all services.
        //
        // At least one of Service and Method must be a non-empty string.
	// +optional
	// +kubebuilder:default=""
	// +kubebuilder:validation:MaxLength=1024
	Service *string `json:"value,omitempty"`

	// Value of the method to match against. If left empty or omitted, will
	// match all services.
        //
        // At least one of Service and Method must be a non-empty string.
	// +optional
	// +kubebuilder:default=""
	// +kubebuilder:validation:MaxLength=1024
	Method *string `json:"value,omitempty"`
}

// MethodMatchType specifies the semantics of how gRPC methods and services should be compared.
// Valid MethodMatchType values are:
//
// * "Exact"
// * "RegularExpression"
//
// Exact paths must be syntactically valid:
//
// - Must not contain `/` character
//
// +kubebuilder:validation:Enum=Exact;PathPrefix;RegularExpression
// +kubebuilder:validation:Enum=Exact;RegularExpression
type GRPCMethodMatchType string

const (
	// Matches the service and/or method exactly and with case sensitivity.
	PathMatchExact PathMatchType = "Exact"

	// Matches if the service and/or method matches the given regular expression with
	// case sensitivity.
	//
	// Since `"RegularExpression"` has custom conformance, implementations
	// can support POSIX, PCRE, RE2 or any other regular expression dialect.
	// Please read the implementation's documentation to determine the supported
	// dialect.
	PathMatchRegularExpression PathMatchType = "RegularExpression"
)

// GRPCHeaderMatch describes how to select a gRPC route by matching gRPC request
// headers.
type GRPCHeaderMatch struct {
	// Type specifies how to match against the value of the header.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name is the name of the gRPC Header to be matched.
	//
	// If multiple entries specify equivalent header names, only the first
	// entry with an equivalent name MUST be considered for a match. Subsequent
	// entries with an equivalent header name MUST be ignored. Due to the
	// case-insensitivity of header names, "foo" and "Foo" are considered
	// equivalent.
	Name HeaderName `json:"name"`

	// Value is the value of the gRPC Header to be matched.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}

// +kubebuilder:validation:Enum=Exact;RegularExpression
type HeaderMatchType string

// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=256
// +kubebuilder:validation:Pattern=`^[A-Za-z0-9!#$%&'*+\-.^_\x60|~]+$`
type HeaderName string

// GRPCBackendRef defines how a GRPCRoute should forward a gRPC request.
type GRPCBackendRef struct {
	// BackendRef is a reference to a backend to forward matched requests to.
	//
	// If the referent cannot be found, this GRPCBackendRef is invalid and must
	// be dropped from the Gateway. The controller must ensure the
	// "ResolvedRefs" condition on the Route is set to `status: False` and not
	// configure this backend in the underlying implementation.
	//
	// If there is a cross-namespace reference to an *existing* object
	// that is not covered by a ReferenceGrant, the controller must ensure the
	// "ResolvedRefs"  condition on the Route is set to `status: False`,
	// with the "RefNotPermitted" reason and not configure this backend in the
	// underlying implementation.
	//
	// In either error case, the Message of the `ResolvedRefs` Condition
	// should be used to provide more detail about the problem.
	//
	// Support: Implementation-specific
	//
	// +optional
	BackendRef `json:",inline"`

	// Filters defined at this level should be executed if and only if the
	// request is being forwarded to the backend defined here.
	//
	// Support: Implementation-specific (For broader support of filters, use the Filters field
	// in GRPCRouteRule.)
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Filters []GRPCRouteFilter `json:"filters,omitempty"`
}

// GRPCRouteFilter defines processing steps that must be completed during the
// request or response lifecycle. GRPCRouteFilters are meant as an extension
// point to express processing that may be done in Gateway implementations. Some
// examples include request or response modification, implementing
// authentication strategies, rate-limiting, and traffic shaping. API
// guarantee/conformance is defined based on the type of the filter.
type GRPCRouteFilter struct {
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
	// If a reference to a custom filter type cannot be resolved, the filter
	// MUST NOT be skipped. Instead, requests that would have been processed by
	// that filter MUST receive a HTTP error response.
	//
	// +unionDiscriminator
	// +kubebuilder:validation:Enum=RequestHeaderModifier;RequestMirror;ExtensionRef
	// <gateway:experimental:validation:Enum=RequestHeaderModifier;RequestMirror;ExtensionRef>
	Type GRPCRouteFilterType `json:"type"`

	// RequestHeaderModifier defines a schema for a filter that modifies request
	// headers.
	//
	// Support: Core
	//
	// Support: Core
	//
	// +optional
	RequestHeaderModifier *HTTPRequestHeaderFilter `json:"requestHeaderModifier,omitempty"`

	// RequestMirror defines a schema for a filter that mirrors requests.
	// Requests are sent to the specified destination, but responses from
	// that destination are ignored.
	//
	// Support: Extended
	//
	// +optional
	RequestMirror *HTTPRequestMirrorFilter `json:"requestMirror,omitempty"`

	// ExtensionRef is an optional, implementation-specific extension to the
	// "filter" behavior.  For example, resource "myroutefilter" in group
	// "networking.example.net"). ExtensionRef MUST NOT be used for core and
	// extended filters.
	//
	// Support: Implementation-specific
	// Support: Implementation-specific
	//
	// +optional
	ExtensionRef *LocalObjectReference `json:"extensionRef,omitempty"`
}
```
{% endraw%}




## Beta Graduation Criteria

- `GRPCRoute` has been implemented by at least 2 controllers.
- Conformance tests are in place for the majority of the API surface.
- It is known that users of `GRPCRoute` exist.
- An API review has been performed by upstream Kubernetes reviewers.

## GA Graduation Criteria

- `GRPCRoute` has been implemented by at least 4 controllers.
- Exhaustive conformance tests are in place.
- Adoption of `GRPCRoute` has been shown to have expanded beyond its initial set of users.

## Future Enhancements

Many more ideas have been discussed for the `GRPCRoute` resource, but in the
interest of keeping this particular proposal tractable, they have been deferred
for future proposals. Enough thought has been given to these use cases at the
moment, however, that all of the following may be added at a later date in a
backward-compatible manner.

Some of these ideas are:

- Integration with Service Meshes (both sidecar-proxied and proxyless)
- Better UX for enabling reflection support
- gRPC Web support
- HTTP/JSON transcoding at the gateway
- Protobuf payload-based routing
