# Developer Guide

## Project Management

We are using the GitHub issues and project dashboard to manage the list of TODOs
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
[gh-dashboard]: https://github.com/kubernetes-sigs/gateway-api/projects
[gh-milestones]: https://github.com/kubernetes-sigs/gateway-api/milestones
[semver]:https://semver.org/
[prio-labels]:https://github.com/kubernetes-sigs/gateway-api/labels?q=priority
[issue-cmds]:https://prow.k8s.io/command-help?repo=kubernetes-sigs%2Fgateway-api

## Prerequisites

Before you start developing with Gateway API, we'd recommend having the
following prerequisites installed:

* [KinD](https://kubernetes.io/docs/tasks/tools/#kind): This is a standalone local Kubernetes cluster. At least one container runtime is required.
* [Docker](https://docs.docker.com/engine/install/): This is a prerequisite for running KinD. While you can opt for alternatives like [Podman](https://podman.io/docs/installation), please be aware that doing so is at your own risk.
* [BuildX](https://github.com/docker/buildx): Prerequisite for `make verify` to run.
* [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl): This is the Kubernetes command-line tool.
* [Go](https://golang.org/doc/install): It is the main programming language in this project. Please check this [file](https://github.com/kubernetes-sigs/gateway-api/blob/main/go.mod#L3) to find out the least `Go` version otherwise you might encounter compilation errors.
* [Digest::SHA](https://metacpan.org/pod/Digest::SHA): It is a required dependency. You can obtain it by installing the `perl-Digest-SHA` package.


## Development: Building, Deploying, Testing, and Verifying

Clone the repo:

```
mkdir -p $GOPATH/src/sigs.k8s.io
cd $GOPATH/src/sigs.k8s.io
git clone https://github.com/kubernetes-sigs/gateway-api
cd gateway-api
```

This project works with Go modules; you can choose to setup your environment
outside $GOPATH as well.


### Build the Code

The project uses `make` to drive the build. `make` will clean up previously generated code, run code generators, and
run static analysis against the code and generate Kubernetes CRDs. You can kick
off an overall build from the top-level makefile:

```shell
make generate
```


#### Add Experimental Fields

All additions to the API must start in the Experimental release channel.
Experimental fields must be marked with the `<gateway:experimental>` annotation
in Go type definitions. Gateway API CRD generation will only include these
fields in the experimental set of CRDs.

If experimental fields are removed or renamed, the original field name should be
removed from the go struct, with a tombstone comment
([example](https://github.com/kubernetes/kubernetes/blob/707b8b6efd1691b84095c9f995f2c259244e276c/staging/src/k8s.io/api/core/v1/types.go#L4444-L4445))
ensuring the field name will not be reused.

### Deploy the Code

Use the following command to deploy CRDs to the preexisting `Kind` cluster.

```shell
make crd
```

Use the following command to check if the CRDs have been deployed.

```shell
kubectl get crds
```

### Test Manually

Install a [gateway API implementation](../implementations.md) and test out the change. Take a look at some
[examples](../guides/index.md).

### Verify

Make sure you run the static analysis over the repo before submitting your
changes. The [Prow presubmit][prow-setup] will not let your change merge if
verification fails.

```shell
make verify
```

[prow-setup]: https://github.com/kubernetes/test-infra/tree/master/config/jobs/kubernetes-sigs/gateway-api


## Post-Development: Pull Request, Documentation, and more Tests
### Submit a Pull Request

Gateway API follows a similar pull request process as
[Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/guide/pull-requests.md).
Merging a pull request requires the following steps to be completed before the
pull request will be merged automatically.

- [Sign the CLA](https://git.k8s.io/community/CLA.md) (prerequisite)
- [Open a pull request](https://help.github.com/articles/about-pull-requests/)
- Pass [verification](#verify) tests
- Get all necessary approvals from reviewers and code owners


### Documentation

The site documentation is written in Markdown and compiled with
[mkdocs](https://www.mkdocs.org/). Each PR will automatically include a
[Netlify](https://netlify.com/) deploy preview. When new code merges, it will
automatically be deployed with Netlify to
[gateway-api.sigs.k8s.io](). If you want to
manually preview docs changes locally, you can install mkdocs and run:

```shell
 make docs
```

To make it easier to use the right version of mkdocs, you can build and serve the docs in a container:

```shell
$ make build-docs
...
INFO    -  Documentation built in 6.73 seconds
$ make live-docs
...
INFO    -  [15:16:59] Serving on http://0.0.0.0:3000/
```

You can then view the docs at http://localhost:3000/.

For more information on how documentation should be written, refer to our
[Documentation Style Guide](style-guide.md).

### Conformance Tests

To develop or run conformance tests, refer to the [Conformance Test
Documentation](../concepts/conformance.md#running-tests).

### Adding new tools
The tools used to build and manage this project are self-contained on their own
directory at the `tools` directory.

To add a new tool, use `go get -tool -modfile tools/go.mod the.tool.repo/toolname@version`
and tidy the specific module with `go mod tidy -modfile=tools/go.mod`.

To execute the new tool, use `go tool -modfile=tools/go.mod toolname`.

## API Documentation

When writing API documentation (the `godoc` for API fields and structures) it should be done
 in a meaningful and concise way, where information are provided for the different 
 Gateway API personas (Ian, Chihiro and Ana) without leaking implementation details.

The implementation details are still important for a Gateway API implementation
developer, and they should still be provided but without being exposed on the
CRD generation, that can end leaking to users on a diverse set of ways, like
on Gateway API documentation website, or via `kubectl explain`.

Additionally, it is worth noticing that API documentation reflects on the CRD generation
size, which impacts directly on resource consumption like a maximum Kubernetes resource size 
(which is limited by etcd maximum value size) and avoiding problems with `last-applied-configuration` 
annotation, when doing a client-side apply.

There are two kind of API documentations: 

* User facing - MUST define how a user should be consuming an API and its field, on a concise way.
* Developer facing - MUST define how a controller should implement an API and its fields. 

### User facing Documentation

The API documentation, when meaningful, helps users of it on doing proper configuration
in a way that Gateway API controllers react and configure the proxies the right way.

A good API documentation should cover:

* What is the main feature of the API and Field - Eg.: "`Foo` allows configuring how a
a header should be forwarded to backends"
* What is the support level of the field - Eg.: "Support: Core/Extended/Implementation Specific"
* Caveats of that field - Eg.: "Setting `Foo` field can have conflicts with `Bar` field, and in this
case it will be shown as a Condition". (we don't need to cover all the conditions).

In a simple way, a user reading the field documentation should understand, on one or two 
phrases what happens when the field is configured, what can be configured and what are 
the impacts of that field

When adding a documentation, it is very important to remove your "Developer hat" 
and put yourself on a user that is trying to solve a problem: Does setting a field
solves my needs? How can I use it?

On an implementation, a user facing documentation belongs to the field documentation. Taking
`Listeners`, one of the most complex fields as an example:

```golang
// Listeners define logical endpoints that are bound on this Gateway's addresses.
// At least one Listener MUST be specified. When setting a Listener, conflicts can
// happen depending on its configuration like protocol, hostname and port, and in 
// this case a status condition will be added representing what was the conflict.
// 
// The definition of a Listener protocol implies what kind of Route can be attached 
// to it
Listeners []Listener `json:"listeners"`
```

We don't specify what are the Protocol types (saving this to the `Protocol` field),
what a hostname means, when a TLS configuration is required. All of these information
belongs to each field, so when a user does something like `kubectl explain gateway.spec.listeners`
they will also get the information of each field.

### Developer facing documentation

Developer facing documentation helps during implementations to define the expected
behavior of it, and should answer questions like:

* How that field should be reconciled?
* What conditions should be set during the reconciliation? 
* What should be validated during the reconciliation of that field?

In this case, as the API documentation serves as a guide for implementors on how 
their implementations should behave, it is very important to be as much verbose as
required to avoid any ambiguity. These information are used also to define expected
conformance behavior, and can even point to existing GEPs so a developer looking 
at it can know where to look for more references on what and why are those the expected
behavior of this field.

Still taking the `Listeners` field as an example, it does good definitions of situations 
like:

* Two listeners have different protocols but the same hostname. Should this be a conflict?
* A listener of type `XXX` sets the field `TLS`. Is this a problem? How to expose this to 
the user?

Because these information don't matter for a user, they should be hidden from the CRD/OpenAPI
generation and also from the website API Reference.

This can be achieved putting these information between the tags 
`<gateway:util:excludeFromCRD></gateway:util:excludeFromCRD>` and preferably 
contain a callout that those are a Note for implementors:

```golang
// Mode defines the TLS behavior for the TLS session initiated by the client.
// There are two possible modes:
//
// - Terminate: The TLS session between the downstream client and the
//   Gateway is terminated at the Gateway. This mode requires certificates
//   to be specified in some way, such as populating the certificateRefs
//   field.
// - Passthrough: The TLS session is NOT terminated by the Gateway. This
//   implies that the Gateway can't decipher the TLS stream except for
//   the ClientHello message of the TLS protocol. The certificateRefs field
//   is ignored in this mode.
//
// Support: Core
//
// <gateway:util:excludeFromCRD>
// Notes for implementors:
//
// Setting TLSModeType to Passthrough is only supported on Listeners that are of 
// type HTTP, HTTPS and TLS. In case a user sets a different type, the implementation
// MUST set a condition XXX with value XXX and a message specifying why the condition 
// happened.
// </gateway:util:excludeFromCRD>
Mode *TLSModeType `json:"mode,omitempty"`
```

### Advices when writing the API documentation
As an advice, the person writing the documentation should always being questioning:

**As a user**:

* Does the documentation provide meaningful information and removes any doubt 
about what will happen when setting this field?
* Does the documentation provide information about where should I look if something
goes wrong?
* If I do `kubectl explain` or look into the API Reference, do I have enough
information to achieve my goals without being buried with information I don't care?

**As a developer/implementor**:

* Does the documentation provide enough information for another developer on 
how they should implement their controller?
* Does the documentation provide enough information on what other fields/resources
should be verified to provide the right behavior?
* Does the documentation provide enough information on how I should signal to the 
users what went right/wrong and how to fix it?

It is important to exercise changing the personas for which you are writing the 
documentation.