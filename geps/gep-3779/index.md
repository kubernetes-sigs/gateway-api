# GEP-3779: Identity Based Authz for East-West Traffic

* Issue: [#3779](https://github.com/kubernetes-sigs/gateway-api/issues/3779)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)


## TLDR
 


## Goals

(Using the [Gateway API Personas](../../concepts/roles-and-personas.md))

* A way for Ana the Application Developer to configure a Gateway API implementation to perform identity based authorization that **allows** or **denies** the requests for some K8s clients to the K8s workloads.

* A way for Ana, the Application Developer, to configure a Gateway API implementation to perform identity based authorization that **allows** or **denies** the requests from some K8s clients to all the K8s workload in some namespace.

## Non-Goals

* Supporting identity based authorization for north-south traffic.


## Introduction

An identity-based authorization API is essential because it provides a structured way to control access to network traffic based on client identities within a Kubernetes cluster, a capability particularly vital for enforcing fine-grained security policies in complex multi-tenant or large-scale environments.

All the open source meshes have their own implementaition of idenity based authorization and it is now important use case for Gateway APIs for east-west traffic.

### State of the World

Here are the examples of some of the service meshes. 

* Istio
Istio [authorization policy] (https://istio.io/latest/docs/reference/config/security/authorization-policy/) provides a way to validate the request based on client identities derived from peer certificate used in mTLS. Users can apply to Kubernetes pod labels. This same API is used in Istio's Ambient Mesh as well.

* Linkerd
Linkerd [authorization policy] (https://linkerd.io/2-edge/reference/authorization-policy/) also provides a way to validate the request based on client identities. Linkerd also provides an option to pick the peer identity from the client certs used in mTLS.

* Cilium
[TODO] Add more details ...

* Kuma
[TODO] Add more details ...


## Outstanding Questions and Concerns (TODO)


## API



## Conformance Details


#### Feature Names


### Conformance tests 


## Alternatives


## References