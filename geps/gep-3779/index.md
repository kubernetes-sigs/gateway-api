# GEP-3779: Identity Based Authz for East-West Traffic

* Issue: [#3779](https://github.com/kubernetes-sigs/gateway-api/issues/3779)
* Status: Implementable

(See [status definitions](../overview.md#gep-states).)


## TLDR

Provide a method for configuring Gateway API Mesh implementations to enforce east-west identity-based Authorization controls. At the time of writing this we leave Authentication for specific implementation and outside of this proposal scope.


## Goals

(Using the [Gateway API Personas](../../concepts/roles-and-personas.md))

* A way for Ana the Application Developer to configure a Gateway API for Mesh implementation to enforce authorization policy that **allows** or **denies** identity or multiple identities to talk with some set (could be namespace or more granualr) of the workloads she controls.

* A way for both Ana and Chihiro to restrict the scope of the policies they deploy to specific ports.

## TBD Goals

* A way for Chihiro, the Cluster Admin, to configure a Gateway API for Mesh implementation to enforce non-overridable cluster-wide, authorization policies that **allows** or **denies** identity or multiple identities to talk with some set of the workloads in the cluster.

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
| **Policy CRDs** | `AuthorizationPolicy` (APIs `security.istio.io/v1`) | `AuthorizationPolicy` (CRD `policy.linkerd.io/v1alpha1`), plus supporting CRDs (`Server`, `HTTPRoute`, `MeshTLSAuthentication`) | `CiliumNetworkPolicy` and `CiliumClusterwideNetworkPolicy` (superset of K8s NetworkPolicy) |
| **Identity model** | Identities derived from mTLS peer certificates (bound to SA): <ul><li>SPIFFE-like principal `<trust-domain>/ns/<namespace>/sa/<serviceaccount>`. </li> <li>ServiceAccount name </li> <li>Namespaces</li></ul></br> identity within JWT derived from `request.auth.principal`<br/><br/>IPBlocks and x-forwarded-for ipBlocks | Identities derived from mTLS peer certificates (bound to SA trust domain `identity.linkerd.cluster.local`. Policies reference Serviceaccounts or explicit mesh identities (e.g. `webapp.identity.linkerd.cluster.local`). <br/><br/>Policies use `requiredAuthenticationRefs` to reference the entities who get authorization. This is a list of targetRefs and it can include: <ul><li>ServiceAccounts</li> <li>`MeshTLSAuthentication` - which represents a set of mesh identities either with a mesh identities strings or reference to serviceAccounts</li> <li>`NetworkAuthentication` - represents sets of IPs or subnets.</li></ul>  |Cilium service mesh can leverage SPIFFE identities in certs that are used for handshake. These SPIFFEE identities are mapped to CiliumIdentities. You can read more about cilium identities in [CiliumIdentity](#CiliumIdentity). <br/><br/>Policies target abstractions like Serviceaccounts in the form of labels, pod labels, namespace label, node selectors, CIDR blocks and Cilium predefined [entities](https://docs.cilium.io/en/stable/security/policy/language/#entities-based). All policy targeting is coalesced by Cilium into one or more Cilium Identities for translation into the BPF datapath|
| **Enforcement** | For Istio with sidecars - a proxy on each pod. For ambient, ztunnel node agent enforces mTLS based L4 authorization, L7 authorization is being enforced in waypoints if any. <br/><br/> Istio supports ALLOW, DENY, CUSTOM (often used for external authorization), and AUDIT. DENY policies in istio's context are used to enforce higher priority deny policies. The allow semantics is that whatever is not allowed explicitly (and assuming there is any policy for the same match) is implicitly denied  | Linkerd data-plane proxy (injected into each pod). The proxy enforces policies via mTLS identity checks. <br/><br/> Linkerd supports AUDIT and ALLOW. There is not DENY policies, whats not allowed (and assuming there is any policy for the same match) is implicitly denied. | For L3/4 Ingress Rules, CiliumNetworkPolicy enforcement - an eBPF-based datapath in the Linux kernel on the destination node. If L7 http rules are specified, the packet is redirected for a node-local envoy for further enforcement.<br/><br/>Cilium supports ALLOW and DENY semantics - all policies generate audit logs. <br/><br/>Cilium service mesh also offers a kind of AuthN where a Cilium agent on the src node validates a workloads SPIFFE identity by talking to another agent on the destination node, performing the initial TLS handshake to do authentication.|
| **Request Match criteria** | Policies can target a group of pods using label selector, a Gateway/Service (this means targeting a waypoint proxy) or a GatewayClass - meaning all the gateways created from this class.  Policies without a label selector in a namespace implies the whole namespace is targeted. <br/><br/> Fine-grained L7 and L4 matching: HTTP/gRPC methods, paths, headers, ports, SNI, etc.Policies use logical OR over rules. <br/><br/>All match criterias are inline in the policy. See https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-To and https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-when | Policies can target: <ul><li>A `Server` which describes a set of pods (using fancy label match expressions), and a single port on those pods.</li> <li>A user can optionally restrict the authorization to a smaller subset of the traffic by targeting an HTTPRoute. (TODO: any plans to support sectionNames?)</li> <li> A namespace - this indicates that the policy applies to all traffic to all Servers and HTTPRoutes defined in the namespace.</li></ul> Note: We leave `ServerAuthorization` outside the scope as it planned to be deprecated (per linkerd website)  | Policies can target groups of pods using label selector (`endpointSelector`), or by node-labels (`nodeSelector`). Cilium supports L7 via built-in HTTP parsing: rules can match HTTP methods, paths, etc. For example, a CiliumNetworkPolicy can allow only specific HTTP methods/paths on a port. |
| **Default policies and admin policies** | If **no** ALLOW policy matches, traffic is **allowed** by default. You can deploy an overridable - default deny by default by deploying an **allow-nothing** policy in either the namespace or istio-system <br/><br/>AuthorizationPolicies in the `istio-system` namespace apply to the whole mesh and take precedence. These are not overridable by namespace-level policies.  | Default inbound policy can be set at install time using `proxy.defaultInboundPolicy`. Supported values are: <ul><li>`all-unauthenticated:` allow all traffic. This is the default.</li>  <li>`all-authenticated:` allow traffic from meshed clients in the same or from a different cluster (with multi-cluster).</li>  <li>`cluster-authenticated:` allow traffic from meshed clients in the same cluster.</li>  <li>`cluster-unauthenticated:` allow traffic from both meshed and non-meshed clients in the same cluster.</li>  <li>`deny:` all traffic are denied. </li> <li>`audit:` Same as all-unauthenticated but requests get flagged in logs and metrics.</li> </ul> <br/>Users can override the default policies for namespaces/pods or by setting the [config.linkerd.io/default-inbound-policy](http://config.linkerd.io/default-inbound-policy) annotation There is no support for admin, non overridable policies. | Follows Kubernetes NetworkPolicy semantics by default: if no `CiliumNetworkPolicy` allows the traffic, it is allowed (no implicit deny). Once at least one `CiliumNetworkPolicy` or `CiliumClusterwideNetworkPolicy` allows some traffic, all other traffic is implicitly denied.
<br/><br/> Operators must apply explicit deny rules or “default-deny” policies to block traffic in the absence of allow rules. <br/><br/> `CiliumClusterwideNetworkPolicy` exists for whole-cluster enforcement.)|


Every mesh vendor has their own API of such authorization. Below we describe brief UX for different implementations:

#### Istio
For the full spec and sematics of Istio AuthorizationPolicy: [Istio authorization policy docs](https://istio.io/latest/docs/reference/config/security/authorization-policy/)

Istio's AuthorizationPolicy can enforce access control by specifying allowed istio-formatted identities using the `source.principals` field, which matches authenticated Serviceaccount identities via mTLS. You can also use other source constructs which are described in the table above and in https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source.

```
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

```
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

For the full spec and sematics of Linkerd AuthorizationPolicy: [Linkerd authorization policy docs](https://linkerd.io/2-edge/reference/authorization-policy/)

In Linkerd, identity-based authorization is enforced using AuthorizationPolicy and MeshTLSAuthentication, where MeshTLSAuthentication specifies allowed ServiceAccounts or mTLS identities (e.g., sleep.default.serviceaccount.identity.linkerd.cluster.local), ensuring that only authenticated workloads can access a resource.

Linkerd Policy can by applied to two different targets.

##### Pod Labels with Server Resource

```
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

```
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

-----

apiVersion: policy.linkerd.io/v1beta1
kind: MeshTLSAuthentication
metadata:
  name: sleep-authn
  namespace: default
spec:
  identities:
    - sleep.default.serviceaccount.identity.linkerd.cluster.local
-----

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

For the full spec and sematics of CiliumNetworkPolicy: https://docs.cilium.io/en/stable/network/kubernetes/policy/#ciliumnetworkpolicy & https://docs.cilium.io/en/stable/network/servicemesh/gateway-api/gateway-api/#cilium-s-ingress-config-and-ciliumnetworkpolicy

Beyond what's explained in the table above, Cilium also automatically labels each pod with its associated Serviceaccount using the label io.cilium.k8s.policy.serviceaccount. This label can be used in CiliumNetworkPolicy to enforce identity-based access controls using [ServiceAccounts Based Identities](https://docs.cilium.io/en/latest/security/policy/kubernetes/#serviceaccounts) within CiliumNetworkPolicy;

See below for example.

```
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


## API

This GEP introduces a new policy resource, `AuthorizationPolicy`, for **identity-based** authorization. The policy defines a target, a single action (`ALLOW` or `DENY`), and a set of rules that include sources (the “who”) and an optional port attribute.

### **Policy Rules**

Each `AuthorizationPolicy` resource contains a list of rules. A request matches the policy if it matches **any** rule in the list (logical OR). Each rule defines multiple matching criteria; a request matches a rule only if it matches **all** criteria within that rule (logical AND).

A rule may specify:

  * **Sources:** The source identities to which the rule applies. A request’s identity must match one of the listed sources. Supported source kinds are:
    * **SPIFFE ID**
    * **Kubernetes ServiceAccount**
    * **Kubernetes Namespace**

  * **Attributes:** Conditions on the target workload, at the time of writing this, only port is supported. If no attributes are specified, the rule applies to all traffic toward the target.

### **ALLOW Policies**

  * An **ALLOW** policy is permissive.
  * A request is allowed if:
    * It matches at least one rule in any ALLOW policy targeting the workload **and**
    * It is not explicitly denied by any DENY policy.

  * If no ALLOW policy exists for a workload, traffic is permitted by default, unless any DENY policy applies.

### **DENY Policies**

* A **DENY** policy is restrictive and takes precedence over ALLOW.
* If a request matches any rule in a DENY policy, it is immediately rejected, regardless of matching ALLOW rules elsewhere.
* DENY policies enable to define global blocks or exceptions (for example: “block all traffic from Namespace X”).

### **ALLOW vs. DENY Semantics**

* **DENY always wins.** If both an ALLOW and a DENY policy match a request, the DENY policy blocks it.
* The presence of any authorization policy causes the system to default to **deny-by-default** for matching workloads.
* Another bullet to re-clarify the one above - the default behavior when no policies select a target workload is to allow all traffic. However, **as soon as at least one `AuthorizationPolicy` targets a workload, the model becomes implicitly deny-if-not-allowed**.

### **Target of Authorization**

The `targetRef` of the policy specifies the workload(s) to which the policy applies. Two options are available for `targetRef`:

#### **Option 1: Targeting a Service**

The `targetRef` can point to a Kubernetes `Service`.

Two implementation options when targeting a Service:

1. Apply authorization policy to all traffic addressed to this Service.

1. Apply authorization policy to all workloads (pods) selected by the Service.

#### Benefits

* **No API Extension Required:** Works with the current PolicyAttachment model in Gateway API without modification.
* **Simplicity:** Intuitive for users familiar with Kubernetes networking concepts.

#### Downsides and Open Questions
However, targeting a `Service` introduces several challenges;

##### Loss Of Service Context

If we go with option 1, apply authorization policy to all traffic addressed to a Service;

This option is very tricky to implement for sidecar-based meshes where the destination sidecar has no knowledge of which Service the request came through.

Here is the very high-level tarffic flow for sidecar-based meshes:

  ```sh
  Client → Request to backend-service:8080
      → Source sidecar resolves service → backend-pod-1 (<ip>>:8080)
      → Destination sidecar receives traffic on pod IP
      → Destination sidecar has NO context that this came via "backend-service"
  ```

Solving this problem either introduces security concern (e.g adding request metadata to indicate which Service was dialed), or unnecessarily complex or in-efficient to solve.

#### A Workload is Part of Multiple Services

Assuming we go with option 2, apply authorization policy to all workloads (pods) selected by the Service.

If a Pod belongs to multiple Services targeted by different authorization policies, precedence rules, may become unclear, leading to unpredictable or insecure outcomes. Even if such rules are explicitly defined, UX could potentially be confusing for users. For example we _could_ apply the following algorithm:

```sh
## Algorithm
For traffic to workload W:
1. Collect all Service-targeted AuthorizationPolicies where Service.selector matches W.labels
2. If no policies collected → Allow (default behavior)
3. If policies collected → Evaluate using union semantics


## Union semantics:
1. For each DENY policy: if request matches → DENY immediately
2. For each ALLOW policy: collect matching rules  
3. If any ALLOW rule matches → ALLOW
4. Otherwise → DENY
```

#### Enforcement & Consistency

Assuming we go with option 2, apply authorization policy to all workloads (pods) selected by the Service.

The UX becomes very confusing. Are we going to enforce only if the traffic arrived through the specific Service? Probably not, cause then we are back to the first problem of Service context.

The UX gets weird becuase even though you target a Service, you essentially get a **workload** policy, thats enforced regardless. 

> Note: with Service as a targetRef, of course we are going to need a Service in order to enforce Authorization -- meaning pods/jobs without a Serivce are completely out of scope.

#### **Option 2: Targeting xRoutes**

The main benefit of this option is that we can more easily scope the authorization enforcement only for "GAMMA" traffic. Whether its more or less confusing is still unclear.

We can target xRoutes (TCPRoute, HTTPRoute, GRPCRoute). However we are back to [Loss Of Service Context](#loss-of-service-context). In sidecar mode, same as we don't have the context of which Service was dialed, we don't have the route information that was responsible for the routing.

> Note: Linkerd solved this with [Reusing HTTPRoute Schema](https://linkerd.io/2.15/features/httproute/) to distinguish between Inbound and Outbound HTTPRoute. However, I doubt we want that as a community feature. (sorry @kflynn)

Another (perhaps, easier-to-address) concern is that if we target xRoutes, it is likely that users would expect this Authorization to work for N/S traffic. Documentation and guidance are nice, but I think this still ends up more confusing for no real value.

#### **Option 3: Targeting Pods via Label Selectors**

Alternatively, the `targetRef` can specify a set of pods using a `LabelSelector` for a more flexible and direct approach.

**Benefits:**

* Aligns with established practices. Mesh implementations (Istio, Linkerd, Cilium) already use label selectors as the primary mechanism for targeting workloads in their native authorization policies, creating a consistent user experience.
* Directly applies policy to pods, avoiding ambiguity present when targeting Services. Ensures policies are enforced exactly where intended, regardless of how many Services a pod might belong to.
* Policies can apply to any workload, including pods not exposed via a `Service`, providing a comprehensive authorization solution.

**Downsides and Open Questions:**

The main downside of `LabelSelector` is the huge increase to the complexity of policy discoverability. See below for more info.

**Requirement: Enhancing Policy Attachment:**

This option depends on enhancements to Gateway API’s policy attachment model to support `LabelSelector` as a valid `targetRef`. This capability was discussed and received consensus at KubeCon North America 2024 and was originally in scope for GEP-713 but deferred for a future PR to keep GEP-713 focused on stabilizing what we already have (See [https://github.com/kubernetes-sigs/gateway-api/pull/3609#discussion_r2053376938](https://github.com/kubernetes-sigs/gateway-api/pull/3609#discussion_r2053376938)).

##### **Experimental Pattern**

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

##### Namespace-Wide Policies

A common case for authorization is to target a whole namespace. We can achieve the above with a few patterns:

###### Option 1

Target a namespace, name MUST be empty (as this is already namespaced policy)

```yaml
targetRefs:
- Kind: Namespace
```

This however gives the (confusing) impression that Gateways (whether in-cluster or off-cluster gateways) are targeted as well. Beyond the question of "does it really make sense for an E/W policy to be applicable to a N/S Gateway?" it leaves no way for users to do a namespace wide policy that excludes Gateways (no way to exclude Off-cluster Gateways with labels -- see option 3 for alternative).

###### Option 2

Leaving an empty targetRefs. While this is theoretically possible without a breaking change, this would be another fundamental change to policy attachment (to allow policies without a targetRef). 
Additionally, this is also suffer from the same problem of being perceived as an applicable scope for both Gateways and Workloads in the namespace.

This option is also inconsistent with other API fields where an empty field or absence of a it does not mean select all. See [recent comment](https://github.com/kubernetes-sigs/gateway-api/pull/3887#discussion_r2176125600) on the proposal for empty ParentRef field.

###### Option 3

An empty pod selector. Kubernetes official docs clarify that the semantics of empty selectors are the decision of the API owner. In fact, many Kubernetes APIs (I know Service API does the opposite :/) using empty selectors as a select-all mechanism. See [NetworkPolicy podSelector](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#networkpolicy-v1-networking-k8s-io), PodDisruptionBudget, ResourceQuota, and more.

```yaml
targetRefs:
- Kind: Pod
  Selector: {}
```

This provides a clearer separation between workloads and gateways. Gateways that happen to be Pods in the cluster, will also get selected. However they can be excluded with the right `matchExpression` if desired.

Below is a pseudo example:

```yaml
targetRefs:
- Kind: Pod
  Selector:
    matchExpression:
    - { key: "purpose", operator: NotIn, values: ["gateway"] }
```

##### **Enhanced Discoverability with `gwctl`**

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

#### Recommended Option - Hybrid TargetRef and VAP

Adding only the recommendation below, please see [https://docs.google.com/document/d/1CeagBnHDPbzpYAxBmtJqTshxAW8l-aRPwvwgAGbnv2I/edit?tab=t.0](https://docs.google.com/document/d/1CeagBnHDPbzpYAxBmtJqTshxAW8l-aRPwvwgAGbnv2I/edit?tab=t.0) for a comprehensive comparison for design options for TargetRef.

Create one unified authorization policy API with implementation-specific validation using Validating Admission Policies (VAP). Additionally introducing a new `enforcementLevel` enum to simplify validation and enable the user to clarify intent.

```go

// EnforcementLevel defines the scope at which an AuthorizationPolicy is enforced.
//
// There are two enforcement levels:
//
//   - NETWORK: Enforces the policy at the network layer (L4), typically at
//     network proxies or gateway dataplanes. Only supports attributes available
//     at connection time (e.g., source identity, port). Recommended for broad,
//     coarse-grained access controls.
//
//   - APPLICATION: Enforces the policy at the application layer (L7), typically at
//     HTTP/gRPC-aware sidecars or L7 proxies. Enables fine-grained authorization
//     using protocol-specific attributes (e.g., HTTP paths, methods).
//
// This field clarifies policy intent and informs where enforcement is expected
// to happen. It also enables implementation-specific validation and behavior.
//
// +kubebuilder:validation:Enum=NETWORK;APPLICATION
type EnforcementLevel string

```

```
# Network-scoped policy (L4 enforcement points)
kind: AuthorizationPolicy
metadata:
  name: network-authz
spec:
  enforcementLevel: "network"  # Enforced at L4 proxies
  targetRef: 
    kind: Pod
    selector: {}
  rules:
    - authorizationSource:
        serviceAccount: ["default/productpage"]
      networkAttributes:
        ports: [9080]

# Application-scoped policy (L7-only enforcement points and sidecars)
kind: AuthorizationPolicy  
metadata:
  name: app-authz
spec:
  enforcementLevel: "application"  # Enforced at L7 proxies
  targetRef: # Service for Ambient implementations, label-selector for sidecar implementations.
  rules:
    - authorizationSource:
        serviceAccount: ["default/productpage"]
      networkAttributes:
        ports: [9080]
      applicationAttributes: 
        paths: ["/api/*"]
        methods: ["GET", "POST"]


```

### Mesh Implementation Behavior

Sidecar Meshes

* **Network Level**: Sidecar enforces policy but only uses L4 attributes  
* **Application Level**: Sidecar enforces policy with full L4+L7 capabilities

Ambient Meshes

* **Network Level**: Node-L4-proxy enforces policy using labelSelector targeting  
* **Application Level**: Waypoint proxy enforces policy using Service targetRef targeting

**VAP Validation**: 

* **Sidecar mesh**: Ensures application-level policies don't use targetRef: Service (since sidecars use labelSelector)  
  * **Ambient**: Ensures application-level policies don't use label-selectors.  
  * **Both sidecar and ambient:** ensures network level policies **do** use label selectors

Advantages

* **Clear Intent**: Users explicitly state where they want enforcement  
* **Flexible Targeting**: Different targetRef types for different enforcement points  
* **Migration Path**: Users can start with network-level and upgrade to application-level easily  
* **Implementation Alignment**: supporting different meshes architectures  
* **Single API**: No duplication of schemas or concepts  
* **Validation Clarity**: Clear rules about what's allowed at each level

Disadvantages

* **Complexity**: Users must understand enforcement levels  
* **VAP Dependency**: Requires validation rules

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
// +kubebuilder:validation:Enum=ALLOW;DENY
type AuthorizationPolicyAction string

const (
    // ActionAllow allows requests that match the policy rules.
    ActionAllow AuthorizationPolicyAction = "ALLOW"
    // ActionDeny denies requests that match the policy rules.
    ActionDeny AuthorizationPolicyAction = "DENY"
)

// AuthorizationPolicySpec defines the desired state of AuthorizationPolicy.
type AuthorizationPolicySpec struct {
    // TargetRef identifies the resource this policy is attached to.
    // +kubebuilder:validation:Required
    TargetRefs []gatewayv1.PolicyTargetReferenceWithLabelSelectors `json:"targetRefs"`

    // Action specifies the action to take when a request matches the rules.
    // +kubebuilder:validation:Required
    Action AuthorizationPolicyAction `json:"action"`

    // Rules defines the list of authorization rules.
    // A request matches the policy if it matches any of these rules.
    // +optional
    Rules []AuthorizationRule `json:"rules,omitempty"`
}

// AuthorizationRule defines a single authorization rule.
// A request matches the rule if it matches ALL fields specified.
type AuthorizationRule struct {
    // AuthorizationSource specifies who is making the request.
    // If omitted, matches any source.
    // +optional
    AuthorizationSource *AuthorizationSource `json:"authorizationSource,omitempty"`

    // NetworkAttributes specifies TCP-level matching criteria.
    // If omitted, matches any TCP traffic.
    // +optional
    NetworkAttributes *AuthorizationNetworkAttributes `json:"networkAttributes,omitempty"`

    // FOR FUTURE ENHANCEMENT!!

    // ApplicationAttributes defines optional application-layer (L7) matching criteria
    // used to authorize requests when enforcement is at the APPLICATION level.
    //
    // All specified fields must match for the rule to apply (logical AND). If omitted,
    // the rule matches all application-layer traffic.
    //
    // Typical use cases include protocol-aware authorization for HTTP, gRPC, or other L7 traffic.
    // Fields may vary in applicability based on the protocol and implementation support.
    //
    // NOTE: ApplicationAttributes are only meaningful when `enforcementLevel` is set to
    // APPLICATION. Implementations MUST ignore or reject these fields if used at NETWORK level.

    // +optional
    ApplicationAttributes *AuthorizationApplicationAttributes `json:"applicationAttributes,omitempty"`
}



// AuthorizationSource specifies the source of a request.
//
// At least one field may be set. If multiple fields are set,
// a request matches this AuthorizationSource if it matches
// **any** of the specified criteria (logical OR across fields).
//
// For example, if both `Identities` and `ServiceAccounts` are provided,
// the rule matches a request if either:
// - the request's identity is in `Identities`
// - OR the request's Serviceaccount matches an entry in `ServiceAccounts`.
//
// Each list within the fields (e.g. `Identities`) is itself an OR list.
//
// If this struct is omitted in a rule, it matches any source.
//
// NOTE: In the future, if there’s a need to express more complex
// logical conditions (e.g. requiring a request to match multiple
// criteria simultaneously—logical AND), we may evolve this API
// to support richer match expressions or logical operators.
type AuthorizationSource struct {

    // Identities specifies a list of identities that are matched by this rule.
    // A request's identity must be present in this list to match the rule.
    //
    // Identities must be specified as SPIFFE-formatted URIs following the pattern:
    //   spiffe://<trust_domain>/<workload-identifier>
    //
    // While the exact workload identifier structure is implementation-specific,
    // implementations are encouraged to follow the convention of
    // `spiffe://<trust_domain>/ns/<namespace>/sa/<serviceaccount>`
    // when representing Kubernetes workload identities.
    //
    // Identities for authorization can be derived in various ways by the underlying
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
    Identities []string `json:"identities,omitempty"`

    // ServiceAccounts specifies a list of Kubernetes Service Accounts that are
    // matched by this rule. A request originating from a pod associated with
    // one of these Serviceaccounts will match the rule.
    //
    // Values must be in one of the following formats:
    //   - "<namespace>/<serviceaccount-name>": A specific Serviceaccount in a namespace.
    //   - "<namespace>/*": All Serviceaccounts in the given namespace.
    //   - "<serviceaccount-name>": a Serviceaccount in the same namespace as the policy.
    //
    // Use of "*" alone (i.e., all Serviceaccounts in all namespaces) is not allowed.
    // To select all Serviceaccounts in the current namespace, use "<namespace>/*" explicitly.
    //
    // Example:
    //   - "default/bookstore" → Matches Serviceaccount "bookstore" in namespace "default"
    //   - "payments/*" → Matches any Serviceaccount in namespace "payments"
    //   - "frontend" → Matches "frontend" Serviceaccount in the same namespace as the policy
    //
    // The ServiceAccounts listed here are expected to exist within the same
    // trust domain as the targeted workload, which in many environments means
    // the same Kubernetes cluster. Cross-cluster or cross-trust-domain access
    // should instead be expressed using the `Identities` field.
    // 
    // +optional 
    ServiceAccounts []string `json:"serviceAccounts,omitempty"`
}

// AuthorizationAttribute defines L4 properties of a request destination.
type AuthorizationNetworkAttributes struct {
    // Ports specifies a list of destination ports to match on.
    // Traffic is matched if it is going to any of these ports.
    // If not specified, the rule applies to all ports.
    // +optional
    Ports []gatewayv1.PortNumber `json:"ports,omitempty"`
}

```

Note: Existing AuthorizationAPIs recognized the need to support negation fields like `not{field}`. To avoid duplicating fields with negation, we plan to support richer match expressions for fields in `AuthorizationSource` such as `matchExpressions: { operator: In|NotIn, values: []}`.

## Conformance Details

### Feature Names

Two new feature sets will be added - AuthorizationPolicyCoreFeatures, and later, AuthorizationPolicyExtendedFeatures.

TBD exact FeatureNames.

### Conformance tests


## Alternatives


## References