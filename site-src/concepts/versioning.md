# Versioning

## Overview
Each new release of Gateway API is defined with a "bundle version" that
represents the Git tag of a release, such as v0.4.0. This contains the
following:

* API Types (Go bindings for the resources)
* CRDs (Kubernetes definitions of the resources)
* Validating Webhook

### API Versions (e.g. v1alpha2, v1beta1)
Within Gateway API, API versions are primarily used to indicate the stability of
a resource. For example, if a resource has not yet graduated to beta, it is
still possible that it could either be removed from the API or changed in
backwards incompatible ways. For more information on API versions, refer to the
[full Kubernetes API versioning
documentation](https://kubernetes.io/docs/reference/using-api/#api-versioning).

![Lifecycle of New Gateway API Resources](/images/lifecycle-new-resources.png)
<!-- Source: https://docs.google.com/presentation/d/1sfZTV-vlisDUIie_iK_B2HqKia_querT6m6T2_vbAk0/edit -->

### Release Channels (e.g. Experimental, Standard)
Gateway API provides 2 release channels: an Experimental one and a Standard one.

The Standard release channel includes:

* Resources that have graduated to beta
* All fields that have graduated to standard and are no longer considered
  experimental

The Experimental release channel includes everything in the Standard release
channel, plus:

* Alpha API resources
* New fields that are considered experimental and have not yet graduated to the
  standard channel

![Release Channel Overlap](/images/release-channel-overlap.svg)
<!-- Source: https://docs.google.com/presentation/d/1sfZTV-vlisDUIie_iK_B2HqKia_querT6m6T2_vbAk0/edit -->

Release channels are used to indicate feature stability within Gateway API. For
example, even if a resource has already graduated to beta, we need a way to test
out new fields and features before committing to their long term stability. With
upstream Kubernetes APIs, this is accomplished with [feature
gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/).
Unfortunately, CRDs do not have a similar concept yet.

Instead of trying to recreate feature gates in this project, we've introduced
release channels. Starting in v0.5.0, all new fields and features will start in
the Experimental release channel. From that point, these may graduate to the
Standard release channel or be dropped from the API entirely. See
[GEP-922](/geps/gep-922) for a more detailed discussion of this approach to new
fields and features.

![Lifecycle of New Gateway API Fields](/images/lifecycle-new-fields.svg)
<!-- Source: https://docs.google.com/presentation/d/1sfZTV-vlisDUIie_iK_B2HqKia_querT6m6T2_vbAk0/edit -->

## Version Indicators
Each CRD will be published with annotations that indicate their bundle version
and channel:

```
gateway.networking.k8s.io/bundle-version: v0.4.0
gateway.networking.k8s.io/channel: standard|experimental
```

## What can Change
When using or implementing this API, it is important to understand what can
change across API versions.

### Patch version (e.g. v0.4.0 -> v0.4.1)
* Clarifications to godocs
* Updates to CRDs and/or code to fix a bug
* Conformance test fixes
* Additional conformance test coverage for existing features
* Fixes to typos

### Minor version (e.g. v0.4.0 -> v0.5.0)
* Everything that is valid in a patch release
* New experimental API fields or resources
* Changes to recommended conditions or reasons in status
* Loosened validation
* Making required fields optional
* Removal of experimental fields
* Removal of experimental resources
* Graduation of fields or resources from experimental to standard track
* Changes to conformance tests to match spec updates
* Introduction of a new **API version**, which may also include:
  * Renamed fields
  * Anything else that is valid in a new Kubernetes API version
  * Removal/tombstoning of beta fields
* Removal of an API resource following [Kubernetes deprecation
  policy](https://kubernetes.io/docs/reference/using-api/deprecation-policy/)

### Major version (e.g. v0.x to v1.0)
* There are no API compatibility guarantees when the major version changes.

## Graduation Criteria

### Resources

#### Alpha -> Beta
A resource to graduate from alpha to beta must meet the following criteria:

* Implemented by several implementations.
* Conformance test framework is in place, with some coverage of basic
  functionality.
* Validation is well thought out.
* Most of the API surface is being exercised by users.
* Approval from subproject owners + KEP reviewers.

#### Beta -> GA

A resource to graduate from beta to GA must meet the following criteria:

* Almost all of the fields and behavior have conformance test coverage.
* Multiple conformant implementations.
* Widespread implementation and usage.
* At least 6 months of soak time as a beta API.
* Approval from subproject owners + KEP reviewers.

### Fields

#### Experimental -> Standard
As described above, field level stability is layered on top of beta and GA
resources (no fields in alpha resources can be considered standard). The
requirements for a field to graduate from experimental to standard depend on the
API version of the resource it is a part of. For a field to be considered
standard, it needs to meet the same criteria of the resource it is contained in.

If a resource has graduated to beta, an experimental field must meet all of the
beta graduation criteria before graduating to standard. Similarly, if a resource
has graduated to GA, a field must meet all of the beta and GA graduation
criteria. There is one slight variation here, instead of 6 months of soak time
as a beta API, a field graduating to standard requires 6 months of soak time as an
experimental field.

## Out of Scope
### Unreleased APIs
This project will have frequent updates to the main branch. There are no
compatibility guarantees associated with code in any branch, including main,
until it has been released. For example, changes may be reverted before a
release is published. For the best results, use the latest published release of
this project.

### Source Code
We do not provide stability guarantees for source code imports. The Interfaces
and behavior may change in an unexpected and backwards-incompatible way in any
future release.

## Supported Versions

This project aims to provide support for a wide range of Kubernetes versions with
consistent upgrade experiences across versions. To accomplish that, we commit to:

1. Support a minimum of the most recent 5 Kubernetes minor versions.
2. Ensure that all standard channel changes between v1beta1 and v1 are fully
   compatible and convertible.
3. Take every possible effort to avoid introduction of a conversion webhook. If
   a conversion webhook needs to be introduced, it will be supported for the
   lifetime of the API, or at least until an alternative is available.
