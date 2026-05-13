# GEP-4488: Backend Resource

* Issue: [#4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)
  * Incubated by the [AI Gateway Working Group](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/20)
* Status: Experimental

## TLDR

This GEP proposes a new `Backend` resource that fills the [backend role](/geps/gep-2907/) — a **general-purpose decorator for Service** (and other backend types) within Gateway API. The Kubernetes `Service` resource is mature and stable, but it is effectively frozen and SIG-Network leadership is very careful about any potential features that would further bloat `Service`'s responsibilities. Previous approaches to extend Service behavior (like `BackendTLSPolicy`) have significant limitations around discoverability, implementation complexity, and the conflation of producer and consumer concerns. `BackendTLSPolicy` in particular was the right solution at the time, but feedback has shown that policy attachment was not the right approach for TLS configuration.

The `Backend` resource provides a namespace-scoped, consumer-focused resource that can:

1. **Decorate existing Services** via `EndpointSelector`, adding Gateway-specific configuration (TLS, protocol, etc.) without modifying the Service itself.
2. **Represent external destinations** via `ExternalHostname`, replacing the need for insecure synthetic `ExternalName` Services.
3. **Serve as a foundation for future Gateway-level backend configuration** such as retries, session persistence, load balancing algorithms, and other features that are tightly bound to the destination rather than the route.

While egress and AI use cases provided the initial urgent motivation, the `Backend` resource is designed to be useful for **all backend types**. At its core, a `Backend` of type `EndpointSelector` does what `Service` already does — but in a Gateway-native way that allows configuration to grow over time. Egress support via `ExternalHostname` is the first Extended feature built on this foundation.

### Clarifying the Semantics of "Backend"

It is critically important that we emphasize that the `Backend` resource describes what a specific gateway client connection **MUST** do on the wire when connecting to a destination.

In the common ingress persona, Ana often owns both the `HTTPRoute` and the Service behind it. In this scenario, Ana is being delegated the ability to control how the Gateway consumes her service. In many egress, mesh, and AI-oriented deployments, that ownership model changes: Ana does NOT own the destination behind the route, but she still needs a way to express "how should the gateway connect to this destination" without needing to coordinate with the producer or cluster admin. In other words, the `Backend` resource is a consumer-side resource that describes connection requirements for a particular client path, regardless of who owns the destination.

This distinction is foundational to this GEP:

- `Backend` captures consumer-side connection requirements for a particular client path (for example, SNI, TLS validation, client cert presentation, or higher-level protocol expectations).
- Producer/server guidance is still valuable, out of scope for this resource: producer hints describe what servers generally **SHOULD** accept, while this resource describes what this client/gateway connection **MUST** attempt.
- In practice, server hints were always actuated as client configuration anyway. This proposal makes that behavior explicit and auditable.

Said differently: `Backend` is intentionally scoped to "how this client should connect" rather than "what all clients of this server must do."

## Motivation

The Kubernetes `Service` resource conflates two distinct concerns that have become increasingly problematic as Gateway API adoption grows:

1. **Frontend concerns**: How services are discovered and called (DNS names, ClusterIPs, service discovery)
2. **Backend concerns**: Where traffic should be routed and how to connect to destinations (endpoints, TLS configuration, protocol settings)

This conflation creates friction in a few notable areas:

### External Destination Limitations

Currently, representing external destinations in Gateway API requires synthetic `Service` objects with `type: ExternalName`. There are several drawbacks to this approach:

- **Security vulnerabilities**: ExternalName Services are subject to DNS rebinding attacks ([CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675))
- **Policy limitations**: Cannot apply backend-specific policies (TLS, authentication, rate limiting) without affecting all consumers
- **Synthetic resource overhead**: Creates artificial Kubernetes resources for external dependencies

### Limitations of Policy-Based Decoration of Service

The current approach to adding Gateway-specific behavior to Services is through policy attachment (e.g., `BackendTLSPolicy`). While this works, it has significant limitations:

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
- **Standardize producer-owned backend policy in this GEP**: Producer guidance and hints may inform client behavior, but defining producer-authoritative policy semantics remains out of scope for this proposal.
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
│  EndpointSlice (my-svc)                     │
│  (unchanged — still provides endpoints,     │
│   as before)                                │
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
│  externalHostname:                          │
│    hostname: api.openai.com                 │
│  tls: { ... }                               │
│  (no Service needed)                        │
└─────────────────────────────────────────────┘
```

Existing `Service`-based `backendRef`s in HTTPRoutes continue to work indefinitely. The `Backend` resource is an opt-in addition for users who need Gateway-specific backend configuration.

### Conformance Tiers

The Backend resource is designed with a clear separation between Core and Extended features:

| Feature | Conformance Level | Description |
| --- | --- | --- |
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

## API Specification

### Backend Resource Schema

```go
type Backend struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`
  Spec   BackendSpec   `json:"spec"`
  Status BackendControllerStatus `json:"status,omitempty"`
}

// +kubebuilder:validation:Enum=ExternalHostname;EndpointSelector
type BackendType string

const (
  BackendTypeExternalHostname             BackendType = "ExternalHostname"
  BackendTypeEndpointSelector             BackendType = "EndpointSelector"
)

// +kubebuilder:validation:XValidation:rule="self.type == 'ExternalHostname' ? has(self.externalHostname) : !has(self.externalHostname)",message="externalHostname must be set when type is ExternalHostname and must be unset otherwise"
// +kubebuilder:validation:XValidation:rule="self.type == 'EndpointSelector' ? has(self.endpointSelector) : !has(self.endpointSelector)",message="endpointSelector must be set when type is EndpointSelector and must be unset otherwise"
type BackendSpec struct {
  // Type defines the backend type
  // +unionDiscriminator
  // +required
  Type BackendType `json:"type"`
  // Port defines the port that the implementation should use when connecting to this backend.
  // +required
  Port PortNumber `json:"port,omitempty"`

  // ExternalHostname specifies the configuration for an ExternalHostname backend. Only used if type is ExternalHostname.
  // Support: Extended
  // +optional

  ExternalHostname *ExternalHostnameBackend `json:"externalHostname,omitempty"`

  // EndpointSelector specifies the configuration for an EndpointSelector backend. Only used if type is EndpointSelector.
  // As defined in GEP-4731, creation of a `Backend` of type `EndpointSelector` should result in `Backend` controllers
  // creating the requisite `EndpointSelector` resource and setting ownerReferences appropriately.
  // TODO: Add link when GEP-4731 merges.
  // +optional
  EndpointSelector *EndpointSelectorBackend `json:"endpointSelector,omitempty"`

  // Protocol defines the protocol for backend communication.
  // In the common case, the underlying transport protocol for the
  // proxied traffic will already have been determined and processed
  // by the dataplane at the routing step. Where this field is useful
  // is either for higher level protocols or asymmetrical protocol
  // configurations (e.g. version upgrades or h2c). In cases where the
  // protocol is negotiated on the wire (e.g. HTTP/1.1 Upgrade or ALPN),
  // implementations MUST include the protocol set here in the negotiation
  // options presented to the backend. It is currently undefined whether this
  // means required, optional, or most preferred (e.g. first in the set).
  // TODO: Define full semantics in protocol negotation.
  //
  // These protocols are also used for validation of future protocol-specific
  // fields that may be added to the Backend resource (e.g. retries, session persistence, etc.)
  //
  // Support: Extended for MCP, Core for TCP, HTTP, HTTP2, and H2C
  // TODO: Not sure if the above is allowed or viable.
  // +optional
  Protocol BackendProtocol `json:"protocol"`

  // TLS defines the TLS configuration that a client should use when talking to the backend.
  // N.B: ExternalHostname backends SHOULD have TLS configured; the lack of TLS for external hostnames
  // should be considered insecure and a security risk.
  // +optional
  TLS *BackendTLS `json:"tls,omitempty"`
}

// BackendTLSMode defines the TLS mode for backend connections.
// +kubebuilder:validation:Enum=None;ServerOnly;ClientAndServer
type BackendTLSMode string

const (
  // Disable TLS when connecting to the backend.
  BackendTLSModeNone BackendTLSMode = "None"
  // Enable TLS with simple server certificate verification.
  BackendTLSModeServerOnly BackendTLSMode = "ServerOnly"
  // Enable mutual TLS.
  BackendTLSModeClientAndServer BackendTLSMode = "ClientAndServer"
)

// +kubebuilder:validation:ExactlyOneOf=SelectorRef,Selector
type EndpointSelectorBackend struct {
  // SelectorRef specifies the reference to the EndpointSelector resource that manages the EndpointSlices for this backend.
  // If omitted, the controller creating this Backend is expected to create an EndpointSelector
  // resource on behalf of the user and set ownerReferences appropriately so that the lifecycle
  // of the EndpointSelector is tied to this Backend (as described in GEP-4731).
  // +optional
  SelectorRef *LocalObjectReference `json:"selectorRef"`

  // Selector defines the label selector used to identify the endpoints that this backend
  // should route traffic to. This field is only used if SelectorRef is not specified.
  // +optional
  Selector *metav1.LabelSelector `json:"selector,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="self.mode == 'ClientAndServer' ? has(self.clientCertificateRef) : !has(self.clientCertificateRef)",message="clientCertificateRef must be set if and only if mode is ClientAndServer"
type BackendTLS struct {
  // Mode defines the TLS mode for the backend.
  // +required
  Mode BackendTLSMode `json:"mode"`

  // ClientCertificateRef defines the reference to the client certificate for mutual
  // TLS. Only used if mode is ClientAndServer.
  // +optional
  ClientCertificateRef *SecretObjectReference `json:"clientCertificateRef,omitempty"`

  // Re-use BackendTLS policy validation fields. This is currently missing InsecureSKipVerify
  // but that will be added in GEP-4152.
  Validation BackendTLSPolicyValidation `json:"validation,omitempty"`
}

type BackendControllerStatus struct {
  // Name is a domain/path string that indicates the name of the controller that manages the
  // Backend. Name corresponds to the GatewayClass controllerName field when the
  // controller will manage parents of type "Gateway". Otherwise, the name is implementation-specific.
  //
  // Example: "example.net/import-controller".
  //
  // The format of this field is DOMAIN "/" PATH, where DOMAIN and PATH are valid Kubernetes
  // names (https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names).
  //
  // A controller MUST populate this field when writing status and ensure that entries to status
  // populated with their controller name are removed when they are no longer necessary.
  //
  // +required
  Name ControllerName `json:"name"`
  // For Kubernetes API conventions, see:
  // https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
  // conditions represent the current state of the Backend resource.
  // Each condition has a unique type and reflects the status of a specific aspect of the resource.
  //
  // Standard condition types include:
  // - "Available": the resource is fully functional
  // - "Progressing": the resource is being created or updated
  // - "Degraded": the resource failed to reach or maintain its desired state
  //
  // The status of each condition is one of True, False, or Unknown.
  // +listType=map
  // +listMapKey=type
  // +optional
  Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

### ExternalHostname Backend Configuration

```go
type ExternalHostnameBackend struct {
  // Hostname specifies the destination address used to reach this hostname.
  // IP addresses are not allowed in this field (enforced by validation on the type).
  // If implementations are aware of custom trust domains being used for `Service` FQDNs,
  // the MUST also enforce that hostnames ending with those trust domains (e.g. `.cluster.local`) are not allowed.
  // +kubebuiler:validation:XValidation:rule="!endsWith(self.hostname, '.cluster.local')))",message="hostname must not be an IP address or end with .cluster.local"
  Hostname PreciseHostname `json:"hostname"`
}
```

### Protocol and Extension Support

```go
// +kubebuilder:validation:Enum=TCP,HTTP,HTTP2,HTTP11,H2C,MCP
type BackendProtocol string

const (
  BackendProtocolMCP   BackendProtocol = "MCP"
  BackendProtocolTCP   BackendProtocol = "TCP"
  BackendProtocolHTTP  BackendProtocol = "HTTP"
  BackendProtocolHTTP2 BackendProtocol = "HTTP2"
  BackendProtocolH2C   BackendProtocol = "H2C"
  BackendProtocolHTTP11 BackendProtocol = "HTTP11"
)
```

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
     - Implementations SHOULD provide control-plane guardrails (for example, an allow-list of permitted egress domains)
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
- **Health checks**: Active health checking configuration for the destination from the consumer dataplane perspective

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
