# TLS Configuration

Gateway API allows for a variety of ways to configure TLS. This document lays
out various TLS settings and gives general guidelines on how to use them
effectively.

Although this doc covers the most common forms of TLS configuration with Gateway
API, some implementations may also offer implementation-specific extensions that
allow for different or more advanced forms of TLS configuration. In addition to
this documentation, it's worth reading the TLS documentation for whichever
implementation(s) you're using with Gateway API.

!!! info "Experimental Channel"

    The `TLSRoute` and `BackendTLSPolicy` resources described below are currently only included in the
    "Experimental" channel of Gateway API. For more information on release
    channels, refer to our [versioning guide](../concepts/versioning.md).

## Client/Server and TLS

![overview](../images/tls-overview.svg)

For Gateways, there are two connections involved:

- **downstream**: This is the connection between the client and the Gateway.
- **upstream**: This is the connection between the Gateway and backend resources
   specified by routes. These backend resources will usually be Services.

With Gateway API, TLS configuration of downstream and
upstream connections is managed independently.

For downstream connections, depending on the Listener Protocol, different TLS modes and Route types are supported.

| Listener Protocol | TLS Mode    | Route Type Supported |
|-------------------|-------------|---------------------|
| TLS               | Passthrough | TLSRoute            |
| TLS               | Terminate   | TCPRoute            |
| HTTPS             | Terminate   | HTTPRoute           |
| GRPC              | Terminate   | GRPCRoute           |

Please note that in case of `Passthrough` TLS mode, no TLS settings take
effect as the TLS session from the client is NOT terminated at the Gateway, but rather
passes through the Gateway, encrypted.

For upstream connections, `BackendTLSPolicy` is used, and neither listener protocol nor TLS mode apply to the
upstream TLS configuration. For `HTTPRoute`, the use of both `Terminate` TLS mode and `BackendTLSPolicy` is supported.
Using these together provides what is commonly known as a connection that is terminated and then re-encrypted at
the Gateway.

## Downstream TLS

Downstream TLS settings are configured using listeners at the Gateway level.

### Listeners and TLS

Listeners expose the TLS setting on a per domain or subdomain basis.
TLS settings of a listener are applied to all domains that satisfy the
`hostname` criteria.

In the following example, the Gateway serves the TLS certificate
defined in the `default-cert` Secret resource for all requests.
Although the example refers to HTTPS protocol, one can also use the same
feature for TLS-only protocol along with TLSRoutes.

```yaml
listeners:
- protocol: HTTPS # Other possible value is `TLS`
  port: 443
  tls:
    mode: Terminate # If protocol is `TLS`, `Passthrough` is a possible mode
    certificateRefs:
    - kind: Secret
      group: ""
      name: default-cert
```

### Examples

#### Listeners with different certificates

In this example, the Gateway is configured to serve the `foo.example.com` and
`bar.example.com` domains. The certificate for these domains is specified
in the Gateway.

```yaml
{% include 'standard/tls-basic.yaml' %}
```

#### Wildcard TLS listeners

In this example, the Gateway is configured with a wildcard certificate for
`*.example.com` and a different certificate for `foo.example.com`.
Since a specific match takes priority, the Gateway will serve
`foo-example-com-cert` for requests to `foo.example.com` and
`wildcard-example-com-cert` for all other requests.

```yaml
{% include 'standard/wildcard-tls-gateway.yaml' %}
```

#### Cross namespace certificate references

In this example, the Gateway is configured to reference a certificate in a
different namespace. This is allowed by the ReferenceGrant created in the
target namespace. Without that ReferenceGrant, the cross-namespace reference
would be invalid.

```yaml
{% include 'standard/tls-cert-cross-namespace.yaml' %}
```

## Upstream TLS

Upstream TLS settings are configured using the experimental `BackendTLSPolicy`
attached to a `Service` via a target reference.

This resource can be used to describe the SNI the Gateway should use to connect to the
backend and how the certificate served by the backend Pod(s) should be verified.

### TargetRefs and TLS

BackendTLSPolicy contains specification for the `TargetRefs` and `Validation`.  TargetRefs is required and
identifies one or more `Service`s for which your HTTPRoute requires TLS. The `Validation` configuration contains a
required `Hostname`, and either `CACertificateRefs` or `WellKnownCACertificates`.

Hostname refers to the SNI the Gateway should use to connect to the backend, and
must match the certificate served by the backend pod.

CACertificateRefs refer to one or more PEM-encoded TLS certificates. If there are no specific certificates
to use, then you must set WellKnownCACertificates to "System" to tell the Gateway to use a set of trusted
CA Certificates. There may be some variation in which system certificates are used by each implementation.
Refer to documentation from your implementation of choice for more information.

!!! info "Restrictions"

    - Cross-namespace certificate references are not allowed.
    - Wildcard hostnames are not allowed.

### Examples

#### Using System Certificates

In this example, the `BackendTLSPolicy` is configured to use system certificates to connect with a
TLS-encrypted upstream connection where Pods backing the `dev` Service are expected to serve a valid
certificate for `dev.example.com`.

```yaml
{% include 'experimental/v1alpha3/backendtlspolicy-system-certs.yaml' %}
```

#### Using Explicit CA Certificates

In this example, the `BackendTLSPolicy` is configured to use certificates defined in the configuration
map `auth-cert` to connect with a TLS-encrypted upstream connection where Pods backing the `auth` Service
are expected to serve a valid certificate for `auth.example.com`.

```yaml
{% include 'experimental/v1alpha3/backendtlspolicy-ca-certs.yaml' %}
```

## Extensions

Gateway TLS configurations provides an `options` map to add additional TLS
settings for implementation-specific features. Some examples of features that
could go in here would be TLS version restrictions, or ciphers to use.
