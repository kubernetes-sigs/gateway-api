# Implementer's Guide

Everything you wanted to know about building a Gateway API implementation
but were too afraid to ask.

This document is a place to collect tips and tricks for _writing a Gateway API
implementation_ that have no straightforward place within the godoc fields of the
underlying types.

It's also intended to be a place to write down some guidelines to
help implementers of this API to skip making common mistakes.

It may not be very relevant if you are intending to _use_ this API as an end
user as opposed to _building_ something that uses it.

This is a living document, if you see something missing, PRs welcomed!

## Important things to remember about Gateway API

Hopefully most of these are not surprising, but they sometimes have non-obvious
implications that we'll try and lay out here.

### Gateway API is a `kubernetes.io` API

Gateway API uses the `gateway.networking.k8s.io` API group. This means that,
like APIs delivered in the core Kubernetes binaries, each time a release happens,
the APIs have been reviewed by upstream Kubernetes reviewers, just like the APIs
delivered in the core binaries.

### Gateway API is delivered using CRDs

Gateway API is supplied as a set of CRDs, version controlled using our [versioning
policy][versioning].

The most important part of that versioning policy is that what _appears to be_
the same object (that is, it has the same `group`,`version`, and `kind`) may have
a slightly different schema. We make changes in ways that are _compatible_, so
things should generally "just work", but there are some actions implementations
need to take to make "just work"ing more reliable; these are detailed below.

The CRD-based delivery also means that if an implementation tries to use (that is
get, list, watch, etc) Gateway API objects when the CRDs have _not_ been installed,
then it's likely that your Kubernetes client code will return serious errors.
Tips to deal with this are also detailed below.

The CRD definitions for Gateway API objects all contain two specific
annotations:

- `gateway.networking.k8s.io/bundle-version: <semver-release-version>`
- `gateway.networking.k8s.io/channel: <channel-name>`

The concepts of "bundle version" and "channel" (short for "release channel") are
explained in our [versioning][versioning] documentation.

Implementations may use these to determine what schema versions are installed in
the cluster, if any.

[versioning]: /concepts/versioning

### Changes to the Standard Channel CRDs are backwards compatible

Part of the contract for Standard Channel CRDs is that changes _within an API
version_ must be _compatible_. Note that CRDs that are part of Experimental
Channel do not provide any backwards compatibility guarantees.

Although the [Gateway API versioning policy](/concepts/versioning) largely
aligns with upstream Kubernetes APIs, it does allow for "corrections to
validation". For example, if the API spec stated that a value was invalid but
the corresponding validation did not cover that, it's possible that a future
release may add validation to prevent that invalid input.

This contract also means that an implementation will not fail with a higher
version of the API than the version it was written with, because the newer
schema being stored by Kubernetes will definitely be able to be serialized into
the older version used in code by the implementation.

Similarly, if an implementation was written with a _higher_ version, the newer
values that it understands will simply _never be used_, as they are not present
in the older version.

## Implementation Rules and Guidelines

### CRD Management

For information on how to manage Gateway API CRDs, including when it is
acceptable to bundle CRD installation with your implementation, refer to our
[CRD Management Guide](/guides/crd-management).

### Conformance and Version Compatibility

A conformant Gateway API implementation is one that passes the conformance tests
that are included in each Gateway API bundle version release.

An implementation MUST pass the conformance suite with _no_ skipped tests to be
conformant. Tests may be skipped during development, but a version you want to
be conformant MUST have no skipped tests.

Extended features may, as per the contract for Extended status, be disabled.

Gateway API conformance is version-specific. An implementation that passes
conformance for version N may not pass conformance for version N+1 without changes.

Implementations SHOULD submit a report from the conformance testing suite back
to the Gateway API Github repo containing details of their testing.

The conformance suite output includes the Gateway API version supported.

#### Version Compatibility

Once v1.0 is released, for implementations supporting Gateway and GatewayClass,
they MUST set a new Condition, `SupportedVersion`, with `status: true` meaning
that the installed CRD version is supported, and `status: false` meaning that it
is not.

### Standard Status Fields and Conditions

Gateway API has many resources, but when designing this, we've worked to keep
the status experience as consistent as possible across objects, using the
Condition type and the `status.conditions` field.

Most resources have a `status.conditions` field, but some also have a namespaced
field that _contains_ a `conditions` field.

For the latter, Gateway's `status.listeners` and the Route `status.parents`
fields are examples where each item in the slice identifies the Conditions
associated with some subset of configuration.

For the Gateway case, it's to allow Conditions per _Listener_, and in the Route
case, it's to allow Conditions per _implementation_ (since Route objects can
be used in multiple Gateways, and those Gateways can be reconciled by different
implementations).

In all of these cases, there are some relatively-common Condition types that have
similar meanings:

- `Accepted` - the resource or part thereof contains acceptable config that will
produce some configuration in the underlying data plane that the implementation
controls. This does not mean that the _whole_ configuration is valid, just that
_enough_ is valid to produce some effect.
- `Programmed` - this represents a later phase of operation, after `Accepted`,
when the resource or part thereof has been Accepted and programmed into the
underlying dataplane. Users should expect the configuration to be ready for
traffic to flow _at some point in the near future_. This Condition does _not_
say that the dataplane is ready _when it's set_, just that everything is valid
and it _will become ready soon_. "Soon" may have different meanings depending
on the implementation.
- `ResolvedRefs` - this Condition indicates that all references in the resource
or part thereof were valid and pointed to an object that both exists and allows
that reference. If this Condition is set to `status: false`, then _at least one_
reference in the resource or part thereof is invalid for some reason, and the
`message` field should indicate which one are invalid.

Implementers should check the godoc for each type to see the exact details of
these Conditions on each resource or part thereof.

Additionally, the upstream `Conditions` struct contains an optional
`observedGeneration` field - implementations MUST use this field and set it to
the `metadata.generation` field of the object at the time the status is generated.
This allows users of the API to determine if the status is relevant to the current
version of the object.

## TLS

TLS is a large topic in Gateway API, with the set of capabilities continuing to
expand. There is a [TLS guide](/guides/tls) that covers this topic in more depth
from a user-facing perspective, but this section attempts to fill in some gaps
from an implementer's perspective.

### Listener Isolation
Within Gateways, TLS config is currently tied exclusively to Listeners. To make
that approach manageable, we're encouraging all implementations to work towards
the goal of providing full "Listener Isolation" as defined below:

Requests SHOULD match at most one Listener. For example, if Listeners are
defined for "foo.example.com" and "*.example.com", a request to
"foo.example.com" SHOULD only be routed using routes attached to the
"foo.example.com" Listener (and not the "*.example.com" Listener).

Implementations that do not support Listener Isolation MUST clearly document
this. In the future, we plan to add HTTPS Listener Isolation conformance tests
to ensure this behavior is consistent across implementations that claim support
for the feature. See
[#2803](https://github.com/kubernetes-sigs/gateway-api/issues/2803) for the
latest updates on these tests.

### Indirect Configuration
There are a variety of instances where TLS Certificates may not be managed
directly by the owner of a Gateway. Although this is not meant to be an
exhaustive list, it documents some of the approaches we expect to see used to
manage TLS Certificates with Gateway API:

#### 1. Certificates from other places
Some providers offer the ability to configure and host TLS Certificates outside
of Kubernetes altogether. Implementations that can connect to those external
providers may expose that capability via a TLS option on the Gateway Listener,
for example:

```
  listeners:
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      options:
        vendor.example.com/certificate-name: store-example-com
```

In this example, the `store-example-com` name would refer to the name of a
certificate stored by the external `vendor.example.com` TLS Certificate
provider.

#### 2. Automatically generated TLS certs that are populated later
Many users would prefer that TLS certs will be automatically generated on their
behalf. One potential implementation of that would involve a controller that
watches Gateways and HTTPRoutes, generates TLS certs, and attaches them to the
Gateway. Depending on the implementation details, Gateway owners may need to
configure something at the Gateway or Listener level to explicitly opt-in to
this feature. For example, let's say someone created `acme-cert-generator` to
generate certs following this pattern. That generator may choose to only
generate and populate certs on Gateway Listeners with `acme.io/cert-generator`
set in `tls.options` or a similar annotation set for the entire Gateway.

Note that this is actually fairly similar to [how Cert Manager works
today](https://cert-manager.io/docs/usage/gateway/), but that requires Gateway
owners to reference a Kubernetes Secret that it will then populate. This
specific approach was required because TLS CertificateRefs were required to be
specified until Gateway API v1.1.

With the relaxing of Gateway API validation in v1.1, TLS Certificates can be
left unspecified on creation, allowing for less configuration when working with
generated TLS certificates.

#### 3. Certs that are specified by other personas
In some organizations, Application Developers are responsible for managing TLS
Certificates (see [Roles and Personas](/concepts/roles-and-personas) for more on
this and other roles).

To enable this use case, a new controller and CRD would be created. This
CRD would link hostnames to user-provided certs, and then the controller would
populate the certs specified by that CRD on Gateway listeners that matched those
hostnames. This would also likely benefit from a Listener or Gateway-level
opt-in for the behavior.

### Overall Guidelines for TLS Extensions
When building TLS extensions on top of Gateway API, it's important to follow the
following guidelines:

1. Use domain-prefixed names that are unique to your implementation for any TLS
   Options or Annotations. (For example, use `example.com/certificate-name`
   instead of just `certificate-name`).
2. Do not encode sensitive information like certificates in the option or
   annotation value. Instead favor references via concise names that are easy
   to understand. Although these values can technically be as long as 253
   characters, we strongly recommend keeping the values below 50 characters to
   maintain overall readability and UX.
3. To enable these extensions, Gateway API v1.1+ will no longer require TLS
   Config to be specified on Gateway Listeners. When a Gateway Listener does not
   have sufficient TLS configuration specified, implementations MUST set the
   `Programmed` condition to `False` for that Listener with the
   `InvalidTLSConfig` reason.
4. Regardless of any extensions you may choose to support, it is important to
   support the core TLS configuration that is intended to be portable across all
   implementations. Extensions have their place in this API, but all
   implementations MUST still support the core capabilities of the API.

## Resource Details

For each currently available conformance profile, there are a set of resources
that implementations are expected to reconcile.

The following section goes through each Gateway API object and indicates expected
behaviors.

### GatewayClass

GatewayClass has one main `spec` field - `controllerName`. Each implementation
is expected to claim a domain-prefixed string value (like
`example.com/example-ingress`) as its `controllerName`.

Implementations MUST watch _all_ GatewayClasses, and reconcile GatewayClasses
that have a matching `controllerName`. The implementation must choose at least
one compatible GatewayClass out of the set of GatewayClasses that have a matching
`controllerName`, and indicate that it accepts processing of that GatewayClass
by setting an `Accepted` Condition to `status: true` in each. Any GatewayClasses
that have a matching `controllerName` but are _not_ Accepted must have the
`Accepted` Condition set to `status: false`.

Implementations MAY choose only one GatewayClass out of the pool of otherwise
acceptable GatewayClasses if they can only reconcile one, or, if they are capable
of reconciling multiple GatewayClasses, they may also choose as many as they like.

If something in the GatewayClass renders it incompatible (at the time of writing,
the only possible reason for this is that there is a pointer to a `paramsRef`
object that is not supported by the implementation), then the implementation
SHOULD mark the incompatible GatewayClass as not `Accepted`.

### Gateway

Gateway objects MUST refer in the `spec.gatewayClassName` field to a GatewayClass
that exists and is `Accepted` by an implementation for that implementation to
reconcile them.

Gateway objects that fall out of scope (for example, because the GatewayClass
they reference was deleted) for reconciliation MAY have their status removed by
the implementation as part of the delete process, but this is not required.

### Routes

All Route objects share some properties:

- They MUST be attached to an in-scope parent for the implementation to consider
  them reconcilable.
- The implementation MUST update the status for each in-scope Route with the
  relevant Conditions, using the namespaced `parents` field. See the specific
  Route types for details, but this usually includes `Accepted`, `Programmed`
  and `ResolvedRefs` Conditions.
- Routes that fall out of scope SHOULD NOT have status updated, since it's
  possible that these updates may overwrite any new owners. The
  `observedGeneration` field will indicate that any remaining status is out of
  date.

#### HTTPRoute

HTTPRoutes route HTTP traffic that is _unencrypted_ and available for inspection.
This includes HTTPS traffic that's terminated at the Gateway (since that is then
decrypted), and allows the HTTPRoute to use HTTP properties, like path, method,
or headers in its routing directives.

#### TLSRoute

TLSRoutes route encrypted TLS traffic using the SNI header, _without decrypting
the traffic stream_, to the relevant backends.

#### TCPRoute

TCPRoutes route a TCP stream that arrives at a Listener to one of the given
backends.

#### UDPRoute

UDPRoutes route UDP packets that arrive at a Listener to one of the given
backends.

### ReferenceGrant

ReferenceGrant is a special resource that is used by resource owners in one
namespace to _selectively_ allow references from Gateway API objects in other
namespaces.

A ReferenceGrant is created in the same namespace as the thing it's granting
reference access to, and allows access from other namespaces, from other Kinds,
or both.

Implementations that support cross-namespace references MUST watch ReferenceGrant
and reconcile any ReferenceGrant that points to an object that's referred to by
an in-scope Gateway API object.
