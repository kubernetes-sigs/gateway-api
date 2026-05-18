---
title: "GEP-2644: TCPRoute"
---

- Issue: [#2644](https://github.com/kubernetes-sigs/gateway-api/issues/2644)
- Status: Provisional

(See definitions in [GEP States](../overview.md#gep-states).)

## TLDR

Gateway API needs a first-class Route type for TCP because a meaningful class of Kubernetes workloads cannot be represented by the standard Gateway API routing model without it.
HTTPRoute only addresses a single class of TCP-based traffic (HTTP/1.1 and HTTP/2 over TCP), and TLSRoute handles routing via SNI, leaving the broader universe of TCP workloads that don't work in either of those cases without a portable routing API, which prevents Gateway API from serving as a common and portable configuration model for them.
TCPRoute standardizes the minimal interoperable API surface for exposing TCP workloads through Gateway API.

This GEP retroactively documents the rationale, scope, and design constraints of the existing TCPRoute resource.

## Goals

- Define a standard Gateway API Route resource that matches TCP traffic based off the inbound port and forwards it to a TCP based backend.
- Standardize the minimal interoperable forwarding model for TCP traffic: listener attachment and backend forwarding.
- Support forwarding to one or more backends, including weighted backend selection where implemented.
- Provide a stable baseline for evaluating future TCPRoute enhancements.
- Promote TCPRoute to v1 to signal stability.

## Non-Goals

- Define rich TCP-specific matching semantics such as address matching, payload inspection, idle timeouts, or client ip preservation.
- Define TLS termination behavior at the Gateway. TLS-aware L4 routing is covered by `TLSRoute`.
- Define SNI-based or hostname-based routing. Use `TLSRoute` for SNI routing.
- Define HTTP or gRPC routing behavior. Use `HTTPRoute` or `GRPCRoute` for L7 routing.

## Longer Term Goals

The following topics are out of scope for this GEP but are candidates for future GEPs:

- Client IP preservation (e.g. PROXY protocol)
- Connection draining and timeout semantics
- Extended routing options (e.g. 5-tuple or 3-tuple matching)
- Listener port range support for TCP

## Introduction

This GEP proposes a standard resource for TCP traffic routing. Presently, the Gateway API lacks a way to describe how
to route plain TCP traffic meaning that it's hard (or impossible) to natively define such applications using the Gateway API:

- Databases (e.g. PostgreSQL, MySQL, Redis)
- Message brokers (e.g. Kafka, RabbitMQ, NATS)
- SMTP, IMAP, and other mail protocols
- Legacy or proprietary TCP-based services

Without TCPRoute, users must rely on implementation-specific extensions or fall back to traditional Kubernetes Service resources.
TCPRoute allows for consistent network configuration management using the Gateway API. Another benefit is that
organizations may consolidate their load balancing infrastructure under one Gateway, instead of having one physical
load balancer per Service.

## API

### TCPRoute Resource

These resources follow the same pattern as other route types. Notably, as TCP doesn't have traffic control options
exposed at this layer, the route rule only includes backends to forward TCP traffic to.

```go
// TCPRoute provides a way to route TCP traffic. When combined with a Gateway
// listener, it can be used to forward traffic on the port specified by the
// listener to a set of backends specified by the TCPRoute.
type TCPRoute struct {
    metav1.TypeMeta `json:",inline"`
    // +optional
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of TCPRoute.
    // +required
    Spec TCPRouteSpec `json:"spec"`

    // Status defines the current state of TCPRoute.
    // +optional
    Status TCPRouteStatus `json:"status,omitempty"`
}

// TCPRouteSpec defines the desired state of TCPRoute.
type TCPRouteSpec struct {
    CommonRouteSpec `json:",inline"`

    // Rules are a list of TCP matchers and actions.
    //
    // +required
    // +listType=atomic
    // +kubebuilder:validation:MinItems=1
    // +kubebuilder:validation:MaxItems=1
    // <gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) || self.exists_one(l2, has(l2.name) && l1.name == l2.name))">
    Rules []TCPRouteRule `json:"rules"`
}

// TCPRouteRule is the configuration for a given rule.
type TCPRouteRule struct {
    // Name is the name of the route rule. This name MUST be unique within a Route if it is set.
    //
    // Support: Extended
    // +optional
    Name *SectionName `json:"name,omitempty"`

    // BackendRefs defines the backend(s) where matching requests should be
    // sent. If unspecified or invalid (refers to a nonexistent resource or a
    // Service with no endpoints), the underlying implementation MUST actively
    // reject connection attempts to this backend. Connection rejections must
    // respect weight; if an invalid backend is requested to have 80% of
    // the connections, then 80% of connections must be rejected instead.
    //
    // Support: Core for Kubernetes Service
    //
    // Support: Extended for Kubernetes ServiceImport
    //
    // Support: Implementation-specific for any other resource
    //
    // Support for weight: Extended
    //
    // +required
    // +listType=atomic
    // +kubebuilder:validation:MinItems=1
    // +kubebuilder:validation:MaxItems=16
    BackendRefs []BackendRef `json:"backendRefs,omitempty"`
}

// TCPRouteStatus defines the observed state of TCPRoute.
type TCPRouteStatus struct {
	RouteStatus `json:",inline"`
}
```

## Request flow

Following are some of the request flows covered by TCPRoute, and the expected
behavior:

### Basic TCP Forwarding

In this workflow, TCP traffic arriving at a specific listener port is forwarded
directly to the backend service. This is the base use case for the TCPRoute
object, and is included in the core `TCPRoute` feature.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tcp-gateway
  namespace: gateway-conformance-infra
spec:
  gatewayClassName: example-gateway-class
  listeners:
    - name: postgres
      protocol: TCP
      port: 5432
      allowedRoutes:
        kinds:
          - kind: TCPRoute
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-postgres
  namespace: gateway-conformance-infra
spec:
  parentRefs:
    - name: tcp-gateway
      sectionName: postgres
  rules:
    - backendRefs:
        - name: postgres
          port: 5432
```

A typical [north/south](/docs/glossary/#northsouth-traffic)
request flow for a gateway implemented using a `TCPRoute` is:

* A client opens a TCP connection to the Gateway address on port 5432.
* The Gateway receives the connection on the `Listener` matching port 5432.
* The Gateway identifies the `TCPRoute` attached to that `Listener`.
* The Gateway opens a new TCP connection to one of the backend `Service`
  endpoints based on the `backendRefs` of the `TCPRoute`.
* The Gateway proxies the TCP stream bidirectionally between the client
  and the backend for the lifetime of the connection.

### Multiple Weighted Backends

In this workflow, TCP connections are distributed across multiple backends
according to configured weights. Weight support is Extended for TCPRoute.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: tcp-gateway
  namespace: gateway-conformance-infra
spec:
  gatewayClassName: example-gateway-class
  listeners:
    - name: kafka
      protocol: TCP
      port: 9092
      allowedRoutes:
        kinds:
          - kind: TCPRoute
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: kafka-route
  namespace: gateway-conformance-infra
spec:
  parentRefs:
    - name: tcp-gateway
      sectionName: kafka
  rules:
    - backendRefs:
        - name: kafka-broker-1
          port: 9092
          weight: 70
        - name: kafka-broker-2
          port: 9092
          weight: 30
```

* A client opens a TCP connection to the Gateway address on port 9092.
* The Gateway receives the connection on the `Listener` matching port 9092.
* The Gateway identifies the `TCPRoute` attached to that `Listener`.
* The Gateway selects a backend based on the configured weights: approximately
  70% of new connections are sent to `kafka-broker-1` and 30% to
  `kafka-broker-2`. Each TCP SYN (new connection) is a single unit for
  weighting purposes.
* The Gateway proxies the TCP stream bidirectionally between the client
  and the selected backend.

### Mixing TCP and UDP Protocols

A common use case is to expose the same service over TCP and UDP (e.g. DNS).
An implementation MAY listen for both TCP and UDP traffic utilizing the same
Listener port. In this example, all TCP traffic MUST be routed to the TCPRoute
and all UDP traffic MUST be routed to the UDPRoute.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: dns-gateway
  namespace: default
spec:
  gatewayClassName: example-gatewayclass
  listeners:
    - name: dns-tcp
      protocol: TCP
      port: 53
      allowedRoutes:
        kinds:
          - kind: TCPRoute
            group: gateway.networking.k8s.io

    - name: dns-udp
      protocol: UDP
      port: 53
      allowedRoutes:
        kinds:
          - kind: UDPRoute
            group: gateway.networking.k8s.io
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: dns-tcp-route
  namespace: default
spec:
  parentRefs:
    - name: dns-gateway
      sectionName: dns-tcp
  rules:
    - backendRefs:
        - name: dns-tcp-service
          port: 53
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: dns-udp-route
  namespace: default
spec:
  parentRefs:
    - name: dns-gateway
      sectionName: dns-udp
  rules:
    - backendRefs:
        - name: dns-udp-service
          port: 53
```

* A client sends a DNS query over TCP to the Gateway address on port 53.
* The Gateway receives the connection on the `dns-tcp` `Listener`.
* The Gateway identifies the `TCPRoute` attached to that `Listener` and
  forwards the TCP connection to `dns-tcp-service`.
* Separately, a client sends a DNS query over UDP to the same Gateway
  address on port 53.
* The Gateway receives the datagram on the `dns-udp` `Listener`.
* The Gateway identifies the `UDPRoute` attached to that `Listener` and
  forwards the UDP datagram to `dns-udp-service`.

## Conflict management and precedences

A conflict can happen when two or more distinct listeners on a Gateway definition
have conflicting behavior, or when multiple routes attempt to attach to the same
listener.

Unlike TLSRoute (which uses SNI hostnames to disambiguate traffic on the same port),
TCPRoute traffic is classified only by `protocol:port`. This means that a TCP
listener's only distinctness field is **port** — there is no further attribute
to discriminate between connections arriving on the same listener.

### Two TCP listeners on the same port

Two TCP listeners on the same port are always indistinct, because there is no
hostname or other field to differentiate them. The implementation MUST mark both
listeners as conflicted:

```yaml
spec:
  listeners:
  - name: listener1
    port: 5432
    protocol: TCP
  - name: listener2
    port: 5432
    protocol: TCP
status:
  listeners:
  - name: listener1
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: ProtocolConflict
      status: "True"
      type: Conflicted
  - name: listener2
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: ProtocolConflict
      status: "True"
      type: Conflicted
```

### TCP listener on the same port as an HTTP, HTTPS, or TLS listener

When a TCP protocol listener shares a port with an HTTP, HTTPS, or TLS protocol
listener, all listeners sharing that port are indistinct and MUST NOT be accepted.
TCP operates at a lower layer and cannot coexist with protocols that require
additional parsing (e.g. HTTP Host header or TLS SNI) on the same port:

```yaml
spec:
  listeners:
  - name: tcp-listener
    port: 443
    protocol: TCP
  - name: https-listener
    port: 443
    protocol: HTTPS
    hostname: "app.example.com"
    tls:
      mode: Terminate
      certificateRefs:
      - name: app-cert
        kind: Secret
status:
  listeners:
  - name: tcp-listener
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: ProtocolConflict
      status: "True"
      type: Conflicted
  - name: https-listener
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: ProtocolConflict
      status: "True"
      type: Conflicted
```

### Multiple TCPRoutes attaching to the same listener

Because a TCP listener has no mechanism to distinguish between connections (no
hostname, no SNI, no path), attaching multiple TCPRoutes to the same listener
results in only one route effectively receiving traffic.

When multiple TCPRoutes reference the same listener, the implementation MUST
follow the general Gateway API route precedence rules defined in `AllowedRoutes`:

1. The oldest Route based on `metadata.creationTimestamp`.
2. If timestamps are equal, the Route appearing first in alphabetical order
   (`namespace/name`).

All attached TCPRoutes are `Accepted`, consistent with how other route types
handle precedence in the Gateway API. Only the winning route's backends receive
traffic.

> **Note:** Accepting all conflicting routes without surfacing which one is
> actively receiving traffic is not optimal for user experience. A future GEP
> may introduce a dedicated route condition reason to explicitly signal that a
> route has been superseded by another route on the same listener.

```yaml
# Two TCPRoutes targeting the same listener
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-route-1
  creationTimestamp: "2026-01-01T00:00:00Z"
spec:
  parentRefs:
    - name: tcp-gateway
      sectionName: postgres
  rules:
    - backendRefs:
        - name: postgres-primary
          port: 5432
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: tcp-route-2
  creationTimestamp: "2026-01-02T00:00:00Z"
spec:
  parentRefs:
    - name: tcp-gateway
      sectionName: postgres
  rules:
    - backendRefs:
        - name: postgres-replica
          port: 5432
```

Both routes are accepted, but only `tcp-route-1` (the oldest) receives traffic:

```yaml
# tcp-route-1 status
status:
  parents:
  - parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: tcp-gateway
      sectionName: postgres
    conditions:
    - type: Accepted
      status: "True"
      reason: Accepted

# tcp-route-2 status (accepted, but does not receive traffic)
status:
  parents:
  - parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: tcp-gateway
      sectionName: postgres
    conditions:
    - type: Accepted
      status: "True"
      reason: Accepted
```

### TCP and UDP listeners on the same port

TCP and UDP are distinct transport protocols. An implementation MAY support
listeners for both TCP and UDP on the same port without conflict. See the
[Mixing TCP and UDP Protocols](#mixing-tcp-and-udp-protocols) request flow for
an example.

## Conformance Details

### Feature Names

* TCPRoute

### Conformance test scenarios

The following scenarios will be validated:

1. TCPRoute attaches to a TCP listener by port specified
   - A Gateway has a TCP listener on port 5432
   - A TCPRoute specifies a `parentRef` with `port: 5432`

   - The TCPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         port: 5432
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - TCP traffic sent to port 5432 is forwarded to the backend.

   Features: `TCPRoute`

1. TCPRoute attaches to a TCP listener by sectionName
   - A Gateway has a TCP listener named `postgres` on port 5432
   - A TCPRoute specifies a `parentRef` with `sectionName: postgres`

   - The TCPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - TCP traffic sent to the `postgres` listener is forwarded to the backend.

   Features: `TCPRoute`

1. TCPRoute attaches to a TCP listener by sectionName and port
   - A Gateway has a TCP listener named `postgres` on port 5432
   - A TCPRoute specifies a `parentRef` with `sectionName: postgres` and `port: 5432`

   - The TCPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
         port: 5432
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - TCP traffic sent to port 5432 on the `postgres` listener is forwarded to the backend.

   Features: `TCPRoute`

1. TCPRoute attaches to all TCP listeners in a Gateway when sectionName and port are omitted
   - A Gateway has multiple TCP listeners: `postgres` on port 5432 and `kafka` on port 9092
   - A TCPRoute specifies a `parentRef` with only the Gateway name (no `sectionName` or `port`)

   - The TCPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - The TCPRoute attaches to all TCP listeners on the Gateway.

   Features: `TCPRoute`

1. TCPRoute fails attachment to a non-TCP listener when port or sectionName is specified
   - A Gateway has a UDP listener named `udp-listener` on port 5432 and no TCP listeners on that port/name
   - A TCPRoute specifies a `parentRef` targeting the UDP listener via `sectionName: udp-listener`

   - The TCPRoute is not accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: mixed-gateway
         sectionName: udp-listener
       conditions:
       - type: Accepted
         status: "False"
         reason: NotAllowedByListeners
     ```

   - No TCP traffic is routed through the UDP listener.

   Features: `TCPRoute`

1. TCPRoute references a backend Service that does not exist
   - A Gateway has a TCP listener named `postgres` on port 5432
   - A TCPRoute specifies a `parentRef` with `sectionName: postgres` and a `backendRef` pointing to a Service `nonexistent-service` that does not exist

   - The TCPRoute has the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "False"
         reason: BackendNotFound
     ```

   - TCP connections to this route MUST be rejected.

   Features: `TCPRoute`

1. TCPRoute with a cross-namespace backendRef and no valid ReferenceGrant
   - A Gateway has a TCP listener named `postgres` on port 5432
   - A TCPRoute in the `gateway-conformance-infra` namespace specifies a `backendRef` pointing to a Service in another namespace without a valid ReferenceGrant

   - The TCPRoute has the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "False"
         reason: RefNotPermitted
     ```

   - TCP connections to this route MUST be rejected.

   Features: `TCPRoute`, `ReferenceGrant`

1. TCPRoute with multiple weighted backends distributes connections according to configured weights
   - A Gateway has a TCP listener named `kafka` on port 9092
   - A TCPRoute specifies a `parentRef` with `sectionName: kafka` and multiple `backendRefs`:
     - `kafka-broker-1` with `weight: 70`
     - `kafka-broker-2` with `weight: 30`

   - The TCPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: kafka
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "True"
         reason: ResolvedRefs
     ```

   - TCP connections sent to port 9092 are distributed across backends respecting the configured weights (approximately 70% to `kafka-broker-1` and 30% to `kafka-broker-2`). In this situation, TCP SYN packets denote one connection for weighting purposes.

   Features: `TCPRoute` (weight support is Extended)

1. Two TCP listeners on the same port are marked as conflicted
   - A Gateway has two TCP listeners on port 5432: `listener1` and `listener2`

   - Both listeners are marked as conflicted with the following status:

     ```
     listeners:
     - name: listener1
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: Conflicted
         status: "True"
         reason: ProtocolConflict
     - name: listener2
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: Conflicted
         status: "True"
         reason: ProtocolConflict
     ```

   - No traffic is routed through either listener.

   Features: `TCPRoute`

1. TCP listener on the same port as an HTTPS listener causes conflict
   - A Gateway has a TCP listener on port 443 and an HTTPS listener on port 443

   - Both listeners are marked as conflicted with the following status:

     ```
     listeners:
     - name: tcp-listener
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: Conflicted
         status: "True"
         reason: ProtocolConflict
     - name: https-listener
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: Conflicted
         status: "True"
         reason: ProtocolConflict
     ```

   - No traffic is routed through either listener.

   Features: `TCPRoute`

1. Multiple TCPRoutes attaching to the same listener results in only the oldest receiving traffic
   - A Gateway has a TCP listener named `postgres` on port 5432
   - Two TCPRoutes (`tcp-route-1` created first, `tcp-route-2` created second) both specify a `parentRef` with `sectionName: postgres`

   - Both routes are accepted:

     ```
     # tcp-route-1 status
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

     ```
     # tcp-route-2 status
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: tcp-gateway
         sectionName: postgres
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - TCP traffic sent to port 5432 is forwarded only to the backends of `tcp-route-1` (the oldest route).

   Features: `TCPRoute`

## Standard Graduation Criteria

The TCPRoute resource pre-dates the current GEP process and has existed in the
Gateway API since early releases as a `v1alpha2` resource. This GEP retroactively
documents the rationale, scope, and design constraints of the existing resource.

Because TCPRoute has been available and implemented for a significant period of
time, it is being grandfathered into the current process and is not subject to the
standard probationary period requirement. The graduation criteria for Standard are:

* At least one Feature Name must be listed: `TCPRoute`.
* The Conformance Details must be filled out, with conformance test scenarios listed.
* Conformance tests must be implemented that test all the listed test scenarios.
* At least three (3) implementations must have submitted conformance reports that
  pass those conformance tests.

## References

- [TCPRoute and UDPRoute Specification](/docs/concepts/api-overview/#tcproute-and-udproute)
- [TLSRoute Specification](/reference/api-types/tlsroute/)
- [GEP-735: TCP and UDP addresses matching](../gep-735/index.md) (Declined, but relevant context)
- [Gateway API Use Cases](/docs/concepts/use-cases/)

