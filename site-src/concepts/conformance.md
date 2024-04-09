# Conformance

This API covers a broad set of features and use cases and has been implemented
widely. This combination of both a large feature set and variety of
implementations requires clear conformance definitions and tests to ensure the
API provides a consistent experience wherever it is used.

When considering Gateway API conformance, there are three important concepts:

## 1. Release Channels

Within Gateway API, release channels are used to indicate the stability of a
field or resource. The "standard" channel of the API includes fields and
resources that have graduated to "beta". The "experimental" channel of the API
includes everything in the "standard" channel, along with experimental fields
and resources that may still be changed in breaking ways **or removed
altogether**. For more information on this concept, refer to our
[versioning](/concepts/versioning) documentation.

## 2. Support Levels

Unfortunately some implementations of the API will not be able to support every
feature that has been defined. To address that, the API defines a corresponding
support level for each feature:

* **Core** features will be portable and we expect that there is a reasonable
  roadmap for ALL implementations towards support of APIs in this category.
* **Extended** features are those that are portable but not universally
  supported across implementations. Those implementations that support the
  feature will have the same behavior and semantics. It is expected that some
  number of roadmap features will eventually migrate into the Core. Extended
  features will be part of the API types and schema.
* **Implementation-specific** features are those that are not portable and are
  vendor-specific. Implementation-specific features will not have API types and
  schema except via generic extension points.

Behavior and feature in the Core and Extended set will be defined and validated
via behavior-driven conformance tests. Implementation-specific features will not
be covered by conformance tests.

By including and standardizing Extended features in the API spec, we expect to
be able to converge on portable subsets of the API among implementations without
compromising overall API support. Lack of universal support will not be a
blocker towards developing portable feature sets. Standardizing on spec will
make it easier to eventually graduate to Core when support is widespread.

### Overlapping Support Levels

It is possible for support levels to overlap for a specific field. When this
occurs, the minimum expressed support level should be interpreted. For example,
an identical struct may be embedded in two different places. In one of those
places, the struct is considered to have Core support while the other place only
includes Extended support. Fields within this struct may express separate Core
and Extended support levels, but those levels must not be interpreted as
exceeding the support level of the parent struct they are embedded in.

For a more concrete example, HTTPRoute includes Core support for filters defined
within a Rule and Extended support when defined within BackendRef. Those filters
may separately define support levels for each field. When interpreting
overlapping support levels, the minimum value should be interpreted. That means
if a field has a Core support level but is in a filter attached in a place with
Extended support, the interpreted support level must be Extended.

## 3. Conformance Tests

Gateway API includes a set of conformance tests. These create a series of
Gateways and Routes with the specified GatewayClass, and test that the
implementation matches the API specification.

Each release contains a set of conformance tests, these will continue to
expand as the API evolves. Currently conformance tests cover the majority
of Core capabilities in the standard channel, in addition to some Extended
capabilities.

### Running Tests

There are two main contrasting sets of conformance tests:

* Gateway related tests (can also be thought of as ingress tests)
* Service Mesh related tests

For `Gateway` tests you must enable the `Gateway` test feature, and then
opt-in to any other specific tests you want to run (e.g. `HTTPRoute`). For
Mesh related tests you must enable `Mesh`.

We'll cover each use case separately, but it's also possible to combine these
if your implementation implements both. There are also options which pertain
to the entire test suite regardless of which tests you're running.

#### Gateway Tests

By default `Gateway` oriented conformance tests will expect a GatewayClass
named `gateway-conformance` to be installed in the cluster, and tests will be
run against that. Most often, you'll use a different class, which can be
specified with the `-gateway-class` flag along with the corresponding test
command. Check your instance for the `gateway-class` name to use. You must
also enable `Gateway` support and test support for any `*Routes` your
implementation supports.

The following runs all the tests relevant to `Gateway`, `HTTPRoute`, and
`ReferenceGrant`:

```shell
go test ./conformance -run TestConformance -args \
    --gateway-class=my-gateway-class \
    --supported-features=Gateway,HTTPRoute
```

Other useful flags may be found in [conformance flags][cflags].

[cflags]:https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/flags/flags.go

#### Mesh Tests

Mesh tests can be run by simply enabling the `Mesh` feature:

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh
```

If your mesh also includes ingress support with an API such as `HTTPRoute`, you
can run the relevant tests in the same test run by enabling the `Gateway`
feature and any relevant API features, e.g:

```shell
go test ./conformance -run TestConformance -args --supported-features=Mesh,Gateway,HTTPRoute
```

#### Namespace Labels and Annotations

If labels are needed on namespaces used for testing, you can use the
`-namespace-labels` flag to pass one or more `name=value` labels to set on the
test namespaces. Likewise, `-namespace-annotations` can be used to specify
annotations to be applied to the test namespaces. For mesh testing, this flag
can be used if an implementation requires labels on namespaces that host mesh
workloads, for example, to enable sidecar injection.

As an example, when testing Linkerd, you might run

```shell
go test ./conformance -run TestConformance -args \
   ...
   --namespace-annotations=linkerd.io/inject=enabled
```

so that the test namespaces are correctly injected into the mesh.

#### Excluding Tests

The `Gateway` and `ReferenceGrant` features are enabled by default.
You do not need to explicitly list them using the `-supported-features` flag.
However, if you don't want to run them, you will need to disable them using
the `-exempt-features` flag. For example, to run only the `Mesh` tests,
and nothing else:

```shell
go test ./conformance -run TestConformance -args \
    --supported-features=Mesh \
    --exempt-features=Gateway,ReferenceGrant
```

#### Suite Level Options

When running tests of any kind you may not want the test suite to cleanup the
test resources when it completes (i.e. so that you can inspect the cluster
state in the event of a failure). You can skip cleanup by setting the `--cleanup-base-resources=false`
flag.

It may be helpful (particularly when working on implementing a specific
feature) to run a very specific test by name. This can be done by setting the
`--run-test` flag.

#### Network Policies

In clusters that use [Container Network Interface (CNI) plugins][network_plugins]
which enforce network policies some conformance tests might require custom
[`NetworkPolicy`][netpol] resources to be added to the cluster in order to allow
the traffic to reach the required destinations.

Users are expected to add those required network policies so that their
implementations pass those tests.

[network_plugins]: https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/
[netpol]: https://kubernetes.io/docs/concepts/services-networking/network-policies/

### Conformance Profiles

The Conformance profiles are a tool to group multiple supported features, with the
goal of certifying their support, through the conformance reports. The supported
profiles can be configured with the `--conformance-profiles=PROFILE1,PROFILE2`
flag.

## Conformance Reports

In case the suite has been configured to run a set of conformance profiles, and
all the Conformance reports fields have been properly set according to [this guide][conformance-guide],
the conformance report can be uploaded into [this folder][reports-folder] through
a PR. Follow the [guide][conformance-guide] to have all the details on how to structure
and submit reports for a specific implementation.

[reports-folder]: ./conformance/reports/
[conformance-guide]: ./conformance/reports/README.md

## Contributing to Conformance

Many implementations run conformance tests as part of their full e2e test suite.
Contributing conformance tests means that implementations can share the
investment in test development and ensure that we're providing a consistent
experience.

All code related to conformance lives in the "/conformance" directory of the
project. Test definitions are in "/conformance/tests" with each test including
a pair of files. A YAML file contains the manifests to be applied as part of
running the test. A Go file contains code that confirms that an implementation
handles those manifests appropriately.

Issues related to conformance are [labeled with
"area/conformance"](https://github.com/kubernetes-sigs/gateway-api/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fconformance).
These often cover adding new tests to improve our test coverage or fixing flaws
or limitations in our existing tests.
