# GEP-3798: Client IP-Based Session Persistence

* Issue: [#3798](https://github.com/kubernetes-sigs/gateway-api/issues/3798)
* Status: Deferred

(See [status definitions](../overview.md#gep-states).)

## Notes and Disclaimers

* **DEFERRED**: This originally targeted release as `Experimental` in [v1.4.0].
  Notably (in [PR#3844]) there was concern that it may be difficult to get
  multiple implementations to support this. During the release cycle, this GEP
  was not able to meet the timeline requirements to progress, so it is now
  considered deferred. If anyone is interested in picking this back up, it will
  need to be re-submitted for consideration in a future release with a written
  plan about how it will achieve implementation from multiple implementations.
  If this remains in deferred state for a prolonged period, it may eventually
  be moved to `Withdrawn`, or moved into the alternatives considered for
  [GEP-1619].

[v1.4.0]:https://github.com/kubernetes-sigs/gateway-api/milestone/22
[PR#3844]:https://github.com/kubernetes-sigs/gateway-api/pull/3844
[GEP-1619]:https://gateway-api.sigs.k8s.io/geps/gep-1619/

## TLDR

### What
 This GEP proposes the addition of Client IP-based session persistence to the Gateway API. This feature will allow Gateway API implementations to ensure that requests originating from a specific client IP address (or a subnet defined by an IP mask) are consistently routed to the same backend endpoint for a configurable duration. This aims to provide a standardized and centralized mechanism for client IP persistence across various Gateway API implementations.

 As mentioned in the [GEP-1619](https://gateway-api.sigs.k8s.io/geps/gep-1619/#api), `SessionPersistence` can be applied via `BackendLBPolicy` and `RouteRule` API. Similar [edge case behaviour](https://gateway-api.sigs.k8s.io/geps/gep-1619/#edge-case-behavior) and [API Granularity](https://gateway-api.sigs.k8s.io/geps/gep-1619/#api-granularity) for ClientIP Persistence type should be applicable as well.

 An important addition/difference compared to [GEP-1619](https://gateway-api.sigs.k8s.io/geps/gep-1619) is that the identity of the backend assigned to a client (or a group of clients in the same subnet) is stored on the server (load balancer / gateway) side as opposed to the client side.

## Goals

* Define an API extension within Gateway API to enable client IP-based session persistence.

* Allow configuration of a session duration for which a client IP should stick to a backend.

* Provide an optional mechanism to specify an IP mask for subnet-based persistence, allowing multiple clients from the same subnet to be routed to the same backend.

* Improve portability of applications requiring client IP persistence across different Gateway API implementations.

## Non-Goals

This GEP does not dictate the specific algorithm or implementation details for how an individual Gateway controller maintains the client IP-to-backend mapping (e.g., in-memory, distributed cache).

## Introduction

### Why: The Problem This Solves
Currently, achieving client IP-based session persistence within Kubernetes Gateway API environments often requires vendor-specific annotations or out-of-band configurations on the underlying load balancer or ingress controller. This approach has several drawbacks:

* Inconsistent User Experience: Users have to learn different methods for configuring the same logical feature depending on their chosen Gateway API implementation. Configurations are not easily transferable between different Gateway API implementations, leading to vendor lock-in and increased operational overhead when migrating or using multiple controllers.

* Reduced Visibility: The desired session persistence behavior is not explicitly declared within the Gateway API resources, making it harder to audit, manage, and understand the routing logic from a single source of truth.

This GEP addresses these issues by providing a first-class API mechanism for client IP-based session persistence, enhancing the Gateway API's capabilities and promoting consistency and portability.

### Who: Beneficiaries
* Application Developers: Can define session persistence requirements directly in their Gateway API configurations, ensuring consistent behavior regardless of the underlying Gateway implementation. This simplifies application deployment and management for stateful workloads.

* Platform Operators/Administrators: Gain a standardized way to configure and manage client IP-based session persistence across their clusters, reducing the need for custom scripts or deep knowledge of individual controller implementations. This improves operational efficiency and consistency.

* Gateway API Implementers: Receive a clear specification for implementing client IP-based session persistence, fostering interoperability and reducing divergent approaches.

* Users with Stateful/Legacy Applications: Applications that rely on client IP Persistence (e.g., certain legacy applications, gaming servers, or applications with in-memory session stores) will directly benefit from a reliable and configurable persistence mechanism.

## API

TODO: when we get to the implementation iterations, we need to consider whether we can implement this using functionality established in [GEP-1619](https://gateway-api.sigs.k8s.io/geps/gep-1619).

### Conformance tests 

NA

## Alternatives

Yet to do

## References

Below are references showing how ClientIP persistence is currently supported across some implementations:

* [AVI](https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-load-balancer/30-2/load-balancing-overview/persistence.html)
* [Native k8s](https://kubernetes.io/docs/reference/networking/virtual-ips/#session-affinity)

Below are some implementations of ClientIP persistence which allows configuring subnet Mask 

* [F5](https://techdocs.f5.com/content/kb/en-us/products/big-ip_ltm/manuals/product/ltm-concepts-11-5-1/11.html#:~:text=is%20persisted%20properly.-,Source%20address%20affinity%20persistence,-Source%20address%20affinity)
* [Fortinet](https://help.fortinet.com/fadc/4-8-0/olh/Content/FortiADC/handbook/slb_persistence.htm)
* [Huawei](https://info.support.huawei.com/hedex/api/pages/EDOC1100149308/AEJ0713J/18/resources/cli/session_persistence.html)
* [NetScaler](https://docs.netscaler.com/en-us/citrix-adc/current-release/load-balancing/load-balancing-persistence/no-rule-persistence#:~:text=For%20IP%2Dbased%20persistence%2C%20you%20can%20also%20set%20the%20persistMask%20parameter)
* [AVI](https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-load-balancer/30-2/load-balancing-overview/persistence/client-ip-persistence.html)
