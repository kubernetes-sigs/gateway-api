# Gateway API Implementer's Guide

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

### Changes to the Gateway API CRDs are backwards compatible

Part of the contract for Gateway API CRDs is that changes _within an API version_
must be _compatible_.

"Within an API Version" means changes to a CRD that occur while the same API version
(`v1alpha2` or `v1` for example) is in use, and "compatible" means that any new
fields, values, or validation will be added to ensure that _previous_
objects _will still be valid objects_ after the change.

This means that once Gateway API objects move to the `v1` API version, then _all_
changes must be compatible.

This contract also means that an implementation will not fail with a higher version
of the API than the version it was written with, because the newer schema being
stored by Kubernetes will definitely be able to be serialized into the older version
used in code by the implementation.

Similarly, if an implementation was written with a _higher_ version, the newer
values that it understands will simply _never be used_, as they are not present
in the older version.

## Implementation Rules and Guidelines

### CRD Management

For a Gateway API implementation to work, the Gateway API CRDs must be installed
in the Kubernetes cluster the implementation is watching.

Implementations have two options: automatically installing CRDs or requiring
installation before working. Both have tradeoffs.

Either way has certain things that SHOULD be true, however:

Whatever method is used, cluster admins SHOULD attempt to ensure that
the Bundle version of the CRDs is not _downgraded_. Although we ensure that
API changes are backwards compatible, changing CRD definitions can change the
storage version of the resource, which could have unforseen effects. Most of the
time, things will probably work, but if it doesn't work, it will most likely 
break in weird ways.

Additionally, older versions of the API may be missing fields or features, which
could be very disruptive for users.

Try your best to ensure that the bundle version doesn't roll backwards. It's safer.

Implementations SHOULD also handle the Gateway API CRDs _not_ being present in
the cluster without crashing or panicking. Exiting with a clear fatal error is
acceptable in this case, as is disabling Gateway API support even if enabled in
configuration.

Practically, for implementations using tools like `controller-runtime` or
similar tooling, they may need to check for the _presence_ of the CRDs by 
getting the list of installed CRDs before attempting to watch those resources.
(Note that this will require the implementation to have `read` access to those
resources though.)

#### Automatic CRD installation

Automatic CRD installation also includes automatic installation mechanisms such
as Helm, if the CRDs are included in a Helm chart with the implementation's
installation.

CRD definitions MAY be installed automatically by implementations, and if they do,
they MUST have a way to ensure:

- there are no other Gateway API CRDs installed in the cluster before starting, or
- that the CRD definitions are only installed if they are a higher bundle version
  than any existing Gateway API CRDs. Note that even this may not be safe if there
  are breaking changes in the experimental channel resources, so implementations
  should be _very_ careful with doing this.

This avoids problems if another implementation is also installed in the cluster
and expects a higher version of the CRDs to be installed.

If the implementation can guarantee that no other implementation will interact
with the cluster, then it MAY automatically install a relevant version of the CRDs.

The ideal method for an automatic installation would require the implementation
to:

- Check if there are any Gateway API CRDs installed in the cluster.
- If not, install its most compatible version of the CRDs.
- If so, only install its version of the CRDs if the bundle version is higher
  than the existing one, and the mechanism may also need to check if there are
  incompatible changes included in any versions as well.


Because of our backwards compatibility guarantees, it's also safe for a controller
to flip the install channel between "standard" and "experimental", although
implementations MUST NOT do this without requiring confirmation from the cluster
administrator.

Automatic CRD installation has the advantage that there is less for the
implementation user to do; any required version checking can be performed by
code instead of by the cluster admin.

It should also be noted that many infra and cluster admins do this sort of
automatic CRD management using external management systems that will not be
visible to a Gateway implementation, so automatic installation MUST be able to
be disabled by the installation owner (whether that is the infra or cluster
admin).

Because of all these caveats, we DO NOT recommend doing automatic CRD management
at this time.

#### Manual CRD installation

Manual CRD installation has the advantage that the implementer needs to maintain
less code; however, it pushes the responsibility for correctly managing the
Gateway API CRDs back to the cluster admin, who may not have as much context
as is provided here.

Implementations MAY require the installation to be done manually; if so, the
installation instructions SHOULD include commands to check if there are any other
CRDs installed already and verify that the installation will not be a downgrade.

### Conformance and Version compatibility

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

#### Version compatibility

Once v1.0 is released, for implementations supporting Gateway and GatewayClass,
they MUST set a new Condition, `SupportedVersion`, with `status: true` meaning
that the installed CRD version is supported, and `status: false` meaning that it
is not.

### Standard Status fields and Conditions

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


### Resources details

For each currently available conformance profile, there are a set of resources
that implementations are expected to reconcile.

The following section goes through each Gateway API object and indicates expected
behaviors.

#### GatewayClass

GatewayClass has one main `spec` field - `controllerName`. Each implementation
is expected to claim a domain-prefixed string value (like
`example.com/example-ingress`) as its `controllerName`.

Implementations MUST watch _all_ GatewayClasses, and reconcile GatewayClasses
that have a matching `controllerName`. The implementation must choose at least
one compatible GatewayClass out of the set of GatewayClasses that have a matching
`controllerName`, and indicate that it accepts processing of that GatewayClass
by setting an `Accepted` Condition to `status: true` in each. Any GatewayClasses
that have a matching `controllerName` but are _not_ Accepted must have the
`Accepted` Condition sett to `status: false`.

Implementations MAY choose only one GatewayClass out of the pool of otherwise
acceptable GatewayClasses if they can only reconcile one, or, if they are capable
of reconciling multiple GatewayClasses, they may also choose as many as they like.

If something in the GatewayClass renders it incompatibie (at the time of writing,
the only possible reason for this is that there is a pointer to a `paramsRef`
object that is not supported by the implementation), then the implementation
SHOULD mark the incompatible GatewayClass as not `Accepted`.

#### Gateway

Gateway objects MUST refer in the `spec.gatewayClassName` field to a GatewayClass
that exists and is `Accepted` by an implementation for that implementation to
reconcile them.

Gateway objects that fall out of scope (for example, because the GatewayClass
they reference was deleted) for reconciliation MAY have their status removed by
the implementation as part of the delete process, but this is not required.

#### General Route information

All Route objects share some properties:

- They MUST be attached to an in-scope parent for the implementation to consider
them reconcilable.
- The implementation MUST update the status for each in-scope Route with the
relevant Conditions, using the namespaced `parents` field. See the specific Route
types for details, but this usually includes `Accepted`, `Programmed` and
`ResovledRefs` Conditions.
- Routes that fall out of scope SHOULD NOT have status updated, since it's possible
that these updates may overwrite any new owners. The `observedGeneration` field
will indicate that any remaining status is out of date.


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

#### ReferenceGrant

ReferenceGrant is a special resource that is used by resource owners in one
namespace to _selectively_ allow references from Gateway API objects in other
namespaces.

A ReferenceGrant is created in the same namespace as the thing it's granting
reference access to, and allows access from other namespaces, from other Kinds,
or both.

Implementations that support cross-namespace references MUST watch ReferenceGrant
and reconcile any ReferenceGrant that points to an object that's referred to by
an in-scope Gateway API object.
