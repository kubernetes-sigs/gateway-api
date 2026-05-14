# GEP-2645: UDPRoute

- Issue: [#2645](https://github.com/kubernetes-sigs/gateway-api/issues/2645)
- Status: Provisional

(See definitions in [GEP States](../overview.md#gep-states).)

## TLDR

Gateway API needs a first-class Route type for UDP because a meaningful class of Kubernetes workloads cannot be represented by the standard Gateway API routing model without it.
Without UDPRoute, these UDP based workloads must either fall back to Service-based exposure or rely on implementation-specific APIs,
which prevents Gateway API from serving as a common and portable configuration model for them.
UDPRoute standardizes the minimal interoperable API surface for exposing UDP workloads through Gateway API.

This GEP retroactively documents the rationale, scope, and design constraints of the existing UDPRoute resource.

## Goals

- Define a standard Gateway API Route resource that matches UDP traffic based off the inbound port and forwards it to a UDP based backend.
- Standardize the minimal interoperable forwarding model for UDP traffic: listener attachment and backend forwarding.
- Support forwarding to one or more backends, including weighted backend selection where implemented.
- Provide a stable baseline for evaluating future UDPRoute enhancements.

## Non-Goals

- Define rich UDP-specific matching semantics such as address matching, payload inspection, or idle timeouts.
- Require stateful UDP session tracking or connection management semantics. Implementations are expected to document how they implement such semantics.
- Define DTLS termination behavior at the Gateway.
- Define HTTP/3 or QUIC-specific behavior.

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

### Example Usage

#### Basic UDP Service Routing (DNS)

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

#### Multiple Backend Distribution

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: UDPRoute
metadata:
  name: game-server-route
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

### Mixing Protocols

A common use-case is to expose the same service over UDP and TCP. An Implementation MAY listen for both UDP and TCP traffic
utilizing the same Listener port. In this example, all UDP traffic MUST be routed to the UDP route and all TCP traffic
must be routed to the TCP route.

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

## Conformance Details

The following Gateway Conformance features will be added:

```
	// SupportUDPRoute option indicates support for UDPRoute
	SupportUDPRoute FeatureName = "UDPRoute"
```

They will validate the following scenarios:

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
         status: "False"
         reason: BackendNotFound
       - type: ResolvedRefs
         status: "False"
         reason: BackendNotFound
     ```

   - No UDP traffic is forwarded for this route.

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

   - UDP flows are sent to port 7777 is distributed across backends respecting the configured weights (approximately 70% to `game-server-1` and 30% to `game-server-2`).

Conformance Level: **Extended**

## References

- [TCPRoute Specification](/docs/concepts/api-overview/#tcproute-and-udproute)
- [GEP-735: TCP and UDP addresses matching](../gep-735/index.md) (Declined, but relevant context)
- [Gateway API Use Cases](/docs/concepts/use-cases/)

## Provisional TODOs

- Define behavior for multiple UDP routes attaching to same listener [Do we merge? reject?]
- Declare optional behaviors
  - Client IP perservation
  - Flow management [flow timeout]
  - Routing options [5 tuple, 3 tuple]
- Port range support
