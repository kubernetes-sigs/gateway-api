# GEP-4747: L7 Reverse-Proxy Egress Gateway

* Issue: [#4747](https://github.com/kubernetes-sigs/gateway-api/issues/4747)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

[Chihiro]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas
[Ian]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas
[Ana]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#key-roles-and-personas

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

> **Note**: This GEP is Provisional. It proposes that existing Gateway API
> resources are sufficient for L7 reverse-proxy egress. A companion GEP
> ([GEP-4748]) explores an alternative approach using a dedicated
> `EgressGateway` resource. Community feedback on both approaches is
> explicitly requested.

[GEP-4748]: ../gep-4748/index.md
[GEP-4751]: https://github.com/kubernetes-sigs/gateway-api/issues/4751

## TLDR

Define how Gateway API resources compose for L7 reverse-proxy egress: routing
traffic from in-cluster workloads to external destinations through a Gateway,
using HTTPRoute for routing and the Backend resource
([PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)) for destination representation. This GEP
argues that no new Gateway-level resource is required.

## User Stories

### Platform Operator

> **[Ian] needs to provide workloads access to external AI services (OpenAI,
> Vertex AI, Bedrock) with centralized credential management, TLS policy,
> and observability -- without distributing API keys to individual workloads.**

### Application Developer

> **[Ana] needs her application to call external inference APIs through a
> managed gateway, with automatic failover between providers when the primary
> is unavailable.**

### Cluster Administrator

> **[Chihiro] needs to enforce that all outbound traffic to third-party
> services passes through a policy-enforced gateway, with per-namespace
> rate limiting and regulatory region locks.**

## Goals

* Establish egress as a first-class usage pattern of Gateway API
* Define how Gateway + HTTPRoute + Backend ([PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)) compose for egress
* Document egress-specific guidance for listeners, routes, and policy scoping
* Define Endpoint routing mode (direct to external destination); gateway chaining
  is covered in [GEP-4751]

## Non-Goals

* Introduce a new `EgressGateway` resource (see [GEP-4748] and
  [Alternatives Considered](#alternatives-considered))
* Define the Backend resource (see [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488))
* Address forward-proxy egress with dynamic routing
  (see [#4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704))
* Address L3/L4 network-level egress
* Address mesh-attached egress (sidecar/waypoint enforcement without Gateway)
* Solve workload-to-Gateway addressing
  (see [#1651](https://github.com/kubernetes-sigs/gateway-api/issues/1651))

## Introduction

### The Problem

Applications increasingly need managed access to services outside the cluster:
cloud AI APIs, cross-cluster inference endpoints, third-party services.
Kubernetes currently lacks standardized APIs for routing this traffic through a
policy-enforced Gateway. Workarounds include:

- **Synthetic ExternalName Services**: Subject to confused deputy attacks
  ([CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675)),
  break SNI/Host alignment, and cannot carry per-destination TLS policy.
- **Implementation-specific resources**: Istio `ServiceEntry`, Linkerd
  `EgressNetwork`, Cilium `CiliumEgressGatewayPolicy` -- each with different
  models, none standardized.

### Why No New Resource

A field-by-field analysis of the Gateway resource shows that some fields carry different meanings
depending on ingress or egress use cases: listeners, addresses, TLS, allowedRoutes. However, they remain
semantically valuable within both contexts:

| Field | Ingress Meaning | Egress Meaning |
|-------|----------------|----------------|
| `addresses` | How external clients reach the gateway | How internal workloads reach the gateway |
| `listeners[].port` | Port to accept incoming traffic | Port to accept outbound traffic from workloads |
| `listeners[].hostname` | Virtual host to match (SNI/Host header) | Filter workload requests by SNI/Host header |
| `listeners[].tls` | Cert presented to external clients (server TLS) | Cert presented to internal workloads (server TLS) |

The difference between ingress and egress is an emergent property of:

1. **Where the Gateway is deployed** (facing internal workloads, not the
   internet)
2. **What backends routes reference** (external Backend resources, not internal
   Services)

Both are already expressible in Gateway API. GatewayClass can provide a
mechanism for implementations to distinguish egress controllers from ingress
controllers.

### Prior Art

| Implementation | Egress Model | Separate Resource? |
|---|---|---|
| Istio | [Same `Gateway` resource](https://istio.io/latest/docs/tasks/traffic-management/egress/egress-gateway/) | No |
| Linkerd | [`EgressNetwork`](https://linkerd.io/2-edge/reference/egress-network/) (classifies traffic, not a Gateway) | Different model |
| Cilium | `CiliumEgressGatewayPolicy` (L3/L4) | Different layer |

Implementations that use Gateway API's Gateway resource for egress do NOT
require a separate resource type.

## API

### No New API Types

This GEP does not introduce new API types. It defines how existing types
compose for egress:

```
                     parentRef              backendRef
Workload --> Gateway <-------- HTTPRoute ----------> Backend (PR #4488)
              |                   |                     |
         GatewayClass        hostnames:            destination:
         (egress)          ["*.openai.com"]       type: Hostname
                                                  hostname:
                                                    address: api.openai.com
                                                  ports:
                                                  - number: 443
                                                    tls: ...
```

### Egress Gateway Configuration

An egress gateway is a `Gateway`. An egress-specific `GatewayClass` MAY be
used to apply egress-specific validation, defaults, and to denote that the
Gateway is being used for egress:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: egress
spec:
  controllerName: example.com/egress-gateway-controller
---
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: egress-gateway
  namespace: gateway-system
spec:
  gatewayClassName: egress
  listeners:
  - name: https
    port: 8443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
      - name: egress-gw-cert
```

**Listener guidance for egress**:

- Egress gateways typically use a single listener with no hostname filter
  (accepting all destinations), though setting `hostname` to filter by
  destination IS valid and means "only accept requests going to this host."
- `listeners[].tls` configures TLS for workloads connecting to the gateway.
  The TLS configuration is identical to ingress. TLS to external destinations
  is configured on the Backend resource ([PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)).

### Egress Routing

HTTPRoute attaches to the egress Gateway and references Backend resources:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: openai-route
  namespace: app-team
spec:
  parentRefs:
  - name: egress-gateway
    namespace: gateway-system
  hostnames:
  - "api.openai.com"
  rules:
  - backendRefs:
    - group: gateway.networking.k8s.io
      kind: Backend
      name: openai-backend
```

`hostnames` on an egress HTTPRoute matches requests whose destination
Host header is one of these values. The matching behavior is identical
to ingress.

### Routing Modes

#### Endpoint Mode

Traffic flows directly from egress gateway to external destination:

```
Workload --> Egress Gateway --> External API (Backend)
```

The gateway applies policies, resolves the Backend destination, (optionally) originates
TLS to the external endpoint, and forwards the request.

### Policy Application Scopes

Egress policies apply at three levels:

| Scope | Mechanism | Egress Examples |
|-------|-----------|-----------------|
| **Gateway** | Policy attachment to Gateway | CIDR deny lists, global rate limits, default deny |
| **Route** | HTTPRoute filters, ExtensionRef | Payload transforms, compliance checks, guardrails |
| **Backend** | Backend.spec.tls, Backend.spec.filters [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488) | Per-destination TLS, credential injection |

**Conflict resolution**: Route > Backend > Gateway (most specific wins).
Same-level ties: oldest resource by `creationTimestamp` wins. Implementations
MUST surface conflicts in status conditions.

### Workload-to-Gateway Addressing (Interim)

Until Gateway Routability ([#1651](https://github.com/kubernetes-sigs/gateway-api/issues/1651))
is resolved, workloads connect to the egress gateway via an implementation-managed
address reachable only within the cluster (often a ClusterIP Service).

Implementations should:

- Expose an internally-reachable address for egress Gateways
- Report this address in `gateway.status.addresses`
- Use a stable DNS name (e.g., `<gateway-name>.<namespace>.svc.cluster.local`)

### Traffic Enforcement

For egress gateways to be effective, operators should deny direct egress
from workloads and force outbound traffic through the Gateway using
enforcement mechanisms (like NetworkPolicy, sidecar configuration, or
CNI-level controls). This GEP does not define enforcement mechanisms -- it defines what the
Gateway does once traffic arrives.

## Security Considerations

### Egress-Specific Risks

1. **Data exfiltration**: A compromised workload could use the egress gateway
   to send data to attacker-controlled endpoints. Mitigation: restrict Backend
   creation via RBAC, use NetworkPolicy to limit which namespaces can reach the
   gateway.

2. **Credential exposure**: Backend extensions may inject credentials into
   requests. The egress gateway becomes a high-value target. Mitigation:
   standard Kubernetes secrets RBAC, audit logging.

3. **Gateway bypass**: Workloads with direct internet access bypass all egress
   policy. NetworkPolicy enforcement may be used to mitigate this.

## Open Questions

> These are explicitly marked as open for community feedback during
> Provisional status.

### 1. Gateway Reuse vs EgressGateway Resource

This GEP argues that no new resource is needed. [GEP-4748] proposes a
dedicated `EgressGateway` resource as an alternative. Community input is
needed to decide which approach to pursue. See
[Alternatives Considered](#alternatives-considered).

### 2. Listener Hostname Guidance

Should the GEP recommend specific listener configurations for egress (e.g.,
"use a single wildcard listener") or leave this entirely to implementations?

### 3. Mixed Ingress/Egress Gateways

Should a single Gateway be allowed to serve both ingress and egress traffic
(via multiple listeners)?

## Alternatives Considered

### New EgressGateway Resource (GEP-4748)

A dedicated `EgressGateway` resource was prototyped
([wg-ai-gateway PR #45](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/45))
and is proposed in companion [GEP-4748].

**Arguments for**: User clarity, prevents accidental misuse, access controls, reserves design
space for egress-specific fields.

**Arguments against**: Zero UX differences at the Gateway level, API
fragmentation, precedent risk, may introduce confusion when paired with mixed-mode
architectures (using Gateway).

See [GEP-4748] for the full proposal.

### EgressRoute (Prior GEP #1971)

A [previous attempt](https://github.com/kubernetes-sigs/gateway-api/pull/1971)
proposed an `EgressRoute` resource. It was closed without merge. This GEP
takes a different approach: egress routing uses existing HTTPRoute with Backend
as the destination type.

### Forward Proxy Model

Dynamic routing to arbitrary hostnames (forward proxy) is out of scope. See
[#4704](https://github.com/kubernetes-sigs/gateway-api/issues/4704).

## Dependencies

| Dependency | Status | Impact |
|---|---|---|
| [PR #4488: Backend Resource](https://github.com/kubernetes-sigs/gateway-api/pull/4488) | PR open | Required -- egress routes need Backend destinations |
| [#1651: Gateway Routability](https://github.com/kubernetes-sigs/gateway-api/issues/1651) | Issue open | Nice-to-have -- defines how workloads address gateways |

## References

* [WG AI Gateway egress proposal](https://github.com/kubernetes-sigs/wg-ai-gateway/blob/main/proposals/10-egress-gateways.md)
* [GEP-4488: Backend Resource](https://github.com/kubernetes-sigs/gateway-api/pull/4488)
* [GEP-1897: BackendTLSPolicy](../gep-1897/index.md)
* [GEP-2907: TLS Terminology](../gep-2907/index.md)
* [Issue #1651: Gateway Routability](https://github.com/kubernetes-sigs/gateway-api/issues/1651)
* [Issue #4704: Forward Proxy Egress](https://github.com/kubernetes-sigs/gateway-api/issues/4704)
* [PR #1971: Prior Egress GEP](https://github.com/kubernetes-sigs/gateway-api/pull/1971)
* [CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675)

## Graduation Criteria

### Provisional -> Implementable

- [ ] Community decision on Gateway reuse (this GEP) vs EgressGateway ([GEP-4748])
- [ ] [PR #4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488) (Backend) reaches at least Provisional status
- [ ] Open questions resolved

### Alpha (Experimental)

- [ ] At least one implementation supports egress usage pattern
- [ ] Documentation for egress usage patterns

### Beta

- [ ] At least two implementations support egress usage pattern
- [ ] Production usage reports from 2+ organizations
- [ ] Gateway Routability (#1651) resolved or workaround documented

### GA (Standard)

- [ ] Three implementations supporting egress usage pattern
- [ ] No API-level changes needed for 6+ months
- [ ] Security review complete

