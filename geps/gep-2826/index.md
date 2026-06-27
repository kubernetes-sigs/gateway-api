---
title: "GEP-2826: DirectResponse HTTPRouteFilter"
---

* Issue: [#2826](https://github.com/kubernetes-sigs/gateway-api/issues/2826)
* Status: Experimental

(See [status definitions](/geps/overview/#gep-states).)

## TLDR

Add a `DirectResponse` filter to `HTTPRouteFilter` that allows a Gateway to
respond directly to a matched request with a configured HTTP status code and
optional body, without forwarding the request to any backend.

## Goals

* Allow an Application Developer to configure a Gateway to return a fixed HTTP
  response for a matched route rule, without requiring a backend service.

## Non-Goals

* Supporting binary/base64-encoded response bodies (plain strings only).
* Setting arbitrary response headers (use `ResponseHeaderModifier` for that).
* Supporting `DirectResponse` at the `HTTPBackendRef` level in this iteration.

## Introduction

A common need in HTTP routing is the ability to respond directly to certain
requests — without proxying to any upstream — for example:

* Returning a `404` or `403` for a blocked or deprecated path.
* Serving a static `robots.txt` or health-check endpoint without a dedicated
  backend pod.
* Adding a catch-all fallback rule that returns a meaningful error instead of
  silently dropping requests.
* Responding to pre-flight or probe requests at the gateway layer.

Several implementations already support this:

* **Istio** — [`HTTPDirectResponse`](https://istio.io/latest/docs/reference/config/networking/virtual-service/#HTTPDirectResponse)
* **Contour** — [`HTTPDirectResponsePolicy`](https://projectcontour.io/docs/1.28/config/api/#projectcontour.io/v1.HTTPDirectResponsePolicy)
* **Envoy Gateway** — [`HTTPDirectResponseFilter`](https://gateway.envoyproxy.io/docs/api/extension_types/#httpdirectresponsefilter) (released in v1.2 with a 4096-byte body limit)

The convergence of behaviour across multiple implementations makes this a good
candidate for promotion to an extended conformance filter in Gateway API.

### Interaction with `backendRefs`

`DirectResponse` is a terminal filter: it generates the response itself and
never forwards the request. Therefore it MUST NOT be used in the same
`HTTPRouteRule` as `backendRefs`. This is enforced by a CEL validation rule,
consistent with how `RequestRedirect` is handled today.

### Body size limit

To keep Gateway API implementations portable and to avoid using an
`HTTPRouteFilter` as a content-delivery mechanism, the response body is capped
at **4096 bytes**. This limit matches the Envoy Gateway precedent and is
intentionally not configurable. Larger bodies should be served by a dedicated
backend.

## API

```golang
const (
    // HTTPRouteFilterDirectResponse can be used to respond directly to a
    // matched request without forwarding it to a backend.
    //
    // Support in HTTPRouteRule: Extended
    //
    // Feature Name: HTTPRouteDirectResponse
    //
    // <gateway:experimental>
    HTTPRouteFilterDirectResponse HTTPRouteFilterType = "DirectResponse"
)

// HTTPDirectResponseFilter defines a filter that replies to the request
// directly from the gateway, without forwarding to a backend.
//
// <gateway:experimental>
type HTTPDirectResponseFilter struct {
    // StatusCode is the HTTP status code to be used in the direct response.
    //
    // +kubebuilder:validation:Minimum=100
    // +kubebuilder:validation:Maximum=599
    // +required
    StatusCode int `json:"statusCode"`

    // Body is the HTTP response body to be returned.
    //
    // +optional
    Body *HTTPDirectResponseBody `json:"body,omitempty"`
}

// HTTPDirectResponseBody defines the body of an HTTP direct response.
//
// <gateway:experimental>
type HTTPDirectResponseBody struct {
    // String is the response body as a plain string.
    //
    // +kubebuilder:validation:MaxLength=4096
    // +required
    String string `json:"string"`
}
```

### Example: blocking a path

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: block-metrics
  namespace: default
spec:
  parentRefs:
  - name: http-gateway
  rules:
  - matches:
    - path:
        type: Exact
        value: /metrics
    filters:
    - type: DirectResponse
      directResponse:
        statusCode: 403
        body:
          string: "Forbidden"
  - backendRefs:
    - name: my-app
      port: 8080
```

### Example: serving robots.txt

```yaml
  rules:
  - matches:
    - path:
        type: Exact
        value: /robots.txt
    filters:
    - type: DirectResponse
      directResponse:
        statusCode: 200
        body:
          string: "User-agent: *\nDisallow: /"
```

### Example: catch-all 404

```yaml
  rules:
  - filters:
    - type: DirectResponse
      directResponse:
        statusCode: 404
        body:
          string: "Not Found"
```

### Example: JSON response with Content-Type header

`DirectResponse` MAY be combined with `ResponseHeaderModifier` to set response
headers such as `Content-Type`. This is the recommended pattern for returning
structured (e.g. JSON) bodies:

```yaml
  rules:
  - matches:
    - path:
        type: Exact
        value: /api/gone
    filters:
    - type: DirectResponse
      directResponse:
        statusCode: 410
        body:
          string: '{"error":"this endpoint has been removed"}'
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        set:
        - name: Content-Type
          value: application/json
```

## Conformance

A new extended conformance feature `HTTPRouteDirectResponse` is introduced.
Implementations that support this feature MUST:

1. Return the configured `statusCode` for all matched requests.
2. Return the configured `body.string` as the response body (UTF-8 plain text)
   when provided; omit the body when `body` is absent.
3. NOT forward the request to any backend when this filter is active.
4. Reject (via the `Accepted` condition set to `False`) any `HTTPRouteRule`
   that specifies both `DirectResponse` and non-empty `backendRefs`.
5. Support combining `DirectResponse` with `ResponseHeaderModifier` in the same
   rule to allow setting response headers (e.g. `Content-Type`).

## Alternatives Considered

### Use `ExtensionRef` custom filter

Implementations already support this via vendor-specific `ExtensionRef` filters.
However, this requires users to know the implementation-specific type names,
makes routes non-portable, and provides no standard conformance guarantee.

### Add a top-level `directResponse` field on `HTTPRouteRule`

Discussed in the issue thread. Placing it as an `HTTPRouteFilter` is more
consistent with how `RequestRedirect` works (also a terminal filter that skips
the backend) and keeps the rule structure uniform.

## References

* Issue: https://github.com/kubernetes-sigs/gateway-api/issues/2826
* Istio HTTPDirectResponse: https://istio.io/latest/docs/reference/config/networking/virtual-service/#HTTPDirectResponse
* Contour HTTPDirectResponsePolicy: https://projectcontour.io/docs/1.28/config/api/#projectcontour.io/v1.HTTPDirectResponsePolicy
* Envoy Gateway HTTPDirectResponseFilter: https://gateway.envoyproxy.io/docs/api/extension_types/#httpdirectresponsefilter
