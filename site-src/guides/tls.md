# TLS details

Gateway API allow for a variety of ways to configure TLS. This document lays
out various TLS settings and gives general guidelines on how to use them
effectively.

## Client/Server and TLS

![overview](/images/tls-overview.svg)

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
      group: core
      name: default-cert
    routeOverride:
      certificate: Deny
```

If `hostname.match` is set to `Exact`, then the TLS settings apply to only the
specific hostname that is set in `hostname.name`.

Specifying `tls.routeOverride.certificate: Deny` is recommended because it
centralizes TLS configuration within the Gateway specification and should
suffice for the majority of use-cases. Please take a look at the examples below
for various alternatives.

### Routes and TLS

If `listeners[].tls.routeOverride.certificate` is set to `Allow`, TLS certificates
can be configured on routes that are bound to the Gateway. This feature is
primarily meant for a cluster with a self-service model where Application developers
bring their own TLS certificates. This feature also mirrors the behavior of
TLS as defined in the Ingress v1 resource. One should use this feature only
when the Cluster Operator wishes to delegate TLS configuration to the Application Developer.
With this feature, the certificate defined in the route overrides any certificate defined in
the Gateway.

When using this feature, please note that the TLS certificate to serve is chosen
before an HTTPRoute is selected. This is because the TLS handshake is completed
before an HTTP request is sent from the client.

[TLS Certificate in Route](#tls-certificate-in-route) provides an example
of how this feature can be used.

Also, as mentioned above, the Route Kind (`HTTPRoute`, `TLSRoute`, `TCPRoute`) 
is dependent on the protocol on the listener level. Listeners with `HTTPS` or 
`HTTP` protocols can use `HTTPRoute` as the TLS Termination is done at the 
listener level and thus, only HTTP information is used for routing.

Listeners with the `TLS` protocol must use `TLSRoute` when the mode is set to `Passthrough` and `TCPRoute` when the mode is `Terminate`.

Listeners with the TCP protocol must use `TCPRoute` for plain TCP Routing.

### Examples

#### TLS in listener

In this example, the Gateway is configured to serve the `foo.example.com` and
`bar.example.com` domains. The certificate for these domains is specified
in the Gateway.

```
{% include 'tls-basic.yaml' %}
```

#### Wildcard TLS listeners

In this example, the Gateway is configured with a wildcard certificate for
`*.example.com` and a different certificate for `foo.example.com`.
Since a specific match takes priority, the Gateway will serve
`foo-example-com-cert` for requests to `foo.example.com` and
`wildcard-example-com-cert` for all other requests.

```yaml
{% include 'wildcard-tls-gateway.yaml' %}
```

#### TLS Certificate in Route

In this example, the Gateway is configured with a default certificate that will be
served for all hostnames. In addition, `tls.routeOverride.certificate` is set to
`Allow`, meaning routes can specify TLS certificates for any domains. Next,
there are two HTTPRoute resources which specify certificates for
`foo.example.com` and `bar.example.com`.

```yaml
{% include 'tls-cert-in-route.yaml' %}
```

## Upstream TLS

Upstream TLS configuration applies to the connection between the Gateway
and Service.

There is only one way to configure upstream TLS: using the `BackendPolicy`
resource.

Please note that the TLS configuration is related to the Service or backend
resource and not related to a specific route resource.

### Example

The following example shows how upstream TLS can be configured. We have
omitted downstream TLS configuration for simplicity. As noted before, it
doesn't matter how downstream TLS is configured for the specific listener or
route.

```yaml
{% include 'upstream-tls.yaml' %}
```

## Extensions

Both upstream and downstream TLS configs provide an `options` map to add
additional TLS settings for implementation-specific features.
Some examples of features that could go in here would be TLS version restrictions,
or ciphers to use.
