# GEP-XXX: Backend Resource for Egress and Service Decoration

* Issue: [#XXXX](https://github.com/kubernetes-sigs/gateway-api/issues/XXXX)
* Status: Provisional

## TLDR

This document specifies the `Backend` resource, a namespace-scoped consumer resource that represents destinations for Gateway API routing. The Backend resource supports external FQDNs, IP addresses, and Kubernetes Services while providing inline TLS configuration, protocol options, and extension points for vendor-specific functionality.

## Goals

- **Define Backend resource schema** supporting multiple destination types with connection metadata
- **Establish consumer resource pattern** for consistent client-side connection semantics
- **Provide inline TLS configuration** for external destinations without synthetic Services
- **Support protocol extensions** for vendor-specific protocols and connection options
- **Enable policy attachment framework** for backend-specific policies and configurations
- **Document security model** for external FQDN access and RBAC patterns

## Non-Goals

- **Replace existing BackendTLSPolicy** for Service targets (remains recommended)
- **Support producer-side policies** (authorization, authentication to the backend)
- **Provide cluster-scoped backends** (security boundary enforcement)
- **Standardize service mesh integration** (implementation-specific)

## Introduction

The Backend resource addresses a fundamental gap in Gateway API's backend reference system: first-class support for external destinations. Currently, external APIs, databases, and services must be represented as synthetic Kubernetes Services with `type: ExternalName`, which creates security vulnerabilities, policy application challenges, and resource management overhead.

The Backend resource is explicitly designed as a **consumer resource** - it describes how a gateway should connect to a destination from the client perspective, regardless of whether that destination is internal or external to the cluster.

## API Specification

### Backend Resource Schema

```go
type Backend struct {
    metav1.TypeMeta      `json:",inline"`
    metav1.ObjectMeta    `json:"metadata,omitempty"`
    Spec   BackendSpec   `json:"spec"`
    Status BackendStatus `json:"status,omitempty"`
}

type BackendSpec struct {
    // Destination defines where traffic should be sent
    Destination BackendDestination `json:"destination"`

    // Filters defines filters that should be executed when
    // sending traffic to this backend. Filters should not
    // be duplicated on a backendRef (targeting this `Backend`)
    // and on the `Backend` itself.
    // +optional
    // TODO: Specify filter type definition. Should,
    // at minimum, include ExtensionRef pattern.
    Filters []BackendFilters `json:"extensions,omitempty"`
}

type BackendDestination struct {
    // Type defines the destination type
    Type BackendType `json:"type"`

    // Ports defines the destination ports and protocols
    // +optional
    Ports []BackendPort `json:"ports,omitempty"`

    // FQDN specifies an external fully qualified domain name
    // +optional
    FQDN *FQDNBackend `json:"fqdn,omitempty"`

    // Service references an existing Service
    // +optional
    Service *ServiceBackend `json:"kubernetesService,omitempty"`
}

type BackendType string

const (
    BackendTypeFQDN    BackendType = "FQDN"
    BackendTypeService BackendType = "Service"
)
```

### FQDN Backend Configuration

```go
type FQDNBackend struct {
    // Hostname specifies the destination FQDN
    Hostname string `json:"hostname"`

    // TLS configuration for the connection
    // +optional
    TLS *BackendTLSConfig `json:"tls,omitempty"`

    // Protocol specifies the application protocol
    // +optional
    Protocol *BackendProtocol `json:"protocol,omitempty"`
}

type BackendTLSConfig struct {
    // ServerName for TLS SNI
    // +optional
    ServerName *string `json:"serverName,omitempty"`

    // CACertificates for server verification
    // +optional
    CACertificates []BackendTLSCertificate `json:"caCertificates,omitempty"`

    // ClientCertificate for mTLS
    // +optional
    ClientCertificate *BackendTLSCertificate `json:"clientCertificate,omitempty"`

    // InsecureSkipVerify disables server certificate verification
    // +optional
    InsecureSkipVerify *bool `json:"insecureSkipVerify,omitempty"`
}

type BackendTLSCertificate struct {
    // Secret reference containing certificate data
    SecretRef SecretObjectReference `json:"secretRef"`

    // Key within the secret containing the certificate
    // +optional
    Key *string `json:"key,omitempty"`
}
```

### Protocol and Extension Support

```go
type BackendProtocol struct {
    // Type specifies the protocol type
    Type string `json:"type"`

    // Options provides protocol-specific configuration
    // +optional
    Options map[string]string `json:"options,omitempty"`
}
```

## TLS Policy Consolidation Analysis

One of the most significant design decisions for the Backend resource concerns TLS configuration: should it be inline within the Backend resource or provided through separate policy resources like `BackendTLSPolicy`?

### Tradeoffs: Inline TLS vs. Policy-Based TLS

#### Arguments for Inline TLS Configuration

1. **Simplified UX for External Destinations**
   - External FQDNs often require TLS configuration that is specific to that destination
   - Better discoverability for users who want to understand what TLS settings apply for backends within a route.
   - Much simpler for implementations to integrate due to TLS settings being colocated with destination

2. **BackendTLSPolicy Limitations**
   - Current `BackendTLSPolicy` is designed around Service-based backends only.
   - `BackendTLSPolicy` currently does not support per-consumer overrides
   - It is unclear whether `BackendTLSPolicy` is a producer or consumer oriented resource.

3. **Per-Backend Client Certificates**
   - External destinations may require different client certificates
   - Current Gateway API patterns only support one client certificate per Gateway
   - Backend-specific client certificates are essential for many external integrations

#### Arguments for Policy-Based Configuration

1. **API Consistency**
   - Gateway API uses policy attachment for most configuration beyond basic routing
   - Inline configuration creates another place to define TLS settings
   - Multiple places for TLS configuration could lead to misalignment of features

2. **Reusability and Standardization**
   - Policies can be shared across multiple Backends
   - Consistent TLS configuration patterns across resource types (e.g. `Service`, `Backend`, `InferencePool`, etc.)

#### Proposed Compromise Approach

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
  # Deprecate gateway.spec.tls.backend in favor of Backend resource configuration

---
# Backend resource with inline TLS for external destination
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
metadata:
  name: openai-api
  namespace: ai-apps
spec:
  destination:
    type: FQDN
    fqdn:
      hostname: api.openai.com
      tls:
        serverName: api.openai.com
        clientCertificate:
          secretRef:
            name: openai-client-cert
            key: tls.crt
    ports:
    - name: https
      port: 443
      protocol: HTTPS

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
```

## Security Model and RBAC Considerations

### FQDN Security Analysis

Allowing namespace-scoped Backend resources to reference external FQDNs raises legitimate security concerns that must be carefully considered:

#### Identified Security Risks

1. **DNS Spoofing Attacks**
   - Malicious DNS responses could redirect traffic to attacker-controlled servers
   - Particularly concerning for internal proxy endpoints or localhost addresses (i.e. the [Confused Deputy Problem](https://en.wikipedia.org/wiki/Confused_deputy_problem))
    - Risk: `api.external.com` resolves to `127.0.0.1`, `169.254.169.2554` or other privileged, trusted addresses

2. **Cross-Namespace Service Access**
   - FQDNs could target internal cluster services via `svc.namespace.svc.cluster.local`
   - Potential bypass of namespace isolation and RBAC controls
   - Risk: Accessing services in other namespaces without proper authorization

#### Risk Assessment and Mitigations

**DNS Trust Model Decision**

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
   - Trivial for attackers to register FQDNs that resolve to internal addresses
   - Restrictive validation would break legitimate external integrations
   - Security focus should be on network-level controls, not resource-level validation
   - No initial support for wildcard FQDNs to limit attack surface and the need for Dynamic Forward Proxy support from implementations
    - Implementations may also be able to implement data-plane/proxy-level protections for common attack vectors
    - NOTE: This may be a decision we revisit in the future based on user feedback

#### Example of a Recommended Network

```yaml
# Network policy enforcement
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: egress-through-gateway
  namespace: application-team
spec:
  podSelector: {}
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: gateway-system
    ports:
    - protocol: TCP
      port: 8080  # Gateway proxy port
  # Block direct external egress, force through gateway
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

## Extension Framework

The Backend resource provides three levels for applying extensions and policies:

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

**Use cases**: Request modification, rate limiting, authentication injection

### 2. Backend-Level Extensions

Applied at the Backend resource, affecting all requests to that destination.

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha1
kind: Backend
spec:
  destination:
    type: FQDN
    fqdn:
      hostname: api.openai.com
  extensions:
  - type: "vendor.io/connection-pool"
    config:
      maxConnections: 100
      timeout: 30s
  - type: "vendor.io/circuit-breaker"
    config:
      failureThreshold: 5
```

**Use cases**: Connection management, circuit breaking, load balancing

### 3. Policy Attachment

Separate policy resources attached to Backend resources.

```yaml
apiVersion: networking.example.com/v1
kind: RetryPolicy
metadata:
  name: openai-retry
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Backend
    name: openai-api
  retries: 3
  backoff: exponential
```

**Use cases**: Retry policies, observability configuration, vendor-specific policies

### Protocol Extensions

The Backend resource supports vendor-specific protocols through a structured extension mechanism:

```yaml
spec:
  destination:
    type: FQDN
    fqdn:
      hostname: anthropic.ai
      protocol:
        type: "anthropic.io/claude-api"
        options:
          version: "2023-06-01"
          streaming: "true"
```

<!-- NOTE: It's unclear to me if we want to start off with such a sprawling vendor extension model from the get-go. -->
<!-- Maybe, initially, we only include standard protocols like MCP and disallow vendor prefixed protocols at first. -->
**Vendor Prefix Pattern**: Following Gateway API conventions, implementations can define custom protocols using vendor prefixes (e.g., `istio.io/grpc-web`, `linkerd.io/profile`).

## Alternatives Considered

### Cluster-Scoped Backend Resource

Cluster-scoped Backend resources were considered but rejected due to:
- **Management complexity**: Requires coordination between cluster admins and app developers
- **Incorrect Persona Alignment**: Application developers are the primary consumers of backend resources, and they typically operate within namespace boundaries

### Policy-Only Approach

Using only policy attachment without a dedicated Backend resource was considered but rejected due to:
- **Destination representation gap**: No clear way to represent external FQDNs without synthetic Services
- **Policy target ambiguity**: Policies would still need to target synthetic Services
- **Extension limitations**: Protocol and connection options don't fit policy patterns well
