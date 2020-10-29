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

## Developing Service APIs

You must have a working [Go environment] and then clone the repo:

```
mkdir -p $GOPATH/src/sigs.k8s.io
cd $GOPATH/src/sigs.k8s.io
git clone https://github.com/kubernetes-sigs/service-apis
cd service-apis
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
We use [milestones][gh-milestones] to track our progress towards releases. Looking at our current milestone can help identify our highest priority issues.

[gh-issues]: https://github.com/kubernetes-sigs/service-apis/issues
[gh-dashboard]: https://github.com/kubernetes-sigs/service-apis/projects/1
[gh-milestones]: https://github.com/kubernetes-sigs/service-apis/milestones

## Building the code

The project uses `make` to drive the build.
`make` will run code generators, and run static analysis against the code and
generate Kubernetes CRDs.
You can kick off an overall build from the top-level makefile:

```shell
make
```

## Install CRDs

To install service-apis CRDs into a Kubernetes cluster:

```shell
make install
```

To uninstall CRDs and associated resources:

```shell
make uninstall
```

## Submitting a Pull Request

Service APIs follows a similar pull request process as [Kubernetes].
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

[prow-setup]: https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/service-apis

## Documentation

The site documentation is written in [mkdocs][mkdocs] format. The files are
contained in `docs-src/`. Generated files are in `docs/` and published to
Github Pages.

Building the docs:

```shell
make docs
```

Live preview for editing (view on [http://localhost:8000](http://localhost:8000), CTRL-C to quit):

```shell
make serve
```

Remove generated documentation files:

```shell
make clean
```

### Publishing

The docs are published automatically to [Github pages][ghp]. When making changes to the
documentation, generate the new documentation and make the generated code a
self-contained commit (e.g. the changes to `docs/`). This will keep the code
reviews simple and clearly delineate user vs generated content.

[ghp]: https://kubernetes-sigs.github.io/service-apis/
[mkdocs]: https://www.mkdocs.org/
[Go environment]: https://golang.org/doc/install
[Kubernetes]: https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md
