# GEP-820: Drop extension points from Route matches

* Issue: [#820](https://github.com/kubernetes-sigs/gateway-api/issues/820)
* Status: Standard

## TLDR

Drop extension points within Route match block. These extension points are
not well understood.

## Goals

- Drop the extension points within Route match block.

## Non-Goals

- Figure out a replacement solution for the use-case that these extension
  points addressed

## Introduction

As the API moves towards `v1alpha2`, the maintainers intend to make the API
standard and forward compatible for the foreseeable future. To that end,
maintainers intend to minimize (eliminate if possible) breaking changes post
`v1alpha2`. This GEP is part of that initiative.

Extension points for match criteria in Routes were added to enable use-cases
where match criteria defined by implementation was a super-set of match
criteria defined within this API. To the best of our knowledge, even though
extension points were added, no concrete examples or use-cases were known at
that time and none have been discovered so far.

This proposal advocates removal of these extension points because:

- It goes against the unwritten design principles this API has followed so far: 

  - minimize number of API types as much as possible
  - minimize strongly coupled API types and instead shoot for self-contained
    types.
  - extension points are introduced with clear use-cases and possibilities in
    mind. Vague extension points are avoided as they become harder to maintain.
- It is unlikely that the user experience resulting from defining two k8s
  resources for defining a Route will be optimal.
- There is not prior art on splitting match criteria and backend forwarding
  semantics (spec.backends) in the community. We believe they are kept together
  for good reasons.

## API

The following fields and all associated documentation will be removed:

- HTTPRouteMatch.ExtensionRef
- TCPRouteMatch.ExtensionRef will be removed. This results in a struct without
  any members: TCPRouteMatch. The struct will be kept as it is expected that
  more match criterias might be added to L4 routes.

  - Do the same to UDPRoute and TLSRoute

## Alternatives

N/A

