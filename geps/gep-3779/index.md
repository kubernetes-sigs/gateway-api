# GEP-3779: Identity Based Authz for East-West Traffic

* Issue: [#3779](https://github.com/kubernetes-sigs/gateway-api/issues/3779)
* Status: Implementable

(See [status definitions](../overview.md#gep-states).)

## TLDR

Provide a method for configuring Gateway API Mesh implementations to enforce east-west identity-based Authorization controls. At the time of writing this we leave Authentication for specific implementation and outside of this proposal scope.

## Goals

(Using the [Gateway API Personas](../../concepts/roles-and-personas.md))

* A way for Ana the Application Developer to configure a Gateway API for Mesh implementation to enforce authorization policy that **allows** identity or multiple identities to talk with some set (could be namespace or more granular) of the workloads she controls.

* A way for both Ana and Chihiro to restrict the scope of the policies they deploy to specific ports.

## Non-Goals

* Support identity based authorization for north-south traffic or define the composition with this API.

* A way for Chihiro, the Cluster Admin, to configure a Gateway API for Mesh implementation to enforce non-overridable cluster-wide, authorization policies that **allows** or **denies** identity or multiple identities to talk with some set of the workloads in the cluster.

* A way for Chihiro, the Cluster Admin, to configure a Gateway API for Mesh implementation to enforce default, overridable, cluster-wide, authorization policies that **allows** identity or multiple identities to talk with some set of the workloads in the cluster.

## Deferred Goals

* Support enforcement on attributes beyond identities and ports.

## Introduction

Authorization is positioned as one of core mesh values. Every mesh supports some kind of east/west authorization between the workloads it controls.

Kubernetes core provides NetworkPolicies as one way to do it. Network Policies however falls short in many ways including:

* Network policies leverage labels as identities.
  * Labels are mutable at runtime. This opens a path for escalating privileges
  * Most implementations of network policies translate labels to IPs, this involves an eventual consistency nature which can and has led to over permissiveness in the past.

* Scale. Network Policies are enforced using IPs (different selectors in the APIs get translated to IPs). This does not scale well with large clusters or beyond a single cluster

An identity-based authorization API is essential because it provides a structured way to control authorization between identities within the cluster.

### Current AuthZ Support within Meshes

Istio, Linkerd, and Cilium all support identity-aware authorization via vendored policies, but differ in mechanics and philosophy. Istio and Linkerd rely on mTLS-derived identities tightly coupled with service accounts, while Cilium broadens the scope using BPF-based identities tied to labels, IPs, and SPIFFE. Istio offers policy layering (namespace/system-wide), including support for DENY and CUSTOM rules, enforced at sidecars or ztunnel/waypoints. Linkerd injects proxies, emphasizes mTLS, and supports ALLOW/AUDIT — there's no DENY. Cilium stands apart with L3–L7 policy enforcement (via kernel or envoy), and broader match targets including pod/node selectors and CIDRs. It also uniquely maps identities into the datapath ([#CiliumIdentity](#CiliumIdentity)), and supports explicit default-deny enforcement patterns. See [#state-of-the-world](#state-of-the-world) for more detailed comparison.

## API

This GEP introduces a new policy resource, `AuthorizationPolicy`, for **identity-based** authorization. The policy defines a target destination, an action, and a set of rules that include sources (the “who”) and attributes to limit the scope of the action.

This GEP does not define support for L7 authorization policy (see the [Future Enhancement](#future-enhancements) section for more information).

### Policy Rules

Each `AuthorizationPolicy` resource contains a list of rules. The policy action is applied if the request matches **any** rule in the list (logical OR). Each rule defines multiple matching criteria; a request matches a rule only if it matches **all** criteria within that rule (logical AND).

A rule may specify:

  * **Sources:** The source identities to which the rule applies. A request’s identity must match one of the listed sources. Supported sources are:
    * **Kubernetes ServiceAccount** (with the ability to specify all ServiceAccounts in a given namespace)
    * **SPIFFE ID**

  * **Attributes:** Used to narrow the scope of the rule to apply to only some subset of traffic for the target workload destination. Currently, only port is supported. If no attributes are specified, the rule applies to all traffic toward the target.


### Policy Actions

The only currently-defined policy action is `ALLOW`. A request is allowed if and only if it matches at least one rule in any ALLOW policy targeting the workload.

If no authorization policies exist for a workload, and no other policies exist in the cluster or namespace which would deny traffic to the workload, traffic is permitted by default.

### Target of Authorization

The `targetRef` of the policy specifies the workload(s) to which the policy applies.

Note: Its worth noting that a policy will only ever target objects within the its namespace. Targeting objects in different namespaces brings security concerns, and while some implementations has some semantics to target across all namespaces (cluster-wide), the overall preference is to have a specific ClusterAuthorizationPolicy resource for this purpose.

> Note: in sidecar mode, it happens to also be the policy enforcement point, in ambient however, this is not true.

Before we are jumping into the options, lets start with some background.

##### Sidecar-Based Meshes

* **Architecture**: Sidecar proxy deployed alongside every pod  
* **Capabilities**: Can enforce all request attributes (L4 and L7)  
* **Targeting**: Uses label selectors to target sidecars for policy enforcement points (and destination)
* **Enforcement Point**: Single enforcement point (the destination sidecar)

##### Ambient Meshes

  * **Architecture**: Two-tier proxy system
    * Node-L4-proxy: Handles identity-based policies and port enforcement  
    * L7 enforcement point (waypoint proxy): Handles advanced L7 features  

  * **Targeting**:  
    * Label selectors for node proxy distribution
    * Service targetRef for L7 enforcement at waypoints
    * **Constraint**: Selectors and targetRefs cannot be used together

  * **Enforcement Point**: When using label selectors, enforcement at the node-proxy, when targeting a Service, enforcement happens at the Waypoint delegate for that service.

##### Key Challenges

  1. **Architectural Differences**: Ambient meshes separate L4 and L7 enforcement, while sidecars handle both  
  2. **Targeting Mechanisms**:
    * Ambient struggles with label selectors for L7 enforcement on waypoints ([#label-selectors-aren't-good-for-ambient-l7](#label-selectors-arent-good-for-ambient-l7))
    * Sidecars have difficulty supporting Service targeting for L7 enforcement ([#loss-of-service-context](#loss-of-service-context))

  3. **API Consistency**: Need a unified approach that works across all implementations

#### Label Selectors

Given these challenges, this GEP focuses on L4 only, leaving L7 authorization policy is for potential future work.
Since Label selectors are widely supported by all implementations for L4 authorization, we will begin with a `LabelSelector` inside targetRef to allow targeting a set of pods.

The [Future-Enhancements](#future-enhancements) section of this GEP gets into more background, complications, and recommended solution for incorporating L7 support to this policy.

**Benefits:**

* Aligns with established practices. Mesh implementations (Istio, Linkerd, Cilium) already use label selectors as the primary mechanism for targeting workloads in their native authorization policies, creating a consistent user experience.
* Directly applies policy to pods, avoiding ambiguity present when targeting Services. Ensures policies are enforced exactly where intended, regardless of how many Services a pod might belong to.
* Policies can apply to any workload, including pods not exposed via a `Service`, providing a comprehensive authorization solution.

**Downsides and Open Questions:**

The main downside of `LabelSelector` is the huge increase to the complexity of policy discoverability. See below for more info.

**Requirement: Enhancing Policy Attachment:**

This option depends on enhancements to Gateway API’s policy attachment model to support `LabelSelector` as a valid `targetRef`. This capability was discussed and received consensus at KubeCon North America 2024 and was originally in scope for GEP-713 but deferred for a future PR to keep GEP-713 focused on stabilizing what we already have (See [https://github.com/kubernetes-sigs/gateway-api/pull/3609#discussion_r2053376938](https://github.com/kubernetes-sigs/gateway-api/pull/3609#discussion_r2053376938)).

##### Experimental Pattern

To mitigate some of the concerns, `LabelSelector` support in policy attachment is designated as an **experimental pattern**.

* **Gateway API Community First:** Allows experimentation within Gateway API policies (like the one in this GEP).
* Implementations **should not** adopt `LabelSelector` targeting in their own custom policies attached to Gateway API resources until the pattern is sufficiently battle-tested and promoted to a standard feature. This staged approach mitigates risks of ecosystem fragmentation.

Here is how it is going to look like:

```go

// PolicyTargetReferenceWithLabelSelectors specifies a reference to a set of Kubernetes
// objects by Group and Kind, with an optional label selector to narrow down the matching
// objects.
//
// Currently, we only support label selectors when targeting Pods.
// This restriction is intentional to limit the complexity and potential
// ambiguity of supporting label selectors for arbitrary Kubernetes kinds.
// Unless there is a very strong justification in the future, we plan to keep this
// functionality limited to selecting Pods only.
//
// This is currently experimental in the Gateway API and should only be used
// for policies implemented within Gateway API. It is currently not intended for general-purpose
// use outside of Gateway API resources.
// +kubebuilder:validation:XValidation:rule="!(has(self.selector)) || (self.kind == 'Pod' && (self.group == '' || self.group == 'core'))",message="Selector may only be set when targeting Pods."
type PolicyTargetReferenceWithLabelSelectors struct {
  // Group is the group of the target object.
  Group Group `json:"group"`

  // Kind is the kind of the target object.
  Kind Kind `json:"kind"`

  // Selector is the label selector of target objects of the specified kind.
  Selector *metav1.LabelSelector `json:"selector"`
}

```

##### Enhanced Discoverability with `gwctl`

A key challenge with `LabelSelector` is the loss of discoverability. It’s easier to see which policies target a `Service` but difficult to determine which policies might affect a specific pod.

To address this, **investment in tooling is required.** Specifically, the `gwctl` CLI tool should be enhanced to provide insights such as:

```sh
gwctl describe pods pod1 

...
InheritedPolicies:
  Type                   Name                                 Target Kind   Target Name/Expression
  ----                   ----                                 -----------   -----------
  AuthorizationPolicy  demo-authz-policy-on-payment-pods      Pod           Selector={foo: a, bar: 2}
...

# List all AuthorizationPolicy resources in json format in the active namespace
gwctl get authorizationPolicy -o json

```

Without dedicated tooling, the `LabelSelector` approach could significantly degrade the user experience and observability.

##### Targeting All **Pods** in a Namespace

A common case for authorization is to target all the workloads in the namespace (where the policy resource lives). We can achieve the above with a few patterns:

###### Option 1

Target a namespace, `name` MUST be empty (will be the namespace where the policy resource lives).

```yaml
targetRefs:
- Kind: Namespace
```

This however gives the (confusing) impression that Gateways (whether in-cluster or off-cluster gateways) are targeted as well. Beyond the question of "does it really make sense for an E/W policy to be applicable to a N/S Gateway?" it leaves no way for users to do a namespace wide policy that excludes Gateways (no way to exclude Off-cluster Gateways with labels -- see option 3 for alternative).

###### Option 2

Leaving an empty targetRefs. While this is theoretically possible without a breaking change, this would be another fundamental change to policy attachment (to allow policies without a targetRef).
Additionally, this also suffers from the same problem of being perceived as an applicable scope for both Gateways and Workloads in the namespace.

This option is also inconsistent with other API fields where an empty field or absence of it does not mean select all. See [recent comment](https://github.com/kubernetes-sigs/gateway-api/pull/3887#discussion_r2176125600) on the proposal for empty ParentRef field.

###### Option 3 (Recommended)

An empty pod selector to target all **workloads** in the namespace. Kubernetes official docs clarify that the semantics of empty selectors are the decision of the API owner. In fact, many Kubernetes APIs (I know Service API does the opposite :/) using empty selectors as a select-all mechanism. See [NetworkPolicy podSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#networkpolicy-v1-networking-k8s-io), PodDisruptionBudget, ResourceQuota, and more.

```yaml
targetRefs:
- kind: Pod
  selector: {}
```

This provides a clearer separation between workloads and gateways. Gateways that happen to be Pods in the cluster, will also get selected. However they can be excluded with the right `matchExpression` if desired.

Below is a pseudo example:

```yaml
targetRefs:
- kind: Pod
  selector:
    matchExpression:
    - { key: "purpose", operator: NotIn, values: ["gateway"] }
```

##### New EnforcementLevel Enum

To mitigate the challenges L7 support across dataplanes is going to present, and to encourage policy authors to be more explicit, we introduce a new EnforcementLevel Enum.

We start by only supporting the Network value, and Application is reserved for future iteration.

Note: The whole point of this Enum is to encourage explicitness, we **do not** want to start without it and default to Network if and when introduced later.

```go

// EnforcementLevel defines the scope at which an AuthorizationPolicy is enforced.
//
// There are two enforcement levels:
//
//   - Network: Enforces the policy at the network layer (L4), typically at
//     network proxies or gateway dataplanes. Only supports attributes available
//     at connection time (e.g., source identity, port). Recommended for broad,
//     coarse-grained access controls.
//
//     NOTE FOR REVIEWERS -- THIS IS FOR FUTURE ENHANCEMENT
//
//   - Application: Enforces the policy at the application layer (L7), typically at
//     HTTP/gRPC-aware sidecars or L7 proxies. Enables fine-grained authorization
//     using protocol-specific attributes (e.g., HTTP paths, methods).
//
// This field clarifies policy intent and informs where enforcement is expected
// to happen. It also enables implementation-specific validation and behavior.
//
// +kubebuilder:validation:Enum=Network;
type EnforcementLevel string
```

### API Design

```go

type AuthorizationPolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of AuthorizationPolicy.
    Spec AuthorizationPolicySpec `json:"spec,omitempty"`

    // Status defines the current state of AuthorizationPolicy.
    Status PolicyStatus `json:"status,omitempty"`
}

// AuthorizationPolicyAction specifies the action to take.
// +kubebuilder:validation:Enum=ALLOW
type AuthorizationPolicyAction string

const (
    // ActionAllow allows requests that match the policy rules.
    ActionAllow AuthorizationPolicyAction = "ALLOW"
)

// AuthorizationPolicySpec defines the desired state of AuthorizationPolicy.
type AuthorizationPolicySpec struct {
    // TargetRef identifies the resource this policy is attached to.
    // (will be translated to CEL later):
    //   - When Kind is Pod, Name MUST be Empty, Selector MUST be set, Max targetRef length is 1. 
    //   - Otherwise multiple refs as long as Kind isn't Pod, selector MUST NOT be set
    // +kubebuilder:validation:Required
    TargetRefs []gatewayv1.PolicyTargetReferenceWithLabelSelectors `json:"targetRefs"`

    // Action specifies the action to take when a request matches the rules.
    // +kubebuilder:validation:Required
    Action AuthorizationPolicyAction `json:"action"`

    // Rules defines the list of authorization rules.
    // The policy action is applied if the request matches any of these rules.
    // +optional
    Rules []AuthorizationRule `json:"rules,omitempty"`

    // Informs where the policy is expected to be enforced, and what attributes are
    // allowed in policy rules.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum=Network;
    EnforcementLevel string `json:"enforcementLevel"`
}

// AuthorizationRule defines a single authorization rule.
// A request matches the rule if it matches both Sources and NetworkAttributes specified (logical AND).
type AuthorizationRule struct {
    // Sources specify a list of sources to match. 
    // a request matches this policy if it matches **any** of the specified sources (logical OR).
    // If specified as an empty list, matches **no** sources ("allow nothing").
    // If omitted, matches any source.
    // +optional
    Sources []*AuthorizationSource `json:"sources,omitempty"`

    // NetworkAttributes specifies TCP-level matching criteria.
    // If omitted, matches any TCP traffic.
    // +optional
    NetworkAttributes *AuthorizationNetworkAttributes `json:"networkAttributes,omitempty"`
}

// AuthorizationSourceType identifies a type of source for authorization.
// +kubebuilder:validation:Enum=ServiceAccount;SPIFFE
type AuthorizationSourceType string

const (
  // AuthorizationSourceTypeSPIFFE is used to identify a request matches a SPIFFE Identity.

  AuthorizationSourceTypeSPIFFE AuthorizationSourceType = "SPIFFE"

  // AuthorizationSourceTypeServiceAccount is used to identify a request matches a ServiceAccount from within the cluster.

  AuthorizationSourceTypeServiceAccount AuthorizationSourceType = "ServiceAccount"
)

// Source specifies the source of a request.
//
// Type must be set to indicate the type of source type.
// Similarly, either SPIFFE or Serviceaccount can be set based on the type.
//
type Source struct {

    // +unionDiscriminator
    // +kubebuilder:validation:Enum=ServiceAccount;SPIFFE
    // +kubebuilder:validation:Required
    Type AuthorizationSourceType `json:"type"`

    // spiffe specifies an identity that is matched by this rule.
    //
    // spiffe identities must be specified as SPIFFE-formatted URIs following the pattern:
    //   spiffe://<trust_domain>/<workload-identifier>
    //
    // The exact workload identifier structure is implementation-specific.
    //
    // spiffe identities for authorization can be derived in various ways by the underlying
    // implementation. Common methods include:
    // - From peer mTLS certificates: The identity is extracted from the client's
    //   mTLS certificate presented during connection establishment.
    // - From IP-to-identity mappings: The implementation might maintain a dynamic
    //   mapping between source IP addresses (pod IPs) and their associated
    //   identities (e.g., Service Account, SPIFFE IDs).
    // - From JWTs or other request-level authentication tokens.
    //
    // Note for reviewers: While this GEP primarily focuses on identity-based
    // authorization where identity is often established at the transport layer,
    // some implementations might derive identity from authenticated tokens or sources
    // within the request itself.
    //
    // +optional
    SPIFFE *AuthorizationSourceSPIFFE `json:"spiffe,omitempty"`

    // ServiceAccount specifies a Kubernetes Service Account that is
    // matched by this rule. A request originating from a pod associated with
    // this serviceaccount will match the rule.
    //
    // The ServiceAccount listed here is expected to exist within the same
    // trust domain as the targeted workload. Cross-trust-domain access
    // should instead be expressed using the `SPIFFE` field.
    // +optional
    ServiceAccount AuthorizationSourceServiceAccount `json:"serviceAccount,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="self.startsWith("spiffe://") && size(self.split("/")) >= 3", message="spiffe source must start with spiffe:// and include at least / separating trust domain from the workload identity."
type AuthorizationSourceSPIFFE AbsoluteURI

type AuthorizationSourceServiceAccount struct {
  // Namespace is the namespace of the ServiceAccount
  // If not specified, current namespace (the namespace of the policy) is used.
  Namespace *Namespace `json:"namespace,omitempty"`

  // Name is the name of the ServiceAccount. 
  // Use "*" to indicate all serviceaccounts in the namespace.
  // +kubebuilder:validation:Required
  Name string `json:"name"`
}


// AuthorizationAttribute defines L4 properties of a request destination.
type AuthorizationNetworkAttributes struct {
    // Ports specifies a list of destination ports to match on.
    // Traffic is matched if it is going to any of these ports.
    // If not specified, the rule applies to all ports.
    // <gateway:util:excludeFromCRD>
    // Note: We *may* introduce breaking change to this field if we decide to support port ranges.
    // </gateway:util:excludeFromCRD>
    // +optional
    Ports []gatewayv1.PortNumber `json:"ports,omitempty"`
}

```

Note: More advanced logic like negation or explicit Boolean operations is left for potential future work.

## Future Enhancements

### DENY Policies

We start without DENY policies in scope, DENY policies _may_ be added in future iterations.

### Future L7 Support

It is very likely that graduation of this API will also require adding L7 support to it. However, after numerous conversations on this topic, the decision is to start with L4 only and leave L7 for future iterations. The details of how L7 support could look like is not in scope for this GEP.

## Graduation Criteria and Guardrails

This GEP is starting as Experimental with X-prefix. We would like to define some guardrails to avoid being stuck with experimental api in the project forever.

GEP is removed after 3 releases unless:

  1. There is more than 1 implementation (implementations that are nearing completion can count here)
  2. There are conformance tests in place
3. The label selector pattern is added to the policy attachment definition in GEP-713.

GEP is removed after 6 releases if it hasn't graduated to GA.

## Conformance Details

### Feature Names

Two new feature sets will be added - AuthorizationPolicyCoreFeatures, and later, AuthorizationPolicyExtendedFeatures.

AuthorizationPolicyExtendedFeatures features, if and when introduced, would likely have individual granularity, like AuthorizationPolicyDeny, AuthorizationPolicyHTTP, AuthorizationPolicyGRPC, etc.

TBD exact FeatureNames.

### Conformance tests

## Appendix

### State of the World

| Aspect | Istio | Linkerd | Cilium |
| ----- | ----- | ----- | ----- |
| **Policy CRDs** | `AuthorizationPolicy` (APIs `security.istio.io/v1`) | `AuthorizationPolicy` (CRD `policy.linkerd.io/v1alpha1`), plus supporting CRDs (`Server`, `HTTPRoute`, `MeshTLSAuthentication`) | `CiliumNetworkPolicy` and `CiliumClusterwideNetworkPolicy` (superset of K8s NetworkPolicy) |
| **Identity model** | Identities derived from mTLS peer certificates (bound to SA): <ul><li>SPIFFE-like principal `<trust-domain>/ns/<namespace>/sa/<serviceaccount>`. </li> <li>ServiceAccount name </li> <li>Namespaces</li></ul></br> identity within JWT derived from `request.auth.principal`<br/><br/>IPBlocks and x-forwarded-for ipBlocks | Identities derived from mTLS peer certificates (bound to SA trust domain `identity.linkerd.cluster.local`. Policies reference Serviceaccounts or explicit mesh identities (e.g. `webapp.identity.linkerd.cluster.local`). <br/><br/>Policies use `requiredAuthenticationRefs` to reference the entities who get authorization. This is a list of targetRefs and it can include: <ul><li>ServiceAccounts</li> <li>`MeshTLSAuthentication` - which represents a set of mesh identities either with a mesh identities strings or reference to serviceAccounts</li> <li>`NetworkAuthentication` - represents sets of IPs or subnets.</li></ul>  |Cilium service mesh can leverage SPIFFE identities in certs that are used for handshake. These SPIFFE identities are mapped to CiliumIdentities. You can read more about cilium identities in [CiliumIdentity](#ciliumidentity). <br/><br/>Policies target abstractions like Serviceaccounts in the form of labels, pod labels, namespace label, node selectors, CIDR blocks and Cilium predefined [entities](https://docs.cilium.io/en/stable/security/policy/language/#entities-based). All policy targeting is coalesced by Cilium into one or more Cilium Identities for translation into the BPF datapath|
| **Enforcement** | For Istio with sidecars - a proxy on each pod. For ambient, ztunnel node agent enforces mTLS based L4 authorization, L7 authorization is being enforced in waypoints if any. <br/><br/> Istio supports ALLOW, DENY, CUSTOM (often used for external authorization), and AUDIT. DENY policies in istio's context are used to enforce higher priority deny policies. The allow semantics is that whatever is not allowed explicitly (and assuming there is any policy for the same match) is implicitly denied  | Linkerd data-plane proxy (injected into each pod). The proxy enforces policies via mTLS identity checks. <br/><br/> Linkerd supports AUDIT and ALLOW. There is not DENY policies, whats not allowed (and assuming there is any policy for the same match) is implicitly denied. | For L3/4 Ingress Rules, CiliumNetworkPolicy enforcement - an eBPF-based datapath in the Linux kernel on the destination node. If L7 http rules are specified, the packet is redirected for a node-local envoy for further enforcement.<br/><br/>Cilium supports ALLOW and DENY semantics - all policies generate audit logs. <br/><br/>Cilium service mesh also offers a kind of AuthN where a Cilium agent on the src node validates a workloads SPIFFE identity by talking to another agent on the destination node, performing the initial TLS handshake to do authentication.|
| **Request Match criteria** | Policies can target a group of pods using label selector, a Gateway/Service (this means targeting a waypoint proxy) or a GatewayClass - meaning all the gateways created from this class.  Policies without a label selector in a namespace implies the whole namespace is targeted. <br/><br/> Fine-grained L7 and L4 matching: HTTP/gRPC methods, paths, headers, ports, SNI, etc.Policies use logical OR over rules. <br/><br/>All match criteria are inline in the policy. See https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-To and https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-when | Policies can target: <ul><li>A `Server` which describes a set of pods (using fancy label match expressions), and a single port on those pods.</li> <li>A user can optionally restrict the authorization to a smaller subset of the traffic by targeting an HTTPRoute. (TODO: any plans to support sectionNames?)</li> <li> A namespace - this indicates that the policy applies to all traffic to all Servers and HTTPRoutes defined in the namespace.</li></ul> Note: We leave `ServerAuthorization` outside the scope as it planned to be deprecated (per linkerd website)  | Policies can target groups of pods using label selector (`endpointSelector`), or by node-labels (`nodeSelector`). Cilium supports L7 via built-in HTTP parsing: rules can match HTTP methods, paths, etc. For example, a CiliumNetworkPolicy can allow only specific HTTP methods/paths on a port. |
| **Default policies and admin policies** | If **no** ALLOW policy matches, traffic is **allowed** by default. You can deploy an overridable - default deny by default by deploying an **allow-nothing** policy in either the namespace or istio-system <br/><br/>AuthorizationPolicies in the `istio-system` namespace apply to the whole mesh and take precedence. These are not overridable by namespace-level policies.  | Default inbound policy can be set at install time using `proxy.defaultInboundPolicy`. Supported values are: <ul><li>`all-unauthenticated:` allow all traffic. This is the default.</li>  <li>`all-authenticated:` allow traffic from meshed clients in the same or from a different cluster (with multi-cluster).</li>  <li>`cluster-authenticated:` allow traffic from meshed clients in the same cluster.</li>  <li>`cluster-unauthenticated:` allow traffic from both meshed and non-meshed clients in the same cluster.</li>  <li>`deny:` all traffic are denied. </li> <li>`audit:` Same as all-unauthenticated but requests get flagged in logs and metrics.</li> </ul> <br/>Users can override the default policies for namespaces/pods or by setting the [config.linkerd.io/default-inbound-policy](http://config.linkerd.io/default-inbound-policy) annotation There is no support for admin, non overridable policies. | Follows Kubernetes NetworkPolicy semantics by default: if no `CiliumNetworkPolicy` allows the traffic, it is allowed (no implicit deny). Once at least one `CiliumNetworkPolicy` or `CiliumClusterwideNetworkPolicy` allows some traffic, all other traffic is implicitly denied.
<br/><br/> Operators must apply explicit deny rules or “default-deny” policies to block traffic in the absence of allow rules. <br/><br/> `CiliumClusterwideNetworkPolicy` exists for whole-cluster enforcement.)|


Every mesh vendor has their own API of such authorization. Below we describe brief UX for different implementations:

#### Istio

For the full spec and semantics of Istio AuthorizationPolicy: [Istio authorization policy docs](https://istio.io/latest/docs/reference/config/security/authorization-policy/)

Istio's AuthorizationPolicy can enforce access control by specifying allowed istio-formatted identities using the `source.principals` field, which matches authenticated Serviceaccount identities via mTLS. You can also use other source constructs which are described in the table above and in https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source.

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-sleep
  namespace: default
spec:
  selector:
    matchLabels:
      app: httpbin  # The policy applies to pods with this label
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/sleep"]
```

OR targeting a gateway for example.

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-sleep
  namespace: default
spec:
  targetRefs:
  - name: waypoint
    kind: Gateway # note: supported target Refs are Gateway, GatewayClass, Service, and ServiceEntry
    group: gateway.networking.k8s.io
  action: ALLOW
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/sleep"]
```

#### Linkerd

For the full spec and semantics of Linkerd AuthorizationPolicy: [Linkerd authorization policy docs](https://linkerd.io/2-edge/reference/authorization-policy/)

In Linkerd, identity-based authorization is enforced using AuthorizationPolicy and MeshTLSAuthentication, where MeshTLSAuthentication specifies allowed ServiceAccounts or mTLS identities (e.g., sleep.default.serviceaccount.identity.linkerd.cluster.local), ensuring that only authenticated workloads can access a resource.

Linkerd Policy can by applied to two different targets.

##### Pod Labels with Server Resource

```yaml
apiVersion: policy.linkerd.io/v1beta1
kind: Server
metadata:
  namespace: default
  name: httpbin-server
spec:
  podSelector:
    matchLabels:
      app: httpbin
  port: 8080
  proxyProtocol: HTTP/2

----
apiVersion: policy.linkerd.io/v1beta1
kind: MeshTLSAuthentication
metadata:
  name: sleep-authn
  namespace: default
spec:
  identities:
    - sleep.default.serviceaccount.identity.linkerd.cluster.local
----

apiVersion: policy.linkerd.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-sleep
  namespace: default
spec:
  targetRef:
    group: policy.linkerd.io
    kind: Server
    name: httpbin-server
  requiredAuthenticationRefs:
    - name: sleep-authn
      kind: MeshTLSAuthentication
      group: policy.linkerd.io/v1beta1

---
```

##### HTTPRoutes

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: httpbin-route
  namespace: default
spec:
  parentRefs:
    - name: httpbin
      kind: Service
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /
      backendRefs:
        - name: httpbin
          port: 80

---

apiVersion: policy.linkerd.io/v1beta1
kind: MeshTLSAuthentication
metadata:
  name: sleep-authn
  namespace: default
spec:
  identities:
    - sleep.default.serviceaccount.identity.linkerd.cluster.local
---

apiVersion: policy.linkerd.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: allow-sleep-http
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: httpbin-route
  requiredAuthenticationRefs:
    - name: sleep-authn
      kind: MeshTLSAuthentication
      group: policy.linkerd.io/v1beta1
---
```

#### Cilium

For the full spec and semantics of CiliumNetworkPolicy: https://docs.cilium.io/en/stable/network/kubernetes/policy/#ciliumnetworkpolicy & https://docs.cilium.io/en/stable/network/servicemesh/gateway-api/gateway-api/#cilium-s-ingress-config-and-ciliumnetworkpolicy

Beyond what's explained in the table above, Cilium also automatically labels each pod with its associated Serviceaccount using the label io.cilium.k8s.policy.serviceaccount. This label can be used in CiliumNetworkPolicy to enforce identity-based access controls using [ServiceAccounts Based Identities](https://docs.cilium.io/en/latest/security/policy/kubernetes/#serviceaccounts) within CiliumNetworkPolicy;

See below for example.

```yaml
apiVersion: "cilium.io/v2"
kind: CiliumNetworkPolicy
metadata:
  name: "k8s-svc-account-policy"
spec:
  endpointSelector:
    matchLabels:
      io.cilium.k8s.policy.serviceaccount: httpbin
  ingress:
  - fromEndpoints:
    - matchLabels:
        io.cilium.k8s.policy.serviceaccount: sleep
    toPorts:
    - ports:
      - port: '80'
        protocol: TCP
      rules:
        http:
        - method: GET
          path: "/"
```

##### CiliumIdentity

Cilium has the concept of CiliumIdentity. Pods are assigned identities derived from their Kubernetes labels (namespace, app labels, etc.). Cilium’s policy matches based on these label-derived identities. The CiliumIdentity implementation maps an integer to a group of IP addresses (the pod IPs associated with a group of pods). This “integer” and its mapping to pod IP addresses represents the core identity primitive in Cilium.

More on https://docs.cilium.io/en/stable/internals/security-identities/ & https://docs.cilium.io/en/stable/security/network/identity/

### Loss Of Service Context

When applying authorization policy to all traffic addressed to a Service;

This option is very tricky to implement for sidecar-based meshes where the destination sidecar has no knowledge of which Service the request came through.

Here is the very high-level traffic flow for sidecar-based meshes:

  ```sh
  Client → Request to backend-service:8080
      → Source sidecar resolves service → backend-pod-1 (<ip>:8080)
      → Destination sidecar receives traffic on pod IP
      → Destination sidecar has NO context that this came via "backend-service"
  ```

Solving this problem either introduces security concern (e.g adding request metadata to indicate which Service was dialed), or unnecessarily complex or in-efficient to solve.

### Label-Selectors aren't Good for Ambient L7

In L7 Ambient, AuthorizationPolicy targets a Service. This Service has to have a waypoint proxy. The policy enforcement point is the waypoint, but it is also the point where the Service VIP resolution happens.

If we were to target Label Selectors, the waypoint proxy would get the request, do VIP resolution and select an endpoint, and then it would require round-tripping to itself for doing policy enforcement.

@howardjohn has actually added that this was actually the first way ambient had implemented authorization, but it has proved to be much less performant.