# GEP-4488: Backend Resource

* Issue: [#4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)
  * Incubated by the [AI Gateway Working Group](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/20)
* Status: Provisional

## TLDR

This GEP proposes a new `Backend` resource that fills the [backend role](/geps/gep-2907/) — a Gateway-native resource that can be referenced via `backendRefs` in Routes, just like `Service` is today. The Kubernetes `Service` resource is mature and stable, but it is effectively frozen and cannot be extended with Gateway-specific configuration. Every time we've wanted to add backend-level behavior (TLS settings, protocol metadata, connection policies), we've had to create separate policy CRDs like `BackendTLSPolicy` that attach to Service. This approach has significant limitations around discoverability, implementation complexity, and the conflation of producer and consumer concerns.

The `Backend` resource provides a namespace-scoped, consumer-focused resource that can:

1. **Decorate existing Services** via `EndpointSelector`, adding Gateway-specific configuration (TLS, protocol, etc.) without modifying the Service itself.
2. **Represent external destinations** via `ExternalHostname`, replacing the need for insecure synthetic `ExternalName` Services.
3. **Serve as a foundation for future Gateway-level backend configuration** such as retries, session persistence, load balancing algorithms, and other features that are tightly bound to the destination rather than the route.

While egress and AI use cases provide the initial urgent motivation, the `Backend` resource is designed to be useful for **all backend types**. At its core, a `Backend` of type `EndpointSelector` does what `Service` already does — but in a Gateway-native way that allows configuration to grow over time. Egress support via `ExternalHostname` is the first Extended feature built on this foundation.

## Motivation

The Kubernetes `Service` resource conflates two distinct concerns that have become increasingly problematic as Gateway API adoption grows:

1. **Frontend concerns**: How services are discovered and called (DNS names, ClusterIPs, service discovery)
2. **Backend concerns**: Where traffic should be routed and how to connect to destinations (endpoints, TLS configuration, protocol settings)

This conflation creates friction in a few notable areas:

### External Destination Limitationsr. The Backend resource provides

Currently, representing external destinations in Gateway API requires synthetic `Service` objects with `type: ExternalName`. There are several drawbacks to this approach:

- **Security vulnerabilities**: ExternalName Services are subject to DNS rebinding attacks ([CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675))
- **Policy limitations**: Cannot apply backend-specific policies (TLS, authentication, rate limiting) without affecting all consumers
- **Synthetic resource overhead**: Creates artificial Kubernetes resources for external dependencies

### Limitations of Policy-Based Decoration of Service

The current approach to adding Gateway-specific behavior to Services is through policy attachment (e.g., `BackendTLSPolicy`). While this works, it has significant limitations that become more acute with each new backend-level concern we want to add:

- **Discoverability**: When looking at a route, it is not clear what policies affect a given backend. Users must search for policy resources targeting the same Service, which scales poorly.
- **Producer vs Consumer ambiguity**: `BackendTLSPolicy` targets a Service, but it is unclear whether it represents the producer's TLS configuration or a consumer's desired TLS settings. Previous attempts to add consumer overrides (e.g., [GEP 3875](https://github.com/kubernetes-sigs/gateway-api/pull/3876)) introduced significant complexity and were ultimately abandoned.
- **Implementation complexity**: Implementations must reconcile potentially conflicting policies from multiple sources. Inline configuration on a dedicated Backend resource makes conflict resolution simpler, as the tightest scope wins by default.
- **CRD proliferation**: Each new backend-level concern (retries, session persistence, load balancing, timeouts) requires either a new policy CRD or extension of an existing one. The Backend resource provides a single, natural home for all configuration that describes "how to connect to a destination."

## Goals

- **Introduce Backend resource** as a namespace-scoped, consumer-focused resource for representing destinations and their Gateway-specific connection metadata
- **Decorate existing Services**: Allow `Backend` of type `EndpointSelector` to wrap a `Service` (or other endpoint-producing resource) with TLS, protocol, and other connection configuration, without modifying the underlying Service
- **Support external destinations**: Provide first-class `ExternalHostname` support as an Extended feature, replacing the need for synthetic `ExternalName` Services
- **Provide a home for backend-level configuration**: Inline TLS, protocol metadata, and (in the future) retries, session persistence, load balancing, and other destination-bound settings
- **Maintain Service compatibility**: Existing Service-based `backendRef`s continue to work indefinitely; Backend is additive, not a replacement
- **Enable incremental adoption**: At Core, a `Backend` of type `EndpointSelector` does what `Service` already does for Gateway API. Extended features (ExternalHostname, TLS, MCP protocol, etc.) can be adopted independently by implementations.

## Non-Goals

- **Deprecate or replace Services**: Services remain the primary backend type for internal destinations. Backend is a decorator, not a replacement.
- **Support producer-side policies**: Backend resource is explicitly consumer-focused
- **Provide cluster-scoped backends**: Backend resource is namespace-scoped for security boundaries
- **Solve all backend configuration at once**: This GEP establishes the Backend resource and its first features (EndpointSelector, ExternalHostname, inline TLS). Additional features (retries, session persistence, load balancing, etc.) will be proposed in follow-on GEPs.

## Relationship to Service

The `Backend` resource is **not** a replacement for `Service`. Instead, it is a decorator that adds a Gateway-specific configuration layer:

```
┌─────────────────────────────────────────────┐
│  HTTPRoute                                  │
│  backendRefs:                               │
│    - name: my-backend                       │
│      kind: Backend  ◄── Gateway-native ref  │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│  Backend (my-backend)                       │
│  type: EndpointSelector                     │
│  tls: { ... }         ◄── Gateway config    │
│  protocol: MCP        ◄── Gateway config    │
│  endpointSelector:                          │
│    selectorRef: my-svc  ◄── delegates to    │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│  Service (my-svc)                           │
│  (unchanged — still provides endpoints,     │
│   DNS, service discovery as before)         │
└─────────────────────────────────────────────┘
```

For external destinations, the `Backend` resource replaces the need for synthetic Services entirely:

```
┌─────────────────────────────────────────────┐
│  HTTPRoute                                  │
│  backendRefs:                               │
│    - name: openai-api                       │
│      kind: Backend                          │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│  Backend (openai-api)                       │
│  type: ExternalHostname                     │
│  hostname: api.openai.com                   │
│  tls: { ... }                               │
│  (no Service needed)                        │
└─────────────────────────────────────────────┘
```

Existing `Service`-based `backendRef`s in HTTPRoutes continue to work indefinitely. The `Backend` resource is an opt-in addition for users who need Gateway-specific backend configuration.

### Conformance Tiers

The Backend resource is designed with a clear separation between Core and Extended features:

| Feature | Conformance Level | Description |
|---|---|---|
| `EndpointSelector` type | Core | Backend wraps an existing Service; behaves equivalently to a Service `backendRef` |
| `ExternalHostname` type | Extended | First-class external FQDN support, replacing `ExternalName` Services |
| Inline TLS | Extended | TLS configuration inlined on the Backend resource |
| `MCP` protocol | Extended | Higher-level protocol metadata for AI/agentic use cases |

This layering allows the Backend resource itself to move to Standard quickly (Core tests just validate "does Backend do what Service does"), while Extended features mature independently.

## User Stories

### As an Application Developer (Service Decoration)

> "I want to add TLS configuration to my existing Service-backed backend without modifying the Service itself. Today, I have to create a separate `BackendTLSPolicy`, but it's not obvious from my HTTPRoute that TLS is configured, and I can't easily have different TLS settings for different consumers of the same Service."

### As an Application Developer (Egress)

> "I want to configure my application to call external APIs (like OpenAI) with specific TLS settings and authentication without creating synthetic Services that expose security risks or affect other applications."

### As a Platform Engineer

> "I want to enforce that all external API calls go through specific gateways with proper logging and policy enforcement, without having to manage complex Service configurations for every external dependency."

### As a Security Administrator

> "I want to avoid ExternalName Services due to DNS rebinding vulnerabilities while still allowing applications to declare their external dependencies in a structured, auditable way."

### As an Application Developer (Ingress)

> "I want to tell the Gateway how to connect to my internal Service — what TLS mode to use, what timeouts are appropriate, and what higher-level protocol my app speaks — without requiring a cluster admin to set up separate policy resources for each of these concerns. A single Backend resource that wraps my Service lets me express all of this in one place."

## Proposal

The `Backend` resource is a general-purpose, Gateway-native backend abstraction. It serves two complementary roles:

1. **Service Decorator (Core)**: A `Backend` of type `EndpointSelector` wraps an existing `Service` (or other endpoint-producing resource) and layers on Gateway-specific configuration — TLS, protocol metadata, and future features like retries or session persistence. At Core conformance, this type does what `Service` already does as a `backendRef`, but provides a dedicated resource where backend-level configuration can live and grow. This avoids the need for a separate policy CRD for each new backend-level concern.

2. **External Destination (Extended)**: A `Backend` of type `ExternalHostname` provides first-class support for external FQDNs, replacing the need for synthetic `ExternalName` Services. This is an Extended feature that addresses the urgent egress and AI use cases.

The Backend resource is explicitly designed as a **consumer resource** — it describes how a gateway should connect to a destination from the client perspective, regardless of whether that destination is internal or external to the cluster.

## TLS Policy Consolidation Analysis

One of the most significant design decisions for the Backend resource concerns TLS configuration: should it be inline within the Backend resource or provided through separate policy resources like `BackendTLSPolicy`?

### Tradeoffs: Inline TLS vs. Policy-Based TLS

#### Arguments for Inline TLS Configuration

1. **Simplified UX for External Destinations**
   - External FQDNs often require TLS configuration that is specific to that destination
   - Better discoverability for users who want to understand what TLS settings apply for backends within a route
   - Much simpler for implementations to integrate due to TLS settings being colocated with destination

2. **BackendTLSPolicy Limitations**
   - Current `BackendTLSPolicy` is designed around Service-based backends only
   - `BackendTLSPolicy` currently does not support per-consumer overrides
   - It is unclear whether `BackendTLSPolicy` is a producer or consumer oriented resource
     - [GEP 3875](https://github.com/kubernetes-sigs/gateway-api/pull/3876) proposed, among other things, adding consumer overrides to `BackendTLSPolicy`; however, that proposal introduced several new fields to the resource, including a `from` selector on `targetRef` that would have added significant complexity to both the API and implementations. Furthermore, the GEP has a [limitation](https://github.com/kubernetes-sigs/gateway-api/pull/3876/changes#diff-67a0076fb272af6273ce353d4687732735e03ddec8ae2bbc35b0a905281f9057R88) that it would not be possible to enforce consumer-side policies originating from the same namespace as the producer. This proposal was ultimately abandoned.

3. **Per-Backend Client Certificates**
   - Each external destination may require a different client certificate
   - Current Gateway API patterns only support one client certificate per Gateway
   - Backend-specific client certificates are essential for many external integrations

#### Arguments for Policy-Based Configuration

1. **API Consistency**
   *Note: the following points, while true, are only true because `Service` is effectively immutable.*
   - Gateway API uses policy attachment for most configuration beyond basic routing
   - Inline configuration creates another place to define TLS settings
   - Multiple places for TLS configuration could lead to misalignment of features

2. **Reusability and Standardization**
   - Policies can be shared across multiple Backends
   - Consistent TLS configuration patterns across resource types (e.g. `Service`, `Backend`, `InferencePool`, etc.)

#### Proposed Approach

Based on community feedback and practical considerations, this proposal recommends:

1. **Inline TLS for Backend Resources**: Provide inline TLS configuration within the Backend resource for simplicity and external destination requirements

2. **Explicit BackendTLSPolicy Exclusion**: Backend resources are explicitly disallowed as targets for `BackendTLSPolicy` to avoid confusion and conflicts

3. **Type Definition Alignment**: Align the inline TLS types with `BackendTLSPolicy` types as closely as possible for consistency

### Implementation Example

```yaml
# Gateway-level TLS remains authoritative for incoming connections
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
spec:
  listeners:
  - name: https
    protocol: HTTPS
    tls:
      certificateRefs:
      - name: gateway-cert
---
# Backend resource with inline TLS for external destination
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
metadata:
  name: openai-api
  namespace: ai-apps
spec:
  type: ExternalHostname
  externalHostname:
    hostname: api.openai.com
  tls:
    mode: ClientAndServer
    clientCertificateRef:
      name: openai-client-cert
  port: 443

---
# HTTPRoute referencing Backend
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
spec:
  rules:
  - backendRefs:
    - name: openai-api
      kind: Backend
      group: gateway.networking.k8s.io
      port: 443
```

## Security Model and RBAC Considerations

### FQDN Security Analysis

Allowing namespace-scoped Backend resources to reference external FQDNs raises legitimate security concerns that must be carefully considered:

#### Identified Security Risks

1. **DNS Spoofing Attacks**
   - Malicious DNS responses could redirect traffic to attacker-controlled servers
   - Particularly concerning for internal proxy endpoints or localhost addresses (i.e. the [Confused Deputy Problem](https://en.wikipedia.org/wiki/Confused_deputy_problem))
   - Risk: `api.external.com` resolves to `127.0.0.1`, `169.254.169.254` or other privileged, trusted addresses

2. **Cross-Namespace Service Access**
   - FQDNs could target internal cluster services via `svc.namespace.svc.cluster.local`
   - Potential bypass of namespace isolation and authorization controls
   - Risk: Accessing services in other namespaces without proper authorization

#### Risk Assessment and Mitigations

##### DNS Trust Model Decision

After extensive community discussion, this proposal adopts a **DNS trust model** for the following reasons:

1. **Egress Inherently Requires DNS Trust**
   - Any meaningful egress functionality must trust DNS resolution
   - Malicious DNS responses can redirect any external call regardless of validation
   - We will still provide some common sense validations (e.g., disallow all IPs, things ending in .cluster.local) but cannot fully mitigate DNS-based attacks

2. **RBAC and Admission Control as Primary Security Control**
   - Application developers are the persona target by the `Backend` resource
   - Admission control (e.g. VAP, Gatekeeper, Kyverno) can enforce organizational policies on FQDN usage
   - Network policies can restrict egress traffic regardless of Backend configuration (forcing DNS resolution to happen at the gateway only)

3. **Practical Effectiveness**
   - Trivial for attackers to register DNS records that resolve to internal addresses
     - Because of this, implementations implementing `Backend` MUST add either TLS or JWT validation on sensitive localhost endpoints to prevent confused deputy attacks
   - Restrictive validation would break legitimate external integrations
   - Security focus should be on network-level controls, not resource-level validation
   - No initial support for wildcard hostnames to limit attack surface and the need for Dynamic Forward Proxy support from implementations
     - Future proposals to add this functionality should include a comprehensive DNS trust specification and threat model.
   - Implementations may also be able to implement data-plane/proxy-level protections for common attack vectors
   - NOTE: This may be a decision we revisit in the future based on user feedback

#### Example of a Recommended Network Policy

```yaml
# Block traffic outside of the cluster
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-internal-egress-only
spec:
  podSelector: {} # Apply to all pods in namespace
  policyTypes:
  - Egress
  egress:
  # Allow traffic to all pods in all namespaces
  - to:
    - namespaceSelector: {}
  # Allow DNS resolution (Required)
  - to:
    - namespaceSelector:
        matchLabels:
          kubernetes.io/metadata.name: kube-system
      podSelector:
        matchLabels:
          k8s-app: kube-dns
    ports:
    - protocol: UDP
      port: 53
    - protocol: TCP
      port: 53
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-gateway-external
spec:
  podSelector:
    matchLabels:
      app: egress-gateway # Match your gateway's label
  policyTypes:
  - Egress
  egress:
  - {} # Allows everything, including external internet
```

#### Security Boundaries and Personas

**Namespace-Scoped App Developer Persona**
- Can create Backend resources within their namespace
- Limited to secrets within their namespace for TLS configuration
- Subject to network policies enforcing egress through gateways
- RBAC controls prevent cross-namespace resource access

**Cluster-Admin Risk Acceptance**
- Cluster administrators who grant Backend creation permissions accept DNS trust model
- Network-level controls (firewalls, proxy configuration) provide defense in depth
- Backend resources provide audit trail for external dependencies

## EndpointSelector Type

TODO: Reference GEP 4731 once it merges.

## Extension Framework

The Backend resource provides two levels for applying extensions and policies:

### 1. Route-Level Extensions (HTTPRoute Filters)

Applied to individual requests as they are routed to a Backend.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
spec:
  rules:
  - matches:
    - path: { value: "/api/models" }
    filters:
    - type: ExtensionRef
      extensionRef:
        name: rate-limiter
        kind: RateLimitPolicy
    backendRefs:
    - name: openai-api
      kind: Backend
```

### 2. Inline Backend Configuration (Future GEPs)

The Backend resource is designed to be the home for backend-level connection concerns that have historically required separate policy CRDs. Future GEPs will propose adding inline configuration for common concerns such as:

- **Retries**: Max retries, backoff strategy, retryable status codes
- **Session persistence**: Cookie-based, header-based, or connection-based affinity
- **Timeouts**: Connection timeout, request timeout, idle timeout
- **Load balancing**: Algorithm selection (round-robin, least-connections, consistent hashing)
- **Health checks**: Active health checking configuration for the destination

By inlining these into the Backend resource rather than requiring separate policy attachments, users get a single resource that fully describes "how to connect to this destination," improving discoverability and reducing the number of resources to manage.

```yaml
# Example of what future inline configuration might look like
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
metadata:
  name: my-api
spec:
  type: EndpointSelector
  endpointSelector:
    selectorRef:
      name: my-service
  port: 8080
  tls:
    mode: ServerOnly
  # Future inline fields (illustrative, not proposed in this GEP):
  # retries:
  #   attempts: 3
  #   backoff: exponential
  # timeouts:
  #   connect: 5s
  #   request: 30s
  # sessionPersistence:
  #   type: Cookie
```

Note: Policy attachment to Backend remains available for vendor-specific or niche configuration that doesn't warrant standardization in the upstream API.

## Graduation Criteria

This GEP follows the standard [Gateway API graduation criteria](/docs/concepts/versioning/#graduation-criteria). The following are additional criteria specific to this GEP:

### Implementable
- [ ] Backend resource CRD with full schema validation
- [ ] Documentation and examples for common use cases

### Experimental
- [ ] Reference implementation in at least one Gateway API implementation
- [ ] Basic conformance tests for FQDN and Service destination types
- [ ] Security review and RBAC documentation

### Standard
- [ ] At least 3 implementations with production usage
- [ ] Comprehensive conformance test suite
- [ ] Compatibility testing with existing BackendTLSPolicy patterns
- [ ] Migration guide from synthetic Services to Backend resources
- [ ] Integration with policy attachment framework

## Alternatives Considered

### Enhanced Service Resource

Extending the existing Service resource to support external destinations was considered but rejected due to:
- **Backward compatibility concerns**: Changes would affect all existing Service users
- **Security model conflicts**: External destination support conflicts with internal service patterns
- **API surface complexity**: Adding external destination fields to Service creates confusion

### Cluster-Scoped Backend Resource

Cluster-scoped Backend resources were considered but rejected due to:
- **Management complexity**: Requires coordination between cluster admins and app developers
- **Incorrect Persona Alignment**: Application developers are the primary consumers of backend resources, and they typically operate within namespace boundaries

### Policy-Only Approach

Using only policy attachment without a dedicated Backend resource was considered but rejected due to:

- **Destination representation gap**: No clear way to represent external FQDNs without synthetic Services
- **Policy target ambiguity**: Policies would still need to target synthetic Services
- **Extension limitations**: Protocol and connection options don't fit policy patterns well
- **CRD proliferation**: Each new backend-level concern would require its own policy CRD, leading to poor discoverability and implementation complexity. The Backend resource provides a single home for configuration that describes "how to connect to a destination"
