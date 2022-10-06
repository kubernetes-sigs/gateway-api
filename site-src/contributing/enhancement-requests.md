# Enhancement Requests

Inspired by [Kubernetes enhancements][enhance], Gateway API provides a process for
introducing new functionality or considerable changes to the project. The
enhancement process will evolve over time as the project matures.

[enhance]: https://github.com/kubernetes/enhancements

Enhancements provides the basis of a community roadmap. Enhancements may be
filed by anyone, but require approval from a maintainer to accept the
enhancement into the project.

## Quick start

1. Create an [Issue][issue] and select "Enhancement Request".
2. Follow the instructions in the enhancement request template and submit the
   Issue.
3. (depending on size of change) Start a [draft Gateway Enhancement Proposal
   (GEP)][gep]

[issue]: https://github.com/kubernetes-sigs/gateway-api/issues/new/choose
[gep]: /geps/overview

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

## When to Create a New Enhancement

Create an enhancement once you have:

- Circulated your idea to see if there is interest.
- Identified community members who agree to work on and maintain the enhancement.
- Enhancements may take several releases to complete.
- A prototype in your own fork (optional)

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
