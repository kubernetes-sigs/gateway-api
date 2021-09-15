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

### Overlapping Support Levels
It is possible for support levels to overlap. When this occurs, the minimum
expressed support level should be interpreted. For example, an identical struct
may be embedded in two different places. In one of those places, the struct is
considered to have CORE support while the other place only includes EXTENDED
support. Fields within this struct may express separate CORE and EXTENDED
support levels, but those levels may never be interpreted as exceeding the
support level of the parent struct they are embedded in.

For a more concrete example, HTTPRoute includes CORE support for filters defined
within a Rule and EXTENDED support when defined within ForwardTo. Those filters
may separately define support levels for each field. When interpreting
overlapping support levels, the minimum value should be interpreted. That means
if a field has a CORE support level but is in a filter attached in a place with
EXTENDED support, the interpreted support level should be EXTENDED.

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

Gateway API follows Kubernetes API [conventions][1]. These conventions
are intended to ease client development and ensure that configuration
mechanisms can consistently be implemented across a diverse set of use
cases. In addition to the Kubernetes API conventions, Gateway API has the
following conventions:

### List Names

Another convention this project uses is for plural field names for lists
in our CRDs. We use the following rules:

- If the field name is a noun, use a plural value.
- If the field name is a verb, use a singular value.

So for example, in HTTPRoute, `hostnames` uses a plural, but `forwardTo` is singular,
although they are both lists.

[1]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
