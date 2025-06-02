# GEP-3798: Client IP-Based Session Persistence

* Issue: [#3798](https://github.com/kubernetes-sigs/gateway-api/issues/3798)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

## TLDR

### What
 This GEP proposes the addition of Client IP-based session persistence to the Gateway API. This feature will allow Gateway API implementations to ensure that requests originating from a specific client IP address (or a subnet defined by an IP mask) are consistently routed to the same backend endpoint for a configurable duration. This aims to provide a standardized and centralized mechanism for client IP persistence across various Gateway API implementations. As per the nomenclature established in [#1619](https://gateway-api.sigs.k8s.io/geps/gep-1619), this feature is being referred as Session Affinity . 


## Goals

* Define an API extension within Gateway API to enable client IP-based session persistence.

* Allow configuration of a session duration for which a client IP should stick to a backend.

* Provide an optional mechanism to specify an IP mask for subnet-based persistence, allowing multiple clients from the same subnet to be routed to the same backend.

* Ensure the solution is generic enough to be implemented by various Gateway API controllers.

* Improve portability of applications requiring client IP persistence across different Gateway API implementations.

## Non-Goals

This GEP does not dictate the specific algorithm or implementation details for how an individual Gateway controller maintains the client IP-to-backend mapping (e.g., in-memory, distributed cache).

## Introduction

### Why: The Problem This Solves
Currently, achieving client IP-based session persistence within Kubernetes Gateway API environments often requires vendor-specific annotations or out-of-band configurations on the underlying load balancer or ingress controller. This approach has several drawbacks:

* Lack of Portability: Configurations are not easily transferable between different Gateway API implementations, leading to vendor lock-in and increased operational overhead when migrating or using multiple controllers.

* Inconsistent User Experience: Users have to learn different methods for configuring the same logical feature depending on their chosen Gateway API implementation.

* Limited API Expressiveness: Important traffic management capabilities are not directly exposed or controlled through the Gateway API, making it less comprehensive for certain application requirements.

* Reduced Visibility: The desired session persistence behavior is not explicitly declared within the Gateway API resources, making it harder to audit, manage, and understand the routing logic from a single source of truth.

This GEP addresses these issues by providing a first-class API mechanism for client IP-based session persistence, enhancing the Gateway API's capabilities and promoting consistency and portability.

### Who: Beneficiaries
* Application Developers: Can define session persistence requirements directly in their Gateway API configurations, ensuring consistent behavior regardless of the underlying Gateway implementation. This simplifies application deployment and management for stateful workloads.

* Platform Operators/Administrators: Gain a standardized way to configure and manage client IP-based session persistence across their clusters, reducing the need for custom scripts or deep knowledge of individual controller implementations. This improves operational efficiency and consistency.

* Gateway API Implementers: Receive a clear specification for implementing client IP-based session persistence, fostering interoperability and reducing divergent approaches.

* Users with Stateful Applications: Applications that rely on client IP affinity (e.g., certain legacy applications, gaming servers, or applications with in-memory session stores) will directly benefit from a reliable and configurable persistence mechanism.

## API

As mentioned in the [GEP-1619](https://gateway-api.sigs.k8s.io/geps/gep-1619/#api), `SessionPersistence` can be applied via `BackendLBPolicy` and `RouteRule` API .Similar [edge case behaviour](https://gateway-api.sigs.k8s.io/geps/gep-1619/#edge-case-behavior) and [API Granularity](https://gateway-api.sigs.k8s.io/geps/gep-1619/#api-granularity) for ClientIP Persistence type should be applicable as well.  

Requirement is to introduce a new `SessionPersistenceType` called `ClientIP`

Example (illustrative, exact field names and structure are subject to review):

```
# Existing SessionPersistence (simplified for example)
# apiVersion: gateway.networking.k8s.io/v1beta1
# kind: HTTPRoute

spec:
  rules:
  - backendRefs:
    - name: my-service
      port: 80
    sessionPersistence:
      # New field for client IP based persistence
      type: "ClientIP"
      absoluteTimeout: "5m"
      ipMask: 24 # Optional: IP mask for subnet persistence (e.g., "24" for /24 subnet)
```
```
type SessionPersistence struct {
	...

    // IPMask defines the IP mask to be applied on client this may be
	// used to persist clients from a same subnet to stick to same session
	//
	// Support: Implementation-specific
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=128
	IPMask *uint32 `json:"ipMask,omitempty"`

}

type SessionPersistenceType string

const (
    // CookieBasedSessionPersistence specifies cookie-based session
    // persistence.
    //
    // Support: Core
    CookieBasedSessionPersistence   SessionPersistenceType = "Cookie"

    // HeaderBasedSessionPersistence specifies header-based session
    // persistence.
    //
    // Support: Extended
    HeaderBasedSessionPersistence   SessionPersistenceType = "Header"

    // ClientIPBasedSessionPersistence specifies Client IP based session
    // persistence.
    //
    // Support: Implementation-specific
    ClientIPBasedSessionPersistence   SessionPersistenceType = "ClientIP"
)
```

### Conformance tests 

NA

## Alternatives

Yet to do

## References

Below are references showing how ClientIP persistence is currently supported across some implementations:

* [AVI](https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-load-balancer/30-2/load-balancing-overview/persistence.html)
* [Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-msg-config-route-v3-routeaction-hashpolicy-connectionproperties) (the connection property hash policy can be used with Ring Hash load balancing to ensure session persistence for a particular source IP)
* [Nginx](https://nginx.org/en/docs/http/ngx_http_upstream_module.html#ip_hash)
* [Native k8s](https://kubernetes.io/docs/reference/networking/virtual-ips/#session-affinity)

