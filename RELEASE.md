# Release Process

This details how releases of Gateway API are delivered. This process is
exercised by a [Release Manager](/RELEASE_MANAGEMENT.md).

## Overview

The Gateway API project is an API project that has the following two components:
- Kubernetes Custom Resource Definitions (CRDs)
- Corresponding Go API in the form of `sigs.k8s.io/gateway-api` Go package

This repository is the home for both of the above components.

## Versioning strategy
The versioning strategy for this project is covered in detail in [the release
documentation].

[the release documentation]: https://gateway-api.sigs.k8s.io/concepts/versioning/

## Releasing a monthly version

### Starting point

Make sure all the changes that should be part of the monthly release are
merged into `main`.

### Tagging the monthly release

Start by tagging the `main` branch with a tag of the form
`monthly-YYYY.MM` (for example, `monthly-2025.11`). Push this to GitHub.

### Start the CI workflow

Trigger the [`monthly-release`] GitHub workflow, passing it the
`monthly-YYYY.MM` tag just created.

[`monthly-release`]: https://github.com/kubernetes-sigs/gateway-api/actions/workflows/monthly-release.yml

CI handles the rest of the release process, including creating a
**draft** GitHub release which includes an automatically-generated
changelog and the various release artifacts. **You will need to
publish** this release after making sure that the correct artifacts
are attached to it and that the CHANGELOG is what you want.

### Writing the Release Changelog

In many cases, the changelog that GitHub generates is going to be OK.
However, if there are significant changes in a given monthly, it can be
helpful to write a changelog that's more human-readable. Given the `$TAG`
used for this monthly release as well as the `$PREV_TAG` used for the
previous monthly release or GA release, you can get a more complete set
of commit messages with:

```
git log --stat ${PREV_TAG}..${TAG} -- config/crd/experimental
```

and a full set of diffs with:

```
git diff ${PREV_TAG}..${TAG} -- config/crd/experimental
```

It can be helpful to then summarize anything significant in a more human
way in the GitHub release notes. (You won't commit these notes back to
the repository: they will only be saved in the GitHub release itself.)

## Releasing a new standard-channel version

Every new major release gets a `release-1.Y` branch -- so all the work
for 1.5.0, 1.5.1, etc, plus all their RCs, needs to happen on
`release-1.5`. Typically, this release branch will be cut from `main` at
the point where most of the work for the first RC is complete, to
minimize cherry-picks.

Once the release branch is cut, the typical pattern is:

- Work primarily meant for that release should be PR'd into the release
  branch. For example, updating the CHANGELOG and such will happen in a
  PR into the release branch.

- Work that needs to go into a release but should also be carried into
  future major releases should be PR'd into `main`, and cherry-picked
  into the release. For example, a fix for a bug found in 1.5.0-rc.2
  would be PR'd into `main`, but we'd also give the `/cherry-pick
  release-1.5` command to Prow for that PR.

These are rules of thumb, not laws of the universe. Judgment is always
required.

### Writing a Changelog

To simplify release notes generation, we recommend using the [Kubernetes
release notes generator].

[Kubernetes release notes generator]: https://github.com/kubernetes/release/blob/master/cmd/release-notes

```
go install k8s.io/release/cmd/release-notes@latest
export GITHUB_TOKEN=your_token_here
release-notes generate \
  --repo gateway-api --org kubernetes-sigs \
  --required-author ""
  --branch release-1.X \
  --start-sha EXAMPLE_COMMIT --end-sha EXAMPLE_COMMIT \
  --output relnotes.md
```

This will take longer than you might expect it to, but assuming you've
picked good start and end SHAs, it should give you something useful to
start with. You'll all but certainly need to reorganize it and clean it
up a bit; notably, it's a very good idea to break out a more
human-readable high-level summary for the beginning of the notes.

The CHANGELOG will go into `CHANGELOG/1.X-CHANGELOG.md`.

### Release Steps

The following steps must be done by a release manager, who will need elevated Git permissions from one of the [Gateway API maintainers][gateway-api-team].

**ALWAYS start by making sure that the release branch exists.**

If the release branch does not exist:

- If you're doing a "first RC", which is to say a 1.x.0-rc.1 release, go ahead and cut the release branch from `main`.
- If you're doing **any other** kind of release: **STOP. GO FIND A MAINTAINER. DO NOT TOUCH A THING.**

Given that the release branch exists:

- Make sure that everything meant to be in the patch release has been landed on the release branch.
- Create a `release-$tag` branch off the release branch.
- Update the CHANGELOG as described above.
- Update `pkg/consts/consts.go` with the new semver tag and any updates to the API review URL.
- Update regex `spec.validations.expression` in
  `config/crd/standard/gateway.networking.k8s.io_vap_safeupgrades.yaml`
  to match older versions. (Look for a regex like `v1.[0-n].`, and make
  sure that  the `n` in `0-n` is the current minor version number -1.)
- Run the following command `BASE_REF=vmajor.minor.patch make generate` which
  will update generated docs with the correct version info. (Note that you can't
  test with these YAMLs yet as they contain references to elements which wont
  exist until the tag is cut and image is promoted to production registry.)
- Commit all of the above to the new `release-x.x.x` branch.
- Create a pull request of the `release-x.x.x` branch into the `release-x.x` branch upstream. Add a hold on this PR waiting for at least one maintainer/codeowner to provide a `lgtm`. Approval
of the PR is the community consensus for a new release.
- Verify the CI tests pass and merge the PR into `release-x.x`.
- Tag `HEAD` of the `release-x.x` branch with the version number (including the initial `v`, so e.g. `v1.5.0-rc.1` or `v1.6.1`). This can be done using the `git` CLI or
  the GitHub UI, but **note well**: if the release manager can't create the tag due to Git permissions, a maintainer will need to do it, and in that case it's more polite for the maintainer to create and push the _tag_, then let the release manager create the _release_, so that it's easier for people to find the manager if there are problems!
- Run the `make build-install-yaml` command which will generate install files in the `release/` directory.
  Attach these files to the GitHub release.
- Update the `README.md` and `site-src/guides/index.md` files to point links and examples to the new release.

#### For a **MAJOR** or **MINOR** release:
- Cut a `release-major.minor` branch that we can tag things in as needed.
- Check out the `release-major.minor` release branch locally.
- Update `pkg/consts/consts.go` with the new semver tag and any updates to the API review URL.
- Update `config/crd/standard/gateway.networking.k8s.io_vap_safeupgrades.yaml`
  - Update the `gateway.networking.k8s.io/bundle-version`.
  - Update regex `spec.validations.expression` to match older versions. (Look for a regex like `v1.[0-3].`, and replace the `3` with the new minor version number -1).
- Run the following command `BASE_REF=vmajor.minor.patch make generate` which
  will update generated docs with the correct version info. (Note that you can't
  test with these YAMLs yet as they contain references to elements which wont
  exist until the tag is cut and image is promoted to production registry.)
- Verify the CI tests pass before continuing.
- Create a tag using the `HEAD` of the `release-x.x` branch. This can be done using the `git` CLI or
  GitHub's [release][release] page.
- Run the `make build-install-yaml` command which will generate install files in the `release/` directory.
  Attach these files to the GitHub release.
- Update the `README.md` and `site-src/guides/index.md` files to point links and examples to the new release.
- Edit the text blurb in `hack/docsy-generate-conformance.py` to reflect the added past version if necessary.

#### For an **RC** release:
- Update `pkg/consts/consts.go` with the new semver tag (like `v1.2.0-rc1`) and any updates to the API review URL.
- Run the following command `make generate` which
  will update generated docs with the correct version info. (Note that you can't
  test with these YAMLs yet as they contain references to elements which wont
  exist until the tag is cut and image is promoted to production registry.)
- Include the changelog update in this PR.
- Merge the update PR.
- Tag the release using the commit on `main` where the changelog update merged.
  This can  be done using the `git` CLI or GitHub's [release][release]
  page.
- Run the `make build-install-yaml` command which will generate
  install files in the `release/` directory.
- Attach these files to the GitHub release.

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
