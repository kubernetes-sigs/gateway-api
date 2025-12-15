
# GEP-2643: TLS/SNI based routing / TLSRoute

* Issue: [#2643](https://github.com/kubernetes-sigs/gateway-api/issues/2643)
* Status: Standard

## TLDR

This GEP documents the implementation of a route type that uses the Server Name
Indication attribute (aka SNI) of a TLS handshake to chose the route destination.

While this feature is also known sometimes as TLS passthrough, where after the 
server name is identified, the gateway does a full encrypted passthrough of the 
communication, this GEP will also cover cases where a TLS communication is 
terminated on the Gateway before being passed to a backend.

## Goals

* Provide a TLS route type, based on the SNI identification.
* Support both Passthrough (core support) and Termination (extended support) TLS modes for TLSRoute
* Provide load balancing of a TLS route, allowing a user to define a route, 
based on the SNI identification that should pass the traffic to N load balanced backends

## Longer Term Goals

* Implement capabilities for [ECH](https://blog.cloudflare.com/encrypted-client-hello/)
* Extend TLSRouteRule to support hostname being a matcher

## Non-Goals

* Provide an interface for users to define different listeners or ports for the
`TLSRoute` - This will be covered by the [ListenerSet enhancement](../gep-1713/index.md).
* Describe re-encryption capabilities when TLS Termination is being used - This will
be proposed in [GEP-4274](https://github.com/kubernetes-sigs/gateway-api/issues/4274), extending
`BackendTLSPolicy` to allow `TLSRoute`.
* When using `TLSRoute` passthrough, support `PROXY` protocol on the gateway listener.
* When using `TLSRoute` passthrough, support `PROXY` protocol on the communication
between the Gateway and the backend.
* Support TLS over UDP or UDP-based protocols such as `QUIC`.
* Support attributes other than the SNI/hostname for the route selection
* Support multiplexing (HTTPS and TLS protocols) on the same Listener - This should be covered
on [GEP-4271](https://github.com/kubernetes-sigs/gateway-api/issues/4271)
* Define timeout for TLSRoute - This will be proposed in [GEP-4280](https://github.com/kubernetes-sigs/gateway-api/issues/4280)

## Introduction

While many application routing cases can be implemented using HTTP/L7 matching
(the tuple `protocol:hostname:port:path`), there are some specific cases where direct,
encrypted communication to the backend may be required, without further assertion. For example:

* A backend that is TLS based but not HTTP based (e.g., a Kafka service, or a
Postgres service, with its listener being TLS enabled).
* Some WebRTC solutions.
* Backends that can require direct client-certificate authentication (e.g., OAuth).

For the example cases above, it is desired that the routing is made as a passthrough
mode, where the `Gateway` passes the packets to the backend without terminating TLS.

On some other cases, it is desired that the termination is done on the `Gateway`
and the proxy passes the unencrypted packets to the backend without caring about 
attributes other than a TCP communication.

While this routing is possible using other mechanisms such as a `TCPRoute` or a
Kubernetes service of type `LoadBalancer`, it may be desired to have a single point
of traffic convergence (like one single `LoadBalancer`) for reasons like cost,
traffic control, etc.

An implementation of `TLSRoute` achieves this end using the
[server_name TLS attribute](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
to determine what backend should be used for a given request.

An example scenario on how TLSRoute may help:

* [Ana], the application developer has 3 services that are Postgres. They all listen on port 5432. They all are TLS enabled. [Ana] can create a service of type LoadBalancer for each of them, but this will incur in costs.
* [Ana] realizes she could use TCPRoute, but then she would need different ports for each service on the listener (because `TCPRoute` traffic is classified only by the dstip:dstport)
* [Ana] figures out she can use the `SNI attribute` to listen to the traffic all on the same port (5432) but classify it based on the SNI attribute. This way the same port can be used, but given this is a TLS traffic, the requested hostname would be used to route the traffic to the right Postgres instance.

[Ana]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana


## API

```golang
// TLSRouteSpec defines the expected behavior of a TLSRoute.
// A TLSRoute MUST be attached to a Listener of protocol TLS.
// Core: The listener CAN be of type Passthrough
// Extended: The listener CAN be of type Terminate 
type TLSRouteSpec struct {
  CommonRouteSpec `json:",inline"`
  
  // Hostnames defines a set of SNI hostnames that should match against the
  // SNI attribute of TLS ClientHello message in TLS handshake. This matches
  // the RFC 1123 definition of a hostname with 2 notable exceptions:
  //
  // 1. IPs are not allowed in SNI hostnames per RFC 6066.
  // 2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard
  //    label must appear by itself as the first label.
  //
  // <gateway:util:excludeFromCRD>
  // If a hostname is specified by both the Listener and TLSRoute, there
  // must be at least one intersecting hostname for the TLSRoute to be
  // attached to the Listener. For example:
  //
  // * A Listener with `test.example.com` as the hostname matches TLSRoutes
  //   that have specified at least one of `test.example.com` or
  //   `*.example.com`.
  // * A Listener with `*.example.com` as the hostname matches TLSRoutes
  //   that have specified at least one hostname that matches the Listener
  //   hostname. For example, `test.example.com` and `*.example.com` would both
  //   match. On the other hand, `example.com` and `test.example.net` would not
  //   match.
  // * A listener with `something.example.com` as the hostname matches a 
  //   TLSRoute with hostname `*.example.com`.
  //
  // If both the Listener and TLSRoute have specified hostnames, any
  // TLSRoute hostnames that do not match any Listener hostname MUST be
  // ignored. For example, if a Listener specified `*.example.com`, and the
  // TLSRoute specified `test.example.com` and `test.example.net`,
  // `test.example.net` must not be considered for a match.
  //
  // If both the Listener and TLSRoute have specified hostnames, and none
  // match with the criteria above, then the TLSRoute is not accepted for that 
  // Listener. If the TLSRoute does not match any Listener on its parent, the 
  // implementation must raise an 'Accepted' Condition with a status of
  // `False` in the corresponding RouteParentStatus.
  // 
  // A Listener MUST be have protocol set to TLS when a TLSRoute attaches to it. The 
  // implementation MUST raise an 'Accepted' Condition with a status of
  // `False` in the corresponding RouteParentStatus with the reason 
  // of "UnsupportedValue" in case a Listener of the wrong type is used.
  // Core: Listener with `protocol` `TLS` and `tls.mode` `Passthrough`.
  // Extended: Listener with `protocol` `TLS` and `tls.mode` `Terminate`. The feature name for this Extended feature is `TLSRouteTermination`.
  // </gateway:util:excludeFromCRD>
  // +required
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=16
  // +kubebuilder:validation:XValidation:message="Hostnames cannot contain an IP",rule="self.all(h, !isIP(h))"
  // +kubebuilder:validation:XValidation:message="Hostnames must be valid based on RFC-1123",rule="self.all(h, !h.contains('*') ? !format.dns1123Subdomain().validate(h).hasValue() : true )"
	// +kubebuilder:validation:XValidation:message="Wildcards on hostnames must be the first label, and the rest of hostname must be valid based on RFC-1123",rule="self.all(h, h.contains('*') ? (h.startsWith('*.') && !format.dns1123Subdomain().validate(h.substring(2)).hasValue()) : true )"
  Hostnames []Hostname `json:"hostnames,omitempty"`
  // Rules is a list of TLS matchers and actions.
  //
  // +required
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=1
  Rules []TLSRouteRule `json:"rules,omitempty"`
}

// TLSRouteStatus defines the observed state of TLSRoute
type TLSRouteStatus struct {
	RouteStatus `json:",inline"`
}

type TLSRouteRule struct {
  // Name is the name of the route rule. This name MUST be unique within a Route if it is set.
  //
  // +optional
  Name *SectionName `json:"name,omitempty"`
  // BackendRefs (same as other BackendRef here)
  // +required
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=16
  BackendRefs []BackendRef `json:"backendRefs,omitempty"`
}
```

A validation change is required on `Gateway.spec.listeners` as users MUST define what
mode should be used on a TLS Listener, otherwise the mode may be defaulted to `Terminate`
per the current API specification:

```golang
type GatewaySpec struct {

  ...
  // +kubebuilder:validation:XValidation:message="tls mode must be set for protocol TLS",rule="self.all(l, (l.protocol == 'TLS' ? has(l.tls) && has(l.tls.mode) && l.tls.mode != '' : true))"
  ...
  Listeners []Listener `json:"listeners"`

}
```

## Request flow

Following are some of the requests flow covered by TLSRoute, and the expected 
behavior:

### TLSRoute + TLS Passthrough
In this workflow, the TLS traffic will be matched against the `SNI attribute` of 
the request and then directed to the backends.

This is the base use case for the TLSRoute object, and is included in the base TLSRoute feature, `TLSRoute`. To put this another way, this feature has Core support within the TLSRoute object.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway-tlsroute
spec:
  gatewayClassName: "my-class"
  listeners:
  - name: somelistener
    port: 443
    protocol: TLS
    hostname: "*.example.com"
    allowedRoutes:
      namespaces:
        from: Same
      kinds:
      - kind: TLSRoute
    tls:
      mode: Passthrough
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: TLSRoute
metadata:
  name: my-tls-route
spec:
  parentRefs:
  - name: gateway-tlsroute
  hostnames:
  - foo.example.com
  rules:
  - backendRefs:
    - name: tls-backend
      port: 443
```

A typical [north/south](../../concepts/glossary.md#northsouth-traffic)
API request flow for a gateway implemented using a `TLSRoute` is:

* A client makes a request to https://foo.example.com.
* DNS resolves the name to a `Gateway` address.
* The reverse proxy receives the request on a `Listener` and uses the
[Server Name Indication](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
attribute to match an `TLSRoute`.
* The reverse proxy passes through the request directly to one object,
i.e. `Service`, in the cluster based on `backendRefs` rules of the `TLSRoute`.

### TLSRoute + TLS Termination
In this workflow, the TLS traffic will be matched against the `SNI attribute` of 
the request and terminated on the `Gateway`. 

This use case is Extended for TLSRoute, and so MAY be supported, with the featurename `TLSRouteTermination`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway-tlsroute
spec:
  gatewayClassName: "my-class"
  listeners:
  - name: my-terminated-listener
    port: 443
    protocol: TLS
    hostname: "*.example.com"
    allowedRoutes:
      namespaces:
        from: Same
      kinds:
      - kind: TLSRoute
    tls:
      mode: Terminate
      certificateRefs:
      - name: listener
        kind: Secret
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: TLSRoute
metadata:
  name: my-rtmp-route
spec:
  parentRefs:
  - name: gateway-tlsroute
  hostnames:
  - rtmp.example.com
  rules:
  - backendRefs:
    - name: rtmp-backend
      port: 12345
```

A typical [north/south](../../concepts/glossary.md#northsouth-traffic)
API request flow for a gateway implemented using both `TLSRoute` is:

* A client makes a request to `rtmps://rtmp.example.com:443`.
* DNS resolves the name to a `Gateway` address.
* The reverse proxy receives the request on a `Listener` and uses the
[Server Name Indication](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
attribute to match an `TLSRoute`.
* The reverse proxy terminates the TLS negotiation on the `Gateway`.
* The reverse proxy passes unencrypted TCP packets to one or more objects,
i.e. `Service`, in the cluster based on `backendRefs` rules of the `TLSRoute`.
It is not valid to make any assumptions about the content of the traffic in this case,
it MUST be treated as _only_ TCP traffic.

In case the implementation does not support this TLSRoute termination, it MUST 
add the Listener status as:

```yaml
listeners:
  - attachedRoutes: 0
    name: my-terminated-listener
    supportedKinds: []
```

Also, in case the implementation does not support TLSRoute termination, the TLSRoute 
resource status MUST be updated to contain the following condition:

```yaml
status:
  parents:
  - conditions:
    - reason: NotAllowedByListeners
      status: "False"
      type: Accepted
```

### TLSRoute + Mixed Termination
In this workflow, the TLS traffic will be matched against the `SNI attribute` of 
the request, and based on the SNI attribute be directed to the backends on Passthrough
mode or be terminated on the `Gateway` and passed unencrypted to the backends.

This use case is Extended for TLSRoute, and so MAY be supported, with the feature name `TLSRouteMixedMode`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway-tlsroute
spec:
  gatewayClassName: "my-class"
  listeners:
  - name: terminatelistener
    port: 443
    protocol: TLS
    hostname: "rtmp.example.com"
    allowedRoutes:
      namespaces:
        from: Same
      kinds:
      - kind: TLSRoute
    tls:
      mode: Terminate
      certificateRefs:
      - name: listener
        kind: Secret
  - name: passthroughlistener
    port: 443
    protocol: TLS
    hostname: "direct.example.com"
    allowedRoutes:
      namespaces:
        from: Same
      kinds:
      - kind: TLSRoute
    tls:
      mode: Passthrough
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: TLSRoute
metadata:
  name: my-tls-route
spec:
  parentRefs:
  - name: gateway-tlsroute
  hostnames:
  - direct.example.com
  rules:
  - backendRefs:
    - name: tls-backend
      port: 443
---
apiVersion: gateway.networking.k8s.io/v1alpha3
kind: TLSRoute
metadata:
  name: my-rtmp-route
spec:
  parentRefs:
  - name: gateway-tlsroute
  hostnames:
  - rtmp.example.com
  rules:
  - backendRefs:
    - name: rtmp-backend
      port: 12345
```

A typical [north/south](../../concepts/glossary.md#northsouth-traffic)
API request flow for a gateway implemented using a `TLSRoute` is:

* A client makes a request to `rtmps://rtmp.example.com:443`.
* DNS resolves the name to a `Gateway` address.
* The reverse proxy receives the request on a `Listener` and uses the
[Server Name Indication](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
attribute to match an `TLSRoute`.
* The reverse proxy terminates the TLS negotiation on the `Gateway`.
* The reverse proxy passes unencrypted request to one or more objects,
i.e. `Service`, in the cluster based on `backendRefs` rules of the `TLSRoute`.
* A new request is made to `direct.example.com` and the same identification flow
happens.
* The reverse proxy receiving the request identifies that this is a Passthrough request
and passes through the request directly to one or more objects,
i.e. `Service`, in the cluster based on `backendRefs` rules of the `TLSRoute`.

In case mixed mode is not supported, the Listeners MUST be marked as conflicted with 
the reason `ProtocolConflict`.

## Conflict management and precedences

A conflict can happen when two or more distinct listeners on a Gateway definition
have conflicting behavior.

As an example, in case a listener of protocol `TLS` and a listener of protocol `HTTP`
are both specified on the same Gateway and port, this generates a conflicting situation.

Following are the possible extra conflict situations for TLSRoute and how the controller should react to it.

### Two listeners of type TLS and intersecting hostnames

In this case, there is no conflict as the most specific hostname MUST be chosen:

```yaml
spec:
  listeners:
  - hostname: app.user1.tld
    name: listener1
    port: 443
    protocol: TLS
    tls:
      mode: Passthrough
  - hostname: "*.user1.tld"
    name: listener2
    port: 443
    protocol: TLS
    tls:
      mode: Passthrough
status:
 listeners:
  - name: listener1
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: NoConflicts
      status: "False"
      type: Conflicted
  - name: listener2
    conditions:
    - reason: Accepted
      status: "True"
      type: Accepted
    - reason: NoConflicts
      status: "False"
      type: Conflicted
```

### Two listeners of type TLS with different termination modes

As the support for this feature is `Extended`, in case the implementation
does not support mixed modes it MUST mark the listeners as conflicted with the 
reason `ProtocolConflict`:

```yaml
spec:
  listeners:
  - hostname: "*.user2.tld"
    name: listener1
    port: 443
    protocol: TLS
    tls:
      mode: Passthrough
  - hostname: "*.user1.tld"
    name: listener2
    port: 443
    protocol: TLS
    tls:
      mode: Terminate
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

## Hostname intersection between Gateway and TLSRoutes

The following are hostname intersection cases between the Gateway and TLSRoutes

* If a hostname is specified by both the `Listener` and `TLSRoute`, there MUST
be at least one intersecting hostname for the `TLSRoute` to be attached to the
`Listener`.
  * A `Gateway listener` with `test.example.com` as the hostname matches a `TLSRoute` that
  specifies at least one of `test.example.com` or `*.example.com`
  * A `Gateway listener` with `*.example.com` as the hostname matches a `TLSRoute` that
  specifies at least one hostname that matches the `Gateway listener` hostname.
  For example, `test.example.com` and `*.example.com` would both match. On the
  other hand, `example.com` and `test.example.net` would not match.
  * If both the `Gateway listener` and `TLSRoute` specify hostnames, any `TLSRoute`
  hostnames that do not match the `Gateway listener` hostname MUST be ignored
  for that Listener. For example, if a `Gateway listener` specified `*.example.com`,
  and the `TLSRoute` specified `test.example.com` and `test.example.net`,
  the later must not be considered for a match. In case none of the hostnames specified
  on the `TLSRoute` matches the Gateway hostnames, the route MUST have a `Condition` of 
  `Accepted=False` with the reason `NoMatchingListenerHostname`.
  * In any of the cases above, the `TLSRoute` should have a `Condition` of `Accepted=True`.

## Conformance Details

###  Feature Names

* TLSRoute
* TLSRouteTermination
* TLSRouteMixedMode

### Conformance tests

| Description | Outcome | Features |
| :---- | :---- | :---- |
| A single TLSRoute in the gateway-conformance-infra namespace attaches to a Gateway in the same namespace. <br/> Code: [Simple] <br/>Issue: [TLSRoute conformance] | A request to a hostname served by the TLSRoute should be passthrough directly to the backend. Check if the termination happened, if no additional Gateway header was added. | TLSRoute |
| A single TLSRoute in the gateway-conformance-infra namespace, with a backendRef in another namespace without valid ReferenceGrant, should have the ResolvedRefs condition set to False. <br/> Code: [ReferenceGrant] and the request to the route MUST fail [Invalid ReferenceGrant Request] | TLSRoute conditions must have a `RefNotPermitted` status. A request to the route MUST fail | TLSRoute + ReferenceGrant |
| A TLSRoute trying to attach to a gateway without a “tls” listener MUST be rejected  | TLSRoute should have a parent condition of type `Accepted=False` with Reason `NoMatchingParent`. Request to the route MUST fail. | TLSRoute |
| A TLSRoute with a hostname that does not match the Gateway hostname MUST be rejected (eg.: route with hostname [www.example.com](http://www.example.com), gateway with hostname www1.example.com) <br/> Issue: [TLSRoute conformance] | Condition on the TLSRoute parent of type `Accepted=False` with Reason `NoMatchingListenerHostname`. Request to the route MUST fail. | TLSRoute |
| A Gateway containing a Listener of type TLS/Terminate MUST be accepted, and MUST direct the requests to the right TLSRoute when TLSRoute termination is supported. <br/> Issue: [Termination] | Being able to do a request to a TLS route being terminated on gateway (eg.: terminated.example.tld/xpto) | TLSRoute + TLSRouteTermination |
| A Gateway containing a Listener of type TLS/Passthrough and a Listener of type TLS/Terminate MUST be accepted, and MUST direct the requests to the right TLSRoute when mixed mode is supported. <br/> Issue: [Termination] | Being able to do a request to a TLS route being terminated on gateway (eg.: terminated.example.tld/xpto) and to a TLS Passthrough route on the same gateway, but different host (passthrough.example.tld) | TLSRoute + TLSRouteTermination + TLSRouteMixedMode |
| A Gateway containing a Listener of type TLS/Passthrough and a Listener of type TLS/Terminate MUST NOT be accepted when mixed mode is not supported | Listener condition MUST have Condition Conflicted=True with reason `ProtocolConflict`. A request to any route attached to any of the listeners MUST fail. | TLSRoute + TLSRouteTermination + !TLSRouteMixedMode |
| A Gateway with \*.example.tld on a TLS listener MUST allow a TLSRoute with hostname some.example.tld to be attached to it (and the same, but with a non wildcard hostname) <br/> Issue: [TLSRoute conformance] | TLSRoute MUST be able to attach to the Gateway using the matching hostname, a request MUST succeed | TLSRoute  |
| Expose support for TLSRoute termination <br/> Issue: [Termination] | For a [Listener] setting mode: "terminate", TLSRoute MUST be present in [ListenerStatus.SupportedKinds] in case TLSRoute termination is supported | TLSRouteTermination |
| Explicitly expose that TLSRoute termination is not supported. <br/> Issue: [Termination] | For a [Listener] setting mode: "terminate" and not being supported, the Listener entry MUST NOT be Accepted containing the condition `Accepted: False` and `Reason: UnsupportedValue`. | TLSRoute + !TLSRouteTermination |
| Explicitly expose that tls listener accepts termination for TCPRoute only. | For a [Listener] setting mode: "terminate" that supports only `TCPRoute`, a Listener entry MUST exist but only TCPRoute MUST be present in [ListenerStatus.SupportedKinds]. | TLSRoute + !TLSRouteTermination |

[Simple]: https://github.com/kubernetes-sigs/gateway-api/blob/edd7cbeac3ff1458c75ed21636af52ba1536b73a/conformance/tests/tlsroute-simple-same-namespace.go
[ReferenceGrant]: https://github.com/kubernetes-sigs/gateway-api/blob/edd7cbeac3ff1458c75ed21636af52ba1536b73a/conformance/tests/tlsroute-invalid-reference-grant.go
[Invalid ReferenceGrant Request]: https://github.com/kubernetes-sigs/gateway-api/issues/2153
[Termination]: https://github.com/kubernetes-sigs/gateway-api/issues/3466
[TLSRoute conformance]: https://github.com/kubernetes-sigs/gateway-api/issues/1579
[Listener]: ../../reference/spec.md#listener
[ListenerStatus.SupportedKinds]: ../../reference/spec.md#listenerstatus

## Changes between v1alpha3 and v1

The following changes exists between v1alpha3 and v1

### Route hostnames are validated with CEL

The following rules are now enforced using CEL validations directly on the API:

* The hostname is not an IP
* The hostname is a valid domain, per RFC-1123
* In case the hostname contains a wildcard prefix (`*.example.tld`), the remaining 
content of the hostname (`example.tld`) is a valid domain per RFC-1123
* Only a single wildcard prefix can be used.

### Explicit support for mixed modes on a TLS listener

Previously there wasn't explicit support for mixed TLS modes on a Gateway. This means 
that it wasn't clear if one listener being of type "Passthrough" and the other of type 
"Terminate" was supported.

On the current specification it was made explicit that this mode is supported on `Extended`
support level.

## Alternatives considered

### Hostname as a rule matcher
Moving the `hostname` to a matcher inside `.spec.rules`  was considered as part of this GEP. 
This alternative will be considered as a long term discussion, as of by the time of the creation of 
this GEP moving this field to another place would be a breaking change, and duplicating
the field can be considered too complex.

## References
* [Existing API](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1alpha3/tlsroute_types.go)
* [TLS Terminology - Needs update](https://github.com/kubernetes-sigs/gateway-api/blob/d28cd59d37887be07b879f098cff7b14a87c0080/geps/gep-2907/index.md?plain=1#L29)
* [TLSRoute promotion issue](https://github.com/kubernetes-sigs/gateway-api/issues/3165)
* [TLSRoute intersecting hostnames issue](https://github.com/kubernetes-sigs/gateway-api/issues/3541)
* [TLSRoute termination feature request](https://github.com/kubernetes-sigs/gateway-api/issues/2111)
* [GatewayAPI TLS Use Cases](https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8)
