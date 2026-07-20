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
headers, paths, and query parameters — but cannot inspect or act on the request
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

While the API surface supports both InProcess and ExtProcess processor types,
this GEP's initial scope is limited to InProcess header and body field mutation
from request body content, which has been validated by a
[proof-of-concept implementation]. The ExtProcess processing protocol
standardization is deferred to a follow-on GEP.

[GEP-713]: https://gateway-api.sigs.k8s.io/geps/gep-713/
[proof-of-concept implementation]: https://github.com/kubernetes-sigs/wg-ai-gateway/pull/56

## Motivation

Gateway API provides a powerful, extensible framework for configuring HTTP
routing in Kubernetes. However, its current processing model is fundamentally
limited to metadata-level operations — headers, paths, query parameters, and
method. There is no standardized mechanism for Gateway API implementations to
inspect or act on the **body** of a request or response. This gap creates
friction in several areas:

### No Body Access in Gateway API

Gateway API's `HTTPRoute` filters (`RequestHeaderModifier`,
`RequestRedirect`, `URLRewrite`, `RequestMirror`, `ExtensionRef`) all operate
on request metadata. None can read or act on the request body. This means
common patterns like "route based on a field in the JSON body" require
implementation-specific extensions with no portability.

### AI Inference Requires Body-Level Decisions

AI inference workloads send model selection, prompt content, and configuration
in the request body (typically JSON). Key decisions — which model to route to,
whether the prompt contains PII or injection attacks, whether to cache the
response — all require reading the body. Today, llm-d has an implementation of
a Body-Based Router (BBR) to extract model names for routing. This is the primary
implementation of the pluggable BBR framework proposed by [Gateway API Inference Extension (GAIE)].
This proposal is in a draft state and the reference implementation is no longer
within the GAIE repo.

### External Processing Varies Per Proxy

Envoy's `ext_proc` filter, NGINX's `mirror` and Lua scripting, and other proxy
mechanisms provide body processing capabilities, but each uses a different
protocol and configuration model. There is no Kubernetes-native abstraction for
"send this request's body to an external service for processing before
routing."

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
* Support **InProcess** processors that use CEL expressions to extract data from
  request bodies and mutate headers and body fields, enabling body-based routing
  without external services.
* Support **ExtProcess** processors that delegate payload processing to external
  gRPC services referenced via `backendRef`, enabling arbitrary processing
  logic (security scanning, PII detection, semantic analysis).
* Provide per-processor **failure modes** (`FailClosed`, `FailOpen`) to enable
  safe composition of security-critical and optimization processors.
* Define **ordered, sequential execution** with short-circuit rejection — if
  any processor rejects, subsequent processors are skipped.
* Support attachment to both `Gateway` (pre-routing, applies to all traffic)
  and `HTTPRoute` (post-routing, applies to matched traffic) via the standard
  policy attachment pattern ([GEP-713]).
* Ensure the API is extensible for future capabilities (response body mutation,
  body rewriting, metadata extraction) without breaking changes.

## Non-Goals

* **Standardizing the ExtProc protocol**: The wire protocol between the gateway
  and external processor services (gRPC service definition, message format,
  streaming semantics) is explicitly deferred to a follow-on GEP. This GEP
  defines the Kubernetes resource API only; implementations may use Envoy
  ext_proc, a custom gRPC protocol, or other mechanisms.
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
* **Develop a CEL standard library**: While CEL is the recommended expression language for InProcess processors, the
  development of a standard library of CEL functions for common payload processing
  tasks (JSON parsing, string manipulation, semantic similarity) is deferred to
  future work. However, for the payload processor resource to be portable and have
  consistent behavior across implementations, CEL standardization is a critical
  dependency that must be addressed in parallel.

[agentgateway's implementation]: https://github.com/agentgateway/agentgateway/pull/1787

## User Stories

### As an AI Platform Engineer

> "I want to route inference requests to the correct model backend based on the
> `model` field in the JSON request body, without modifying my application or
> using implementation-specific extensions. Today I use a custom
> Body-Based Router API and implementation, but I want a portable Gateway API
> solution."

### As a Security Engineer

> "I want to add a processing step that scans inference request bodies for
> prompt injection attacks and PII before they reach the model backend. If the
> scan detects a threat, the request should be rejected with a clear error. If
> the scanning service is unavailable, the request should be rejected
> (fail-closed) for security processors but allowed through (fail-open) for
> non-critical enrichment processors."

### As a Compliance Officer

> "I want to examine both inference requests and responses for personally
> identifiable information so that PII can be blocked, sanitized, or reported.
> I need this to be declarative, auditable, and composable with other
> processing steps."

### As a Developer of Agentic AI Platforms

> "I need to process Model Context Protocol (MCP) request payloads to extract
> tool names and session identifiers for routing decisions. I want to set
> headers based on payload attributes so the gateway can route to the correct
> backend MCP server."

### As a Cluster Administrator

> "I want to add semantic caching to inference requests — detecting repeated
> or semantically similar requests and returning cached results to reduce
> inference costs and improve latency. This requires reading the request body
> to compute similarity, which no current Gateway API resource supports."

### As a Gateway API Implementation Author

> "I want a clear, standardized resource definition for payload processing so
> I can implement it consistently. I need the specification to be unambiguous
> about ordering, failure modes, and the boundary between in-process and
> external processing."

## Proposal

The `PayloadProcessor` resource is a namespace-scoped, policy-attached resource
that declares an ordered list of processors to be applied to HTTP request
and/or response payloads.

### Resource Overview

```
┌─────────────────────────────────────────────────┐
│  PayloadProcessor                               │
│  targetRef: Gateway or HTTPRoute                │
│  phase: PreRouting | PostRouting                 │
│  processors:                                    │
│    ┌─────────────────────────────────────────┐   │
│    │ [0] extract-model (InProcess)           │   │
│    │     CEL: json(request.body).model       │   │
│    │     → Set X-Gateway-Model-Name header   │   │
│    │     failureMode: FailClosed             │   │
│    ├─────────────────────────────────────────┤   │
│    │ [1] scan-pii (ExtProcess)               │   │
│    │     backendRef: pii-scanner:4444        │   │
│    │     failureMode: FailClosed             │   │
│    ├─────────────────────────────────────────┤   │
│    │ [2] enrich-context (ExtProcess)         │   │
│    │     backendRef: context-service:8080    │   │
│    │     failureMode: FailOpen               │   │
│    └─────────────────────────────────────────┘   │
└──────────────────┬──────────────────────────────┘
                   │ targetRef
                   ▼
┌─────────────────────────────────────────────────┐
│  Gateway or HTTPRoute                           │
│  (standard routing continues after processing)  │
└─────────────────────────────────────────────────┘
```

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

  # phase determines when processors execute relative to route selection.
  # PreRouting: before HTTPRoute matching (targets Gateway or ListenerSet)
  # PostRouting: after route selected (targets Gateway, ListenerSet, or HTTPRoute)
  phase: PreRouting

  # processors is an ordered list of processing steps (1-16).
  # Executed sequentially; if any processor rejects, subsequent processors
  # are skipped and the request is rejected.
  processors:
  - name: extract-model             # unique within this resource, 1-63 chars
    type: InProcess                  # InProcess or ExtProcess
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

  - name: pii-scanner
    type: ExtProcess
    failureMode: FailClosed
    timeout: "1s"

    # extProcess: configuration for external processor.
    # Required when type is ExtProcess.
    extProcess:
      backendRef:
        kind: Service
        name: pii-scanner-service
        port: 4444
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

**CEL Context Available:**

TODO: Define a CEL standard library for payload processing with functions like `json()`, `form.decode()`, and `merge()`.

| Variable | Type | Description |
|----------|------|-------------|
| `request.body` | `bytes` | Raw request body (triggers automatic buffering) |
| `request.headers` | `map<string, string>` | Request headers |
| `request.method` | `string` | HTTP method |
| `request.path` | `string` | Request path |
| `json(request.body)` | `map` | Parsed JSON body (convenience function) |

**Header and Body Field Mutation Operations:**

| Operation | Behavior |
|-----------|----------|
| `setHeaders` | Overwrites an existing header or creates a new one. Value is a CEL expression. |
| `removeHeaders` | Removes a header by name. |
| `setBodyFields` | Overwrites or creates a body field addressed by JSONPath. Value is a static value or a CEL expression evaluated over the payload body. |
| `removeBodyFields` | Removes a body field addressed by JSONPath. |

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

### ExtProcess Processors

ExtProcess processors delegate payload processing to an external service
referenced via `backendRef`. The external service receives the request payload
and can signal approval, rejection, or header/body mutations.

**Note:** The wire protocol between the gateway and the ExtProcess service is
**not standardized by this GEP**. Implementations MAY use Envoy's ext_proc
gRPC protocol, a custom protocol, or any other mechanism. A follow-on GEP will
propose a standardized processing protocol.

```yaml
processors:
- name: pii-scanner
  type: ExtProcess
  failureMode: FailClosed
  timeout: "1s"
  extProcess:
    backendRef:
      kind: Service
      name: pii-scanner-service
      port: 4444
```

**ExtProcess Service Requirements:**
* The service MUST be reachable from the gateway data plane.
* Implementations MUST support referencing Kubernetes `Service` resources.
* Implementations MAY support other backend kinds (e.g., `Backend` from
  [GEP-4488]).
* The `timeout` field, if specified, MUST be enforced by the gateway. If the
  external service does not respond within the timeout, the gateway MUST apply
  the processor's `failureMode`.

[GEP-4488]: https://gateway-api.sigs.k8s.io/geps/gep-4488/

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

### Validation

The `PayloadProcessor` CRD uses Kubernetes-native validation mechanisms:

* **Schema validation**: Field types, enums, string lengths, array bounds
  (1-16 processors, 1-63 char names, 1-256 char header names)
* **CEL validation rules** (`x-kubernetes-validations`):
  * Exactly one of `inProcess` or `extProcess` MUST be set per processor
    (enforced by: `has(self.inProcess) != has(self.extProcess)`)
  * `targetRef.kind` MUST be `Gateway` or `ListenerSet` when `phase` is
    `PreRouting`
  * Processor names MUST be unique within the resource

### Status

The `PayloadProcessor` resource reports status following the standard policy
attachment pattern ([GEP-713]):

```yaml
status:
  ancestors:
  - ancestorRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: ai-gateway
    controllerName: example.com/gateway-controller
    conditions:
    - type: Accepted
      status: "True"
      reason: Accepted
      message: "PayloadProcessor accepted by gateway"
    - type: Attached  # XXX: This name is under discussion
      status: "True"
      reason: Attached
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

### Envoy ext_proc

Envoy's [External Processing filter] provides a mature, streaming-capable
protocol for external payload processing. PayloadProcessor's ExtProcess type is
conceptually similar but does not mandate the Envoy ext_proc wire protocol.
Implementations using Envoy MAY map ExtProc processors directly to Envoy
ext_proc filters. The standardization of a common wire protocol is deferred.

[External Processing filter]: https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_proc_filter

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

* At least 2 conformant implementations with production usage
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

### Inline HTTPRoute Filter vs. Separate CRD

Two API shapes were considered for how payload processing is configured:

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

* **Pro**: Familiar filter pattern; processing is visible inline with routing
* **Con**: HTTPRoute rules can already be complex; adding processor
  configuration (potentially 16 processors with CEL expressions, backendRefs,
  failure modes) would make HTTPRoutes unwieldy. Cannot reuse the same
  processor configuration across multiple routes. Cannot target Gateway-level
  (pre-routing) processing.

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

* **Pro**: Reusable across routes; supports Gateway-level attachment;
  consistent with GEP-713 pattern; keeps HTTPRoute focused on routing
* **Con**: Less discoverable from HTTPRoute; requires cross-referencing
  resources

**Decision**: Option B was chosen because:
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

### ExtProc Wire Protocol

> What wire protocol should external processors implement?

Options include extending Envoy's ext_proc v3 protocol, defining a new
Gateway API-specific protocol, or allowing implementation-defined protocols.
This is deferred to a companion GEP. In the interim, implementations MAY
use any protocol.

### Processing Loops

> Can a mutating PayloadProcessor trigger re-evaluation of HTTPRoute matching?

The current design says no — PreRouting processors execute once, mutate
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