# GEP-1742: HTTPRoute Timeouts

* Issue: [#1742](https://github.com/kubernetes-sigs/gateway-api/issues/1742)
* Status: Standard

(See [status definitions](../overview.md#gep-states).)

## TLDR

Create some sort of design so that Gateway API objects can be used to configure
timeouts for different types of connection.

## Goals

- Create some method to configure some timeouts.
- Timeout config must be applicable to most if not all Gateway API implementations.

## Non-Goals

- A standard API for every possible timeout that implementations may support.

## Introduction

In talking about Gateway API objects, particularly HTTPRoute, we've mentioned
timeout configuration many times in the past as "too hard" to find the common
ground necessary to make more generic configuration. This GEP intends firstly
to make this process less difficult, then to find common timeouts that we can
build into Gateway API.

For this initial round, we'll focus on Layer 7 HTTP traffic, while acknowledging
that Layer 4 connections have their own interesting timeouts as well.

The following sections will review all the implementations, then document what
timeouts are _available_ for the various data planes.

### Background on implementations

Most implementations that handle HTTPRoute objects use a proxy as the data plane
implementation, that actually forwards flows as directed by Gateway API configuration.

The following table is a review of all the listed implementations of Gateway API
at the time of writing, with the data plane they use for Layer 7, based on what information
could be found online. If there are errors here, or if the implementation doesn't
support layer 7, please feel free to correct them.

| Implementation | Data Plane       |
|----------------|------------|
| Acnodal EPIC   | Envoy      |
| Apache APISIX  | Nginx      |
| BIG-IP Kubernetes Gateway| F5 BIG-IP  |
| Cilium         | Envoy      |
| Contour        | Envoy      |
| Emissary Ingress| Envoy     |
| Envoy Gateway  | Envoy      |
| Flomesh Service Mesh | Pipy |
| Gloo Edge      | Envoy      |
| Google Kubernetes Engine (GKE) | Similar to Envoy Timeouts |
| HAProxy Ingress | HAProxy   |
| Hashicorp Consul | Envoy    |
| Istio          | Envoy      |
| Kong           | Nginx      |
| Kuma           | Envoy      |
| Litespeed      | Litespeed WebADC |
| NGINX Gateway Fabric | Nginx |
| Traefik        | Traefik    |


### Flow diagrams with available timeouts

The following flow diagrams are based off the basic diagram below, with all the
timeouts I could find included.

In general, timeouts are recorded with the setting name or similar that the data
plane uses for them, and are correct as far as I've parsed the documentation
correctly.

Idle timeouts are marked as such.

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant U as Upstream
    C->>P: Connection Started
    C->>P: Starts sending Request
    C->>P: Finishes Headers
    C->>P: Finishes request
    P->>U: Connection Started
    P->>U: Starts sending Request
    P->>U: Finishes request
    P->>U: Finishes Headers
    U->>P: Starts Response
    U->>P: Finishes Headers
    U->>P: Finishes Response
    P->>C: Starts Response
    P->>C: Finishes Headers
    P->>C: Finishes Response
    Note right of P: Repeat if connection sharing
    U->>C: Connection ended
```

#### Envoy Timeouts

For Envoy, some timeouts are configurable at either the HTTP Connection Manager
(very, very roughly equivalent to a Listener), the Route (equivalent to a HTTPRoute)
level, or the Cluster (usually close to the Service) or some combination. These
are noted in the below diagram with a `CM`, `R`, or `Cluster` prefix respectively.

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Envoy
    participant U as Upstream
    C->>P: Connection Started
    activate P
    Note left of P: transport_socket_connect_timeout for TLS
    deactivate P
    C->>P: Starts sending Request
    activate C
    activate P
    activate P
    C->>P: Finishes Headers
    note left of P: CM request_headers_timeout
		C->>P: Finishes request
    deactivate P
    activate U
    note left of U: Cluster connect_timeout
    deactivate U
    P->>U: Connection Started
		activate U
    note right of U: CM idle_timeout<br />CM max_connection_duration
    P->>U: Starts sending Request
    P->>U: Finishes Headers
    note left of P: CM request_timeout
		P->>U: Finishes request
    deactivate P
    activate U
    U->>P: Starts Response
    U->>P: Finishes Headers
		note right of U: R timeout<br/>R per_try_timeout<br/>R per_try_idle_timeout
    U->>P: Finishes Response
    deactivate U
    P->>C: Starts Response
    P->>C: Finishes Headers
    P->>C: Finishes Response
    Note left of C: CM stream_idle_timeout<br />R idle_timeout<br />CM,R max_stream_duration<br/>TCP proxy idle_timeout<br />TCP protocol idle_timeout
    deactivate C
    Note right of P: Repeat if connection sharing
    U->>C: Connection ended
    deactivate U
```

#### Nginx timeouts

Nginx allows setting of GRPC and general HTTP timeouts separately, although the
purposes seem to be roughly equivalent.

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Nginx
    participant U as Upstream
    C->>P: Connection Started
    activate P
    C->>P: Starts sending Request
    C->>P: Finishes Headers
    Note right of P: client_headers_timeout
    deactivate P
    activate P
    C->>P: Finishes request
    deactivate P
    Note right of P: client_body_timeout
    activate U
    note left of U: proxy_connect_timeout<br/>grpc_connect_timeout
    deactivate U
    P->>U: Connection Started
    Activate U
	  Activate U
    P->>U: Starts sending Request
    P->>U: Finishes Headers
		P->>U: Finishes request
    Note right of U: (between write operations)<br/>proxy_send_timeout<br/>grpc_send_timeout
    deactivate U
		activate U
    U->>P: Starts Response
    U->>P: Finishes Headers
        Note right of U: (between read operations)<br/>proxy_read_timeout<br/>grpc_read_timeout
    U->>P: Finishes Response
    deactivate U
    activate P
    P->>C: Starts Response
    P->>C: Finishes Headers
    P->>C: Finishes Response
    deactivate P
    Note left of P: send_timeout (only between two successive write operations)
    Note left of C: Repeat if connection is shared until server's keepalive_timeout is hit
    Note Right of U: upstream's keepalive_timeout (if keepalive enabled)
    U->>C: Connection ended
		deactivate U
```

#### HAProxy timeouts

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant U as Upstream

    C->>P: Connection Started
    activate U
    activate C
    activate P
    note left of P: timeout client (idle)
    C->>P: Starts sending Request
    C->>P: Finishes Headers
    C->>P: Finishes request
    note left of C: timeout http-request
    deactivate C
			activate C
    note left of C: timeout client-fin
    deactivate C
		deactivate P
    activate U
    note left of U: timeout queue<br/>(wait for available server)
    deactivate U

    P->>U: Connection Started
    activate U
    P->>U: Starts sending Request
    activate U
    P->>U: Finishes Headers
    P->>U: Finishes request

    note right of U: timeout connect
    deactivate U
    note left of U: timeout server<br/>(idle timeout)
    deactivate U
    activate U
    note left of U: timeout server-fin
    deactivate U
    U->>P: Starts Response
    U->>P: Finishes Headers
    U->>P: Finishes Response
    P->>C: Starts Response
    P->>C: Finishes Headers
    P->>C: Finishes Response
    activate C
    note left of C: timeout http-keep-alive
    deactivate C
    Note right of P: Repeat if connection sharing
    Note right of U: timeout tunnel<br/>(for upgraded connections)
    deactivate U
    U->>C: Connection ended

```

#### Traefik timeouts

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant U as Upstream
    C->>P: Connection Started
    activate U
    C->>P: Starts sending Request
    activate P
    C->>P: Finishes Headers
    Note right of P: respondingTimeouts<br/>readTimeout
    C->>P: Finishes request
    deactivate P
    P->>U: Connection Started
    activate U
    Note right of U: forwardingTimeouts<br/>dialTimeout
    deactivate U
    P->>U: Starts sending Request
    P->>U: Finishes request
    P->>U: Finishes Headers
    U->>P: Starts Response
    activate U
    note right of U: forwardingTimeouts<br/>responseHeaderTimeout
    U->>P: Finishes Headers
    deactivate U
    U->>P: Finishes Response
    P->>C: Starts Response
    activate P
    P->>C: Finishes Headers
    Note right of P: respondingTimeouts<br/>writeTimeout
    P->>C: Finishes Response
    deactivate P
    Note right of P: Repeat if connection sharing
    Note right of U: respondingTimeouts<br/>idleTimeout<br/>Keepalive connections only
    deactivate U
    U->>C: Connection ended

```
#### F5 BIG-IP Timeouts

Could not find any HTTP specific timeouts. PRs welcomed. 😊

#### Pipy Timeouts

Could not find any HTTP specific timeouts. PRs welcomed. 😊

#### Litespeed WebADC Timeouts

Could not find any HTTP specific timeouts. PRs welcomed. 😊

## API

The above diagrams show that there are many different kinds of configurable timeouts
supported by Gateway implementations: connect, idle, request, upstream, downstream.
Although there may be opportunity for the specification of a common API for more of
them in the future, this GEP will focus on the L7 timeouts in HTTPRoutes that are
most valuable to clients.

From the above analysis, it appears that most implementations are capable of
supporting the configuration of simple client downstream request timeouts on HTTPRoute
rules. This is a relatively small addition that would benefit many users.

Some implementations support configuring a timeout for individual backend requests,
separate from the overall client request timeout. This is particularly useful if a
client HTTP request to a gateway can result in more than one call from the gateway
to the destination backend service, for example, if automatic retries are supported.
Adding support for this would also benefit many users.

### Timeout values

There are 2 kinds of timeouts that can be configured in an `HTTPRouteRule`:

1. `timeouts.request` is the timeout for the Gateway API implementation to send a
    response to a client HTTP request. Whether the gateway starts the timeout before
    or after the entire client request stream has been received, is implementation dependent.
    This field is optional `Extended` support.

1. `timeouts.backendRequest` is a timeout for a single request from the gateway to a backend.
    This field is optional `Extended` support. Typically used in conjunction with retry configuration,
    if supported by an implementation.
    Note that retry configuration will be the subject of a separate GEP (GEP-1731).

```mermaid
sequenceDiagram
    participant C as Client
    participant P as Proxy
    participant U as Upstream
    C->>P: Connection Started
    note left of P: timeouts.request start time (min)
    C->>P: Starts sending Request
    C->>P: Finishes Headers
    C->>P: Finishes request
    note left of P: timeouts.request start time (max)
    P->>U: Connection Started
    note right of P: timeouts.backendRequest start time
    P->>U: Starts sending Request
    P->>U: Finishes request
    P->>U: Finishes Headers
    U->>P: Starts Response
    U->>P: Finishes Headers
    note right of P: timeouts.backendRequest end time
    note left of P: timeouts.request end time
    U->>P: Finishes Response
    note right of P: Repeat if retry
    P->>C: Starts Response
    P->>C: Finishes Headers
    P->>C: Finishes Response
    Note right of P: Repeat if connection sharing
    U->>C: Connection ended
```

Both timeout fields are [GEP-2257 Duration] values. A zero-valued timeout
("0s") MUST be interpreted as disabling the timeout; a non-zero-valued timeout
MUST be >= 1ms.

[GEP-2257 Duration]:../gep-2257/index.md

### GO

```go
type HTTPRouteRule struct {
	// Timeouts defines the timeouts that can be configured for an HTTP request.
	//
	// Support: Extended
	//
	// +optional
	// <gateway:standard>
	Timeouts *HTTPRouteTimeouts `json:"timeouts,omitempty"`

	// ...
}

// HTTPRouteTimeouts defines timeouts that can be configured for an HTTPRoute.
// Timeout values are represented with Gateway API Duration formatting.
// Specifying a zero value such as "0s" is interpreted as no timeout.
//
// +kubebuilder:validation:XValidation:message="backendRequest timeout cannot be longer than request timeout",rule="!(has(self.request) && has(self.backendRequest) && duration(self.request) != duration('0s') && duration(self.backendRequest) > duration(self.request))"
type HTTPRouteTimeouts struct {
	// Request specifies the maximum duration for a gateway to respond to an HTTP request.
	// If the gateway has not been able to respond before this deadline is met, the gateway
	// MUST return a timeout error.
	//
	// For example, setting the `rules.timeouts.request` field to the value `10s` in an
	// `HTTPRoute` will cause a timeout if a client request is taking longer than 10 seconds
	// to complete.
	//
	// This timeout is intended to cover as close to the whole request-response transaction
	// as possible although an implementation MAY choose to start the timeout after the entire
	// request stream has been received instead of immediately after the transaction is
	// initiated by the client.
	//
	// The value of Request is a Gateway API Duration string as defined by GEP-2257. When this
	// field is unspecified, request timeout behavior is implementation-specific.
	//
	// Support: Extended
	//
	// +optional
	Request *Duration `json:"request,omitempty"`

	// BackendRequest specifies a timeout for an individual request from the gateway
	// to a backend. This covers the time from when the request first starts being
	// sent from the gateway to when the full response has been received from the backend.
	//
	// An entire client HTTP transaction with a gateway, covered by the Request timeout,
	// may result in more than one call from the gateway to the destination backend,
	// for example, if automatic retries are supported.
	//
	// The value of BackendRequest must be a Gateway API Duration string as defined by
	// GEP-2257.  When this field is unspecified, its behavior is implementation-specific;
	// when specified, the value of BackendRequest must be no more than the value of the
	// Request timeout (since the Request timeout encompasses the BackendRequest timeout).
	//
	// Support: Extended
	//
	// +optional
	BackendRequest *Duration `json:"backendRequest,omitempty"`
}

// Duration is a string value representing a duration in time. The format is as specified
// in GEP-2257, a strict subset of the syntax parsed by Golang time.ParseDuration.
//
// +kubebuilder:validation:Pattern=`^([0-9]{1,5}(h|m|s|ms)){1,4}$`
type Duration string
```

### YAML

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: HTTPRoute
metadata:
  name: timeout-example
spec:
  ...
  rules:
  - backendRefs:
    - name: some-service
      port: 8080
    timeouts:
      request: 10s
      backendRequest: 2s
```

### Conformance Details

Gateway implementations can indicate support for the optional behavior in this GEP using
the following feature names:

- `HTTPRouteRequestTimeout`: supports `rules.timeouts.request` in an `HTTPRoute`.
- `HTTPRouteBackendTimeout`: supports `rules.timeouts.backendRequest` in an `HTTPRoute`.

## Alternatives

Timeouts could be configured using policy attachments or in objects other than `HTTPRouteRule`.

### Policy Attachment

Instead of configuring timeouts directly on an API object, they could be configured using policy
attachments. The advantage to this approach would be that timeout policies can be not only
configured for an `HTTPRouteRule`, but can also be added/overridden at a more fine
(e.g., `HTTPBackendRef`) or coarse (e.g. `HTTPRoute`) level of granularity.

The downside, however, is complexity introduced for the most common use case, adding a simple
timeout for an HTTP request. Setting a single field in the route rule, instead of needing to
create a policy resource, for this simple case seems much better.

In the future, we could consider using policy attachments to configure less common kinds of
timeouts that may be needed, but it would probably be better to instead extend the proposed API
to support those timeouts as well.

The default values of the proposed timeout fields could also be overridden
using policy attachments in the future. For example, a policy attachment could be used to set the
default value of `rules.timeouts.request` for all routes under an `HTTPRoute` or `Gateway`.

### Other API Objects

The new timeouts field could be added to a different API struct, instead of `HTTPRouteRule`.

Putting it on an `HTTPBackendRef`, for example, would allow users to set different timeouts for different
backends. This is a feature that we believe has not been requested by existing proxy or service mesh
clients and is also not implementable using available timeouts of most proxies.

Another alternative is to move the timeouts configuration up a level in the API to `HTTPRoute`. This
would be convenient when a user wants the same timeout on all rules, but would be overly restrictive.
Using policy attachments to override the default timeout value for all rules, as described in the
previous section, is likely a better way to handle timeout configuration above the route rule level.

## References

[GEP-2257]:../gep-2257/index.md
