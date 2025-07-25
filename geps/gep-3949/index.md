# GEP-3949: Mesh Resource

* Issue: [#3949](https://github.com/kubernetes-sigs/gateway-api/issues/3949)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

[Chihiro]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian
[Ana]: https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## User Story

**[Chihiro] and [Ian] would like a Mesh resource,
parallel to the Gateway resource,
that allows them to
supply mesh-wide configuration
and
shows what features
a given mesh implementation supports.**

## Background

Gateway API has long had a GatewayClass resource
that represents a class of Gateways
that can be instantiated in a cluster.
GatewayClass both
allows configuring the class as a whole
and provides a way for [Chihiro] and [Ian] to see
what features Gateways in that class support.
We have,
to date,
avoided such a resource for meshes,
but as we work on
improving mesh conformance tests and reports
and start work on
supporting Out-of-Cluster Gateways (OCGs),
we will need ways to
show what features a given mesh implementation supports
and represent mesh-wide configuration.

Unlike Gateways, we do not expect
multiple instances of meshes to be instantiated
in a single cluster.
This implies that a MeshClass resource is not needed;
instead, we will simply define a Mesh resource.

## Goals

- Define a Mesh resource
  that allows for
  mesh-wide configuration
  and feature discovery.

## Non-Goals

- Support multiple instances of a mesh
  in a single cluster at the same time.

   At some point, we may choose to
   change this goal,
   but it is definitely out of scope
   for this GEP.

- Support meshes interoperating with each other.

   As always,
   we will not rule out future work
   in this area,
   but it is out of scope
   for this GEP.

- Support off-cluster gateways.

   This is covered in a separate GEP
   and will not be discussed here.

## API

The purpose
of the Mesh resource
is to support both
mesh-wide configuration
as well as feature discovery.
However,
as of the writing of this GEP,
there is
no mesh-wide configuration
that is portable across implementations.
Therefore,
the Mesh resource
is currently pretty simple:

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XMesh
metadata:
  name: one-mesh-to-mesh-them-all
spec:
  # required, must be domain-prefixed
  controllerName: one-mesh.example.com/one-mesh
  parametersRef:
    # optional ParametersReference
    ...
```

- Although we call this the Mesh resource,
  as an experimental API
  it must be named XMesh
  in the `gateway.networking.x-k8s.io` API group.

   When the API graduates to standard,
   it will be renamed to `Mesh`
   in the `gateway.networking.k8s.io` API group.

- The `controllerName` field
  is analogous to
  the `controllerName` field
  in the GatewayClass resource:
  it defines the name
  of the mesh implementation
  that is responsible for
  this Mesh resource.

   Although we expect
   that there will be
   only one mesh
   in a given cluster, the
   `controllerName` field
   MUST be supplied,
   and a given mesh implementation
   MUST ignore
   a Mesh resource
   that does not have
   a `controllerName` field
   that matches its own name.

   If a MeshClass resource
   is later defined,
   the Mesh resource
   will gain a
   `meshClassName` field,
   the `controllerName` field
   will be deprecated,
   and a
   Mesh resource
   that includes
   both `controllerName` and `meshClassName`
   will be invalid.

- The `parametersRef` field
  is analogous to
  the `parametersRef` field
  in the GatewayClass resource:
  it allows specifying
  a reference to a resource
  that contains configuration
  specific to the mesh
  implementation.

### `status`

The `status` stanza
of the Mesh resource
is used to indicate
whether the Mesh resource
has been accepted by
a mesh implementation,
whether the mesh is
ready to use,
and
what features
the mesh supports.

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: XMesh
metadata:
  name: one-mesh-to-mesh-them-all
  namespace: mesh-namespace
spec:
  controllerName: one-mesh.example.com/one-mesh
status:
  conditions:
    # MUST include Accepted and Ready conditions if the Mesh resource is active.
    - type: Accepted  # Becomes true when the controller accepts the Mesh resource
      status: "True"
      reason: MeshAccepted
      lastTransitionTime: "2023-10-01T12:00:00Z"
      message: Mesh resource accepted by one-mesh v1.2.3 in namespace one-mesh
    - type: Ready     # Becomes true when the mesh is actually running
      status: "True"
      reason: MeshReady
      lastTransitionTime: "2023-10-01T12:00:15Z"
      message: one-mesh v1.2.3 is running
    ...
  supportedFeatures:
    # List of SupportedFeature
    - name: MeshHTTPRoute
    - name: MeshConsumerRoute
    - name: OffClusterGateway
    ...
```

Although it is
explicitly
not a goal of this GEP
to support multiple meshes
running in the same cluster at the same time,
meshes still MUST provide
human-readable information
in the `Accepted` condition
about which mesh instance
has claimed a given Mesh resource,
and SHOULD provide human-readable
information in the `Ready` condition,
in support of future expansion work.
This information is meant to be used
by [Chihiro] and [Ian] as confirmation
that the mesh instance
is doing what they expect it to do.

The mesh implementation
MUST set `status.SupportedFeatures`
to indicate which features
the mesh supports.

### API Type Definitions

TBA.

## Conformance Details

TBA.

#### Feature Names

No feature name is defined
for the Mesh resource itself;
filling out the `status` stanza
of the Mesh resource
is a conformance requirement,
and is sufficient indication
that the Mesh resource is supported.

### Conformance tests

TBA.

## Alternatives

We did not find any
particularly compelling alternatives
to having a Mesh resource
to meet these needs.
We considered having both
MeshClass and Mesh resources,
but decided that
there was no clear need for both,
and that a Mesh resource
better served the use cases.

## References

TBA.
