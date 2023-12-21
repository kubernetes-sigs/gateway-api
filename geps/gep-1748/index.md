# GEP-1748: Gateway API Interaction with Multi-Cluster Services

* Issue: [#1748](https://github.com/kubernetes-sigs/gateway-api/issues/1748)
* Status: Experimental

???+ Prolonged Experimental Phase
    This GEP will be in the "Experimental" channel for a prolonged period of
    time. This explores the interaction of Gateway API with the Multi-Cluster
    Services API. Until the MCS API is also GA, it will be impossible for this
    GEP to graduate beyond "Experimental".

    This GEP is also exempt from the [Probationary Period][expprob] rules as it
    predated them.

[expprob]:https://gateway-api.sigs.k8s.io/geps/overview/#probationary-period

## TLDR

The Kubernetes Multi-Cluster Services API enables Services to span multiple
clusters. Gateway API enables advanced traffic routing, serving as the next
generation Ingress, Load Balancing, and Mesh API. This document describes how
these APIs can be used together.

## Goals

* Define the interaction between Gateway API and Multi-Cluster Services
* Define any situations where Gateway API may span multiple clusters without the
  Multi-Cluster Services API

## Non-Goals

* Make significant changes to either API

## Introduction

Gateway API and the Multi-Cluster Services API represent two of the newest
Kubernetes networking APIs. As they’ve been developed in parallel, there’s been
some cross-project discussion about how they can interact, but that has never
formally been written down. This GEP aims to change that.

## Overview

Multi-Cluster Services can be used within Gateway API wherever Services can be
used. The difference is that Services refer only to cluster-local endpoints while
Multi-Cluster Services can refer to endpoints throughout an entire ClusterSet.

### ServiceImport as a Backend

A Route can forward traffic to the endpoints attached to an imported Service.
This behaves identically to how forwarding to Services work in Kubernetes, with
the exception that the endpoints attached to a ServiceImport may span multiple
clusters. For example, the following HTTPRoute would forward traffic to
endpoints attached to the "store" ServiceImport:

```yaml
{% include 'standard/multicluster/httproute-simple.yaml' %}
```

#### Routing to Specific Clusters

In some cases, it may be helpful to route certain paths to a specific cluster
(or region). Similar to single-cluster Services, this can be accomplished by
creating multiple Multi-Cluster Services, one for each subset of endpoints you
would like to route to. For example, the following configuration will send
requests with paths prefixed with “/west” to the store-west ServiceImport, and
“/east” to the store-east ServiceImport. Requests that don’t match either of
these paths will be routed to the “store” ServiceImport which represents a
superset of the “store-west” and “store-east” ServiceImports.

```yaml
{% include 'standard/multicluster/httproute-location.yaml' %}
```

#### Advanced Routing With ServiceImports

All Routing capabilities in Gateway API should apply equally whether the backend
is a Service or ServiceImport. For example, when routing to a system with
multiple read replicas, it could be beneficial to route requests based on HTTP
Method. In the following example, requests with POST, PUT, and DELETE methods
are routed to `api-primary` while the rest are routed to `api-replicas`:

```yaml
{% include 'standard/multicluster/httproute-method.yaml' %}
```

#### Routing to Both Services and ServiceImports

There are some situations where it will be useful to split traffic between a
Service and ServiceImport. In the following example, 90% of traffic would go to
endpoints attached to the cluster-local "store" Service, and the remaining 10%
would go to endpoints attached to the Multi-Cluster "store-global" Service:

```yaml
{% include 'standard/multicluster/httproute-hybrid.yaml' %}
```

#### Cross-Namespace References with ReferenceGrant

It is possible to use ReferenceGrant to enable cross-namespace references to
ServiceImports. For example, the following HTTPRoute would forward traffic to
endpoints attached to the “bar” Multi-Cluster Service in the “bar” namespace:

```yaml
{% include 'standard/multicluster/httproute-referencegrant.yaml' %}
```

### Mesh: ServiceImport as Parent

In some cases, you may want to override traffic destined for a Multi-Cluster
Service with a mesh. As part of the broader GAMMA initiative, ServiceImport can
be used in the same way that Service is used as a ParentRef. When a Service is
specified as a parent, meshes will intercept traffic destined for the ClusterIP
of the Service and apply any policies or routing decisions defined by the Route.
Similarly, when a ServiceImport is specified as a parent, meshes will intercept
traffic destined for the ClusterSetIP and apply any policies or routing
decisions defined by the Route. In the following example, the mesh would
intercept traffic destined for the store ClusterSetIP matching the `/cart` path
and redirect it to the `cart` Multi-Cluster Service.

```yaml
{% include 'standard/multicluster/httproute-gamma.yaml' %}
```

### Services vs ServiceImports

It is important that all implementations provide a consistent experience. That
means that references to Services SHOULD always be interpreted as references to
endpoints within the same cluster for that Service. References to ServiceImports
MUST be interpreted as routing to Multi-Cluster endpoints across the ClusterSet
for the given ServiceImport. In practice, that means that users should use
“Service” when they want to reference cluster-local endpoints, and
“ServiceImport” when they want to route to Multi-Cluster endpoints across the
ClusterSet for the given ServiceImport. This behavior should be analogous to
using `.cluster.local` versus `.clusterset.local` DNS for a given Service.

## API Changes

* ServiceImport is recognized as backend with “Extended” conformance
* ServiceImport is included in GAMMA GEP(s) with “Extended” conformance
* Clarification that Services refer to endpoints within the same cluster

## Alternatives

### Develop Custom Multi-Cluster Concepts Independently from Multi-Cluster Services

We could theoretically develop an entirely new way to handle multi-cluster routing. We’re choosing not to do that because the existing APIs are sound and can work well together.

## References

* [Original Doc](https://docs.google.com/document/d/1akwzBKtMKkkUV8tX-O7tPcI4BPMOLZ-gmS7Iz-7AOQE/edit#)
