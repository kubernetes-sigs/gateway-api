# TLSRoute

??? success "Standard Channel since v1.5.0"

    The `TLSRoute` resource is GA and has been part of the Standard Channel since
    `v1.5.0`. For more information on release channels, refer to our [versioning
    guide](../concepts/versioning.md).

[TLSRoute][tlsroute] is a Gateway API type for specifying routing behavior
using the [server_name TLS attribute](https://datatracker.ietf.org/doc/html/rfc6066#section-3)
to route requests to backends.

While this feature is also known sometimes as TLS passthrough, where after the server name is identified, the gateway does a full encrypted passthrough of the communication. `TLSRoute` also allows the traffic to be terminated on the Gateway before being passed to a backend. 

TLSRoute is covered by the following features, that may be reported by your implementation
* `TLSRoute` - If reported, means your implementation supports `TLSRoute` with `Passthrough` mode. Any implementation that claims to support the `TLSRoute` API MUST report this feature.
* `TLSRouteModeTerminate` - If reported, means your implementation supports `TLSRoute` with `Terminate` mode in addition to `Passthrough` mode
* `TLSRouteModeMixed` - If reported, means your implementation supports two TLS listeners with distinct modes (`Passthrough` and `Terminate`) on the same port.

## Background

While many application routing cases can be implemented using HTTP/L7 matching (the tuple protocol:hostname:port:path), there are some specific cases where direct, encrypted communication to the backend may be required without terminating TLS. For example:

* A backend that is TLS based but not HTTP based (e.g., a Kafka service, or a Postgres service, with its listener being TLS enabled).
* Some WebRTC solutions.
* Backends that require mutual TLS (mTLS) authentication with client certificates.

For the example cases above, it is desired that the routing uses passthrough mode, where the Gateway passes the packets to the backend without terminating TLS.

In other cases, it is desired that the termination is done on the Gateway and the proxy passes the unencrypted packets to the backend, treating it as a basic TCP connection.

`TLSRoute` can be used in these cases, where the traffic between the client and Gateway is encrypted and contains
the SNI (Server Name Indication), which can be used to decide which backend should be used for this request.

## Spec

The specification of a TLSRoute consists of:

- [ParentRefs][parentRef] - Define which Gateways this Route wants to be attached
  to.
- [Hostnames][hostname] (optional) - Define a list of hostnames to use for
  matching the SNI (Server Name Indication) of a TLS request.
- [Rules][tlsrouterule] - Define a list of rules to perform actions against
  matching TLS requests. For TLSRoute this is limited to which [backendRefs][backendRef]
  should be used.

### Attaching to Gateways

Each Route includes a way to reference the parent resources it wants to attach
to. In most cases, that's going to be Gateways, but there is some flexibility
here for implementations to support other types of parent resources.

The following example shows how a Route would attach to the `acme-lb` Gateway:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: tlsroute-example
spec:
  parentRefs:
  - name: acme-lb
```

Note that the target Gateway needs to allow TLSRoutes from the route's
namespace to be attached for the attachment to be successful.

For a listener of type `TLS`, defining the field `tls.mode` is mandatory.

This field can contain two values:
* Passthrough - Means the traffic will be directed to the backends while remaining encrypted
* Terminate - Means the encrypted traffic will be terminated at the Gateway, and then pass the unencrypted TCP packets to one or more backends.

You can also attach routes to specific sections of the parent resource.
For example, let's say that the `acme-lb` Gateway includes the following
listeners:

```yaml
  listeners:
  - name: passthrough
    protocol: TLS
    port: 8883
    tls:
      mode: Passthrough
    ...
  - name: terminated
    protocol: TLS
    port: 18883
    tls:
      mode: Terminate
    ...
```

You can bind a route to listener `passthrough` only, using the `sectionName` field
in `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    sectionName: passthrough
```

Alternatively, you can achieve the same effect by using the `port` field,
instead of `sectionName`, in the `parentRefs`:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    port: 8883
```

Binding to a port also allows you to attach to multiple listeners at once.
For example, binding to port `8090` of the `acme-lb` Gateway would be more
convenient than binding to the corresponding listeners by name:

```yaml
spec:
  parentRefs:
  - name: acme-lb
    sectionName: bar
  - name: acme-lb
    sectionName: baz
```

However, when binding Routes by port number, Gateway admins will no longer have
the flexibility to switch ports on the Gateway without also updating the Routes.
This approach should only be used when a Route must bind to a specific port
number, rather than to named listeners whose ports may change.

### Hostnames

Hostnames define a list of hostnames to match against the SNI (Server Name Indication)
of the TLS request. When a match occurs, the TLSRoute is selected to route the request
based on its rules. 

The SNI specification adds the following restrictions for a Hostname definition:

- the hostname MUST be a fully qualified domain name
- The usage of IPv4 and IPv6 addresses is not permitted

The following example defines hostname "my.example.com":

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: tlsroute-example
spec:
  hostnames:
  - my.example.com
```

### Rules

Rules define the list of actions to be taken with the traffic.

#### BackendRefs

BackendRefs defines API objects where matching requests should be sent. At least 
one backendRef must be specified.

The following example forwards TLS requests with the hostname `foo.example.com` to service
"foo-svc" on port `443`.

```yaml
{% include 'standard/tls-routing/tls-route.yaml' %}
```

Reference the [backendRef][backendRef] API documentation for additional details
on `weight` and other fields.

This TLSRoute attaches to the Gateway TLS listener of name `tls` as defined in:

```yaml
{% include 'standard/tls-routing/gateway.yaml' %}
```

Because this listener has its TLS mode configured as `Passthrough`, the traffic
that is routed via this listener is sent as a direct TCP stream to the backend.

In case the listener `tls-terminate` is being used, the TLS traffic is terminated and 
the TCP stream is forwarded unencrypted to the backends.

## Status

Status defines the observed state of TLSRoute.

### RouteStatus

RouteStatus defines the observed state that is required across all route types.

#### Parents

Parents define a list of the Gateways (or other parent resources) that are
associated with the TLSRoute, and the status of the TLSRoute with respect to
each of these Gateways. When an TLSRoute adds a reference to a Gateway in
parentRefs, the controller that manages the Gateway should add an entry to this
list when the controller first sees the route and should update the entry as
appropriate when the route is modified.

The following example indicates TLSRoute "tls-example" has been accepted by
Gateway "gw-example" in namespace "gw-example-ns":
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: tls-example
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
Multiple TLSRoutes can be attached to a single Gateway resource. Importantly,
only one Route hostname may match each request. For more information on how conflict
resolution applies to merging, refer to the [API specification][hostname].


[tlsroute]: ../reference/spec.md#tlsroute
[tlsrouterule]: ../reference/spec.md#tlsrouterouterule
[hostname]: ../reference/spec.md#hostname
[backendRef]: ../reference/spec.md#backendref
[parentRef]: ../reference/spec.md#parentreference
[name]: ../reference/spec.md#sectionname
[rfc-6066]: https://tools.ietf.org/html/rfc6066

