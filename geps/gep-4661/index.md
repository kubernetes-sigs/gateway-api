# GEP-4661: In-Cluster: Provisioned service scope and optimizations

* Issue: [#4661](https://github.com/kubernetes-sigs/gateway-api/issues/4661)
* Status: Provisional

## TLDR (What)

This GEP enables Gateway owners to portably select the Kubernetes Service type provisioned by an in-cluster Gateway implementation, 
and establishes production-ready defaults for each service type so that common best practices are applied automatically.

Concretely, this GEP has two goals:

* Allow users to specify the scope of a service provisioned by an `In-Cluster` implementation, whether the provisioned Service should be of type `ClusterIP` or `LoadBalancer`.
* Define normative requirements for each service type so that implementations ship with optimal defaults (e.g. `externalTrafficPolicy`, `healthCheckNodePort`)

### Goals

* Introduce a new `AddressType` value that allows Gateway owners to portably
  provision a Gateway scoped for in-cluster-only visibility (ClusterIP)
* Introduce a new `AddressType` value that allows Gateway owners to opt in to
  production-ready LoadBalancer defaults (e.g. `externalTrafficPolicy: Local`) without changing the behavior of existing Gateways
* Define normative requirements per `AddressType` so implementations ship
  consistent, well-defined behavior for each service scope
* Allow users to specify a `LoadBalancerClass` when provisioning a
  LoadBalancer-backed Gateway

### Non-Goals

* Replicate the full Kubernetes Service API inside Gateway API
* Change existing LoadBalancer provisioning behavior — optimized defaults are
  opt-in via new address types, not retroactively applied
* Replace `infrastructure.parametersRef` for implementation-specific Service
  customization
* Support `ExternalName` Service types

## Motivation (Why)

[GEP-1762](https://gateway-api.sigs.k8s.io/geps/gep-1762/) established the foundation for in-cluster Gateway deployments and acknowledged that Service type matters — its [Gateway IP](https://gateway-api.sigs.k8s.io/geps/gep-1762/#gateway-ip) section references both `ClusterIP` and `LoadBalancer` services — but did not provide a portable mechanism to choose between them. Instead, this was deferred to "arbitrary customization" via `infrastructure.parametersRef` ([GEP-1867](https://gateway-api.sigs.k8s.io/geps/gep-1867/)).

In practice, this means that every implementation has solved service type selection differently — through custom annotations, implementation-specific parameters, or other ad-hoc mechanisms. 
This reproduces the same fragmentation that Gateway API was designed to eliminate: users must learn each implementation's particular approach for what is fundamentally a portable concern.

By promoting service type selection into the Gateway API itself, this GEP gives users a single, declarative way to express their intent. Implementations can then apply well-defined defaults for each service type, ensuring that a newly provisioned Gateway is production-ready without requiring additional configuration.

This GEP does not aim to replicate the full Kubernetes Service API. The scope is deliberately narrow: service type selection and normative defaults for the most impactful fields. Additional Service-level customization remains available through `infrastructure.parametersRef` for implementation-specific needs.

## Who

This GEP benefits Chihiro, the cluster operator as they:

* need to choose the right service type for their workload without learning implementation-specific configuration.
* want consistent, production-ready defaults across Gateway deployments in their clusters.

### Use Cases

* A Gateway owner provisions a Gateway for
  [inference extension](https://gateway-api-inference-extension.sigs.k8s.io/) and wants it reachable only within the cluster. Today, making the provisioned Service a `ClusterIP` requires implementation-specific knowledge. With this GEP, the owner can express this intent portably.
* A Gateway owner provisions a Gateway exposed via `LoadBalancer` and expects production-ready traffic routing out of the box — with `externalTrafficPolicy` set to `Local` to preserve client source IP and avoid unnecessary cross-node hops.
* A cluster operator migrating from implementation A to implementation B expects their Gateway manifests to work without modification, because the service type and scope are expressed portably rather than through implementation-specific annotations.


## API

**TODO**: First PR will not include any implementation details, in favor of
building consensus on the motivation, goals and non-goals first. _"How?"_ we
implement shall be left open-ended until _"What?"_ and _"Why?"_ are solid.

## References

