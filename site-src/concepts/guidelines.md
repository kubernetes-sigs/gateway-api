# Implementation guidelines

There are some general design guidelines used throughout this API.

!!! note
    Throughout the Gateway API documentation and specification,
    keywords such as "MUST", "MAY", and "SHOULD" are used
    broadly. These should be interpreted as described in RFC 2119.

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

## Gracefully Handling Future API Versions

An important consideration when implementing this API is how it might change in
the future. Similar to the Ingress API before it, this API is designed to be
implemented by a variety of different products within the same cluster. That
means that the API version your implementation was developed with may be
different than the API version it is used with. At a minimum, the following
requirements must be met to ensure future versions of the API do not break your
implementation:

* Handle fields with loosened validation without crashing
* Handle fields that have transitioned from required to optional without
  crashing

## Limitations of CRD and Webhook Validation

CRD and webhook validation is not the final validation i.e. webhook is "nice UX"
but not schema enforcement. This validation is intended to provide immediate
feedback to users when they provide an invalid configuration. Write code
defensively with the assumption that at least some invalid input (Gateway API
resources) will reach your controller. Both Webhook and CRD validation is not
fully reliable because it:

* May not be deployed correctly.
* May be loosened in future API releases. (Fields may contain values with less
  restrictive validation in newer versions of the API). 

*Note: These limitations are not unique to Gateway API and apply more broadly to
any Kubernetes CRDs and webhooks.*

Implementers should ensure that, even if unexpected values are encountered in
the API, their implementations are still as secure as possible and handle this
input gracefully. The most common response would be to reject the configuration
as malformed and signal the user via a condition in the status block. To avoid
duplicating work, Gateway API maintainers are considering adding a shared
validation package that implementations can use for this purpose. This is
tracked by [#926](https://github.com/kubernetes-sigs/gateway-api/issues/926).

### Expectations

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


## Kind vs. Resource

Similar to other Kubernetes APIs, Gateway API uses "Kind" instead of "Resource"
in object references throughout the API. This pattern should be familiar to
most Kubernetes users.

Per the [Kubernetes API conventions][1], this means that all implementations of
this API should have a predefined mapping between kinds and resources. Relying
on dynamic resource mapping is not safe.

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

[1]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md
