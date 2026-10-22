---
title: "Levels"
weight: 10
---

There are two levels of Gateway API conformance for implementations. These are
determined based on the submission of conformance reports to the Gateway API
repo.

### Conformant

Conformant implementations have submitted at least one conformance report that:

  * All core conformance tests for at least one combination of Route type and
    Profile
  * All claimed Extended features for one of the two (2) most recent Gateway API releases.

So, it's conformant to support Mesh + HTTPRoute, or Gateway + HTTPRoute, or
Gateway + TLSRoute, or Gateway + Mesh + HTTPRoute, plus any extended features
the implementation claims. But implementations _must_ support at least one
Profile and one Route type in that profile, and must pass all Core conformance
tests for that Profile and Route type in addition to all claimed Extended
features.

### Partially Conformant

These implementations are aiming for full conformance but are not currently
achieving it. They have submitted at least one conformance report passing some
of the tests to be Conformant (as above) for one of the three (3) most recent
Gateway API releases. Note that the requirements to be considered "partially
conformant" may be tightened in a future release of Gateway API.

## Implementation traffic types

Implementations may also support two types of traffic:

* **Gateway** controllers reconcile the Gateway resource and are intended to
handle north-south traffic, mainly concerned with coming from outside the
cluster to inside.
* **Mesh** controllers reconcile Service resources with HTTPRoutes attached
and are intended to handle east-west traffic, within the same cluster or
set of clusters.

Each parent resource has a set of conformance tests associated with it, that lay out
the expected behavior for implementations to be conformant (as above).

Implementations may also handle both parent resources.

## Conformance Reports

In case the suite has been configured to run a set of conformance profiles, and
all the Conformance reports fields have been properly set according to [this guide][conformance-guide],
the conformance report can be uploaded into [this folder][reports-folder] through
a PR. Follow the [guide][conformance-guide] to have all the details on how to structure
and submit reports for a specific implementation.

For help running the test suite, see the [Running the suite](/docs/conformance/running-tests/) guide.

[reports-folder]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/
[conformance-guide]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/README.md

