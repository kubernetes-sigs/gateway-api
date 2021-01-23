# Releases

Although Service APIs are an official Kubernetes project, and represent official
APIs, these APIs will not be installed by default on Kubernetes clusters at this
time. This project will use Custom Resource Definitions (CRDs) to represent the
new API types that Service APIs include.

Similar to other Kubernetes APIs, these will go through a formal Kubernetes
Enhancement Proposal (KEP) review. Unlike other Kubernetes APIs, Service API
releases will be independent from Kubernetes releases initially.

Service API releases will include four components:

* Custom Resource Definitions to define the API.
* Go client libraries.
* Validation webhooks to implement cross field validations.
* Conversion webhooks to convert resources between API versions.

## Versioning

This project uses 2 different but related forms of versioning:

- [Kubernetes API Versioning] (example: v1alpha1)
- [Semantic Versioning] (example: v0.1.0)

Each new API version will be released with a new semantic version. For example,
v1alpha1 was released with v0.1.0. Before we release an API version, we may
provide some release candidates such as v0.1.0-rc1 as a way to evaluate a new
API version before it is formally released. All releases will be compatible with
[Go modules versioning].

As the API evolves, we will make intermediate releases that improve upon an
existing API version. These releases will be fully backwards compatible and will
be limited to either bug fixes or additions to the API.

This project may release one or more additional alpha API versions. New alpha
API versions may include breaking changes such as removing or renaming fields or
resources.

Following [Semantic Versioning], new patch releases will be limited to bug
fixes. New minor releases may include new fields or resources in addition to bug
fixes. New API versions will be released with new minor or major versions.

Our changelog and release notes will always include both the semantic version
and API version(s) included in the release. 

[Kubernetes API Versioning]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api_changes.md#alpha-beta-and-stable-versions
[Semantic Versioning]: https://semver.org/
[Go modules versioning]: https://golang.org/ref/mod#versions

> The first release candidate was tagged as `v1alpha1-rc1`. It predates this
documentation and is an exception.

## Branching
This project will have frequent updates to the master branch. There are no
compatibility guarantees associated with code in any branch, including master,
until it has been released. For example, changes may be reverted before a
release is published. For the best results, use the latest published release of
this project.

## Installation

This project will be responsible for providing straightforward and reliable ways
to install releases of Service APIs.

## Other Official Custom Resources

This is a relatively new concept, and there is only one previous example of
official custom resources being used:
[VolumeSnapshots](https://kubernetes.io/blog/2018/10/09/introducing-volume-snapshot-alpha-for-kubernetes/).
Although VolumeSnapshot CRDs can be installed directly by CSI drivers that
support them, Service APIs must support multiple controllers per cluster, so the
CRDs will live in and be installed from this repo.
