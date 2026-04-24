# GEP-3965: HTTPRoute Implementation-Specific Matches

* Issue: [#3965](https://github.com/kubernetes-sigs/gateway-api/issues/3965)
* Status: Provisional

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
`HTTPRoute`. The exact API shape is intentionally left open for now, but a
reference-based mechanism such as `extensionRef` is the likely direction.

This also revisits [GEP-820](../gep-820/index.md), which removed route match
extension points because there were no concrete use cases at the time. We now
have concrete use cases.

## Purpose (Why and Who)

This is for implementations that already support richer matching and for users
who need it without giving up on `HTTPRoute`.

It also gives the project a cleaner answer than annotations. Core matches stay
standard. Custom matches stay explicitly implementation-specific.

## API

The exact API is left for a follow-up revision of this GEP.

The likely shape is an implementation-specific reference on `HTTPRouteMatch`,
or another narrowly scoped mechanism with the same effect.

Whatever shape we choose should:

- live next to existing HTTP matches
- make implementation-specific behavior explicit
- compose with existing match semantics
- fail clearly when the referenced matcher is unsupported

## Conformance Details

No new conformance requirements are proposed in `Provisional`.

Any implementation-specific match mechanism would be outside core portability
guarantees unless and until a specific part of it is standardized later.

## Alternatives

### Annotations

Easy to add, but opaque and inconsistent.

### Standardize CEL or WASM directly

Too specific. The problem is the extension point, not the choice of engine.

## References

- [GEP-820: Drop extension points from Route matches](../gep-820/index.md)
- [GEP-1364: Status and Conditions Update](../gep-1364/index.md)
