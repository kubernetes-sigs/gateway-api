---
title: "Conformance"
weight: 30
---
This API covers a broad set of features and use cases and has been implemented
widely. This combination of both a large feature set and variety of
implementations requires clear conformance definitions and tests to ensure the
API provides a consistent experience wherever it is used.

## Conformance Concepts

When considering Gateway API conformance, there are three important concepts:

### Release Channels

Within Gateway API, release channels are used to indicate the stability of a
field or resource. The "standard" channel of the API includes fields and
resources that have graduated to "stable". The "experimental" channel of the API
includes everything in the "standard" channel, along with experimental fields
and resources that may still be changed in breaking ways **or removed
altogether**. For more information on this concept, refer to our
[versioning](/docs/concepts/versioning/) documentation.

### Support Levels

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

Behavior and feature in the Core and Extended set are defined and validated
via behavior-driven conformance tests. Implementation-specific features are not
be covered by conformance tests.

{{% alert color="primary" title="Overlapping Support Levels" %}}

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

{{% /alert %}}

### Conformance Tests and Conformance Reports

Gateway API includes a set of conformance tests. These create a series of
Gateways and Routes with the specified GatewayClass, and test that the
implementation matches the API specification.

Each release contains a suite of conformance tests, along with a harness to
run them. Currently conformance tests cover the majority of capabilities in
the Standard channel, for both Core and Extended behaviors for each supported
resource and traffic direction.

## What conformance actually means

Conformance is intended to provide publicly available, reproducible reports that
users and the public can use to verify that implementations support the features
that they claim to.

To this end, implementations are expected to run the conformance test suite, take
the output, and submit the conformance report to the Gateway API repo. Tooling will
then apply conformance rules and update lists of conformant implementations, feature
matrices and other information automatically.

In this way, users can be confident that, when an implementation says it supports
a particular version of Gateway API (and optionally a set of Extended features
associated with that version), it actually does.

The last benefit that conformance gives is that it ensures that behavior for
standardized behaviors is portable across conformant implementations.


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
"area/conformance-test"](https://github.com/kubernetes-sigs/gateway-api/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fconformance-test).
These often cover adding new tests to improve our test coverage or fixing flaws
or limitations in our existing tests.

