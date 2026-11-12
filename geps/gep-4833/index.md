---
title: "GEP-4833: ALPN Protocols for Backend Connections"
---

* Issue: [#4833](https://github.com/kubernetes-sigs/gateway-api/issues/4833)
* Status: Provisional

## TLDR

Define which ALPN protocols a Gateway offers when it opens a TLS connection to a backend, and let users set the list on both `Backend` and `BackendTLSPolicy`.

## Motivation

ALPN ([RFC 7301](https://www.rfc-editor.org/rfc/rfc7301)) is negotiated during the TLS handshake. The Gateway, as the TLS client, offers an ordered list of protocols. The backend makes the final choice, and many TLS servers choose by their own preference order rather than the client's. Some backends need a specific list:

* gRPC servers need `h2`.
* Some older HTTPS servers only work with `http/1.1`.
* Some backends, such as those inside a service mesh, need a non-HTTP or private ID.

For a `Backend`, the list can usually be inferred from `protocol`, and this GEP makes that mapping precise and required. Inference doesn't cover every case:

* A `Service` targeted by `BackendTLSPolicy` has no protocol field, and `appProtocol` has no value for HTTP/2 over TLS, since that's left to ALPN. There's nothing to infer `h2` from.
* No `protocol` value can cover every ID beyond `h2` and `http/1.1`.

Gateway API doesn't cover this today:

* [GEP-4894](/geps/gep-4894/) says the `Backend` `protocol` must be included in the ALPN offer, but leaves open what the full list is.
* `BackendTLSPolicy` covers CA trust, SNI and SAN validation for the backends it targets, but not ALPN.

Today the list can only be set through implementation-specific APIs, which differ between implementations and aren't portable. Configuring ALPN per backend should be standardized in Gateway API, so it works the same way on every implementation.

### User Stories

* As an application developer, I want to offer only `h2` to my gRPC backend, so the Gateway never falls back to HTTP/1.1 for it.
* As an application developer, I want to offer only `http/1.1` to a legacy backend, without changing ALPN for other backends on the same Gateway.
* As an application developer, I want different ALPN lists for different ports of the same backend.

## Goals

* Define how the `Backend` `protocol` maps to the ALPN list offered on TLS connections.
* Configure the ALPN list offered to a backend, inline on `Backend` and through `BackendTLSPolicy` per backend and per port, using the existing `targetRefs` and `sectionName`.
* Use the same ALPN semantics on both resources.
* Define how the list must line up with the protocol already implied for the backend, such as the `Backend` `protocol`, a `Service` port's `appProtocol` or the route type (for example, a `GRPCRoute` needs `h2`).
* Keep today's behavior when the list is unset.

## Non-Goals

* Other TLS options such as cipher suites, TLS versions or curves. These are tracked in [#4795](https://github.com/kubernetes-sigs/gateway-api/issues/4795).
* ALPN on Gateway listeners, where the Gateway is the TLS server.
* ALPN for HTTP/3 over QUIC.

## Overview

### Default list on Backend

When the list isn't set on a `Backend` with TLS enabled, implementations MUST derive it from `protocol`:

| `protocol` | ALPN list offered |
| --- | --- |
| `HTTP11`, `WSS` | `http/1.1` |
| `HTTP2`, `GRPC` | `h2` |
| `HTTP`, `MCP` | Implementation-specific; MUST include at least one of `h2` and `http/1.1` |
| `TCP` | None |
| `H2C` | Not applicable; `H2C` is cleartext |
| unset | Implementation-specific |

The effective backend protocol is `protocol` when set. Otherwise, it is the protocol implied by the route that references the Backend, such as HTTP/2 for a `GRPCRoute`, so it can differ between routes that share a Backend.

### Explicit list

When the list is set, implementations MUST offer exactly that list, in that order. To pin a backend to one HTTP version, offer only that version. On `Backend`, the list must be consistent with `protocol`:

* `HTTP2` and `GRPC`: the list MUST include `h2` and MUST NOT include `http/1.1`.
* `HTTP11`: the list MUST include `http/1.1` and MUST NOT include `h2`.
* `HTTP`, `MCP` and `WSS`: the list MUST include at least one of `h2` and `http/1.1`. Offering `h2` for `WSS` requires WebSockets over HTTP/2 ([RFC 8441](https://www.rfc-editor.org/rfc/rfc8441)) support.
* `TCP` or unset: any list is allowed.

### Protocol IDs

`h2` and `http/1.1` have Extended support. Any other protocol ID, such as a private ID used inside a service mesh, has Implementation-specific support. An implementation that can't offer a listed ID MUST report it in status rather than drop it silently. Protocol IDs MUST NOT contain commas, because some data planes store the list as a comma-separated string.

## API

**TODO**: This first PR leaves out API details, so we can agree on the motivation, goals and non-goals first. The likely direction is:

* An optional `alpnProtocols` list under `tls` on `Backend`, with CEL that enforces the rules above against `protocol`.
* An optional `alpnProtocols` list on `BackendTLSPolicySpec`, added to the Experimental channel first.
* A shared `ALPNProtocol` type in the `v1` API, which `Backend` reuses, since Standard types can't depend on experimental ones.

For example:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
metadata:
  name: web-api
spec:
  type: ExternalHostname
  externalHostname:
    hostname: api.example.com
  port:
    number: 443
  protocol: HTTP
  tls:
    mode: ServerOnly
    validation:
      hostname: api.example.com
      wellKnownCACertificates: System
    # Offer both versions and let the backend choose, instead of relying on
    # the implementation's default list for HTTP.
    alpnProtocols:
    - h2
    - http/1.1
---
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: grpc-backend-tls
spec:
  targetRefs:
  - group: ""
    kind: Service
    name: grpc-backend
    sectionName: https
  validation:
    hostname: grpc.example.com
    wellKnownCACertificates: System
  alpnProtocols:
  - h2
```

## Open Questions

* How should an implementation report an explicit list that conflicts with the route's effective protocol (for example, a `GRPCRoute` with no `h2` in the list), given that CEL can't see the route?

## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](/geps/gep-1897/)
* [GEP-4894: Backend Resource](/geps/gep-4894/)
* [#4795: GEP: Standardize TLS Options](https://github.com/kubernetes-sigs/gateway-api/issues/4795)
