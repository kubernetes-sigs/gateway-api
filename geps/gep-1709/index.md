# GEP-1709: Conformance Profiles

* Issue: [#1709](https://github.com/kubernetes-sigs/gateway-api/issues/1709)
* Status: Standard

## Important Notes

Unlike our APIs, CRDs, or `gwctl` conformance profiles and their reports are
not user facing in the traditional sense. As trends in Gateway API utilization
grow and change, it may make sense to adjust how profiles work to align better
with what implementations are doing. As such conformance profiles (that is,
the test suite and the reports) might be subject to backwards incompatible
changes at times in ways the rest of the API wouldn't be (e.g. significant
features may need to be added or removed on a profile).

## TLDR

Add high level profiles for conformance tests which implementations can select
when running the conformance test suite. Also add opt-in automated conformance
reporting to the conformance test suite to report conformance results back to
the Gateway API project and receive recognition (e.g. badges).

## Goals

- Add high-level profiles which downstream implementations can subscribe to in
  order to run tests for the associated supported feature sets.
- Add a reporting mechanism where conformance results can be reported back to
  the Gateway API project and provide "badges" to visibly decorate the
  implementations as conformant according to their profiles.
- Expand conformance testing documentation significantly so that it becomes the
  "landing zone" for new prospective implementations and provides a clear and
  iterative process for how to get started implementing Gateway API support.

## Non-Goals

- We want to avoid adding any infrastructure for the reporting mechanism if
  feasible.
- For this iteration we don't want to add configuration files for conformance
  tests, instead leaving that to future iterations and working on the raw
  machinery here (see [alternatives considered](#alternatives-considered)).
- For this iteration we don't want to add container images for conformance test
  runs, instead leaving that to future iterations (see
  [alternatives considered](#alternatives-considered).

## Introduction

Since our conformance test suite was conceived of it's been our desire to
provide simple high level profiles that downstream implementations can
subscribe to.

Today we have `SupportedFeatures` which get us some of what we want in terms of
easily configuring the conformance test suite, but in this GEP we will describe
taking that a step further (and a level higher) to create named profiles which
indicate a "level of conformance" which implementations can prove they satisfy
and receive certification for.

An API will be provided as the format for conformance test reports. We will
provide tooling to assist with the reporting and certification process of
submitting those reports and displaying the results.

## API

The API for conformance profiles will include an API resource called
`ConformanceReport` which will be at the center of a workflow that
implementations can opt into to generate and submit those resources.

The workflow implementers will follow will include the following high-level
steps:

1. select a [profile](#profiles)
2. [integrate](#integration) tests in the downstream project
3. [report results and get certified](#certification)

The goal is to make selecting a conformance profile as simple and automatic of a
process as feasible and support both the existing command line integration
approach (e.g. `go test`) as well as a [Golang][go] approach using the
conformance suite as a library.

[go]:https://go.dev

### Profiles

"Profiles" are effectively categories which represent the high-level grouping of
tests related to some feature (or feature set) of Gateway API. When conformance
is reported using one of these profiles extra features can be covered according
to support levels:

- `core`
- `extended`

> **NOTE**: `implementation-specific` doesn't really have much in the way of
> tests today, but it is something users want to be able to display. We leave
> door open for it later and mention it in the [alternatives
> considered](#alternatives-considered) section below.

We will start with the following named profiles:

- `HTTP`
- `TLSPassthrough`

These profiles correspond with `*Route` type APIs that we currently have tests
for. As the tests roll in, we'll also eventually have:

- `UDP`
- `TCP`
- `GRPC`

> **NOTE**: In time we may have higher level groupings, like `Layer4` (which
> would include at least `TCP` and `UDP`) but feedback from the community has
> been strong for a preference on the `*Route` level (see the
> [alternatives considered](#alternatives-considered) for some more notes on
> this) for the moment.

> **NOTE**: APIs that are referenced to or by `*Route` APIs will be tested as
> a part of a profile. For instance, running the `HTTP` profile will also run
> tests for `GatewayClass` and `Gateway` implicitly as these are required
> components of supporting `HTTP`.

The technical implementation of these profiles is very simple: effectively a
"profile" is a static compilation of existing [SupportedFeatures][feat] which
represent the named category. Features that aren't covered under a "core" level
of support are opt-in.

[feat]:https://github.com/kubernetes-sigs/gateway-api/blob/c61097edaa3b1fad29721e787fee4b02c35e3103/conformance/utils/suite/suite.go#L33

### Integration

Integrating the test suite into your implementation can be done using one of
the following methods:

- The [go test][go-test] command line interface which enables projects of any
  language to run the test suite on any platform [Golang][go] supports.
- Using the conformance test suite as a [Golang library][lib] within an already
  existing test suite.

> **NOTE**: Usage as a library is already an established colloquialism in the
> community, this effort simply intends to make that more official.

Conformance profiles are passed as arguments when running the test suite. For
instance when running via command line:

```console
$ go test ./conformance/... -args -gateway-class=acme -conformance-profile=Layer7
```

Or the equivalent configuration using the Golang library:

```go
cSuite, err := suite.New(suite.Options{
    GatewayClassName: "acme",
    Profiles: sets.New(Layer7),
    // other options
})
require.NoError(t, err, "misconfigured conformance test suite")
cSuite.Setup(t)

for i := 0; i < len(tests.ConformanceTests); i++ {
    test := tests.ConformanceTests[i]
    test.Run(t, cSuite)
}
```

> **NOTE**: In the `suite.Options` above it's still possible to add `SkipTests`
> but when used in conjunction with `Profile` this will result in a report that
> the profile is not valid for reporting. Implementations in this state may be
> able to report themselves as "in progress", see the
> [certification section](#certification) for details.

Alternatively for an `Extended` conformance profile where not all of the
features are implemented (as described in the [profiles](#profiles) section
above):

```console
$ go test ./conformance/... -args \
    -gateway-class=acme \
    -conformance-profiles=HTTP,TCP \
    -unsupported-features=HTTPResponseHeaderModification,HTTPRouteMethodMatching,HTTPRouteQueryParamMatching,
```

Or the equivalent configuration using the Golang library:

```go
cSuite, err := suite.New(suite.Options{
    GatewayClassName: "acme",
    Profiles: sets.New(
        HTTP,
        TCP,
    ),
    UnsupportedFeatures: sets.New(
        suite.SupportHTTPResponseHeaderModification,
        suite.SupportHTTPRouteMethodMatching,
        suite.SupportHTTPRouteQueryParamMatching,
    ),
    // other options
})
require.NoError(t, err, "misconfigured conformance test suite")
cSuite.Setup(t)

for i := 0; i < len(tests.ConformanceTests); i++ {
    test := tests.ConformanceTests[i]
    test.Run(t, cSuite)
}
```

> **NOTE**: You can't disable features that are `Core` conformance as `Core` is
> a minimum requirement for the profile to be considered fulfilled.

Some implementations may support more or less extended features than others,
so in some cases it could be cumbersome to have to list ALL features that you
_don't_ support so we optionally and inversely allow `SupportedFeatures` so
you can pick which option makes sense to you, and under the hood the
expressions will compile to the same overall list:

```go
cSuite, err := suite.New(suite.Options{
    GatewayClassName: "acme",
    Profiles: sets.New(
        HTTP,
        TCP,
    ),
    SupportedFeatures: sets.New(
        suite.SupportHTTPRouteMethodMatching,
    ),
    // other options
})
```

> **NOTE**: The `UnsupportedFeatures` and `SupportedFeatures` fields are
> mutually exclusive.

So to have your YAML report include details about extended features you support
you must either opt-in using `SupportedFeatures` to the exact features you
support, or opt-out of the features you _don't_ support using
`UnsupportedFeatures`.

Once an implementation has integrated with the conformance test suite, they can
move on to [certification](#certification) to report the results.

[go-test]:https://go.dev/doc/tutorial/add-a-test
[go]:https://go.dev
[lib]:https://pkg.go.dev/sigs.k8s.io/gateway-api@v0.6.2/conformance/utils/suite

### Implementation mode

The certification process runs against an implementation using a specific mode,
specified in the final report's "mode" field. A "mode" is intended to capture situations
where a Gateway API implementation may have different features and capabilities
depending on how it is deployed (e.g. an implementation might be deployed with an
external, OR internal load balancer, and have different capabilities depending on
the mode chosen). The modes are implementation-specific, and no upstream mode is
defined, except for `default`, which is used in case no mode is specified.

### Gateway API version and channel

The certification is related to a specific API version and a specific channel,
therefore such information must be included in the final report. At test suite
setup time, the conformance profile machinery gets all the CRDs with the field
`.spec.group` equal to `gateway.networking.k8s.io`, and for each of them checks
the annotations `gateway.networking.k8s.io/bundle-version` and
`gateway.networking.k8s.io/channel`. If there are `CRD`s with different
versions, the certification fails specifying that it's not possible to run the
tests as there are different Gateway API versions installed in the cluster. If
there are CRDs with different channels, the certification fails specifying that
it's not possible to run the tests as there are different Gateway API channels
installed in the cluster. If all the Gateway API `CRD`s have the same version
and the same channel, the tests can be run and the detected version and channel
will be set in the `gatewayAPIVersion` and `gatewayAPIChannel` fields of
the final report. Furthermore, the suite must run all the experimental tests when
the channel is `experimental`, and the related features are enabled.

In addition to the `CRD`s version, the suite needs to check its version in
relation to the `CRD`s one. To do so, a new `.go` file containing the current
Gateway API version is introduced in the project and compiled with the
conformance profile suite:

```go
const GatewayAPIVersion = "0.7.0"
```

At test suite setup time the conformance profile suite checks the `CRD`s version
and the suite version; if the two versions differ, the certification fails. A
new generator will be introduced in the project to generate the aforementioned
`.go` file starting from a VERSION file contained in the root folder. Such a
VERSION file contains the semver of the latest release and is manually bumped at
release time. The script hack/verify-all.sh will be updated to ensure the
generated `.go` file is up to date with the VERSION file.

Implementations will be able to report their conformance testing results using
our [reporting process](#reporting-process). Implementations will be able to
visibly demonstrate their conformance results on their downstream projects and
repositories using our [certification process](#certification-process).

## Reporting process

When conformance tests are executed an argument can be provided to the test
suite to emit `ConformanceReport` resource with the test results. This resource
can be configured to emit to the test output itself, or to a specific file.

The following is an example report:

```yaml
apiVersion: v1alpha1
kind: ConformanceReport
implementation:
  organization: acme
  project: operator
  url: https://acme.com
  contact:
  - @acme/maintainers
  version: v1.0.0
date: "2023-02-28 20:29:41+00:00"
gatewayAPIVersion: v0.8.0
gatewayAPIChannel: experimental
mode: default
profiles:
- name: http
  core:
    result: success
    summary: "all core functionality passed"
    statistics:
      passed: 20
      skipped: 0
      failed: 0
  extended:
    result: success
    summary: "all extended features supported"
    statistics:
      passed: 8
      skipped: 0
      failed: 0
    supportedFeatures:
    - ExtendedFeature1
    - ExtendedFeature2
    - ExtendedFeature3
    - ExtendedFeature4
    - ExtendedFeature5
- name: tcp
  core:
    result: success
    summary: "all core functionality passed"
    statistics:
      passed: 4
      skipped: 0
      failed: 0
  extended:
    result: skipped
    summary: "some extended features supported"
    statistics:
      passed: 2
      skipped: 0
      failed: 0
    supportedFeatures:
    - ExtendedFeature1
    - ExtendedFeature2
    unsupportedFeatures:
    - ExtendedFeature3
```

> **WARNING**: It is an important clarification that this is NOT a full
> Kubernetes API. It uses `TypeMeta` for some fields that made sense to re-use
> and were familiar, but otherwise has it's own structure. It is not a [Custom
> Resource Definition (CRD)][crd] nor will it be made available along with our
> CRDs. It will be used only by conformance test tooling.

> **NOTE**: The `implementation` field in the above example includes an
> `organization` and `project` field. Organizations can be an open source
> organization, an individual, a company, e.t.c.. Organizations can
> theoretically have multiple projects and should submit separate reports for
> each of them.

> **NOTE**: The `contact` field indicates the Github usernames or team
> names of those who are responsible for maintaining this file, so they can be
> easily contacted when needed (e.g. for relevant release announcements
> regarding conformance, e.t.c.). Optionally, it can be an email address or
> a support URL (e.g. Github new issue page).

The above report describes an implementation that just released `v1`, uses gateway
API `v0.8.0` `experimental` channel, and has `HTTP` `core` and `extended` and `TCP`
core and partial `extended` support in `default` mode.

### Multiple reports

Multiple `ConformanceReports` can be stored as a list of reports for a specific
Gateway API version. The following shows previous releases of the `acme`/`operator`
implementation and its feature progression:

```yaml
apiVersion: v1alpha1
kind: ConformanceReport
implementation:
  organization: acme
  project: operator
  url: https://acme.com
  contact:
  - @acme/maintainers
  version: v0.9.0
date: "2023-02-28 20:29:41+00:00"
gatewayAPIVersion: v0.8.0
gatewayAPIChannel: standard
mode: default
profiles:
- name: http
  core:
    result: success
    summary: "some core functionality passed"
    statistics:
      passed: 18
      skipped: 2
      failed: 0
    skippedTests:
    - CoreTest8
    - CoreTest15
  extended:
    result: success
    summary: "some extended features supported"
    statistics:
      passed: 8
      skipped: 0
      failed: 0
    supportedFeatures:
    - ExtendedFeature1
    - ExtendedFeature2
    unsupportedFeatures:
    - ExtendedFeature3
    - ExtendedFeature4
    - ExtendedFeature5
```

```yaml
apiVersion: v1alpha1
kind: ConformanceReport
implementation:
  organization: acme
  project: operator
  url: https://acme.com
  contact:
  - @acme/maintainers
  version: v1.0.0
date: "2023-06-1 20:29:41+00:00"
gatewayAPIVersion: v1.0.0
gatewayAPIChannel: experimental
mode: default
profiles:
- name: http
  core:
    result: success
    summary: "all core functionality passed"
    statistics:
      passed: 20
      skipped: 0
      failed: 0
  extended:
    result: success
    summary: "all extended features supported"
    statistics:
      passed: 8
      skipped: 0
      failed: 0
    supportedFeatures:
    - ExtendedFeature1
    - ExtendedFeature2
    - ExtendedFeature3
    unsupportedFeatures:
    - ExtendedFeature4
    - ExtendedFeature5
- name: tcp
  core:
    result: success
    summary: "all core functionality passed"
    statistics:
      passed: 4
      skipped: 0
      failed: 0
  extended:
    result: skipped
    summary: "some extended features supported"
    statistics:
      passed: 2
      skipped: 0
      failed: 0
    supportedFeatures:
    - ExtendedFeature1
    - ExtendedFeature2
    unsupportedFeatures:
    - ExtendedFeature3
```

> **NOTE**: In the above you can see the `acme` implementation's progression. In
> their release `v0.9.0` they had started adding `HTTP` support and added the
> conformance tests to CI, but they were still skipping some core tests. In
> their next release `v1.0.0` they completed adding `HTTP` `Core`
> functionality (and even added three extended features), and also added
> `TCP` functionality with `Core` and partial `Extended` support.

> **NOTE**: Implementations **MUST** report for a specific release version
> (e.g. `v0.7.1`) and not use branches or Git SHAs. Some exceptions will be
> made for initial reports to help make it easier for implementations to get
> started, but as we move to standard everyone should be reporting on specific
> releases.

Creating a pull request to add the `ConformanceReport` will start the
[certification process](#certification-process).

> **NOTE**: No verification process (to prevent intentionally incorrect
> conformance results) will be implemented at this time. We expect that this wont
> be an issue in our community and even if someone were to try and "cheat" on
> the reporting the reputation loss for being caught would make them look very
> bad and would not be worth it.

[crd]:https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/

### Reports submission

Implementers can create as many reports as they want, with different API versions,
API channels, modes, and implementation versions. The reports can be submitted upstream
by creating a pull request to the Gateway API repository. The reports folder tree
is the following:

```console
|-- conformance/reports
|   |-- v1.0
|   |   |-- acme-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v2.13-default-report.yaml
|   |-- v1.1
|   |   |-- acme-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v2.14-default-report.yaml
|   |   |   |-- standard-v2.14-with-the-lot-report.yaml
|   |   |   |-- extended-v2.14-with-the-lot-report.yaml
|   |   |-- umbrella-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v1.8-default-report.yaml
```

The main folder `conformance/reports` contains a set of sub-folders, each for a
Gateway API version. Implementors will create their project folder in all the Gateway
API version folders for which they have a supported implementation. The implementors
folder contains the following set of files:

* `README.md`: contains general information about the implementation, a table of
  contents, and the instructions to reproduce the claimed report.
* `<API-channel>-<implementation-version>-<mode-name>-report.yaml`: the report
  generated by the conformance test suite for the specific mode.

#### README.md structure

The README.md files SHOULD be structured in the following way (enrichment from
implementors is allowed, even though the minimum content is described below):

```md
# Acme operator

General information about the Acme/operator project

## Table of contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
|             |                        |      |        |
|             |                        |      |        |
|             |                        |      |        |

## To reproduce

Instructions on how to reproduce the claimed report.
```

Each row of the table of contents MUST correspond to an uploaded report and contains
the most implementation version and the mode. An example follows:

| API channel  | Implementation version | Mode         | Report                                        |
|--------------|------------------------|--------------|-----------------------------------------------|
| standard     | v2.15                  | default      | ./standard-v2.15-default-report.yaml            |
| standard     | v2.16                  | default      | ./standard-v2.16-default-report.yaml            |
| experimental | v2.16                  | with-the-lot | ./experimental-v2.16-with-the-lot-report.yaml |

> **NOTE**: the column `Report` contains the link to the corresponding uploaded report.

> **NOTE**: This table of contents could be improved in the future with a link to
> proof of conformance, such as artifacts containing logs.

## Certification Process

The raw report data of the `ConformanceReports` _is predominantly meant for machine
consumption_. Report data will be compiled into human-friendly displays during an
automated certification process.

Certification starts with the pull request described during the [reporting
process](#reporting-process). Once the `ConformanceReport` is created or
updated a display layer in the implementations page will need to be updated to
point to the new data.

> **NOTE**: The `ConformanceReport` API will be defined in Golang like our
> other `apis/` so that we can utilize build tags from kubebuilder for defaults
> and validation, and so that there exists a common Golang type for it in the
> conformance test suite. When PRs are created the Gateway API repositories'
> CI will run linting and validation against the reports.

Maintainers will provide [badges][bdg] to implementations at the end of the process
which link to the implementations page for that specific implementation and can
be easily added via markdown to Git repositories.

[impl]:https://gateway-api.sigs.k8s.io/implementations/
[bdg]:https://shields.io

## Alternatives Considered

### Conformance Test Configuration File

Conformance testing is currently done mainly through command line with
`go test` or via use of the conformance test suite as a library in Golang
projects. We considered whether adding the alternative to provide a
configuration file to the test suite would be nice, but we haven't heard
any specific demand for that from implementors yet so we're leaving the
idea here for later iterations to consider.

### Conformance Test Container Image

Providing a container image which could be fed deployment instructions for an
implementation was considered while working on this GET but it seemed like a
scope all unto itself so we're avoiding it here and perhaps a later iteration
could take a look at that if there are asks in the community.

### Implementation-Specific Reporting

Users have mentioned the desire to report on `implementation-specific` features
they support as a part of conformance. At the time of writing, there's not much
in the way of structure or testing for us to do this with but we remain open to
the idea. The door is left open in the `ConformanceReport` API for a future
iteration to add this if desired, but it probably warrants its own GEP as we
need to make sure we have buy-in from multiple stakeholders with different
implementations that are implementing those features.

### High-Level Profiles

We originally started with two high-level profiles:

* `Layer4`
* `Layer7`

However, the overwhelming feedback from the community was to go a step down and
define profiles at the level of each individual API (e.g. `HTTPRoute`,
`TCPRoute`, `GRPCRoute`, e.t.c.). One of the main reasons for this was that we
already have multiple known implementations of Gateway API which only support
a single route type (`UDPRoute`, in particular as it turns out).

We may consider in the future doing some of these higher-level profiles if
there's a technical reason or strong desire from implementers.

## Graduation Criteria

The following are items that **MUST** be resolved to move this GEP to
`Standard` status (and before the end of the probationary period):

- [x] some kind of basic level of display for the report data needs to exist.
  It's OK for a more robust display layer to be part of a follow-up effort.
  - for now we ended up with badges in our implementations page. We have
    [another effort](https://github.com/kubernetes-sigs/gateway-api/issues/2550)
    underway to build an even better display layer, but we consider this
    additive and the current display is sufficient for moving the project to
    standard.
- [x] initially we were OK with storing reports in the Git repository as files.
  While this is probably sufficient during the `Experimental` phase, we need to
  re-evaluate this before `Standard` and see if this remains sufficient or if
  we want to store the data elsewhere.
  - During the experimental phase this has not caused any significant issues,
    so for the purposes of calling this standard we're going to move forward
    as-is. It should be straightforward and reasonable to change the storage
    mechanism later as needs arise.
- [x] We have been actively [gathering feedback from SIG
  Arch][sig-arch-feedback]. Some time during the `experimental` phase needs to
  be allowed to continue to engage with SIG Arch and incorporate their feedback
  into the test suite.
  - SIG Arch did not have any significant feedback during the experimental
    phase. In the same timespan the Network Policy group has started using our
    test suite and APIs as well, so it seems the overall approach has obvious
    merit, and doesn't appear to be redundant.
- [x] Finalize the report organization structure based on feedback during the
  experimental phase.
- [x] Base documentation must exist for implementations to run the tests via
  the CLI and via the Golang library.

[sig-arch-feedback]:https://groups.google.com/g/kubernetes-sig-architecture/c/YjrVZ4NJQiA/m/7Qg7ScddBwAJ

## References

- https://github.com/kubernetes-sigs/gateway-api/issues/1709
- https://github.com/kubernetes-sigs/gateway-api/issues/1329
