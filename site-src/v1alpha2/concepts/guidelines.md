# Implementation guidelines

There are some general design guidelines used throughout this API.

## Single resource consistency

The Kubernetes API guarantees consistency only on a single resource level. There
are a couple of consequences for complex resource graphs as opposed to single
resources:

*   Error checking of properties spanning multiple resource will be asynchronous
    and eventually consistent. Simple syntax checks will be possible at the
    single resource level, but cross resource dependencies will need to be
    handled by the controller.
*   Controllers will need to handle broken links between resources and/or
    mismatched configuration.

## Conflicts

Separation and delegation of responsibility among independent actors (e.g
between cluster ops and application developers) can result in conflicts in the
configuration. For example, two application teams may inadvertently submit
configuration for the same HTTP path.

In most cases, guidance for conflict resolution is provided along with the
documentation for fields that may have a conflict. If a conflict does not have a
prescribed resolution, the following guiding principles should be applied:

* Prefer not to break things that are working.
* Drop as little traffic as possible.
* Provide a consistent experience when conflicts occur.
* Make it clear which path has been chosen when a conflict has been identified.
  Where possible, this should be communicated by setting appropriate status
  conditions on relevant resources.
* More specific matches should be given precedence over less specific ones.
* The resource with the oldest creation timestamp wins.
* If everything else is equivalent (including creation timestamp), precedences
  should be given to the resource appearing first in alphabetical order
  (namespace/name). For example, foo/bar would be given precedence over foo/baz.

## Conformance

As this API aims to cover a wide set of implementations and use cases,
it will not be possible for all implementations to support *all*
features at the present. However, we do expect the set of features
supported to converge eventually. For a given feature, users will be
guaranteed that features in the API will be portable between providers
if the feature is supported.

To model this in the API, we are taking a similar approach as with
[sig-arch][sig-arch-bdd] work on conformance profiles. Features as
described in the API spec will be divided into three major categories:

[sig-arch-bdd]: https://github.com/kubernetes/enhancements/tree/master/keps/sig-architecture/960-conformance-behaviors

* **CORE** features will be portable and we expect that there is a
  reasonable roadmap for ALL implementations towards support of APIs
  in this category.
* **EXTENDED** features are those that are portable but not
  universally supported across implementations. Those implementations
  that support the feature will have the same behavior and
  semantics. It is expected that some number of EXTENDED features will
  eventually migrate into the CORE. EXTENDED features will be part of
  the API types and schema.
* **CUSTOM** features are those that are not portable and are
  vendor-specific. CUSTOM features will not have API types and schema
  except via generic extension points.

Behavior and feature in the CORE and EXTENDED set will be defined and
validated via behavior-driven conformance tests. CUSTOM features will
not be covered by conformance tests.

By including and standardizing EXTENDED features in the API spec, we
expect to be able to converge on portable subsets of the API among
implementations without compromising overall API support. Lack of
universal support will not be a blocker towards developing portable
feature sets. Standardizing on spec will make it easier to eventually
graduate to CORE when support is widespread.

### Conformance expectations

We expect there will be varying levels of conformance among the
different providers in the early days of this API. Users can use the
results of the conformance tests to understand areas where there may
be differences in behavior from the spec.

### Implementation-specific

In some aspects of the API, we give the user an ability to specify usage of the
feature, however, the exact behavior may depend on the underlying
implementation. For example, regular expression matching is present in all
implementations but specifying an exact behavior is impossible due to
subtle differences between the underlying libraries used (e.g. PCRE, ECMA,
Re2). It is still useful for our users to spec out the feature as much as
possible, but we acknowledge that the behavior for some subset of the API may
still vary (and that's ok).

These cases will be specified as defining delimited parts of the API
"implementation-specific".

The "implementation-specific" designation allows a CORE or EXTENDED feature to
be well-defined taking into account the realities of some features that are
mostly but not entirely portable.

## API Conventions

Service APIs follow Kubernetes API [conventions][1]. These conventions
are intended to ease client development and ensure that configuration
mechanisms can consistently be implemented across a diverse set of use
cases. One variation from the Kubernetes conventions is for fields with
default values. These fields are considered optional but are not pointers
since they will never be empty.

[1]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md

### Conformance Tests

Conformance tests are actively being developed to ensure that implementations of
this API are conformant with the spec. Use `make conformance` to run these tests
with the Kubernetes cluster you are currently connected to. These tests require
the `CONTROLLER_NAME` environment variable to be set. This will be used to set
the `controller` field on GatewayClass resources created for the conformance
tests.

Conformance tests are defined with scenarios. Each scenario includes:

* A YAML file with Service APIs resource manifests.
* A Go test file describing what a controller should do with those resources.

These tests are currently in an alpha state. Please file a GitHub issue or ask
in Slack if these are not working as expected.
