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

# Building, testing and deploying

You will need to have Docker installed to perform the steps below.

## Project management

We are using the Github issues and project dashboard to manage the list of TODOs
for this project:

* [Open issues][gh-issues]
* [Project dashboard][gh-dashboard]

[gh-issues]: https://github.com/kubernetes-sigs/service-apis/issues
[gh-dashboard]: https://github.com/kubernetes-sigs/service-apis/projects/1

Issues labeled `good first issue` and `help wanted` are especially good for a
first contribution.

## Building the code

The project uses `make` to drive the build. You can kick off an overall build
from the top-level makefile:

```shell
make
```

## Testing the code

TODO

## Submitting a review

TODO

## Documentation

The site documentation is written in [mkdocs][mkdocs] format. The files are
contained in `docs-src/`. Generated files are in `docs/` and published to
Github Pages.

Building the docs:

```shell
make -f docs.mk
```

Live preview for editing (view on [http://localhost:8000](), CTRL-C to quit):

```shell
make -f docs.mk serve
```

### Publishing

The docs are published automatically to [Github pages][ghp]. When making changes to the
documentation, generate the new documentation and make the generated code a
self-contained commit (e.g. the changes to `docs/`). This will keep the code
reviews simple and clearly delineate user vs generated content.

[ghp]: https://kubernetes-sigs.github.io/service-apis/
[mkdocs]: https://www.mkdocs.org/
