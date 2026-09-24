---
title: "GEP-5293: Dynamic Request Header Modification"
---

* Issue: [#5293](https://github.com/kubernetes-sigs/gateway-api/issues/5293)
  * Related to [#5271](https://github.com/kubernetes-sigs/gateway-api/pull/5271) (Pre-Routing Filters), where the need for dynamic `RequestHeaderModifier` was surfaced
  * Related to [GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/pull/5092), which raised body-based header projection as a motivating case
* Status: Provisional

## TLDR

This GEP proposes that Gateway API extend `RequestHeaderModifier` so that
a header value can be **derived at request time** from something other
than a static string. Today `RequestHeaderModifier`'s `set` and `add`
accept only static string literals in `HTTPHeaderFilter.value`. There is
no portable way to say "set `x-model` from the JSON `model` field in the
request body," or "set `x-tenant` from the `tenant_id` claim on the
caller's JWT," or "set `x-client-cn` from the client certificate's Common
Name," even though data planes can express some form of this natively.

This GEP establishes that dynamic header modification is a first-class capability
that belongs in Gateway API, distinct from body processing, external
callouts, or pre-routing phase attachment. Value derivation is scoped to
running **in-process** in the data plane; out-of-process derivation (via
a `backendRef` to an external service such as `ext_proc`, `ext_authz`,
`ForwardAuth`, or SPOE) is out of scope. The concrete API shape (the
field(s) added to `HTTPHeaderFilter`, the expression language, and which
value sources are in scope) is deliberately left open here and enumerated
as *options under consideration*.

Because `HTTPHeaderFilter` is the type used by both
`HTTPRouteFilter.RequestHeaderModifier` and
`GRPCRouteFilter.RequestHeaderModifier`, any extension defined here
applies uniformly to HTTP and gRPC routes.

## Motivation

`RequestHeaderModifier` is one of the most widely implemented filters in
Gateway API and one of the most commonly requested extension points. Every
request today for "body-based routing," "route on a JWT claim," or "project
a client-cert field into a header before matching" terminates in the same
gap: `RequestHeaderModifier` can only set a header to a static value, so
the header cannot express any signal that is not already known at
configuration time.

### The Static-Value Limitation

`HTTPHeaderFilter.Set` and `Add` take an `HTTPHeader{Name, Value}` where
`Value` is a plain string:

> "Set overwrites the request with the given header (name, value) [...]"
> `HTTPHeaderFilter.Set`, [apis/v1/httproute_types.go](../../apis/v1/httproute_types.go)

There is no notion of a value source, expression, or template. The only
way to get a request-time-computed value into a header today is via
implementation-specific extension (for example, an implementation that
silently interprets `%REQ(...)%` inside `value`) or an out-of-band
component in front of the Gateway.

### Real Workloads Need Request-Time Header Values

Several established use cases need header values that cannot be known at
configuration time:

* **Body-derived routing keys.** LLM inference gateways route on the
  `model` field of a JSON request body; the Gateway API Inference
  Extension's Body-Based Router exists precisely to project that field
  into a header so standard `HTTPRoute` matching can act on it.
* **Verified identity projection.** A tenant identifier lives inside a JWT
  claim, or a caller identity in a client certificate SAN, and the
  backend expects that identifier in a header it can read directly. Every
  major data plane already offers a claim-to-header or cert-to-header
  primitive; Gateway API does not surface any of them.
* **Request-context enrichment.** The backend expects a header carrying
  the original client IP, TLS version, cipher suite, SNI, or a specific
  query parameter or cookie value. Managed cloud load balancers already
  expose a fixed set of predefined variables for exactly this purpose.
* **Normalization for downstream matching.** Canonicalizing or projecting
  a value from one place to another so that a later filter or a
  post-routing rule can act on it uniformly.

None of these can be expressed today with a static
`RequestHeaderModifier.value`.

### There Is No Portable Way to Express This

Because Gateway API does not model dynamic values, implementations
diverge. Envoy Gateway silently interprets `RequestHeaderModifier.value`
as an Envoy substitution string. agentgateway uses a separate
`AgentgatewayPolicy` with CEL. Istio exposes
JWT claim projection natively but requires `WasmPlugin` or `ext_proc` for
body access. NGINX Gateway Fabric *rejects* variable-like syntax in the
field and is tracking a separate design
([nginx/nginx-gateway-fabric#5737](https://github.com/nginx/nginx-gateway-fabric/issues/5737)).
Cloud load balancers each expose their own closed set of predefined
variables. A user who wants any of these behaviors currently picks an
implementation-specific mechanism and accepts lock-in, or deploys a
bespoke component in front of the Gateway.

### Relationship to Pre-Routing

Dynamic `RequestHeaderModifier` is a filter capability that could be
utilized in the pre-routing phase, as described in
[GEP-5224 (Pre-Routing Filters)](https://gateway-api.sigs.k8s.io/geps/gep-5224/).
Dynamic `RequestHeaderModifier` is a general capability that applies to
filters in any phase. When used in a post-routing filter, it can enrich
requests for later filters and backends, but it cannot influence route
selection. When used in a pre-routing filter, the derived header value
is available as an input to route matching.

Body-based routing is the motivating use case for both GEPs and requires
both capabilities together: the value must be derived from the request
body (dynamic `RequestHeaderModifier`), *and* it must be derived before
route selection (pre-routing). Neither GEP alone is sufficient for
body-based routing, but each is independently useful.

## Goals

* Establish agreement that Gateway API needs a portable way to express
  **request-time header values** in `RequestHeaderModifier`, rather than
  leaving each implementation to invent its own extension or requiring
  users to deploy an out-of-band component.
* Establish that this capability is a natural extension of the existing
  `RequestHeaderModifier` filter (via the shared `HTTPHeaderFilter` type),
  and applies uniformly to `HTTPRoute` and `GRPCRoute`.
* Enumerate the classes of value sources that motivate the feature. This
  includes body fields, JWT and other verified security context, and
  request context (headers, query parameters, cookies, connection info,
  client certificate fields). Do not prematurely commit to a specific set.
* Scope value derivation to **in-process** evaluation in the data plane.
  Out-of-process derivation via a `backendRef` to an external service
  (mirroring `ext_proc`, `ext_authz`, `ForwardAuth`, or SPOE) is
  explicitly out of scope for this GEP.
* Preserve full backwards compatibility: any existing configuration using
  static string values in `RequestHeaderModifier` continues to work
  unchanged.

## Non-Goals

* Defining the exhaustive set of supported value sources.
* Extending `ResponseHeaderModifier`. Response header dynamism has a
  different set of value sources (backend response headers, response
  body) and should be considered separately once the request-side model
  is settled.
* Specifying a body-processing surface (buffering size, streaming modes,
  content-type restrictions). This overlaps with
  [GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5092) and
  will be resolved in coordination with it.
* Standardizing pre-routing attachment. That is
  [GEP-5224](https://gateway-api.sigs.k8s.io/geps/gep-5224/)'s scope.

## User Stories

**As a Platform Engineer routing LLM inference traffic:**

> "My clients POST OpenAI-style JSON to a single endpoint. I want to route
> requests for `gpt-4` to one backend and `claude-3` to another, based on
> the `model` field in the request body, without deploying a bespoke
> body-based router in front of my Gateway."

**As a Platform Engineer for a Multi-Tenant Service:**

> "My callers present a JWT with a `tenant_id` claim. I want to project
> that claim into an `x-tenant` header so both my route matching and my
> backends can act on it consistently, without a per-implementation
> JWT-authn extension."

**As a Gateway API Implementation Author:**

> "My data plane already supports request-time header values through a
> substitution language or expression evaluation. I want a Gateway
> API surface I can implement conformantly, rather than silently
> extending `RequestHeaderModifier.value` or asking my users to leave
> Gateway API for an implementation-specific CRD."

## Static vs Dynamic Request Header Modification

Static and dynamic header values are different capabilities that a single
filter should be able to express. Recognizing dynamic values as a distinct
capability has three consequences:

1. **Its own conformance surface.** Not every implementation can support
   every value source. Managed cloud load balancers, for example,
   categorically cannot read the request body in their configuration
   surface. Naming the capability lets it be feature-flagged, tiered, and
   conformance-tested independently of the base `RequestHeaderModifier`.
2. **Its own cost model.** Dynamic values may require body buffering or
   expression evaluation. Users opting into a dynamic value are opting
   into that cost; the API should make the choice explicit.
3. **Composition with pre-routing.** Only a dynamic value can encode a
   signal that route matching needs to see. The pre-routing filter list
   ([GEP-5224](https://gateway-api.sigs.k8s.io/geps/gep-5224/)) is
   substantially more useful once dynamic values exist.

## Prior Art

The dynamic header modification pattern exists in many data planes;
what varies is *how* it is expressed and what value sources are reachable.

* **Envoy** exposes substitution command operators (`%REQ()%`,
  `%DYNAMIC_METADATA()%`, `%DOWNSTREAM_PEER_SUBJECT%`, `%CEL(...)%`, and
  others) in `request_headers_to_add` and the `header_mutation` filter.
  These run in-process and cover request context, TLS/mTLS, and dynamic
  metadata, but not the body. Body access is layered on via
  `json_to_metadata`, Lua, or Wasm in-process, or via `ext_proc`
  out-of-process. JWT claim projection is first-class through
  `jwt_authn.claim_to_headers`. **Envoy Gateway**
  already interprets `RequestHeaderModifier.value` as an Envoy substitution
  string, delivering dynamic values today off-spec.
* **agentgateway** evaluates CEL expressions like `json(request.body).model` in a
  `PreRouting` phase.
* **Istio** wraps `jwt_authn` as `RequestAuthentication.outputClaimToHeaders`. Body
  projection in Istio goes through `WasmPlugin` or `EnvoyFilter`.
* **HAProxy** is the closest existing prior art for *declarative* dynamic
  values without scripting. `http-request set-header X-Model
  %[req.body,json_query('$.model')]` composes a sample fetch (`req.body`,
  `http_auth_bearer`, `ssl_c_s_dn(cn)`, `src`, and others) with converters
  (`json_query`, `jwt_payload_query`, `regsub`, and others). Body access
  requires `option http-buffer-request` or `http-request wait-for-body`
  and is bounded by `tune.bufsize`.
* **NGINX** expands variables in `proxy_set_header` covering headers
  (`$http_<name>`), query (`$arg_<name>`), cookies, connection info, and
  TLS/mTLS fields; JWT claims are available in NGINX Plus. Body access
  requires `lua-nginx-module`, `njs`, or the announced-but-unshipped
  `client_body_preread` / `json_parse` directives referenced by
  [nginx/nginx-gateway-fabric#5737](https://github.com/nginx/nginx-gateway-fabric/issues/5737).
  `auth_request` + `auth_request_set` is the out-of-process pattern.
* **Kong**, **Apache**, and **Traefik** each expose comparable in-process
  substitution or expression languages (Kong `request-transformer-advanced`
  with `$(...)` Lua expressions and a `shared` scratchpad; Apache
  `mod_headers` with `ap_expr`; Traefik middlewares plus Yaegi plugins),
  and each has an out-of-process equivalent (Kong external plugin server,
  Apache handler chain, Traefik `ForwardAuth`).
* **Cloud load balancers** (GCP URL maps, Azure Application Gateway, Azure
  Front Door, AWS ALB) expose closed sets of predefined substitution
  variables covering client IP, TLS parameters, geo, and mTLS-derived
  fields. None of them offer request-body access in their configuration
  surface. Their out-of-process equivalents are vendor-specific (Lambda
  authorizers, ALB `authenticate-oidc`).

Across the surveyed implementations the *capability* is universal, the
common *value sources* overlap heavily on request context and verified
identity, and the *body-access* path is where the strongest divergence
sits. Every implementation supports an in-process pattern for at least
some value sources; managed cloud load balancers are the only class that
cannot support body access in-process. Out-of-process callout patterns
(such as `ext_proc`, `ext_authz`, `ForwardAuth`, and SPOE) exist in most
data planes as well but are out of scope for this GEP.

## Options Under Consideration

These design dimensions are enumerated so that agreement on motivation is
not blocked on any one of them. All are deferred to the Experimental stage.

* **In-process vs out-of-process.** *Resolved.* Value derivation is
  scoped to in-process evaluation in the data plane. Out-of-process
  derivation via a `backendRef` (mirroring `ext_proc`, `ext_authz`,
  `ForwardAuth`, or SPOE) is out of scope for this GEP. Users needing
  that pattern are directed to existing external-service GEPs
  ([GEP-1494 ExternalAuth](https://gateway-api.sigs.k8s.io/geps/gep-1494/))
  or implementation-specific mechanisms.
* **Enumerated value sources.** Which sources are in scope. Candidates
  include body fields, JWT and other verified security context,
  individual request headers, query parameters, cookies, TLS/mTLS fields,
  and connection metadata. Also open is whether they are exposed as a
  fixed enumeration (like GCP's or Azure's server variables) or through
  an open-ended expression language.
* **Expression language.** Whether to standardize on CEL (already used
  elsewhere in Kubernetes and in Envoy's `%CEL()%` formatter), on a bounded
  substitution vocabulary (like GCP's or the Envoy command operators), on
  JSONPath for body access, or on multiple.
* **API shape.** How the dynamic value is expressed relative to the
  existing static `value` field. Options include a sibling `valueFrom`
  field on `HTTPHeader` (as sketched in the [PR #5092 discussion](https://github.com/kubernetes-sigs/gateway-api/pull/5092#discussion_r3852055666)),
  a new filter type, or a per-source structured field.
* **Body-access surface.** If body-based projection is in scope, whether
  body buffering size, streaming behavior, content-type constraints, and
  failure modes are specified here or by reference to
  [GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5092).
* **Conformance tiering.** Whether dynamic values are one Extended feature
  or several (for example one Extended feature per value source), so that
  a managed cloud load balancer can conform to the request-context tier
  without being required to support body access.
* **Interaction with pre-routing.** Whether the dynamic
  `RequestHeaderModifier` is admitted into the
  [GEP-5224](https://gateway-api.sigs.k8s.io/geps/gep-5224/) pre-routing
  filter list from the start, and if so under what conformance conditions.

## Open Questions

* **How is failure handled?** If a JWT is absent, a body is not JSON, or
  an expression does not evaluate, does the header stay unset, take a
  default, or does the request fail? Does the answer differ by value
  source?
* **How does this interact with static `set` on the same header name?**
  If the same header is set statically in one filter and dynamically in
  another, or in the same filter, what is the precedence?
* **How is this presented to the user in status?** When an implementation
  cannot honor a particular value source, does it reject the filter at
  admission time, degrade at runtime, or surface a route/listener
  condition?
* **How does this relate to implementation-specific matches
  ([GEP-3965](https://gateway-api.sigs.k8s.io/geps/gep-3965/))?** Some
  data planes can match directly on a body field or claim without a
  project-into-header intermediate step. The two mechanisms should not
  contradict each other.

## Relationship to Other Work

* **[GEP-5224: Pre-Routing Filters](https://gateway-api.sigs.k8s.io/geps/gep-5224/)**.
  Pre-routing is the phase where a dynamic `RequestHeaderModifier` can
  influence route selection. The two GEPs are separable but each is more
  useful with the other.
* **[GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/pull/5092)**.
  Raised body-based header projection as a motivating case. Body-access
  semantics defined there (buffering size, streaming, content type) will
  need to be reused if body sources are in scope here.
* **[Issue #5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194)**.
  Once a later filter can consume a value produced by an earlier dynamic
  `RequestHeaderModifier`, ordering ceases to be commutative. This GEP
  inherits the ordering discussion.
* **[GEP-1494: HTTP Auth / ExternalAuth](https://gateway-api.sigs.k8s.io/geps/gep-1494/)**.
  `ExternalAuth` already injects headers as a side effect of
  authorization. Ordering and precedence with a dynamic
  `RequestHeaderModifier` must be reconciled.
* **[GEP-3965: HTTPRoute Implementation-Specific Matches](https://gateway-api.sigs.k8s.io/geps/gep-3965/)**.
  Some data planes can match on a derived signal directly. The two
  mechanisms should compose cleanly.
* **Gateway API Inference Extension Body-Based Router.** The strongest
  existing in-tree example of the body-derived header projection pattern,
  implemented as an out-of-process `ext_proc` component today because the
  core API has no portable way to express it.

## References

* [Issue #5293: Extend `RequestHeaderModifier` for dynamic modification](https://github.com/kubernetes-sigs/gateway-api/issues/5293)
* [PR #5271: GEP-5224 Pre-Routing Filters (discussion of dynamic `RequestHeaderModifier`)](https://github.com/kubernetes-sigs/gateway-api/pull/5271)
* [PR #5092: GEP-5091 PayloadProcessor Resource (sketch of `valueFrom` on `HTTPHeader`)](https://github.com/kubernetes-sigs/gateway-api/pull/5092#discussion_r3852055666)
* [GEP-5224: Pre-Routing Filters](https://gateway-api.sigs.k8s.io/geps/gep-5224/)
* [Issue #5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* [nginx/nginx-gateway-fabric#5737: Predicate Routing](https://github.com/nginx/nginx-gateway-fabric/issues/5737)
* [Gateway API Inference Extension Body-Based Router](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
* [Envoy: `jwt_authn` `claim_to_headers`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/jwt_authn_filter)
* [Envoy: `json_to_metadata` filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/json_to_metadata_filter)
* [Envoy: `ext_proc` filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter)
* [Envoy Gateway: HTTP request headers](https://gateway.envoyproxy.io/docs/tasks/traffic/http-request-headers/)
* [Istio: `RequestAuthentication.outputClaimToHeaders`](https://istio.io/latest/docs/tasks/security/authentication/claim-to-header/)
* [HAProxy: configuration reference (sample fetches and converters)](https://docs.haproxy.org/3.0/configuration.html)
* [NGINX: `proxy_set_header`](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header)
* [Kong: `request-transformer-advanced`](https://developer.konghq.com/plugins/request-transformer-advanced/)
* [GCP: custom headers on external Application Load Balancers](https://cloud.google.com/load-balancing/docs/https/custom-headers-global)
* [Azure Application Gateway: rewrite HTTP headers and URL](https://learn.microsoft.com/en-us/azure/application-gateway/rewrite-http-headers-url)
