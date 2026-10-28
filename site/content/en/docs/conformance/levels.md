---
title: "Levels"
weight: 10
---

There are two levels of Gateway API conformance for implementations. These are
determined based on the submission of conformance reports to the Gateway API
repo.

### Conformant

Implementations may submit multiple conformance reports.

For a report to be Conformant, it:

* MUST be from a _publicly available_ version of the implementation, available to download and/or use via a tagged version of the implementation itself, hosted platform version, or evalation version of closed-source software. Commit hashes are not acceptable, but this MAY be a prerelease, of a form like `v1.2.3-pre.0`, `v1.2.3-alpha1` or similar. This ensures that reports are reproducible.
* MUST have _all_ profiles in the report pass conformance, with no skipped tests. Skipped tests result in Partial Conformance.
* MUST have at least one report Conformant if multiple reports are submitted for different modes or similar reason.
* MUST be for the most recent Gateway API release (`n`), or _one_ version before (`n-1`).

So, it's conformant to support Mesh + HTTPRoute, or Gateway + HTTPRoute, or
Gateway + TLSRoute, or Gateway + Mesh + HTTPRoute, plus any extended features
the implementation claims. But implementations _must_ support at least one
Profile and one Route type in that profile, and must pass all Core conformance
tests for that Profile and Route type in addition to all claimed Extended
features.

### Partial Conformance

For an implementation to have Partial conformance, it:

* MAY have a profile report a `partial` conformance result, while having the rest of the first three points above be true (released version, at least one report passes these rules)
* AND/OR the supported Gateway API version is from exactly _two_ versions before the current release (`n-2`)

That is, a Conformant implementation which has the supported Gateway APIO version `n-2` can only ever be Partially Conformant.

Implementations with `partial` conformance can only ever be Partially Conformant, no matter the Gateway API version.

The Partial Conformance state has been deprecated and will be removed entirely in the v1.8 release.

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

## Prerelease reports

For handling prerelease reports, which we can use to prove that an implementation _will_ support a forthcoming feature, we will handle differently. The intent here is that they are used as evidence that we can graduate a feature from Experimental to Standard.

Prerelease reports are stored in a different location and are _not_ used to generate conformance information such as the Implementations List, the feature matrices, or the feature wizard.

Prerelease reports MAY use a commit hash rather than a tagged version, as they are evidence a feature has been implemented, not a record for end users.

Prerelease reports MAY also only run tests for features that the implementation needs to prove they support.
