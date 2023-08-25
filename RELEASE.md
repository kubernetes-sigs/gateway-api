# Release Process

## Overview

The Gateway API project is an API project that has the following two components:
- Kubernetes Custom Resource Definitions (CRDs)
- Corresponding Go API in the form of `sigs.k8s.io/gateway-api` Go package

This repository is the home for both of the above components.

## Versioning strategy
The versioning strategy for this project is covered in detail in [the release
documentation].

[the release documentation]: https://gateway-api.sigs.k8s.io/concepts/versioning/

## Releasing a new version

### Writing a Changelog

To simplify release notes generation, we recommend using the [Kubernetes release
notes generator](https://github.com/kubernetes/release/blob/master/cmd/release-notes):

```
go install k8s.io/release/cmd/release-notes@latest
export GITHUB_TOKEN=your_token_here
release-notes --start-sha EXAMPLE_COMMIT --end-sha EXAMPLE_COMMIT --branch main --repo gateway-api --org kubernetes-sigs
```

This output will likely need to be reorganized and cleaned up a bit, but it
provides a good starting point. Once you're satisfied with the changelog, create
a PR. This must go through the regular PR review process and get merged into the
`main` branch. Approval of the PR indicates community consensus for a new
release.

### Release Steps

The following steps must be done by one of the [Gateway API maintainers][gateway-api-team]:

For a **PATCH** release:
- Create a new branch in your fork named something like `<githubuser>/release-x.x.x`. Use the new branch
  in the upcoming steps.
- Use `git` to cherry-pick all relevant PRs into your branch.
- Update `pkg/generator/main.go` with the new semver tag and any updates to the API review URL.
- Run the following command `BASE_REF=vmajor.minor.patch make generate` which
  will update generated docs and webhook with the correct version info. (Note
  that you can't test with these YAMLs yet as they contain references to
  elements which wont exist until the tag is cut and image is promoted to
  production registry.)
- Create a pull request of the `<githubuser>/release-x.x.x` branch into the `release-x.x` branch upstream
  (which should already exist since this is a patch release). Add a hold on this PR waiting for at least
  one maintainer/codeowner to provide a `lgtm`.
- Verify the CI tests pass and merge the PR into `release-x.x`.
- Create a tag using the `HEAD` of the `release-x.x` branch. This can be done using the `git` CLI or
  Github's [release][release] page.
- Run the `make build-install-yaml` command which will generate install files in the `release/` directory.
  Attach these files to the Github release.
- Update the `README.md` and `site-src/guides/index.md` files to point links and examples to the new release.

For a **MAJOR** or **MINOR** release:
- Cut a `release-major.minor` branch that we can tag things in as needed.
- Check out the `release-major.minor` release branch locally.
- Update `pkg/generator/main.go` with the new semver tag and any updates to the API review URL.
- Run the following command `BASE_REF=vmajor.minor.patch make generate` which
  will update generated docs and webhook with the correct version info. (Note
  that you can't test with these YAMLs yet as they contain references to
  elements which wont exist until the tag is cut and image is promoted to
  production registry.)
- Verify the CI tests pass before continuing.
- Create a tag using the `HEAD` of the `release-x.x` branch. This can be done using the `git` CLI or
  Github's [release][release] page.
- Run the `make build-install-yaml` command which will generate install files in the `release/` directory.
  Attach these files to the Github release.
- Update the `README.md` and `site-src/guides/index.md` files to point links and examples to the new release.

For an **RC** release:
- Update `pkg/generator/main.go` with the new semver tag and any updates to the API review URL.
- Run the following command `BASE_REF=vmajor.minor.patch make generate` which
  will update generated docs and webhook with the correct version info. (Note
  that you can't test with these YAMLs yet as they contain references to
  elements which wont exist until the tag is cut and image is promoted to
  production registry.)
- Include the changelog update in this PR.
- Merge the update PR.
- Tag the release using the commit on `main` where the changelog update merged.
  This can  be done using the `git` CLI or Github's [release][release]
  page.
- Run the `make build-install-yaml` command which will generate
  install files in the `release/` directory.
- Attach these files to the Github release.

### Promoting images to production registry
Gateway API follows the standard kubernetes image promotion process described [here][kubernetes-image-promotion]. 

1. Once the tag has been cut and the image is available in the staging registry,
   identify the SHA-256 image digest of the image that you want to promote.
2. Modify the
   [k8s-staging-gateway-api/images.yaml](https://github.com/kubernetes/k8s.io/blob/main/registry.k8s.io/images/k8s-staging-gateway-api/images.yaml)
   file under [kubernetes/k8s.io](https://github.com/kubernetes/k8s.io)
   repository and add the image digest along with the new tag under the correct
   component.
   1. Currently, the following images are included: `admission-server`, `echo-server`
3. Create a PR with the above changes.
4. Image will get promoted by [automated prow jobs][kubernetes-image-promotion]
   once the PR merges

[release]: https://github.com/kubernetes-sigs/gateway-api/releases
[gateway-api-team]: https://github.com/kubernetes/org/blob/main/config/kubernetes-sigs/sig-network/teams.yaml
[kubernetes-image-promotion]: https://github.com/kubernetes/k8s.io/tree/main/registry.k8s.io#image-promoter
