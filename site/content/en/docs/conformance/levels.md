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

### Security Contact and Embargo Policy

Conformance reports include an `implementation.securityContact` email address (required starting in Gateway API v1.8) used to coordinate security advisories with Gateway API implementations, similar to the [Kubernetes Security Response Committee (formerly Product Security Committee) private distributors list][k8s-src-distributors].

Because conformance reports are stored in a public repository, implementations should provide an **actively monitored project security email alias or mailing list** (e.g., `security@example.com` or `psirt@example.com`) rather than an individual's personal email address. This list is intended to provide actionable information to implementation teams rather than individuals, and using a dedicated alias allows implementations to manage spam filtering and moderation.

By submitting a conformance report with a `securityContact` email address, you are agreeing to abide by a non-disclosure agreement (NDA) and embargo policy, and to avoid any form of early disclosure of any security advisories released by Gateway API:

* **Dedicated Security Alias**: Provide an actively monitored security email alias for your project or organization rather than a personal email address.
* **Confidentiality & Need-to-Know**: Information received via the security contact may only be shared internally within your team on a strict need-to-know basis to fix the issue in your implementation, and anyone with whom it is shared must agree to these same terms.
* **No Early Disclosure**: Vulnerability details, patches, and derived binaries or releases must not be made public, distributed to external parties, or hinted at prior to the agreed-upon public disclosure date and time.
* **Leak Notification**: If embargoed information leaks prematurely, you must urgently inform the Gateway API maintainers and `security@kubernetes.io` with details of what leaked and to whom.

{{% alert color="warning" title="Disclaimer" %}}
Security advisories are released on a **best-effort basis**. There is no guarantee that all potential security issues will be released in this way or with advance notice.
{{% /alert %}}

[reports-folder]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/
[conformance-guide]: https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/README.md
[k8s-src-distributors]: https://github.com/kubernetes/committee-security-response/blob/main/private-distributors-list.md

## Prerelease reports

For handling prerelease reports, which we can use to prove that an implementation _will_ support a forthcoming feature, we will handle differently. The intent here is that they are used as evidence that we can graduate a feature from Experimental to Standard.

Prerelease reports are stored in a different location and are _not_ used to generate conformance information such as the Implementations List, the feature matrices, or the feature wizard.

Prerelease reports MAY use a commit hash rather than a tagged version, as they are evidence a feature has been implemented, not a record for end users.

Prerelease reports MAY also only run tests for features that the implementation needs to prove they support.
