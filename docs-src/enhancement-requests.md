# Enhancement Tracking and Backlog

Inspired by [Kubernetes enhancements](https://github.com/kubernetes/enhancements), service-api's
provides a process for introducing new functionality or considerable changes to the project. The
enhancement process will evolve over time as the project matures.

Enhancements provides the basis of a community roadmap. Enhancements may be filed by anyone, but
require approval from a maintainer to accept the enhancement into the project.

## Quick start

1. Create an [Issue](https://github.com/kubernetes-sigs/gateway-api/issues/new/choose) and select
"Enhancement Request".
2. Follow the instructions in the enhancement request template and submit the Issue.

## What is Considered an Enhancement?

An enhancement is generally anything that:

- impacts how a cluster is operated including addition or removal of significant
  capabilities
- introduces changes to an api
- needs significant effort to implement
- requires documentation to utilize

It is unlikely to require an enhancement if it:

- fixes a bug
- adds more testing
- code refactors
- minimal impact to a release

If you're unsure the proposed work requires an enhancement, file an issue
and ask.

## When to Create a New Enhancement

Create an enhancement once you have:

- circulated your idea to see if there is interest.
- identified community members who agree to work on and maintain the enhancement.
- enhancements may take several releases to complete.
- a prototype in your own fork (optional)


## Why are Enhancements Tracked

As the project evolves, it's important that the service-api's community understands
how the enhancement affects the project.  Individually, it's hard to understand how
all parts of the system interact, but as a community we can work together to build
the right design and approach before getting too deep into an implementation.

## When to Comment on an Enhancement Issue

Please comment on the enhancement issue to:

- request a review or clarification on the process
- update status of the enhancement effort
- link to relevant issues in other repos
