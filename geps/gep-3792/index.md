# GEP-3792: Out-of-Cluster Gateways

* Issue: [#3792](https://github.com/kubernetes-sigs/gateway-api/issues/3792)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

## User Story

**[Chihiro] and [Ian] want a way for out-of-cluster Gateways to be able to
usefully participate in a GAMMA-compliant in-cluster service mesh.**

Historically, API gateways and ingress controllers have often been implemented
using a Service of type LoadBalancer fronting a Kubernetes pod running a
proxy. This is simple to reason about, easy to manage for sidecar meshes, and
will presumably be an important implementation mechanism for the foreseeable
future. Some cloud providers, though, are moving the proxy outside of the
cluster, for various reasons which are out of the scope of this GEP. [Chihiro]
and [Ian] want to be able to use these out-of-cluster proxies effectively and
safely, though they recognize that this may require additional configuration.

[Chihiro]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian
[Ana]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana

### Nomenclature and Background

In this GEP:

1. We will use _out-of-cluster Gateway_ (OCG) to refer to a conformant
   implementation of Gateway API's `GATEWAY` profile that's running outside of
   the cluster. This would most commonly be a managed implementation from a
   cloud provider, but of course there are many other possibilities -- and in
   fact it's worth noting that anything we define here to support OCGs could
   also be used by workloads that run in-cluster but which, for whatever
   reason, can't be brought into the mesh in the mesh's usual way.

2. We'll also distinguish between _mTLS meshes_, which rely on standard mTLS
   for secure communication (authentication, encryption, and integrity
   checking) between workloads, and _non-mTLS meshes_, which do anything else.
   We'll focus on mTLS meshes in this GEP; this isn't because of a desire to
   exclude non-mTLS meshes, but because we'll have enough trouble just
   wrangling the mTLS meshes! Supporting non-mTLS meshes will be a separate
   GEP.

   **Note:** It's important to separate mTLS and HTTPS here. Saying that the
   mTLS meshes use mTLS for secure communication does not preclude them from
   using custom protocols on top of mTLS, and certainly does not mean that
   they must use only HTTPS.

3. _Authentication_ is the act of verifying the identity of some _principal_;
   what the principal actually is depends on context. For this GEP we will
   primarily be concerned with _workload authentication_, in which the
   principal is a workload, as opposed to _user authentication_, in which the
   principal is the human on whose behalf a piece of technology is acting. We
   expect that the OCG will handle user auth, but of course meshed workloads
   can't trust what the OCG says about the user unless the OCG successfully
   authenticates itself as a workload.

   **Note:** A single workload will have only one identity, but in practice we
   often see a single identity being used for multiple workloads (both because
   multiple replicas of a single workload need to share the same identity, and
   because some low-security workloads may be grouped together under a single
   identity).

4. Finally, we'll distinguish between _inbound_ and _outbound_ behaviors.

   Inbound behaviors are those that are applied to a request _arriving_ at a
   given workload. Authorization and rate limiting are canonical examples
   of inbound behaviors.

   Outbound behaviors are those that are applied to a request _leaving_ a
   given workload. Load balancing, retries, and circuit breakers are canonical
   examples of outbound behaviors.

## Goals

- Allow [Chihiro] and [Ian] to configure an OCG and a mesh such that the OCG
  can usefully participate in the mesh, including:

    - The OCG must be able to securely communicate with meshed workloads in
      the cluster, where "securely communicate" includes encryption,
      authentication, and integrity checking.

    - The OCG must have a proper identity within the mesh, so that the mesh
      can apply authorization policy to requests from the OCG.

    - Whatever credentials the OCG and the mesh use to authenticate each other
      must be able to be properly maintained over time (for example, if they
      use mTLS, certificates will need rotation over time).

    - The OCG must be able to distinguish meshed workloads from non-meshed
      workloads, so that it can communicate appropriately with each.

- Allow [Ana] to develop and operate meshed applications without needing to
  know whether the Gateway she's using is an OCG or an in-cluster Gateway.

- Define a basic set of requirements for OCGs and meshes that want to
  interoperate with each other (for example, the OCG and the mesh will likely
  need to agree on how workload authentication principals are represented).

- Define how responsibility is shared between the OCG and the mesh for
  outbound behaviors applied to requests leaving the OCG. (Note that "the OCG
  has complete responsibility and authority over outbound behaviors for
  requests leaving the OCG" is very much a valid definition.)

## Non-Goals

- Support multicluster operations. It may be the case that functional
  multicluster (with, e.g., a single OCG fronting multiple clusters) ends up
  falling out of this GEP, but it is not a goal.

- Support meshes interoperating with each other. It's possible that this GEP
  will lay a lot of groundwork in that direction, but it is not a goal.

- Support multiple meshes running in the same cluster. This GEP assumes a
  single mesh per cluster.

- Support non-mTLS meshes in Gateway API 1.4. We'll make every effort not to
  rule out non-mTLS meshes, but since starting with the mTLS meshes should
  tackle a large chunk of the industry with a single solution, that will be
  the initial focus.

- Solve the problem of extending a mesh to cover non-Kubernetes workloads (AKA
  _mesh expansion_). In many ways, mesh expansion is adjacent to the OCG
  situation, but the where the OCG is aware of the cluster and mesh, mesh
  expansion deals with a non-Kubernetes workload that is largely not aware of
  either.

- Solve the problem of how to support an OCG doing mTLS directly to a
  _non_-meshed workload (AKA the _backend TLS problem_). Backend TLS to
  non-meshed workloads is also adjacent to the OCG situation, but its
  configuration has different needs: backends terminating TLS on their own are
  likely to need per-workload configuration of certificates, cipher suites,
  etc., where the mesh as a whole should share a single configuration.

- Prevent the OCG API from being used by an in-cluster workload. We're not
  going to make in-cluster workloads a primary use case for this GEP, but
  neither are we disallowing them.

## Overview

Making an OCG work with an in-cluster mesh at the most basic level doesn't
really require any special effort. As long as the OCG has IP connectivity to
pods in the cluster, and the mesh is configured with permissive security, the
OCG can simply forward traffic from clients directly to meshed pods, and
things will "function" in that requests from clients, through the OCG, can be
handled by workloads in the cluster.

Of course, this sort of non-integration has obvious and terrible security
implications, since the traffic between the OCG and the application pods in
the cluster will be cleartext in the scenario above. The lack of encryption is
awful in its own right, but the fact that any mTLS mesh uses mTLS for
_authentication_ also means that the mesh loses any way to enforce
authorization policy around the OCG. Combined, these items amount to a major
problem.

An additional concern is that the OCG needs to be able to implement features
(e.g. sticky sessions) which require it to speak directly to endpoint IPs,
which can limit what the mesh will be able to do. This is likely a more minor
concern since a conformant OCG should itself be able to provide advanced
functionality; however, at minimum it can create some friction in
configuration.

### The Problems

To allow the OCG to _usefully_ participate in the mesh, we need to solve at
least four significant problems. Thankfully, these are mostly problems for
[Chihiro] -- if we do our jobs correctly, [Ana] will never need to know.

#### 1. The Trust Problem

The _trust problem_ is fairly straightforward to articulate: the OCG and the
mesh both need access to whatever information will allow each of them to trust
the other.

In the case of mTLS meshes, we are helped by the fact that basically every OCG
candidate already speaks mTLS, so the trust problem becomes "only" one of
setting things up for the OCG and the mesh to each include the other's CA
certificate in their trust bundle. (They may be using the same CA certificate,
but we shouldn't rely on that.)

In the case of non-mTLS meshes, the trust problem is more complex; this is the
major reason that this GEP is focused on mTLS meshes.


#### 2. The Protocol Problem

The _protocol problem_ is that the data-plane elements of the mesh may assume
that they'll always be talking only to other mesh data-plane elements, which
the OCG will not be. If the mesh data-plane elements use a specific protocol,
then either the OCG will need to speak that protocol, or the mesh will need to
relax its requirements (perhaps on a separate port?) to accept requests
directly from the OCG.

For example, Linkerd and Istio Legacy both use standard mTLS for
proxy-to-proxy communication -- however, both also use ALPN to negotiate
custom (and distinct!) "application" protocols during mTLS negotiation, and
depending on the negotiated protocol, both can require the sending proxy to
send additional information after mTLS is established, before any client data
is sent. (For example, Linkerd requires the originating proxy to send
transport metadata right after the TLS handshake, and it will reject a
connection which doesn't do that correctly.)

#### 4. The Discovery Problem

When using a mesh, not every workload in the cluster is required to be meshed
(for example, it's fairly common to have some namespaces meshed and other
namespaces not meshed, especially during migrations). The _discovery problem_
here is that the OCG needs to be know which workloads are meshed, so that it
can choose appropriate communication methods for them.

#### 4. The Outbound Behavior Problem

The OCG will need to speak directly to endpoints in the cluster, as described
above. This will prevent most meshes from being able to tell which service was
originally requested, which makes it impossible for the mesh to apply outbound
behaviors. This is the _outbound behavior problem_: it implies that either the
OCG must be responsible for outbound behaviors for requests leaving the OCG
for a meshed workload, or that the OCG must supply the mesh with enough
information about the targeted service to allow the mesh to apply those
outbound behaviors (if that's even possible: sidecar meshes may very well
simply not be able to do this.)

This is listed last because it shouldn't be a functional problem to simply
declare the OCG solely responsible for outbound behaviors for requests leaving
the OCG. It is a UX problem: if a given workload needs to be used by both the
OCG or other meshed workloads, you'll need to either provide two Routes with
the same configuration, or you'll need to provide a single Route with multiple
`parentRef`s.

## Graduation Criteria

In addition to the [general graduation
criteria](../concepts/versioning.md#graduation-criteria), this GEP must also
guarantee that **all four** of the problems listed above need resolutions, and
must have implementation from at least two different Gateways and two
different meshes.

## API

There are three important aspects to the OCG API:

1. The API must allow for mutually authenticated communication between the OCG
   and meshed workloads. This includes providing the OCG with the information
   it needs to both authenticate itself to workloads, and to authenticate the
   workloads it communicates with.

2. The API must allow the mesh to be configured to accept requests from the
   OCG, including providing the mesh with the information it needs to
   authenticate the OCG as a workload.

3. Since the API should affect only Gateway API resources, it is not a good
   candidate for policy attachment. It is likely to be much more reasonable to
   simply provide whatever extra configuration we need inline in the Gateway
   or Mesh resources.

The API must also solve all four of the problems listed above, so we'll start
with an overview of the solutions before diving into the API details.

### Solving the Trust Problem

The trust problem is that the OCG and the mesh need to be able to trust each
other. The simplest solution to this problem is to add a _trust bundle_ to the
Gateway resource, and to create a Mesh resource which will also have a trust
bundle.

This is a straightforward way to permit each component to verify the identity
of the other, which will provide sufficient basis for verifying identity when
mTLS meshes are involved.


This leads to a critical architectural decision: What is the most secure and
reliable method for the Gateway and the Mesh to obtain each other's trust
bundles?

Two primary models were considered to answer this question:

#### Proposal 1: Configuration Model (Recommended)

This model is an explicit grant of trust, where an administrator directly
configures the relationship.

How it works:

* The Gateway resource is configured with a reference to the Mesh's trust bundle
  Secret.

* The Mesh resource is configured with a reference to the Gateway's trust bundle
  Secret.

Pro: Highly Secure. An administrator's explicit action prevents impersonation
attacks. There is no ambiguity about who to trust.

Pro: Simple & Clear. The configuration is direct and easy to understand, even in
complex clusters with multiple gateways.


```yaml
# 1. The Mesh's Trust Bundle Secret
# This Secret contains the mesh's CA certificate, which the Gateway will be configured to trust.
# The certificate data must be base64-encoded.
apiVersion: v1
kind: Secret
metadata:
  name: mesh-ca-secret
  # This should be in a namespace the Gateway's controller can access.
  namespace: mesh-system
type: Opaque
data:
  # The key can be anything, but 'ca.crt' is a common convention.
  ca.crt: |
    # Paste the base64-encoded CA certificate for the
    # in-cluster service mesh here. For example:
    # LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN...
---
# 2. The Gateway's Trust Bundle Secret
# This Secret contains the OCG's CA certificate, which the Mesh will be configured to trust.
apiVersion: v1
kind: Secret
metadata:
  name: ocg-ca-secret
  # This should be in a namespace the mesh's controller can access.
  namespace: gateway-system
type: Opaque
data:
  ca.crt: |
    # Paste the base64-encoded CA certificate for the
    # Out-of-Cluster Gateway (OCG) here.
    # LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN...
---
# 3. The Gateway Resource (Trusting the Mesh)
# The OCG controller reads this and learns to trust the mesh by
# fetching the 'mesh-ca-secret'.
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ocg-gateway
spec:
  gatewayClassName: my-ocg-class
  # ... other gateway spec fields ...
  mesh:
    trustBundle:
      # Reference to the Secret containing the mesh's CA.
      # Note: The 'kind' defaults to Secret if not specified,
      # but it is shown here for clarity.
      kind: Secret
      name: mesh-ca-secret
      namespace: mesh-system # Namespace of the Secret
---
# 4. The Mesh Resource (Trusting the Gateway)
# The mesh controller reads this and learns to trust the OCG by
# fetching the 'ocg-ca-secret'.
apiVersion: networking.x-k8s.io/v1alpha1 # Hypothetical API version for Mesh resource
kind: Mesh
metadata:
  name: in-cluster-mesh
spec:
  ocg:
    trustBundle:
      # Reference to the Secret containing the OCG's CA.
      kind: Secret
      name: ocg-ca-secret
      namespace: gateway-system # Namespace of the Secret
```

#### Proposal 2: Discovery Model (Alternative)

This model relies on publication of identity, where each component announces itself and the other must find it.

How it works:

* The Gateway resource points to a Secret containing its own identity.

* The Mesh resource points to a Secret containing its own identity.

* Each controller must then find the other's resource to learn which identity to trust.

Con: Security Risk. An attacker can create a fake Mesh or Gateway resource. The discovery process could mistakenly trust this malicious identity, leading to a security breach.

Con: Operationally Complex. It's unclear how the system should behave when multiple gateways exist. This ambiguity makes the system more fragile.



#### Proposal 3: Single Trust Bundle for OCG and Mesh

This proposal defines a single trust bundle, requiring the OCG and
the mesh to each use the same CA certificate. This adds considerable operational
complexity (especially in the world of enterprise PKI) without any real benefit.

#### Proposal 4: Symmetrical Model with ClusterTrustBundle (Recommended, Forward-Looking)

This model represents the ideal state, assuming a modern environment where both
the cluster and the OCG implementation can leverage the ClusterTrustBundle
resource.

How it works:

* The Mesh's CA is placed into a ClusterTrustBundle (e.g.,
mesh-identity-bundle). The Gateway resource is configured to trust this bundle
by name. The OCG controller reads this bundle to establish trust.

* The Gateway's CA is placed into a separate ClusterTrustBundle (e.g.,
gateway-identity-bundle). The Mesh resource is configured to trust this bundle
by name. The in-cluster mesh proxies mount this bundle to establish trust.

Pro: Architecturally Elegant & Symmetrical. This is the cleanest approach. It
uses the same modern, purpose-built Kubernetes primitive (ClusterTrustBundle)
for both sides of the trust exchange.

Pro: Centralized Trust Management. It allows administrators to manage all major
cluster trust anchors in one place as ClusterTrustBundle resources. These
bundles can then be easily reused by other applications in the cluster, not just
the mesh or gateway.

Pro: Highly Secure. It is still an explicit configuration model, inheriting all
the security benefits of Proposal 1 and completely avoiding the risks of the
discovery model.

Con: Strongest Requirement on Environment. This is the most forward-looking
approach and requires two conditions: a modern Kubernetes cluster (v1.29+ for
Beta support) and an OCG implementation that explicitly supports reading
ClusterTrustBundle resources via the API.

```yaml
# 1. The Mesh's Trust Bundle
# This bundle represents the mesh's identity. It will be referenced by the Gateway.
apiVersion: certificates.k8s.io/v1beta1
kind: ClusterTrustBundle
metadata:
  name: mesh-identity-bundle
spec:
  signerName: my-mesh.io/identity
  trustBundle: |
    # -----BEGIN CERTIFICATE-----
    #
    #   The public PEM-encoded CA certificate for the
    #   in-cluster service mesh goes here.
    #
    # -----END CERTIFICATE-----
---
# 2. The Gateway's Trust Bundle
# This bundle represents the OCG's identity. It will be referenced by the Mesh.
apiVersion: certificates.k8s.io/v1beta1
kind: ClusterTrustBundle
metadata:
  name: ocg-identity-bundle
spec:
  signerName: my-cloud-provider.com/ocg
  trustBundle: |
    # -----BEGIN CERTIFICATE-----
    #
    #   The public PEM-encoded CA certificate for the
    #   Out-of-Cluster Gateway (OCG) goes here.
    #
    # -----END CERTIFICATE-----
---
# 3. The Gateway Resource (Trusting the Mesh)
# The OCG controller reads this and learns to trust the mesh by
# fetching the 'mesh-identity-bundle'.
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: ocg-gateway
spec:
  gatewayClassName: my-ocg-class
  # ... other gateway spec fields ...
  mesh:
    trustBundle:
      # Reference to the mesh's identity bundle.
      kind: ClusterTrustBundle
      name: mesh-identity-bundle
---
# 4. The Mesh Resource (Trusting the Gateway)
# The mesh controller reads this and learns to trust the OCG by
# configuring its sidecars to use the 'ocg-identity-bundle'.
apiVersion: networking.x-k8s.io/v1alpha1 # Hypothetical API version for Mesh resource
kind: Mesh
metadata:
  name: in-cluster-mesh
spec:
  ocg:
    trustBundle:
      # Reference to the OCG's identity bundle.
      kind: ClusterTrustBundle
      name: ocg-identity-bundle
```


### Solving the Protocol Problem

The protocol problem is that the OCG needs a way to indicate to the mesh that
it intends to participate in the mesh for a given connection, and the mesh
needs to accept the OCG's participation.

As a starting point for OCG/mTLS mesh interaction:

- The OCG MUST use an mTLS connection to communicate with meshed workloads.

- The OCG MUST use an mTLS certificate ultimately signed by a certificate in
  the trust bundle provided to the mesh.

- The mesh MUST use an mTLS certificate ultimately signed by a certificate in
  the trust bundle provided to the OCG.

- The OCG MUST send the `ocg.gateway.networking.k8s.io/v1` ALPN protocol
  during mTLS negotiation. The mesh MUST interpret this ALPN selection as a
  signal that the OCG intends to participate in the mesh.

- The OCG MUST NOT send any additional information in the data stream before
  client data. (This is a contrast from e.g. Linkerd's default behavior, and
  has implications for the outbound behavior problem.)

This will (obviously) require implementation work on the part of both the OCG
implementations and the mesh implementations, so it's worth looking at some
alternatives:

* The OCG could be required to exactly mimic the ALPN/transport metadata
combination used by an existing mesh. However, the existing meshes don't share a
single common mechanism, so this would require a lot of work on the OCG's part,
and it wouldn't be portable between meshes.

* The OCG could simply skip ALPN, and hand a "bare" mTLS connection to the
mesh. In general, existing meshes don't support this in the way that we want:
depending on the destination of the connection, they may interpret it as
application-level mTLS that they should treat as an opaque data stream, or they
may simply refuse it. This alternative, therefore, both shifts the entire burden
of implementation to the meshes, and probably makes it impossible to correctly
handle application-level mTLS. Whether meshes should support application-level
mTLS in this way is a separate discussion, and is out of scope for this GEP.

* We could perhaps abuse the [PROXY protocol] for this, or define something
similar. This would appear to increase the implementation burden on both sides
without providing appreciable benefit.

* A configuration knob could be added to the Gateway resource, allowing an
administrator to specify a custom ALPN or protocol variant. For example, a user
running Istio could configure their OCG to use the istio-http/1.1 ALPN. While
this provides a path to interoperability with unmodified, existing meshes, it
comes with significant downsides. It places a large implementation burden on OCG
providers, who would need to support the distinct transport protocols of
multiple meshes. Furthermore, it would make the Gateway configuration
non-portable across different service meshes and would require the user to be
aware of low-level implementation details of their mesh.

#### Initial Approach

The protocol problem addresses how an Out-of-Cluster Gateway (OCG) and an
in-cluster mesh proxy should communicate at the application protocol
level. While a future release may introduce a standardized protocol for richer
metadata exchange, the initial goal is to establish a functional baseline that
is simple for both OCG and mesh implementers.

To lower the barrier to entry, the full solution to this problem, which may or
may not include the definition of a standard `ocg.gateway.networking.k8s.io/v1`
ALPN, will be deferred to a future GEP, targeting Gateway API v1.5.

For the initial release, the requirements are simplified to rely on standard
mTLS without a custom protocol.

##### Protocol Requirements

* The OCG MUST use a standard mTLS connection to communicate with meshed
workloads. The OCG's identity is established by the client certificate it
presents, which the mesh proxy validates against its configured trust bundle.

* The OCG MUST NOT be required to use a mesh-specific ALPN (Application-Layer
Protocol Negotiation) value (e.g., istio-http/1.1, linkerd-proxy) or send any
proprietary transport-level metadata after the TLS handshake.

* The mesh proxy, after successfully authenticating the OCG via its client
certificate, MUST treat the connection as a standard application data stream
(e.g., HTTP/1.1, HTTP/2). Routing and policy decisions will be based on the
content of this stream, such as the HTTP Host or :authority header, which aligns
with standard GAMMA routing behavior.

##### Justification for Deferral

This simplified approach provides several key benefits for an initial release:

* Lowers Implementation Burden: OCG implementers are not required to support a
variety of mesh-specific protocols. Mesh implementers are only required to
handle a standard mTLS connection from a trusted peer, which is a common
interoperability requirement.

* Provides Immediate Value: This baseline solves the core security and routing
problems, allowing the OCG to securely participate in the mesh for standard HTTP
traffic.

* Allows for Future Enhancement: Deferring the decision allows the community to
gain practical experience with OCG-mesh interactions before standardizing on a
more advanced protocol that could support richer features, like forwarding
proxy-protocol data or other out-of-band metadata.

* Future Work (Post-Initial Release) A subsequent GEP will revisit the Protocol
Problem to define a standard mechanism. This will enable more advanced use cases
that require passing metadata between the OCG and the mesh proxy outside of the
primary application stream (e.g., original client IP, custom trace headers).

### Solving the Discovery Problem

The discovery problem is that not every workload in the cluster is required to
be meshed, which means that the OCG needs a way to know whether a given
connection must participate in the mesh or not. In practice, this isn't
actually a question of _workloads_ but of _Routes_: the point of interface
between a Gateway and a workload in the cluster is not a Pod or a Service, but
rather a Route.

We could approach this in a few different ways:

1. Assume that _all_ Routes in the cluster are meshed. This is obviously a
   nonstarter.

2. Assume that _all_ Routes bound to the OCG are meshed. This is also a
   nonstarter.

3. Add a field to the Route resource that indicates whether the Route is
   meshed. This feels like quite a bit of an imposition on [Ana] (and, again,
   if we do our jobs correctly, [Ana] shouldn't need to think about this).

4. Add a field to the Gateway resource that enumerates Routes that are meshed.
   This would be a lot of work for [Chihiro] and [Ian] to maintain; while
   that's better than putting the burden on [Ana], it still isn't good.

5. Add a field to the Gateway resource that enumerates namespaces that are
   meshed. This isn't quite as bad as option 4, but it's still a lot of work
   for [Chihiro] and [Ian], _and_ it limits us to only having entire
   namespaces meshed.

6. Chihiro adds a label selector to the Gateway resource, which declares that
   Routes with the label, and Routes in namespaces with the label, are meshed.

7. Chihiro defines a label selector for Meshed Namespaces on the Mesh Resource.


Let's explore options 6 and 7 in more detail:

#### Option 6: Selector on the Gateway (GEP Recommended)

In this model, Chihiro adds a selector field to the Gateway resource. The
Out-of-Cluster Gateway (OCG) uses this selector to determine if a Route is
meshed by checking the labels on the Route itself or its containing Namespace.

This approach is considered the least imposition on everyone involved for
several reasons:

* It's an active, explicit choice by the operator, rather than relying on risky
assumptions about the cluster state.

* It allows for flexible operation at both the namespace level (for broad policy)
and the Route level (for fine-grained exceptions).

* It efficiently reuses existing configurations, as it can leverage the same
labels a service mesh already uses for sidecar injection.

* It employs a standard, Kubernetes-native mechanism (LabelSelector) for
selection, which is familiar to all cluster operators.

Overall, option 6 is probably the least imposition on everyone.

#### Option 7: Namespace Selector on the Mesh (Alternative)

In this alternative model, Chihiro configures a selector on the Mesh resource to
specify which namespaces are part of the mesh. The OCG controller would then
need to find and read this resource to make its routing decisions. When routing
to a backend Route, the OCG would check if the Route's namespace is in the list
it discovered from the Mesh resource. If it is, the Route is considered meshed.

##### Tradeoffs and Justification for the GEP's Approach

While this alternative has the benefit of centralizing the mesh's definition in
one place, the GEP's proposed solution of using a selector on the Gateway
resource was chosen for several key reasons:

1. Controller Coupling - This is the most significant drawback. For this model
to work, the OCG controller must discover, watch, and interpret the Mesh
resource, which is managed by a completely different controller. This creates a
tight coupling between the two systems.

The GEP's Approach (Decoupled): The OCG controller only needs to look at its own
Gateway resource. All the information it needs to function is self-contained.

The Alternative (Coupled): The OCG's functionality becomes dependent on the Mesh
resource's API. A change to the Mesh resource could break the OCG. This makes
the overall system more brittle.

2. Reduced Granularity - This proposal limits the definition of "meshed" to the
namespace level. The GEP's solution of using a label selector on the Gateway is
more flexible because it can target either Namespaces or individual Route
resources. This allows an operator to, for example, have a mostly meshed
namespace but exclude one specific Route within it from mesh communication.

3. Duplication of Configuration - The service mesh already knows which
namespaces are part of the mesh—this is fundamental to its operation (e.g., for
sidecar injection). The GEP's label selector approach allows the Gateway to
reuse this existing source of truth (the labels). The alternative would require
an administrator to define this list of namespaces a second time in the Mesh
resource, which is less efficient and prone to configuration drift.

In summary, while placing the namespace list on the Mesh resource is
conceptually clean, the GEP's approach was chosen because it results in a more
decoupled, flexible, and robust system that avoids unnecessary controller
dependencies and configuration duplication.


### Solving the Outbound Behavior Problem

The outbound behavior problem is that the OCG will need to speak directly to
endpoints in the cluster, which will prevent most meshes from being able to
apply outbound behaviors directly.

As a starting point, we will explicitly declare that the OCG is responsible
for all outbound behaviors for meshed requests, and that it is OK for the mesh
to not be able to apply these behaviors. This leaves the UX problem that a
Route meant to apply equally to N/S traffic and E/W traffic will involve some
duplication of configuration, but all the alternatives create operational
problems.

- If the mesh is responsible for all outbound behaviors, what happens if the
  OCG needs to speak to a non-meshed Route? Would we require [Ana] to
  duplicate the Route, or (arguably worse) require the OCG to interpret GAMMA
  Routes?

- If the OCG and the mesh share responsibility in some way, how do we describe
  the split?

Overall, the alternatives to the OCG being responsible for all outbound
behaviors for requests leaving the OCG would all seem to create much worse
problems than the UX problem of having to duplicate configuration for a Route
that applies equally to N/S and E/W traffic.


### Open Questions

While this GEP provides a strong foundation for Out-of-Cluster Gateway (OCG) and
mesh interoperability, several questions remain open for discussion and
potential standardization in future revisions.

1. How to configure the Gateway's identity certificate and private key?

This GEP defines how the OCG and the mesh should be configured to trust each
other by exchanging CA bundles. However, it does not standardize how an
administrator configures the specific client certificate and private key that
the OCG uses to identify itself to the mesh.

Currently, this is left as an implementation detail, likely handled via a
provider-specific CRD referenced from the GatewayClass or through an out-of-band
mechanism. The open question is: Should a future version of this GEP standardize
this configuration to ensure a consistent user experience? This could involve
adding a new identityCertificateRef field to the Gateway spec.

2. Should use cases where mesh workloads disable mTLS be supported?

This GEP focuses on meshes where mTLS is strictly enforced for
communication. However, some service meshes support a "DISABLE" mode where mTLS
can be disabled for certain workloads. This raises the question: How should an
OCG behave when a target workload is discovered as "meshed" but does not require
or accept an mTLS connection?

3. What are the graduation requirements for the initial release?

The original graduation criteria required resolving all four major problems,
including the "Protocol Problem." Now that the standardization of a custom ALPN
protocol has been deferred to a future release (targeting v1.5), the graduation
criteria for this GEP should be re-evaluated.

The open question is: Should the graduation requirements be adjusted to reflect
the reduced scope? For example, should graduation now require successful
interoperability from at least two Gateway and two Mesh implementations using
only the standard "bare" mTLS approach, proving the core value of the trust and
discovery mechanisms without waiting for the advanced protocol work?

### API Details (North/South)

Since the OCG itself is assumed to be a conformant Gateway API implementation,
we can extend the Gateway resource to include the necessary configuration for
the OCG to securely communicate with meshed workloads in the cluster. We'll
add a new `mesh` field to the Gateway resource, which will currently allow
specifying two things:

1. A trust bundle that contains the CA certificate(s) that the OCG should use
   to verify workloads in the mesh.

2. A label selector that allows the OCG to find namespaces that are meshed.

For example:

```yaml
apiVersion: networking.x-k8s.io/v1
kind: Gateway
metadata:
  name: ocg-gateway
spec:
  gatewayClassName: ocg-gateway-class
  ...
  mesh:   # All mesh-related configuration goes here
    trustBundle:
      # List of SecretObjectReference; at least one is required
      - name: ocg-mesh-root-ca          # mandatory
        namespace: ocg-mesh-namespace   # defaults to the Gateway's namespace
      - ...
    selector:
      matchLabels:
        linkerd.io/inject: "true"       # or whatever label is appropriate
```

### API Details (East/West)

We'll define a new Mesh resource to allow the mesh to be configured to accept
requests from the OCG. The Mesh resource will allow specifying a trust bundle
that contains the CA certificate(s) that the mesh should use to verify
requests from the OCG.

(The Mesh resource should clearly also have a way for the mesh implementation
to indicate supported features, in parallel to the GatewayClass resource, but
that will be a separate GEP.)

Note that the Mesh resource doesn't need a selector for meshed workloads: the
mesh implementation will already understand this.

For example:

```yaml
apiVersion: networking.x-k8s.io/v1
kind: Mesh
metadata:
  name: ocg-mesh
spec:
  ocg:    # All OCG-related configuration goes here
    trustBundle:
      # List of SecretObjectReference; at least one is required
      - name: ocg-mesh-root-ca          # mandatory
        namespace: ocg-mesh-namespace   # defaults to the Mesh's namespace
      - ...
```

#### Mesh Resource Lifecycle and Management

A key operational question for the Mesh resource is who is responsible for its
creation and management. The lifecycle of this resource directly impacts the
user experience for [Chihiro], the cluster operator. We considered two primary
approaches.

##### Option 1: User-Created Resource

In this model, [Chihiro] would be fully
responsible for authoring a Mesh resource manifest from scratch and applying it
to the cluster. The resource would include a controllerName field to associate
it with the specific mesh implementation installed.

While this approach is explicit and requires clear intent from the user, it
presents a notable barrier to entry. The user must consult documentation to find
the correct API group, version, and fields, increasing the initial effort and
potential for error.

##### Option 2: Controller-Created, User-Configured Resource (Recommended)

A second, more user-friendly approach is a controller-managed lifecycle. In this
model, the service mesh controller, upon its installation, automatically creates
a single, default Mesh resource within the cluster. This resource would
initially be configured with sensible defaults, but with advanced features like
OCG integration disabled.

[Chihiro]'s workflow then shifts from creation to configuration. To enable OCG
integration, they would simply retrieve the existing Mesh resource (e.g., via
kubectl get mesh) and modify it to add the necessary trustBundle and other
OCG-related settings.


### API Type Definitions


This spec incorporates the flexibility to reference either a Secret or a ClusterTrustBundle, accommodating the different proposals we discussed.

#### Shared Types

First, we define a flexible reference type that can point to either a Secret or a ClusterTrustBundle. This will be used by both the Gateway and Mesh resources.

```Go

package v1alpha1 // or a relevant API version

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)


// TrustBundleReference is a reference to a resource containing a trust bundle.
// It can reference either a Secret or a ClusterTrustBundle. Only one of these
// fields should be set.
type TrustBundleReference struct {
	// Group is the group of the referent.
	// When empty, the core API group is assumed.
	// For ClusterTrustBundle, this would be "certificates.k8s.io".
	//
	// +optional
	Group *gwv1.Group `json:"group,omitempty"`

	// Kind is the kind of the referent.
	// Valid values are "Secret" and "ClusterTrustBundle".
	//
	// +required
	Kind *gwv1.Kind `json:"kind"`

	// Name is the name of the referent.
	//
	// +required
	Name gwv1.ObjectName `json:"name"`

	// Namespace is the namespace of the Secret referent.
	// This field is ignored for cluster-scoped resources like ClusterTrustBundle.
	//
	// +optional
	Namespace *gwv1.Namespace `json:"namespace,omitempty"`
}
```

#### Gateway API Spec Additions

The following shows the proposed mesh field to be added to the existing GatewaySpec.

```Go

// GatewaySpec defines the desired state of Gateway.
// This is an EXTENSION of the existing GatewaySpec.
type GatewaySpec struct {
    // ... existing fields like gatewayClassName, listeners, addresses ...

	// Mesh defines the configuration for Gateway participation in a service mesh.
	// If this field is unspecified, the Gateway does not participate in a mesh.
	//
	// +optional
	Mesh *GatewayMeshConfig `json:"mesh,omitempty"`
}

// GatewayMeshConfig defines the configuration for a Gateway's service mesh integration.
type GatewayMeshConfig struct {
	// TrustBundle defines the trust anchor(s) that this Gateway will use to
	// verify the identity of in-cluster mesh workloads.
	// This references the mesh's CA.
	//
	// +required
	TrustBundle TrustBundleReference `json:"trustBundle"`

	// Selector defines the criteria for determining which Routes are part of the mesh.
	// A Route is considered meshed if it or its parent Namespace matches this selector.
	// If unspecified, all Routes attached to this Gateway are considered meshed.
	//
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty"`
}
```

#### New Mesh API Resource Spec

This defines the new, cluster-scoped Mesh resource.

```Go

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster,shortName=mesh
// +kubebuilder:storageversion

// Mesh is a cluster-scoped resource that provides configuration for a service mesh
// to integrate with other components, like an Out-of-Cluster Gateway.
type Mesh struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Mesh.
	Spec MeshSpec `json:"spec,omitempty"`

	// Status defines the current state of the Mesh.
	// +optional
	Status MeshStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MeshList contains a list of Mesh resources.
type MeshList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mesh `json:"items"`
}

// MeshSpec defines the desired state of Mesh.
type MeshSpec struct {
	// ControllerName is the name of the controller that is managing this Mesh.
	// This field must be set.
	// For example: "my-mesh.io/controller".
	//
	// +required
	ControllerName gwv1.GatewayController `json:"controllerName"`

	// OCG defines the configuration for allowing an Out-of-Cluster Gateway
	// to securely participate in the mesh.
	//
	// +optional
	OCG *OCGConfig `json:"ocg,omitempty"`
}

// OCGConfig defines the configuration for Out-of-Cluster Gateway integration.
type OCGConfig struct {
	// TrustBundle defines the trust anchor(s) that the mesh will use to
	// verify the identity of the Out-of-Cluster Gateway.
	// This references the OCG's CA.
	//
	// +required
	TrustBundle TrustBundleReference `json:"trustBundle"`
}

// MeshStatus defines the observed state of Mesh.
type MeshStatus struct {
	// Conditions describe the current conditions of the Mesh.
	//
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

## Conformance Details

TBA.

#### Feature Names

This GEP will use the feature name `OffClusterGateway`, and MUST be listed in
both the `Gateway` and `Mesh` resources, so that [Chihiro] can know that
they're choosing a combination of OCG and mesh that will work together.

### Conformance tests

## Alternatives

## References
