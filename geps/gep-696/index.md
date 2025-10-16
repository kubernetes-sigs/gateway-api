# GEP-696: GEP template

* Issue: [#696](https://github.com/kubernetes-sigs/gateway-api/issues/696)
* Status: Provisional|Prototyping|Implementable|Experimental|Standard|Completed|Memorandum|Deferred|Declined|Withdrawn

(See [status definitions](../overview.md#gep-states).)

## TLDR

[required_in]: # (Provisional status and above)

(1-2 sentence summary of the proposal)

## Goals

[required_in]: # (Provisional status and above)

(Primary goals of this proposal.)

## Longer Term Goals (optional)

(goals that are not covered initially on this proposal but may be considered long term)

## Non-Goals

[required_in]: # (Provisional status and above)

(What is out of scope for this proposal.)

## Introduction/Overview

[required_in]: # (Provisional status and above)

(Can link to external doc -- but we should bias towards copying
the content into the GEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

Write here "What" we want to do. What is the proposal aiming to do?

## Purpose (Why and Who)

[required_in]: # (Provisional status and above)

Write here "Why" we want to do it. What problems are being solved? What personas are
the target of this proposal, and why will this proposal will make their lives better?

## API

[required_in]: # (Implementable status and above)

(... details, can point to PR with changes)

### Gateway for Ingress (North/South)
(Include API details for North/South use cases)

### Gateway For Mesh (East/West)
(Include East/West API considerations, examples, and if different - APIs)

## Request flow
Example on the usage flow of this proposal/enhancement. It is suggested to contain
at least one manifest as example.

Example of a flow description:

* A client makes a request to https://foo.example.com.
* DNS resolves the name to a `Gateway` address.
* The reverse proxy receives the request on a `Listener` and does something with it
* The reverse proxy passes the request through `XPTORoute` modifying the headers to contain `XYZ`

## Conformance Details

[required_in]: # (Provisional status and above)

(from https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-2162/index.md#standardize-features-and-conformance-tests-names)

#### Feature Names

[required_in]: # (Experimental status and above)

(Does it require separate feature(s) for mesh? Please add them if necessary)

Every feature should:

1. Start with the resource name. i.e HTTPRouteXXX
2. Follow the PascalCase convention. Note that the resource name in the string should come as is and not be converted to PascalCase, i.e HTTPRoutePortRedirect and not HttpRoutePortRedirect.
3. Not exceed 128 characters.
4. Contain only letters and numbers

GEPs cannot move to Experimental without a Feature Name.

### Conformance test scenarios

This section records the scenarios that Conformance tests will check.

It _does not_ need to include code, although code _may_ be used to illustrate the scenarios
if required. Tables are also acceptable for describing complex interactions.

Scenarios _should_ be able to be summarized without code.

These scenario summaries can then be used to determine the names of the tests and their files.

#### Example test scenario 1, please remove: HTTPRoute Simple, Same Namespace

A HTTPRoute with a basic routing configuration, in the same namespace as its
parent Gateway, should route traffic to the specified backend.

#### Example test scenario 2, please remove: HTTPRoute Path Rewrite

A HTTPRoute with a Path Rewrite filter should rewrite the path according to
the specification, routing traffic to the backend.

* A Match of `/prefix/one` with a `ReplacePrefixMatch` of `/one` should route requests
  to `/prefix/one/two` to `/one/two` instead.
* A Match of `/strip-prefix` with a `ReplacePrefixMatch` of `/` should route requests to
  `/strip-prefix/three` to `/three` instead.
* A Match of `/full/one` with a `ReplaceFullPath` of `/one` should route requests to
  `/full/one/two` to `/one` instead.
* ... and so on.

#### Conformance test file names

Conformance tests file names should try to follow the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.

Treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.

In any case, the conformance tests file names should be meaningful and easy to understand.

(Make sure to also include conformance tests that cover mesh)


## `Standard` Graduation Criteria

( This section outlines the criteria required for graduation to Standard. It MUST
contain at least the items in the template, but more MAY be added if necessary. )

( Required for Experimental status and above)

* At least one Feature Name must be listed.
* The `Conformance Details` must be filled out, with conformance test scenarios listed.
* Conformance tests must be implemented that test all the listed test scenarios.
* At least three (3) implementations must have submitted conformance reports that pass
  those conformance tests.
* At least six months must have passed from when the GEP moved to `Experimental`.


## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
