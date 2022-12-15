<!--
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

## Developing Gateway API

You must have a working [Go environment] and then clone the repo:

```
mkdir -p $GOPATH/src/sigs.k8s.io
cd $GOPATH/src/sigs.k8s.io
git clone https://github.com/kubernetes-sigs/gateway-api
cd gateway-api
```

This project works with Go modules; you can chose to setup your environment
outside $GOPATH as well.

# Building, testing and deploying

You will need to have Docker installed to perform the steps below.

## Project management

We are using the Github issues and project dashboard to manage the list of TODOs
for this project:

* [Open issues][gh-issues]
* [Project dashboard][gh-dashboard]

Issues labeled `good first issue` and `help wanted` are especially good for a
first contribution.

We use [milestones][gh-milestones] to track our progress towards releases.
These milestones are generally labeled according to the [semver][semver]
release version tag that they represent, meaning that in general we only focus
on the next release in the sequence until it is closed and the release is
finished. Only Gateway API maintainers are able to create and attach issues to
milestones.

We use [priority labels][prio-labels] to help indicate the timing importance of
resolving an issue, or whether an issue needs more support from its creator or
the community to be prioritized. These labels can be set with the [/priority
command in PR and issue comments][issue-cmds]. For example,
`/priority important-soon`.

[gh-issues]: https://github.com/kubernetes-sigs/gateway-api/issues
[gh-dashboard]: https://github.com/kubernetes-sigs/gateway-api/projects/1
[gh-milestones]: https://github.com/kubernetes-sigs/gateway-api/milestones
[semver]:https://semver.org/
[prio-labels]:https://github.com/kubernetes-sigs/gateway-api/labels?q=priority
[issue-cmds]:https://prow.k8s.io/command-help?repo=kubernetes-sigs%2Fgateway-api

## Building the code

The project uses `make` to drive the build.
`make` will run code generators, and run static analysis against the code and
generate Kubernetes CRDs.
You can kick off an overall build from the top-level makefile:

```shell
make
```

## Adding Experimental Fields

Starting with v0.5.0, all additions to the API must start in the experimental
release channel. Experimental fields must be marked with the
`<gateway:experimental>` annotation in Go type definitions. Gateway API CRD
generation will only include these fields in the experimental set of CRDs.

If experimental fields are removed or renamed, the original field name should be
removed from the go struct, with a tombstone comment
([example](https://github.com/kubernetes/kubernetes/blob/707b8b6efd1691b84095c9f995f2c259244e276c/staging/src/k8s.io/api/core/v1/types.go#L4444-L4445))
ensuring the field name will not be reused.

## Install CRDs

To install gateway-api CRDs into a Kubernetes cluster:

```shell
make install
```

To uninstall CRDs and associated resources:

```shell
make uninstall
```

## Submitting a Pull Request

Gateway API follows a similar pull request process as [Kubernetes].
Merging a pull request requires the following steps to be completed before the
pull request will be merged automatically.

- [Sign the CLA](https://git.k8s.io/community/CLA.md) (prerequisite)
- [Open a pull request](https://help.github.com/articles/about-pull-requests/)
- Pass [verification](#verify) tests
- Get all necessary approvals from reviewers and code owners

### Verify

Make sure you run the static analysis over the repo before submitting your
changes. The [Prow presubmit][prow-setup] will not let your change merge if
verification fails.

```shell
make verify
```

[prow-setup]: https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/gateway-api

## Documentation

The site documentation is written in Markdown and compiled with [mkdocs]. Each PR
will automatically include a [Netlify] deploy preview. When new code merges, it will
automatically be deployed with Netlify to [gateway-api.sigs.k8s.io]. If you want to manually
preview docs changes locally, you can install mkdocs and run:

```shell
 make docs
```

To make it easier to use the right version of [mkdocs], there is a `.venv`
target to create a Python virtualenv that includes [mkdocs]. To use the
[mkdocs] live preview server while you edit, you can run [mkdocs] from
the virtualenv:

```shell
$ make .venv
Creating a virtualenv in .venv... OK
To enter the virtualenv type "source .venv/bin/activate", to exit type "deactivate"
(.venv) $ source .venv/bin/activate
(.venv) $ mkdocs serve
INFO    -  Building documentation...
...
```

[mkdocs]: https://www.mkdocs.org/
[Netlify]: https://netlify.com/
[gateway-api.sigs.k8s.io]: https://gateway-api.sigs.k8s.io
[Go environment]: https://golang.org/doc/install
[Kubernetes]: https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md
