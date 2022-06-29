# Release Process

## Overview

The Gateway API project is an API project that has the following two components:
- Kubernetes Custom Resource Definitions (CRDs)
- Corresponding Go API in the form of `sigs.k8s.io/gateway-api` Go package

This repository is the home for both of the above components.

## Versioning strategy
The versioning strategy for this project is covered in detail in [the release
documentation].

[the release documentation]: https://gateway-api.sigs.k8s.io/releases/#versioning

## Releasing a new version

- Write the [changelog](CHANGELOG.md) with user-visible API changes. This must
  go through the regular PR review process and get merged into the `main` branch.
  Approval of the PR indicates community consensus for a new release.

The following steps must be done by one of the [Gateway API maintainers][gateway-api-team]:

For a major or minor release:
- Cut a `release-major.minor` branch that we can tag things in as needed.
For all releases:
- Check out the `release-major.minor` branch locally
- Update `pkg/generator/main.go` with the new semver tag and any updates to the API review URL.
- Run the following commnand `BASE_REF=vmajor.minor.patch make generate` which will update generated docs
  and webhook with the correct version info. Note that the YAMLs will not work until the tag is actually
  published in the next step.
- Publish a new Git tag. This can  be done using the `git` CLI or Github's [release][release]
  page.
- Run the `make build-install-yaml` command which will generate
  install files in the `release/` directory and attach these files to
  the Github release.

[release]: https://gateway-api.sigs.k8s.io/references/releases/
[gateway-api-team]: https://github.com/kubernetes/org/blob/main/config/kubernetes-sigs/sig-network/teams.yaml

