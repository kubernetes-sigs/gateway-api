# TLS routing

The [TLSRoute resource](../api-types/tlsroute.md) allows you to match on TLS traffic and
direct it to Kubernetes backends. This guide shows how the TLSRoute matches
traffic on hostname and forwards it to different Kubernetes Services, using passthrough 
or terminating it on the Gateway.

TLSRoute is covered by the following features, that may be reported by your implementation
* `TLSRoute` - If reported, means your implementation supports `TLSRoute` with `Passthrough` mode. Any implementation that claims to support the `TLSRoute` API MUST report this feature.
* `TLSRouteModeTerminate` - If reported, means your implementation supports `TLSRoute` with `Terminate` mode additionally to  `Passthrough` mode
* `TLSRouteModeMixed` - If reported, means your implementation supports two TLS listeners with distinct modes (`Passthrough` and `Terminate`) on the same port.

The following diagram describes a required traffic flow across two different
Services:

- Traffic to `foo.example.com` is forwarded as is to `foo-svc`
- Traffic to `bar.example.com` has its TLS traffic terminated on the Gateway and then 
forwarded as a TCP stream to `bar-svc`

![TLS Routing](../images/tls-routing.svc)

The dotted lines show the Gateway resources deployed to configure this routing
behavior. There are two TLSRoute resources that create routing rules on the
same `prod-tls` Gateway.

In order to receive traffic from a [Gateway][gateway] a `TLSRoute` resource
must be configured with `ParentRefs` which reference the parent gateway(s) that it
should be attached to. The following example shows how the combination
of `Gateway` and `TLSRoute` would be configured to serve TLS traffic using passthrough
and terminate, when supported by the Gateway API implementation:

```yaml
{% include 'standard/tls-routing/gateway.yaml' %}
```

A TLSRoute can match against a [single set of hostnames][spec].
Since `foo.example.com` and `bar.example.com` are separate hosts with
different routing requirements, each is deployed as its own TLSRoute -
`foo-route` and `bar-route`.

The following `foo-route` will match any traffic for `foo.example.com` and apply
its routing rules to forward the traffic to the correct backend. Since it is attached
to a listener that is configured as `Passthrough` mode it will pass TCP stream traffic directly
to the backend:

```yaml
{% include 'standard/tls-routing/tls-route.yaml' %}
```

Similarly, the `bar-route`  matches traffic for `bar.example.com`, but given it is attached to
a listener that is configured as `Terminate` mode, it will first terminate the TLS stream 
on the Gateway, and then pass an unencrypted TCP stream to the backend.

```yaml
{% include 'standard/tls-routing/tls-route-terminate.yaml' %}
```

[gateway]: ../reference/spec.md#gateway
[spec]: ../reference/spec.md#tlsroutespec
