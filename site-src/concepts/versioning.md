# Versioning

## Summary
Each Gateway API release is represented by a bundle version that represents
that specific combination of CRDs, API versions, and validating webhook. To
enable experimental fields, future releases of the API will include stable and
experimental CRD tracks. Users will be able to access experimental features by
installing the experimental CRDs. A cluster can contain either an experimental
or stable CRD for any resource at a given time. 

## Version Terminology
Gateway API has two related but different versioning schemes:

### 1. API Versions (ex: v1alpha2)
Each API version provides a unique way to interact with the API. Significant
changes such as removing or renaming fields will be represented here.

### 2. Bundle Versions (ex: v0.4.0)
Bundle versions are broader in scope. They use semantic versioning and include
the following:

* API Types/CRDs
* Validating Webhook

Each bundle may include multiple API versions, potentially introducing new ones
and/or removing old ones as part of a new bundle.

## Limitations of Webhook and CRD Validation
CRD and webhook validation is not the final validation i.e. webhook is “nice UX”
but not schema enforcement. This validation is intended to provide immediate
feedback to users when they provide an invalid configuration, but can not
completely be relied on because it:

* Is not guaranteed to be present or up to date in all clusters.
* Will likely never be sufficient to cover all edge cases.
* May be loosened in future API releases.

## Persona Requirements
When implementing or using Gateway API, each persona has a unique set of
responsibilities to ensure we're providing a consistent experience.

### API Authors:
* MUST provide conversion between API versions (excluding experimental fields),
  starting with v1alpha2.
* MAY include the following changes to an existing API version with a new bundle
 **patch** version:
    * Clarifications to godocs
    * Updates to CRDs and/or code to fix a bug
    * Fixes to typos
* MAY include the following changes to an existing API version with a new bundle
  **minor** version:
    * Everything that is valid in a patch release
    * New experimental API fields or resources
    * Loosened validation
    * Making required fields optional
    * Removal of experimental fields
    * Removal of experimental resources
    * Graduation of fields or resources from experimental to stable track
* MAY introduce a new **API version** with a new bundle minor version, which may
  include:
    * Everything that is valid in a minor release
    * Renamed fields
    * Anything else that is valid in a new Kubernetes API version
    * Removal/tombstoning of beta fields
* MAY release a new major bundle version (v1.0) as part of graduating the API to
  GA and releasing a new API version.

Note that each new bundle version, no matter how small, may include updated
CRDs, webhook, or both. Implementations may read annotations on Gateway API CRDs
(defined below) to determine the version and channel of CRDs that have been
installed in the cluster.

### Implementers:
* MUST handle fields with loosened validation without crashing
* MUST handle fields that have transitioned from required to optional without
  crashing
* MUST NOT rely on webhook or CRD validation as a security mechanism. If field
  values need to be escaped to secure an implementation, both webhook and CRD
  validation can be bypassed and cannot be relied on. Instead, implementations
  should implement their own escaping or validation as necessary. To avoid
  duplicating work, Gateway API maintainers are considering adding a shared
  validation package that implementations can use for this purpose. This is
  tracked by [#926](https://github.com/kubernetes-sigs/gateway-api/issues/926).

### Installers:
* MUST install a full Gateway API bundle, with matching CRD and webhook
  versions.

## Adding Experimental Fields
Over time, it will be useful to add experimental fields to the API. In upstream
Kubernetes, those would generally be guarded with feature gates. With Gateway
API we will accomplish by releasing experimental versions of our CRDs.

With this approach, we achieve a similar result. Instead of using feature gates
and validation to prevent fields from being set, we just release separate CRDs.
Once the API reaches beta, each bundle release can include 2 sets of CRDs,
stable and experimental.

New fields will be added to the experimental set of CRDs first, and may graduate
to stable APIs later. Experimental fields will be marked with the
`+experimental` annotation in Go type definitions. Gateway API CRD generation
will exclude these fields from stable CRDs. Experimental fields may be removed
from the API. Due to the experimental nature of these CRDs, they are not
recommended for production use.

If experimental fields are removed or renamed, the original field name should be
removed from the go struct, with a tombstone comment ensuring the field name
will not be reused. 

Each CRD will be published with annotations that indicate their bundle version
and channel:

```
gateway.networking.k8s.io/bundle-version: v0.4.0
gateway.networking.k8s.io/channel: stable|experimental
```
