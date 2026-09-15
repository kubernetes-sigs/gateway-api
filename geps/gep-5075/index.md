---
title: "GEP-5075: Support ClusterTrustBundle as a valid caCertificateRef"
---

* Issue: [#5075](https://github.com/kubernetes-sigs/gateway-api/issues/5075)
  * Related to [#1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897)
  * Related to but distinct from [#3787](https://github.com/kubernetes-sigs/gateway-api/issues/3787)
* Status: Provisional

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
2.  **Support Label Selector Resolution:** Allow selecting different ClusterTrustBundles using label selectors and `signerName` matching, leveraging all native selection features.
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
*   **Default ServiceAccount Access:** Kubernetes grants default read permissions (`get`, `list`, `watch`) on ClusterTrustBundles to all authenticated users and all ServiceAccounts in the cluster via default system ClusterRoles.
*   **Alignment with Security Boundary:** Because any pod in a namespace can already read these bundles through standard projected volumes, referencing them in a `BackendTLSPolicy` or `Gateway` listener does not introduce new attack vectors or leak private data. It aligns perfectly with standard cluster RBAC models.

## 6. Next Steps & Graduation Criteria

This GEP is proposed as **Provisional** to establish community alignment on goals and schemas. Upon approval, we will graduate the GEP to **Implementable** by delivering:

1.  API Schema modifications in the experimental release channel.
2.  Integration test suites verifying controller resolution of explicit name references.
3.  Comprehensive conformance tests ensuring correct TLS handshake failures when invalid `ClusterTrustBundle` references are provided.
