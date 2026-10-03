---
title: "GEP-5152: Rename Core and Extended"
---

* Issue: [#5152](https://github.com/kubernetes-sigs/gateway-api/issues/5152)
* Status: Memorandum

## TLDR

Gateway API should rename Core and Extended support levels to Required and Optional to increase clarity.

## Goals

* Increase clarity in documentation and discussions by renaming Core and Extended to Required and Optional

## Non-Goals

* Changing how conformance works
* Changing install channels (Standard and Experimental)

## Introduction/Overview

Currently, Gateway API has three types of support:

* **Core**: The field or behavior is required for a specific resource.  Conformance tests for the top-level feature test these fields and behaviors. Example: Weighted Load Balancing in `HTTPRoute`, the resource and the Conformance feature.
* **Extended**: The field or behavior is optional, but must have a specific behavior if supported. Conformance tests for the feature require the feature to be supported by the implementation. Example: TLS Termination in `TLSRoute`, has the feature name `TLSRouteModeTerminate`, and if an implementation claims to support that feature, then the conformance suite will run tests against the implementation to validate the behavior.
* **Implementation Specific (or ImplementationSpecific)** : The field or behavior is described in the API, but there is no testing for this behavior. Implementations are free to implement it how they wish, if they do. Note that all ImplementationSpecific support is optional, but there are no feature names or conformance tests. (That is the key difference between **Extended** and **ImplementationSpecific**)

Even among Gateway API community members, these levels have sometimes led to confusion (particularly because we also have two installation channels **Standard** and **Experimental**, and many folks find the distinction between **Extended** and **Experimental** confusing).

This naming has been part of Gateway API since the very beginning, and can be seen in some of the earliest work on what became Gateway API:

* [A shared Google Doc][api-design-sketch] outlining details of the API
* [Slides and speaker notes from a talk][kubecon-barcelona] by Bowei Du and Rohit Ramkumar (again, in the section on Portability)
* [Slides from a talk][kubecon-san-diego] by Bowei Du and Chris Luciano (in particular the slides on Portability)

## Purpose (Why and Who)

The aim here is change the language we use so that we do not need to explain what we mean by "Core" and "Extended" any more.

The contention here is that "Required" and "Optional" carry enough meaning by themselves that we can safely use them instead.

As part of making this change, it's also important to change the docs to note that:

**Required** and **Optional** are scoped to particular _resources_ or combinations of resources, since _all resources are optional_ for Gateway API.

Additionally:

**Resources** can be Required or Optional in any specific Conformance Profile.

For example, the GATEWAY-HTTP profile requires the resources:

* GatewayClass
* Gateway
* HTTPRoute
* ReferenceGrant

But has the following Optional features:

* BackendTLSPolicy (the whole resource)
* Optional HTTProute features (a long list)

This is _implied_ in many places in the docs, but we have never been explicit about it:

* Implementations can support either Gateway or Mesh Gateway API, or both.
* Implementations can, if they support Gateway, support any number of combination of Routes, and some behaviors are Core for some Routes when combined with Gateway API.
* Implementations can, if they support Mesh, can support either HTTPRoute, GRPCRoute, or both for those use cases.
* Even parts of the API that are required in most cases (such as ReferenceGrant) are not required in all cases.

As part of doing this rename, we should also clarify this.

The people who will benefit are everyone developing and using Gateway API.

## Implementing the rename

If the community agrees on this rename, we will need to:

- [ ] Update all instances of Core and Extended on the site documentation
- [ ] Update field documentation (this will also include clarifying what resources a feature is Required for.)



## Alternatives

We don't rename, and leave things how they are, accepting a cost in confusion.

## References

[api-design-sketch]: https://docs.google.com/document/d/1BxYbDovMwnEqe8lj8JwHo8YxHAt3oC7ezhlFsG_tyag/edit?tab=t.0
[kubecon-san-diego]: https://static.sched.com/hosted_files/kccncna19/a5/Kubecon%20San%20Diego%202019%20-%20Evolving%20the%20Kubernetes%20Ingress%20APIs%20to%20GA%20and%20Beyond%20%5BPUBLIC%5D.pdf
[kubecon-barcelona]: https://static.sched.com/hosted_files/kccnceu19/97/%5Bwith%20speaker%20notes%5D%20Kubecon%20EU%202019_%20Ingress%20V2%20%26%20Multi-Cluster%20Services.pdf
