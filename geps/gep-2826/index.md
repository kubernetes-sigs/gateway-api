---
title: "GEP-2826: DirectResponse HTTPRouteFilter"
---

* Issue: [#2826](https://github.com/kubernetes-sigs/gateway-api/issues/2826)
* Status: Provisional

(See [status definitions](/geps/overview/#gep-states).)

## TLDR

Add a `DirectResponse` filter to `HTTPRouteFilter` that allows a Gateway to
respond directly to a matched request with a configured HTTP status code and
optional body, without forwarding the request to any backend.

## Goals

* Allow an Application Developer to configure a Gateway to return a fixed HTTP
  response for a matched route rule, without requiring a dedicated backend
  service.

## Non-Goals

* Supporting binary/base64-encoded response bodies.
* Setting arbitrary response headers as part of this filter (use
  `ResponseHeaderModifier` for that).
* Supporting `DirectResponse` at the `HTTPBackendRef` level in this iteration.
* Serving large bodies — this filter is not a content-delivery mechanism.

## Introduction/Overview

Today, when a user wants to return a fixed HTTP response for a specific path
(e.g. `403 Forbidden`, a static `robots.txt`, or a catch-all `404`), they must
deploy a dedicated backend service whose only purpose is to return that
response. This is wasteful and adds operational overhead.

A `DirectResponse` filter would allow the Gateway itself to generate and return
the response, with no backend involved. The filter is *terminal*: when active,
the request is never forwarded upstream.

This is analogous to how `RequestRedirect` works today — it is also a terminal
filter that generates a response directly, and it is already prohibited from
being combined with `backendRefs`.

## Purpose (Why and Who)

**Target personas:**

* **Ana the Application Developer** — wants to block a sensitive path (e.g.
  `/metrics`, `/admin`) with a `403`, or serve a simple static response (e.g.
  `robots.txt`) without provisioning an extra Service and Deployment.

* **Chihiro the Cluster Admin** — wants to add a catch-all fallback rule to an
  `HTTPRoute` that returns a meaningful `404` instead of silently dropping
  unmatched requests.

**Problems being solved:**

* No standard way to return a fixed response from a Gateway without a backend.
* Users currently work around this with vendor-specific `ExtensionRef` filters
  (Envoy Gateway, Istio, Contour all have proprietary implementations), making
  routes non-portable.
* Simple use cases like blocking a path or serving `robots.txt` require a
  running Pod, which is operationally wasteful.

## API

[required_in]: # (Implementable status and above)

API details will be defined once this GEP is accepted as Provisional and moves
to Implementable. A prototype implementation is available in
[PR #5020](https://github.com/kubernetes-sigs/gateway-api/pull/5020) for
reference.

## Conformance Details

### Feature Names

* `HTTPRouteDirectResponse`

### Conformance test scenarios

#### HTTPRoute DirectResponse with body

An `HTTPRoute` with a `DirectResponse` filter configured with a status code and
body string should cause the Gateway to return that status code and body
directly, without forwarding to any backend.

#### HTTPRoute DirectResponse without body

An `HTTPRoute` with a `DirectResponse` filter configured with a status code but
no body should return that status code with an empty body.

#### HTTPRoute DirectResponse incompatible with backendRefs

An `HTTPRoute` rule that specifies both a `DirectResponse` filter and non-empty
`backendRefs` must be rejected by the Gateway (the `Accepted` condition set to
`False`).

#### Conformance test file names

* `httproute-direct-response.go`

## `Standard` Graduation Criteria

* At least one Feature Name must be listed. ✓ (`HTTPRouteDirectResponse`)
* The `Conformance Details` must be filled out with conformance test scenarios. ✓
* Conformance tests must be implemented that test all listed scenarios.
* At least three (3) implementations must have submitted conformance reports
  that pass those conformance tests.
* At least six months must have passed from when the GEP moved to `Experimental`.

## Alternatives

### Use `ExtensionRef` custom filter

Implementations already support direct responses via vendor-specific
`ExtensionRef` filters. However, this requires users to know
implementation-specific type names, makes routes non-portable, and provides no
standard conformance guarantee.

### Add a top-level `directResponse` field on `HTTPRouteRule`

Placing it as an `HTTPRouteFilter` is more consistent with how `RequestRedirect`
works (also a terminal filter that skips the backend) and keeps the rule
structure uniform.

## References

* Issue: https://github.com/kubernetes-sigs/gateway-api/issues/2826
* Istio HTTPDirectResponse: https://istio.io/latest/docs/reference/config/networking/virtual-service/#HTTPDirectResponse
* Contour HTTPDirectResponsePolicy: https://projectcontour.io/docs/1.28/config/api/#projectcontour.io/v1.HTTPDirectResponsePolicy
* Envoy Gateway HTTPDirectResponseFilter: https://gateway.envoyproxy.io/docs/api/extension_types/#httpdirectresponsefilter
