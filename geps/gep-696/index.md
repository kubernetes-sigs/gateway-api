# GEP-696: GEP template

* Issue: [#696](https://github.com/kubernetes-sigs/gateway-api/issues/696)
* Status: Provisional|Prototyping|Implementable|Experimental|Standard|Completed|Memorandum|Deferred|Declined|Withdrawn

(See [status definitions](../overview.md#gep-states).)

## TLDR

(1-2 sentence summary of the proposal)

## Goals

(Primary goals of this proposal.)

## Longer Term Goals (optional)

(goals that are not covered initially on this proposal but may be considered long term)

## Non-Goals

(What is out of scope for this proposal.)

## Introduction/Overview

(Can link to external doc -- but we should bias towards copying
the content into the GEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

Write here "What" we want to do. What is the proposal aiming to do?

## Purpose (Why and Who)

Write here "Why" we want to do it. What problems are being solved? What personas are
the target of this proposal, and why will this proposal will make their lives better?

## API
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

(from https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-2162/index.md#standardize-features-and-conformance-tests-names)

#### Feature Names

(Does it require separate feature(s) for mesh? Please add them if necessary)

Every feature should:

1. Start with the resource name. i.e HTTPRouteXXX
2. Follow the PascalCase convention. Note that the resource name in the string should come as is and not be converted to PascalCase, i.e HTTPRoutePortRedirect and not HttpRoutePortRedirect.
3. Not exceed 128 characters.
4. Contain only letters and numbers

GEPs cannot move to Experimental without a Feature Name.

### Conformance tests 

Conformance tests file names should try to follow the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.

Treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.

In any case, the conformance tests file names should be meaningful and easy to understand.

(Make sure to also include conformance tests that cover mesh)

When describing the new feature, write down some conformance test scenarios the feature should manage,
to guarantee that future implementors understand what "Conformance" means and what will be tested.

At least _some_ tests should be added at each phase, starting with Provisional.

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
