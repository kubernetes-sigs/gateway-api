# GEP-3779: Identity Based Authz for East-West Traffic

* Issue: [#3779](https://github.com/kubernetes-sigs/gateway-api/issues/3779)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)


## TLDR

Provide a method for configuring Gateway API Mesh implementations to enforce east-west identity-based Authorization controls. At the time of writing this we leave Authentication for specific implementation and outside of this proposal scope.


## Goals

(Using the [Gateway API Personas](../../concepts/roles-and-personas.md))

* A way for Ana the Application Developer to configure a Gateway API for Mesh implementation to enforce authorization policy that **allows** or **denies** identity or multiple identities to talk with some set of the workloads she controls.

* A way for Chihiro, the Cluster Admin, to configure a Gateway API for Mesh implementation to enforce non-overridable cluster-wide, authorization policies that **allows** or **denies** identity or multiple identities to talk with some set of the workloads in the cluster.

* A way for both Ana and Chihiro to restrict the scope of the policies they deploy to specific ports.

## Stretch Goals

* A way for Chihiro, the Cluster Admin, to configure a Gateway API for Mesh implementation to enforce default, overridable, cluster-wide, authorization policies that **allows** or **denies** identity or multiple identities to talk with some set of the workloads in the cluster.

## Non-Goals

* Support identity based authorization for north-south traffic or define the composition with this API.

## Deferred Goals

* (Potentially) Support enforcement on attributes beyond identities and ports.

## Introduction

Authorization is positioned as one of core mesh values. Every mesh supports some kind of east/west authorization between the workloads it controls.

Kubernetes core provides NetworkPolicies as one way to do it. Network Policies however falls short in many ways including:

* Network policies leverage labels as identities.
  * Labels are mutable at runtime. This opens a path for escalating privileges
  * Most implementations of network policies translate labels to IPs, this involves an eventual consistency nature which can and has lea to over permissiveness in the past.

* Scale. Network Policies are enforced using IPs (different selectors in the APIs get translated to IPs). This does not scale well with large clusters or beyond a single cluster

An identity-based authorization API is essential because it provides a structured way to control authorization between identities within the cluster.

### State of the World


| Aspect | Istio | Linkerd | Cilium |
| ----- | ----- | ----- | ----- |
| **Policy CRDs** | `AuthorizationPolicy` (APIs `security.istio.io/v1`) | `AuthorizationPolicy` (CRD `policy.linkerd.io/v1alpha1`), plus supporting CRDs (`Server`, `HTTPRoute`, `MeshTLSAuthentication`) | `CiliumNetworkPolicy` and `CiliumClusterwideNetworkPolicy`(`cilium.io/v2`) (superset of K8s NetworkPolicy) |
| **Identity model** | Identities derived from mTLS peer certificates (bound to SA): <ul><li>SPIFFE-like principal `<trust-domain>/ns/<namespace>/sa/<serviceaccount>`. </li> <li>ServiceAccount name </li> <li>Namespaces</li></ul></br> identity within JWT derived from `request.auth.principal`<br/><br/>IPBlocks and x-forwarded-for ipBlocks | Identities derived from mTLS peer certificates (bound to SA trust domain `identity.linkerd.cluster.local`. Policies reference service accounts or explicit mesh identities (e.g. `webapp.identity.linkerd.cluster.local`). <br/><br/>Policies use `requiredAuthenticationRefs` to reference the entities who get authorization. This is a list of targetRefs and it can include: <ul><li>ServiceAccounts</li> <li>`MeshTLSAuthentication` - which represents a set of mesh identities either with a mesh identities strings or reference to serviceAccounts</li> <li>`NetworkAuthentication` - represents sets of IPs or subnets.</li></ul>  | TODO(liorlieberman) |
| **Enforcement** | For Istio with sidecars - a proxy on each pod. For ambient, ztunnel node agent enforces mTLS based L4 authorization, L7 authorization is being enforced in waypoints if any. <br/><br/> Istio supports ALLOW, DENY, CUSTOM (often used for external authorization), and AUDIT. DENY policies in istio's context are used to enforce higher priority deny policies. The allow semantics is that whatever is not allowed explicitly (and assuming there is any policy for the same match) is implicitly denied  | Linkerd data-plane proxy (injected into each pod). The proxy enforces policies via mTLS identity checks. <br/><br/> Linkerd supports AUDIT and ALLOW. There is not DENY policies, whats not allowed (and assuming there is any policy for the same match) is implicitly denied. | TODO(liorlieberman) |
| **Request Match criteria** | Fine-grained L7 and L4 matching: HTTP/gRPC methods, paths, headers, ports, SNI, etc., plus source identity (namespace, service account). Policies use logical OR over rules.<br/><br/> All match criterias are inline in the policy. See [https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-To](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-To) and [https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-when](https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-when)  | Policies can target: <ul><li>A `Server` which describes a set of pods (using fancy label match expressions), and a single port on those pods.</li> <li>A user can optionally restrict the authorization to a smaller subset of the traffic by targeting an HTTPRoute. (TODO: any plans to support sectionNames?)</li> <li> A namespace - this indicates that the policy applies to all traffic to all Servers and HTTPRoutes defined in the namespace.</li></ul> Note: We leave `ServerAuthorization` outside the scope as it planned to be deprecated (per linkerd website)  | TODO(liorlieberamn) |
| **Default policies and admin policies** | If **no** ALLOW policy matches, traffic is **allowed** by default. You can deploy an overridable - default deny by default by deploying an **allow-nothing** policy in either the namespace or istio-system <br/><br/>AuthorizationPolicies in the `istio-system` namespace apply to the whole mesh and take precedence. These are not overridable by namespace-level policies.  | Default inbound policy can be set at install time using `proxy.defaultInboundPolicy`. Supported values are: <ul><li>`all-unauthenticated:` allow all traffic. This is the default.</li>  <li>`all-authenticated:` allow traffic from meshed clients in the same or from a different cluster (with multi-cluster).</li>  <li>`cluster-authenticated:` allow traffic from meshed clients in the same cluster.</li>  <li>`cluster-unauthenticated:` allow traffic from both meshed and non-meshed clients in the same cluster.</li>  <li>`deny:` all traffic are denied. </li> <li>`audit:` Same as all-unauthenticated but requests get flagged in logs and metrics.</li> </ul> <br/>Users can override the default policies for namespaces/pods or by setting the [config.linkerd.io/default-inbound-policy](http://config.linkerd.io/default-inbound-policy) annotation There is no support for admin, non overridable policies. | TODO(liorlieberman)|
| **YAML Configuration** | Istio `AuthorizationPolicy` CRD. E.g., to **allow** a namespace or service-account to call a service: |  |  |


Every mesh vendor has their own API of such authorization. Below we describe the UX for different implementations:

#### Istio
Link: [Istio authorization policy docs](https://istio.io/latest/docs/reference/config/security/authorization-policy/)

TODO

#### Linkerd

Link: [Linkerd authorization policy docs](https://linkerd.io/2-edge/reference/authorization-policy/)

TODO

#### Cilium

TODO 

## API



## Conformance Details


#### Feature Names


### Conformance tests 


## Alternatives


## References