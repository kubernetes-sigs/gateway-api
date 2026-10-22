---
title: "Changelog"
weight: 999
---

This page lists changes made to the Gateway API conformance process, descending by version.
Only versions that included changes to conformance are listed.

## PROPOSED for Gateway API v1.8

* Core behavior conformance test skips are not allowed, and render a profile nonconformant.

## Gateway API v1.7

* ALL profiles in a report must now be conformant for the report to be conformant.
* Implementations are conformant as long as one entire report is Conformant.
* The Partial Conformance state is removed. Implementations are either conformant or they are not.
* Re-introduce the idea of Stale conformance, and apply it only to Gateway API version `n-2`.
This effectively moves the Partial Conformance due to stale supports to the Stale state.
* No core test skips are allowed, and only 5 tests _in total_ across a report are allowed.
* Conformance reports may only be submitted from released versions, which must be tagged
and usable by the public.
* Add a new process for prerelase reports, which may be commit hashes, and are used
for proving that an implementation _will_ support a Gateway API version.


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

