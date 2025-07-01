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

1. The API must allow the OCG to be configured to securely communicate with
   meshed workloads in the cluster, including providing the OCG with the
   information it needs to authenticate itself as a workload.

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
other. The simplest solution to this problem is to add a _trust bundle_ the
Gateway resource, and to create a Mesh resource which will also have a trust
bundle:

- The trust bundle in the Gateway resource will define the CA certificate(s)
  that the OCG should trust when communicating with meshed workloads in the
  cluster.

- The trust bundle in the Mesh resource will define the CA certificate(s)
  that the mesh should trust when communicating with the OCG.

This is a straightforward way to permit each component to verify the identity
of the other, which will provide sufficient basis for verifying identity when
mTLS meshes are involved.

- An alternative would be to define a single trust bundle, requiring the OCG
  and the mesh to each use the same CA certificate. This adds considerable
  operational complexity (especially in the world of enterprise PKI) without
  any real benefit.

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
alternatives.

- The OCG could be required to exactly mimic the ALPN/transport metadata
  combination used by an existing mesh. However, the existing meshes don't
  share a single common mechanism, so this would require a lot of work on the
  OCG's part, and it wouldn't be portable between meshes.

- The OCG could simply skip ALPN, and hand a "bare" mTLS connection to the
  mesh. In general, existing meshes don't support this in the way that we
  want: depending on the destination of the connection, they may interpret it
  as application-level mTLS that they should treat as an opaque data stream,
  or they may simply refuse it. This alternative, therefore, both shifts the
  entire burden of implementation to the meshes, and probably makes it
  impossible to correctly handle application-level mTLS.

  (Whether meshes _should_ support application-level mTLS in this way is a
  separate discussion, and is out of scope for this GEP.)

- We could perhaps abuse the [PROXY protocol] for this, or define something
  similar. This would appear to increase the implementation burden on both
  sides without providing appreciable benefit.

[PROXY protocol]: https://github.com/haproxy/haproxy/blob/master/doc/proxy-protocol.txt

None of these alternatives would appear to be better than using ALPN in a
similar way to how (some) existing meshes already use it.

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

6. Add a label selector to the Gateway resource and declare that Routes with
   the label, and Routes in namespaces with the label, are meshed.

Overall, option 6 is probably the least imposition on everyone:

- It's still an active choice, rather than assuming things about the whole
  cluster.

- It allows operating at the namespace level or at the Route level.

- It takes advantage of the case where a mesh already uses a label to indicate
  which resources are meshed.

- It uses a reasonably Kubernetes-native mechanism for selection.

Therefore, we'll add a label selector to the Gateway resource, and OCG MUST
assume that any Route that either matches this selector, or is in a namespace
that matches this selector, is meshed.

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

### API Type Definitions

TBA.

## Conformance Details

TBA.

#### Feature Names

This GEP will use the feature name `OffClusterGateway`, and MUST be listed in
both the `Gateway` and `Mesh` resources, so that [Chihiro] can know that
they're choosing a combination of OCG and mesh that will work together.

### Conformance tests

## Alternatives

## References
