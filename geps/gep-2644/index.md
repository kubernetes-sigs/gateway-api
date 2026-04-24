# GEP-2644: TCPRoute

- Issue: [#2644](https://github.com/kubernetes-sigs/gateway-api/issues/2644)
- Status: Provisional

(See definitions in [GEP States](../overview.md#gep-states).)

## TLDR

Gateway API needs a first-class Route type for TCP because a meaningful class of Kubernetes workloads cannot be represented by the standard Gateway API routing model without it.
HTTPRoute only addresses a single class of TCP-based traffic (HTTP/1.1 and HTTP/2 over TCP), leaving the broader universe of TCP workloads without a portable routing API.
Without TCPRoute, these TCP based workloads must either fall back to Service-based exposure or rely on implementation-specific APIs,
which prevents Gateway API from serving as a common and portable configuration model for them.
TCPRoute standardizes the minimal interoperable API surface for exposing TCP workloads through Gateway API.

This GEP retroactively documents the rationale, scope, and design constraints of the existing TCPRoute resource.

## Goals

- Define a standard Gateway API Route resource that matches TCP traffic based off the inbound port and forwards it to a TCP based backend.
- Standardize the minimal interoperable forwarding model for TCP traffic: listener attachment and backend forwarding.
- Support forwarding to one or more backends, including weighted backend selection where implemented.
- Provide a stable baseline for evaluating future TCPRoute enhancements.

## Non-Goals

- Define rich TCP-specific matching semantics such as address matching or payload inspection.
- Define TLS termination behavior at the Gateway. TLS-aware L4 routing is covered by `TLSRoute`.
- Define SNI-based or hostname-based routing. Use `TLSRoute` for SNI routing.
- Define HTTP or gRPC routing behavior. Use `HTTPRoute` or `GRPCRoute` for L7 routing.

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

### Example Usage

#### Basic TCP Service Routing (PostgreSQL)

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

#### Multiple Backend Distribution

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TCPRoute
metadata:
  name: kafka-route
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

### Mixing Protocols

A common use-case is to expose the same service over TCP and UDP. An Implementation MAY listen for both TCP and UDP traffic
utilizing the same Listener port. In this example, all TCP traffic MUST be routed to the TCP route and all UDP traffic
must be routed to the UDP route.

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
	// SupportTCPRoute option indicates support for TCPRoute
	SupportTCPRoute FeatureName = "TCPRoute"
```

They will validate the following scenarios:

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
         status: "False"
         reason: BackendNotFound
       - type: ResolvedRefs
         status: "False"
         reason: BackendNotFound
     ```

   - No TCP traffic is forwarded for this route.

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

   - TCP connections sent to port 9092 are distributed across backends respecting the configured weights (approximately 70% to `kafka-broker-1` and 30% to `kafka-broker-2`).

Conformance Level: **Core**

## References

- [UDPRoute Specification](https://gateway-api.sigs.k8s.io/reference/spec/#udproute)
- [TLSRoute Specification](https://gateway-api.sigs.k8s.io/reference/spec/#tlsroute)
- [GEP-735: TCP and UDP addresses matching](../gep-735/index.md) (Declined, but relevant context)
- [Gateway API Use Cases](https://gateway-api.sigs.k8s.io/concepts/use-cases/)

## Provisional TODOs

- Define behavior for multiple TCP routes attaching to same listener [Do we merge? reject?]
- Declare optional behaviors
  - Client IP preservation (e.g. PROXY protocol)
  - Connection draining and timeout semantics
  - Routing options [5 tuple, 3 tuple]
- Port range support
