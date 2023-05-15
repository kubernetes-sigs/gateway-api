# Changelog

## Table of Contents

- [v0.7.0](#v070)
- [v0.7.0-rc2](#v070-rc2)
- [v0.7.0-rc1](#v070-rc1)
- [v0.6.2](#v062)
- [v0.6.1](#v061)
- [v0.6.0](#v060)
- [v0.6.0-rc2](#v060-rc2)
- [v0.6.0-rc1](#v060-rc1)
- [v0.5.1](#v051)
- [v0.5.0](#v050)
- [v0.5.0-rc2](#v050-rc2)
- [v0.5.0-rc1](#v050-rc1)
- [v0.4.3](#v043)
- [v0.4.2](#v042)
- [v0.4.1](#v041)
- [v0.4.0](#v040)
- [v0.4.0-rc2](#v040-rc2)
- [v0.4.0-rc1](#v040-rc1)
- [v0.3.0](#v030)
- [v0.2.0](#v020)
- [v0.1.0](#v010)
- [v0.1.0-rc2](#v010-rc2)
- [v0.1.0-rc1](#v010-rc1)

# v0.7.0

The v0.7.0 release focuses on refining and stabilizing existing APIs. This
included a focus on both conformance tests and clarifying ambiguous parts of the
API spec.

## Features Graduating to Standard
In addition to those broad focuses, 2 features are graduating to the
standard channel:

* GEP-1323: Response Header Modifiers (#1905, @robscott)
* GEP-726: Path Redirects and Rewrites (#1905, @robscott)

## GEPs
There are a lot of interesting GEPs in the pipeline right now, but only some of
these GEPs have made it to experimental status in time for v0.7.0. The GEPs
highlighted below are both in an experimental state and are either entirely new
(GEP-1748) or had significant new concepts introduced (GEP-713):

### GEP-713: Policy Attachment
This GEP received a major update, splitting policy attachment into two
categories "Direct" and "Inherited". The new "Direct" mode enables a simplified
form of policy attachment for targeting a single resource (#1565, @youngnick).

### GEP-1748: Gateway API Interaction with Multi-Cluster Services
A new GEP was introduced to define how Gateway API interacts with Multi-Cluster
Services. At a high level, this states that ServiceImports have "Extended"
support and can be used anywhere Services can throughout the API. There's a lot
more nuance here, so for the full details, refer to the GEP. (#1843, @robscott)

## Other Changes by Kind

### Status Changes

- The "Ready" Gateway and Listener condition has been reserved for future use.
  (#1888, @howardjohn)
- The UnsupportedAddress Listener condition reason has been moved to a Gateway
  condition reason.  (#1888, @howardjohn)
- The AddressNotAssigned Gateway condition reasons has moved from Accepted to
  Programmed. (#1888, @howardjohn)
- The NoResources Gateway condition reasons has moved from Ready to Programmed.
  (#1888, @howardjohn)

### Spec Cleanup

- Clarification that port redirects should not add port number to Location
  header for HTTP and HTTPS requests on 80 and 443. (#1908, @robscott)
- Port redirect when empty will depend on the configured Redirect scheme (#1880,
  @gauravkghildiyal)
- Updated spec to clarify that Exact matches have precedence over Prefix matches
  and RegularExpression matches have implementation specific precedence. (#1855,
  @Xunzhuo)
- The `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer is no
  longer required and is now just recommended. (#1917, @howardjohn)

### Validation Fixes

- Removes GRPCRoute method match defaulting to allow for matching all requests,
  or matching only by header. (#1753, @skriss)
- Update route validation to comply with RFC-3986 "p-char" characters. (#1644,
  @jackstine)
- Illegal names like " " will be not allowed for query param name in
  HTTPQueryParamMatch. (#1796, @gyohuangxin)
* Webhook: Port is now considered when validating that ParentRefs are unique
  (#1995, @howardjohn)

### Conformance

- No conformance tests run by default anymore, including tests for GatewayClass
  and Gateway. A new SupportGateway feature must be opted into in order to run
  those tests (similar to what we've done previously for ReferenceGrant and
  HTTPRoute). Also with this release, `EnableAllSupportedFeatures` enables all
  Gateway AND Mesh features (where previously that was just Gateway). (#1894,
  @shaneutt)
- Gateways must publish the "Programmed" condition. (#1732, @robscott)
- Add `all-features` flag to enable all supported feature conformance tests.
  (#1642, @gyohuangxin)
- A new SkipTests field has been added to the conformance test options to
  opt-out of specific tests. (#1578, @mlavacca)
- Added: conformance tests for http rewrite host and path filters. (#1622,
  @LiorLieberman)
- In Conformance tests, when a Route references a gateway having no listener
  whose allowedRoutes criteria permit the route, the reason
  NotAllowedByListeners should be used for the accepted condition. (#1669,
  @mlavacca)
- Support configurable timeout for GatewayObservedGenerationBump (#1887,
  @Xunzhuo)
- The conformance test HTTPRouteInvalidCrossNamespaceParentRef now requires the
  HTTPRoute accepted condition to be failing with the ParentRefNotPermitted
  reason. (#1694, @mlavacca)
- The conformance tests always check that the HTTPRoute ResolvedRefs condition
  is enforced, even when the status is true. (#1668, @mlavacca)
- Checks for the NotAllowedByListeners reason on the HTTPRoute's Accepted: false
  condition in the HTTPRouteInvalidCrossNamespaceParentRef conformance test.
  (#1714, @skriss)
- Added conformance test to verify that path matching precedence is
  implemented correctly. (#1855, @Xunzhuo)
- Remove a test that only covered redirect status without any other changes.
  (#2007, @robscott)
- Port redirect when empty will depend on the configured Redirect scheme (#1880,
  @gauravkghildiyal)
- Fixes for mesh conformance tests (#2017, @keithmattix)

### Documentation

- Updated outdated content on list of resources in installation guide page.
  (#1857, @randmonkey)
- Fix description of ReferenceGrant example in documentation by making it use
  the correct resources. (#1864, @matteoolivi)
- Fix grammar mistake in ReferenceGrant implementation guidelines. (#1865,
  @matteoolivi)

# v0.7.0-rc2

We expect this to be our final release candidate before launching v0.7.0. This
release candidate includes a variety of clarifications and conformance updates.
The changelog below represents the changes since v0.7.0-rc1.

## Changes by Kind

### Spec Clarification

- Port redirect when empty will depend on the configured Redirect scheme (#1880,
  @gauravkghildiyal)

### Conformance

- Remove a test that only covered redirect status without any other changes.
  (#2007, @robscott)
- Port redirect when empty will depend on the configured Redirect scheme (#1880,
  @gauravkghildiyal)

### Validation Fixes

* Webhook: Port is now considered when validating that ParentRefs are unique
  (#1995, @howardjohn)

# v0.7.0-rc1

## Changes by Kind

### Graduating to Standard

- GEP-1323: Response Header Modifier has graduated to standard (#1905,
  @robscott)
- GEP-726: Path Redirects and Rewrites has graduated to the standard channel.
  (#1874, @robscott)

### Experimental GEPs

- The Policy Attachment GEP received a major update, splitting policy attachment
  into two categories "Direct" and "Inherited". The new "Direct" mode enables a
  simplified form of policy attachment for targeting a single resource (#1565,
  @youngnick)
- A new GEP was introduced to define how Gateway API interacts with
  Multi-Cluster Services (#1843, @robscott)

### Status Changes

- The "Ready" Gateway and Listener condition has been reserved for future use.
  (#1888, @howardjohn)
- The UnsupportedAddress Listener condition reason has been moved to a Gateway
  condition reason.  (#1888, @howardjohn)
- The AddressNotAssigned Gateway condition reasons has moved from Accepted to
  Programmed. (#1888, @howardjohn)
- The NoResources Gateway condition reasons has moved from Ready to Programmed.
  (#1888, @howardjohn)

### Spec Cleanup

- Clarification that port redirects should not add port number to Location
  header for HTTP and HTTPS requests on 80 and 443. (#1908, @robscott)
- Updated spec to clarify that Exact matches have precedence over Prefix matches
  and RegularExpression matches have implementation specific precedence. (#1855,
  @Xunzhuo)
- The `gateway-exists-finalizer.gateway.networking.k8s.io` finalizer is no
  longer required and is now just recommended. (#1917, @howardjohn)

### Validation Fixes

- Removes GRPCRoute method match defaulting to allow for matching all requests,
  or matching only by header. (#1753, @skriss)
- Update route validation to comply with RFC-3986 "p-char" characters. (#1644,
  @jackstine)
- Illegal names like " " will be not allowed for query param name in
  HTTPQueryParamMatch. (#1796, @gyohuangxin)

### Conformance

- No conformance tests run by default anymore, including tests for GatewayClass
  and Gateway. A new SupportGateway feature must be opted into in order to run
  those tests (similar to what we've done previously for ReferenceGrant and
  HTTPRoute). Also with this release, `EnableAllSupportedFeatures` enables all
  Gateway AND Mesh features (where previously that was just Gateway). (#1894,
  @shaneutt)
- Gateways must publish the "Programmed" condition. (#1732, @robscott)
- Add `all-features` flag to enable all supported feature conformance tests.
  (#1642, @gyohuangxin)
- A new SkipTests field has been added to the conformance test options to
  opt-out of specific tests. (#1578, @mlavacca)
- Added: conformance tests for http rewrite host and path filters. (#1622,
  @LiorLieberman)
- In Conformance tests, when a Route references a gateway having no listener
  whose allowedRoutes criteria permit the route, the reason
  NotAllowedByListeners should be used for the accepted condition. (#1669,
  @mlavacca)
- Support configurable timeout for GatewayObservedGenerationBump (#1887,
  @Xunzhuo)
- The conformance test HTTPRouteInvalidCrossNamespaceParentRef now requires the
  HTTPRoute accepted condition to be failing with the ParentRefNotPermitted
  reason. (#1694, @mlavacca)
- The conformance tests always check that the HTTPRoute ResolvedRefs condition
  is enforced, even when the status is true. (#1668, @mlavacca)
- Checks for the NotAllowedByListeners reason on the HTTPRoute's Accepted: false
  condition in the HTTPRouteInvalidCrossNamespaceParentRef conformance test.
  (#1714, @skriss)
- Added conformance test to verify that path matching precedence is
  implemented correctly. (#1855, @Xunzhuo)

### Documentation

- Updated outdated content on list of resources in installation guide page.
  (#1857, @randmonkey)
- Fix description of ReferenceGrant example in documentation by making it use
  the correct resources. (#1864, @matteoolivi)
- Fix grammar mistake in ReferenceGrant implementation guidelines. (#1865,
  @matteoolivi)

# v0.6.2

This is a patch release that predominantly includes updated conformance tests
for implementations to implement.

For all major changes since the `v0.5.x` release series, please see the
[v0.6.0](/#v060) release notes.

## Maintenance

- As per [changes in upstream to container image registries] we replaced all
  usage of the k8s.gcr.io registry with registry.k8s.io.
  (#1736, @shaneutt)

[changes in upstream to container image registries]:https://github.com/kubernetes/k8s.io/issues/4738

## Bug Fixes

- Fix invalid HTTP redirect/rewrite examples.
  (#1787, @Xunzhuo)

## Conformance Test Updates

- The `HTTPRouteInvalidCrossNamespaceParentRef` conformance test now checks for
  the `NotAllowedByListeners` reason on the `HTTPRoute`'s `Accepted: false`
  condition to better indicate why the route was note accepted.
  (#1714, @skriss)
- A conformance test was added for `HTTPRoute` to cover the behavior of a
  non-matching `SectionName` similar to what was already present for
  `ListenerPort`.
  (#1719, @zaunist)
- Fixed an issue where tests may fail erroneously on the removal of resources
  that are already removed.
  (#1745, @mlavacca)
- Logging in conformance utilities related to resource's `ObservedGeneration`
  has been improved to emit the `ObservedGenerations that are found for the
  purpose of making it easier to debug test failures and be more verbose about
  the objects in question.
  (#1761, @briantkennedy)
  (#1763, @briantkennedy)
- Patch instead of update in some places in conformance tests to reduce noise
  in logs.
  (#1760, @michaelbeaumont)
- Added `AttachedRoutes` testing to conformance tests.
  (#1624, @ChaningHwang)
- The conformance tests always check that the HTTPRoute ResolvedRefs condition
  is enforced, even when the status is true.
  (#1668, @mlavacca)

# v0.6.1

This is a patch release that predominantly includes updated conformance tests
for implementations to implement.

For all major changes since the `v0.5.x` release series, please see the
[v0.6.0](/#v060) release notes.

## Bug Fixes

- Our regex for validating path characters was updated to accurately identify
  "p-chars" as per RFC-3986.
  (#1644, @jackstine)
- An erroneous "namespace" field was present in our webhook ClusterRoleBindings
  and has been removed.
  (#1684, @tao12345666333)

## New Features

- Conditions for Policies have been added to the Golang library, enabling
  Go-based implementations to re-use those for their downstream Policies.
  (#1682, @mmamczur)

## Conformance Test Updates

- Added conformance tests for checking Port, Scheme and Path to the extended and
  experimental features.
  (#1611, @LiorLieberman)
- Added conformance tests for HTTP rewrite
  (#1622, #1628, @LiorLieberman)
- Added more conformance tests for path matching to catch known edge cases.
  (#1627, @sunjayBhatia)
- Added some initial conformance tests for TLSRoute passthrough.
  (#1579, @candita)
- Added conformance tests that exercise NotAllowedByListeners reason.
  (#1669, @mlavacca)
- Loosen the Accepted check in GatewayClass observed generation tests to
  provide a more realistic test for implementations.
  (#1655, @arkodg)
- A "SkipTests" field has been added to accomodate implementations in
  running subsets of the tests as needed, this can be particularly helpful
  for new implementations that want to add conformance iteratively.
  (#1578, @mlavacca)
- Fixed a broken test for GRPCRoute that caused an erronous failure.
  (#1692, @arkodg)
- Added "all-features" flag to conformance test to enable all supported
  features on test runs.
  (#1642, @gyohuangxin)
- Fixed usage of `net/http` default client in conformance test suite
  (#1617, @howardjohn)
- Fixed missing reference to NoMatchingParent in godoc
  (#1671, @mlavacca)

# v0.6.0

## Major Changes

### ReferenceGrant moves to `v1beta1`, ReferencePolicy removed

With more implementations now supporting ReferenceGrant (and more conformance coverage of the resource), we've moved ReferenceGrant to `v1beta1` in this release. **Note** that moving to beta also moves the object to the Standard channel (it was Experimental previously).

We've also removed the already-deprecated ReferencePolicy resource, so please move over to the shiny new ReferenceGrant, which has all the same features.

- Promotes ReferenceGrant to the v1beta1 API and the standard release channel
  (#1455, @nathancoleman)
- ReferencePolicy has been removed from the API in favor of ReferenceGrant.
  (#1406, @robscott)

### Introduce GRPCRoute

The `GRPCRoute` resource has been introduced in order to simplify the routing of GRPC requests.
Its design is described in [GEP-1016](https://gateway-api.sigs.k8s.io/geps/gep-1016/).
As it is a new resource, it is introduced in the experimental channel.

Thanks to @gnossen for pushing this ahead.

- Introduce GRPCRoute resource. (#1115, @gnossen)

### Status updates

As described in [GEP-1364](https://gateway-api.sigs.k8s.io/geps/gep-1364/), status conditions have been updated within the Gateway resource to make it more consistent with the rest of the API. These changes, along with some other status changes, are detailed below.

Gateway:

* New `Accepted` and `Programmed` conditions introduced.
* `Scheduled` condition deprecated.
* Core Conditions now `Accepted` and `Programmed`.
* Moves to Extended: `Ready`.

Gateway Listener:

* New `Accepted` and `Programmed` conditions introduced.
* `Detached` condition deprecated.
* Core Conditions now `Accepted`, `Programmed`, `ResolvedRefs`, and `Conflicted`.
* Moves to Extended: `Ready`.

All Resources:

* The `Accepted` Condition now has a `Pending` reason, which is the default until
  the condition is updated by a controller.

Route resources:

* The `Accepted` Condition now has a `NoMatchingParent` reason, to be set on routes
  when no matching parent can be found.

The purpose of these changes is to make the status flows more consistent across objects, and to provide a clear pattern for new objects as we evolve the API.

> **Note**: This change will require updates for implementations to be able to pass conformance tests. Implementations may choose to publish both new and old conditions, or only new conditions.

- Adds `Accepted` and deprecates `Detached` Listener conditions and reasons (#1446, @mikemorris)
- Adds `Accepted` and deprecates `Scheduled` Gateway conditions and reasons (#1447, @mikemorris)
- Adds `Pending` reason for use with all `Accepted` conditions throughout the API (#1453, @youngnick)
- Adds `Programmed` Gateway and Listener conditions, moves `Ready` to extended
  conformance (#1499, @LCaparelli)
- Add `RouteReasonNoMatchingParent` reason for `Accepted` condition. (#1516, @pmalek)

## Other Changes by type

### Deprecations

- GatewayClass, Gateway, and HTTPRoute are now only supported with the v1beta1
  version of the API. The v1alpha2 API versions of these resources will be fully
  removed in a future release. Additionally, v1alpha2 is marked as deprecated
  everywhere. (#1348 and #1405, @robscott)

### API Changes

- A new field `responseHeaderModifier` is added to `.spec.rules.filters`, which
  allows for modification of HTTP response headers (#1373, @aryan9600)
- Display the Programmed condition instead of the Ready condition in the output
- HTTPRoute: Validating webhook now ensures that Exact and Prefix path match
  values can now only include valid path values per RFC-3986. (RegularExpression
  path matches are not affected by this change). (#1599, @robscott)
- `RegularExpression` type selectors have been clarified to all be
  `ImplementationSpecific` conformance. (#1604, @youngnick)

### Documentation

- Clarify that BackendObjectReference's Port field specifies a service port, not
  a target port, for Kubernetes Service backends. (#1332, @Miciah)
- HTTPRequestHeaderFilter and HTTPResponseHeaderFilter forbid configuring
  multiple actions for the same header. (#1497, @rainest)
- Changes "custom" conformance level to "implementation-specific" (#1436,
  @LCaparelli)
- Clarification that changes to ReferenceGrants MUST be reconciled (#1429,
  @robscott)

### Conformance Tests

- ExemptFeatures have been merged into SupportedFeatures providing implementations
  a uniform way to specify the features they support.
  (#1507, @robscott) (#1394, @gyohuangxin)
- To be conformant with the API, if there is no ReferenceGrant that grants a
  listener to reference a secret in another namespace, the
  ListenerConditionReason for the condition ResolvedRefs must be set to
  RefNotPermitted instead of InvalidCertificateRef. (#1305, @mlavacca)
- A new test has been added to cover HTTP Redirects (#1556, @LiorLieberman)
- Fix Gateway reference in HTTPRouteInvalidParentRefNotMatchingListenerPort
  (#1591, @sayboras)

### Build Changes

- We now provide a [multi-arch](https://www.docker.com/blog/multi-arch-images/)
  image including new support for `arm64` in addition to `amd64` for our
  validating webhook.
  (#627, @wilsonwu & @Xunzhuo)

### Developer Notes

- Deprecated `v1alpha2` Go types are now aliases to their `v1beta1` versions
  (#1390, @howardjohn)

# v0.6.0-rc2

We expect this to be our final release candidate before launching v0.6.0. This
release candidate includes a variety of cleanup and documentation updates. The
changelog below represents the changes since v0.6.0-rc1.

### Conformance Tests

- A new test has been added to cover HTTP Redirects (#1556, @LiorLieberman)
- Fix Gateway reference in HTTPRouteInvalidParentRefNotMatchingListenerPort
  (#1591, @sayboras)

### General Cleanup

- Display the Programmed condition instead of the Ready condition in the output
  of `kubectl get gateways`. (#1602, @skriss)
- GRPCRoute: Regex validation for Method and Service has been tightened to match
  GRPC spec. (#1599, @robscott)
- GRPCRoute: Webhook validation of GRPCRoute has been expanded to closely match
  HTTPRoute validation. (#1599, @robscott)
- HTTPRoute and Gateway: Gaps between webhook validation for v1alpha2 and
  v1beta1 have been closed. (#1599, @robscott)
- HTTPRoute: Validating webhook now ensures that Exact and Prefix path match
  values can now only include valid path values per RFC-3986. (RegularExpression
  path matches are not affected by this change). (#1599, @robscott)
- The Gateway default conditions list now includes the Programmed condition.
  (#1604, @youngnick)
- `RegularExpression` type selectors have been clarified to all be
  `ImplementationSpecific` conformance. (#1604, @youngnick)

# v0.6.0-rc1

## Major Changes

### ReferenceGrant moves to `v1beta1`, ReferencePolicy removed

With more implementations now supporting ReferenceGrant (and more conformance coverage of the resource), we've moved ReferenceGrant to `v1beta1` in this release. **Note** that moving to beta also moves the object to the Standard channel (it was Experimental previously).

We've also removed the already-deprecated ReferencePolicy resource, so please move over to the shiny new ReferenceGrant, which has all the same features.

- Promotes ReferenceGrant to the v1beta1 API and the standard release channel
  (#1455, @nathancoleman)
- ReferencePolicy has been removed from the API in favor of ReferenceGrant.
  (#1406, @robscott)

### Introduce GRPCRoute

The `GRPCRoute` resource has been introduced in order to simplify the routing of GRPC requests.
Its design is described in [GEP-1016](https://gateway-api.sigs.k8s.io/geps/gep-1016/).
As it is a new resource, it is introduced in the experimental channel.

Thanks to @gnossen for pushing this ahead.

- Introduce GRPCRoute resource. (#1115, @gnossen)

### Status updates

As described in [GEP-1364](https://gateway-api.sigs.k8s.io/geps/gep-1364/), status conditions have been updated within the Gateway resource to make it more consistent with the rest of the API. These changes, along with some other status changes, are detailed below.

Gateway:

* New `Accepted` and `Programmed` conditions introduced.
* `Scheduled` condition deprecated.
* Core Conditions now `Accepted` and `Programmed`.
* Moves to Extended: `Ready`.

Gateway Listener:

* New `Accepted` and `Programmed` conditions introduced.
* `Detached` condition deprecated.
* Core Conditions now `Accepted`, `Programmed`, `ResolvedRefs`, and `Conflicted`.
* Moves to Extended: `Ready`.

All Resources:

* The `Accepted` Condition now has a `Pending` reason, which is the default until
  the condition is updated by a controller.

Route resources:

* The `Accepted` Condition now has a `NoMatchingParent` reason, to be set on routes
  when no matching parent can be found.

The purpose of these changes is to make the status flows more consistent across objects, and to provide a clear pattern for new objects as we evolve the API.

> **Note**: This change will require updates for implementations to be able to pass conformance tests. Implementations may choose to publish both new and old conditions, or only new conditions.

- Adds `Accepted` and deprecates `Detached` Listener conditions and reasons (#1446, @mikemorris)
- Adds `Accepted` and deprecates `Scheduled` Gateway conditions and reasons (#1447, @mikemorris)
- Adds `Pending` reason for use with all `Accepted` conditions throughout the API (#1453, @youngnick)
- Adds `Programmed` Gateway and Listener conditions, moves `Ready` to extended
  conformance (#1499, @LCaparelli)
- Add `RouteReasonNoMatchingParent` reason for `Accepted` condition. (#1516, @pmalek)

## Other Changes by type

### Deprecations

- GatewayClass, Gateway, and HTTPRoute are now only supported with the v1beta1
  version of the API. The v1alpha2 API versions of these resources will be fully
  removed in a future release. Additionally, v1alpha2 is marked as deprecated
  everywhere. (#1348 and #1405, @robscott)

### API Changes

- A new field `responseHeaderModifier` is added to `.spec.rules.filters`, which
  allows for modification of HTTP response headers (#1373, @aryan9600)

### Conformance Tests

- ExemptFeatures have been merged into SupportedFeatures providing implementations
  a uniform way to specify the features they support.
  (#1507, @robscott) (#1394, @gyohuangxin)
- To be conformant with the API, if there is no ReferenceGrant that grants a
  listener to reference a secret in another namespace, the
  ListenerConditionReason for the condition ResolvedRefs must be set to
  RefNotPermitted instead of InvalidCertificateRef. (#1305, @mlavacca)

### Developer Notes

- Deprecated `v1alpha2` Go types are now aliases to their `v1beta1` versions
  (#1390, @howardjohn)
- Moved type translation helpers from the `utils` package to a new package named
  `translator`. (#1337, @carlisia)

### Documentation

- Clarify that BackendObjectReference's Port field specifies a service port, not
  a target port, for Kubernetes Service backends. (#1332, @Miciah)
- HTTPRequestHeaderFilter and HTTPResponseHeaderFilter forbid configuring
  multiple actions for the same header. (#1497, @rainest)
- Changes "custom" conformance level to "implementation-specific" (#1436,
  @LCaparelli)
- Clarification that changes to ReferenceGrants MUST be reconciled (#1429,
  @robscott)

## v0.5.1

API versions: v1beta1, v1alpha2

This release includes a number of bug fixes and clarifications:

### API Spec

* The spec has been clarified to state that the port specified in BackendRef
  refers to the Service port number, not the target port, when a Service is
  referenced. [#1332](https://github.com/kubernetes-sigs/gateway-api/pull/1332)
* The spec has been clarified to state that "Accepted" should be used instead of
  "Attached" on HTTPRoute.
  [#1382](https://github.com/kubernetes-sigs/gateway-api/pull/1382)

### Webhook:

* The duplicate gateway-system namespace definitions have been removed.
  [#1387](https://github.com/kubernetes-sigs/gateway-api/pull/1387)
* The webhook has been updated to watch v1beta1.
  [#1365](https://github.com/kubernetes-sigs/gateway-api/pull/1368)

### Conformance:

* The expected condition for a cross-namespace certificate reference that has
  not been allowed by a ReferenceGrant has been changed from
  "InvalidCertificateRef" to "RefNotPermitted" to more closely match the spec.
  [#1351](https://github.com/kubernetes-sigs/gateway-api/pull/1351)
* A new test has been added to cover when a Gateway references a Secret that
  does not exist
  [#1334](https://github.com/kubernetes-sigs/gateway-api/pull/1334)


## v0.5.0

API versions: v1beta1, v1alpha2

This release is all about stability.

Changes in this release can largely be divided into the following categories:

- Release Channels
- Resources graduating to beta
- New experimental features
- Bug Fixes
- General Improvements
- Breaking Changes
  - Validation improvements
  - Internal type cleanup

Note: This release is largely identical to v0.5.0-rc2, this changelog tracks
the difference between v0.5.0 and v0.4.3.

### Release channels

In this release, we've made two release channels available, `experimental` and
`standard`.

The `experimental` channel contains all resources and fields, while `standard`
contains only resources that mave moved to beta status.

We've also added a way to flag particular fields within a resource as
experimental, and any fields marked in this way are only present in the
`experimental` channel. Please see the [versioning][vers] docs for a more
detailed explanation.

One caveat for the standard channel - due to work on the new ReferenceGrant
resource: conformance tests may not pass with the `standard` set of CRDs.

[vers]:https://gateway-api.sigs.k8s.io/concepts/versioning/

### Resources Graduating to BETA

The following APIs have been promoted to a `v1beta1` maturity:

- `GatewayClass`
- `Gateway`
- `HTTPRoute`

[#1192](https://github.com/kubernetes-sigs/gateway-api/pull/1192)

### New Experimental Features

- Routes can now select `Gateway` listeners by port number
  [#1002](https://github.com/kubernetes-sigs/gateway-api/pull/1002)
- Gateway API now includes "Experimental" release channel. Consequently, CRDs now
  include `gateway.networking.k8s.io/bundle-version` and
  `gateway.networking.k8s.io/channel` annotations.
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)
- URL Rewrites and Path redirects have been added as new "Experimental" features
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)

### Bug Fixes

- Fixes a problem that would cause webhook deployment to fail on Kubernetes
  v1.22 and greater.
  [#991](https://github.com/kubernetes-sigs/gateway-api/pull/991)
- Fixes a bug where the `Namespace` could be unspecified in `ReferencePolicy`
  [#964](https://github.com/kubernetes-sigs/gateway-api/pull/964)
- Fixes a bug where v1alpha2 GatewayClass controller names were not being
  shown in the output of `kubectl get gatewayclasses`
  [#909](https://github.com/kubernetes-sigs/gateway-api/pull/909)

### General Improvements

- Conformance tests were introduced with [GEP-917][gep-917] and multiple
  conformance tests were added from a variety of contributors under the
  `conformance/` directory.
- The status of the GatewayClass "Accepted" condition for the `GatewayClass`
  is now present in `kubectl get` output.
  [#1168](https://github.com/kubernetes-sigs/gateway-api/pull/1168)
- New `RouteConditionReason` types `RouteReasonNotAllowedByListeners` and
  `RouteReasonNoMatchingListenerHostname` were added.
  [#1155](https://github.com/kubernetes-sigs/gateway-api/pull/1155)
- New `RouteConditionReason` type added with `RouteReasonAccepted`,
  `RouteReasonResolvedRefs` and `RouteReasonRefNotPermitted` constants.
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)
- Introduced PreciseHostname which prevents wildcard characters in relevant
  Hostname values.
  [#956](https://github.com/kubernetes-sigs/gateway-api/pull/956)

[gep-917]:https://gateway-api.sigs.k8s.io/geps/gep-917/

### Validation Improvements

- Webhook validation now ensures that a path match exists when required by path
  modifier in filter.
  [#1171](https://github.com/kubernetes-sigs/gateway-api/pull/1171)
- Webhook validation was added to ensure that only type-appropriate fields are
  set in `HTTPPathModifier`.
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- The Gateway API webhook is now deployed in a `gateway-system` namespace
  instead of `gateway-api`.
  [#1051](https://github.com/kubernetes-sigs/gateway-api/pull/1051)
- Adds webhook validation to ensure that no HTTP header or query param is
  matched more than once in a given route rule. (#1230, @skriss)

### Breaking Changes

- The v1alpha1 API version was deprecated and removed.
  [#1197](https://github.com/kubernetes-sigs/gateway-api/pull/1197)
  [#906](https://github.com/kubernetes-sigs/gateway-api/issues/906)
- The `NamedAddress` value for `Gateway`'s `spec.addresses[].type` field has
  been deprecated, and support for domain-prefixed values (like
  `example.com/NamedAddress`) has been added instead to better represent the
  custom nature of this support.
  [#1178](https://github.com/kubernetes-sigs/gateway-api/pull/1178)
- Implementations are now expected to use `500` instead of `503` responses when
  the data-plane has no matching route.
  [#1151](https://github.com/kubernetes-sigs/gateway-api/pull/1151),
  [#1258](https://github.com/kubernetes-sigs/gateway-api/pull/1258)

#### UX and Status Improvements

The following are **breaking changes** related to status updates and end-user
experience changes.

- The `UnsupportedExtension` named `ListenerConditionReason` has been removed.
  [#1146](https://github.com/kubernetes-sigs/gateway-api/pull/1146)
- The `RouteConflict` named `ListenerConditionReason` has been removed.
  [#1145](https://github.com/kubernetes-sigs/gateway-api/pull/1145)

#### Internal Type Cleanup

These changes will only affect implementations. Implementors will need to adjust
for the type changes when updating the Gateway API dependency in their projects.

**NOTE**: These kinds of changes are not always present in the CHANGELOG so
          please be aware that the CHANGELOG is not an exhaustive list of Go
          type changes. In this case there were a significant number of changes
          in a single release, so we included them for extra visibility for
          implementors.

- `ReferencePolicy` has been renamed to `ReferenceGrant`.
  [#1179](https://github.com/kubernetes-sigs/gateway-api/pull/1179)
- `GatewayTLSConfig`'s `CertificateRefs` field is now a slice of pointers to
  structs instead of the structs directly.
  [#1176](https://github.com/kubernetes-sigs/gateway-api/pull/1176)
- `HTTPPathModifer` field `Absolute` renamed to `ReplaceFullPath`
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- the `ParentRef` type was renamed to `ParentReference`
  [#982](https://github.com/kubernetes-sigs/gateway-api/pull/982)
- Types `ConditionRouteAccepted` and `ConditionRouteResolvedRefs` are now
  deprecated in favor of `RouteConditionAccepted` & `RouteConditionResolvedRefs`
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)


## v0.5.0-rc2

API versions: v1beta1, v1alpha2

We expect this to be our final release candidate before launching v0.5.0. This
release candidate includes a variety of cleanup and documentation updates.

### Webhook

- Adds webhook validation to ensure that no HTTP header or query param is
  matched more than once in a given route rule. (#1230, @skriss)

### Documentation

- Add examples and documentation for v1beta1 (#1238, @EmilyShepherd)
- Add policy attachment example (#1233, @keithmattix)
- Add warning headers for experimental resources/concepts (#1234, @keithmattix)
- All Enum API fields have had updates to clarify that we may add values at any
  time, and that implementations must handle unknown Enum values. (#1258,
  @youngnick)
- Spacing has been improved around the documentation of feature-level
  core/extended support for better readability and clarity. (#1241, @acnodal-tc)
- Update ReferenceGrant docs to include Gateways that reference a Secret in a
  different namespace (#1181, @nathancoleman)

### Cleanup

- ReferencePolicyList Items is an array of ReferencePolicy again (#1239,
  @dprotaso)
- This release of experimental-install.yaml will apply successfully. Previous
  releases had some extraneous yaml. (#1232, @acnodal-tc)
- The NamedAddress type is back to support backwards compatibility but it is
  still formally deprecated. (#1252, @robscott)

## v0.5.0-rc1

API versions: v1beta1, v1alpha2

This release is all about stability.

Changes in this release can largely be divided into the following categories:

- Release Channels
- Resources graduating to beta
- New experimental features
- Bug Fixes
- General Improvements
- Breaking Changes
  - Validation improvements
  - Internal type cleanup

### Release channels

In this release, we've made two release channels available, `experimental` and
`standard`.

The `experimental` channel contains all resources and fields, while `standard`
contains only resources that mave moved to beta status.

We've also added a way to flag particular fields within a resource as
experimental, and any fields marked in this way are only present in the
`experimental` channel. Please see the [versioning][vers] docs for a more
detailed explanation.

One caveat for the standard channel - due to work on the new ReferenceGrant
resource: conformance tests may not pass with the `standard` set of CRDs.

[vers]:https://gateway-api.sigs.k8s.io/concepts/versioning/

### Resources Graduating to BETA

The following APIs have been promoted to a `v1beta1` maturity:

- `GatewayClass`
- `Gateway`
- `HTTPRoute`

[#1192](https://github.com/kubernetes-sigs/gateway-api/pull/1192)

### New Experimental Features

- Routes can now select `Gateway` listeners by port number
  [#1002](https://github.com/kubernetes-sigs/gateway-api/pull/1002)
- Gateway API now includes "Experimental" release channel. Consequently, CRDs now
  include `gateway.networking.k8s.io/bundle-version` and
  `gateway.networking.k8s.io/channel` annotations.
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)
- URL Rewrites and Path redirects have been added as new "Experimental" features
  [#945](https://github.com/kubernetes-sigs/gateway-api/pull/945)

### Bug Fixes

- Fixes a problem that would cause webhook deployment to fail on Kubernetes
  v1.22 and greater.
  [#991](https://github.com/kubernetes-sigs/gateway-api/pull/991)
- Fixes a bug where the `Namespace` could be unspecified in `ReferencePolicy`
  [#964](https://github.com/kubernetes-sigs/gateway-api/pull/964)
- Fixes a bug where v1alpha2 GatewayClass controller names were not being
  shown in the output of `kubectl get gatewayclasses`
  [#909](https://github.com/kubernetes-sigs/gateway-api/pull/909)

### General Improvements

- Conformance tests were introduced with [GEP-917][gep-917] and multiple
  conformance tests were added from a variety of contributors under the
  `conformance/` directory.
- The status of the GatewayClass "Accepted" condition for the `GatewayClass`
  is now present in `kubectl get` output.
  [#1168](https://github.com/kubernetes-sigs/gateway-api/pull/1168)
- New `RouteConditionReason` types `RouteReasonNotAllowedByListeners` and
  `RouteReasonNoMatchingListenerHostname` were added.
  [#1155](https://github.com/kubernetes-sigs/gateway-api/pull/1155)
- New `RouteConditionReason` type added with `RouteReasonAccepted`,
  `RouteReasonResolvedRefs` and `RouteReasonRefNotPermitted` constants.
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)
- Introduced PreciseHostname which prevents wildcard characters in relevant
  Hostname values.
  [#956](https://github.com/kubernetes-sigs/gateway-api/pull/956)

[gep-917]:https://gateway-api.sigs.k8s.io/geps/gep-917/

### Validation Improvements

- Webhook validation now ensures that a path match exists when required by path
  modifier in filter.
  [#1171](https://github.com/kubernetes-sigs/gateway-api/pull/1171)
- Webhook validation was added to ensure that only type-appropriate fields are
  set in `HTTPPathModifier`.
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- The Gateway API webhook is now deployed in a `gateway-system` namespace
  instead of `gateway-api`.
  [#1051](https://github.com/kubernetes-sigs/gateway-api/pull/1051)

### Breaking Changes

- The v1alpha1 API version was deprecated and removed.
  [#1197](https://github.com/kubernetes-sigs/gateway-api/pull/1197)
  [#906](https://github.com/kubernetes-sigs/gateway-api/issues/906)
- The `NamedAddress` value for `Gateway`'s `spec.addresses[].type` field has
  been deprecated, and support for domain-prefixed values (like
  `example.com/NamedAddress`) has been added instead to better represent the
  custom nature of this support.
  [#1178](https://github.com/kubernetes-sigs/gateway-api/pull/1178)
- Implementations are now expected to use `500` instead of `503` responses when
  the data-plane has no matching route.
  [#1151](https://github.com/kubernetes-sigs/gateway-api/pull/1151)

#### UX and Status Improvements

The following are **breaking changes** related to status updates and end-user
experience changes.

- The `UnsupportedExtension` named `ListenerConditionReason` has been removed.
  [#1146](https://github.com/kubernetes-sigs/gateway-api/pull/1146)
- The `RouteConflict` named `ListenerConditionReason` has been removed.
  [#1145](https://github.com/kubernetes-sigs/gateway-api/pull/1145)

#### Internal Type Cleanup

These changes will only affect implementations. Implementors will need to adjust
for the type changes when updating the Gateway API dependency in their projects.

**NOTE**: These kinds of changes are not always present in the CHANGELOG so
          please be aware that the CHANGELOG is not an exhaustive list of Go
          type changes. In this case there were a significant number of changes
          in a single release, so we included them for extra visibility for
          implementors.

- `ReferencePolicy` has been renamed to `ReferenceGrant`.
  [#1179](https://github.com/kubernetes-sigs/gateway-api/pull/1179)
- `GatewayTLSConfig`'s `CertificateRefs` field is now a slice of pointers to
  structs instead of the structs directly.
  [#1176](https://github.com/kubernetes-sigs/gateway-api/pull/1176)
- `HTTPPathModifer` field `Absolute` renamed to `ReplaceFullPath`
  [#1124](https://github.com/kubernetes-sigs/gateway-api/pull/1124)
- the `ParentRef` type was renamed to `ParentReference`
  [#982](https://github.com/kubernetes-sigs/gateway-api/pull/982)
- Types `ConditionRouteAccepted` and `ConditionRouteResolvedRefs` are now
  deprecated in favor of `RouteConditionAccepted` & `RouteConditionResolvedRefs`
  [#1114](https://github.com/kubernetes-sigs/gateway-api/pull/1114)

## v0.4.3

API version: v1alpha2

This release includes improvements to our webhook, including:

* Migrating kube-webhook-certgen to k8s.gcr.io/ingress-nginx:v1.1.1.
  [#1126](https://github.com/kubernetes-sigs/gateway-api/pull/1126)
* New validation to ensure that a HTTPRouterFilter Type matches its value
  [#1071](https://github.com/kubernetes-sigs/gateway-api/pull/1071)
* A fix to ensure that Path match validation actually works
  [#1071](https://github.com/kubernetes-sigs/gateway-api/pull/1071)

## v0.4.2

API version: v1alpha2

This release is intended to verify our webhook image tagging process.

### Bug Fixes

* Update image generation process with more consistent naming
  [#1034](https://github.com/kubernetes-sigs/gateway-api/pull/1034)

## v0.4.1

API version: v1alpha2

This release contains minor bug fixes for v1alpha2.

### Bug Fixes

* ControllerName now prints correctly in kubectl output for GatewayClass
  [#909](https://github.com/kubernetes-sigs/gateway-api/pull/909)
* Namespace can no longer be left unspecified in ReferencePolicy
  [#964](https://github.com/kubernetes-sigs/gateway-api/pull/964)
* Wildcard characters can no longer be used in redirect Hostname values
  [#956](https://github.com/kubernetes-sigs/gateway-api/pull/956)

## v0.4.0

API version: v1alpha2

This release contains significant breaking changes as we strive for a concise
API. We anticipate that this API will be very similar to a future v1beta1
release.

The following changes have been made since v0.3.0:

### Major Changes

* The Gateway API APIGroup has moved from `networking.x-k8s.io` to
  `gateway.networking.k8s.io`. This means that, as far as the apiserver is
  concerned, this version is wholly distinct from v1alpha1, and automatic
  conversion is not possible. As part of this process, Gateway API is now
  subject to Kubernetes API review, the same as changes made to core API
  resources. More details in
  [#780](https://github.com/kubernetes-sigs/gateway-api/pull/780) and
  [#716](https://github.com/kubernetes-sigs/gateway-api/issues/716).

* Gateway-Route binding changes ([GEP-724](https://gateway-api.sigs.k8s.io/geps/gep-724/)):
  In v1alpha1, Gateways chose which Routes were attached using a combination of
  object and namespace selectors, with the option of also specifying object
  names. This resulted in a very complex config, that's easy to misinterpret. As
  part of v1alpha2, we're changing to:
  * Gateways *may* specify what kind of Routes they support (defaults to same
    protocol if not specified), and where those Routes can be (defaults to same
    namespace).
  * Routes *must* directly reference the Gateways the want to attach to, this is
    a list, so a Route can attach to more than one Gateway.
  * The Route becomes attached only when the specifications intersect.

  We believe this is quite a bit easier to understand, and still gives good
  flexibility for most use cases.
  GEP added in [#725](https://github.com/kubernetes-sigs/gateway-api/pull/725).
  Implemented in [#754](https://github.com/kubernetes-sigs/gateway-api/pull/754).
  Further documentation was added in [#762](https://github.com/kubernetes-sigs/gateway-api/pull/762).

* Safer cross-namespace references ([GEP-709](https://gateway-api.sigs.k8s.io/geps/gep-709/)):
  This concerns (currently), references from Routes to Backends, and Gateways to
  Secrets. The new behavior is:
  * By default, references across namespaces are not permitted; creating a
    reference across a namespace (like a Route referencing a Service in another
    namespace) must be rejected by implementations.
  * These references can be accepted by creating a ReferencePolicy in the
    referent (target) namespace, that specifies what Kind is allowed to accept
    incoming references, and from what namespace and Kind the references may be.

  The intent here is that the owner of the referent namespace must explicitly
  accept incoming references, otherwise we can run into all sorts of bad things
  from breaking the namespace security model.
  Implemented in [#741](https://github.com/kubernetes-sigs/gateway-api/pull/741).

* Attaching Policy to objects ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/)):
  This has been added so that we have an extensible mechanism for adding a
  cascading set of policy to Gateway API objects.

  What policy? Well, it's kind of up to the implementations, but the best example
  to begin with is timeout policy.

  Timeout policy for HTTP connections is highly depedent on how the underlying
  implementation handles policy - it's very difficult to extract commonalities.

  This is intended to allow things like:
  * Attach a policy that specifies the default connection timeout for backends
    to a GatewayClass. All Gateways that are part of that Class will have Routes
    get that default connection timeout unless they specify differently.
  * If a Gateway that's a member of the GatewayClass has a different default
    attached, then that will beat the GatewayClass (for defaults, more specific
    object beats less specific object).
  * Alternatively, a Policy that mandates that you can't set the client timeout
    to "no timeout" can be attached to a GatewayClass as an override. An
    override will always take effect, with less specific beating more specific.

  This one is a bit complex, but will allow implementations to solve some things
  that currently require tools like admission control.
  Implemented in [#736](https://github.com/kubernetes-sigs/gateway-api/pull/736).

* As part of GEP-713, `BackendPolicy` has been removed, as its functionality is
  now better handled using that mechanism.
  [#732](https://github.com/kubernetes-sigs/gateway-api/pull/732).

* Removal of certificate references from HTTPRoutes ([GEP-746](https://gateway-api.sigs.k8s.io/geps/gep-746/)):
  In v1alpha1, HTTPRoute objects have a stanza that allows referencing a TLS
  keypair, intended to allow people to have a more self-service model, where an
  app owner can provision a TLS keypair inside their own namespace, attach it to
  a HTTPRoute they control, and then have that used to secure their app.
  When implementing this, however, there are a large number of edge cases that
  are complex, hard to handle, and poorly defined - about checking SNI, hostname,
  and overrides, that made even writing a spec on how to implement this very
  difficult, let alone actually implementing it.

  In removing certificate references from HTTPRoute, we're using the
  ReferencePolicy from GEP-709 to allow Gateways to securely create a
  cross-namespace reference to TLS keypairs in app namespaces.
  We're hopeful that this will hit most of the self-service use case, and even
  if not, provide a basis to build from to meet it eventually.
  GEP added in [#749](https://github.com/kubernetes-sigs/gateway-api/pull/749).
  Implemented in [#768](https://github.com/kubernetes-sigs/gateway-api/pull/768).

  [GEP-851](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-851.md)
  was a follow up on this change that allowed multiple Certificate Refs per
  Gateway Listener. This was implemented in
  [#852](https://github.com/kubernetes-sigs/gateway-api/pull/852).

* The `RouteForwardTo` (YAML: `routeForwardTo`) struct/stanza has been reworked
  into the `BackendRef` (YAML: `backendRef`) struct/stanza,
  [GEP-718](https://gateway-api.sigs.k8s.io/geps/gep-718/). As part of this
  change, the `ServiceName` (YAML: `serviceName`) field has been removed, and
  Service references must instead now use the `BackendRef`/`backendRef`
  struct/stanza.

### Small Changes
* Extension points within match blocks from all Routes have been removed
  [#829](https://github.com/kubernetes-sigs/gateway-api/pull/829). Implements
  [GEP-820](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-820.md).
  These extension points have been removed because they are currently not used,
  are poorly understood, and we don't have good use cases for them. We may
  consider re-adding them in the future.

* Controller is now a required field in Gateway references from Route status.
  [#671](https://github.com/kubernetes-sigs/gateway-api/pull/671).

* Header Matching, Query Param Matching, and HTTPRequestHeaderFilter now use
  named subobjects instead of maps.
  [#657](https://github.com/kubernetes-sigs/gateway-api/pull/657) and
  [#681](https://github.com/kubernetes-sigs/gateway-api/pull/681)

* [#796](https://github.com/kubernetes-sigs/gateway-api/pull/796) API Review
  suggestions:
  * listener.routes has been renamed to listener.allowedRoutes
  * The `NoSuchGatewayClass` has been removed after it was deprecated in
    v1alpha1
  * `*` is no longer a valid hostname. Instead, leaving hostname unspecified is
    interpreted as `*`.

* The `scope` field has been removed from all object references.
  [#882](https://github.com/kubernetes-sigs/gateway-api/pull/882)

* "Controller" has been renamed to "ControllerName"
  [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)

* "Admitted" condition has been renamed to "Accepted" and now defaults to an
  "Unknown" state instead of "False"
  [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)

* HTTPRequestRedirectFilter's Protocol field has been renamed to Scheme.
  [#863](https://github.com/kubernetes-sigs/gateway-api/pull/863)

* ImplementationSpecific match types in HTTPRoute's path, query, and header
  matches have been removed.
  [#850](https://github.com/kubernetes-sigs/gateway-api/pull/850)

* The "Prefix" path match type has been renamed "PathPrefix".
  [#898](https://github.com/kubernetes-sigs/gateway-api/pull/898)

### Small Additions
* HTTP Method matching is now added into HTTPRoute, with Extended support:
  [#733](https://github.com/kubernetes-sigs/gateway-api/pull/733).

* GatewayClass now has a 'Description' field that is printed as a column in
  `kubectl get` output. You can now end up with output that looks like this:
  ```shell
  $> kubectl get gatewayclass
  NAME       CONTROLLER                            DESCRIPTION
  internal   gateway-controller-internal   For non-internet-facing Gateways.
  external   gateway-controller-external   For internet-facing Gateways.
  ```
  See [#610](https://github.com/kubernetes-sigs/gateway-api/issues/610) and
  [#653](https://github.com/kubernetes-sigs/gateway-api/pull/653) for the
  details.

### Validation changes
* Ensure TLSConfig is empty when the protocol is HTTP, TCP, or UDP
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Ensure Hostname is empty when the protocol is TCP or UDP.
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Listener ProtocolType now has validation.
  [#871](https://github.com/kubernetes-sigs/gateway-api/pull/871)
* HTTP Path match values are now validated for PathMatchExact and
  PathMatchPrefix match types.
  [#894](https://github.com/kubernetes-sigs/gateway-api/pull/894)
* TLS options keys are now subject to the same validation as Kubernetes
  annotations. [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* TLS options values now have a max length of 4096 characters.
  [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Make `MirrorFilter.BackendRef` a required field when the mirror filter is used
  [#837](https://github.com/kubernetes-sigs/gateway-api/pull/837).

### Clarifications
* Updated guidance on how HTTP and TLS Route status should be populated when
  hostnames do not match.
  [#859](https://github.com/kubernetes-sigs/gateway-api/pull/859)
* Aligned path prefix matching with Ingress by clarifying that it is a prefix of
  path elements. [#869](https://github.com/kubernetes-sigs/gateway-api/pull/869)
* HTTP listeners may now be used for Cleartext HTTP/2.
  [#879](https://github.com/kubernetes-sigs/gateway-api/pull/879)
* Added clarification that implementation-specific TLS options MUST be
  domain-prefixed.
  [#899](https://github.com/kubernetes-sigs/gateway-api/pull/899)

### Documentation Updates
* [#782](https://github.com/kubernetes-sigs/gateway-api/pull/782) : Restructure docs and split into versioned and unversioned
* [#777](https://github.com/kubernetes-sigs/gateway-api/pull/777) : Fix typo
* [#765](https://github.com/kubernetes-sigs/gateway-api/pull/765) : document multi-value headers as undefined
* [#761](https://github.com/kubernetes-sigs/gateway-api/pull/761) : minor improvements to navigation on docs site
* [#760](https://github.com/kubernetes-sigs/gateway-api/pull/760) : Remove references of vendor configurations in GatewayTLSConfig
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#755](https://github.com/kubernetes-sigs/gateway-api/pull/755) : Document the supported kubernetes versions
* [#745](https://github.com/kubernetes-sigs/gateway-api/pull/745) : Remove RouteTLSConfig requirement for gateway TLS passthrough.
* [#744](https://github.com/kubernetes-sigs/gateway-api/pull/744) : automate nav for GEPs
* [#743](https://github.com/kubernetes-sigs/gateway-api/pull/743) : Add READY and ADDRESS to gateway printer columns
* [#742](https://github.com/kubernetes-sigs/gateway-api/pull/742) : Moving method match to v1alpha2 example
* [#729](https://github.com/kubernetes-sigs/gateway-api/pull/729) : Adding suggested reasons for when conditions are healthy
* [#728](https://github.com/kubernetes-sigs/gateway-api/pull/728) : Fixing wording in enhancement template
* [#723](https://github.com/kubernetes-sigs/gateway-api/pull/723) : Clarifying Redirect Support levels
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#880](https://github.com/kubernetes-sigs/gateway-api/pull/880) : Reworking Policy vs. Filter Documentation
* [#878](https://github.com/kubernetes-sigs/gateway-api/pull/878) : Clarifying the fields that all Route types must include
* [#875](https://github.com/kubernetes-sigs/gateway-api/pull/875) : Fix HTTP path match documentation.
* [#864](https://github.com/kubernetes-sigs/gateway-api/pull/864) : Merging v1alpha2 concepts docs into unversioned docs
* [#858](https://github.com/kubernetes-sigs/gateway-api/pull/858) : Fixing broken link to spec page
* [#857](https://github.com/kubernetes-sigs/gateway-api/pull/857) : Adding missing references pages to docs navigation
* [#853](https://github.com/kubernetes-sigs/gateway-api/pull/853) : docs: Use v0.4.0-rc1 in "Getting started with Gateway APIs" for v1alpha2
* [#845](https://github.com/kubernetes-sigs/gateway-api/pull/845) : Fix markdown list formatting.
* [#844](https://github.com/kubernetes-sigs/gateway-api/pull/844) : docs: add ssl passthrough note in FAQ
* [#843](https://github.com/kubernetes-sigs/gateway-api/pull/843) : Add APISIX implementation
* [#834](https://github.com/kubernetes-sigs/gateway-api/pull/834) : Fixes some broken links
* [#807](https://github.com/kubernetes-sigs/gateway-api/pull/807) : docs: update multiple-ns guide for v1alpha2
* [#888](https://github.com/kubernetes-sigs/gateway-api/pull/888) : Corrected broken getting started
* [#885](https://github.com/kubernetes-sigs/gateway-api/pull/885) : Fix incorrect urls
* [#890](https://github.com/kubernetes-sigs/gateway-api/pull/890) : Updating HTTPRoute docs for v1alpha2
* [#870](https://github.com/kubernetes-sigs/gateway-api/pull/870) : Adding guidance on Kind vs. Resource in implementation guidelines
* [#865](https://github.com/kubernetes-sigs/gateway-api/pull/865) : Route cleanup for v1alpha2 sig-network review

### Tooling and infra updates
* [#766](https://github.com/kubernetes-sigs/gateway-api/pull/766) : comment out the GEP notice
* [#758](https://github.com/kubernetes-sigs/gateway-api/pull/758) : bump up mkdocs and deps
* [#751](https://github.com/kubernetes-sigs/gateway-api/pull/751) : bump up deps to k8s v1.22
* [#748](https://github.com/kubernetes-sigs/gateway-api/pull/748) : fix kustomize to install v1a2 crds
* [#747](https://github.com/kubernetes-sigs/gateway-api/pull/747) : Cleaning up GEP Template
* [#889](https://github.com/kubernetes-sigs/gateway-api/pull/889) : remove outdated version label
* [#883](https://github.com/kubernetes-sigs/gateway-api/pull/883) : validating webhook cleanup
* [#872](https://github.com/kubernetes-sigs/gateway-api/pull/872) : Remove duplicate validation from CRD & Webhook

## v0.4.0-rc2

API version: v1alpha2

The group expects that this release candidate has no changes before we release
v1alpha2 final, but are cutting here to allow implementations a chance to check
before we go to the final release.

In general, most of the changes below have been made to reduce the complexity of
the API for v1alpha2, on the assumption that we can add functionality in later
in the API's lifecycle, but cannot remove it.

The following changes have been made since v0.4.0-rc1:

### GEP implementations
* Replace `CertificateRef` field with `CertificateRefs` in `GatewayTLSConfig`.
[#852](https://github.com/kubernetes-sigs/gateway-api/pull/852). This implements
[GEP-851](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-851.md),
Allow Multiple Certificate Refs per Gateway Listener.
* Extension points within match blocks from all Routes have been removed
[#829](https://github.com/kubernetes-sigs/gateway-api/pull/829). Implements
[GEP-820](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-820.md).
These extension points have been removed because they are currently not used,
are poorly understood, and we don't have good use cases for them. We may
consider re-adding them in the future.

### Field changes
* Make `MirrorFilter.BackendRef` a required field when the mirror filter is used
[#837](https://github.com/kubernetes-sigs/gateway-api/pull/837).
* ImplementationSpecific match types in HTTPRoute's path, query, and header
matches have been removed.
[#850](https://github.com/kubernetes-sigs/gateway-api/pull/850)
* The "Prefix" path match type has been renamed "PathPrefix".
* The "ClassName" field in PolicyTargetReference has been removed.
* A new optional "Name" field has been added to ReferencePolicyTo.
[#898](https://github.com/kubernetes-sigs/gateway-api/pull/898)

### Field Renames
* "Controller" has been renamed to "ControllerName"
* "Admitted" condition has been renamed to "Accepted" and now defaults to an
"Unknown" state instead of "False" [#839](https://github.com/kubernetes-sigs/gateway-api/pull/839)
* HTTPRequestRedirectFilter's Protocol field has been renamed to Scheme.
[#863](https://github.com/kubernetes-sigs/gateway-api/pull/863)


### Validation changes
*  Validation: Ensure TLSConfig is empty when the protocol is HTTP, TCP, or UDP
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
*  Validation: Ensure Hostname is empty when the protocol is TCP or UDP.
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Validation: Listener ProtocolType now has validation.
[#871](https://github.com/kubernetes-sigs/gateway-api/pull/871)
* Validation: HTTP Path match values are now validated for PathMatchExact and
PathMatchPrefix match types. [#894](https://github.com/kubernetes-sigs/gateway-api/pull/894)

### Documentation and specification updates
* Updated guidance on how HTTP and TLS Route status should be populated when
hostnames do not match.
[#859](https://github.com/kubernetes-sigs/gateway-api/pull/859)
* Aligned path prefix matching with Ingress by clarifying that it is a prefix of
path elements. [#869](https://github.com/kubernetes-sigs/gateway-api/pull/869)
* HTTP listeners may now be used for Cleartext HTTP/2.
[#879](https://github.com/kubernetes-sigs/gateway-api/pull/879)
* The `scope` field has been removed from all object references.
* ParentRefs can no longer refer to cluster-scoped resources.
[#882](https://github.com/kubernetes-sigs/gateway-api/pull/882)
* TLS options keys are now subject to the same validation as Kubernetes
annotations. [#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* TLS options values now have a max length of 4096 characters.
[#886](https://github.com/kubernetes-sigs/gateway-api/pull/886)
* Added clarification that implementation-specific TLS options MUST be domain-prefixed.
[#899](https://github.com/kubernetes-sigs/gateway-api/pull/899)

### Other changes
* [#890](https://github.com/kubernetes-sigs/gateway-api/pull/890) : Updating HTTPRoute docs for v1alpha2
* [#889](https://github.com/kubernetes-sigs/gateway-api/pull/889) : remove outdated version label
* [#888](https://github.com/kubernetes-sigs/gateway-api/pull/888) : Corrected broken getting started
* [#885](https://github.com/kubernetes-sigs/gateway-api/pull/885) : Fix incorrect urls
* [#883](https://github.com/kubernetes-sigs/gateway-api/pull/883) : v1alpha2 validation fix/update
* [#880](https://github.com/kubernetes-sigs/gateway-api/pull/880) : Reworking Policy vs. Filter Documentation
* [#878](https://github.com/kubernetes-sigs/gateway-api/pull/878) : Clarifying the fields that all Route types must include
* [#875](https://github.com/kubernetes-sigs/gateway-api/pull/875) : Fix HTTP path match documentation.
* [#872](https://github.com/kubernetes-sigs/gateway-api/pull/872) : Remove duplicate validation from CRD & Webhook
* [#870](https://github.com/kubernetes-sigs/gateway-api/pull/870) : Adding guidance on Kind vs. Resource in implementation guidelines
* [#865](https://github.com/kubernetes-sigs/gateway-api/pull/865) : Route cleanup for v1alpha2 sig-network review
* [#864](https://github.com/kubernetes-sigs/gateway-api/pull/864) : Merging v1alpha2 concepts docs into unversioned docs
* [#858](https://github.com/kubernetes-sigs/gateway-api/pull/858) : Fixing broken link to spec page
* [#857](https://github.com/kubernetes-sigs/gateway-api/pull/857) : Adding missing references pages to docs navigation
* [#853](https://github.com/kubernetes-sigs/gateway-api/pull/853) : docs: Use v0.4.0-rc1 in "Getting started with Gateway APIs" for v1alpha2
* [#845](https://github.com/kubernetes-sigs/gateway-api/pull/845) : Fix markdown list formatting.
* [#844](https://github.com/kubernetes-sigs/gateway-api/pull/844) : docs: add ssl passthrough note in FAQ
* [#843](https://github.com/kubernetes-sigs/gateway-api/pull/843) : Add APISIX implementation
* [#834](https://github.com/kubernetes-sigs/gateway-api/pull/834) : Fixes some broken links
* [#807](https://github.com/kubernetes-sigs/gateway-api/pull/807) : docs: update multiple-ns guide for v1alpha2


## v0.4.0-rc1

API version: v1alpha2

The working group expects that this release candidate is quite close to the final
v1alpha2 API. However, breaking API changes are still possible.

This release candidate is suitable for implementors, but the working group does
not recommend shipping products based on a release candidate API due to the
possibility of incompatible changes prior to the final release.

### Major Changes

* The Gateway API APIGroup has moved from `networking.x-k8s.io` to
`gateway.networking.k8s.io`. This means that, as far as the apiserver is
concerned, this version is wholly distinct from v1alpha1, and automatic conversion
is not possible. As part of this process, Gateway API is now subject to Kubernetes
API review, the same as changes made to core API resources. More details in
[#780](https://github.com/kubernetes-sigs/gateway-api/pull/780) and [#716](https://github.com/kubernetes-sigs/gateway-api/issues/716).

* Gateway-Route binding changes:
[GEP-724](https://gateway-api.sigs.k8s.io/geps/gep-724/). Currently, Gateways
choose which Routes are attached using a combination of object and namespace
selectors, with the option of also specifying object names. This has made a very
complex config, that's easy to misinterpret. As part of v1alpha2, we're changing to:
  * Gateways *may* specify what kind of Routes they support (defaults to same
  protocol if not specified), and where those Routes can be (defaults to same
  namespace).
  * Routes *must* directly reference the Gateways the want to attach to, this is
  a list, so a Route can attach to more than one Gateway.
  * The Route becomes attached only when the specifications intersect.

  We believe this is quite a bit easier to understand, and still gives good
  flexibility for most use cases.
  GEP added in [#725](https://github.com/kubernetes-sigs/gateway-api/pull/725).
  Implemented in [#754](https://github.com/kubernetes-sigs/gateway-api/pull/754).
  Further documentation was added in [#762](https://github.com/kubernetes-sigs/gateway-api/pull/762).


* Safer cross-namespace references:
([GEP-709](https://gateway-api.sigs.k8s.io/geps/gep-709/)): This concerns
(currently), references from Routes to Backends, and Gateways to Secrets. The
new behavior is:
  * By default, references across namespaces are not permitted; creating a
  reference across a namespace (like a Route referencing a Service in another
  namespace) must be rejected by implementations.
  * These references can be accepted by creating a ReferencePolicy in the
  referent (target) namespace, that specifies what Kind is allowed to accept
  incoming references, and from what namespace and Kind the references may be.

  The intent here is that the owner of the referent namespace must explicitly
  accept incoming references, otherwise we can run into all sorts of bad things
  from breaking the namespace security model.
  Implemented in [#741](https://github.com/kubernetes-sigs/gateway-api/pull/741).

* Attaching Policy to objects:
[GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/): This has been added
so that we have an extensible mechanism for adding a cascading set of policy to
Gateway API objects.

  What policy? Well, it's kind of up to the implementations, but the best example
  to begin with is timeout policy.

  Timeout policy for HTTP connections is highly depedent on how the underlying
  implementation handles policy - it's very difficult to extract commonalities.

  This is intended to allow things like:
  * Attach a policy that specifies the default connection timeout for backends
  to a GatewayClass. All Gateways that are part of that Class will have Routes
  get that default connection timeout unless they specify differently.
  * If a Gateway that's a member of the GatewayClass has a different default
  attached, then that will beat the GatewayClass (for defaults, more specific
  object beats less specific object).
  * Alternatively, a Policy that mandates that you can't set the client timeout
  to "no timeout" can be attached to a GatewayClass as an override. An override
  will always take effect, with less specific beating more specific.

  This one is a bit complex, but will allow implementations to solve some things
  that currently require tools like admission control.
  Implemented in [#736](https://github.com/kubernetes-sigs/gateway-api/pull/736).

* As part of GEP-713, `BackendPolicy` has been removed, as its functionality is
now better handled using that mechanism. [#732](https://github.com/kubernetes-sigs/gateway-api/pull/732).

* Removal of certificate references from HTTPRoutes:
[GEP-746](https://gateway-api.sigs.k8s.io/geps/gep-746/):
  In v1alpha1, HTTPRoute objects have a stanza that allows referencing a TLS
  keypair, intended to allow people to have a more self-service model, where an
  app owner can provision a TLS keypair inside their own namespace, attach it to
  a HTTPRoute they control, and then have that used to secure their app.
  When implementing this, however, there are a large number of edge cases that
  are complex, hard to handle, and poorly defined - about checking SNI, hostname,
  and overrides, that made even writing a spec on how to implement this very
  difficult, let alone actually implementing it.

  In removing certificate references from HTTPRoute, we're using the
  ReferencePolicy from GEP-709 to allow Gateways to securely create a
  cross-namespace reference to TLS keypairs in app namespaces.
  We're hopeful that this will hit most of the self-service use case, and even
  if not, provide a basis to build from to meet it eventually.
  GEP added in [#749](https://github.com/kubernetes-sigs/gateway-api/pull/749).
  Implemented in [#768](https://github.com/kubernetes-sigs/gateway-api/pull/768).

* The `RouteForwardTo` (YAML: `routeForwardTo`) struct/stanza has been reworked
into the `BackendRef` (YAML: `backendRef`) struct/stanza,
[GEP-718](https://gateway-api.sigs.k8s.io/geps/gep-718/). As part of this change,
the `ServiceName` (YAML: `serviceName`) field has been removed, and Service
references must instead now use the `BackendRef`/`backendRef` struct/stanza.

### Other changes
* HTTP Method matching is now added into HTTPRoute, with Extended support:
[#733](https://github.com/kubernetes-sigs/gateway-api/pull/733).

* GatewayClass now has a 'Description' field that is printed as a column in
`kubectl get` output. You can now end up with output that looks like this:
  ```shell
  $> kubectl get gatewayclass
  NAME       CONTROLLER                            DESCRIPTION
  internal   gateway-controller-internal   For non-internet-facing Gateways.
  external   gateway-controller-external   For internet-facing Gateways.
  ```
  See [#610](https://github.com/kubernetes-sigs/gateway-api/issues/610) and
  [#653](https://github.com/kubernetes-sigs/gateway-api/pull/653) for the details.

*  [#671](https://github.com/kubernetes-sigs/gateway-api/pull/671): Controller is
now a required field in Gateway references from Route status. Fixes
[#669](https://github.com/kubernetes-sigs/gateway-api/pull/671).

*  [#657](https://github.com/kubernetes-sigs/gateway-api/pull/657): and
[#681](https://github.com/kubernetes-sigs/gateway-api/pull/681) Header Matching,
Query Param Matching, and HTTPRequestHeaderFilter now use named subobjects
instead of maps.

* [#796](https://github.com/kubernetes-sigs/gateway-api/pull/796) API Review suggestions:
  * listener.routes has been renamed to listener.allowedRoutes
  * The `NoSuchGatewayClass` has been removed after it was deprecated in v1alpha1
  * `*` is no longer a valid hostname. Instead, leaving hostname unspecified is interpreted as `*`.

### Documentation Updates
* [#782](https://github.com/kubernetes-sigs/gateway-api/pull/782) : Restructure docs and split into versioned and unversioned
* [#777](https://github.com/kubernetes-sigs/gateway-api/pull/777) : Fix typo
* [#765](https://github.com/kubernetes-sigs/gateway-api/pull/765) : document multi-value headers as undefined
* [#761](https://github.com/kubernetes-sigs/gateway-api/pull/761) : minor improvements to navigation on docs site
* [#760](https://github.com/kubernetes-sigs/gateway-api/pull/760) : Remove references of vendor configurations in GatewayTLSConfig
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName
* [#755](https://github.com/kubernetes-sigs/gateway-api/pull/755) : Document the supported kubernetes versions
* [#745](https://github.com/kubernetes-sigs/gateway-api/pull/745) : Remove RouteTLSConfig requirement for gateway TLS passthrough.
* [#744](https://github.com/kubernetes-sigs/gateway-api/pull/744) : automate nav for GEPs
* [#743](https://github.com/kubernetes-sigs/gateway-api/pull/743) : Add READY and ADDRESS to gateway printer columns
* [#742](https://github.com/kubernetes-sigs/gateway-api/pull/742) : Moving method match to v1alpha2 example
* [#729](https://github.com/kubernetes-sigs/gateway-api/pull/729) : Adding suggested reasons for when conditions are healthy
* [#728](https://github.com/kubernetes-sigs/gateway-api/pull/728) : Fixing wording in enhancement template
* [#723](https://github.com/kubernetes-sigs/gateway-api/pull/723) : Clarifying Redirect Support levels
* [#756](https://github.com/kubernetes-sigs/gateway-api/pull/756) : Clarify docs on invalid serviceName

### Tooling and infra updates
* [#766](https://github.com/kubernetes-sigs/gateway-api/pull/766) : comment out the GEP notice
* [#758](https://github.com/kubernetes-sigs/gateway-api/pull/758) : bump up mkdocs and deps
* [#751](https://github.com/kubernetes-sigs/gateway-api/pull/751) : bump up deps to k8s v1.22
* [#748](https://github.com/kubernetes-sigs/gateway-api/pull/748) : fix kustomize to install v1a2 crds
* [#747](https://github.com/kubernetes-sigs/gateway-api/pull/747) : Cleaning up GEP Template


## v0.3.0

API Version: v1alpha1

### API changes

#### Gateway
- The `NoSuchGatewayClass` status reason has been deprecated.
  [#635](https://github.com/kubernetes-sigs/gateway-api/pull/635)

#### HTTPRoute
- `.spec.rules.matches.path` now has a default `prefix` match on the `/` path.
  [#584](https://github.com/kubernetes-sigs/gateway-api/pull/584)
- Conflict resolution guidance has been added for rules within a route.
  [#620](https://github.com/kubernetes-sigs/gateway-api/pull/620)
- HTTPRoute now supports query param matching.
  [#631](https://github.com/kubernetes-sigs/gateway-api/pull/631)

#### All Route Types
- Route status now includes controller name for each Gateway.
  [#616](https://github.com/kubernetes-sigs/gateway-api/pull/616)
- Conflict resolution guidance has been added for non-HTTP routes.
  [#626](https://github.com/kubernetes-sigs/gateway-api/pull/626)

#### Misc
- Fields of type LocalObjectRef do not default to "secrets". All LocalObjectRef
  fields must be specified.
  [#570](https://github.com/kubernetes-sigs/gateway-api/pull/570)
- CRDs have been added to gateway-api category
  [#592](https://github.com/kubernetes-sigs/gateway-api/pull/592)
- New "Age" column has been added to all resources in `kubectl get` output.
  [#592](https://github.com/kubernetes-sigs/gateway-api/pull/592)
- A variety of Go types have been changed to pointers to better reflect their
  optional status.
  [#564](https://github.com/kubernetes-sigs/gateway-api/pull/564)
  [#572](https://github.com/kubernetes-sigs/gateway-api/pull/572)
  [#579](https://github.com/kubernetes-sigs/gateway-api/pull/579)

#### Validation
- A new experimental validation package and validating webhook have been added.
  [#597](https://github.com/kubernetes-sigs/gateway-api/pull/597)
  [#617](https://github.com/kubernetes-sigs/gateway-api/pull/617)


## v0.2.0

API Version: v1alpha1

### API changes

Service APIs has been renamed to Gateway API.
[#536](https://github.com/kubernetes-sigs/service-apis/issues/536).


#### GatewayClass
- The default status condition of GatewayClass resource is now `Admitted:false`
  instead of `InvalidParameters:Unknown`.
  [#471](https://github.com/kubernetes-sigs/service-apis/pull/471).
- `GatewayClass.spec.parametersRef` now has an optional `namespace` field to
  refer to a namespace-scoped resource in addition to cluster-scoped resource.
  [#543](https://github.com/kubernetes-sigs/service-apis/pull/543).

#### Gateway
- `spec.listeners[].tls.mode` now defaults to `Terminate`.
  [#518](https://github.com/kubernetes-sigs/service-apis/pull/518).
- Empty `hostname` in a listener matches all request.
  [#525](https://github.com/kubernetes-sigs/service-apis/pull/525).

#### HTTPRoute
- New `set` property has been introduced for `HTTPRequestHeader` Filter. Headers
  specified under `set` are overriden instead of added.
  [#475](https://github.com/kubernetes-sigs/service-apis/pull/475).

#### Misc
- Maximum limit for `forwardTo` has been increased from `4` to `16` for all
  route types.
  [#493](https://github.com/kubernetes-sigs/service-apis/pull/493).
- Various changes have been made in the Kubernetes and Go API to align with
  upstream Kubernetes API conventions. Some of the fields have been changed to
  pointers in the Go API for this reason.
  [#538](https://github.com/kubernetes-sigs/service-apis/pull/538).

### Documentation

There are minor improvements to docs all around.
New guides, clarifications and various typos have been fixed.

## v0.1.0

API Version: v1alpha1

### API changes since v0.1.0-rc2
#### GatewayClass
- CRD now includes `gc` short name.
- Change the standard condition for GatewayClass to `Admitted`, with
  `InvalidParameters` as a sample reason for it to be false.

#### Gateway
- CRD now includes `gtw` short name.
- The `DroppedRoutes` condition has been renamed to `DegradedRoutes`.
- `ListenerStatus` now includes `Protocol` and `Hostname` to uniquely link the
  status to each listener.

#### Routes
- HTTPRoute clarifications:
  - Header name matching must be case-insensitive.
  - Match tiebreaking semantics have been outlined in detail.
- TCPRoute, TLSRoute, and UDPRoute:
  - At least 1 ForwardTo must be specified in each rule.
  - Clarification that if no matches are specified, all requests should match a
    rule.
- TCPRoute and UDPRoute: Validation has been added to ensure that 1-16 rules are
  specified, matching other route types.
- TLSRoute: SNIs are now optional in matches. If no SNI or extensionRef are
  specified, all requests match.

#### BackendPolicy
- CRD now includes `bp` short name.
- A new `networking.x-k8s.io/app-protocol` annotation can be used to specify
  AppProtocol on Services when the AppProtocol field is unavailable.


## v0.1.0-rc2

API Version: v1alpha-rc2

### API changes since v0.1.0-rc1
#### GatewayClass
- A recommendation to set a `gateway-exists-finalizer.networking.x-k8s.io`
  finalizer on GatewayClass has been added.
- `allowedGatewayNamespaces` has been removed from GatewayClass in favor of
  implementations with policy agents like Gatekeeper.

#### Gateway
- Fields in `listeners.routes` have been renamed:
  - `routes.routeSelector` -> `routes.selector`
  - `routes.routeNamespaces`-> `routes.namespaces`
- `clientCertificateRef` has been removed from BackendPolicy.
- In Listeners, `routes.namespaces` now defaults to `{from: "Same"}`.
- In Listeners, support has been added for specifying custom, domain prefixed
  protocols.
- In Listeners, `hostname` now closely matches Route hostname matching with wildcard
  support.
- A new `UnsupportedAddress` condition has been added to Listeners to indicate
  that a requested address is not supported.
- Clarification has been added to note that listeners may be merged in certain
  instances.

#### Routes
- HeaderMatchType now includes a RegularExpression option.
- Minimum weight has been decreased from 1 to 0.
- Port is now required on all Routes.
- On HTTPRoute, filters have been renamed:
  - `ModifyRequestHeader` -> `RequestHeaderModifier`
  - `MirrorRequest` -> `RequestMirror`
  - `Custom` -> `ExtensionRef`
- TLSRoute can now specify as many as 16 SNIs instead of 10.
- Limiting the number of Gateways that may be stored in RouteGatewayStatus to
  100.
- Support level of filters defined in ForwardTo has been clarified.
- Max weight has been increased to 1 million.


## v0.1.0-rc1

API Version: v1alpha-rc1

- Initial release candidate for v1alpha1.
