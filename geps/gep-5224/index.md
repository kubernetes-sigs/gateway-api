---
title: "GEP-5224: Pre-Routing Filters"
---

* Issue: [#5224](https://github.com/kubernetes-sigs/gateway-api/issues/5224)
  * Split out of [GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/pull/5092)
  * Resolves the pre-routing portion of [#5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* Status: Provisional

## TLDR

This GEP proposes that Gateway API define a way to express **pre-routing
filters**: processing steps that run **before** a route is selected and that
may therefore influence which route, and ultimately which backend, is chosen.
Every filter Gateway API defines today, from `RequestHeaderModifier` to
`ExternalAuth`, is a **post-routing** filter: it is attached to an
`HTTPRouteRule` and only runs *after* a route has already been matched. There
is no portable mechanism for a filter to act on a request before route
selection.

This provisional revision is scoped to the *what*, *who*, and *why*. It
establishes that pre-routing processing is a distinct and necessary phase, and
that the distinction between pre-routing and post-routing is a first-class
architectural property rather than an implementation detail. The concrete API
shape (a Gateway/listener-level attachment vs. a new construct vs. something
else), the set of filters allowed in the pre-routing phase, and the ordering
and re-matching semantics are deliberately left open here and enumerated as
*options under consideration*. They will be resolved at the Experimental stage.

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
> — `HTTPRouteRule.Filters`, [apis/v1/httproute_types.go](../../apis/v1/httproute_types.go)

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
> — `HTTPRouteRule.Filters`, [apis/v1/httproute_types.go](../../apis/v1/httproute_types.go)

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
* Establish that ordering among pre-routing filters is significant and should
  be a guarantee from the introduction of the phase, providing a concrete
  answer to the pre-routing portion of [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194).
* Enumerate the API-shape and semantic decisions that must be resolved at the
  Experimental stage, without committing to any of them here.

## Non-Goals

* **Selecting a concrete API shape.** Whether pre-routing filters attach at the
  `Gateway` level, the listener level, via a new construct on `HTTPRoute`, via
  policy attachment, or some combination, is deferred. See
  [Options Under Consideration](#options-under-consideration).
* **Defining which specific filters are pre-routing-capable.** Whether existing
  filter types can run in the pre-routing phase, and which new ones are
  introduced, is out of scope for a provisional GEP.
* **Body-level processing semantics.** How the body is addressed, buffered, or
  mutated is the subject of the body-based routing / [GEP-5091](https://github.com/kubernetes-sigs/gateway-api/pull/5091) work.
  This GEP is about the *phase*, not the *body*.
* **External processing wire protocols.** Invoking an external service is the
  subject of the external-callouts split. Pre-routing filters may eventually
  host such callouts, but the protocol is out of scope here.
* **Re-litigating post-routing filter ordering.** [#5194](https://github.com/kubernetes-sigs/gateway-api/issues/5194) asks a
  broader question about the existing GA filter list. This GEP only commits to
  strict ordering *within the new pre-routing phase*; the post-routing question
  is left to that issue.
* **Replacing existing filters or route matching.** Pre-routing filters
  complement `HTTPRoute` matching and existing filters; they do not replace
  header/path/method matching or post-routing transformation.
* **TCP/UDP/TLS-layer processing.** SNI/TLS-based selection
  ([GEP-2643](https://github.com/kubernetes-sigs/gateway-api/pull/2643)) is a different,
  lower-layer form of pre-routing and is not in scope here.

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

## Options Under Consideration

These design dimensions are enumerated so that agreement on motivation and the
phase model is not blocked on any one of them. All are deferred to the
Experimental stage.

* **Attachment point and API shape.** Because pre-routing filters cannot be
  scoped to a matched route, they need an attachment point that exists before
  routing, for example a `Gateway`, a listener / `ListenerSet`
  ([GEP-1713](https://gateway-api.sigs.k8s.io/geps/gep-1713/)), or a new
  construct. Whether this is expressed via policy attachment
  ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/)), a dedicated
  field, or a new resource is open.
* **Which filters are allowed pre-routing.** Whether the pre-routing phase
  reuses existing filter types, defines a new set, or both, is unresolved.
* **Re-matching semantics.** Whether routing is evaluated exactly once after
  all pre-routing filters run (the provisional position, mirroring
  [GEP-5091][gep-5091]), or whether any re-entry is allowed, must be defined.
  Avoiding feedback loops argues for single-pass matching.
* **Ordering guarantee mechanics.** This GEP takes the position that
  pre-routing ordering is significant and should be guaranteed from the start;
  exactly how that is specified and conformance-tested is a design detail for
  the Experimental stage.
* **Interaction with post-routing filters and `ExternalAuth`.** The ordering
  and layering between the pre-routing phase and the existing post-routing
  filter list, especially authorization, must be specified rather than left
  implementation-defined.

## Open Questions

* **How does a pre-routing mutation interact with route selection precisely?**
  The provisional position is single-pass: pre-routing filters run in order,
  then matching evaluates the mutated request once. This must be reconciled
  with data-plane behaviors such as Envoy's route-cache invalidation and
  NGINX's rewrite-triggered re-selection.
* **What is the relationship to identity and `ExternalAuth`?** Some routing
  decisions depend on verified identity. Whether identity establishment can or
  should run in the pre-routing phase, and how that orders against the existing
  post-routing `ExternalAuth` filter ([GEP-1494][gep-1494]), is open.
* **What is the relationship to implementation-specific matches
  ([GEP-3965](https://gateway-api.sigs.k8s.io/geps/gep-3965/))?** Some data
  planes can match on a derived signal natively without a project-into-header
  step. The user-facing UX should be consistent whether or not the
  intermediate projection is needed.
* **How do the three splits compose?** Body-based routing and external callouts
  are kinds of processing that will often want to run pre-routing. The seam
  between "the phase" (this GEP) and "the processing" (the other two) must stay
  clean so the pieces compose without overlap.

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
* **[GEP-91: Client Certificate Validation](https://gateway-api.sigs.k8s.io/geps/gep-91/)**
  and **[GEP-2643: TLS/SNI-based routing](https://gateway-api.sigs.k8s.io/geps/gep-2643/)**
  — lower-layer pre-routing decisions (TLS handshake, SNI) that establish
  context an HTTP-layer pre-routing filter might consume.
* **[GEP-713: Policy Attachment](https://gateway-api.sigs.k8s.io/geps/gep-713/)**
  and **[GEP-1713: ListenerSets](https://gateway-api.sigs.k8s.io/geps/gep-1713/)**
  — candidate structural foundations for a before-routing attachment point.

[gep-1494]: https://gateway-api.sigs.k8s.io/geps/gep-1494/

## References

* [GEP-5091: PayloadProcessor Resource (PR #5092)](https://github.com/kubernetes-sigs/gateway-api/pull/5092)
* [Issue #5194: Filter ordering — support guaranteed order of filters](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* [GEP-1494: HTTP Auth in Gateway API](https://gateway-api.sigs.k8s.io/geps/gep-1494/)
* [GEP-1767: CORS Filter](https://gateway-api.sigs.k8s.io/geps/gep-1767/)
* [GEP-713: Metaresources and Policy Attachment](https://gateway-api.sigs.k8s.io/geps/gep-713/)
* [GEP-1713: ListenerSets](https://gateway-api.sigs.k8s.io/geps/gep-1713/)
* [GEP-3965: HTTPRoute Implementation-Specific Matches](https://gateway-api.sigs.k8s.io/geps/gep-3965/)
* [GEP-2643: TLS/SNI-based routing](https://gateway-api.sigs.k8s.io/geps/gep-2643/)
* [GEP-91: Client Certificate Validation](https://gateway-api.sigs.k8s.io/geps/gep-91/)
* [Envoy: HTTP filters](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_filters)
* [Envoy: HTTP routing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/http/http_routing)
* [NGINX: request processing phases](https://nginx.org/en/docs/dev/development_guide.html#http_phases)
* [Predicate Routing (NGINX Gateway Fabric #5737)](https://github.com/nginx/nginx-gateway-fabric/issues/5737)
* [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
