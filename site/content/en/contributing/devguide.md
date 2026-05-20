---
title: "Developer Guide"
weight: 2
---

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
[semver]: https://semver.org/
[prio-labels]: https://github.com/kubernetes-sigs/gateway-api/labels?q=priority
[issue-cmds]: https://prow.k8s.io/command-help?repo=kubernetes-sigs%2Fgateway-api

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

```bash
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
removed from the go struct, with a tombstone comment ensuring the field name will not be reused.

Example:

```golang
// DeprecatedField is tombstoned to show why 16 is reserved protobuf tag.
// DeprecatedField string `json:"deprecatedField,omitempty" protobuf:"bytes,16,opt,name=deprecatedField"`
```

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

Install a [gateway API implementation](/docs/implementations/list/) and test out the change. Take a look at some
[examples](/guides/).

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
[Hugo](https://gohugo.io/). Each PR will automatically include a
[Netlify](https://netlify.com/) deploy preview. When new code merges, it will
automatically be deployed with Netlify to
[gateway-api.sigs.k8s.io](). If you want to
manually preview docs changes locally, you can install Hugo and run:

```shell
 make docs
```

To make it easier to use the right version of Hugo, you can build and serve the docs in a container:

```shell
$ make build-docs
...
INFO    -  Documentation built in 6.73 seconds
$ make live-docs
...
Web Server is available at //localhost:1313/ (bind address 127.0.0.1)
```

You can then view the docs at http://localhost:1313/.

For more information on how documentation should be written, refer to our
[Documentation Style Guide](/contributing/style-guide/).

### Conformance Tests

To develop or run conformance tests, refer to the [Conformance Test
Documentation](/docs/concepts/conformance/#running-tests).

### Adding new tools
The tools used to build and manage this project are self-contained on their own
directory at the `tools` directory.

To add a new tool, use `go get -tool -modfile tools/go.mod the.tool.repo/toolname@version`
and tidy the specific module with `go mod tidy -modfile=tools/go.mod`.

To execute the new tool, use `go tool -modfile=tools/go.mod toolname`.

## API Documentation

When writing API documentation (the `godoc` for API fields and structures) it should be done
in a meaningful and concise way, where information is provided for the different 
[Gateway API personas](../docs/concepts/roles-and-personas.md#key-roles-and-personas)  without leaking implementation details.

The implementation details are still important for Gateway API implementation
developers, and should still be provided, but we need to ensure they are not 
exposed in the generated CRDs. That can end up leaking to users in multiple ways,
such as on the Gateway API documentation website, or via `kubectl explain`.

Additionally, it is worth noting that API documentation affects the size of generated CRD manifests,
which directly impacts the Kubernetes resource size, 
which is in turn limited by etcd maximum value size.  The situation is exacerbated by the `last-applied-configuration` 
annotation from client-side apply, which roughly doubles the size of the Kubernetes resource.

There are two kind of API documentation: 

* User facing - MUST define how a user should be consuming an API and its fields, in a concise way.
* Developer facing - MUST define how a controller should implement an API and its fields. 

### User facing Documentation

The API documentation, when meaningful, helps users of it understand and create configuration
in a way that ensures that their Gateway API implementation does what they want it to do.

A good API documentation should cover:

* What is the main feature of the API and Field - Eg.: "`foo` allows configuring how a
a header should be forwarded to backends"
* What is the support level of the field - Eg.: "Support: Core/Extended/Implementation Specific"
* Caveats of that field - Eg.: "Setting `foo` field can have conflicts with `bar` field, and in this
case it will be shown as a Condition". (we don't need to cover all the conditions).

In one or two phrases, API documentation should help a user who is reading the documentation understand
what happens when the field is configured, what values can be specified, and what any
additional important implications of setting the field may be. 

When adding a documentation, it is very important to remove your "Developer hat" 
and put yourself in the role of a user who is trying to solve a problem: Does setting a field
solve my needs? How can I use it?

The godoc for an API field contains user-facing documentation. Taking
`Listeners`, one of the most complex fields, as an example:

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

We don't specify here what the Protocol types are (this information goes in the `Protocol` field itself),
how a hostname is defined, or when TLS configuration is required. All of this information
belongs to the respective API fields, which are sub-fields of `Listener`.  Note that `kubectl explain` prints information about both
the specified field and its sub-fields, so when a user executes the `kubectl explain gateway.spec.listeners` command,
they get all of this information.


### Developer facing documentation

Developer-facing documentation helps implementers to define the expected
behavior of it, and should answer questions such as the following:

* How should that field be reconciled?
* What conditions should be set during the reconciliation? 
* What should be validated during the reconciliation of that field?

In this case, as the API documentation serves as a guide for implementers on how 
their implementations should behave, it is very important to be as verbose as
required to avoid any ambiguity. This information is used also to define expected
conformance behavior, and it can point to existing GEPs so a developer looking 
at it can know where to look for more references on what the expected
behavior of this field is and why.

Continuing with the `Listeners` field as an example, it shows how API documentation can provide guidance for situations 
such as the following:

* Two listeners have different protocols but the same hostname. Should this be a conflict?
* A listener of type `XXX` sets the field `TLS`. Is this a problem? How to expose this to 
the user?

Because these details don't matter for a user, they should be hidden from the CRD/OpenAPI
generation and also from the website API Reference.

This can be achieved putting these information between the tags 
`<gateway:util:excludeFromCRD></gateway:util:excludeFromCRD>` and preferably 
contain a callout that those are a Note for implementers:

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
// Notes for implementers:
//
// Setting TLSModeType to Passthrough is only supported on Listeners that are of 
// type HTTP, HTTPS and TLS. In case a user sets a different type, the implementation
// MUST set a condition XXX with value XXX and a message specifying why the condition 
// happened.
// </gateway:util:excludeFromCRD>
Mode *TLSModeType `json:"mode,omitempty"`
```

### Advice when writing the API documentation
When writing the documentation, you should always consider questions like:

**As a user**:

* Does the documentation provide meaningful information and remove any doubt 
about what will happen when setting this field?
* Does the documentation provide information about where should I look if something
goes wrong?
* If I do `kubectl explain` or look into the API Reference, do I have enough
information to achieve my goals without being buried with information I don't care about?

**As a developer/implementor**:

* Does the documentation provide enough information for another developer on 
how they should implement their controller?
* Does the documentation provide enough information on what other fields/resources
should be verified to provide the right behavior?
* Does the documentation provide enough information on how I should signal to the 
users what went right/wrong and how to fix it?

It is important to you consider the personas for which you are writing for, when
working on each type of documentation.
