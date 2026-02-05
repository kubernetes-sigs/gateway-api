# GEP-4488: Backend Resource for Egress and Service Decoration

* Issue: [#4488](https://github.com/kubernetes-sigs/gateway-api/pull/4488)
  * Incubated by the [AI Gateway Working Group](https://github.com/kubernetes-sigs/wg-ai-gateway/pull/20)
* Status: Provisional

## TLDR

This GEP proposes a new `Backend` resource to address the fact that Service is a mature and stable, if complex API. It is useful to be able to add Gateway-specific behaviors to Service, but previous approaches (like BackendTLSPolicy) have significant limltations, particuarly around discoverability and implementation complexity.

In particular, AI use cases have an urgent need to be able to better handle external destination representation, policy application, and security concerns than the current Service-based backend system can manage.

The namespace-scoped `Backend` resource serves in this use case as a consumer-focused alternative to synthetic Services for external hosts and provides a foundation for service decoration and enhanced policy application, as a first use for this new resources.

Further use of the `Backend` resource for other configuration that is tightly bound to Service is expected to follow.

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

### 2. Policy Attachment

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

A third extension mechanism, filters applied on the `Backend` resource itself, may be the subject of a future GEP.

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
