---
title: "Changelog"
weight: 999
---

This page lists changes made to the Gateway API conformance process, descending by version.
Only versions that included changes to conformance are listed.

## PROPOSED for Gateway API v1.8

* Partial Conformance is removed. Only fully conformant implementations are counted as conformant.
* Conformance reports with skipped tests will not be accepted. For extended features, you will
not be able to claim support for a feature unless you pass all tests for that feature.

## Gateway API v1.7

* ALL profiles in a report must now be at least partially conformant for the report to be conformant.
* Implementations are conformant as long as one entire report is Conformant.
* Partial Conformance is deprecated and will be removed in the next release. In this release, a
maximum of 5 skipped tests _in total_ across a report are allowed for partial conformance.
* Conformance reports must be reproducible, so Conformance reports may only be submitted from released
versions, which must be usable by the public. This could mean a tagged version for a downloadable
implementation, a specific platform version for a hosted service, or an evaluation version for a closed-source product.
* A concept of prerelease reports has been added. These reports will be used to ensure that features
will have enough conformant implementations to graduate to Standard.


## Gateway API v1.6

* Stale category removed.

As of this version:

If the current release of Gateway API is version `n`:

* Implementations that submitted reports for version `n` or `n-1` may be **Conformant** if they
meet the other requirements.
* Implementation that submitted reports for version `n-2` may only be **Partially Conformant**,
even they meet all other criteria for conformance.
* Implememtations that submitted reports for version `n-3` or older, are no longer conformant.

## Gateway API v1.3

* Stale category introduced. 

As of this version:

If the current release of Gateway API is version `n`:

* Implementations that submitted reports for version `n` or `n-1` may be conformant if they
meet the other requirements.
* Implementation that submitted reports for version `n-2` may only be **Partially Conformant**,
even they meet all other criteria for conformance.
* Implememtations that submitted reports for version `n-3` or older, are **Stale**.


## Before Gateway API v1.3

There were no limitations on conformance aside from requiring conformance reports to have
at least one conformance profile that passed Core conformance for at least one resource.

