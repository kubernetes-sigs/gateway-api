---
description: >
  We are excited to announce the v0.8.0 release of Gateway API, where the GAMMA
  initiative has now reached Experimental status, conformance profiles are supported,
  and more!
---

# Gateway API: Welcome, GAMMA!

<small style="position:relative; top:-30px;">
  :octicons-calendar-24: August 01, 2023 ·
  :octicons-clock-24: 5 min read
</small>

We are thrilled to announce the v0.8.0 release of the Gateway API! With this
release, the work the GAMMA initiative has been doing over the past year has
reached [Experimental status][status], with two conformant implementations at
present (Linkerd and Istio). We look forward to your feedback!

As [Experimental][status] features, you will find the GAMMA CRDs in the
Gateway API [`experimental`][ch] channel.

## What is the GAMMA initiative?

Almost from the point where the Gateway API was itself implementable,
questions have been raised about how it could be used for configuring service
meshes. The GAMMA initiative, started in 2022, is a dedicated vendor-neutral
workstream within the Gateway API, working on examining how best to fit the
service mesh into the framework of the Gateway API resources, without
requiring users of the Gateway API to relearn everything they understand about
the Gateway API itself.

Over the last year, GAMMA has dug deeply into the challenges and possible
solutions around using the Gateway API for service mesh. The end result is a
small number of [enhancement
proposals](https://gateway-api.sigs.k8s.io/v1beta1/contributing/gep/) that
subsume many hours of thought and debate, and provide a minimum viable path to
allow the Gateway API can be used for service mesh. Of particular note:

- [GEP-1324](https://gateway-api.sigs.k8s.io/geps/gep-1324/) provides an
  overview of the GAMMA goals and some important definitions. This GEP is well
  worth a read for its discussion of the problem space.
- [GEP-1426](https://gateway-api.sigs.k8s.io/geps/gep-1426/) defines how to
  use Gateway API route resources, such as HTTPRoute, to manage traffic within
  a service mesh.
- [GEP-1686](https://gateway-api.sigs.k8s.io/geps/gep-1686/) builds on the
  work of [GEP-1709](https://gateway-api.sigs.k8s.io/geps/gep-1709/) to define
  a _conformance profile_ for service meshes to be declared conformant with
  the Gateway API.

### How will mesh routing work when using the Gateway API?

All the details are in
[GEP-1426](https://gateway-api.sigs.k8s.io/geps/gep-1426/), but the short
version for Gateway API 0.8.0 is that an HTTPRoute can now have a `parentRef`
that is a Service, rather than just a Gateway. We anticipate future GEPs in
this area as we gain more experience with service mesh use cases -- binding to
a Service makes it possible to use the Gateway API with a service mesh, but
there are several interesting use cases that remain difficult to cover.

### How does Gateway API conformance work for a service mesh?

One of the challenges that the GAMMA initiative ran into is that Gateway API
conformance was strongly tied to the idea that a given implementation provides
an ingress controller. Many service meshes don't, and requiring a
GAMMA-conformant mesh to also implement an ingress controller seemed
impractical at best. This resulted in work restarting on Gateway API
_conformance profiles_, as discussed in
[GEP-1709](https://gateway-api.sigs.k8s.io/geps/gep-1709/).

The basic idea of conformance profiles is that we can define subsets of the
Gateway API, and allow implementations to choose - and document! - which
subsets they conform to. GAMMA is adding a new profile, named `Mesh`, which
checks only the mesh functionality as defined by GAMMA; SIG-Network-Policy is
going to be using this concept as well.

## What else is in Gateway API 0.8.0?

In addition to GAMMA becoming experimental, Gateway API 0.8.0 also includes a
way to configure routing scope for a given Gateway (see
[GEP-1651](https://gateway-api.sigs.k8s.io/geps/gep-1651/) and updates the
HTTPRoute resource to include native configuration for timeouts (see
[GEP-1742](https://gateway-api.sigs.k8s.io/geps/gep-1742/).

Additionally, we have (by necessity) taken a hard look at some
[Experimental][status] GEPs which have been lingering long enough that
projects have come to rely on them in production use. This is a bit of a
breakdown of the GEP process; in order to prevent it happening in the future,
we have changed the GEP process such that reaching [Experimental][status]
_requires_ that a GEP include both the graduation criteria by which the GEP
will become [Standard][status], and a probationary period after which the GEP
will be dropped if does not meet its graduation criteria.

For an exhaustive list of changes included in the `v0.8.0` release, please see
the [v0.8.0 release
notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.0).
For more information on Gateway API versioning, refer to the [official
documentation](https://gateway-api.sigs.k8s.io/concepts/versioning/).

## How can I get started with the Gateway API?

At this point, the Gateway API is supported by a number of
[implementations][impl] of both ingress controllers and service meshes. To get
started, take a look at the [API concepts documentation][concepts] and check
out some of the [Guides][guides] to learn about the Gateway API, or check out
the [implementations page][impl] and select an implementation that you're
familiar with to try it out. Gateway API is a [Custom Resource Definition
(CRD)][crd] based API so you'll need to [install the CRDs][install-crds] onto
a cluster to use the API.

If you're specifically interested in helping to contribute to Gateway API, we
would love to have you! Please feel free to [open a new issue][issue] on the
repository, or join in the [discussions][disc]. Also check out the [community
page][community] which includes links to the Slack channel and community
meetings. We look forward to seeing you!!

[status]:https://gateway-api.sigs.k8s.io/geps/overview/#status
[ch]:https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard
[crd]:https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/
[concepts]:https://gateway-api.sigs.k8s.io/concepts/api-overview/
[guides]:https://gateway-api.sigs.k8s.io/guides/getting-started/
[impl]:https://gateway-api.sigs.k8s.io/implementations/
[install-crds]:https://gateway-api.sigs.k8s.io/guides/getting-started/#install-the-crds
[issue]:https://github.com/kubernetes-sigs/gateway-api/issues/new/choose
[disc]:https://github.com/kubernetes-sigs/gateway-api/discussions
[community]:https://gateway-api.sigs.k8s.io/contributing/community/

