---
title: "GEP-5250: Multi Stage Routing"
---

* Issue: [#5250](https://github.com/kubernetes-sigs/gateway-api/issues/5250)
  * Related to [GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/issues/5091)
  * Related to [#5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194)
* Status: Provisional

## TLDR

This GEP was incubated in [WG AI-Gateway](https://github.com/kubernetes-sigs/wg-ai-gateway), and proposes to define standards in Gateway API for expressing a single logical client 
request that is fulfilled by multiple, ordered backend interactions, rather than by a single backend interaction.

The concrete API shape is deliberately left open here, and will be proposed once the problem statement is agreed.

This GEP consumes the payload-processing work rather than competing with it: deriving one stage's input from an earlier stage's response is expressed
with whatever [GEP-5091] lands on, in the phase established by [GEP-5224].

## Motivation

Routing in Kubernetes today assumes that a client request maps to one backend interaction:

```
Client Request → Route → Backend
```

This assumption has served traditional applications well, and it is deeply embedded in how routes, policies and observability are defined. 
A growing class of workloads breaks it. For these workloads, what the client sees as one request is fulfilled by a sequence of distinct backend interactions, 
each of which is independently routable and may need its own networking configuration:

* **Inference disaggregation**: a single inference request is served by separate prefill and decode backends, and increasingly by separate encode, prefill 
  and decode backends, each is a distinct pool of endpoints with distinct routing and scaling characteristics, and selecting the right endpoint within a pool 
  is itself a load-aware decision made per request. The interactions are also coupled: state established while serving one is consumed by the next.

* **Representation and protocol transformation**: an inference request arrives over HTTP and is first sent to a tokenizer service, 
  after which later interactions operate on a tokens-in/tokens-out representation, potentially over a different protocol like gRPC.

* **Service composition and fan-out**: a single client request requires calls to several independent backends, possibly in parallel, 
  before a response can be constructed. This case is not AI specific, and predates the AI workloads above.
  Fan-out is deferred to a follow-up proposal, see [Open Points](#open-points).

Because there is no way to express this in routing configuration, every project that needs it builds the capability outside the gateway
(see [Prior Art](#prior-art)). This has three consequences, and together they are the motivation for this proposal:

**It fragments the ecosystem.** Projects solving the same problem arrive at incompatible architectures. Prefill/decode disaggregation is 
the clearest example: each project (like llm-d, AIBrix or Dynamo) implements it differently, with no shared configuration surface between them. 
A platform team that adopts one of these cannot carry its routing configuration to another, and a gateway implementation cannot support the 
pattern generically because there is nothing generic to support. The pattern is common; only the expression of it is bespoke.

**It duplicates capabilities the gateway already provides.** Once orchestration lives outside the gateway, the sub-requests it issues 
fall outside the gateway's reach — so retries, backoff, timeouts, fail-open/close behavior, connection management and traffic policy 
all have to be reimplemented alongside it, per project. This is a substantial amount of subtle, load-bearing behavior to rebuild, 
and rebuilding it is not the point of any of these projects.

**It leaves operators without a coherent view of a request.** The stages of a logical request are where the interesting behavior 
lives: which backend was slow, which interaction failed, where the request was retried, where it ended. When those interactions are 
invisible to the gateway, they cannot be uniformly observed, and an operator is left correlating them by hand across whatever each 
project happens to expose.

Kubernetes has a strong tradition of standardizing the networking substrate so that workloads do not each solve it again. Multi-stage request 
processing is a gap in that substrate, and it is being filled today by duplicated, non-portable effort.

## Definitions

* **Logical request**: the request as the client issues and perceives it. One logical request, one client-visible response.

* **Stage**: one of the backend interactions that together fulfill a logical request. A stage is independently routable and may carry its 
  own networking configuration.

* **Orchestration**: the per-request decisions taken while a logical request is being fulfilled: which stage runs next, whether a stage is skipped, 
  how the input to a stage is derived from what came before, and whether the exchange should end early. Configuration declares the rules for these 
  decisions; the decisions themselves are made per request, against that request's data.

> **Note**: these definitions are intended to describe the shape of the problem, not to prescribe a mechanism. In particular, "stage" is not meant 
> to imply any existing filter, chain, or extension mechanism in any specific implementation.

## Goals

* Provide a declarative way to express that a logical request is fulfilled by multiple backend interactions, including which stages exist, in what order
  they may execute, and what each targets.

* Allow each stage to carry its own networking configuration, so that stages with different performance and failure characteristics can be configured differently.

* Establish a clear separation between the rules declared in configuration — the boundaries of the workflow, and the conditions under which each stage runs —
  and the evaluation of those rules against an individual request as it is served.

* Reuse existing Kubernetes routing semantics for each stage, so that capabilities such as retries, timeouts and failure behavior do not have to be reimplemented.

* Remain workload agnostic. The abstraction must be usable for service composition and fan-out, and must not require any AI-specific understanding to implement.

* Make the stages of a logical request observable as related parts of one exchange.

## Non-Goals

* **Becoming a workflow engine.** Stages are forward-only and each runs at most once per logical request: no loops, no re-entry, and no branching DAG.
  Configuration declares the boundaries of the composition; the per-request decisions taken within those boundaries may skip a stage or end the exchange
  early, but never go back.

* **Encoding application semantics.** Tokens, KV cache, model state, scheduling algorithms, model placement and model-serving protocols are all out of
  scope.

* **Defining a new vocabulary for payloads.** Inspecting, rewriting or rejecting request and response content remains the concern of payload processing
  ([GEP-5091]) and of the phase established by [GEP-5224].

* **Replacing configuration that belongs in a single route.** If a piece of networking configuration can be expressed in one route, it should stay there.
  Multi-stage routing is not a second way to configure what a route already configures.

* **Replacing endpoint selection.** Choosing which endpoint within a pool serves a given interaction stays where it is today. This GEP is concerned with
  how many interactions a logical request consists of, not with which endpoint serves any one of them.

* **Defining a Gateway-to-orchestrator protocol.** The composition is expressed declaratively in configuration. This GEP does not introduce an external
  component that the gateway consults per request, nor a protocol for communicating with one. External callouts are a separate piece of the
  payload-processing split and are out of scope here.

## User Stories

* As an **application developer** consuming an AI Gateway:

  * I want to send one request and receive one response, without knowing how many backend interactions were required to serve it, so that the serving
    architecture can evolve without changing my client.

* As an **inference platform provider**:

  * I want to express that requests to a model are served by a prefill backend followed by a decode backend, so that I can adopt disaggregated serving
    while reducing the overhead of building networking behavior myself.

  * I want each stage to carry its own routing and networking configuration, so that a prefill interaction and a decode interaction can have different
    timeouts and failure behavior, as they have genuinely different performance characteristics.

  * I want this expressed in a portable way, so that I can change gateway implementation without rewriting how my workload is composed.

* As a **cluster or gateway administrator**:

  * I want to see the backend interactions caused by a logical request, so that I can attribute latency and failures to a specific stage rather than to 
    the request as a whole.

  * I want to apply policy to individual stages, so that an interaction with an external service and an interaction with an in-cluster backend can be
    governed differently within the same logical request.

* As a **Gateway API implementer**:

  * I want a standard way for a workload to describe that its requests span multiple backend interactions, so that I can support these workloads
    generically instead of integrating with each project's bespoke mechanism.

  * I want the boundary between what the routing layer decides and what the workload decides to be explicit, so that I can implement the routing side
    without encoding any workload's semantics.


## Prior Art

Multi-stage request processing is not hypothetical. Four projects implement it today: three distributed inference stacks and one proxy, each with a different architecture.

### llm-d (CNCF Sandbox project)

llm-d has implemented this twice, in this section we walkthrough both implementations.

#### First Try - The sidecar model

Orchestration lived in a sidecar running only on the vLLM decode worker; no coordination logic ran on prefill or encode nodes.
The EPP's `disagg-profile-handler` selected pods for every phase the request needed in a single workers selection cycle: decode always, then encode if multimodal content was detected, then prefill if the P/D decider judged it beneficial — and passed them to the sidecar as request headers.
The gateway forwarded the request to the decode pod and took no further part.

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Inference Gateway
    participant EPP as EPP
    participant SC as Decode Worker Sidecar
    participant P as Prefill Worker
    participant D as Decode Worker(vLLM)

    C->>GW: POST /v1/chat/completions
    GW->>EPP: Select all endpoint for all phases in a single cycle
    EPP-->>GW: Decode pod plus 'x-prefiller-host-port' header
    GW->>SC: Request with phase headers
    Note right of SC: Composition lives here.<br>The gateway sees one request<br>and one response and is not <br> aware of this sub-request.
    SC->>P: Remote prefill with max_tokens=1
    P-->>SC: KV transfer parameters
    SC->>D: Decode with KV transfer parameters
    D-->>SC: Stream
    SC-->>GW: Stream
    GW-->>C: Response
```

#### Second Try - The coordinator model

The sidecar is removed and orchestration is pulled out to a standalone coordinator in front of the Inference Gateway.
Client traffic enters through the Gateway: an HTTPRoute sends the client-facing inference paths to the coordinator, which then
makes one EPP-mediated call per phase back through the same Gateway, tagging each with an EPP-Profile header.
Each call is single-phase worker selection, so a pod is selected only when that phase's call is made.
Cross-phase state (which is essentially payload processing) lives on the coordinator's code and is very tailored to inference phases.

Two kinds of stage behave differently, and the difference matters. The inference phases are routed through the Gateway, because they need
EPP endpoint selection against an InferencePool. The render (tokenization service) step is called directly by the coordinator — llm-d's own
diagram labels the tokenizer a "side service".

The pipeline also begins with a conditional probe: the coordinator first tries decode alone, and only falls back to the full encode,
prefill and decode sequence when decode answers 412 Precondition Failed. That is per-request control flow, where a stage's response
decides what runs next, implemented here as application logic in a project-specific service.

```mermaid
sequenceDiagram
    participant Client
    participant Gateway
    participant Coordinator
    participant EPP
    participant Render as Render (side service)
    participant Encode as Encode pool
    participant Prefill as Prefill pool
    participant Decode as Decode pool

    Client->>Gateway: Request
    Gateway->>Coordinator: Forward request

    opt render step configured
        Coordinator->>Render: Normalize multimodal input (tokenize)
        Render-->>Coordinator: TokenIDs
    end

    Note over Coordinator,Decode: conditional-decode, try decode alone first
    Coordinator->>Gateway: Call (EPP-Profile: decode)
    Gateway->>EPP: ext_proc, pick pod (profile=decode)
    EPP-->>Gateway: Chosen decode pod
    Gateway->>Decode: Forward request

    alt Decode can serve the request
        Decode-->>Gateway: 200 OK (tokens)
        Gateway-->>Coordinator: Response
        Coordinator-->>Client: Final response
    else Decode cannot serve it (e.g. no KV cache for this prompt)
        Decode-->>Gateway: 412 Precondition Failed
        Gateway-->>Coordinator: 412
        Note over Coordinator: Fall back to the full pipeline

        opt request has multimodal entries
            par one call per media entry
                Coordinator->>Gateway: Call (EPP-Profile: encode)
                Gateway->>EPP: ext_proc, pick pod (profile=encode)
                EPP-->>Gateway: Chosen encode pod
                Gateway->>Encode: Forward request
                Encode-->>Gateway: Embeddings
                Gateway-->>Coordinator: Embeddings
            end
        end

        Coordinator->>Gateway: Call (EPP-Profile: prefill)
        Gateway->>EPP: ext_proc, pick pod (profile=prefill)
        EPP-->>Gateway: Chosen prefill pod
        Gateway->>Prefill: Forward request
        Prefill-->>Gateway: KV transfer descriptor
        Gateway-->>Coordinator: Response

        Coordinator->>Gateway: Call (EPP-Profile: decode)
        Gateway->>EPP: ext_proc, pick pod (profile=decode)
        EPP-->>Gateway: Chosen decode pod
        Gateway->>Decode: Forward request (pulls KV via NIXL)
        Decode-->>Gateway: 200 OK (tokens)
        Gateway-->>Coordinator: Response
        Coordinator-->>Client: Final response
    end
```

Every inference phase is a real routed request through the Gateway, with an EPP worker selection decision of its own.
The loop that drives them is not, the conditional-decode fallback is not, and the render stage is not.
Eight Gateway round-trips in the worst case, while the order decisions and payload processing lives outside the Gateway entirely.

llm-d is the only one of the three distributed inference projects that attempts to describe its stages in routing configuration,
and it takes two cooperating HTTPRoutes to do it. Abbreviated from the coordinator disaggregation guide:

#### The client-facing paths go to the coordinator.

```yaml
kind: HTTPRoute
metadata: {name: coordinator}
spec:
  parentRefs: [{kind: Gateway, name: llm-d-inference-gateway}]
  rules:
  - backendRefs: [{kind: Service, name: llm-d-coordinator, port: 8080}]
    matches:
    - {path: {type: Exact, value: /v1/completions}}
    - {path: {type: Exact, value: /v1/chat/completions}}
    - {path: {type: Exact, value: /inference/v1/generate}}
    timeouts: {request: 300s}
```

#### The coordinator's own per-phase calls come back in, distinguished only by a header.
```yaml
kind: HTTPRoute
spec:
  parentRefs: [{kind: Gateway, name: llm-d-inference-gateway}]
  rules:
  - backendRefs: [{group: inference.networking.k8s.io, kind: InferencePool, name: <pool>}]
    matches:
    - path: {type: Exact, value: /v1/completions}
      headers: [{type: Exact, name: EPP-Profile, value: encode}]
    - path: {type: Exact, value: /v1/completions}
      headers: [{type: Exact, name: EPP-Profile, value: prefill}]
    - path: {type: Exact, value: /v1/completions}
      headers: [{type: Exact, name: EPP-Profile, value: decode}]
    # ... the same three profiles repeated for each of the other two paths
    - path: {type: PathPrefix, value: /}
    timeouts: {request: 300s}
```

Several limits are visible in this example:

- The stage selector is client-forgeable. A stage can only be expressed as a header match, and a header is whatever the client sends,
  so the composition can be bypassed from outside. This is also an independent witness for the premise of [GEP-5224]: the reason it cannot
  be defended is that filters run after route selection.
- Stage dispatch rides on match-precedence arithmetic. The two routes must tie on path specificity so that the header match breaks the
  tie; as the guide notes, a prefix match "would lose to the Coordinator's exact match even when EPP-Profile is present." The composition
  works because of precedence rules, not because anything declares it.
- Combinatorial match explosion. Three paths times three profiles, plus a catch-all, to say that a request has three stages.
- One timeout for every stage. Each rule carries request: 300s, because there is nowhere to say that a prefill interaction and a
  decode interaction have different budgets.

### AIBrix

AIBrix puts orchestration inside the gateway, as a routing algorithm in its Envoy ext_proc plugin (pd_disaggregation.go, RouterPD = "pd").
Per request it groups pods by roleset, scores both groups and produces a (prefillPod, decodePod) pair, then issues the prefill request itself,
synchronously for vLLM, extracting kv_transfer_params, and for TRT-LLM, extracting disaggregated_params;
asynchronously for SGLang, which performs a bootstrap handshake instead before setting the target to the decode pod.

```mermaid
sequenceDiagram
    participant C as Client
    participant EG as Envoy Gateway
    participant PL as ext_proc plugin (pd router)
    participant P as Prefill pod
    participant D as Decode pod

    C->>EG: POST /v1/chat/completions
    EG->>PL: ext_proc
    Note over PL: scores pods, selects<br/>(prefillPod, decodePod)
    Note right of PL: Composition lives here.<br>The gateway sees one request<br>and one response and is not <br> aware of this sub-request.
    PL->>P: prefill sub-request — direct HTTP, not using route
    P-->>PL: kv_transfer_params
    PL-->>EG: target = decodePod
    EG->>D: decode + kv_transfer_params
    D-->>EG: stream
    EG-->>C: response
```

AIBrix HttpRoute as it looks today:
```yaml
kind: HTTPRoute
metadata: {name: qwen3-8b-router, namespace: aibrix-system}
spec:
  parentRefs: [{kind: Gateway, name: aibrix-eg, namespace: aibrix-system}]
  rules:
  - backendRefs: [{kind: Service, name: qwen3-8b, namespace: brixbench-adhoc, port: 8000}]
    matches:
    - headers: [{name: model, type: Exact, value: qwen3-8b}]
      path: {type: PathPrefix, value: /v1/chat/completions}
    timeouts: {request: 120s}
```

The prefill call is a plain HTTP request made from an extension process. It has no route, no retry policy, no timeout configuration, no
traffic policy and no status, because nothing in the routing layer knows it happened.
The 120s timeout covers the pair as though it were one exchange, and there is no way to give the prefill call a different budget, a different
retry policy, or a status of its own. An operator reading this cluster's routing configuration has no way to learn that disaggregation is happening at all.

### Dynamo

Dynamo uses a bespoke Frontend with an integrated Router. The frontend activates an internal prefill router when it discovers a decode
worker registered as WorkerType.Decode alongside a prefill service in the same namespace with a matching model name and
WorkerType.Prefill. The router selects a prefill worker by KV-aware routing on cache-overlap scores and load, receives NIXL transfer
metadata in disaggregated_params, injects it into the decode request and routes to the decode worker. Failure handling is deliberately
narrow: once dispatched to prefill, a prefill or handoff failure is returned to the client rather than retried through a decode-only path.

```mermaid
sequenceDiagram
    participant C as Client
    participant FE as Frontend + Router
    participant P as Prefill worker
    participant D as Decode worker

    C->>FE: POST /v1/chat/completions
    Note over FE: KV-aware selection<br/>(cache overlap + load)
    Note right of FE: Composition lives here.<br>The gateway sees one request<br>and one response and is not <br> aware of this sub-request.
    FE->>P: prefill
    P-->>FE: disaggregated_params (NIXL metadata)
    FE->>D: decode + disaggregated_params
    D-->>FE: stream
    FE-->>C: response
    P-->>FE: KV events (async)
```

An earlier documented design placed the decision in the decode worker instead: based on prefill length and queue depth it chose local
versus remote prefill, pushing a RemotePrefillRequest onto a global PrefillQueue that prefill workers consumed.

Dynamo HttpRoute as it looks today:
```yaml
kind: HTTPRoute
metadata: {name: qwen3-0-6b-agg}
spec:
  parentRefs: [{kind: Gateway, name: inference-gateway}]
  rules:
  - matches:
    - headers: [{name: X-Gateway-Model-Name, type: Exact, value: Qwen/Qwen3-0.6B}]
      path: {type: PathPrefix, value: /}
    backendRefs: [{group: inference.networking.k8s.io, kind: InferencePool, name: qwen3-0-6b-agg-pool, port: 8000}]
    timeouts: {request: 300s}
```

Dynamo shares the same shape and the same gaps as AIBrix: one visible request, while the sub-request's networking configuration is invisible.

### Praxis (CNCF Sandbox Project)

The three projects above are distributed inference systems, and each built request composition because its workload demanded it.
Praxis is a different kind of witness: It is a [CNCF Sandbox project][cncf-sandbox-praxis], proxy and gateway, the same category of
software as a Gateway API implementation, the same category of Gateways as the implementations that would have to support this GEP.
Praxis arrived at this Multi-Stage-Routing abstraction independently.

Praxis builds it in three layers.

#### A sub-request primitive
provides request and response types, a client with a bounded execution API, a pooled connector, pull-based streaming bodies, and
typed transport errors: admission timeout, connect, I/O, deadline exceeded, stream idle timeout, circuit open, response too large. 

#### An executor
runs a named filter pipeline against a sub-request, reconstructs a nested filter context, runs the request, request-body, response and
response-body phases, and applies the same forwarding boundary as the normal upstream path — a stage is not a side call, it is a request through the proxy's own machinery.

#### A composition filter
iterative_request_router, holds named steps, each backed by a pre-built sub-pipeline; per request it executes a step's filters, makes the call, executes the step's
response filters, evaluates transition rules, and either continues to another step or returns the response.

```mermaid
sequenceDiagram
    participant C as Client
    participant PX as Praxis iterative_request_router
    participant S1 as Step primary
    participant S2 as Step fallback

    C->>PX: Request
    Note right of PX: Named ordered steps, each a routed<br>sub-pipeline. Composition is<br>configuration, inside the proxy.
    PX->>S1: Step request via router and load_balancer
    S1-->>PX: 503
    Note right of PX: on_result matches 502/503/504,<br>next is fallback
    PX->>S2: Step request
    S2-->>PX: 200
    Note right of PX: on_result default, done
    PX-->>C: Response
```

Two things distinguish this from the three implementations above.

It is general. A step is an arbitrary filter sub-pipeline, not a role. Nothing in the mechanism knows what prefill is, and nothing
would need to change to express a guardrail, a tokenizer, or a plain service composition. This is what it looks like when composition is a
primitive rather than a special case.

It converged on the same shape this GEP describes: named ordered steps, per-step routing and networking configuration,
response-driven transitions, and both a whole-exchange deadline and a per-step one. Steps may reference reusable named chains or define
them inline, mixed freely.

It also answers two questions any design in this space has to face. On streaming, header-safe failover rules run before any bytes are exposed
downstream, while remaining transition rules run after clean EOF and may resume another step inside the same already-committed response;
the end-to-end deadline explicitly covers a streamed final response, not only the step exchanges. On accumulated state, cross-stage
state is bounded by an explicit ceiling. Two differences are instructive: Praxis permits backward re-entrance between steps and therefore
needs both an iteration cap and a reserved depth header to prevent loops, mechanisms this proposal does not need because stages are
forward-only and run at most once; and its transitions match response origin (upstream, local, transport) and transport error kind, a finer
distinction than status alone can express.

And none of it is expressible in Gateway API. Praxis implements Gateway API and configures this in its own YAML, because Gateway API
has no vocabulary for it. That is the gap in one sentence: an implementation has already built the capability, shipped documentation for
it, and still cannot let an operator ask for it portably. The feature is experimental and gated behind an off-by-default build flag, so the
right weight to give it is not a production deployment, but a considered design by maintainers of a Gateway API implementation who had to
answer the same questions this GEP raises.

### Beyond disaggregation

Prefill and decode is the loudest example, but not the only one.
The other shapes expose the problem more sharply because they are exposing the broader repeating patterns.

#### Representation transformation
The tokenizer described in [Motivation](#motivation) is an ordinary service: no InferencePool, no KV cache, no load-aware endpoint selection.
It is simply a stage that happens first, and inference projects do this today (e.g., Dynamo also issues a sub-request for tokenization).

#### Guardrails
An input guard inspects the prompt and may end the exchange with a canned response, so that a detected jailbreak attempt never reaches a model,
while an output guard inspects the response before it reaches the client. Both guards are themselves inference calls, to different models,
with latency and failure characteristics quite unlike the main model's. This is a stage before, a stage after, and early termination in between.

Neither is exotic, and neither is AI-specific in structure: one logical request, several ordered backend interactions, with the input of
one derived from the output of the last. The question is whether the existing implementations can express them.

AIBrix cannot. Its stages are "a prefill pod" and "a decode pod", identified by roleset labels; there is no notion of an arbitrary stage to configure.

Dynamo cannot. The set of expressible stages is the worker-type enumeration. There is no WorkerType.Guardrail, and adding one is a change to the
frontend rather than to configuration.

llm-d partly can. Its coordinator is a configurable pipeline of reorderable plugin steps, but the render step described above is called outside
the Gateway, with the same gaps as AIBrix's prefill call. Adding a guardrail stage means writing a Go plugin for llm-d's coordinator, which does
nothing for a platform team running AIBrix or Dynamo.

Praxis can, as described above, but only in its own configuration.

The specificity is the problem. Each project solved the composition it needed, in the vocabulary of that composition: roleset labels, a
worker-type enumeration, a profile header. None produced a way to say "an input guard, then a tokenizer, then prefill, then decode, then an
output guard", and a platform team that wants exactly that has nowhere to ask for it.

### Takeaway

The same thing is being built four times, in four places. A pod sidecar, an ext_proc plugin, a bespoke frontend process, and a proxy
filter. The handle carried between stages differs per engine — kv_transfer_params for vLLM, disaggregated_params for TRT-LLM, a
bootstrap room for SGLang — which is why this proposal keeps payload shaping in payload processing and stays agnostic to what is carried.

They are converging on the front door, and on the same vocabulary. Dynamo moved the decision out of the decode worker into its
frontend router; llm-d moved it out of the decode pod's sidecar into a coordinator that dispatches every phase through the Inference
Gateway. Praxis independently defined named ordered steps with response-driven transitions. The shape is not in dispute. Only its expression is.

Standardizing this is therefore not a matter of inventing an abstraction and hoping implementations adopt it: one Gateway API implementation
has already built the general form, and three more have built use-case-specific approximations because they had no portable alternative.

This can already be done, and that is precisely the problem: none of these projects can express the composition in routing configuration,
so each pays the costs described in [Motivation](#motivation).

What this proposal is aiming at

```mermaid
sequenceDiagram
    participant C as Client
    participant GW as Gateway
    participant R0 as Input guard
    participant R1 as Tokenizer
    participant R2 as Prefill
    participant R3 as Decode

    C->>GW: POST /v1/chat/completions
    Note right of GW: Composition is configuration:<br>which stages, in what order,<br>under what conditions
    GW->>R0: Stage input guard
    R0-->>GW: Allowed
    Note right of GW: A blocked verdict would end<br>the exchange here
    GW->>R1: Stage tokenize
    R1-->>GW: Token IDs
    GW->>R2: Stage prefill, with its<br>own timeouts and retries
    R2-->>GW: kv_transfer_params
    Note right of GW: Payload processing derives<br>the next stage's request body
    GW->>R3: Stage decode
    R3-->>GW: Stream
    GW-->>C: Response
```

No component beside the Gateway, no sub-request nor payload processing outside its reach, and every stage with its own networking configuration.

## API

**TODO**: API is deliberately left open for next phase until the What and Why are agreed.

## Open Points

**Fan-out is deferred to a follow-up proposal.** It is listed above as a motivating use case, but this proposal covers stages that execute at most once
per logical request. A follow-up would need to address how parallel interactions are expressed, how their responses are aggregated, failure
behavior per branch, and per-branch status.

**Processing order between stages.** Deriving a stage's input may take several processing steps whose order matters: deriving a body field from an
earlier stage and setting a routing header are not commutative. Gateway API currently states that implementations SHOULD apply filters in listed order,
not MUST. This is being discussed in [gateway-api#5194], which this proposal depends on.

**Streaming.** A single whole-request timeout is blunt when the last stage streams its response to the client, and there is no portable way today to
express a budget that covers the earlier stages but not the stream. This is the same gap [GEP-1742] and the draft GRPCRoute timeout work both leave
open, and it should be resolved with them rather than separately here.

## References

* [GEP-5091: PayloadProcessor Resource](https://github.com/kubernetes-sigs/gateway-api/issues/5091) — the Gateway API GEP incubated from the Payload Processing proposal.
* [Gateway API Inference Extension](https://github.com/kubernetes-sigs/gateway-api-inference-extension) — establishes inference-aware endpoint
  selection for a single backend pool.
* [gateway-api#5194: Filter ordering](https://github.com/kubernetes-sigs/gateway-api/issues/5194) — whether listed filter order is a guarantee.
  Ordering the processing between stages depends on it.

[GEP-5091]:https://github.com/kubernetes-sigs/gateway-api/issues/5091
[GEP-5224]:https://github.com/kubernetes-sigs/gateway-api/issues/5224
[GEP-1742]:https://gateway-api.sigs.k8s.io/geps/gep-1742/
[gateway-api#5194]:https://github.com/kubernetes-sigs/gateway-api/issues/5194
[cncf-sandbox-praxis]: https://github.com/cncf/sandbox/issues/506
