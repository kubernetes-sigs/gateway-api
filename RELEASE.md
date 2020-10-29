# Release Process

## Overview

The Service APIs project is an API project that has the following two components:
- Kubernetes Custom Resource Definitions (CRDs)
- Corresponding Go API in the form of `sigs.k8s.io/service-apis` Go package

This repository is the home for both of the above components.

## Versioning strategy

Releases in this repository follow Go module versioning conventions and
semantic versioning.
During the `alpha` and `beta` stage of the API, version tags will take the form
of `v0.x.y`.
Minor version must be incremented whenever a new API version is introduced for
any resource or even when smaller backwards-compatible additions are made to the API.
Bug fixes and clarifications in the spec will lead to patch number being incremented.

During GA (when `apiVersion` changes to `v1`), the Git version tag will be bumped
up to `v1.0.0`.

> The first release candidate was tagged as `v1alpha1-rc1`. It predates this
document and is an exception.

## Releasing a new version

- Write the [changelog](CHANGELOG.md) with user-visible API changes. This must
  go through the regular PR review process and get merged into the `master` branch.
  Approval of the PR indicates community consensus for a new release.
- Once the above PR is merged, the author must publish a new Git tag. This can
  be done using the `git` CLI or Github's [release][release]
  page. This step can be performed only by [Service APIs maintainers][service-apis-team].

[release]: https://github.com/kubernetes-sigs/service-apis/releases
[service-apis-team]: https://github.com/kubernetes/org/blob/master/config/kubernetes-sigs/sig-network/teams.yaml

