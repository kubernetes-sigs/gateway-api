---
title: "GEP-5224: Pre-Routing Filters"
---

* Issue: [#5224](https://github.com/kubernetes-sigs/gateway-api/issues/5224)
  * Split out of [GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/pull/5092)
  * Resolves the pre-routing portion of [#5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* Status: Implementable

## TLDR

This GEP proposes that Gateway API define a way to express **pre-routing
filters**: processing steps that run **before** a route is selected and that
may therefore influence which route, and ultimately which backend, is chosen.
Every filter Gateway API defines today, from `RequestHeaderModifier` to
`ExternalAuth`, is a **post-routing** filter: it is attached to an
`HTTPRouteRule` and only runs *after* a route has already been matched. There
is no portable mechanism for a filter to act on a request before route
selection.

The Provisional stage was scoped to the *what*, *who*, and *why*. It
established that pre-routing processing is a distinct and necessary phase, and
that the distinction between pre-routing and post-routing is a first-class
architectural property rather than an implementation detail.

This GEP defines the concrete API shape (see [API](#api)). Pre-routing
filters are attached to a Gateway `Listener` via a new field,
`httpFilters []HTTPListenerFilter`, whose element type reuses **only** a
narrow subset of the existing `HTTPRouteFilter` payload variants — those
that can produce a *dynamic* signal capable of influencing route selection.
Today that subset is `ExternalAuth` (to establish identity before matching)
and `ExtensionRef` (the escape hatch for custom pre-routing behavior such
as body-based routing or JWT-claim projection). The remaining
`HTTPRouteFilter` variants are excluded — either because they don't mutate
the request (`RequestMirror`, `ResponseHeaderModifier`, `CORS`), because
their static mutation cannot usefully influence matching
(`RequestHeaderModifier`, `URLRewrite`), or because they short-circuit
routing entirely (`RequestRedirect`), which is already achievable with a
catch-all HTTPRoute. Ordering within the list is a strict guarantee (MUST),
route matching runs single-pass after all pre-routing filters complete, and
Listener selection cannot be affected by pre-routing filter mutations —
enforced structurally by TLS for HTTPS listeners and contractually by this
GEP for cleartext HTTP listeners. The initial API surface covers
HTTP-family listener protocols and `HTTPRoute`; extension to `GRPCRoute` and
to L4 route kinds is captured as a deliberate design consideration, not
deferred by accident.

This GEP is one of three that the payload-processing proposal
([GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5091) was split into during Gateway API community discussion:

* **Pre-routing filters** (this GEP): a way to configure filters that run
  before route selection.
* **Body-based routing and/or modification**: acting on the request/response
  body, largely covered by the existing [GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5091) proposal.
* **External callouts**: invoking an external processing service via a defined
  wire protocol.

The three are related but separable. Pre-routing filters are the *phase and
attachment* mechanism; body-based routing and external callouts are *kinds of
processing* that frequently, but not always, need to run in that phase.

## Motivation

Gateway API provides a powerful, extensible filter model for HTTP. However,
that model has a structural gap: it can only express processing that happens
**after** a route has been selected. A growing class of use cases, from AI
inference routing to identity-derived routing, needs to act on a request
*before* the routing decision is made. Today there is no portable way to
express that.

### Every Filter Today Is a Post-Routing Filter

Gateway API filters are attached to an `HTTPRouteRule` (or, more narrowly, to
an `HTTPBackendRef`). By construction, a rule's filters run only once that
rule has matched the request:

> "Filters define the filters that are applied to requests that match this
> rule."
> — [`HTTPRouteRule.Filters` reference](../../reference/api-types/httproute.md#filters-optional)

This is true of all eight filter types: `RequestHeaderModifier`,
`ResponseHeaderModifier`, `RequestRedirect`, `URLRewrite`, `RequestMirror`,
`CORS`, `ExternalAuth`, and `ExtensionRef`. Each one operates on a request
whose destination has already been decided. None of them can change the route.
There is no listener-level or Gateway-level filter attachment point in the
current API, and therefore no supported place for a filter to run before
`HTTPRoute` matching.

The consequence is that Gateway API can express "once you know where this
request is going, transform it" but cannot express "before you decide where
this request goes, do X." The capabilities of the latter are what pre-routing filters
are about.

### Real Workloads Need to Act Before the Route Is Chosen

Several established and emerging patterns require processing *before* route
selection:

* **Routing on a derived signal.** A request's routing-relevant attributes are
  frequently not present in the raw request. The `model` field lives in a JSON
  body; a tenant identifier lives inside a signed JWT; a client's
  organization lives in a presented certificate. To route on any of these, an
  implementation must first extract the signal and project it somewhere the
  router can match on, *before* matching runs.
* **Normalizing the inputs that matching depends on.** Canonicalizing a path,
  host, or header so that `HTTPRoute` matching behaves predictably is only
  meaningful if it happens before matching.
* **Establishing verified identity that routing can trust.** When routing
  depends on *who* the caller is, the identity has to be established before the
  route is selected, not after.

None of these can be expressed natively in Gateway API with a post-routing filter, because in each case
the very information the route depends on is produced by the processing step.
[GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5091) already elevates this to a goal for the specific case of
payload processing; this GEP generalizes the underlying phase so that it is not
tied to any single kind of processing.

### There Is No Portable Way to Express This Today

Because Gateway API has no pre-routing attachment point, implementations that
support pre-routing behavior expose it through implementation-specific
extensions, or users work around it by deploying an out-of-band component in
front of the Gateway. The most prominent in-tree example is the Gateway API
Inference Extension's Body-Based Router, which extracts a model name from the
request body and surfaces it as a header so that a downstream `HTTPRoute` can
match on it. This is a pre-routing pattern implemented as a bespoke component
because the core API has nowhere to put it. The result is that a foundational
capability is neither portable nor discoverable.

### Ordering Matters Most in the Pre-Routing Phase

Filter ordering in Gateway API is currently a `SHOULD`, not a `MUST`:

> "Wherever possible, implementations SHOULD implement filters in the order
> they are specified. Implementations MAY choose to implement this ordering
> strictly [...]"
> — [`HTTPRouteRule.Filters` reference](../../reference/api-types/httproute.md#filters-optional)

[Issue #5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194) revisits whether that ordering should become a
guarantee. As noted there, when filters were "largely commutative header and
path manipulations," soft ordering was tolerable; now that filters can
authorize, terminate, or mutate a request, order changes the result. Pre-routing
processing is exactly where ordering is load-bearing: "extract a value from a
body and push it into a header, order of those operations will probably
matter." The community's working resolution on [Issue #5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194) is to introduce the
pre-routing phase (this GEP) and to make ordering strict for pre-routing
filters from the moment they are introduced, rather than retrofitting a
guarantee onto the existing GA post-routing filter list. Establishing a clean
pre-routing phase with strict ordering from day one avoids the backwards
compatibility problem that a global `SHOULD`->`MUST` change would create for
existing post-routing filters.

## Goals

* Establish agreement that Gateway API needs a first-class way to express
  **pre-routing filters**, processing that runs before route selection and may
  influence it, and that this belongs in the Gateway API family of APIs rather
  than only in implementation-specific extensions or out-of-band components.
* Establish **pre-routing** and **post-routing** as distinct, named phases,
  and clarify that today's `HTTPRouteRule` filters are, by definition,
  post-routing.
* Articulate *why* the distinction is beneficial and load-bearing: only
  pre-routing processing can affect routing, the two phases have different
  scoping rules, and ordering has different weight in each.
* Cover the classes of use cases that require a pre-route decision, including
  but not limited to body-based routing, identity-derived routing, and
  matching-input normalization.
* Define a concrete API — attachment point, element type,
  ordering guarantee, and route-matching interaction — for HTTP-family
  listeners and `HTTPRoute`. Specifically, add
  `Listener.httpFilters []HTTPListenerFilter` with strict list
  ordering. See [API](#api).
* Establish that ordering among pre-routing filters is significant and is a
  guarantee (MUST, not SHOULD) from the introduction of the phase, providing
  a concrete answer to the pre-routing portion of
  [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194).

## Non-Goals

* **Defining an L4 pre-routing surface.** SNI/TLS-based selection
  ([GEP-2643](../gep-2643/index.md)) and TCP/UDP-layer pre-routing filters
  are lower-layer forms of pre-routing and are not in scope here. The API
  is deliberately structured so a parallel L4 field (see
  [Extensibility to Other Route Kinds](#extensibility-to-other-route-kinds))
  can be added by a follow-up GEP without renaming or moving
  `httpFilters`.
* **Defining pre-routing for `GRPCRoute`.** GRPCRoute is HTTP/2 in practice,
  and the same `httpFilters` field is expected to serve it, but
  formally extending the phase (and conformance) to GRPCRoute is a follow-up.
* **Introducing new filter payload types.** This GEP reuses a subset of the
  existing `HTTPRouteFilter` payloads (`ExternalAuth`, `ExtensionRef`)
  inside a new outer `HTTPListenerFilter` wrapper. The other variants are
  explicitly excluded — see
  [Excluded Filter Variants](#excluded-filter-variants) for per-variant
  rationale. Adding pre-routing-only filter variants (for example, a
  body-projection filter once GEP-5091 lands) is deferred.
* **Body-level processing semantics.** How the body is addressed, buffered,
  or mutated is the subject of the body-based routing / [GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5091) work.
  This GEP defines the *phase*; the body-processing content of that phase is
  defined elsewhere.
* **External processing wire protocols.** Invoking an external service is the
  subject of the external-callouts split. Pre-routing filters may host such
  callouts (via `ExternalAuth` or `ExtensionRef`), but the protocol is out of
  scope here.
* **Re-litigating post-routing filter ordering.** [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194) asks a
  broader question about the existing GA filter list. This GEP commits to
  strict ordering *only within the new `HTTPFilters` list*; the
  post-routing question is left to that issue.
* **Replacing existing filters or route matching.** Pre-routing filters
  complement `HTTPRoute` matching and existing filters; they do not replace
  header/path/method matching or post-routing transformation.

## User Stories

The following stories describe *what* users want to accomplish. They are
intentionally described independently of the eventual API shape and of whether
the underlying processing runs in-data-plane or via an external service.

### Body-Based Routing

**As an AI Platform Engineer:**

> "I want the gateway to read the `model` field from a JSON inference request,
> use it to choose the correct model backend, and do so with a portable Gateway
> API construct instead of a bespoke body-based router deployed in front of my
> Gateway. For that to work, the extraction has to happen before the route is
> selected."

### Identity-Derived Routing

**As a Platform Engineer for a Multi-Tenant Service:**

> "I want to route requests to per-tenant backends based on a tenant claim
> inside the caller's JWT. The claim has to be extracted and made available to
> route matching before the route is chosen, otherwise I cannot express
> 'tenant A goes here, tenant B goes there.'"

### Normalizing Matching Inputs

**As an API Owner:**

> "I want to canonicalize inbound paths and hostnames before route matching so
> that my `HTTPRoute` rules behave predictably regardless of how clients format
> their requests. A transformation that runs after the route is already chosen
> is useless for this."

### Ordered Pre-Routing Pipelines

**As a Gateway User:**

> "I want to run several pre-routing steps in a defined order, for example
> 'validate the request shape, then extract a routing key, then normalize a
> header,' and I need to rely on that order being honored, because each step
> depends on the output of the previous one."

### Implementation Consistency

**As a Gateway API Implementation Author:**

> "I want an unambiguous definition of what 'pre-routing' means: when these
> filters run relative to route matching, how they are ordered, whether and how
> routing is re-evaluated after they run, and how they interact with existing
> post-routing filters and with `ExternalAuth`. My data plane already has a
> route-selection phase; I need the spec to map cleanly onto it."

## Pre-Routing vs. Post-Routing: The Core Distinction

The central claim of this GEP is that "before route selection" and "after route
selection" are fundamentally different phases, and that Gateway API today only
exposes the second one.

| Phase | When it runs | What it can do | Examples |
|-------|--------------|----------------|----------|
| **Pre-Routing** (new) | Before `HTTPRoute` matching | Mutations **can** influence which route and backend are selected. Cannot be scoped by matched-route information, because no route has been matched yet. | Extract a routing key from the body or a JWT and project it into a header; normalize path/host; establish identity used for routing. |
| **Post-Routing** (all filters today) | After the route is selected, before backend dispatch (and, symmetrically, on the response) | Mutations **cannot** re-shape routing. They act only on the request delivered to the chosen backend or the response returned to the client. | `RequestHeaderModifier`, `URLRewrite`, `RequestRedirect`, `RequestMirror`, `CORS`, `ResponseHeaderModifier`, `ExternalAuth`. |

```
Client Request
    │
    ▼
┌──────────────────────┐
│  Pre-Routing Filters │ ◄── NEW. Runs before matching, CAN influence routing.
│  (ordered)           │     (extract, project, normalize, validate)
└──────────┬───────────┘
           │ (headers / metadata potentially mutated)
           ▼
┌──────────────────────┐
│  HTTPRoute Matching  │ ◄── Standard header/path/method matching
└──────────┬───────────┘
           │ (route selected)
           ▼
┌──────────────────────┐
│  Post-Routing Filters│ ◄── EXISTING HTTPRoute filters. CANNOT influence routing.
│  (HTTPRouteRule)     │     (RequestHeaderModifier, URLRewrite, CORS, ExternalAuth, …)
└──────────┬───────────┘
           │
           ▼
       Backend
```

Naming the phases has three concrete consequences:

1. **Only pre-routing processing can affect route selection.** This is the defining
   difference. A post-routing filter, no matter what it does, runs against a
   request whose destination is fixed. If the goal is to *change where a
   request goes* based on some computed signal, it can only be done
   pre-routing. Conflating the two phases hides this causal asymmetry.
2. **The phases have different scoping.** A post-routing filter is naturally
   scoped to the matched `HTTPRouteRule`. A pre-routing filter cannot be, since
   no route has been matched when it runs; it is naturally scoped to a broader
   context such as a `Gateway` or listener. This is why pre-routing cannot
   simply be "another entry in the `HTTPRouteRule` filter list."
3. **Ordering has different weight.** Post-routing filters are frequently
   commutative (the order of two header edits rarely matters). Pre-routing
   filters are frequently *not* commutative, because later steps consume the
   output of earlier ones. This is the basis for treating pre-routing ordering
   as a guarantee from the start, as discussed in [#5194][issue-5194].

## Prior Art

Reviewing how existing data planes model this is instructive, because the
pre-routing vs. post-routing split is not a novel concept for gateway implementations.

### Envoy

Envoy's HTTP Connection Manager resolves and caches a route when it finishes
decoding the request headers. Route matching can consider the request path,
method, authority, headers, cookies, and filter-populated dynamic metadata or
filter state, but not the request body. Downstream decoder filters run before
the terminal `router` filter and may influence the route ultimately used by
mutating these inputs and clearing or replacing the cached route. A filter that
mutates a routing-relevant input can force the route to be recomputed via
`clearRouteCache()`; several filters (Lua, ext_proc, the Golang filter,
`json_to_metadata`) can do exactly this. Conversely, `typed_per_filter_config`
attaches configuration to an already-selected route or virtual host, and is by
definition post-routing. Envoy even documents the security implications of the
boundary: route-dependent authorization filters must be ordered carefully
relative to filters that mutate route-matching inputs and clear the route
cache. In Envoy terms, "pre-routing" is "a filter that runs before the router
filter and may clear the route cache," and "post-routing" is
"`typed_per_filter_config` on the resolved route."

### NGINX

NGINX models request handling as an explicit, ordered sequence of **phases**.
Location (route) selection happens in a dedicated phase,
`NGX_HTTP_FIND_CONFIG_PHASE`. Phases such as `NGX_HTTP_POST_READ_PHASE` and
`NGX_HTTP_SERVER_REWRITE_PHASE` run *before* it and are therefore pre-routing;
`NGX_HTTP_REWRITE_PHASE` runs *after* location selection and can rewrite the
URI, which sends the request back through `FIND_CONFIG` for re-selection. NGINX
thus has both an explicit pre-routing hook (server rewrite) and an explicit
mechanism for a post-selection mutation to trigger re-routing.

### HAProxy

HAProxy splits processing between the `frontend` and the `backend`. `http-request`
rules declared in a `frontend` execute before `use_backend` selects the
destination; `http-request` rules declared in a `backend` execute after
selection. The frontend/backend boundary *is* the routing boundary, so HAProxy
users already reason in terms of "before backend selection" (pre-routing) and
"after backend selection" (post-routing).

### Istio

Istio builds on Envoy (via sidecars and waypoints) and inherits its filter-chain-before-router model, so the
same "filters before the terminal `router` filter can influence routing"
property applies. On top of that, Istio exposes insertion semantics for
extensions: `EnvoyFilter` lets operators splice a custom HTTP filter into the
connection manager's filter chain at a chosen position, including *before* the
terminal `router` filter, which is the position that makes a filter effectively
pre-routing. `WasmPlugin` additionally offers a coarser `phase` field
(`AUTHN`, `AUTHZ`, `STATS`) that orders a plugin relative to Istio's
authentication, authorization, and telemetry filters rather than relative to
route selection directly; because those phases all sit within the pre-router
filter chain, a plugin placed there still runs before routing, but the field is
not a general-purpose "before/after route selection" control. The `EnvoyFilter`
insertion point is therefore the more precise example of an operator explicitly
choosing where, relative to routing, custom processing runs.

### agentgateway

agentgateway is an AI-native, Gateway API-based data plane aimed at LLM, MCP,
and A2A traffic, and is the closest existing analogue to this GEP because it
already exposes a **named pre-routing phase in Gateway API terms**. Every
request flows through four fixed, ordered phases — Frontend, PreRouting,
PostRouting, Backend — with route selection between PreRouting and PostRouting;
PreRouting filters run before a route is chosen and can influence selection,
while PostRouting filters cannot. A PreRouting policy can only target a
`Gateway` or `ListenerSet` (never an `HTTPRoute`, since no route has matched),
and the phase admits only a fixed, ordered subset of filters (JWT, basic, and
API-key auth, external authorization, external processing, and transformation).

### Apache HTTP Server

Apache's request lifecycle is likewise a sequence of hooks, with
`translate_name` and `map_to_storage` resolving the request to a handler
(the routing-equivalent step) after earlier hooks have run. Modules routinely
act before this resolution to influence it, another instance of the pre-routing
pattern.

### Cloud Load Balancers

Managed L7 load balancers (for example AWS Application Load Balancer listener
rules, or GCP URL maps) evaluate conditions to select a target group or backend
service, which is the routing step. Their capacity for *mutation before*
routing is generally more limited than a full proxy, and some expose only
fixed, pre-defined pre-routing behaviors (for example authenticate-then-route
actions). This is relevant to conformance: a portable pre-routing filter API
will need to account for data planes whose pre-routing extensibility is
constrained, likely via conformance levels or the ability to reject
configurations that cannot be honored.

### Takeaway

Across Envoy, NGINX, HAProxy, Istio, agentgateway, Apache, and cloud load
balancers, the same shape recurs: request processing is a sequence of ordered
phases with a distinct route/backend-selection step, and hooks exist both
*before* and *after* it.
Gateway API's filter model currently exposes only the post-routing hook. A
pre-routing filter API would map onto capabilities these data planes already
have, rather than asking them to build something new.

## API

This section defines the concrete API shape proposed.
It resolves the attachment point, the filter type, ordering, and
interaction with route matching. Alternatives that were considered but not
adopted are recorded in [Alternatives Considered](#alternatives-considered).

### At a Glance

* HTTP pre-routing filters attach to a Gateway `Listener` via a new
  `httpFilters` field.
* The list is a new type, `HTTPListenerFilter`, whose element variants
  reuse a subset of the `HTTPRouteFilter` payloads: `ExternalAuth` and
  `ExtensionRef`. The remaining variants are excluded — see
  [Excluded Filter Variants](#excluded-filter-variants).
* Filters MUST execute in the exact order they appear in the list. Ordering is
  a *guarantee*, specified as a `MUST`, from the moment the field is
  introduced.
* HTTPRoute matching runs exactly once after all pre-routing filters have run.
* Pre-routing filters cannot re-select the Listener. For HTTPS listeners this
  is enforced structurally by the TLS handshake (SNI binds the listener to a
  specific certificate before HTTP bytes are exchanged); for cleartext HTTP
  listeners it is enforced contractually — Listener selection uses the
  request's initial `Host` header and filter mutations do not feed back into
  that selection. See [Listener Match Invariance](#listener-match-invariance).
* The initial API surface only applies to HTTP-family listener protocols
  (`HTTP`, `HTTPS`) and to `HTTPRoute`. The type is designed so that support
  for `GRPCRoute` and for L4 route kinds can be added later without renaming
  or restructuring the field.

### Attachment: `Listener.httpFilters`

Pre-routing filters attach to an individual `Listener` in the Gateway spec.
The Listener is the smallest existing Gateway API surface that already
carries a specific protocol and hostname context, both of which are
determined without reference to filter state — HTTPS by the TLS handshake
before HTTP bytes are exchanged, cleartext HTTP by matching the request's
initial `Host` header against `Listener.hostname` (see
[Listener Match Invariance](#listener-match-invariance)). This makes the
Listener a natural attachment point.

```go
type Listener struct {
    // ...existing fields (Name, Hostname, Port, Protocol, TLS, AllowedRoutes)...

    // HTTPFilters is an ordered list of HTTP-layer filters that run
    // on every request accepted on this Listener, before HTTPRoute matching
    // is performed. The list order is load-bearing: implementations MUST
    // execute the filters in the exact order they appear here and MUST NOT
    // reorder them. Each filter's Name MUST be unique within the list.
    //
    // HTTPFilters may mutate inputs that HTTPRoute matching
    // consumes (path, request headers including :authority/Host, method,
    // computed metadata). Implementations MUST evaluate HTTPRoute matching
    // exactly once after the last HTTPListenerFilter has run.
    //
    // HTTPFilters MUST NOT be interpreted as changing which
    // Listener handles the request. Listener selection is decided from
    // inputs the client committed to before any pre-routing filter runs:
    // TLS SNI (for HTTPS) or the request's initial Host header (for HTTP).
    // Filter mutations of Host, :authority, or any other header MUST NOT
    // be re-fed into Listener selection.
    //
    // HTTPFilters MUST be empty when Protocol is not `HTTP` or
    // `HTTPS`. Implementations MUST reject a Listener with
    // Accepted=False / Reason=InvalidHTTPFilter if this constraint is
    // violated.
    //
    // Support: Extended
    //
    // +optional
    // +listType=atomic
    // +kubebuilder:validation:MaxItems=16
    HTTPFilters []HTTPListenerFilter `json:"httpFilters,omitempty"`
}
```

`httpFilters` is only accepted on listeners whose `Protocol` is `HTTP` or
`HTTPS`, because every element type has HTTP semantics (header mutation,
URL rewrite, HTTP-shaped external auth). The HTTP prefix on the field name
signals both the payload shape and this protocol restriction. Follow-up
GEPs may introduce parallel fields for other protocols (`tlsFilters`,
`tcpFilters`, `udpFilters`) with protocol-appropriate element types
(`TLSListenerFilter`, `TCPListenerFilter`, `UDPListenerFilter`); the naming
convention keeps each phase attached to a phase-appropriate payload rather
than forcing a single polymorphic list to shape-shift.

### The `HTTPListenerFilter` Type

`HTTPListenerFilter` is a **new type**, distinct from `HTTPRouteFilter`. It
is deliberately narrower than `HTTPRouteFilter`: only a subset of the
filters are included, and were chosen because they can produce a *dynamic*
signal capable of influencing route selection. The remaining
`HTTPRouteFilter` variants are excluded — see
[Excluded Filter Variants](#excluded-filter-variants) for the per-variant
rationale. The outer type is separate so that the strict-ordering guarantee
lives on the pre-routing container and does not perturb the
`SHOULD`-ordering semantics of
[`HTTPRouteRule.Filters`](../../reference/api-types/httproute.md#filters-optional).

```go
// HTTPListenerFilter is one element of a Listener.HTTPFilters list.
// Unlike HTTPRouteFilter, which runs after a route is selected and whose
// ordering is a SHOULD, HTTPListenerFilter runs before route selection and
// its ordering within the containing list is a MUST.
//
// Only a subset of HTTPRouteFilter variants are permitted:
// ExternalAuth (to establish identity that route matching can consume) and
// ExtensionRef (the escape hatch for custom pre-routing behavior — for
// example body-based routing or JWT-claim projection). The other
// HTTPRouteFilter variants are excluded because they either cannot influence
// route selection or can already be expressed post-routing with equal
// expressiveness. See the GEP text for details.
type HTTPListenerFilter struct {
    // Name is an optional, list-unique identifier for this filter. It is optionally
    // included in Gateway status conditions and in implementation
    // diagnostics (for example, to report which filter rejected a request).
    //
    // +optional
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:MaxLength=253
    Name SectionName `json:"name,omitempty"`

    // Type identifies which variant of the discriminated union below is
    // populated. Uses the same union-discriminator pattern as HTTPRouteFilter.
    //
    // +unionDiscriminator
    // +kubebuilder:validation:Enum=ExternalAuth;ExtensionRef
    // +required
    Type HTTPListenerFilterType `json:"type"`

    // The following fields reuse the corresponding HTTPRouteFilter payload
    // types verbatim. Exactly one MUST be set, and it MUST correspond to
    // Type. CEL validation enforces this, matching the existing
    // HTTPRouteFilter pattern.

    ExternalAuth *HTTPExternalAuthFilter `json:"externalAuth,omitempty"`
    ExtensionRef *LocalObjectReference   `json:"extensionRef,omitempty"`
}

// HTTPListenerFilterType is a distinct enum from HTTPRouteFilterType so
// that the two lists can diverge. Today pre-routing admits only a subset
// of the HTTPRouteFilter variants; future revisions may add pre-routing-only
// variants (for example, a first-class body-projection filter once GEP-5091
// lands) without disturbing HTTPRouteFilter.
type HTTPListenerFilterType string

const (
    HTTPListenerFilterExternalAuth HTTPListenerFilterType = "ExternalAuth"
    HTTPListenerFilterExtensionRef HTTPListenerFilterType = "ExtensionRef"
)
```

The paired-CEL validation rule (already established by `HTTPRouteFilter`)
carries over: the field named by `Type` MUST be set, and every other variant
field MUST be nil.

Example YAML:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: example
spec:
  gatewayClassName: example-class
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    hostname: api.example.com
    tls:
      mode: Terminate
      certificateRefs:
      - kind: Secret
        name: api-tls
    httpFilters:
    - name: verify-jwt
      type: ExternalAuth
      externalAuth:
        backendRef:
          name: jwt-verifier
          port: 8443
        protocol: HTTP
    - name: project-tenant-header
      type: ExtensionRef
      extensionRef:
        group: example.com
        kind: JWTClaimToHeader
        name: tenant-claim
```

The two filters above execute *in that exact order* on every request that
terminates on the `https` listener: first `verify-jwt` establishes verified
identity via an external authorization backend, then `project-tenant-header`
extracts the tenant claim from the JWT and projects it into a request header
that downstream HTTPRoutes match on. Only then does HTTPRoute matching run
against the mutated request.

### Filter Ordering (strict)

The `httpFilters` list is authoritative:

1. **Execution order MUST equal list order.** Implementations MUST NOT
   reorder, deduplicate, or coalesce entries.
2. **No skipping.** Implementations MUST NOT skip an entry because a later
   entry is inapplicable, unsupported, or short-circuits. If a filter is
   unsupported by the implementation, the listener MUST be rejected via
   status (`Accepted=False`, `Reason=UnsupportedValue`) rather than the
   filter silently omitted.
3. **Atomic list semantics.** The list is declared `+listType=atomic`, so
   server-side apply replaces the entire list rather than merging elements.
   This eliminates a whole class of ordering ambiguity that would otherwise
   arise from two appliers touching the same list.
4. **Unique `Name` per element.** `Name` MUST be unique within a single
   listener's `httpFilters` list. Names are consumed by status
   reporting so that a listener condition can name the failing filter (for
   example, `Reason=UnsupportedValue`, `Message="filter 'verify-jwt' type
   ExternalAuth is not supported by this implementation"`).
5. **Short-circuit behavior.** A filter that produces a response to the
   client directly instead of forwarding the request (for example,
   `ExternalAuth` returning deny, or an `ExtensionRef` that returns a
   response such as a rate-limit rejection) MUST prevent subsequent
   pre-routing filters from running for that request and MUST prevent
   HTTPRoute matching from running. The response returned to the client is
   the one produced by the short-circuiting filter.

The intent of this MUST is stated explicitly in godoc on the field, and is
verified by conformance tests (see [Conformance](#conformance)). This is a
deliberate contrast to
[`HTTPRouteRule.Filters`](../../reference/api-types/httproute.md#filters-optional),
which today only guarantees ordering as a `SHOULD`. Establishing the strict
guarantee on the *new* container avoids the compatibility problem that a
retroactive `SHOULD`->`MUST` change on `HTTPRouteRule.Filters` would create,
which is the resolution [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
converged on.

### Route Matching Interaction

Route matching is **single-pass**:

1. Every entry in `httpListenerFilters` executes, in order, on the incoming
   request.
2. After the last pre-routing filter completes, the
   implementation performs HTTPRoute matching against the (potentially
   mutated) request.
3. Implementations MUST NOT re-run any `httpListenerFilters` entries after
   the pre-routing stage completes.

Single-pass matching is chosen because it (a) eliminates loop-detection
semantics that different data planes model differently (NGINX rewrite loops vs. Envoy route-cache
invalidation vs. HAProxy frontend/backend split), and (b) preserves the
mental model that pre-routing is a discrete phase, not an interleaved
feedback loop with routing.

### Listener Match Invariance

Listener selection observes only inputs the client committed to *before* any
pre-routing filter runs, and is not re-evaluated as pre-routing filters
mutate the request. How that invariance is enforced depends on the listener
protocol:

* **Port** is fixed by the TCP handshake, for both HTTP and HTTPS listeners.
* **Protocol** is fixed by the TCP/TLS negotiation (HTTP/1.1, HTTP/2, TLS
  with SNI), for both.
* **Hostname (HTTPS listeners)** The client sends its Server
  Name Indication (SNI) in the TLS ClientHello, as the very first bytes on
  the connection, before any HTTP request bytes are exchanged. The gateway
  must pick the listener at that moment because the certificate presented in
  ServerHello must match the SNI. The TLS session keys are then derived from
  material tied to that specific certificate, so every HTTP byte on the
  connection is bound to the selected listener by cryptography. A pre-routing
  filter can rewrite `Host` / `:authority`, but it cannot renegotiate the
  certificate; the connection remains on the SNI-selected listener for its
  entire lifetime. When a later request's `Host` header disagrees with the
  SNI-selected listener, Gateway API's Listener spec instructs implementations
  to return `421 Misdirected Request` (or `404` if no other listener would
  have matched either) rather than re-dispatch to a different listener — see
  [`Listener.Hostname` spec](../../reference/api-types/gateway.md).
* **Hostname (cleartext HTTP listeners)** Cleartext HTTP has
  no pre-HTTP hostname signal (no SNI equivalent), and the spec allows
  successive HTTP requests on the same TCP connection to carry different
  `Host` headers — either via HTTP/1.1 keep-alive to distinct virtual hosts
  or via HTTP/2 multiplexing. Gateway API's spec therefore matches the
  listener **per request**, against the request's `Host` header. To keep
  pre-routing coherent with that per-request matching, this GEP specifies:
  1. Listener selection MUST use the request's `Host` header **as received
     on the wire**, before any pre-routing filter has run.
  2. Pre-routing filters MUST run within the filter chain of the
     initially-selected listener only. Mutations of `Host` (or any other
     header) by a pre-routing filter MUST NOT be re-fed into listener
     selection.
  3. If pre-routing filter mutations leave HTTPRoute matching within the
     selected listener with no matching rule, the request MUST be answered
     per the existing "no matching rule" rules (typically `404`). It MUST
     NOT be re-dispatched to a different listener's pre-routing chain, even
     if the mutated `Host` would have selected that other listener.

The net effect in both cases is identical: **pre-routing filters can affect
HTTPRoute selection, but not Listener selection.** For HTTPS the invariance
is enforced by the wire protocol (renegotiating the certificate is the only
protocol-legal way to re-select the listener, and pre-routing filters cannot
do that). For cleartext HTTP the invariance is enforced by this API contract
(listener selection observes only the initial `Host`, and filter mutations
do not feed back). Both prevent cross-listener leakage and both rule out
re-selection loops.

Concretely, walking the pre-routing filter variants:

| Filter | Can mutate route-matching inputs? | Can re-select the listener? |
|---|---|---|
| `ExternalAuth` | Yes (injected request headers derived from the external authorization backend's response) | No — HTTPS: cert binding; HTTP: initial-`Host` rule above. |
| `ExtensionRef` | Depends on extension | No — extension MUST NOT re-invoke listener selection. |

This invariance is what makes the Listener a safe attachment point for the
pre-routing filter payloads and disposes of the corresponding open
question from earlier revisions of this GEP.

### Interaction with Post-Routing Filters (including `ExternalAuth`)

Every request that reaches HTTPRoute matching has, by construction, already
passed the pre-routing phase. Post-routing filters
(`HTTPRouteRule.Filters`, including
[`ExternalAuth`](../gep-1494/index.md)) continue to run exactly as they do
today, on the matched rule, with unchanged semantics.

Key layering rules:

1. Pre-routing filters on a Listener run for **every** request on that
   listener, regardless of which HTTPRoute (if any) is later matched.
2. If HTTPRoute matching fails to select a rule, post-routing filters do not
   run, but pre-routing filters already have.
3. The same filter type may appear both pre-routing (on a Listener) and
   post-routing (on an HTTPRoute). Both instances execute; they are
   independent configurations.
4. For identity-driven routing, `ExternalAuth` is expected to be used
   pre-routing so that identity is established before HTTPRoute selection.
   The identical filter payload can still be used post-routing on individual
   `HTTPRouteRule`s to enforce route-scoped authorization. The two positions
   do not conflict: they are two independent executions of the same filter
   type at two different phases.

### Which Filter Types Are Legal Pre-Routing

Only a subset of the `HTTPRouteFilter` variants are permitted in an
`httpFilters` list. The bar for admission is deliberately narrow:
a variant earns a slot only if it can produce a **dynamic** signal — one
computed from the request itself or from a call to a per-request external
service — that route matching can then consume. Static configuration alone
does not clear this bar; static changes to the request could equally well be
made post-routing on the matched HTTPRouteRule without any loss of
expressiveness, so hoisting them to pre-routing adds an ordering surface
without adding capability.

| Filter | Pre-routing use case | Notes |
|---|---|---|
| `ExternalAuth` | **Establish verified identity before route selection.** The external authorization backend is a per-request call whose response (verdict, injected headers) becomes matching input. | Answers the "auth before routing" motivation and resolves the identity-derived routing user story. |
| `ExtensionRef` | **Escape hatch for custom pre-routing behavior.** Body-based routing (extracting a `model` field from a JSON body and projecting it into a header), JWT-claim projection, WASM plugins, and similar per-request dynamic computation. | Same implementation-specific conformance status as post-routing `ExtensionRef`. Every "pre-routing-shaped" capability not covered by `ExternalAuth` today enters via this variant until a follow-up GEP adds a first-class type. |

The excluded variants are enumerated in
[Excluded Filter Variants](#excluded-filter-variants) with per-variant
reasoning.

### Excluded Filter Variants

The remaining `HTTPRouteFilter` variants are deliberately **not** present
in `HTTPListenerFilter`. Each exclusion is a considered decision, not an
oversight; if a compelling motivation for one of these variants emerges,
adding it back is additive (append a variant to a distinct enum type) and
does not break any prior configuration.

* **`RequestHeaderModifier`.** In the current API, this filter can only apply
  a **static** header mutation. The`set`/`add`/`remove` values are literal
  strings baked into the resource, not expressions or per-request
  computations. A static mutation that changes route matching could
  be captured by writing HTTPRoutes that match the intended value in
  the first place. If Gateway API adds expression-based
  value support to `HTTPHeaderFilter`, this variant could be considered for
  pre-routing support.

* **`URLRewrite`.** This variant offers two modes today. **`ReplaceFullPath`**
  fully overwrites the path with a static value which could be accomplished by
  writing an HTTPRoute policy that matches the intended path directly.
  **`ReplacePrefixMatch`** rewrites a matched prefix, but there is no
  path prefix in the pre-routing phase to match against since matching hasn't
  happened yet.

* **`RequestRedirect`.** The same outcome can be achieved today with an
  HTTPRoute containing a single catch-all match and a `RequestRedirect`
  filter.

* **`RequestMirror`.** Mirroring clones the request to a separate backend
  and discards the mirror's response. It does **not** mutate the primary
  request in any way, so by construction it cannot influence which route
  the primary is matched to.

* **`ResponseHeaderModifier`.** This filter mutates headers on the response
  path. During the pre-routing phase, no backend has been selected, no
  request has been forwarded, and no response bytes exist. There are only
  two ways to interpret a `ResponseHeaderModifier` entry in a pre-routing
  list, and both are undesirable: either the filter is registered to run
  on the eventual response (in which case it doesn't actually *execute*
  in the pre-routing phase, and the strict-ordering guarantee has no
  meaning for it), or it is a no-op. The same mutation is already available
  on `HTTPRouteRule.Filters` in the correct phase; there is no capability
  gap.

* **`CORS`.** CORS is fundamentally a response-side concern: its output is
  `Access-Control-Allow-*` response headers, and the CORS preflight
  handshake is a request/response exchange rather than a mutation on the
  incoming request. Its natural home is post-routing, where the response
  headers can be tailored to the matched route, which is exactly what
  `HTTPRouteRule.Filters` already offers.

Future revisions of this GEP MAY re-admit any of the six above if a
concrete pre-routing use case emerges (for example,
`RequestHeaderModifier` becomes a strong candidate if expression-based
values are added to `HTTPHeaderFilter`), or MAY add pre-routing-only
variants (for example, a first-class body-projection filter once GEP-5091
lands). Both moves are additive because `HTTPListenerFilterType` is a
distinct enum from `HTTPRouteFilterType`.

### Extensibility to Other Route Kinds

This GEP intentionally ships HTTPRoute-only, but the API shape is
deliberately structured so that other route kinds can be added without
renaming, moving, or forking `httpFilters`:

* **`GRPCRoute`.** GRPCRoute is HTTP/2 in practice. If community consensus
  extends pre-routing to GRPCRoute, no new field is required: the same
  `Listener.httpFilters` (on an HTTP/HTTPS listener) applies. A
  future revision would loosen the "HTTPRoute only" restriction in godoc and
  extend conformance to cover GRPCRoute matching.
* **`TCPRoute`, `TLSRoute`, `UDPRoute` (L4).** These need protocol-appropriate
  filter payloads (source-IP blocking, SNI-based decisions, per-connection
  rate limits) that do not fit the HTTP-shaped `HTTPListenerFilter` type.
  The extension path is to add a **parallel** field on `Listener` — for
  example `tlsFilters []TLSListenerFilter`, `tcpFilters []TCPListenerFilter`,
  `udpFilters []UDPListenerFilter` — each with an element type defined by a
  follow-up GEP. The existing `httpFilters` field stays put and keeps its
  HTTP semantics; L4 pre-routing does not share a type with HTTP pre-routing.

The naming convention (`<protocol>Filters []<Protocol>ListenerFilter` per
protocol family) keeps each phase attached to a phase-appropriate payload
type without collapsing them into a single polymorphic list — which would
force implementations to inspect and validate cross-protocol combinations
that don't make sense. The HTTP prefix on today's field is not a placeholder;
it is the permanent name for the HTTP-family flavor.

### Status Reporting

Pre-routing filter health is surfaced through the new reasons
listed below and are additions to
[`ListenerConditionReason`](../../reference/api-types/gateway.md).

**New reasons introduced by this GEP:**

* **`Accepted=False`, `Reason=InvalidHTTPListenerFilter`.** Used when a filter in
  `httpListenerFilters` is structurally invalid — for example, if implementation-side
  validation rejects a combination CEL did not catch, or if `httpListenerFilters`
  is present on a Listener whose `Protocol` is not `HTTP` or `HTTPS`.
* **`ResolvedRefs=False`, `Reason=InvalidHTTPListenerFilterRef`.** Used when an
  `ExternalAuth.backendRef` or an `ExtensionRef` in `httpListenerFilters` refers
  to a resource that does not exist, has an unsupported group/kind, or is
  otherwise unresolvable. This is analogous to the existing
  `InvalidCertificateRef` reason for the TLS `certificateRefs` case.

**When these reasons are populated:**

Following the Kubernetes and Gateway API convention, these `Reason` values
are populated **only when the corresponding condition transitions to
`False`**. In the healthy case — every filter variant is supported and every
reference resolves — the standard reasons `Accepted=True (Reason=Accepted)`
and `ResolvedRefs=True (Reason=ResolvedRefs)` continue to apply, unchanged.
Implementations MUST NOT emit `Reason=InvalidHTTPListenerFilter`,
`Reason=InvalidHTTPListenerFilterRef`, or the reused error reasons on a `True`
condition, and MUST NOT emit them on Listeners that do not use `httpListenerFilters`
at all.

**Message content:**

When `httpListenerFilters` is the cause of a `False` condition, the `Message` field
SHOULD name the offending filter by its `Name` (if defined) and describe the specific
failure — for example, `"httpListenerFilters[verify-jwt]: ExternalAuth.backendRef
refers to Service default/jwt-verifier which does not exist"`. This mirrors
existing Gateway API practice for `InvalidCertificateRef` and similar
reasons.

### Conformance

* **Feature name:** `HTTPListenerFilters` (Extended). Implementations
  advertise support via the standard Gateway API feature list.
* **Sub-features:** one per permitted filter variant, so an implementation
  can advertise partial support:
  `HTTPListenerFilterExternalAuth`, `HTTPListenerFilterExtensionRef`.
* **Conformance test-only extension.** Several of the ordering and matching
  tests below require driving a *dynamic* header mutation from within a
  pre-routing filter. Because `RequestHeaderModifier` is excluded from the
  pre-routing set (it can only apply a static mutation, which does not
  exercise the phase's dynamic-signal property), these tests rely on a
  conformance-scoped `ExtensionRef` — call it `SetRequestHeader` — provided
  by the Gateway API conformance suite. Implementations claiming
  `HTTPListenerFilterExtensionRef` conformance MUST recognize this test extension
  in the conformance test namespace. This mirrors how Gateway API already
  uses test-only extensions to exercise implementation-specific surfaces.
* **Required conformance tests:**
  1. **Strict ordering.** Two `ExtensionRef` filters, each configured to set
     the same request header to different values, MUST result in the second
     value winning at the backend. Reversing the list order MUST reverse the
     observed outcome.
  2. **Listener match invariance (HTTPS).** With two HTTPS listeners on the
     same port carrying different hostnames and different pre-routing
     chains, a request whose SNI selects Listener A but whose pre-routing
     `SetRequestHeader` extension rewrites `Host` to Listener B's hostname
     MUST continue to be handled by Listener A. It MUST NOT be re-dispatched
     to Listener B's pre-routing chain. It MAY affect HTTPRoute selection
     within Listener A, and MAY result in a `421 Misdirected Request` or
     `404` per the existing Listener spec if no HTTPRoute in Listener A
     matches.
  3. **Listener match invariance (cleartext HTTP).** With two HTTP listeners
     on the same port carrying different hostnames and different pre-routing
     chains, a request whose initial `Host` header selects Listener A but
     whose pre-routing `SetRequestHeader` extension rewrites `Host` to
     Listener B's hostname MUST continue to be handled by Listener A's
     pre-routing chain and HTTPRoute set. It MUST NOT be re-dispatched to
     Listener B. If no HTTPRoute in Listener A matches the mutated request,
     the response MUST be `404`, not a re-dispatch to Listener B.
  4. **Single-pass route matching.** A pre-routing `SetRequestHeader`
     extension that mutates a header the HTTPRoute matches on MUST cause
     HTTPRoute matching to observe the mutated value; a second HTTPRoute
     that would match the *original* value MUST NOT be selected. HTTPRoute
     matching MUST run exactly once.
  5. **Unsupported-filter rejection.** A listener listing a filter variant
     the implementation does not support MUST be rejected with
     `Accepted=False`, `Reason=UnsupportedValue`, and the message MUST
     name the offending filter.
  6. **Wrong-protocol rejection.** A listener with `Protocol` set to a value
     other than `HTTP` or `HTTPS` (for example, `TCP`) whose `httpFilters`
     is non-empty MUST be rejected with `Accepted=False`,
     `Reason=InvalidListenerFilter`.
  7. **Unresolvable filter reference.** A listener whose `httpListenerFilters`
     contains an `ExternalAuth.backendRef` (or `ExtensionRef`) that does
     not exist MUST have `ResolvedRefs=False`,
     `Reason=InvalidListenerFilterRef`, and the message MUST identify which
     filter's reference failed to resolve. A cross-namespace
     `ExternalAuth.backendRef` without a matching ReferenceGrant MUST have
     `ResolvedRefs=False`, `Reason=RefNotPermitted`.
  8. **Short-circuit on deny.** An `ExternalAuth` filter that returns deny
     MUST prevent subsequent pre-routing filters and HTTPRoute matching from
     running.
  9. **Name uniqueness.** A listener with two entries sharing the same
     `Name` MUST be rejected at admission by CEL validation.

## Alternatives Considered

The following API shapes and semantic choices were evaluated and rejected in
favor of the design in [API](#api). Each is captured here with the reasoning
so that the decision is auditable if the design needs to be revisited.

### Top-level `HTTPListenerFilterPolicy` CRD via Policy Attachment (GEP-713)

**Considered.** Introduce a new CRD (analogous to
[`BackendTrafficPolicy`](../gep-3388/index.md)) with `targetRefs` pointing at
`Gateway`, `Listener` (via `sectionName`), or `ListenerSet`. Filters live in
the policy `spec.filters`.

**Rejected** because:

* Cross-cutting ordering across multiple attached policies requires a
  merge/priority mechanism.
* Ordering is load-bearing here (see [Filter Ordering](#filter-ordering-strict))
  and is unambiguous when it lives on a single ordered list on a single
  resource. A separate policy CRD reintroduces the "whose list wins?"
  question every time two policies target the same listener.
* The Listener already owns the notion of "traffic terminating here"; putting
  the filter list on the Listener keeps the ordering scope obvious.
* A future revision *may* add a Policy overlay for cross-listener reuse
  without removing the embedded field. Nothing about the Listener-embedded
  design forecloses that option.

### Attach at Gateway (`GatewaySpec.httpFilters`)

**Considered.** A single shared list at Gateway scope, applied uniformly to
every listener on the Gateway.

**Rejected** because:

* Filter payloads depend on protocol (an HTTP filter on a TCP listener is
  meaningless). Listener scope makes protocol-appropriateness expressible
  directly, without additional per-protocol validation on a Gateway-wide list.
* Different listeners on the same Gateway routinely need different
  pre-routing configurations. For example, an internal-only listener that
  skips auth. A Gateway-wide list forces every listener to share.
* Users who *want* to share a filter set across listeners can already do so
  with [GEP-1713 ListenerSets](../gep-1713/index.md), which shares listener
  definitions (and their `httpFilters`) across Gateways. We do not
  need a second sharing mechanism specifically for pre-routing.

### Attach at ListenerSet (dedicated ListenerSet field)

**Considered.** Add `httpFilters` directly to the `ListenerSet`
resource rather than to individual `Listener` entries.

**Not adopted, but not conflicting.** A `ListenerSet` is a collection of
`Listener` definitions. Putting `httpFilters` on individual
`Listener` structs inside a `ListenerSet` already works transparently — users
who want to share pre-routing filters across Gateways define a `ListenerSet`
with the appropriate filters set on each `Listener` inside it. No dedicated
ListenerSet-level field is required. If aggregate ListenerSet-scope semantics
are wanted later, they can be added additively.

### Reuse `HTTPRouteFilter` directly in a `[]HTTPRouteFilter` on the Listener

**Considered.** Embed `HTTPRouteFilters []HTTPRouteFilter` on the Listener
and add a MUST-ordering rule scoped to that particular container.

**Rejected** because:

* The MUST-vs-SHOULD ordering distinction between the two phases would live
  on the *container* rather than the *element type*. Godoc that says "the
  ordering rule depends on where this filter appears" is subtle and easy for
  implementations to miss.
* Any future divergence. A pre-routing-only filter variant, or narrowing
  the pre-routing legal set would either bloat `HTTPRouteFilter` or force
  a fork after the fact.
* A dedicated `HTTPListenerFilter` type keeps the ordering guarantee
  visually attached to the element via godoc and gives us room to diverge
  without breaking `HTTPRouteFilter` consumers.

### Add pre-routing filters as a field on `HTTPRoute`

**Considered.** Express pre-routing filters via a new field on `HTTPRoute`
(analogous to `HTTPRouteRule.Filters` but at the top of the resource).

**Rejected** because HTTPRoute matching has not happened yet when pre-routing
filters run, so a filter attached to an HTTPRoute cannot influence *which*
HTTPRoute is selected. This is a fundamental chicken-and-egg conflict with
the phase definition established in
[Pre-Routing vs. Post-Routing](#pre-routing-vs-post-routing-the-core-distinction).

### Multiple ordered filter policies with a priority integer

**Considered.** Allow multiple `HTTPListenerFilterPolicy` resources per
listener and use an integer `priority` field to order them.

**Rejected** because:

* Priority collisions and re-numbering are operationally painful in
  practice.
* A single ordered list is unambiguous, matches
  [`HTTPRouteRule.Filters`](../../reference/api-types/httproute.md#filters-optional)'s
  existing shape, and does not require the API to define collision behavior.

### Multi-pass route matching with a loop guard

**Considered.** Allow HTTPRoute matching to re-invoke the pre-routing chain
if a filter after HTTPRoute selection mutates a matching input.

**Rejected** because:

* Introduces loop-detection semantics that different data planes model very
  differently. For example,NGINX rewrite loops, Envoy `clearRouteCache()`, and
  HAProxy frontend/backend split.
* Single-pass matches this GEP's stated position and mirrors GEP-5091.
* Complex enough to warrant its own follow-up GEP if a concrete use case
  emerges.

### Applying MUST ordering retroactively to `HTTPRouteRule.Filters`

**Considered.** Rather than a new list with strict ordering, elevate
`HTTPRouteRule.Filters` from `SHOULD` to `MUST` and add pre-routing filters
into the same list.

**Rejected** because:

* `HTTPRouteRule.Filters` is GA. A retroactive `SHOULD` -> `MUST` change would
  invalidate previously-conforming implementations, which is a
  compatibility break Gateway API does not typically accept for stable APIs.
* Establishing strict ordering on a *new* container (the pre-routing list)
  gives us the guarantee where we need it without touching existing
  implementations. This is the resolution
  [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
  converged on.

## Open Questions

* **What is the relationship to implementation-specific matches
  ([GEP-3965](../gep-3965/index.md))?** Some data
  planes can match on a derived signal natively without a project-into-header
  step. The user-facing UX should be consistent whether or not the
  intermediate projection is needed.
* **How do the three splits compose?** Body-based routing and external callouts
  are kinds of processing that will often want to run pre-routing. The seam
  between "the phase" (this GEP) and "the processing" (the other two) must stay
  clean so the pieces compose without overlap.
* **How should conformance handle implementations that support pre-routing
  only via an external component?** Some cloud load balancers cannot express
  arbitrary pre-routing filter chains natively. Whether such implementations
  can conform by delegating to a sidecar, whether they must decline
  `httpFilters` entirely, or whether a partial conformance level
  is defined, is left open.
* **Does `ExternalAuth` need pre-routing-specific configuration knobs?** The
  identical `HTTPExternalAuthFilter` payload works in both phases today, but
  pre-routing usage may motivate additional fields (for example, whether the
  extauth response can rewrite `Host` for the purposes of subsequent
  HTTPRoute matching) that don't make sense post-routing. Whether such
  fields belong on `HTTPExternalAuthFilter` or on a new pre-routing-specific
  variant is deferred.
* **Extension to other route kinds.** Extending pre-routing to `GRPCRoute`
  (HTTP/2, likely reuses `httpFilters`), and to L4 route kinds
  (`TCPRoute`, `TLSRoute`, `UDPRoute` — likely parallel fields with distinct
  element types) is deferred to follow-up GEPs. See
  [Extensibility to Other Route Kinds](#extensibility-to-other-route-kinds).

## Relationship to Other Work

* **[GEP-5091: PayloadProcessor Resource][gep-5091]** — the parent proposal.
  Pre-routing filters were split out of it. GEP-5091 already defines the
  pre-routing / post-routing phase model for payload processing; this GEP
  generalizes the phase itself.
* **[#5194: Filter ordering][issue-5194]** — this GEP is the community's chosen
  vehicle for resolving the pre-routing part of that discussion.
* **[GEP-1494: HTTP Auth / ExternalAuth][gep-1494]** — `ExternalAuth` is a
  post-routing filter today; identity-derived routing is a motivating
  pre-routing use case, so the two must be reconciled.
* **[GEP-91: Client Certificate Validation](../gep-91/index.md)**
  and **[GEP-2643: TLS/SNI-based routing](../gep-2643/index.md)**
  — lower-layer pre-routing decisions (TLS handshake, SNI) that establish
  context an HTTP-layer pre-routing filter might consume.
* **[GEP-713: Policy Attachment](../gep-713/index.md)**
  and **[GEP-1713: ListenerSets](../gep-1713/index.md)**
  — candidate structural foundations for a before-routing attachment point.

[gep-1494]: ../gep-1494/index.md

## References

* [GEP-5091: PayloadProcessor Resource (PR #5092)](https://github.com/kubernetes-sigs/gateway-api/pull/5092)
* [Issue #5194: Filter ordering — support guaranteed order of filters](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* [GEP-1494: HTTP Auth in Gateway API](../gep-1494/index.md)
* [GEP-1767: CORS Filter](../gep-1767/index.md)
* [GEP-713: Metaresources and Policy Attachment](../gep-713/index.md)
* [GEP-1713: ListenerSets](../gep-1713/index.md)
* [GEP-3965: HTTPRoute Implementation-Specific Matches](../gep-3965/index.md)
* [GEP-2643: TLS/SNI-based routing](../gep-2643/index.md)
* [GEP-91: Client Certificate Validation](../gep-91/index.md)
* [Envoy: HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_filters)
* [Envoy: HTTP routing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_routing)
* [NGINX: request processing phases](https://nginx.org/en/docs/dev/development_guide.html#http_phases)
* [Predicate Routing (NGINX Gateway Fabric #5737)](https://github.com/nginx/nginx-gateway-fabric/issues/5737)
* [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
