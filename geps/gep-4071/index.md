# GEP-4071: Implementation-Specific Usage of BackendTLSPolicy for Any Kind of Service

* Issue: [#4071](https://github.com/kubernetes-sigs/gateway-api/issues/4071)
* Status: Provisional

## TLDR - WHAT

The ability for `BackendTLSPolicy` to be used as an Implementation-Specific feature
to configure TLS for any Service (or other resource), regardless of whether it is
referenced by a Route resource.

## Motivation - WHY

### Current State

BackendTLSPolicy, as defined in [GEP-1897](https://gateway-api.sigs.k8s.io/geps/gep-1897/),
was designed specifically for the HTTPRoute -> Service use case, where a Route explicitly
references a backend Service. The policy can only be applied to Services that are
actively referenced as `backendRefs` in HTTPRoute resources.

### Problem Statement

Gateway implementations now support features that require TLS communication with
backend services that are **not** referenced by any Route resource. These include:

* **Auxiliary features**: Services consumed by Gateway features like the proposed
  ExternalAuthFilter, which may expose TLS listeners
* **Service mesh scenarios**: Workload-to-service communication where Routes
  may not be defined, but TLS validation requirements still exist

Currently, these use cases cannot leverage BackendTLSPolicy because:

1. The services are not referenced in any Route's `backendRefs`
2. GEP-1897 explicitly scoped BackendTLSPolicy to HTTPRoute scenarios only
3. No clear guidance exists for using BackendTLSPolicy outside of Route contexts
4. Implementations are unclear whether they can use BackendTLSPolicy for non-Route scenarios

### Proposed Solution

This GEP proposes clarifying that BackendTLSPolicy can be used as an
**Implementation-Specific** feature, allowing implementations to apply TLS
configuration to any Service (or other resource types), regardless of whether
that Service is referenced by a Route.

This approach:

* Enables consistent TLS configuration using the existing BackendTLSPolicy API
* Avoids vendor-specific API proliferation
* Supports Gateway features requiring backend TLS without Route association
* Provides a path for service mesh implementations to standardize on Gateway API constructs

### Example Use Case

A Gateway implementation supporting the proposed ExternalAuthFilter feature needs
to communicate with an external authentication service over TLS. This authentication
service:

* Requires proper CA verification and hostname validation
* Should use the same TLS configuration patterns as HTTPRoute backends

With this GEP, implementations can apply BackendTLSPolicy to configure this TLS
communication.

Another example is a service mesh implementation that wants to configure TLS settings
for workload-to-service communication. The mesh may not use Route resources at all,
but still needs to specify CA certificates, hostname validation, and SNI settings
for backend connections.

## User Stories - WHO

* As an application developer, I want to use an [ExternalAuthFilter](../../reference/spec/#httpexternalauthfilter)
  that is exposed with a TLS listener, specifying the appropriate Certificate
  Authorities and hostname validation settings, even though the authentication
  service is not part of my HTTPRoute backendRefs.

* As a platform operator, I want to configure TLS settings for infrastructure
  services (monitoring, logging, tracing, rate-limiting) that my Gateway
  communicates with, without requiring these services to be exposed via Routes.

* As an application developer in a service mesh, I want to signal to my mesh
  provider that clients consuming my service should use specific certificate
  authorities and SAN hostname validation when reaching my service, even when
  no Route resource exists for that service.

* As an application developer in a service mesh, I want to signal that when my
  workload reaches a service (or another type of resource), it should validate
  the Certificate Authorities using the existing system certificate chains.
  * **Stretch Goal**: BackendTLSPolicy currently uses `LocalPolicyTargetReferenceWithSectionName`
    as `targetRefs` and would need to be changed to `NamespacedPolicyTargetReference`
    to support cross-namespace scenarios. This is deferred to a future GEP.

* As a mesh implementation developer, I want to use the already existing
  BackendTLSPolicy API definition to allow my users to specify TLS features of
  services without needing to develop a new custom API.

## Goals

* Enable BackendTLSPolicy to be used as an **Implementation-Specific** feature
  for configuring TLS to Services (or other resources) that are **not** referenced
  by any xRoute resource.

* Clarify that implementations MAY use BackendTLSPolicy for infrastructure and
  auxiliary services required by Gateway features (e.g., ExternalAuth, rate
  limiting, logging) without requiring Route associations.

* Provide clear guidance in the BackendTLSPolicy specification about when and how
  implementations can use this resource outside of the HTTPRoute -> Service pattern
  defined in GEP-1897.

* Enable connecting to other services in the cluster over TLS without requiring
  the service to be specified in an xRoute resource.

## Non-Goals

* **NOT** adding GRPCRoute, TLSRoute, TCPRoute, or UDPRoute support in this GEP
* **NOT** enabling cross-namespace targetRefs in this GEP (deferred to future work; see Long-term Goals)
* **NOT** requiring implementations to support non-Service target types
* **NOT** changing the fundamental API structure of BackendTLSPolicy

## Long-term Goals

* Enable BackendTLSPolicy to target services or other resources in a different
  namespace from the policy's namespace. This would require changing `targetRefs`
  from `LocalPolicyTargetReferenceWithSectionName` to `NamespacedPolicyTargetReference`,
  which may be a breaking change requiring a separate GEP.

* Consider future GEPs for adding BackendTLSPolicy support to GRPCRoute, TLSRoute,
  and other route types.

### What GEP-4071 Adds

* **Clarity** that BackendTLSPolicy CAN be used outside of Route contexts as an Implementation-Specific feature
* **Guidance** for implementations using BackendTLSPolicy for infrastructure services
* **Support level designation** (Implementation-Specific) for non-Route scenarios
* **Documentation** clarifying when and how this usage is appropriate

## API

### Changes Required

#### 1. Documentation Clarification in BackendTLSPolicy Specification

Update the BackendTLSPolicy API documentation in the `BackendTLSPolicySpec` type
to clarify Implementation-Specific usage scenarios.

**Current**:
```go
// TargetRefs identifies an API object to apply the policy to.
// Only Services have Extended support. Implementations MAY support
// additional objects, with Implementation Specific support.
```

**Proposed Enhancement**:
```go
// TargetRefs identifies an API object to apply the policy to.
//
// Support Levels:
//
// * Extended: Kubernetes Service referenced by HTTPRoute backendRefs.
//   This is the primary use case defined in GEP-1897.
//
// * Implementation-Specific: Any Service or other resource, regardless of
//   Route association. Implementations MAY use BackendTLSPolicy for:
//   - Services not referenced by any Route (e.g., infrastructure services)
//   - Gateway feature backends (e.g., ExternalAuth, rate-limiting services)
//   - Service mesh workload-to-service communication
//   - Other resource types beyond Service
//
// When applied to Services or resources not referenced by Routes, the behavior
// is implementation-specific. Implementations SHOULD clearly document how
// BackendTLSPolicy is interpreted in these scenarios, including:
//   - Which resources beyond Service are supported
//   - How the policy is discovered and applied
//   - Any implementation-specific semantics or restrictions
```

#### 2. Update GEP-1897 Documentation

Add a reference in GEP-1897's "Future plans" section pointing to GEP-4071:

```markdown
### Implementation-Specific Usage (GEP-4071)

While GEP-1897 scoped BackendTLSPolicy to the HTTPRoute -> Service use case,
GEP-4071 clarifies that implementations MAY use BackendTLSPolicy for additional
scenarios as an Implementation-Specific feature, such as infrastructure services,
Gateway feature backends, or service mesh communication. See GEP-4071 for details.
```

#### 3. Update Gateway API TLS Guide

Update the TLS guide (site-src/guides/tls.md) to mention Implementation-Specific usage:

```markdown
!!! note "Implementation-Specific Usage"

    While BackendTLSPolicy has Extended support for Services referenced by HTTPRoute,
    implementations MAY support additional scenarios as Implementation-Specific features:

    - Services not referenced by any Route (e.g., infrastructure services)
    - Gateway feature backends (e.g., external authentication services)
    - Service mesh workload-to-service communication
    - Other resource types beyond Kubernetes Service

    Consult your implementation's documentation for details on supported scenarios.
    See GEP-4071 for more information.
```

### No API Changes Required

This GEP does **not** require any changes to the BackendTLSPolicy API structure:

* `targetRefs` remains `LocalPolicyTargetReferenceWithSectionName` (same namespace only)
* `Validation` field structure is unchanged
* Status conditions remain as defined in GEP-1897
* No new fields are added

### Future API Changes (Out of Scope for This GEP)

The following changes are recognized as potentially necessary but are deferred
to future GEPs:

* Change `targetRefs` from `LocalPolicyTargetReferenceWithSectionName` to
  `NamespacedPolicyTargetReference` to enable cross-namespace references
  * This is a potentially breaking change requiring careful consideration
  * See Long-term Goals and GEP-1897 Future Plans

* Add support for GRPCRoute, TLSRoute, TCPRoute backends

## Implementation Guidance

### For Gateway API Implementations

Implementations choosing to support Implementation-Specific usage of BackendTLSPolicy SHOULD:

1. **Specify discovery mechanisms** for how BackendTLSPolicy is discovered and applied
2. **Define precedence rules** if multiple policies could apply to the same backend
3. **Report status consistently** using the standard BackendTLSPolicy status conditions
4. **Respect all validation semantics** defined in GEP-1897 (CA certificates, hostname validation, etc.)

### For Service Mesh Implementations

Service mesh implementations MAY use BackendTLSPolicy to configure TLS for
workload-to-service communication, even when no Route resources exist. Mesh
implementations SHOULD:

1. Document how BackendTLSPolicy integrates with mesh TLS/mTLS semantics
2. Clarify whether BackendTLSPolicy is applied to client-side, server-side, or both
3. Explain interaction with mesh transport security (e.g., automatic mTLS)
4. Define behavior when mesh-specific TLS configuration conflicts with BackendTLSPolicy

### Conformance

Implementation-Specific usage is **not** subject to conformance testing. Only the
"Extended" support level (HTTPRoute → Service) is covered by Gateway API conformance tests.


## References

* [GEP-1897: BackendTLSPolicy - Explicit Backend TLS Connection Configuration](https://gateway-api.sigs.k8s.io/geps/gep-1897/)
* [Gateway API Policy Attachment (GEP-713)](https://gateway-api.sigs.k8s.io/geps/gep-713/)
* [Gateway API TLS Guide](https://gateway-api.sigs.k8s.io/guides/tls/)
* [Issue #4071](https://github.com/kubernetes-sigs/gateway-api/issues/4071)
