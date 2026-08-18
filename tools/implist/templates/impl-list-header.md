---
title: "Implementations"
linkTitle: "List"
weight: 1
---

This document tracks downstream implementations and integrations of Gateway API
and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this
document with status information about their implementations, the versions they
cover, and documentation to help users get started. This status information should
be no longer than a few paragraphs.

## Conformance levels

There are three levels of Gateway API conformance:

### Conformant implementations

These implementations have submitted at least one conformance report that has passes for:

  * All core conformance tests for at least one combination of Route type and
    Profile
  * All claimed Extended features

for one of the two (2) most recent Gateway API releases.

So, it's conformant to support Mesh + HTTPRoute, or Gateway + HTTPRoute, or
Gateway + TLSRoute, or Gateway + Mesh + HTTPRoute, plus any extended features
the implementation claims. But implementations _must_ support at least one
Profile and one Route type in that profile, and must pass all Core conformance
tests for that Profile and Route type in addition to all claimed Extended
features.

### Partially Conformant implementations

These implementations are aiming for full conformance but are not currently
achieving it. They have submitted at least one conformance report passing some
of the tests to be Conformant (as above) for one of the three (3) most recent
Gateway API releases. Note that the requirements to be considered "partially
conformant" may be tightened in a future release of Gateway API.

### Stale implementations

These implementations may not be being actively developed and will be removed
from this page on the next page review unless they submit a conformance report
moving them to one of the other categories.

Page reviews are performed at least one month after every Gateway API release,
with the first being performed after the release of Gateway API v1.3, in late
June 2025. Following the Gateway API v1.5 review process, due in mid-2026,
stale implementations will no longer be listed.

## Implementation profiles

Implementations also generally fall into two categories, which are called
_profiles_:

* **Gateway** controllers reconcile the Gateway resource and are intended to
handle north-south traffic, mainly concerned with coming from outside the
cluster to inside.
* **Mesh** controllers reconcile Service resources with HTTPRoutes attached
and are intended to handle east-west traffic, within the same cluster or
set of clusters.

Each profile has a set of conformance tests associated with it, that lay out
the expected behavior for implementations to be conformant (as above).

Implementations may also fit both profiles.

## Integrations

Also listed on this page are **integrations**, which are other software
projects that are able to make use of Gateway API resources to perform
other functions (like managing DNS or creating certificates).

{{% alert color="primary" %}}
This page contains links to third party projects that provide functionality
required for Gateway API to work. The Gateway API project authors aren't
responsible for these projects, which are listed alphabetically within their
class.

{{% /alert %}}

{{% alert color="info" title="Compare extended supported features across implementations" %}}

[View a table to quickly compare supported features of projects](/docs/implementations/versions/v1.4/). These outline Gateway controller implementations that have passed core conformance tests, and focus on extended conformance features that they have implemented. These tables will be generated and uploaded to the site once at least 3 implementations have uploaded their conformance reports under the [conformance reports](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports).

{{% /alert %}}
