# TLS Configuration

!!! danger
    This page has not been updated for v1alpha2 yet.


Gateway API allow for a variety of ways to configure TLS. This document lays
out various TLS settings and gives general guidelines on how to use them
effectively.

## Client/Server and TLS

![overview](/v1alpha2/images/tls-overview.svg)

For Gateways, there are two connections involved:

- **downstream**: This is the connection between the client and the Gateway.
- **upstream**: This is the connection between the Gateway and backend resources
   specified by routes. These backend resources will usually be Services.

With Gateway API, TLS configuration of downstream and
upstream connections is managed independently.

Depending on the Listener Protocol, different TLS modes and Route types are supported.

Listener Protocol | TLS Mode | Route Type Supported
--- | --- | ---
TLS | Passthrough | TLSRoute
TLS | Terminate | TCPRoute
HTTPS | Terminate | HTTPRoute

Please note that in case of `Passthrough` TLS mode, no TLS settings take
effect as the TLS session from the client is NOT terminated at the Gateway.
The rest of the document assumes that TLS is being terminated at the Gateway,
which is the default setting.

## Downstream TLS

Downstream TLS settings are configured using listeners at the Gateway level.

### Listeners and TLS

Listeners expose the TLS setting on a per domain or sub-domain basis.
TLS settings of a listener are applied to all domains that satisfy the
`hostname` criteria.

In the following example, the Gateway serves the TLS certificate
defined in the `default-cert` Secret resource for all requests.
Although, the example refers to HTTPS protocol, one can also use the same
feature for TLS-only protocol along with TLSRoutes.

```yaml
listeners:
- protocol: HTTPS # Other possible value is `TLS`
  port: 443
  tls:
    mode: Terminate # If protocol is `TLS`, `Passthrough` is a possible mode
    certificateRef:
      kind: Secret
      group: ""
      name: default-cert
```

If `hostname.match` is set to `Exact`, then the TLS settings apply to only the
specific hostname that is set in `hostname.name`.

### Examples

#### Listeners with different certificates

In this example, the Gateway is configured to serve the `foo.example.com` and
`bar.example.com` domains. The certificate for these domains is specified
in the Gateway.

```yaml
{% include 'v1alpha2/tls-basic.yaml' %}
```

#### Wildcard TLS listeners

In this example, the Gateway is configured with a wildcard certificate for
`*.example.com` and a different certificate for `foo.example.com`.
Since a specific match takes priority, the Gateway will serve
`foo-example-com-cert` for requests to `foo.example.com` and
`wildcard-example-com-cert` for all other requests.

```yaml
{% include 'v1alpha2/wildcard-tls-gateway.yaml' %}
```

#### Cross namespace certificate references

In this example, the Gateway is configured to reference a certificate in a
different namespace. This is allowed by the ReferencePolicy created in the
target namespace. Without that ReferencePolicy, the cross-namespace reference
would be invalid.

```yaml
{% include 'v1alpha2/tls-cert-cross-namespace.yaml' %}
```

## Extensions

Gateway TLS configurations provides an `options` map to add additional TLS
settings for implementation-specific features. Some examples of features that
could go in here would be TLS version restrictions, or ciphers to use.
