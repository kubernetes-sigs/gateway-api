# GEP-696: GEP template

* Issue: [#696](https://github.com/kubernetes-sigs/gateway-api/issues/696)
* Status: Provisional|Implementable|Experimental|Standard|Deferred|Rejected|Withdrawn|Replaced

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

(1-2 sentence summary of the proposal)

## Goals

(Primary goals of this proposal.)

## Non-Goals

(What is out of scope for this proposal.)

## Introduction

(Can link to external doc -- but we should bias towards copying
the content into the GEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

## API
(... details, can point to PR with changes)

### Gateway for Ingress (North/South)
(Include API details for North/South usecases)

### Gateway For Mesh (East/West)
(Include East/West API considerations, examples, and if different - APIs)

## Conformance Details

(from https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-2162/index.md#standardize-features-and-conformance-tests-names)

#### Feature Names

(Does it require separate feature(s) for mesh? Please add them if necessary)

Every feature should:

1. Start with the resource name. i.e HTTPRouteXXX
2. Follow the PascalCase convention. Note that the resource name in the string should come as is and not be converted to PascalCase, i.e HTTPRoutePortRedirect and not HttpRoutePortRedirect.
3. Not exceed 128 characters.
4. Contain only letters and numbers

### Conformance tests 

Conformance tests file names should try to follow the the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.

Treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.

In any case, the conformance tests file names should be meaningful and easy to understand.

(Make sure to also include conformance tests that cover mesh)

## Alternatives

(List other design alternatives and why we did not go in that
direction)

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
