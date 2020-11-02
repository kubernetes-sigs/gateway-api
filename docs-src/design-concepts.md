# API Design Concepts

This page describes some of the concepts and ideas that have shaped the API design for users and implementors of Gateway. This is the "why" behind why different decisions were made to hopefully provide clarity about different aspects of the API's design. 

### Single resource consistency

The Kubernetes API guarantees consistency only on a single resource level. There
are a couple of consequences for complex resource graphs as opposed to single
resources:

*   Error checking of properties spanning multiple resource will be asynchronous
    and eventually consistent. Simple syntax checks will be possible at the
    single resource level, but cross resource dependencies will need to be
    handled by the controller.
*   Controllers will need to handle broken links between resources and/or
    mismatched configuration.

### Conflicts

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

### Conformance

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

#### Conformance expectations

We expect there will be varying levels of conformance among the
different providers in the early days of this API. Users can use the
results of the conformance tests to understand areas where there may
be differences in behavior from the spec.

### Extension points

A number of extension points are provided in the API to provide flexibility in
addressing the large number of use-cases that cannot be addressed by a general
purpose API.

Here is a summary of extension points in the API:

- **XRouteMatch.ExtensionRef**: This extension point should be used to extend
  the match semantics of a specific core Route. This is an experimental
  extension point and will be iterated on in future based on feedback.
- **XForwardTo.BackendRef**: This extension point should be used for forwarding
  traffic to network endpoints other than core Kubernetes Service resource.
  Examples include an S3 bucket, Lambda function, a file-server, etc.
- **HTTPRouteFilter**: This API type in HTTPoute provides a way to hook into the
  request/response lifecycle of an HTTP request.
- **Custom Routes**: If none of the above extensions points suffice for a use
  case, Implementers can chose to create custom Route resources for protocols
  that are not currently supported in the API.

Whenever you are using an extension point without any prior art, please let
the community know. As we learn more about usage of extension points, we would
like to find the common denominators and promote the features to core/extended
API conformance.