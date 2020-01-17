<!--
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
-->

# Enhancements Tracking and Backlog

Inspired by the [Kubernetes enhancement](https://github.com/kubernetes/enhancements)
process, provide a mechanism to discuss and reach consensus for introducing new
functionality to the service-api's project.

Enhancements may take multiple releases to complete and thus provide
the basis of a community roadmap.  Enhancements may be filed by anyone in the
community, but require project maintainers for acceptance to the project.

## Quick start

1. Socialize an idea with community members.
2. Follow the process outlined in the
[enhancement template](https://github.com/kubernetes-sigs/service-apis/tree/master/enhancements/template.md).

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

Create an enhancement once you:

- have circulated your idea to see if there is interest
- have done a prototype in your own fork (optionally)
- have identified community members who agree to work on and maintain the
enhancement
  - enhancements may take several releases to complete  

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
