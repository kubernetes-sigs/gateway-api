# TLS routing

The [TLSRoute resource](../api-types/tlsroute.md) allows you to match on TLS
metadata and direct it to Kubernetes backends. This guide shows how the TLSRoute
matches traffic on hostname and forwards it to different Kubernetes Services,
using either `Passthrough` or `Terminate` TLS modes on the Gateway.

In order to receive traffic from a [Gateway][gateway] a TLSRoute resource
must be configured with `ParentRefs` which reference the parent gateway(s) that it
should be attached to. The following example shows how the combination
of Gateway and TLSRoute would be configured to serve TLS traffic using both
`Passthrough` and `Terminate` modes (when supported by the Gateway API
implementation):

```yaml
{% include 'standard/tls-routing/gateway.yaml' %}
```

A TLSRoute can match against a [single set of hostnames][spec].
Since `foo.example.com` and `bar.example.com` are separate hosts with
different routing requirements, each is deployed as its own TLSRoute -
`foo-route` and `bar-route`.

The following `foo-route` TLSRoute will match any traffic for `foo.example.com`
and apply its routing rules to forward the traffic to the configured backend.
Since it is attached to a listener that is configured in `Passthrough` mode, the
Gateway will pass the encrypted TCP stream directly to the backend:

```yaml
{% include 'standard/tls-routing/tls-route.yaml' %}
```

Similarly, the `bar-route` TLSRoute matches traffic for `bar.example.com`.
However, since it is attached to a listener that is configured in `Terminate`
mode, the Gateway will first terminate the TLS stream using the certificate
specified on the listener, and then pass the resulting unencrypted TCP stream to
the backend:

```yaml
{% include 'standard/tls-routing/tls-route-terminate.yaml' %}
```

[gateway]: ../reference/spec.md#gateway
[spec]: ../reference/spec.md#tlsroutespec
