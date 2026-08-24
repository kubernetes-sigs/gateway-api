---
title: "GEP-5091: PayloadProcessor Resource"
---

* Issue: [#5091](https://github.com/kubernetes-sigs/gateway-api/issues/5091)
  * Incubated by the [AI Gateway Working Group](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/7-payload-processing.md)
* Status: Provisional

## TLDR

This GEP proposes a new `PayloadProcessor` resource that enables declarative,
ordered processing of HTTP request and response **payloads** (headers *and*
body) within the Gateway API framework. Today, Gateway API filters operate on
headers, paths, and query parameters but do not define mechanisms for acting on the request and response
body. Modern workloads, particularly AI inference, require body-level
processing for routing, security, and compliance decisions.

This provisional revision is scoped to the *what*, *who*, and
*why*. It establishes that payload processing belongs in Gateway API and
discusses a structural distinction that the API will have to
represent: **pre-routing** vs. **post-routing** processing. The concrete
API shape (policy attachment vs. HTTPRoute filter vs. an inline construct),
the processing strategy (in-data-plane expressions vs. an external
processing service), and the expression / body-addressing language are
deliberately left open here and enumerated as *options under
consideration*. They will be resolved at the Experimental stage.

## Motivation

Gateway API provides a powerful, extensible framework for configuring HTTP
routing in Kubernetes. However, its current processing model is
fundamentally limited to metadata-level operations: headers, paths, query
parameters, and method. There is no standardized mechanism for Gateway API
implementations to inspect or act on the **body** of a request or response.
This gap creates friction in several areas.

### No API Mechanism for Response Access and Modification

Gateway API's `HTTPRoute` filters (`RequestHeaderModifier`, `RequestRedirect`,
`URLRewrite`, `RequestMirror`, `ExtensionRef`, `ExternalAuth`) all operate on
request metadata. `ExternalAuth` can additionally forward the request body
to an external service, but no core filter can read or act on the *response*
body. Patterns that need response access and modification therefore require
implementation-specific extensions with no portability, or are shoehorned
into filters whose semantics do not fit. For example, `ExternalAuth` is
scoped to authorization decisions on the request, not to response mutation.

### AI Inference Requires Body-Level Decisions

AI inference workloads send model selection, prompt content, and
configuration in the request body (typically JSON). Key decisions, which
model to route to, whether the prompt contains PII or an injection attack,
whether the response can be served from a cache, all require reading the
body. Today, [llm-d] implements a Body-Based Router (BBR) to extract model
names for routing. That BBR is the primary implementation of the pluggable
BBR framework proposed by the [Gateway API Inference Extension (GAIE)]. That
proposal is in a draft state and its reference implementation no longer
lives inside the GAIE repo, so downstream users have no portable, upstream
answer for what is a foundational AI-gateway pattern.

### Body-Level Processing Is Not Portable Today

A significant class of payload processing consists of deterministic
field-level operations: reading a value from a JSON body, copying it into a
header, removing a field, or setting a field to a fixed value. These
operations require no external state or networking calls, and can be
performed locally by existing data planes. Today the only portable option
is to forward the entire payload to an external service, which adds a
network round trip and requires a separately deployed, scaled, and secured
workload. Equivalent in-data-plane mechanisms exist across data planes,
for example, Envoy's
[`transform`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/transform_filter)
and
[`json_to_metadata`](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/json_to_metadata_filter)
filters, or NGINX's forthcoming
[native body-matching directives][nginx-predicate-routing], but they are
not united under a common API.

### Composability Gap

Real-world payload processing requires ordered, composable pipelines, for
example, "first extract the model name for routing, then scan for PII, then
check for prompt injection." Current approaches require either monolithic
external services or implementation-specific chaining mechanisms.

[Gateway API Inference Extension (GAIE)]: https://github.com/kubernetes-sigs/gateway-api-inference-extension
[llm-d]: https://github.com/llm-d/llm-d-inference-payload-processor
[nginx-predicate-routing]: https://github.com/nginx/nginx-gateway-fabric/issues/5737

## Goals

* Establish agreement that Gateway API needs a standardized mechanism for
  declarative, ordered processing of HTTP request and response payloads
  (including the body), and that this belongs in the Gateway API family of
  APIs rather than only in implementation-specific extensions.
* Distinguish **pre-routing** and **post-routing** payload processing as
  first-class phases, because the ability of body mutation to change
  routing decisions has direct implications for how the mechanism must be
  defined and implemented.
* Cover both **routing-affecting** operations (extract a value from the
  body and use it to select a route) and **non-routing-affecting**
  operations (scan, redact, enrich, or validate a payload after the route
  is chosen).
* Support **ordered, deterministic composition** of multiple processing
  steps so pipelines such as "extract → validate → scan → enrich" are
  expressible without inventing implementation-specific chaining.
* Enumerate the API-shape and processing-strategy decisions that must be
  resolved at the Experimental stage, without committing to any of them
  here.

## Non-Goals

* **Selecting a concrete API shape**: whether payload processing is
  expressed as a policy-attached resource, an inline HTTPRoute filter, an
  HTTPRoute filter that references a separate resource, or something else,
  is deferred. See [Options Under Consideration](#options-under-consideration).
* **Selecting an expression language or body-addressing scheme**: CEL,
  JSONPath, JSON Pointer, whole-body expressions, and other choices are
  noted but not decided here.
* **Standardizing an external processing wire protocol**: needed before
  the external-processing dimension can graduate, but out of scope for a
  provisional GEP. Envoy `ext_proc` and
  [agentgateway's implementation](https://github.com/agentgateway/agentgateway/tree/main/payload-processor-poc)
  can be referenced as prior art.
* **Streaming / chunked body processing**: this GEP assumes buffered
  bodies at the point of processing. Streaming semantics are deferred.
* **TCP/UDP/TLS payload processing**: these protocols lack the structured,
  inspectable body model this proposal depends on. GRPCRoute is a
  plausible follow-up but is not in scope here.
* **Replacing existing HTTPRoute filters**: payload processing
  complements, not replaces, `RequestHeaderModifier` and similar filters.
  Header-only operations should continue to use existing filters.

## User Stories

The following stories describe *what* users want to accomplish. They are
intentionally described independently of whether the underlying processing
runs in-data-plane, in an external service, or both, that dimension is
one of the API design questions and is discussed under
[Options Under Consideration](#options-under-consideration).

### Body-Based Routing

**As an AI Platform Engineer:**

> "I want to route inference requests to the correct model backend based
> on the `model` field in the JSON request body, without modifying my
> application or using implementation-specific extensions. Today I use a
> custom Body-Based Router API and implementation. I want a portable
> Gateway API answer."

**As a Developer of Agentic AI Platforms:**

> "I need to process Model Context Protocol (MCP) request payloads to
> extract tool names and session identifiers for routing decisions, so
> that the gateway can route to the correct backend MCP server."

### Security and Compliance

**As a Compliance Officer:**

> "I want to strip known sensitive fields from request and response
> bodies, and reject payloads matching a defined pattern, so that
> regulated data never reaches a backend or a client. Because the data
> is sensitive, I do not want it forwarded to an additional service in
> order to be inspected. I need this to be declarative, auditable, and
> composable with other processing steps."

**As a Security Engineer:**

> "I want to add a processing step that classifies inference request
> bodies for prompt injection attacks before they reach the model
> backend. If the scan detects a threat, the request should be rejected
> with a clear error. If the scanning service is unavailable, I want
> per-step control over whether the request is rejected (fail-closed for
> security processors) or allowed through (fail-open for non-critical
> enrichment)."

**As a Compliance Officer (Response Inspection):**

> "I want to examine inference responses for personally identifiable
> information that cannot be expressed as a fixed pattern, so that it
> can be blocked, sanitized, or reported."

### Request and Response Enrichment

**As a Platform Engineer:**

> "I want to normalize inference requests before they reach a backend,
> force usage accounting, apply defaults for fields the client omitted,
> and reject requests that are missing a required field. Today this
> requires either changing every client or deploying an external service
> to edit two keys in a JSON document."

**As an API Owner:**

> "I want to enrich requests with context derived from the payload and
> the verified caller identity, request identifiers for tracing, tenant
> headers for downstream attribution, without adding a network hop to a
> request that is otherwise served entirely from cache."

### Optimization

**As a Cluster Administrator:**

> "I want to add semantic caching to inference requests, detecting
> repeated or semantically similar requests and returning cached results
> to reduce inference costs and improve latency."

### Implementation Consistency

**As a Gateway API Implementation Author:**

> "I want a clear, standardized definition of payload processing so I
> can implement it consistently. I need the specification to be
> unambiguous about ordering and phase semantics, and about how payload
> processing interacts with existing HTTPRoute filters and with
> `ExternalAuth`."

## Processing Phases

Independent of API shape, payload processing has to distinguish between
two phases relative to route selection. This distinction is elevated to a
Goal because body mutation is exactly the case where the two phases
behave differently, and any API for payload processing must be able to
express both.

| Phase | When it runs | Why it matters |
|-------|--------------|----------------|
| **PreRouting** | Before HTTPRoute matching | Mutations can influence the route that is selected, for example, extract `model` from a JSON body, project it into a header, and let standard HTTPRoute header matching select the backend. |
| **PostRouting** | After the route is selected, before backend dispatch (and, symmetrically, on the response before it reaches the client) | Mutations cannot re-shape routing, only the payload delivered to the chosen backend or returned to the client, for example, PII redaction, response scanning, response enrichment. |

```
Client Request
    │
    ▼
┌──────────────────────┐
│  PreRouting Phase    │ ◄── Payload processing that CAN influence routing
│                      │     (extract, project, validate)
└──────────┬───────────┘
           │ (headers / metadata potentially mutated)
           ▼
┌──────────────────────┐
│  HTTPRoute Matching  │ ◄── Standard header/path/method matching
└──────────┬───────────┘
           │ (route selected)
           ▼
┌──────────────────────┐
│  HTTPRoute Filters   │ ◄── Existing filters (RequestHeaderModifier, etc.)
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  PostRouting Phase   │ ◄── Payload processing that CANNOT influence routing
│                      │     (redact, enrich, scan request or response body)
└──────────┬───────────┘
           │
           ▼
       Backend
```

Elevating this to a first-class concern has two consequences for any API
design:

1. Any resource that expresses payload processing must be able to say
   *when* in this pipeline it runs. The two phases are semantically
   distinct, a PostRouting processor cannot change routing regardless
   of what it does, and a PreRouting processor cannot be scoped by
   matched-route information because no route has been matched yet.
2. Interactions with existing HTTPRoute filters and with `ExternalAuth`
   must be defined explicitly. Because payload mutation can affect
   authorization and header-derived filters, the ordering between
   payload processing and the existing filter list is a specification
   decision, not an implementation detail.

## Options Under Consideration

The following design dimensions are enumerated so agreement on the
motivation and phase model is not blocked on any one of them. All are
deferred to the Experimental stage of this GEP.

### API Shape

Whether payload processing is expressed as:

* A **standalone resource with policy attachment** to `Gateway` and/or
  `HTTPRoute`, following the [GEP-713] pattern. Enables reuse across
  routes and expresses Gateway-level pre-routing naturally, at the cost
  of the operational-discoverability concerns that other Gateway API
  policies have surfaced.
* An **inline HTTPRoute filter**. Discoverable in-line with routing,
  orders naturally against existing filters, but cannot express
  pre-routing processing that must run before HTTPRoute matching (there
  is no matched route yet), and duplicates configuration across routes.
* An **HTTPRoute filter that references a separate resource**, so the
  filter list carries ordering while the processing definition is
  reused. Only addresses the post-routing case; pre-routing would still
  need a different attachment point.
* A **new inline construct on `HTTPRoute.spec`** that is *not* an
  existing filter, expressing pre-routing separately from the rule-level
  filter list. Considered on the PR discussion as a way to keep
  configuration inline while still distinguishing pre- and post-routing.

The strongest argument for a filter (or filter-like) shape came from
reviewers: because payload mutation can change routing, auth, and
header-derived decisions, ordering with existing HTTPRoute filters and
with `ExternalAuth` must be strict, and inline expression makes that
ordering easier to specify and reason about. The strongest argument for
a policy-attached shape is reuse and Gateway-level pre-routing.

[GEP-713]: https://gateway-api.sigs.k8s.io/geps/gep-713/

### Processing Strategy

Whether processing runs:

* **In the data plane**, driven by an expression language over a defined
  set of variables (for example, extracting a JSON field and projecting
  it into a header). Deterministic, no extra network hop, no separate
  workload, and sensitive payload data stays inside the proxy.
* **In an external service** invoked via a defined wire protocol.
  Enables processing that needs a trained model, external state, or a
  large and frequently updated ruleset. Requires a standardized protocol
  for portability.
* **Both**, in the same ordered pipeline. Most real-world configurations
  will mix the two, for example, fast in-data-plane extraction for
  routing followed by an external classifier for prompt-injection
  detection.

An earlier draft of this GEP treated in-data-plane vs. external service
as the top-level framing of the resource. Reviewer feedback pushed back
on this: the distinction is a mechanism, not a user-facing feature. It
is retained here as a design dimension to be resolved at the
Experimental stage, alongside the wire-protocol decision for the
external variant.

### Expression Language and Body Addressing

For any in-data-plane variant, the choice of expression language (for
example, CEL) and the choice of body-addressing scheme (JSONPath, JSON
Pointer, whole-body expressions) are open. The prior draft of this GEP
compared CEL against JSONPath / JMESPath, Rego, and Lua/WASM, and
separately compared JSONPath ([RFC 9535]), JSON Pointer ([RFC 6901]),
and whole-body CEL expressions. That comparison will be reintroduced at
the Experimental stage. Reviewers noted that consistency of the
standard library across implementations is a portability concern
regardless of which language is chosen, a MUST-level minimum of
supported functions is likely required for portable conformance.

[RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535
[RFC 6901]: https://www.rfc-editor.org/rfc/rfc6901

### Native Body Matching vs. Header-Extraction Bridge for Body-Based Routing

A specific design question surfaced by
[NGINX Gateway Fabric's body-based routing work][nginx-predicate-routing]
is whether Gateway API should express body-based routing as a two-stage
pattern (extract from body → project into a header → match on the header
via `HTTPRoute`) or as a body-focused match on `HTTPRoute` itself. Some
implementations can match on JSON body content natively and would not
require the intermediate extraction hop. Even where the underlying data
plane can match natively, the user-facing UX for expressing "route on a
body field" should be consistent across implementations. Whether that
consistent UX is the extraction-and-header-match pattern, a dedicated
body-focused match on `HTTPRouteRule`, or both, is left open. A
body-focused `HTTPRouteMatch` would fall outside the scope of a
payload-processing resource but would interact with it directly.

## Open Questions

The following high-level questions must be resolved before this GEP
moves to Experimental. Detailed follow-on questions (per-processor
timeouts, CEL cost budgets, buffer size configuration, parallel
execution, confidential-data injection, and so on) will be addressed
once the API shape is chosen.

### Does this belong in Gateway API core, or as an extension?

The functionality is broad enough to serve users well beyond AI
gateways, compliance redaction, request normalization, and enrichment
are long-standing patterns. Whether the resource ships in the core
Gateway API repository or as an experimental/extension API is
unresolved and will influence the API shape decision above.

### How does payload processing order relative to existing filters and `ExternalAuth`?

Because mutation can change headers that authorization and other
filters depend on, the ordering rule cannot be left
implementation-defined. This is also the reason reviewers pushed for a
filter-like shape: filter ordering is already a specified property of
`HTTPRoute`, and payload processing that mutates the request needs to
slot into that order predictably. A related open question is whether
the current "implementations MAY implement filter ordering strictly"
should be upgraded to MUST as part of this work.

### How do pre-routing mutations interact with route selection?

The provisional position is: pre-routing processors execute, then
HTTPRoute matching evaluates the (potentially mutated) request. There
is no re-entry into pre-routing based on the matched route, and no
re-matching after post-routing mutations. This avoids feedback loops
but limits some advanced use cases and interacts with
implementation-specific behavior such as Envoy's route-cache
invalidation.

### Do pre-routing and post-routing belong in one resource, or in two?

Reviewers suggested that pre-routing and post-routing may want to be
distinct extension points, since only one of them can affect routing.
If they are one resource with a phase discriminator, the discriminator
has to be part of the API. If they are two separate constructs, they
can be scoped independently, for example, pre-routing at the Gateway
or listener level, post-routing per-`HTTPRouteRule`.

### What is the relationship to the Gateway API Inference Extension Body-Based Router?

GAIE's BBR extracts the model name from inference request bodies to
select the appropriate `InferencePool`. A portable payload-processing
mechanism could implement the same pattern in a reusable way. Whether
the intent is to eventually re-express BBR on top of this mechanism, or
to keep them separate, is unresolved.

### What is the relationship to the Firewall GEP?

The [Firewall GEP (#3614)][firewall-gep] proposes firewall-like
filtering capabilities. Payload processing and firewall have
complementary scopes, network- and header-level rules vs. body-level
processing, but the boundary should be made explicit before either
graduates.

[firewall-gep]: https://github.com/kubernetes-sigs/gateway-api/issues/3614

## Proof of Concept

An early [PayloadProcessor proof of concept][poc] in the AI Gateway
working group validates that the core functionality is implementable on
top of an existing data plane without invasive changes:

* A `PayloadProcessor` CRD in `ainetworking.x-k8s.io/v0alpha0` expresses
  both in-data-plane and external-service processing.
* A Go controller plugin translates in-data-plane processors to standard
  `TrafficPolicySpec_Transformation` policies; a Rust data plane
  processes them with automatic body buffering, no data plane changes
  required. The controller translates external-service processors to
  policies that the data plane forwards to a configured backend via
  Envoy `ext_proc`.
* The demo does body-based routing across three backends (gpt-4,
  claude, and a default) by extracting the `model` field from the JSON
  request body and projecting it into `X-Gateway-Model-Name` for
  standard `HTTPRoute` header matching.

The PoC is referenced here as evidence that the *what* is achievable,
not as a proposal for the final API shape.

[poc]: https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56

## References

* [WG AI Gateway Payload Processing Proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/7-payload-processing.md)
* [WG AI Gateway Payload Processing Design](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/payload-processing-design.md)
* [Agentgateway PayloadProcessor POC](https://github.com/agentgateway/agentgateway/tree/main/payload-processor-poc)
* [GEP-713: Policy Attachment](https://gateway-api.sigs.k8s.io/geps/gep-713/)
* [GEP-4488: Backend Resource](https://gateway-api.sigs.k8s.io/geps/gep-4488/)
* [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
* [Envoy External Processing](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter)
* [Gateway API Firewall GEP (#3614)](https://github.com/kubernetes-sigs/gateway-api/issues/3614)
* [Predicate Routing (NGINX Gateway Fabric #5737)](https://github.com/nginx/nginx-gateway-fabric/issues/5737)
* [CEL Specification (cel-expr/cel-spec)](https://github.com/cel-expr/cel-spec)
* [Standard CEL Vocabulary](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/57)
* [RFC 6901: JavaScript Object Notation (JSON) Pointer](https://www.rfc-editor.org/rfc/rfc6901)
* [RFC 6902: JavaScript Object Notation (JSON) Patch](https://www.rfc-editor.org/rfc/rfc6902)
* [RFC 9535: JSONPath, Query Expressions for JSON](https://www.rfc-editor.org/rfc/rfc9535)
