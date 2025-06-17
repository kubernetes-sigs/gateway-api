# GEP-3779: Identity Based Authz for East-West Traffic

* Issue: [#3779](https://github.com/kubernetes-sigs/gateway-api/issues/3779)
* Status: Provisional

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
| **Identity model** | Identities derived from mTLS peer certificates (bound to SA): <ul><li>SPIFFE-like principal `<trust-domain>/ns/<namespace>/sa/<serviceaccount>`. </li> <li>ServiceAccount name </li> <li>Namespaces</li></ul></br> identity within JWT derived from `request.auth.principal`<br/><br/>IPBlocks and x-forwarded-for ipBlocks | Identities derived from mTLS peer certificates (bound to SA trust domain `identity.linkerd.cluster.local`. Policies reference service accounts or explicit mesh identities (e.g. `webapp.identity.linkerd.cluster.local`). <br/><br/>Policies use `requiredAuthenticationRefs` to reference the entities who get authorization. This is a list of targetRefs and it can include: <ul><li>ServiceAccounts</li> <li>`MeshTLSAuthentication` - which represents a set of mesh identities either with a mesh identities strings or reference to serviceAccounts</li> <li>`NetworkAuthentication` - represents sets of IPs or subnets.</li></ul>  |Cilium service mesh can leverage SPIFFE identities in certs that are used for handshake. These SPIFFEE identities are mapped to CiliumIdentities. You can read more about cilium identities in [CiliumIdentity](#CiliumIdentity). <br/><br/>Policies target abstractions like service accounts in the form of labels, pod labels, namespace label, node selectors, CIDR blocks and Cilium predefined [entities](https://docs.cilium.io/en/stable/security/policy/language/#entities-based). All policy targeting is coalesced by Cilium into one or more Cilium Identities for translation into the BPF datapath|
| **Enforcement** | For Istio with sidecars - a proxy on each pod. For ambient, ztunnel node agent enforces mTLS based L4 authorization, L7 authorization is being enforced in waypoints if any. <br/><br/> Istio supports ALLOW, DENY, CUSTOM (often used for external authorization), and AUDIT. DENY policies in istio's context are used to enforce higher priority deny policies. The allow semantics is that whatever is not allowed explicitly (and assuming there is any policy for the same match) is implicitly denied  | Linkerd data-plane proxy (injected into each pod). The proxy enforces policies via mTLS identity checks. <br/><br/> Linkerd supports AUDIT and ALLOW. There is not DENY policies, whats not allowed (and assuming there is any policy for the same match) is implicitly denied. | For L3/4 Ingress Rules, CiliumNetworkPolicy enforcement - an eBPF-based datapath in the Linux kernel on the destination node. If L7 http rules are specified, the packet is redirected for a node-local envoy for further enforcement.<br/><br/>Cilium supports ALLOW and DENY semantics - all policies generate audit logs. <br/><br/>Cilium service mesh also offers a kind of AuthN where a Cilium agent on the src node validates a workloads SPIFFE identity by talking to another agent on the destination node, performing the initial TLS handshake to do authentication.|
| **Request Match criteria** | Policies can target a group of pods using label selector, a Gateway/Service (this means targeting a waypoint proxy) or a GatewayClass - meaning all the gateways created from this class.  Policies without a label selector in a namespace implies the whole namespace is targeted. <br/><br/> Fine-grained L7 and L4 matching: HTTP/gRPC methods, paths, headers, ports, SNI, etc.Policies use logical OR over rules. <br/><br/>All match criterias are inline in the policy. See https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-To and https://istio.io/latest/docs/reference/config/security/authorization-policy/#Rule-when | Policies can target: <ul><li>A `Server` which describes a set of pods (using fancy label match expressions), and a single port on those pods.</li> <li>A user can optionally restrict the authorization to a smaller subset of the traffic by targeting an HTTPRoute. (TODO: any plans to support sectionNames?)</li> <li> A namespace - this indicates that the policy applies to all traffic to all Servers and HTTPRoutes defined in the namespace.</li></ul> Note: We leave `ServerAuthorization` outside the scope as it planned to be deprecated (per linkerd website)  | Policies can target groups of pods using label selector (`endpointSelector`), or by node-labels (`nodeSelector`). Cilium supports L7 via built-in HTTP parsing: rules can match HTTP methods, paths, etc. For example, a CiliumNetworkPolicy can allow only specific HTTP methods/paths on a port. |
| **Default policies and admin policies** | If **no** ALLOW policy matches, traffic is **allowed** by default. You can deploy an overridable - default deny by default by deploying an **allow-nothing** policy in either the namespace or istio-system <br/><br/>AuthorizationPolicies in the `istio-system` namespace apply to the whole mesh and take precedence. These are not overridable by namespace-level policies.  | Default inbound policy can be set at install time using `proxy.defaultInboundPolicy`. Supported values are: <ul><li>`all-unauthenticated:` allow all traffic. This is the default.</li>  <li>`all-authenticated:` allow traffic from meshed clients in the same or from a different cluster (with multi-cluster).</li>  <li>`cluster-authenticated:` allow traffic from meshed clients in the same cluster.</li>  <li>`cluster-unauthenticated:` allow traffic from both meshed and non-meshed clients in the same cluster.</li>  <li>`deny:` all traffic are denied. </li> <li>`audit:` Same as all-unauthenticated but requests get flagged in logs and metrics.</li> </ul> <br/>Users can override the default policies for namespaces/pods or by setting the [config.linkerd.io/default-inbound-policy](http://config.linkerd.io/default-inbound-policy) annotation There is no support for admin, non overridable policies. | Follows Kubernetes NetworkPolicy semantics by default: if no `CiliumNetworkPolicy` allows the traffic, it is allowed (no implicit deny). Once at least one `CiliumNetworkPolicy` or `CiliumClusterwideNetworkPolicy` allows some traffic, all other traffic is implicitly denied.
<br/><br/> Operators must apply explicit deny rules or “default-deny” policies to block traffic in the absence of allow rules. <br/><br/> `CiliumClusterwideNetworkPolicy` exists for whole-cluster enforcement.)|


Every mesh vendor has their own API of such authorization. Below we describe brief UX for different implementations:

#### Istio
For the full spec and sematics of Istio AuthorizationPolicy: [Istio authorization policy docs](https://istio.io/latest/docs/reference/config/security/authorization-policy/)

Istio's AuthorizationPolicy can enforce access control by specifying allowed istio-formatted identities using the `source.principals` field, which matches authenticated service account identities via mTLS. You can also use other source constructs which are described in the table above and in https://istio.io/latest/docs/reference/config/security/authorization-policy/#Source.

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

Beyond what's explained in the table above, Cilium also automatically labels each pod with its associated service account using the label io.cilium.k8s.policy.serviceaccount. This label can be used in CiliumNetworkPolicy to enforce identity-based access controls using [ServiceAccounts Based Identities](https://docs.cilium.io/en/latest/security/policy/kubernetes/#serviceaccounts) within CiliumNetworkPolicy;

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



## Conformance Details


#### Feature Names


### Conformance tests 


## Alternatives


## References