# GEP-3792: External Gateways

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
cluster, for various reasons which are out of the scope of this GEP. Chihiro
and Ian want to be able to use these out-of-cluster proxies effectively and
safely, though they recognize that this may require additional configuration.

[Chihiro]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian

### Nomenclature and Background

In this GEP, we will use _out-of-cluster Gateway_ (OCG) to refer to a
conformant implementation of Gateway API's `GATEWAY` profile that's running
outside of the cluster. This would most commonly be a managed implementation
from a cloud provider, but of course there are many other possibilities.

We'll also distinguish between _mTLS meshes_, which rely on standard mTLS for
secure communications including workload authentication, and _non-mTLS
meshes_, which do anything else. We'll focus on mTLS meshes in this GEP; this
isn't because of a desire to exclude non-mTLS meshes, but because we'll have
enough trouble just wrangling the mTLS meshes! Supporting non-mTLS meshes will
be a separate GEP.

Finally, _authentication_ is the act of verifying the identity of some
_principal_; what the principal actually is depends on context. For this GEP
we will primarily be concerned with _workload authentication_ (also known as
_east/west auth_), in which the principal is a workload. (For completeness'
sake, in _user authentication_ AKA _N/S auth_, the principal is the human on
whose behalf a piece of technology is acting. We're only concerned with it
here in that meshed workloads can't trust what the OCG says about user
authentication without successful workload auth of the OCG itself.)

## Goals

- Allow Chihiro and Ian to configure an OCG and a mesh such that the OCG can
  usefully participate in the mesh, including:

    - The OCG must be able to securely communicate with meshed workloads in
      the cluster, where "securely communicate" includes encryption,
      authentication, and integrity checking.

    - The OCG must have a proper identity within the mesh, so that the mesh
      can apply authorization policy to requests from the OCG.

    - Whatever credentials the OCG and the mesh use to authenticate each other
      must be able to be properly maintained over time (for example, if they
      use mTLS, certificates will need rotation over time).

- Allow Ana to develop and operate meshed applications without needing to know
  whether the Gateway they're using is an OCG or an in-cluster Gateway.

- Define a basic set of requirements for OCGs and meshes that want to
  interoperate with each other (for example, the OCG and the mesh will likely
  need to agree on how workload authentication principals are represented).

## Non-Goals

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
  non-meshed workloads is also adjacent to the OCG situation, but it
  introduces a huge amount of additional configuration complexity.

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
least three significant problems. Thankfully, these are mostly problems for
Chihiro -- if we do our jobs correctly, Ana will never need to know.

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

#### 3. The Service Identity Problem

The OCG will need to speak directly to endpoints in the cluster, as described
above. This will prevent most meshes from being able to tell which service was
originally requested, which in turn means that those meshes will likely lose
any ability to properly perform retries, timeouts, etc. because they won't be
able to find the configuration to apply to the request.

A conformant OCG should be able to supply this functionality on its own, so
this isn't likely to be a functional problem. It does create a rough spot in
terms of configuration, though: if a given endpoint will be used by both the
OCG or other meshed workloads, you'll need to either provide two Routes with
the same configuration, or you'll need to provide a single Route with multiple
`parentRef`s.

## API

Most of the API work for this GEP is TBD at this point, but there are two
important points to note:

First, Gateway API has never defined a Mesh resource because, to date, it's
never been clear what would go into it. This may be the first configuration
item that causes us to need a Mesh resource.

Second, since the API should affect only Gateway API resources, it is not a
good candidate for policy attachment. It is likely to be much more reasonable
to simply provide whatever extra configuration we need inline in the Gateway
or Mesh resources.

### Gateway for Ingress (North/South)

### Gateway For Mesh (East/West)

## Conformance Details

#### Feature Names

This GEP will use the feature name `MeshOffClusterGateway`, under the
assumption that we will indeed need a Mesh resource.

### Conformance tests

## Alternatives

## References
