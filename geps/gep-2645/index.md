---
title: "GEP-2645: UDPRoute"
---

- Issue: [#2645](https://github.com/kubernetes-sigs/gateway-api/issues/2645)
- Status: Provisional

(See definitions in [GEP States](../overview.md#gep-states).)

## TLDR

Gateway API needs a first-class Route type for UDP because a meaningful class of Kubernetes workloads cannot be represented by the standard Gateway API routing model without it.
Without UDPRoute, these UDP based workloads must either fall back to Service-based exposure or rely on implementation-specific APIs, which prevents Gateway API from serving as a common and portable configuration model for them.
UDPRoute standardizes the minimal interoperable API surface for exposing UDP workloads through Gateway API.

This GEP retroactively documents the rationale, scope, and design constraints of the existing UDPRoute resource.

## Goals

- Define a standard Gateway API Route resource that matches UDP traffic based off the inbound port and forwards it to a UDP based backend.
- Standardize the minimal interoperable forwarding model for UDP traffic: listener attachment and backend forwarding.
- Support forwarding to one or more backends, including weighted backend selection where implemented.
- Provide a stable baseline for evaluating future UDPRoute enhancements.
- Promote UDPRoute to v1 to signal stability.

## Non-Goals

- Define rich UDP-specific matching semantics such as address matching, payload inspection, or idle timeouts.
- Require stateful UDP session tracking or connection management semantics. Implementations are expected to document how they implement such semantics.
- Define DTLS termination behavior at the Gateway.
- Define HTTP/3 or QUIC-specific behavior.
- Define a mechanism to describe or require UDP packet routing as well as forwarding

## Longer Term Goals

The following topics are out of scope for this GEP but are candidates for future GEPs:

- Client IP preservation
- Flow management (e.g. flow timeout)
- Extended routing options (e.g. 5-tuple or 3-tuple matching)
- Port range support

## Introduction

This GEP proposes a standard resource for UDP traffic routing. Presently, the Gateway API lacks a way to describe how
to route UDP traffic meaning that it's hard (or impossible) to natively define such applications using the Gateway API:

- DNS (Domain Name System)
- VoIP and real-time communications
- Gaming protocols
- Streaming media (RTP/RTCP)
- IoT and telemetry protocols

Without UDPRoute, users must rely on implementation-specific extensions or fall back to traditional Kubernetes Service resources.
UDPRoute allows for consistent network configuration management using the Gateway API. Another benefit is that
organizations may consolidate their load balancing infrastructure under one Gateway, instead of having one physical
load balancer per Service.

## API

### UDPRoute Resource

These resources follow the same pattern as other route types. Notably, as UDP doesn't have traffic control options,
the route rule only includes backends to forward UDP traffic to.

```go
// UDPRoute provides a way to route UDP traffic. When combined with a Gateway
// listener, it can be used to forward traffic on the port specified by the
// listener to a set of backends specified by the UDPRoute.
type UDPRoute struct {
    metav1.TypeMeta `json:",inline"`
    // +optional
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of UDPRoute.
    // +required
    Spec UDPRouteSpec `json:"spec"`

    // Status defines the current state of UDPRoute.
    // +optional
    Status UDPRouteStatus `json:"status,omitempty"`
}

// UDPRouteSpec defines the desired state of UDPRoute.
type UDPRouteSpec struct {
    CommonRouteSpec `json:",inline"`

    // Rules are a list of UDP matchers and actions.
    //
    // +required
    // +listType=atomic
    // +kubebuilder:validation:MinItems=1
    // +kubebuilder:validation:MaxItems=1
    // <gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) || self.exists_one(l2, has(l2.name) && l1.name == l2.name))">
    Rules []UDPRouteRule `json:"rules"`
}

// UDPRouteRule is the configuration for a given rule.
type UDPRouteRule struct {
    // Name is the name of the route rule. This name MUST be unique within a Route if it is set.
    //
    // Support: Extended
    // +optional
    Name *SectionName `json:"name,omitempty"`

    // BackendRefs defines the backend(s) where matching requests should be
    // sent. If unspecified or invalid (refers to a nonexistent resource or a
    // Service with no endpoints), the underlying implementation MUST actively
    // reject connection attempts to this backend. Packet drops must
    // respect weight; if an invalid backend is requested to have 80% of
    // the packets, then 80% of packets must be dropped instead.
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

// UDPRouteStatus defines the observed state of UDPRoute.
type UDPRouteStatus struct {
	RouteStatus `json:",inline"`
}
```

## Request flow

Following are some of the request flows covered by UDPRoute, and the expected
behavior:

### Basic UDP Forwarding

In this workflow, UDP traffic arriving at a specific listener port is forwarded
directly to the backend service. This is the base use case for the UDPRoute
object, and is included in the core `UDPRoute` feature.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: udp-gateway
  namespace: gateway-conformance-infra
spec:
  gatewayClassName: example-gateway-class
  listeners:
    - name: coredns
      protocol: UDP
      port: 5300
      allowedRoutes:
        kinds:
          - kind: UDPRoute
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: udp-coredns
  namespace: gateway-conformance-infra
spec:
  parentRefs:
    - name: udp-gateway
      sectionName: coredns
  rules:
    - backendRefs:
        - name: coredns
          port: 53
```

A typical [north/south](/docs/glossary/#northsouth-traffic)
request flow for a gateway implemented using a `UDPRoute` is:

* A client sends a UDP datagram to the Gateway address on port 5300.
* The Gateway receives the datagram on the `Listener` matching port 5300.
* The Gateway identifies the `UDPRoute` attached to that `Listener`.
* The Gateway forwards the datagram to one of the backend `Service`
  endpoints based on the `backendRefs` of the `UDPRoute`.
* Response datagrams from the backend are forwarded back to the client.

### Multiple Weighted Backends

In this workflow, UDP datagrams are distributed across multiple backends
according to configured weights. Weight support is Extended for UDPRoute.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: udp-gateway
  namespace: gateway-conformance-infra
spec:
  gatewayClassName: example-gateway-class
  listeners:
    - name: gaming
      protocol: UDP
      port: 7777
      allowedRoutes:
        kinds:
          - kind: UDPRoute
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: game-server-route
  namespace: gateway-conformance-infra
spec:
  parentRefs:
    - name: udp-gateway
      sectionName: gaming
  rules:
    - backendRefs:
        - name: game-server-1
          port: 7777
          weight: 70
        - name: game-server-2
          port: 7777
          weight: 30
```

* A client sends a UDP datagram to the Gateway address on port 7777.
* The Gateway receives the datagram on the `Listener` matching port 7777.
* The Gateway identifies the `UDPRoute` attached to that `Listener`.
* The Gateway selects a backend based on the configured weights: approximately
  70% of new flows are sent to `game-server-1` and 30% to `game-server-2`.
* Response datagrams from the selected backend are forwarded back to the client.

### Mixing UDP and TCP Protocols

A common use case is to expose the same service over UDP and TCP (e.g. DNS).
An implementation MAY listen for both UDP and TCP traffic utilizing the same
Listener port. In this example, all UDP traffic MUST be routed to the UDPRoute
and all TCP traffic MUST be routed to the TCPRoute.

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

* A client sends a DNS query over UDP to the Gateway address on port 53.
* The Gateway receives the datagram on the `dns-udp` `Listener`.
* The Gateway identifies the `UDPRoute` attached to that `Listener` and
  forwards the UDP datagram to `dns-udp-service`.
* Separately, a client sends a DNS query over TCP to the same Gateway
  address on port 53.
* The Gateway receives the connection on the `dns-tcp` `Listener`.
* The Gateway identifies the `TCPRoute` attached to that `Listener` and
  forwards the TCP connection to `dns-tcp-service`.

## Conflict management and precedences

A conflict can happen when two or more distinct listeners on a Gateway definition
have conflicting behavior, or when multiple routes attempt to attach to the same
listener.

Like TCPRoute, UDPRoute traffic is classified only by `protocol:port`. This means
that a UDP listener's only distinctness field is **port** — there is no further
attribute to discriminate between datagrams arriving on the same listener.

### Two UDP listeners on the same port

Two UDP listeners on the same port are always indistinct, because there is no
hostname or other field to differentiate them. The implementation MUST mark both
listeners as conflicted:

```yaml
spec:
  listeners:
  - name: listener1
    port: 5300
    protocol: UDP
  - name: listener2
    port: 5300
    protocol: UDP
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

### Multiple UDPRoutes attaching to the same listener

Because a UDP listener has no mechanism to distinguish between datagrams (no
hostname, no SNI, no path), attaching multiple UDPRoutes to the same listener
results in only one route effectively receiving traffic.

When multiple UDPRoutes reference the same listener, the implementation MUST
follow the general Gateway API route precedence rules defined in `AllowedRoutes`:

1. The oldest Route based on `metadata.creationTimestamp`.
2. If timestamps are equal, the Route appearing first in alphabetical order
   (`namespace/name`).

All attached UDPRoutes are `Accepted`, consistent with how other route types
handle precedence in the Gateway API. Only the winning route's backends receive
traffic.

> **Note:** Accepting all conflicting routes without surfacing which one is
> actively receiving traffic is not optimal for user experience. A future GEP
> may introduce a dedicated route condition reason to explicitly signal that a
> route has been superseded by another route on the same listener.

```yaml
# Two UDPRoutes targeting the same listener
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: udp-route-1
  creationTimestamp: "2026-01-01T00:00:00Z"
spec:
  parentRefs:
    - name: udp-gateway
      sectionName: coredns
  rules:
    - backendRefs:
        - name: coredns-primary
          port: 53
---
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: udp-route-2
  creationTimestamp: "2026-01-02T00:00:00Z"
spec:
  parentRefs:
    - name: udp-gateway
      sectionName: coredns
  rules:
    - backendRefs:
        - name: coredns-secondary
          port: 53
```

Both routes are accepted, but only `udp-route-1` (the oldest) receives traffic:

```yaml
# udp-route-1 status
status:
  parents:
  - parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: udp-gateway
      sectionName: coredns
    conditions:
    - type: Accepted
      status: "True"
      reason: Accepted

# udp-route-2 status (accepted, but does not receive traffic)
status:
  parents:
  - parentRef:
      group: gateway.networking.k8s.io
      kind: Gateway
      name: udp-gateway
      sectionName: coredns
    conditions:
    - type: Accepted
      status: "True"
      reason: Accepted
```

### UDP and TCP listeners on the same port

UDP and TCP are distinct transport protocols. An implementation MAY support
listeners for both UDP and TCP on the same port without conflict. See the
[Mixing UDP and TCP Protocols](#mixing-udp-and-tcp-protocols) request flow for
an example.

## Conformance Details

### Feature Names

* UDPRoute

### Conformance test scenarios

The following scenarios will be validated:

1. UDPRoute attaches to a UDP listener by port specified
   - A Gateway has a UDP listener on port 5300
   - A UDPRoute specifies a `parentRef` with `port: 5300`

   - The UDPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         port: 5300
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - UDP traffic sent to port 5300 is forwarded to the backend.

   Features: `UDPRoute`

1. UDPRoute attaches to a UDP listener by sectionName
   - A Gateway has a UDP listener named `coredns` on port 5300
   - A UDPRoute specifies a `parentRef` with `sectionName: coredns`

   - The UDPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - UDP traffic sent to the `coredns` listener is forwarded to the backend.

   Features: `UDPRoute`

1. UDPRoute attaches to a UDP listener by sectionName and port
   - A Gateway has a UDP listener named `coredns` on port 5300
   - A UDPRoute specifies a `parentRef` with `sectionName: coredns` and `port: 5300`

   - The UDPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
         port: 5300
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - UDP traffic sent to port 5300 on the `coredns` listener is forwarded to the backend.

   Features: `UDPRoute`

1. UDPRoute attaches to all UDP listeners in a Gateway when sectionName and port are omitted
   - A Gateway has multiple UDP listeners: `dns` on port 5300 and `game` on port 7777
   - A UDPRoute specifies a `parentRef` with only the Gateway name (no `sectionName` or `port`)

   - The UDPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - The UDPRoute attaches to all UDP listeners on the Gateway. Importantly, the backend must be prepared to handle the variety of traffic.

   Features: `UDPRoute`

1. UDPRoute fails attachment to a non-UDP listener when port or sectionName is specified
   - A Gateway has a TCP listener named `tcp-listener` on port 5300 and no UDP listeners on that port/name
   - A UDPRoute specifies a `parentRef` targeting the TCP listener via `sectionName: tcp-listener`

   - The UDPRoute is not accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: mixed-gateway
         sectionName: tcp-listener
       conditions:
       - type: Accepted
         status: "False"
         reason: NotAllowedByListeners
     ```

   - No UDP traffic is routed through the TCP listener.

   Features: `UDPRoute`

1. UDPRoute references a backend Service that does not exist
   - A Gateway has a UDP listener named `coredns` on port 5300
   - A UDPRoute specifies a `parentRef` with `sectionName: coredns` and a `backendRef` pointing to a Service `nonexistent-service` that does not exist

   - The UDPRoute has the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "False"
         reason: BackendNotFound
     ```

   - UDP datagrams to this route MUST be dropped.

   Features: `UDPRoute`

1. UDPRoute with a cross-namespace backendRef and no valid ReferenceGrant
   - A Gateway has a UDP listener named `coredns` on port 5300
   - A UDPRoute in the `gateway-conformance-infra` namespace specifies a `backendRef` pointing to a Service in another namespace without a valid ReferenceGrant

   - The UDPRoute has the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "False"
         reason: RefNotPermitted
     ```

   - UDP datagrams to this route MUST be dropped.

   Features: `UDPRoute`, `ReferenceGrant`

1. UDPRoute with multiple weighted backends distributes flows according to configured weights
   - A Gateway has a UDP listener named `game` on port 7777
   - A UDPRoute specifies a `parentRef` with `sectionName: game` and multiple `backendRefs`:
     - `game-server-1` with `weight: 70`
     - `game-server-2` with `weight: 30`

   - The UDPRoute is accepted with the following status:

     ```
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: game
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
       - type: ResolvedRefs
         status: "True"
         reason: ResolvedRefs
     ```

   - UDP flows sent to port 7777 are distributed across backends respecting the configured weights (approximately 70% to `game-server-1` and 30% to `game-server-2`).

   Features: `UDPRoute` (weight support is Extended)

1. Two UDP listeners on the same port are marked as conflicted
   - A Gateway has two UDP listeners on port 5300: `listener1` and `listener2`

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

   Features: `UDPRoute`

1. Multiple UDPRoutes attaching to the same listener results in only the oldest receiving traffic
   - A Gateway has a UDP listener named `coredns` on port 5300
   - Two UDPRoutes (`udp-route-1` created first, `udp-route-2` created second) both specify a `parentRef` with `sectionName: coredns`

   - Both routes are accepted:

     ```
     # udp-route-1 status
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

     ```
     # udp-route-2 status
     parents:
     - parentRef:
         group: gateway.networking.k8s.io
         kind: Gateway
         name: udp-gateway
         sectionName: coredns
       conditions:
       - type: Accepted
         status: "True"
         reason: Accepted
     ```

   - UDP traffic sent to port 5300 is forwarded only to the backends of `udp-route-1` (the oldest route).

   Features: `UDPRoute`

## Standard Graduation Criteria

The UDPRoute resource pre-dates the current GEP process and has existed in the
Gateway API since early releases as a `v1alpha2` resource. This GEP retroactively
documents the rationale, scope, and design constraints of the existing resource.

Because UDPRoute has been available and implemented for a significant period of
time, it is being grandfathered into the current process and is not subject to the
standard probationary period requirement. The graduation criteria for Standard are:

* At least one Feature Name must be listed: `UDPRoute`.
* The Conformance Details must be filled out, with conformance test scenarios listed.
* Conformance tests must be implemented that test all the listed test scenarios.
* At least three (3) implementations must have submitted conformance reports that
  pass those conformance tests.

## References

- [TCPRoute and UDPRoute Specification](/docs/concepts/api-overview/#tcproute-and-udproute)
- [GEP-735: TCP and UDP addresses matching](../gep-735/index.md) (Declined, but relevant context)
- [Gateway API Use Cases](/docs/concepts/use-cases/)
