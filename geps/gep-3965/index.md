# GEP-3965: HTTPRoute Implementation-Specific Matches

* Issue: [#3965](https://github.com/kubernetes-sigs/gateway-api/issues/3965)
* Status: Experimental

(See [status definitions](../overview.md#gep-states).)

## TLDR

Add an implementation-specific match extension point to `HTTPRoute`.

This gives implementations a supported place for custom matching such as CEL,
WASM, or other implementation-defined logic without trying to standardize those
mechanisms in core Gateway API.

## Goals

- Add a clear extension point for custom HTTP request matching.
- Keep the built-in `HTTPRoute` matches portable and unchanged.
- Avoid pushing custom matching into annotations or side APIs.

## Non-Goals

- Standardize CEL, WASM, or any other custom matcher language.
- Replace existing `HTTPRoute` matches.
- Guarantee portability for implementation-specific matches.

## Introduction

`HTTPRoute` has solid built-in matching for common cases: path, headers, query
params, and method. That covers the basics, but some implementations need more.

Examples include matching with CEL, WASM, or platform-specific request
attributes. Today there is no clean way to express that in `HTTPRoute`.
Implementations either rely on annotations or make users step outside the Route
API entirely.

This GEP proposes adding a narrow extension point for custom matching in
`HTTPRoute` with a typed inline string value.

This also revisits [GEP-820](../gep-820/index.md), which removed route match
extension points because there were no concrete use cases at the time. We now
have concrete use cases:

* CEL-based matching for complex request logic. CEL is quickly becoming a popular choice
  for policies both in Gateway API sub-projects (AI Gateway and Agentic Networking) and
  implementation-specific policies.
* A custom matcher could allow a CEL matcher, which would enable use cases like [routing based on JWT claims](https://istio.io/latest/docs/tasks/security/authentication/jwt-route/) or [routing based on body](https://gateway-api-inference-extension.sigs.k8s.io/#concepts-and-definitions).

## Purpose (Why and Who)

This is for implementations that already support richer matching and for users
who need it without giving up on `HTTPRoute`.

It also gives the project a cleaner answer than annotations. Core matches stay
standard. Custom matches stay explicitly implementation-specific.

## API

Add an `extension` field to `HTTPRouteMatch`.

```go
type HTTPRouteMatch struct {
	// Existing fields omitted.

	// Extension is an optional, implementation-specific extension to the
	// request matching rules. Extension MUST NOT be used for core or extended
	// Gateway API matching mechanisms.
	//
	// Support: Implementation-specific
	//
	// +optional
	// <gateway:experimental>
	Extension *HTTPRouteMatchExtension `json:"extension,omitempty"`
}

type HTTPRouteMatchExtension struct {
	// Type identifies how Value should be interpreted.
	//
	// This may take two possible forms:
	//
	// * A predefined CamelCase string identifier.
	// * A domain-prefixed string identifier, such as
	//   `example.com/CustomMatcher`.
	//
	// Support: Implementation-specific
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^([A-Za-z0-9]+|[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*\/[A-Za-z0-9\/\-._~%!$&'()*+,;=:]+)$`
	Type HTTPRouteMatchExtensionType `json:"type"`

	// Value contains the implementation-specific matcher expression.
	// The validity and meaning of this value depend on Type.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`
}
```

```go
type HTTPRouteMatchExtensionType string
```

`extension` composes with the existing fields on `HTTPRouteMatch`. A request
matches an `HTTPRouteMatch` only when it satisfies all specified match criteria,
including the implementation-specific matcher. If a match contains only
`extension`, the extension matcher is the entire predicate for that match.

### Ordering and precedence

Extension matchers do not have a portable Gateway API precision model. Each
extension type defines its own matcher semantics, so Gateway API cannot define a
standard way to compare how specific two extension matches are.

Implementations MAY define how a supported extension type participates in
HTTPRoute match ordering. This can include inserting extension-specific
precedence anywhere relative to the existing HTTPRoute match precedence criteria,
such as path, method, header count, and query param count. Implementations that
define precedence for an extension type MUST document that behavior for that
extension type.

For matches that do not use `extension`, the existing HTTPRoute precedence rules
are unchanged.

For example, an implementation supporting `example.com/CEL` could document an
ordering where path precedence is evaluated first, then an implementation-defined
ranking for `example.com/CEL`, then method, header, and query param criteria.
This GEP does not standardize that ranking.

The `type` plus `value` shape follows the same extensible pattern used by
Gateway addresses: well-known CamelCase types can be added in future releases,
and implementation-specific types use domain-prefixed strings. This GEP does not
define any portable matcher types.

The `HTTPRouteMatchExtension` struct is intentionally a wrapper rather than a
bare string field. This preserves room for a future GEP to add another
representation, such as a reference to a matcher resource, without changing the
top-level `HTTPRouteMatch` field or its composition semantics.

If an implementation does not support the specified extension type, the
implementation MUST set the `Accepted` condition to `False` for the relevant
`RouteParentStatus` with a reason of `UnsupportedValue`. If an implementation
supports the specified extension type but the value is invalid for that type, the
implementation MUST set the `Accepted` condition to `False` for the relevant
`RouteParentStatus` with a reason of `UnsupportedValue`.

Implementations MUST NOT use `extension` to define behavior for portable
Gateway API match mechanisms. Any behavior behind `extension` is
implementation-specific unless standardized by a future GEP.

For example, an implementation could define an `example.com/CEL` matcher type:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example
spec:
  parentRefs:
  - name: example
  rules:
  - matches:
    - extension:
        type: example.com/CEL
        value: 'request.path.contains("hi")'
    backendRefs:
    - name: example
      port: 80
```

## Conformance Details

#### Feature Names

None. This GEP defines an implementation-specific extension point, so there is
no portable feature for conformance reports to claim.

### Conformance test scenarios

None. This GEP defines an implementation-specific extension point and does not
standardize any concrete matcher type or matcher behavior.

Any implementation-specific match mechanism would be outside core portability
guarantees unless and until a specific part of it is standardized later.

No conformance test file names are proposed.

## `Standard` Graduation Criteria

To graduate this GEP to the Standard channel:

* A future GEP or revision of this GEP must define portable matcher behavior.
* Conformance tests must be implemented for that portable behavior.
* At least three (3) implementations must have submitted conformance reports that
  pass those future conformance tests.
* At least six months must have passed from when the GEP moved to
  `Experimental`.

## Alternatives

### Annotations

Easy to add, but opaque and inconsistent.

### Standardize CEL or WASM directly

Too specific. The problem is the extension point, not the choice of engine.

## References

- [GEP-820: Drop extension points from Route matches](../gep-820/index.md)
- [GEP-1364: Status and Conditions Update](../gep-1364/index.md)
