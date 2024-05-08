# Enhancement Requests

Inspired by [Kubernetes enhancements][enhance], Gateway API provides a process for
introducing new functionality or considerable changes to the project. The
enhancement process will evolve over time as the project matures.

[enhance]: https://github.com/kubernetes/enhancements

Enhancements provides the basis of a community roadmap. Enhancements may be
filed by anyone, but require approval from a maintainer to accept the
enhancement into the project.

## What is Considered an Enhancement?

An enhancement is generally anything that:

- Introduces changes to an API.
- Needs significant effort to implement.
- Requires documentation to utilize.
- Impacts how a system is operated including addition or removal of significant
  capabilities.

It is unlikely to require an enhancement if it:

- Fixes a bug
- Adds more testing
- Code refactors

If you're unsure the proposed work requires an enhancement, file an issue
and ask.

## New Enhancement Process

The process for creating new enhancement proposals is documented on the
[GEP Overview][gep] page. Please see that page for all the details about how
to log a new GEP, and the process it will follow on its journey towards
Completed status.

A **documented** discussion of some form **must** exist prior to submitting a
request for enhancement if that enhancement is non-trivial (which we will define
as either: _implicates changes to the API specification_
OR _has some kind of end-user impact_).

Please use our [Github Discussions][discussion] forum as the initial place to
start, and feel free to bring that discussion up for synchronous conversation in
one of our [community meetings][meetings]. If the created request doesn't
include reference to a discussion and/or recordings of discussion in our
community meetings, please note that it _may_ get closed with a request to
create an initial discussion first.

[gep]: /geps/overview
[discussion]: https://github.com/kubernetes-sigs/gateway-api/discussions/new/choose
[meetings]: /contributing/#meetings

## When are Enhancements Accepted?

Gateway API has a predictable release cycle that includes multiple phases. New
enhancements are only considered in the early phases of that release cycle while
the scope of a release is being determined. For more information, refer to our
[release cycle documentation](/contributing/release-cycle).

## Why are Enhancements Tracked

As the project evolves, it's important that the community understands how the
enhancement affects the project.  Individually, it's hard to understand how all
parts of the system interact, but as a community we can work together to build
the right design and approach before getting too deep into an implementation.

## When to Comment on an Enhancement Issue

Please comment on the enhancement issue to:

- Request a review or clarification on the process
- Update status of the enhancement effort
- Link to relevant issues in other repos
