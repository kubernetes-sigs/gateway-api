---
title: "GEP-5075: Support ClusterTrustBundle as a valid caCertificateRef"
---

* Issue: [#5075](https://github.com/kubernetes-sigs/gateway-api/issues/5075)
  * Related to [#1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897)
  * Related to but distinct from [#3787](https://github.com/kubernetes-sigs/gateway-api/issues/3787)
* Status: Experimental

[Ana]: /docs/concepts/roles-and-personas/#ana
[Chihiro]: /docs/concepts/roles-and-personas/#chihiro

(See [status definitions](../overview.md#gep-states).)

## 1. What

This Gateway Enhancement Proposal (GEP) introduces native support for `ClusterTrustBundle` resources within Gateway API's TLS validation fields, specifically targeting `caCertificateRefs` in both `BackendTLSPolicy` and `Gateway` specs. By integrating with `certificates.k8s.io` cluster-scoped resources, this proposal provides a robust mechanism to distribute and consume platform-managed Certificate Authority (CA) trust anchors.

## 2. Why

In Kubernetes clusters, validating backend TLS connections or implementing client certificate authentication on ingress listeners requires the validation engine (the Gateway datapath) to reference a trusted set of Certificate Authorities (CAs) - often referred to as trust anchors.

Currently, Gateway API fields like `caCertificateRefs` (in `BackendTLSPolicy` and `Gateway`) only offer "Core" support for local, namespace-scoped `ConfigMap` resources. Consequently, for cluster-scoped CAs, cluster administrators must replicate CA certificate PEM bundles across all tenant namespaces. This pattern introduces several key operational pain points:

*   **Administrative Toil:** Replicating CA bundles across hundreds of user namespaces requires custom operators, synchronization scripts, or manual intervention, which are fragile and prone to failure.
*   **Security Risks:** Stale CA certificates in outdated namespaces can lead to validation failures, application outages, or accidental trust of insecure endpoints if root rotation is out of sync.
*   **Consumability Barriers:** Application developers should not be responsible for importing, updating, or validating corporate root CAs. They should simply consume a platform-provided, highly available source of trust.
*   **Control Plane Overhead:** Duplicating large `ConfigMap` objects - which can expand to hundreds of kilobytes - across all user namespaces creates significant resource pressure and degrades performance on both etcd and the Kubernetes API server.

With the graduation of `ClusterTrustBundle` (`certificates.k8s.io/v1`) to General Availability in Kubernetes 1.37, the platform now provides a native, cluster-scoped container for X.509 trust anchors. Supporting this resource in Gateway API resolves these issues by allowing namespace-scoped resources to safely reference global, cluster-wide trust anchors.

## 3. Who

### Ana (The Application Developer)

*   **User Story:** As [Ana], the application developer, I want to reference a pre-existing, platform-managed `ClusterTrustBundle` in my `BackendTLSPolicy` so that my upstream connections validate correctly without requiring local namespace-scoped ConfigMaps.

### Chihiro (The Cluster Operator / Platform Engineer)

*   **User Story:** As [Chihiro], the cluster operator, I want to deploy a single `ClusterTrustBundle` at the cluster level during bootstrap, enabling all tenant namespaces to dynamically and securely resolve valid CAs for TLS handshakes.
*   **User Story:** As [Chihiro], the cluster operator, I want to configure frontend client certificate validation across multiple Gateway instances referencing a centralized trust anchor, simplifying operational administration for mTLS authentication.

## 4. Goals and Non-Goals

### Goals

1.  **Allow Single ClusterTrustBundle as Source of Truth:** Enable using `ClusterTrustBundle` as a cluster-scoped source of truth for `caCertificateRefs` across namespaces.
2.  **Support Cluster-Scoped Trust Anchor Reference:** Allow referencing a `ClusterTrustBundle` by name from `BackendTLSPolicyValidation` and Gateway frontend TLS validation, using a dedicated cluster-scoped reference type.
3.  **Simplify Trust Management:** Reduce administrative overhead and avoid synchronization issues by offering native references to cluster-scoped trust anchors.
4.  **Reduce Control Plane Overhead:** Avoid duplicating large `ConfigMap` objects—which can expand to hundreds of kilobytes—across user namespaces to prevent resource pressure and performance degradation on etcd and the Kubernetes API server.
5.  **Ensure RBAC and Boundary Safety:** Leverage `ClusterTrustBundle` to provide secure, robust multi-tenant isolation and boundary safety when referencing cluster-scoped trust objects from namespace-scoped policies.

### Non-Goals

6.  **Managing Private Keys in a Cluster-Scoped Way:** `ClusterTrustBundle` objects carry public certificate PEM blocks only. Private key handling remains strictly in namespace-scoped `Secret`s.
7.  **Defining CA Rotation Engines:** This GEP does not prescribe mechanisms for rotating or issuing CA certificates, focusing solely on consumption via API objects.
8.  **Deprecating Namespaced ConfigMaps:** We do not propose deprecating or removing support for using namespaced ConfigMaps as `caCertificateRefs`.

## 5. Key Considerations & Guardrails

### Feature Support Graduation & Upstream Compliance

To maintain API stability, Gateway API dictates that dependencies on core Kubernetes features must adhere to release-graduation timelines:

*   A feature must reside in Kubernetes GA for at least 5 consecutive releases before Gateway API can mandate its support as a Core level feature.
*   Because `ClusterTrustBundle` graduates to GA in Kubernetes 1.37, this feature must initially be designated as Extended support.
*   Gateway implementations running on older cluster versions or environments lacking cluster-scoped lookup engines are not required to implement this GEP to maintain compliance.

### RBAC and Namespace Isolation Safety

A key question during community GEP reviews is whether referencing a cluster-scoped resource from a namespace-scoped resource introduces security bypasses or privilege escalation.

*   **Globally Readable by Design:** According to the core Kubernetes API design, `ClusterTrustBundle` resources are intended to be globally readable.
*   **Default ServiceAccount Access:** Kubernetes grants default read permissions (`get`, `list`, `watch`) on ClusterTrustBundles to all authenticated users and all ServiceAccounts in the cluster via the `system:cluster-trust-bundle-discovery` ClusterRoleBinding.
*   **Alignment with Security Boundary:** Because any pod in a namespace can already read these bundles through standard projected volumes, referencing them in a `BackendTLSPolicy` or `Gateway` listener does not introduce new attack vectors or leak private data. It aligns perfectly with standard cluster RBAC models.
*   **Hardened Cluster Caveat:** Clusters that remove or restrict the default `system:cluster-trust-bundle-discovery` ClusterRoleBinding may prevent the Gateway controller's ServiceAccount from reading `ClusterTrustBundle` resources. Implementations SHOULD document any RBAC prerequisites required for this feature to function.

## 6. Technical Design & API Changes

This GEP introduces a dedicated `ClusterTrustBundleObjectRef` type for cluster-scoped `ClusterTrustBundle` resources and a new optional `clusterTrustBundleRef` field on the relevant validation structs. Using a named type distinct from `LocalObjectReference` makes the cluster-scope semantics explicit and avoids ambiguity.

### New Type: `ClusterTrustBundleObjectRef`

```go
// ClusterTrustBundleObjectRef identifies a ClusterTrustBundle.
//
// The Group defaults to "certificates.k8s.io" and the Kind defaults to
// "ClusterTrustBundle", so users only need to specify the name.
type ClusterTrustBundleObjectRef struct {
    // Group is the group of the referent.
    //
    // Defaults to "certificates.k8s.io".
    //
    // +optional
    Group *Group `json:"group,omitempty"`

    // Kind is the kind of the referent.
    //
    // Defaults to "ClusterTrustBundle".
    //
    // +optional
    Kind *Kind `json:"kind,omitempty"`

    // Name is the name of the referent.
    // +required
    Name ObjectName `json:"name"`
}
```

A `ReferenceGrant` is not required because `ClusterTrustBundle` is cluster-scoped and has no target namespace. By default, Kubernetes grants read access to `ClusterTrustBundle` resources to all authenticated users via the `system:cluster-trust-bundle-discovery` ClusterRoleBinding.

### API Changes

A new optional `clusterTrustBundleRef` field is added to `BackendTLSPolicyValidation` and `FrontendTLSValidation`. The existing `caCertificateRefs` field is unchanged. When both fields are set, their trust anchors are unioned.

#### `BackendTLSPolicyValidation` (apis/v1/backendtlspolicy_types.go)

```go
type BackendTLSPolicyValidation struct {
    // ... existing fields unchanged ...

    // ClusterTrustBundleRef is an optional reference to a cluster-scoped
    // ClusterTrustBundle (certificates.k8s.io/v1) resource.
    //
    // Support: Extended
    //
    // <gateway:experimental>
    // +optional
    ClusterTrustBundleRef *ClusterTrustBundleObjectRef `json:"clusterTrustBundleRef,omitempty"`
}
```

The struct-level XValidation rule is updated to accept `clusterTrustBundleRef` as a valid trust source alongside `caCertificateRefs` and `wellKnownCACertificates`.

#### `FrontendTLSValidation` (apis/v1/gateway_types.go)

```go
type FrontendTLSValidation struct {
    // ... existing fields unchanged ...

    // ClusterTrustBundleRef is an optional reference to a cluster-scoped
    // ClusterTrustBundle (certificates.k8s.io/v1) resource.
    //
    // Support: Extended
    //
    // <gateway:experimental>
    // +optional
    ClusterTrustBundleRef *ClusterTrustBundleObjectRef `json:"clusterTrustBundleRef,omitempty"`
}
```

For the Experimental phase, this GEP is intentionally scoped to explicit name-based references to `ClusterTrustBundle`. Selector-based discovery is out of scope for this proposal.

### Example Usage

#### Cluster-Scoped Trust Anchor Definition

```yaml
apiVersion: certificates.k8s.io/v1
kind: ClusterTrustBundle
metadata:
  name: example.com:internal-signer:v1
spec:
  signerName: example.com/internal-signer
  trustBundle: |
    -----BEGIN CERTIFICATE-----
    MIIF6TCCA9GgAwIBAgIURX... (Corporate Root CA)
    -----END CERTIFICATE-----
```

#### Consumer Policy

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: BackendTLSPolicy
metadata:
  name: secure-inventory-validation
  namespace: dev-ana
spec:
  targetRefs:
    - group: ""
      kind: Service
      name: inventory-db
  validation:
    hostname: db.internal.example.com
    clusterTrustBundleRef:
      name: example.com:internal-signer:v1
```

### Support and Validation

`ClusterTrustBundle` name references via `clusterTrustBundleRef` have **Extended** support (Experimental). The existing Core support for a single namespaced `ConfigMap` via `caCertificateRefs` remains unchanged.

An implementation that supports this feature MUST:

1. Read `ClusterTrustBundle.spec.trustBundle` from the object named in `clusterTrustBundleRef` and use its PEM-encoded certificates as trust anchors for the relevant backend or frontend TLS validation.
2. Treat a nonexistent bundle, an unreadable bundle, a bundle whose `spec.trustBundle` cannot be parsed as a CA certificate bundle, or a bundle with an empty `spec.trustBundle` as an invalid CA certificate reference.
3. Set `ResolvedRefs=False` with reason `InvalidCACertificateRef` for an unresolved or malformed bundle, and MUST NOT use the bundle for TLS validation.
4. If `clusterTrustBundleRef` is the sole trust source and it is invalid, additionally set `Accepted=False` with reason `NoValidCACertificate`, consistent with the existing contract for `BackendTLSPolicyValidation` and `FrontendTLSValidation`.
5. Reconcile updates to the referenced `ClusterTrustBundle`, including changes to `spec.trustBundle`, deletion of the referenced object, and replacement or recreation of an object with the same name. During the interval between deletion and recreation, the implementation MUST treat the reference as invalid and MUST NOT use any previously cached trust anchors.
6. Not require a `ReferenceGrant` for a valid `clusterTrustBundleRef`.

Implementations that do NOT support this feature MUST set `ResolvedRefs=False` with reason `InvalidKind` (for `BackendTLSPolicy`) or `InvalidCACertificateKind` (for Gateway frontend TLS) when `clusterTrustBundleRef` is specified.

### API Availability

`ClusterTrustBundle` is a core Kubernetes API (`certificates.k8s.io/v1`, GA in Kubernetes v1.37). On clusters where the API is not available — e.g., Kubernetes versions older than v1.30, or clusters with the `ClusterTrustBundle` feature gate disabled — implementations MUST:

1. Detect availability at startup and at periodic intervals by probing the discovery document (`GET /api`) for the `certificates.k8s.io` group, rather than by issuing client requests that would fail.
2. Report support dynamically: the associated conformance feature MUST be advertised only when the API is present, so that the reported feature set reflects the cluster, not just the binary.
3. If a resource with `clusterTrustBundleRef` is configured while the API is unavailable, set `ResolvedRefs=False` with reason `InvalidKind` (the kind is unresolvable) and, if it is the sole trust source, `Accepted=False` with reason `NoValidCACertificate`. Existing workloads using `caCertificateRefs` MUST continue to function, and the controller MUST NOT crash, fail to start, or drop reconciliation of other resources.

This keeps `InvalidKind` as the single reason for "kind not available" and `InvalidCACertificateRef` for "API present but the specific bundle is invalid", which lets users distinguish a cluster-provisioning problem from a misconfiguration.

## 7. Conformance

This is an Extended (Experimental) conformance feature.

### Feature Names

* `ClusterTrustBundle` — Extended (Experimental) feature indicating support for `clusterTrustBundleRef` in `BackendTLSPolicyValidation` and Gateway frontend TLS validation.

### Conformance Tests

| Description | Outcome | Feature |
|---|---|---|
| Resolve a named `ClusterTrustBundle` via `clusterTrustBundleRef` and use it for backend TLS validation. | `BackendTLSPolicy` MUST have `ResolvedRefs=True`. TLS handshake to backend MUST succeed. | `ClusterTrustBundle` |
| Reference a nonexistent `ClusterTrustBundle` via `clusterTrustBundleRef`. | `BackendTLSPolicy` MUST have `ResolvedRefs=False` with reason `InvalidCACertificateRef`. TLS handshake MUST fail. | `ClusterTrustBundle` |
| Configure `clusterTrustBundleRef` on a cluster where the `certificates.k8s.io` API is unavailable. | `ResolvedRefs=False` with reason `InvalidKind` MUST be set; other trust sources and unrelated resources MUST continue to be reconciled. | `ClusterTrustBundle` |
| Reference a `ClusterTrustBundle` with an empty or unparsable `spec.trustBundle`. | `BackendTLSPolicy` MUST have `ResolvedRefs=False` with reason `InvalidCACertificateRef` and `Accepted=False` with reason `NoValidCACertificate`. TLS handshake MUST fail. | `ClusterTrustBundle` |
| Update `spec.trustBundle` of a referenced `ClusterTrustBundle`. | Implementation MUST reconcile the change. Effective trust configuration MUST reflect the updated bundle after reconciliation. | `ClusterTrustBundle` |
| Delete the referenced `ClusterTrustBundle`. | `BackendTLSPolicy` MUST move to `ResolvedRefs=False` with reason `InvalidCACertificateRef`. Previously established TLS sessions MAY continue but new sessions MUST fail. | `ClusterTrustBundle` |
| Resolve a named `ClusterTrustBundle` via `clusterTrustBundleRef` for Gateway frontend client certificate validation. | All targeted HTTPS listeners MUST have `ResolvedRefs=True`. Frontend mTLS MUST succeed with a certificate signed by the bundle's CA. | `ClusterTrustBundle` |

## 8. Alternatives Considered

* **Copy the trust bundle to namespaced ConfigMaps.** This retains Core support but duplicates data and makes rotation the responsibility of every consuming namespace.
* **Reuse `caCertificateRefs` with `LocalObjectReference`.** Adding `ClusterTrustBundle` as an accepted `kind` in the existing `caCertificateRefs` field avoids adding a new field, but `LocalObjectReference` has no cluster-scope semantics in its name or documentation. Reviewers requested a distinct type name to make cluster-scope intent explicit.
* **Use selector-based discovery.** This could enable dynamic selection of trust bundles, but it introduces additional API surface and requires significantly more specification work around selection semantics, multi-match handling, reconciliation behavior, and conformance. This proposal leaves that to possible future work so the Experimental scope stays focused on explicit name-based references.
* **Use `ReferenceGrant`.** `ReferenceGrant` expresses consent by the owner of a target namespace. It cannot apply to a cluster-scoped resource and adds no protection beyond the Kubernetes RBAC already governing `ClusterTrustBundle` access.