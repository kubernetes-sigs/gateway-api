# GEP-XXX: Backend Resource for Service Decoration

* Issue: [#XXXX](https://github.com/kubernetes-sigs/gateway-api/issues/XXXX)
* Status: Provisional

## TLDR

This GEP proposes a new `Backend` resource to address limitations in external destination representation, policy application, and security concerns with the current Service-based backend system. The namespace-scoped `Backend` resource serves as a consumer-focused alternative to synthetic Services for external destinations (FQDNs, IPs) and provides a foundation for service decoration and enhanced policy application.

## Motivation

The Kubernetes `Service` resource conflates two distinct concerns that have become increasingly problematic as Gateway API adoption grows:

1. **Frontend concerns**: How services are discovered and called (DNS names, ClusterIPs, service discovery)
2. **Backend concerns**: Where traffic should be routed and how to connect to destinations (endpoints, TLS configuration, protocol settings)

This conflation creates several significant issues:

### External Destination Limitations

Currently, representing external destinations in Gateway API requires synthetic `Service` objects with `type: ExternalName`, which:

- **Security vulnerabilities**: ExternalName Services are subject to DNS rebinding attacks ([CVE-2021-25740](https://github.com/kubernetes/kubernetes/issues/103675))
- **Policy limitations**: Cannot apply backend-specific policies (TLS, authentication, rate limiting) without affecting all consumers
- **Synthetic resource overhead**: Creates artificial Kubernetes resources for external dependencies
- **Configuration complexity**: Mixing internal and external destinations in the same resource type

### Policy Application Complexity

The overloaded Service resource makes policy attachment ambiguous:

- **Producer vs Consumer confusion**: Services represent both "how to call me" (producer) and "where to route" (consumer) concerns
- **Policy scope unclear**: TLS policies applied to Services affect all consumers, preventing per-route customization
- **Extension limitations**: Current `BackendTLSPolicy` only supports Service targets, limiting external destination configuration

### Gateway API Integration Friction

Gateway API's backend reference system defaults to Services but:

- **Limited extensibility**: Adding new backend types requires significant API changes
- **Inconsistent patterns**: Different implementations handle external destinations differently
- **Missing primitives**: No standard way to represent external FQDNs, IPs, or connection metadata

## Goals

- **Introduce Backend resource** as a namespace-scoped consumer resource for representing destinations and connection metadata
- **Support multiple destination types**: Kubernetes Services and external FQDNs
- **Enable backend-specific policies**: TLS configuration, authentication, health checks, and protocol settings per destination
- **Maintain Service compatibility**: Existing Service-based backends continue to work indefinitely
- **Establish extension framework**: Allow vendor-specific protocol support and connection options
- **Provide security model**: Clear RBAC patterns and risk documentation for external destinations

## Non-Goals

- **Deprecate or replace Services**: Services remain the primary backend type for internal destinations
- **Support producer-side policies**: Backend resource is explicitly consumer-focused
- **Standardize service mesh patterns**: Focus on Gateway API egress and decoration use cases
- **Provide cluster-scoped backends**: Backend resource is namespace-scoped for security boundaries
- **Address service discovery concerns**: Service discovery and DNS remain with existing Kubernetes Service system

## User Stories

### As an Application Developer

> "I want to configure my application to call external APIs (like OpenAI) with specific TLS settings and authentication without creating synthetic Services that expose security risks or affect other applications."

### As a Platform Engineer

> "I want to enforce that all external API calls go through specific gateways with proper logging and policy enforcement, without having to manage complex Service configurations for every external dependency."

### As a Security Administrator

> "I want to avoid ExternalName Services due to DNS rebinding vulnerabilities while still allowing applications to declare their external dependencies in a structured, auditable way."

## Proposal

The Backend resource addresses a fundamental gap in Gateway API's backend reference system: first-class support for external destinations. Currently, external APIs, databases, and services must be represented as synthetic Kubernetes Services with `type: ExternalName`, which creates security vulnerabilities, policy application challenges, and resource management overhead. Furthermore, `Backend`s of type `EndpointSelector` can be used to decorate existing `Service`s (or other resources that fulfill the backend role) with TLS and protocol configuration.

The Backend resource is explicitly designed as a **consumer resource** - it describes how a gateway should connect to a destination from the client perspective, regardless of whether that destination is internal or external to the cluster.

## API Specification

### Backend Resource Schema

```go
type Backend struct {
  metav1.TypeMeta   `json:",inline"`
  metav1.ObjectMeta `json:"metadata,omitempty"`
  Spec   BackendSpec   `json:"spec"`
  Status BackendControllerStatus `json:"status,omitempty"`
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
  Filters []BackendFilters `json:"filters,omitempty"`
}

type BackendType string

const (
  BackendTypeFQDN             BackendType = "FQDN"
  BackendTypeEndpointSelector BackendType = "EndpointSelector"
)

type BackendDestination struct {
  // Type defines the destination type
  Type BackendType `json:"type"`

  // Ports defines the destination ports and protocols
  // kubebuilder:validation:MinItems=1
  // kubebuilder:validation:MaxItems=16
  Ports []BackendPort `json:"ports,omitempty"`

  // FQDN specifies the configuration for an FQDN backend. Only used if type is FQDN.
  // +optional
  FQDN *FQDNBackend `json:"fqdn,omitempty"`

  // EndpointSelector specifies the configuration for an EndpointSelector backend. Only used if type is EndpointSelector.
  // TODO: Reference EndpointSelector GEP once added.
  // +optional
  EndpointSelector *EndpointSelectorBackend `json:"endpointSelector,omitempty"`
}

// BackendTLSMode defines the TLS mode for backend connections.
// +kubebuilder:validation:Enum=Simple;Mutual;None
type BackendTLSMode string

const (
  // Do not modify or configure TLS. If your platform (or service mesh)
  // transparently handles TLS, use this mode.
  BackendTLSModeNone BackendTLSMode = "None"
  // Enable TLS with simple server certificate verification.
  BackendTLSModeSimple BackendTLSMode = "Simple"
  // Enable mutual TLS.
  BackendTLSModeMutual BackendTLSMode = "Mutual"
)

type EndpointSelectorBackend struct {
  // SelectorRef specifies the reference to the EndpointSelector resource that manages the EndpointSlices for this backend.
  // +required
  SelectorRef *LocalObjectReference `json:"selectorRef"`
}

type BackendTLS struct {
  // Mode defines the TLS mode for the backend.
  // +required
  Mode BackendTLSMode `json:"mode"`

  // ClientCertificateRef defines the reference to the client certificate for mutual
  // TLS. Only used if mode is MUTUAL.
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

type BackendPort struct {
  // Number defines the port number of the backend.
  // +required
  // +kubebuilder:validation:Minimum=1
  // +kubebuilder:validation:Maximum=65535
  Number uint32 `json:"number"`
  // Protocol defines the protocol of the backend.
  // +required
  // +kubebuilder:validation:MaxLength=256
  Protocol BackendProtocol `json:"protocol"`
  // TLS defines the TLS configuration that a client should use when talking to the backend.
  // TODO: To prevent duplication on the part of the user, maybe this should be declared once at the
  // top level with per-port overrides?
  // +optional
  TLS *BackendTLS `json:"tls,omitempty"`
  // +optional
  ProtocolOptions *BackendProtocolOptions `json:"protocolOptions,omitempty"`
}
```

### FQDN Backend Configuration

```go
type FQDNBackend struct {
  // Hostname specifies the destination FQDN
  Hostname string `json:"hostname"`
}
```

### Protocol and Extension Support

```go

// BackendProtocol defines the protocol for backend communication.
// +kubebuilder:validation:Enum=HTTP;HTTP2;TCP;MCP
type BackendProtocol string

const (
  BackendProtocolHTTP  BackendProtocol = "HTTP"
  BackendProtocolHTTP2 BackendProtocol = "HTTP2"
  BackendProtocolTCP   BackendProtocol = "TCP"
  BackendProtocolMCP   BackendProtocol = "MCP"
)

// +kubebuilder:validation:ExactlyOneOf=mcp
type BackendProtocolOptions struct {
  // +optional
  MCP *MCPProtocolOptions `json:"mcp,omitempty"`
}

type MCPProtocolOptions struct {
  // MCP protocol version. MUST be a valid MCP version string
  // per the project's strategy: https://modelcontextprotocol.io/specification/versioning
  // +optional
  // +kubebuilder:validation:MaxLength=256
  Version string `json:"version,omitempty"`
  // URL path for MCP traffic. Default is /mcp.
  // +optional
  // +kubebuilder:default:=/mcp
  Path string `json:"path,omitempty"`
}
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
   - Trivial for attackers to register FQDNs that resolve to internal addresses
     - Because of this, implementations implementing `Backend` MUST add either TLS or JWT validation on sensitive localhost endpoints to prevent confused deputy attacks
   - Restrictive validation would break legitimate external integrations
   - Security focus should be on network-level controls, not resource-level validation
   - No initial support for wildcard FQDNs to limit attack surface and the need for Dynamic Forward Proxy support from implementations
     - Future proposals to add this functionality should include a comprehensive DNS trust specification and threat model.
   - Implementations may also be able to implement data-plane/proxy-level protections for common attack vectors
   - NOTE: This may be a decision we revisit in the future based on user feedback

#### Example of a Recommended Network Policy

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
  filters:
  - type: ExtensionRef
    extensionRef:
      name: connection-pool
      kind: ConnectionPoolPolicy
  - type: ExtensionRef
    extensionRef:
      name: circuit-breaker
      kind: CircuitBreakerPolicy
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

## Graduation Criteria

### Alpha
- [ ] Backend resource CRD with full schema validation
- [ ] Reference implementation in at least one Gateway API implementation
- [ ] Basic conformance tests for FQDN and Service destination types
- [ ] Documentation and examples for common use cases

### Beta
- [ ] Multiple Gateway API implementations support Backend resource
- [ ] Comprehensive conformance test suite
- [ ] Performance testing with external destinations
- [ ] Security review and RBAC documentation
- [ ] Extension framework validation with vendor implementations

### GA
- [ ] At least 3 implementations with production usage
- [ ] Extended conformance testing covering edge cases
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
