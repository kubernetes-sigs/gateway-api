
# GEP-2643: TLS/SNI based routing / TLSRoute

* Issue: [#2643](https://github.com/kubernetes-sigs/gateway-api/issues/2643)
* Status: Standard

## TLDR

This GEP documents the implementation of a route type that uses the Server Name
Indication attribute (aka SNI) of a TLS handshake to make the routing decision.

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
* Implement capabilities for [ESNI](https://www.cloudflare.com/learning/ssl/what-is-encrypted-sni/)
* Implement capabilities for [ECH](https://blog.cloudflare.com/encrypted-client-hello/)
* Extend TLSRouteRule to support hostname being a matcher

## Non-Goals
* Provide an interface for users to define different listeners or ports for the
`TLSRoute` - This will be covered by the `ListenerSet` enhancement.
* When using `TLSRoute` passthrough, support `PROXY` protocol on the gateway listener.
* When using `TLSRoute` passthrough, support `PROXY` protocol on the communication
between the Gateway and the backend.
* Support TLS over UDP or UDP-based protocols such as `QUIC`.
* Support attributes other than the SNI/hostname for the route selection
* Support multiplexing (HTTPS and TLS protocols) on the same Listener - This should be covered
on a different GEP

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
to determine what backend should be used for a given request. `TLSRoute` thereby
enables the use of a single gateway listener to handle traffic for multiple routes.

## API

```golang
// TLSRouteSpec defines the expected behavior of a TLSRoute.
// A TLSRoute MUST be attached to a Listener of protocol TLS.
// Core: The listener CAN be of type Passthrough
// Extended: The listener CAN be of type Terminate 
type TLSRouteSpec struct {
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
  // Listener. If the TLSRoute does not match any Listener on the parent, the 
  // implementation must raise an 'Accepted' Condition with a status of
  // `False` in the corresponding RouteParentStatus.
  // 
  // A Listener MUST be of type TLS when a TLSRoute attaches to it. The 
  // implementation MUST raise an 'Accepted' Condition with a status of
  // `False` in the corresponding RouteParentStatus with the reason 
  // of "UnsupportedValue" in case a Listener of the wrong type is used.
  // Core: A TLS listener CAN have mode Passthrough
  // Extended: A TLS listener CAN have mode Terminate 
  // </gateway:util:excludeFromCRD>
  // +required
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=16
  Hostnames []Hostname `json:"hostnames,omitempty"`
  // Rules is a list of TLS matchers and actions.
  //
  // +required
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=1
  Rules []TLSRouteRule `json:"rules,omitempty"`
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

## Request flow

Following are some of the requests flow covered by TLSRoute

### TLSRoute + TLS Passthrough
In this workflow, the TLS traffic will be matched against the `SNI attribute` of 
the request and then directed to the backends.

This workflow MUST be supported on Core support level.

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

A typical [north/south](https://gateway-api.sigs.k8s.io/concepts/glossary/#northsouth-traffic)
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

This workflow CAN be supported on Extended support level.

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

A typical [north/south](https://gateway-api.sigs.k8s.io/concepts/glossary/#northsouth-traffic)
API request flow for a gateway implemented using both `TLSRoute` is:
* A client makes a request to `rtmps://rtmp.example.com:443`.
* DNS resolves the name to a `Gateway` address.
* The reverse proxy receives the request on a `Listener` and uses the
[Server Name Indication](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
attribute to match an `TLSRoute`.
* The reverse proxy terminates the TLS negotiation on the `Gateway`.
* The reverse proxy passes unencrypted request to one or more objects,
i.e. `Service`, in the cluster based on `backendRefs` rules of the `TLSRoute`.

### TLSRoute + Mixed Termination
In this workflow, the TLS traffic will be matched against the `SNI attribute` of 
the request, and based on the SNI attribute be directed to the backends on Passthrough
mode or be terminated on the `Gateway` and passed unencrypted to the backends.

This workflow CAN be supported on `Extended` support level.

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

A typical [north/south](https://gateway-api.sigs.k8s.io/concepts/glossary/#northsouth-traffic)
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

## Conflict management and precedences

The following conflict situations are covered by TLSRoute cases:

* If a hostname is specified by both the `Listener` and `TLSRoute`, there must
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
  the later must not be considered for a match.  
  * In any of the cases above, the `TLSRoute` should have a `Condition` of `Accepted=True`.

## Conformance Details

###  Feature Names

* TLSRoute
* TLSRouteTermination

### Conformance tests

| Description | Outcome | Notes |
| :---- | :---- | :---- |
| A single TLSRoute in the gateway-conformance-infra namespace attaches to a Gateway in the same namespace | A request to a hostname served by the TLSRoute should be passthrough directly to the backend. Check if the termination happened, if no additional Gateway header was added | Already [implemented](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/tests/tlsroute-simple-same-namespace.go) needs review |
| A single TLSRoute in the gateway-conformance-infra namespace, with a backendRef in another namespace without valid ReferenceGrant, should have the ResolvedRefs condition set to False | TLSRoute conditions must have a Not Permitted status | Already [implemented](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/tests/tlsroute-invalid-reference-grant.go) needs review |
| A TLSRoute trying to attach to a gateway without a “tls” listener should be rejected  | Condition on the TLSRoute that it was rejected (discuss with community the right condition to be used here) | [https://github.com/kubernetes-sigs/gateway-api/issues/1579](https://github.com/kubernetes-sigs/gateway-api/issues/1579)  |
| A TLSRoute with a hostname that does not match the Gateway hostname should be rejected (eg.: route with hostname [www.example.com](http://www.example.com), gateway with hostname www1.example.com) | Condition on the TLSRoute that it was rejected |  |
| A TLSRoute with an IP on its hostname should be rejected | Condition on the TLSRoute that it was rejected |  |
| A Gateway containing a Listener of type TLS/Passthrough and a Listener of type TLS/Terminate should be accepted, and should direct the requests to the right TLSRoute | Being able to do a request to a TLS route being terminated on gateway (eg.: terminated.example.tld/xpto) and to a TLS Passthrough route on the same gateway, but different host (passthrough.example.tld) | Only as `Extended` support |
| A Gateway with \*.example.tld on a TLS listener should allow a TLSRoute with hostname some.example.tld to be attached to it (and the same, but with a non wildcard hostname) | TLSRoute should be able to attach to the Gateway using the matching hostname, a request should succeed | [https://github.com/kubernetes-sigs/gateway-api/issues/1579](https://github.com/kubernetes-sigs/gateway-api/issues/1579)  |
| For a [Listener](https://gateway-api.sigs.k8s.io/reference/spec/#listener) setting mode: "terminate", TLSRoute should be present in [ListenerStatus.SupportedKinds](https://gateway-api.sigs.k8s.io/reference/spec/#listenerstatus) in case TLSRoute termination is supported |  | [https://github.com/kubernetes-sigs/gateway-api/issues/1579](https://github.com/kubernetes-sigs/gateway-api/issues/1579)  |
| A Gateway containing a Listener of type TLS/Passthrough and a Listener of type HTTPS/Terminate should not be accepted, and should mark all such listeners as `Conflicted` with `Reason: ListenerConflict` | Expose that multiplexing / using different protocols is not allowed/supported |  |

Pending conformance verifications:

* [https://github.com/kubernetes-sigs/gateway-api/issues/3466](https://github.com/kubernetes-sigs/gateway-api/issues/3466)  
* [https://github.com/kubernetes-sigs/gateway-api/issues/2153](https://github.com/kubernetes-sigs/gateway-api/issues/2153)

## Alternatives considered

### Hostname as a rule matcher
Moving the `hostname` to a matcher inside `.spec.rules`  was considered as part of this GEP. 
This alternative will be considered as a long term discussion, as of by the time of the creation of 
this GEP moving the this field to another place would be a breaking change, and duplicating
the field can be considered too complex.

## References
* [Existing API](https://github.com/kubernetes-sigs/gateway-api/blob/main/apis/v1alpha3/tlsroute_types.go)
* [TLS Terminology - Needs update](https://github.com/kubernetes-sigs/gateway-api/blob/d28cd59d37887be07b879f098cff7b14a87c0080/geps/gep-2907/index.md?plain=1#L29)
* [TLSRoute promotion issue](https://github.com/kubernetes-sigs/gateway-api/issues/3165)
* [TLSRoute intersecting hostnames issue](https://github.com/kubernetes-sigs/gateway-api/issues/3541)
* [TLSRoute termination feature request](https://github.com/kubernetes-sigs/gateway-api/issues/2111)
* [GatewayAPI TLS Use Cases](https://docs.google.com/document/d/17sctu2uMJtHmJTGtBi_awGB0YzoCLodtR6rUNmKMCs8)
