---
title: "GEP-5091: PayloadProcessor Resource - Internal Processing"
---

* Issue: [#5091](https://github.com/kubernetes-sigs/gateway-api/issues/5091)
  * Incubated by the [AI Gateway Working Group](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/7-payload-processing.md)
* Status: Provisional

## Summary

This GEP proposes a new `PayloadProcessor` resource that enables declarative,
ordered processing of HTTP request and response **payloads** (headers *and*
body) within the Gateway API framework. Today, Gateway API filters operate on
headers, paths, and query parameters — but do not define mechanisms for acting on the request and response
body. Modern workloads, particularly AI inference, require body-level
processing for routing, security, and compliance decisions.

The `PayloadProcessor` resource attaches to a `Gateway` or `HTTPRoute` via
policy attachment ([GEP-713]) and defines an ordered list of processors. Each
processor is either **InProcess** (CEL expressions evaluated in the data plane
for header and body field mutation based on body content) or **ExtProcess** (an
external gRPC service that receives the payload for arbitrary processing).
Processors execute sequentially with per-processor failure modes, enabling
composable processing pipelines such as "extract model name from body → set
routing header → reject if PII detected."

While the envision API would supports both InProcess and ExtProcess processor types,
this GEP's scope is limited to InProcess header and body field mutation
from request and response body content, which has been validated by a
[proof-of-concept implementation]. The ExtProcess API and protocol
standardization is deferred to a follow-up GEP.

[GEP-713]: https://gateway-api.sigs.k8s.io/geps/gep-713/
[proof-of-concept implementation]: https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56

## Motivation

Gateway API provides a powerful, extensible framework for configuring HTTP
routing in Kubernetes. However, its current processing model is fundamentally
limited to metadata-level operations — headers, paths, query parameters, and
method. There is no standardized mechanism for Gateway API implementations to
inspect or act on the **body** of a request or response. This gap creates
friction in several areas:

### No API Mechanism for Response Access and Modification

Gateway API's `HTTPRoute` filters (`RequestHeaderModifier`,
`RequestRedirect`, `URLRewrite`, `RequestMirror`, `ExtensionRef`, `ExternalAuth`) all operate
on request metadata. They can read the request body but none can read or act on the response body. This means
patterns for response access and modification require
implementation-specific extensions with no portability or are incompatible
with the required protocol (ext_authz).

### AI Inference Requires Body-Level Decisions

AI inference workloads send model selection, prompt content, and configuration
in the request body (typically JSON). Key decisions — which model to route to,
whether the prompt contains PII or injection attacks, whether to cache the
response — all require reading the body. Today, llm-d has an implementation of
a Body-Based Router (BBR) to extract model names for routing. This is the primary
implementation of the pluggable BBR framework proposed by [Gateway API Inference Extension (GAIE)].
This proposal is in a draft state and the reference implementation is no longer
within the GAIE repo.

### No Standard Mechanism for In-Process Payload Actions

A significant class of payload processing consists of deterministic field-level
operations: reading a value from a JSON body, copying it into a header,
removing a field, or setting a field to a fixed value. These operations require
no external state or networking calls, and can be performed locally by
existing data planes. The only portable option is to forward the entire payload to an
external service, which adds a network round trip and requires a separately
deployed, scaled, and secured workload. Equivalent in-process mechanisms exist
across data planes but are not united under a common API and expression language.
For example, Envoy's [`transform` filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/transform_filter) and [`json_to_metadata` filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/json_to_metadata_filter) provide similar functionality, but are not portable across implementations like agentgateway. 

### Composability Gap

Real-world payload processing requires ordered, composable pipelines — for
example, "first extract the model name for routing, then scan for PII, then
check for prompt injection." Current approaches require either monolithic
external services or implementation-specific chaining mechanisms.

[Gateway API Inference Extension (GAIE)]: https://github.com/kubernetes-sigs/gateway-api-inference-extension
[BBR framework Proposed]: https://github.com/kubernetes-sigs/gateway-api-inference-extension/tree/main/docs/proposals/1964-pluggable-bbr-framework
[llm-d]: https://github.com/llm-d/llm-d-inference-payload-processor

## Goals

* Introduce a `PayloadProcessor` resource as a namespace-scoped, policy-attached
  resource for declaring ordered payload processing steps on HTTP requests and
  responses.
* Support **InProcess** processors that support a standard set of CEL expressions and functions to extract data from
  request and response bodies and mutate headers and body fields, enabling body-based routing
  without external services.
* Provide per-processor **failure modes** (`FailClosed`, `FailOpen`) to enable
  safe composition of security-critical and optimization processors.
* Define **ordered, sequential execution** with short-circuit rejection — if
  any processor rejects, subsequent processors are skipped.
* Support attachment to both `Gateway` (pre-routing, applies to all traffic)
  and `HTTPRoute` (post-routing, applies to matched traffic) via the standard
  policy attachment pattern ([GEP-713]).
* Ensure the API is extensible for future capabilities (`ExtProcess`) without breaking changes.

## Non-Goals

* **Replacing existing HTTPRoute filters**: PayloadProcessor complements, not
  replaces, existing filters. Header-only operations should continue to use
  `RequestHeaderModifier` and similar filters.
* **Streaming body processing**: This GEP requires full body buffering for
  InProcess processors. Streaming/chunked body processing is deferred to
  future work.
* **External processing distinction between request and response**: Specifying when
  or how external processing is invoked is out of scope for this GEP and is
  dependent upon a standardized ExtProc protocol. Pre and post-routing phases are
  defined, but the API does not currently distinguish between request and response
  processing for ExtProc. [agentgateway's implementation] can be referenced as prior
  art.
* **TCP/UDP/TLS payload processing**: These protocols lack the structured,
  inspectable body and header-matching model this resource depends on. Support
  for GRPCRoute is possible for future work, but is not in scope for this GEP.

## User Stories

A processor belongs in **InProcess** when it is deterministic, requires no
external state, no network call, and no model inference, and completes in
bounded CPU time. Reading a field, rewriting a field, applying a default,
removing a field, and matching a fixed pattern all satisfy this test. These
operations are strong candidates for in-process execution for three reasons
beyond latency:

* **Buffering can be derived from the expression.** An implementation can
  determine from the processor's expressions whether the body is needed at all,
  and buffer only when it is. An external processor must fix its buffering
  behavior at configuration time, before any request is seen.
* **Sensitive data does not leave the proxy.** Prompts, credentials, and
  regulated fields are not serialized across a pod boundary to a second
  workload.
* **It runs before routing without an added hop.** Pre-routing processing sits
  ahead of every other decision in the request path.

A processor belongs in **ExtProcess** when it needs something the data plane
fundamentally does not have: a trained model, external state, or a large and
frequently updated ruleset.

Note that this boundary does not split cleanly by *problem domain*. PII
handling appears on both sides — deterministic pattern-based detection and
field redaction are in-process, while classifier-based detection is external.
The test is the nature of the computation, not the category of the use case.
The exact boundary is implementation-defined, but the following stories
illustrate the distinction.

### InProcess Use Cases

#### As an AI Platform Engineer

> "I want to route inference requests to the correct model backend based on the
> `model` field in the JSON request body, without modifying my application or
> using implementation-specific extensions. Today I use a custom
> Body-Based Router API and implementation, but I want a portable Gateway API
> solution."

#### As a Platform Engineer

> "I want to normalize inference requests before they reach a backend. I want to
> force usage accounting, apply defaults for fields the client omitted,
> and reject requests that are missing a required field. Today this requires
> either changing every client or deploying an external service to edit two
> keys in a JSON document."

#### As a Compliance Officer

> "I want to strip known sensitive fields from request and response bodies, and
> reject payloads matching a defined pattern, so that regulated data never
> reaches a backend or a client. Because the data is sensitive, I specifically
> do not want it forwarded to an additional service in order to be inspected. I
> need this to be declarative, auditable, and composable with other processing
> steps."

#### As a Developer of Agentic AI Platforms

> "I need to process Model Context Protocol (MCP) request payloads to extract
> tool names and session identifiers for routing decisions. I want to set
> headers based on payload attributes so the gateway can route to the correct
> backend MCP server."

#### As an API Owner

> "I want to enrich requests with context derived from the payload and the
> verified caller identity — request identifiers for tracing, tenant headers
> for downstream attribution — without adding a network hop to a request that
> is otherwise served entirely from cache."

### ExtProcess Use Cases

The following stories motivate the `ExtProcess` extension point. The
`ExtProcess` API and its wire protocol are deferred to a follow-up GEP, but the
resource is shaped so they can be added without breaking changes.

#### As a Security Engineer

> "I want to add a processing step that classifies inference request bodies for
> prompt injection attacks before they reach the model backend. Detection
> requires a trained model, so it cannot run in the data plane. If the scan
> detects a threat, the request should be rejected with a clear error. If the
> scanning service is unavailable, the request should be rejected (fail-closed)
> for security processors but allowed through (fail-open) for non-critical
> enrichment processors."

#### As a Compliance Officer

> "I want to examine inference responses for personally identifiable
> information that cannot be expressed as a fixed pattern, so that it can be
> blocked, sanitized, or reported. This requires a detection model, and I accept
> the additional hop in exchange for the detection quality."

#### As a Cluster Administrator

> "I want to add semantic caching to inference requests — detecting repeated
> or semantically similar requests and returning cached results to reduce
> inference costs and improve latency. This requires computing embeddings and
> querying a vector store, neither of which belongs in the data plane."

### Applies to Both

#### As a Gateway API Implementation Author

> "I want a clear, standardized resource definition for payload processing so
> I can implement it consistently. I need the specification to be unambiguous
> about ordering, failure modes, and the boundary between in-process and
> external processing."

## Proposal

The `PayloadProcessor` resource is a namespace-scoped, policy-attached resource
that declares an ordered list of processors to be applied to HTTP request
and/or response payloads.

### API Definition

**NOTE**: This is an *early draft* of the API definition. Primarily defined
here for discussion.

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: PayloadProcessor
metadata:
  name: example-processor
  namespace: default
spec:
  # targetRef identifies the Gateway or HTTPRoute this policy applies to.
  # Follows the standard policy attachment pattern (GEP-713).
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway          # or HTTPRoute
    name: my-gateway
    sectionName: http # optional, targets a specific Listener or ListenerSet (or HTTPRouteRule name)

  # phase determines when processors execute relative to route selection.
  # PreRouting: before HTTPRoute matching (targets Gateway or ListenerSet)
  # PostRouting: after route selected (targets Gateway, ListenerSet, or HTTPRoute)
  phase: PreRouting

  # processors is an ordered list of processing steps (1-16).
  # Executed sequentially; if any processor rejects, subsequent processors
  # are skipped and the request is rejected.
  processors:
  - name: extract-model             # unique within this resource, 1-63 chars
    type: InProcess                  # InProcess or ExtProcess (deferred to follow-up GEP)
    failureMode: FailClosed          # FailClosed (default) or FailOpen
    timeout: "500ms"                 # optional per-processor timeout

    # inProcess: configuration for in-process (data plane) processing.
    # Required when type is InProcess.
    inProcess:
      request:
        # setHeaders: overwrite or create headers with CEL expression values
        setHeaders:
        - name: X-Gateway-Model-Name
          value: 'json(request.body).model'   # CEL expression
        - name: X-Gateway-Custom-Header
          value: '"my-custom-value"'          # string literal interpreted by CEL
        # removeHeaders: remove headers by name
        removeHeaders: []
        # setBodyFields: overwrite or create body fields (JSONPath) with values
        setBodyFields:
        - name: '$.stream'                    # JSONPath
          value: 'true'
        - name: '$.stream_options'            # JSONPath
          value: '{"include_usage": true}'
        # removeBodyFields: remove body fields by name (JSONPath)
        removeBodyFields:
        - name: '$.user_email'                # JSONPath
```

### Phase Model

PayloadProcessor defines two processing phases that determine when processors
execute relative to HTTPRoute matching:

| Phase | When | Allowed targetRef Kinds | Use Cases |
|-------|------|------------------------|-----------|
| `PreRouting` | Before HTTPRoute matching | `Gateway`, `ListenerSet` | Body-based routing (extract field → set header → HTTPRoute matches on header), request validation |
| `PostRouting` | After route selected, before backend dispatch | `Gateway`, `ListenerSet`, `HTTPRoute` | PII scanning, content enrichment |

**PreRouting** processors execute on all traffic entering the Gateway (or
listener), before any HTTPRoute rules are evaluated. This enables the core
body-based routing pattern: extract a value from the body, set it as a header,
and let standard HTTPRoute header matching select the backend.

**PostRouting** processors execute after a route has been selected. They can
perform processing specific to the matched route, such as scanning request
bodies for PII before forwarding to a particular backend.

```
Client Request
    │
    ▼
┌──────────────────────┐
│  PreRouting Phase    │ ◄── PayloadProcessor (targetRef: Gateway)
│  InProcess/ExtProc   │     Mutate headers/body from content
└──────────┬───────────┘
           │ (headers mutated)
           ▼
┌──────────────────────┐
│  HTTPRoute Matching  │ ◄── Standard header/path/method matching
└──────────┬───────────┘
           │ (route selected)
           ▼
┌──────────────────────┐
│  PostRouting Phase   │ ◄── PayloadProcessor (targetRef: HTTPRoute)
│  InProcess/ExtProc   │     PII scanning, enrichment, etc.
└──────────┬───────────┘
           │
           ▼
┌──────────────────────┐
│  Backend             │
└──────────────────────┘
```

### InProcess Processors

InProcess processors run within the gateway data plane and use CEL expressions
to extract data from the request body and mutate request headers and body
fields. This is the primary mechanism for body-based routing and lightweight
request transformation.

**CEL Standard Library:**

Multiple GEPs have called out the need for a standard set of CEL attributes (ex. Telemetry). This motivates the creation of a set of supported attributes for Gateway API. The [OTEL semantic conventions](https://opentelemetry.io/docs/specs/semconv/registry/attributes/http/) has laid a strong foundation for a dictionary of attributes. Defining the standard attributes is not in scope for the provisional stage of this GEP, but the following attributes are proposed for discussion:

| Variable | Type | Description |
|----------|------|-------------|
| `request.body` | `bytes` | Raw request body (triggers automatic buffering) |
| `request.headers` | `map<string, string>` | Request headers |
| `request.method` | `string` | HTTP method |
| `request.path` | `string` | Request path |
| `json(request.body)` | `map` | Parsed JSON body (convenience function) |

**Body Buffering:** When any CEL expression references `request.body`, or a
processor sets or removes body fields, the gateway implementation MUST buffer
the entire request body before evaluating expressions. Implementations SHOULD
define a maximum buffer size (recommended default: 2 MiB) and MUST reject
requests exceeding the buffer limit when `failureMode` is `FailClosed`.

**Example — Body-Based Routing:**\

**NOTE**: This is an *early draft* of the API definition. Primarily defined
here for discussion.

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: PayloadProcessor
metadata:
  name: model-header-setter
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: ai-gateway
    sectionName: http
  phase: PreRouting
  processors:
  - name: extract-model
    type: InProcess
    failureMode: FailClosed
    inProcess:
      request:
        setHeaders:
        - name: X-Gateway-Model-Name
          value: 'json(request.body).model'
---
# HTTPRoute matches on the header set by PayloadProcessor
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: gpt4-route
spec:
  parentRefs:
  - name: ai-gateway
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /v1/chat/completions
      headers:
      - name: X-Gateway-Model-Name
        value: gpt-4
    backendRefs:
    - name: gpt4-backend
      port: 8080
```

### Failure Modes

Each processor declares its own failure mode, enabling fine-grained control
over behavior when processing fails:

| Mode | Behavior | Use Case |
|------|----------|----------|
| `FailClosed` (default) | Reject the request if the processor errors or times out | Security processors (PII detection, prompt injection scanning) |
| `FailOpen` | Skip the processor and continue if it errors or times out | Optimization processors (caching, enrichment, analytics) |

Failure modes apply to:
* CEL expression evaluation errors (InProcess)
* Body buffering failures (body too large, malformed)
* External service timeouts or connection failures (ExtProcess)
* External service returning an error response (ExtProcess)

### Ordering and Execution Semantics

Processors within a `PayloadProcessor` resource execute **sequentially** in
array order. This provides deterministic, predictable behavior:

1. Processor `[0]` executes first.
2. If processor `[0]` **rejects** the request, processing stops immediately.
   Subsequent processors are not invoked.
3. If processor `[0]` **succeeds** (or fails with `FailOpen`), processor `[1]`
   executes with the (potentially mutated) request.
4. This continues until all processors have executed or one rejects.

**Multiple PayloadProcessor Resources:** When multiple `PayloadProcessor`
resources target the same Gateway or HTTPRoute, implementations MUST apply them
deterministically. The recommended ordering is by resource creation timestamp
(oldest first), consistent with Gateway API policy attachment precedence.

**Interaction with HTTPRoute Filters:** Processors execute in their declared
phase (PreRouting or PostRouting). Standard HTTPRoute filters execute at their
normal point in the request lifecycle. The relative ordering is:

```
PreRouting PayloadProcessors → HTTPRoute Matching → HTTPRoute Filters → PostRouting PayloadProcessors → Backend
```

## Conformance Tiers

The PayloadProcessor resource is designed with a clear separation between Core
and Extended features:

| Feature | Level | Description |
|---------|-------|-------------|
| InProcess header mutation (setHeaders/removeHeaders) | Core | CEL expressions extract body fields and set/remove headers |
| InProcess body field mutation (setBodyFields/removeBodyFields) | Extended | CEL/JSONPath expressions set or remove request body fields |
| PreRouting phase | Core | Processors execute before HTTPRoute matching |
| `FailClosed` / `FailOpen` per processor | Core | Per-processor failure mode selection |
| Sequential processor ordering | Core | Deterministic array-order execution with short-circuit rejection |
| Policy attachment to Gateway | Core | `targetRef` to Gateway resource |
| ExtProcess with `backendRef` | Extended | External gRPC service for arbitrary processing |
| PostRouting phase | Extended | Processors execute after route selection |
| Policy attachment to HTTPRoute | Extended | `targetRef` to HTTPRoute resource |
| Per-processor timeout | Extended | Timeout enforcement for individual processors |

## Relationship to Existing Concepts

### Gateway API Inference Extension (GAIE) Body-Based Router

GAIE implements a Body-Based Router (BBR) that extracts the model name from
inference request bodies to select the appropriate `InferencePool`. The
`PayloadProcessor` InProcess type can implement the same pattern in a
portable, reusable way:

* **BBR**: Implementation-specific, tightly coupled to GAIE's model routing
* **PayloadProcessor**: Generic, reusable for any body field extraction and
  header-based routing

A future proposal may explore re-implementing BBR using PayloadProcessor as
the underlying mechanism, providing consistency and reducing implementation
complexity. However, this GEP does not propose deprecating or replacing BBR.

### Gateway API Firewall GEP

The [Firewall GEP] ([#3614]) proposes firewall-like filtering capabilities for
Gateway API. PayloadProcessor and Firewall have complementary scopes:

* **Firewall**: Network-level and header-level security rules
* **PayloadProcessor**: Body-level processing, transformation, and routing

A PayloadProcessor with an ExtProc service could implement WAF-like body
scanning, while Firewall handles metadata-level rules.

[Firewall GEP]: https://github.com/kubernetes-sigs/gateway-api/issues/3614
[#3614]: https://github.com/kubernetes-sigs/gateway-api/issues/3614

### HTTPRoute Filters

PayloadProcessor is designed to coexist with, not replace, existing HTTPRoute
filters. The key distinction is **body access**: filters operate on metadata;
processors can operate on the full payload. The execution model places
processors in distinct phases (PreRouting/PostRouting) that bracket the
standard filter execution point.

## Graduation Criteria

This GEP follows the standard [Gateway API graduation criteria]. The following
are additional criteria specific to this GEP:

[Gateway API graduation criteria]: https://gateway-api.sigs.k8s.io/concepts/versioning/#graduation-criteria

### Implementable

* PayloadProcessor CRD with full schema validation
* Documentation and examples for InProcess body-based routing
* CEL expression specification for body access

### Experimental

* Reference implementation in at least one Gateway API implementation
  (agentgateway POC serves as initial validation)
* Basic conformance tests for InProcess header mutation from body
* At least one ExtProcess implementation demonstrating external processing

### Standard

* At least 3 conformant implementations with production usage
* Comprehensive conformance test suite covering Core and Extended features
* ExtProc wire protocol standardized in a companion GEP
* Documentation of body buffering limits and performance characteristics

## Alternatives Considered

### CEL vs. Other Expression Languages for Inline Body Processing

The choice of CEL (Common Expression Language) for InProcess body extraction
is a significant design decision. We evaluated several alternatives:

#### CEL (Recommended)

**Strengths:**
* **Kubernetes-native**: CEL is the standard expression language in Kubernetes
  (stable for CRD validation since v1.29, used in ValidatingAdmissionPolicy,
  Gateway API Inference Extension)
* **Type-safe and sandboxed**: No arbitrary code execution, bounded evaluation
  cost, no filesystem or network access
* **Proven for body processing**: The agentgateway project uses CEL expressions
  like `json(request.body).model` in production proxy data plane code with
  automatic body buffering
* **Extensible**: Custom functions (`json()`, `form.decode()`, `merge()`) can
  be added without changing the language
* **Performance**: Expressions are compiled at policy creation time, not
  per-request; adequate for data plane execution

**Limitations:**
* **Requires full body buffering**: CEL cannot process streaming bodies;
  entire body must be in memory before evaluation
* **Buffer size limits**: Recommended 2 MiB default; payloads exceeding this
  limit cannot be processed in-process
* **Complexity cost**: Large documents or deeply nested expressions may exceed
  Kubernetes CEL cost budgets
* **Binary data**: Non-UTF-8 binary payloads require base64 encoding/decoding
* **Standardization**: No enforcement of a consistent CEL standard library
  across implementations may lead to portability issues

#### JSONPath / JMESPath

* **Strengths**: Simple syntax for field extraction; familiar to many users
* **Weaknesses**: No transformation capability (read-only); no type safety;
  not a Kubernetes-native standard; JMESPath adds an external dependency
* **Verdict**: Too limited — PayloadProcessor needs transformation (header
  value construction from body fields), not just extraction

#### Rego (OPA)

* **Strengths**: Powerful policy language; well-suited for security decisions
* **Weaknesses**: Heavier runtime; different syntax from Kubernetes CEL;
  requires OPA deployment; not Kubernetes-native
* **Verdict**: Over-scoped — Rego is better suited for complex policy
  decisions via ExtProc, not inline data plane expressions

#### Lua / WASM

* **Strengths**: Full programming capability; WASM provides sandboxing
* **Weaknesses**: Arbitrary code execution risks (Lua); runtime overhead;
  not Kubernetes-native; poor observability
* **Verdict**: Too powerful and too risky for inline expressions; better
  suited for ExtProc implementations

### Body Field Addressing for Body Mutation

The expression language question above concerns how mutation *values* are
computed. A separate decision is how body mutations are **addressed**: whether
the API identifies individual fields to change, or replaces the body wholesale.
This choice is independent of the selection of CEL for value expressions.

#### JSONPath (current draft)

* **Strengths**: Familiar from `kubectl -o jsonpath`; compact syntax for nested
  access; wildcards and filters allow one expression to select many nodes.
* **Weaknesses**: [RFC 9535] specifies JSONPath as a *query* language that
  returns a nodelist. It does not define assignment semantics. `$.messages[*]
  .content` is a valid query with no defined write behavior, and the results of
  zero-match and multi-match queries would be implementation-defined. There is
  no defined escaping for keys that themselves contain `.`, and no notation for
  appending to an array.
* **Verdict**: Selection semantics exceed what field mutation requires, and the
  resulting ambiguity would have to be specified by this GEP rather than
  inherited from the RFC.

#### JSON Pointer

* **Strengths**: [RFC 6901] identifies exactly one location in a document,
  matching the semantics of `setBodyFields` and `removeBodyFields`. It defines
  escaping (`~0` for `~`, `~1` for `/`) so keys containing separators are
  addressable, and defines the `-` token for the position after the last array
  element, giving array append a standard notation. Kubernetes users already
  encounter the syntax through `kubectl patch --type=json`, which takes
  [RFC 6902] JSON Patch documents whose `path` values are JSON Pointers.
* **Weaknesses**: No wildcard or filter selection, so bulk operations such as
  "redact the content of every message" cannot be expressed as a single entry.
  Visually less familiar than JSONPath.
* **Verdict**: Matches the single-location mutation model without inventing
  semantics. The absence of wildcards limits only the restructuring cases,
  which a whole-body expression serves better.

#### Whole-Body CEL Expression

* **Strengths**: Maximally expressive — parse, transform, and re-serialize the
  document in a single expression, for example
  `toJson(json(request.body).filterKeys(k, !k.startsWith("x_")).merge({...}))`.
  Supports restructuring, conditional logic, and bulk operations that field
  addressing cannot express. This is the mechanism used by the
  [proof-of-concept implementation].
* **Weaknesses**: Requires a full parse and re-serialize even to change one
  key, which does not guarantee round-trip fidelity for key order, number
  formatting, or unknown extension fields. The set of affected fields is not
  statically visible in the resource, which weakens auditability and makes
  per-operation conformance testing harder.
* **Verdict**: Necessary for restructuring use cases. A candidate for a
  distinct field (for example `setBody`) that can be added later without
  breaking a field-level API.

Whether field-level addressing and whole-body replacement should both exist,
and how they would be ordered relative to each other, is unresolved. See
[Header and Body Modification Order](#header-and-body-modification-order).

[RFC 9535]: https://www.rfc-editor.org/rfc/rfc9535
[RFC 6901]: https://www.rfc-editor.org/rfc/rfc6901
[RFC 6902]: https://www.rfc-editor.org/rfc/rfc6902

### Inline HTTPRoute Filter vs. Separate CRD

The following API shapes were considered for how payload processing is configured.
The final design design decision is out of scope for the provisional stage of this GEP.

#### Option A: Inline HTTPRoute Filter

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
spec:
  rules:
  - filters:
    - type: PayloadProcessing
      payloadProcessing:
        processors:
        - name: extract-model
          ...
```

* **Pros**: Familiar filter pattern; processing is visible inline with routing
* **Cons**: HTTPRoute rules can already be complex; adding processor
  configuration (potentially 16 processors with CEL expressions, backendRefs,
  failure modes) would make HTTPRoutes unwieldy. Cannot reuse the same
  processor configuration across multiple routes. Cannot target Gateway-level
  (pre-routing) processing introducing the need to reevaluate routes

#### Option B: Separate CRD with Policy Attachment (Chosen)

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: PayloadProcessor
metadata:
  name: model-extractor
spec:
  targetRef:
    kind: Gateway
    name: ai-gateway
  processors:
  - name: extract-model
    ...
```

* **Pros**: Reusable across routes; supports Gateway-level attachment;
  consistent with GEP-713 pattern; keeps HTTPRoute focused on routing
* **Cons**: Less discoverable from HTTPRoute; requires cross-referencing
  resources

#### Option C: Rule-level HTTPRoute Filters with PayloadProcessorRef and Gateway PolicyRef

This option allows a PayloadProcessor to be defined as a separate resource and referenced from an HTTPRoute
filter. This would allow for reusability while keeping the configuration inline and ordered with other
HTTPRoutefilters. This configuration would support the post-routing phase. However, it introduces complexity in
the APIand doesn't capture the pre-routing use case, which is a primary motivation for this GEP. To address this
gap, we could allow the PayloadProcessor resource to target a Gateway, but this doesn't fix the issue of ordering in the case other policy resources target the gateway.

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: HTTPRoute
metadata:
  name: example-httproute
spec:
  parentRefs:
  - name: example-gateway
  hostnames:
  - "www.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /query
    backendRefs:
    - name: example-svc
      port: 8080
    filters:
    - type: PayloadProcessingRef
      payloadProcessingRef:
        name: prompt-injection-protection
---
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: PayloadProcessingPipeline
metadata:
  name: prompt-injection-protection
spec:
  phase: PostRouting
  processors:
  - name: pii-scanner
    type: InProcess
    failureMode: FailClosed
    timeout: "1s"
    inProcess:
      removeBodyFields:
      - name: '$.user_email'
      - name: '$.metadata.customer_id'
```

**Summary of considerations**
1. Pre-routing processing (the primary use case) requires Gateway-level
   attachment, which inline filters cannot express
2. Processing pipelines can be complex and benefit from dedicated resources
3. Reusability across routes reduces configuration duplication
4. Consistency with the policy attachment pattern (GEP-713) used by other
   Gateway API extensions

### Single PayloadProcessor vs. Pipeline Resource

We considered whether a single resource should define individual processors
or entire pipelines:

* **PayloadProcessor with embedded processor list (Chosen)**: A single
  resource contains an ordered list of processors. Simple, self-contained,
  and sufficient for the common case.
* **Separate PayloadProcessorPipeline**: A separate resource defining a
  reusable pipeline of processor references. More flexible but adds
  complexity and indirection. Can be introduced in a future GEP if needed.

## Open Questions

The following questions are under active discussion and will be resolved
before this GEP moves to Experimental:

### Processing Loops

> Can a mutating PayloadProcessor trigger re-evaluation of HTTPRoute matching?

PreRouting processors execute once, mutate
headers, and then HTTPRoute matching occurs on the mutated headers. There is
no re-entry. This avoids infinite loops but limits some advanced use cases.
PostRouting processors can mutate headers, but those mutations do not affect
the routing decision that has already been made.

### Gateway-Level and HTTPRoute-Level Co-existence

> How do PayloadProcessors targeting a Gateway interact with those targeting
> an HTTPRoute?

The current proposal applies them in phase order: Gateway-targeted
PreRouting processors execute first, then HTTPRoute matching, then
HTTPRoute-targeted PostRouting processors. If both target the same phase,
Gateway-level processors execute before HTTPRoute-level processors. If two
PayloadProcessors target the same phase with the same target reference, the
newer resource is ignored and the older resource is used; the resulting
conflict is reflected in the status of the newer resource.

### CEL Cost Budgets

> Should there be a maximum CEL expression cost for InProcess processors?

Kubernetes enforces cost budgets for CEL in admission webhooks. A similar
mechanism may be needed for data plane CEL evaluation, but the cost model
differs (per-request vs. per-admission). Implementations SHOULD document
their CEL cost limits.

### Body Buffer Size Configuration

> Should the maximum body buffer size be configurable per-PayloadProcessor?

The POC uses a gateway-wide default (2 MiB). Per-processor configuration
adds flexibility but also complexity. The initial proposal defers this to
implementation-defined configuration.

### Parallel Processing

> Should multiple processors be able to execute in parallel?

The initial design executes processors sequentially in array order. The ability
to specify and process multiple payload processors in parallel (both InProcess
and ExtProcess) adds complexity but should be considered for performance in a
future phase.

### Header and Body Modification Order

> In what order are header and body modifications applied within a processor?

There is currently no defined order for when header and body modifications occur
relative to each other. This could lead to unexpected behavior when the order
matters for the processing logic and needs to be specified.

### InProcess and ExtProcess Ordering

> Should ExtProcess processors always run before InProcess processors?

ExtProcess processors are considered the heavy lifters of processing, while
InProcess processors are more lightweight and suited for final formatting and
transformation tasks. One option under discussion is to always process
ExtProcess processors before InProcess processors, independent of array order.

### Request and Response Handling

> How should buffering be controlled for responses?

Buffering a response can negatively impact time to first token. When a processor
does not require buffering, the response can be processed in chunks. The current
API does not provide a way for users to control this behavior.

### Injecting Confidential Data

> How should confidential data be injected into payloads or headers?

The current design does not provide a mechanism for injecting confidential data
(e.g. API keys, secrets) into request or response payloads and/or headers. One
option is a per-processor `secretRef` field naming the secrets to inject.
Another is a set of confidential-data references, defined once and accessible to
all processors, that each processor references for injection via a predefined
key (e.g. `credential.<cred name>.<cred field>`). The exact mechanism for
securely injecting confidential data will be addressed in a future phase.

### ExtProc Buffering

> Are we okay with buffering being the only supported mode for extProc?

https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56/changes/BASE..3cb22badd015dd720f300855f5cdcd290d06b0a9#r3306621885

## Proof of Concept

The [agentgateway PayloadProcessor POC](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56) validates the core design:

* **CRD**: `PayloadProcessor` in `ainetworking.x-k8s.io/v0alpha0` with
  InProcess and ExtProc schema
* **Implementation**: Go controller plugin translates `InProcess` processors
  to standard `TrafficPolicySpec_Transformation` policies; Rust data plane
  processes them with automatic body buffering — no data plane changes required.
  The controller also translates `ExtProcess` processors to policies which the
  data plane translates to Envoy `ext_proc` calls to the specified backendRef.
* **Demo**: Body-based routing with three backends (gpt-4, claude, default)
  using `json(request.body).model` CEL expression (for `InProcess`) or an external
  server (for `ExtProcess`) to extract model name and
  set `X-Gateway-Model-Name` header for HTTPRoute matching

```
# Route to gpt-4 backend
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "gpt-4", "messages": [{"role": "user", "content": "hello"}]}'

# Route to claude backend
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model": "claude", "messages": [{"role": "user", "content": "hello"}]}'
```

[PayloadProcessor Prototype]: https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56
## References

* [WG AI Gateway Payload Processing Proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/7-payload-processing.md)
* [WG AI Gateway Payload Processing Design](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/payload-processing-design.md)
* [Agentgateway PayloadProcessor POC](https://github.com/agentgateway/agentgateway/tree/main/payload-processor-poc)
* [GEP-713: Policy Attachment](https://gateway-api.sigs.k8s.io/geps/gep-713/)
* [GEP-4488: Backend Resource](https://gateway-api.sigs.k8s.io/geps/gep-4488/)
* [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension)
* [Envoy External Processing](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter)
* [Gateway API Firewall GEP (#3614)](https://github.com/kubernetes-sigs/gateway-api/issues/3614)
* [CEL Specification](https://github.com/google/cel-spec)
* [Standard CEL Vocabulary](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/57)
* [RFC 6901: JavaScript Object Notation (JSON) Pointer](https://www.rfc-editor.org/rfc/rfc6901)
* [RFC 6902: JavaScript Object Notation (JSON) Patch](https://www.rfc-editor.org/rfc/rfc6902)
* [RFC 9535: JSONPath — Query Expressions for JSON](https://www.rfc-editor.org/rfc/rfc9535)