# GEP-3951: Minimal Out-of-Cluster Gateway API

* Issue: [#3951](https://github.com/kubernetes-sigs/gateway-api/issues/3951)
* Status: Provisional

See [status definitions](../overview.md#gep-states).

[Chihiro]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian
[Ana]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## User Story

[GEP-3792] defines the rationale
for allowing out-of-cluster Gateways (OCGs)
to participate in a
GAMMA-compliant in-cluster service mesh,
and the problems that must be solved
to allow them to do so.
This GEP defines
an extremely minimal API
to permit experimentation
with OCGs and
in-cluster mTLS meshes.

Nomenclature,
background,
goals,
non-goals,
problems that must be solved,
and some discussion
of possible solutions to those problems
are all included
in [GEP-3792].

[GEP-3792]: https://gateway-api.sigs.k8s.io/geps/gep-3792/

## Goals of the Extra-Minimal API

- Allow [Chihiro] and [Ian]
  to configurate and operate
  an OCG and
  an in-cluster mTLS mesh
  that know how to work together
  to experiment with OCG support
  in Gateway API.

## Non-Goals

- Support production use of OCGs
  in Gateway API.

- Solve all of the problems
  defined in [GEP-3792].

This is an **extra-minimal** API.
Its purpose is to allow **experimentation**
with OCGs and in-cluster meshes,
**not** to provide
a production-ready solution.

Using this API in production
is **guaranteed** to result
in anguish, heartbreak, tears, and pain.

## Overview

The extra-minimal OCG API
solves two of the [GEP-3792] problems
in a very minimal way.

- It solves the
  [trust problem]
  by extending
  the Mesh and Gateway resources
  to permit specifying
  a _trust bundle_
  that contains the CA certificates
  that the OCG and the mesh
  will use to trust each other.

- It solves the
  [discovery problem]
  by adding a label selector
  to the Gateway resource
  that indicates which Routes
  are meshed.

- It does not solve the
  [protocol problem]
  or the
  [outbound behavior problem].

[trust problem]: https://gateway-api.sigs.k8s.io/geps/gep-3792/#1-the-trust-problem
[protocol problem]: https://gateway-api.sigs.k8s.io/geps/gep-3792/#2-the-protocol-problem
[discovery problem]: https://gateway-api.sigs.k8s.io/geps/gep-3792/#3-the-discovery-problem
[outbound behavior problem]: https://gateway-api.sigs.k8s.io/geps/gep-3792/#4-the-outbound-behavior-problem

### Additions to the Mesh Resource

The Mesh resource
gains an `ocg` stanza
containing a `trustBundle` field
that refers to a ConfigMap
that contains the CA certificate(s)
that the mesh should trust
when validating connections
from the OCG:

```yaml
...
spec:
  ...
  ocg:
    trustBundle:
      name: ocg-trust-bundle
      namespace: ocg-namespace
      # Key in Configmap; defaults to "ca-bundle.crt"
      bundleKey: ca-bundle.crt
```

#### Additions to the Gateway Resource

The Gateway resource
gains a `mesh` stanza
containing two fields:

- a `trustBundle` field
  that refers to a ConfigMap
  that contains the CA certificate(s)
  that the OCG should trust
  when validating connections
  from meshed peers

- a `labelSelector` field
  that indicates which Routes
  are meshed.

```yaml
...
spec:
  ...
  mesh:
    trustBundle:
      name: mesh-trust-bundle
      namespace: mesh-namespace
      # Key in Configmap; defaults to "ca-bundle.crt"
      bundleKey: ca-bundle.crt
    labelSelector:
      matchLabels:
        mesh: one-mesh-to-mesh-them-all
```

### Trust Bundles: Solving the Trust Problem

The trust problem is that
both the OCG and the mesh
need to be able to do mTLS verification
of connections arriving from the other.
The simplest solution to this problem
is to add a _trust bundle_
to the Gateway resource
and to the Mesh resource.

- The trust bundle
  in the Gateway resource
  will define the CA certificate(s)
  that the OCG
  should accept as trusted
  when validating connections
  from meshed peers.

- The trust bundle
  in the Mesh resource
  will define the CA certificate(s)
  that the mesh
  should accept as trusted
  when validating connections
  from the OCG.

This is a straightforward way
to permit each component
to verify the identity of the other,
which will provide
sufficient basis for verifying identity when
mTLS meshes are involved.

#### The `trustBundle` Stanza

The Mesh and Gateway resources
both use a common `trustBundle` stanza:

```yaml
trustBundle:
  name: configmap-name
  namespace: configmap-namespace
  # Key in Configmap; defaults to "ca-bundle.crt"
  bundleKey: ca-bundle.crt
```

The `name` field is always required.
The `namespace` field is required
in the Mesh resource,
but may be omitted
in the Gateway resource
if the ConfigMap is
in the same namespace
as the Gateway resource.

The ConfigMap referred to
by the `trustBundle` stanza
MUST contain
a PEM-encoded trust bundle
in the specified `bundleKey`,
for example:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ocg-trust-bundle
  namespace: ocg-namespace
data:
  ca-bundle.crt: |-
    -----BEGIN CERTIFICATE-----
    ... (PEM-encoded CA certificate) ...
    -----END CERTIFICATE-----
    ... (may be repeated for multiple CA certificates) ...
```

The `trustBundle` in either
the Mesh resource
or the Gateway resource
may refer to a ConfigMap
in any namespace
to which RBAC permits access.

##### Further Considerations

The OCG and the mesh
MAY share the same trust bundle,
but this is not required.
If they do,
the Gateway and Mesh resources
MAY refer to the same ConfigMap;
if they do not,
they must (of course)
refer to different ConfigMaps
that contain the same CA certificate(s).

The `trustBundle` fields
MAY NOT refer to Secrets.
Since CA certificates are not private,
they should not be stored in Secrets.

An alternative to adding
the `trustBundle` stanza
to both the Mesh and Gateway resources
would be to define a single trust bundle,
requiring the OCG and the mesh
to each use the same CA certificate.
This adds considerable operational complexity -
especially in the world of enterprise PKI -
without any real benefit.

### Label Selectors: Solving the Discovery Problem

The discovery problem is that
not every workload in the cluster
is required to be meshed,
and the OCG needs a way
to know which Routes are meshed
since it must ensure that it
correctly uses mTLS
for connections to meshed workloads.

In practice, this isn't
actually a question of _workloads_
but of _Routes_:
the point of interface
between a Gateway
and a workload in the cluster
is not a Pod or a Service, but
rather a Route.

The extra-minimal API
solves this problem
by adding a label selector
to the Gateway resource
that indicates which Routes
are meshed.
When the OCG connects
to any Route
that either directly matches this selector,
or is in a namespace that matches this selector,
it MUST use mTLS
with a certificate
that is ultimately signed
by a CA certificate
in the Mesh resource's `trustBundle`,
and the OCG MUST validate
that the peer presents a certificate
that is ultimately signed
by a CA certificate
in the Gateway resource's `trustBundle`.

The label selector
is a simple mechanism
(especially if
the mesh already uses a label
to indicate which resources are meshed)
but it is still an active choice,
rather than assuming
things about the whole cluster.
Additionally, it permits
operating at the namespace level
or at the Route level.

### Other Problems

The extra-minimal API
does not solve the
[protocol problem]
or the
[outbound behavior problem].
Instead, it assumes that
the OCG and the mesh
have prearranged
protocols and behaviors
that are mutually compatible.

## Graduation Criteria

Since [GEP-3792] mandates
that any OCG API
MUST solve all four problems
defined in [GEP-3792]
before graduating to standard,
this GEP
MUST NOT graduate to standard
without significant further work.

## Conformance Details

#### Feature Names

This GEP will use the feature names
`GatewayExtraMinimalOCG` and
`MeshExtraMinimalOCG`.

### Conformance tests

TBA.
