---
title: "UDPRoute"
weight: 7
---

{{< details title="Standard Channel since v1.6.0" color="success" >}}
The `UDPRoute` resource is GA and has been part of the Standard Channel since
`v1.6.0`. For more information on release channels, refer to our [versioning
guide](/docs/concepts/versioning/).
{{< /details >}}

[UDPRoute][udproute] is a Gateway API type for routing UDP traffic from a
Gateway listener to a backend, i.e. Service. When combined with a Gateway
listener, it forwards datagrams arriving on the listener's port to the backends
specified in the route.

UDPRoute is intentionally minimal. UDP carries no application-layer metadata
that the Gateway can match on, so traffic is classified only by the listener's
`protocol:port`. As a result, UDPRoute has no hostnames, matches, or filters,
and a rule consists only of the backends to forward traffic to.

## Background

While many routing cases can be handled at L7 with [HTTPRoute][httproute] or
[GRPCRoute][grpcroute], a large class of workloads speak plain UDP that fits
none of the L7 models. Common examples include:

- DNS (Domain Name System).
- VoIP and real-time communications.
- Gaming protocols.
- Streaming media (RTP/RTCP).
- IoT and telemetry protocols.

Without UDPRoute, these workloads must rely on implementation-specific
extensions or fall back to traditional Kubernetes Service resources. UDPRoute
lets you manage them with the Gateway API and consolidate load balancing under a
single Gateway rather than one load balancer per Service.

## Spec

The specification of a UDPRoute consists of:

- [ParentRefs][parentRef] - Define which Gateways this Route wants to be
  attached to.
- [Rules][udprouterule] - Define a list of rules to perform actions against
  matching UDP datagrams. For UDPRoute this is limited to which
  [backendRefs][backendRef] should be used. A UDPRoute may contain a single
  rule.

### Attaching to Gateways

Each Route includes a way to reference the parent resources it wants to attach
to. In most cases, that's going to be Gateways, but there is some flexibility
here for implementations to support other types of parent resources.

The following example shows how a Route would attach to the `acme-lb` Gateway:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: UDPRoute
metadata:
  name: udproute-example
spec:
  parentRefs:
    - name: acme-lb
```

Note that the target Gateway needs to allow UDPRoutes from the route's
namespace to be attached for the attachment to be successful.

Because the example above specifies neither a `sectionName` nor a `port`, the
UDPRoute attaches to every UDP listener on the `acme-lb` Gateway. Listeners
using other protocols are not affected.

You can also attach routes to specific sections of the parent resource.
For example, let's say that the `acme-lb` Gateway includes the following
listeners:

```yaml
  listeners:
  - name: coredns
    protocol: UDP
    port: 5300
    ...
  - name: gaming
    protocol: UDP
    port: 7777
    ...
```

You can bind a route to listener `coredns` only, using the `sectionName` field
in `parentRefs`:

```yaml
spec:
  parentRefs:
    - name: acme-lb
      sectionName: coredns
```

Alternatively, you can achieve the same effect by using the `port` field,
instead of `sectionName`, in the `parentRefs`:

```yaml
spec:
  parentRefs:
    - name: acme-lb
      port: 5300
```

However, when binding Routes by port number, Gateway admins will no longer have
the flexibility to switch ports on the Gateway without also updating the Routes.
This approach should only be used when a Route must bind to a specific port
number, rather than to named listeners whose ports may change.

### Rules

Rules define the list of actions to be taken with the traffic. A UDPRoute may
contain a single rule.

#### BackendRefs

BackendRefs defines API objects where matching datagrams should be sent. At
least one backendRef must be specified.

The following example forwards UDP datagrams arriving on the Gateway listener
to service "my-foo-service" on port `6000`:

{{< readfile file="/examples/standard/udp-routing/udp-route.yaml" code="true" lang="yaml" >}}

This UDPRoute attaches to the Gateway UDP listener named `foo`, as defined
below:

{{< readfile file="/examples/standard/udp-routing/gateway.yaml" code="true" lang="yaml" >}}

When multiple backends are specified, datagrams are distributed across them
according to their `weight`. Reference the [backendRef][backendRef] API
documentation for additional details on `weight` and other fields. Weight
support is Extended for UDPRoute.

If a backendRef is invalid, for example it refers to a nonexistent resource or a
Service with no endpoints, the implementation MUST actively drop datagrams
destined for that backend, respecting weight.

### Mixing UDP and TCP

UDP and TCP are distinct transport protocols, so an implementation MAY support a
UDP and a TCP listener on the same port without conflict. A common use case is
exposing the same service over both protocols, for example DNS on port 53, where
UDP traffic is routed by a UDPRoute and TCP traffic by a [TCPRoute][tcproute].

## Status

Status defines the observed state of UDPRoute.

### RouteStatus

RouteStatus defines the observed state that is required across all route types.

#### Parents

Parents define a list of the Gateways (or other parent resources) that are
associated with the UDPRoute, and the status of the UDPRoute with respect to
each of these Gateways. When a UDPRoute adds a reference to a Gateway in
parentRefs, the controller that manages the Gateway should add an entry to this
list when the controller first sees the route and should update the entry as
appropriate when the route is modified.

The following example indicates UDPRoute "udp-example" has been accepted by
Gateway "gw-example" in namespace "gw-example-ns":

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: UDPRoute
metadata:
  name: udp-example
...
status:
  parents:
    - parentRef:
        name: gw-example
        namespace: gw-example-ns
      conditions:
        - type: Accepted
          status: "True"
```

## Merging

Multiple UDPRoutes can be attached to a single Gateway resource. However,
because a UDP listener has no hostname, [Server Name Indication (SNI)][sni], or
path to distinguish between datagrams, attaching multiple UDPRoutes to the same
listener results in only one route effectively receiving traffic. All attached
routes are `Accepted`, but following the general Gateway API route precedence
rules, only the oldest route (by `metadata.creationTimestamp`, then
alphabetically by `namespace/name`) receives traffic.

[udproute]: /reference/api-spec/main/spec/#udproute
[udprouterule]: /reference/api-spec/main/spec/#udprouterule
[tcproute]: /reference/api-types/tcproute/
[httproute]: /reference/api-types/httproute/
[grpcroute]: /reference/api-types/grpcroute/
[backendRef]: /reference/api-spec/main/spec/#backendref
[parentRef]: /reference/api-spec/main/spec/#parentreference
[sni]: https://datatracker.ietf.org/doc/html/rfc6066#section-3
