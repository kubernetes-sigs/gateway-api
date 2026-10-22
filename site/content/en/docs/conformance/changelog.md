---
title: "Changelog"
weight: 999
---

This page lists changes made to the Gateway API conformance process, descending by version.
Only versions that included changes to conformance are listed.

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

