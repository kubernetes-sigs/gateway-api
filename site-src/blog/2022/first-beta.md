---
description: >
  We're pleased to announce the start of beta releases for the Gateway API
  project with the release of version `v0.5.0`. This release includes `v1beta1`
  versions of `GatewayClass`, `Gateway` and `HTTPRoute` as well as several other
  significant improvements which this post will cover some of the highlights of.
---

# Introducing Gateway API v1beta1

<small style="position:relative; top:-30px;">
  :octicons-calendar-24: June 24, 2022 ·
  :octicons-clock-24: 5 min read
</small>

We're pleased to announce that key Gateway API resources are graduating to
beta with the `v0.5.0` release of the API. In addition to `GatewayClass`, `Gateway` 
and `HTTPRoute` graduating to beta, this release introduces several new concepts
to the API that we'll cover below.

## What is Gateway API

Gateway API is a collection of resources centered around `Gateway` resources
(which represent the underlying network gateways / proxy servers) to enable
robust Kubernetes service networking through expressive, extensible and
role-oriented interfaces that are implemented by many vendors and have broad
industry support.

Originally conceived as a successor to the well known `Ingress` API, the
benefits of Gateway API include (but are not limited to) explicit support for
many commonly used networking protocols (e.g. `HTTP`, `TLS`, `TCP`, `UDP`) as
well as tightly integrated support for Transport Layer Security (TLS). The
`Gateway` resource in particular enables implementations to manage the lifecycle
of network gateways as a native Kubernetes API.

If you're an end-user interested in some of the benefits of Gateway API we
invite you to jump in and find an implementation that suits you. At the time of
this release there are over a dozen [implementations][impl] for popular API
gateways and service meshes and guides are available to start exploring quickly.

[impl]:https://gateway-api.sigs.k8s.io/implementations

### Getting Started

The Gateway API project itself is a Kubernetes API like `Ingress` and you can
think of the implementations of it as similar to the implementations for
`IngressClass` or `StorageClass` in that Kubernetes itself has no default
implementation of it (though your cluster provider might).

Take a look at our [API concepts documentation][concepts] and check out some of
our [Guides][guides] to start familiarizing yourself with the APIs and how they
work. When you're ready for a practical application open our [implementations
page][impl] and select an implementation that belongs to an existing technology
you may already be familiar with or the one your cluster provider uses as a
default (if applicable). Gateway API is a [Custom Resource Definition
(CRD)][crd] based API so you'll need to [install our CRDs][install-crds] onto a
cluster to use the API.

If you're specifically interested in helping to contribute to Gateway API, we
would love to have you! Please feel free to [open a new issue][issue] on the
repository, or join in our [discussions][disc]. Also check out our [community
page][community] which includes links to our chatrooms and community meetings.

[crd]:https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/
[concepts]:https://gateway-api.sigs.k8s.io/concepts/api-overview/
[guides]:https://gateway-api.sigs.k8s.io/v1alpha2/guides/getting-started/
[impl]:https://gateway-api.sigs.k8s.io/implementations
[install-crds]:https://gateway-api.sigs.k8s.io/v1alpha2/guides/getting-started/#install-the-crds
[issue]:https://github.com/kubernetes-sigs/gateway-api/issues/new/choose
[disc]:https://github.com/kubernetes-sigs/gateway-api/discussions/new
[community]:https://gateway-api.sigs.k8s.io/contributing/community/

## Release Highlights

### Graduation to Beta

The `v0.5.0` release is particularly historic because it marks the growth in
maturity to a `v1beta1` release for some of our key APIs:

- `GatewayClass`
- `Gateway`
- `HTTPRoute`

This achievement was marked by the completion of several graduation criteria:

- over a dozen [implementations][impl]
- the first release of our conformance test suite
- most of the API surface is actively being used
- approval from Kubernetes SIG Network API reviewers

To see what's in store for future releases check out the [next steps
section](/first-beta.md#next-steps).

[impl]:https://gateway-api.sigs.k8s.io/implementations

### Release Channels

This release introduces the `experimental` and `standard` [release channels][ch]
which enable a better balance of maintaining stability while still enabling
experimentation and iterative development.

The `standard` release channel includes:

- resources that have graduated to beta
- fields that have graduated to standard (no longer considered experimental)

The `experimental` release channel includes everything in the `standard` release
channel, plus:

- `alpha` API resources
- un-graduated fields that are considered experimental

Release channels are used internally to enable iterative development with
quick turnaround, and externally to indicate feature stability to implementors
and end-users.

For this release we've added the following experimental features:

- [Route select Gateway listeners by port number](https://github.com/kubernetes-sigs/gateway-api/pull/1002)
- [URL rewrites and path redirects](https://github.com/kubernetes-sigs/gateway-api/pull/945)

[ch]:https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard

## Other Improvements

For an exhaustive list of changes included in the `v0.5.0` release, please see
the [v0.5.0 CHANGELOG entry][changelog].

[changelog]:https://github.com/kubernetes-sigs/gateway-api/blob/master/CHANGELOG.md#v050

## Next Steps

As we continue to mature the API for production use cases, the following is a
(non-exhaustive) list of some of the highlights for upcoming releases:

- [GRPCRoute][gep1016] for [gRPC][grpc] traffic routing
- [route delegation][pr1085]
- Layer 4 API maturity: `v1beta1` for `TCPRoute`, `UDPRoute` and `TLSRoute`
- Expanded support and coverage for service mesh use cases (to be discussed in Working Group Mesh)

If there's something on this list you want to get involved in, or there's
something not on this list that you want to advocate for to get on the roadmap
please get in touch via the [community][community]!

[gep1016]:https://github.com/kubernetes-sigs/gateway-api/blob/master/site-src/geps/gep-1016.md
[grpc]:https://grpc.io/
[pr1085]:https://github.com/kubernetes-sigs/gateway-api/pull/1085
[community]:https://gateway-api.sigs.k8s.io/contributing/community/
