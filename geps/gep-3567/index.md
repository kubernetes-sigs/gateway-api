# GEP-3567: Gateway TLS Updates for HTTP/2 Connection Coalescing

* Issue: [#3567](https://github.com/kubernetes-sigs/gateway-api/issues/3567)
* Status: Experimental

## TLDR

As described in the [previous
doc](https://docs.google.com/document/d/1g_TNN8eOaVDC3xesO9JFdvQbPFdSTHp1vb70TD3-Vrs/edit?tab=t.0#heading=h.qiz1tfw67tbp),
the current state of TLS configuration on Gateways can lead to confusing
behavior when combined with HTTP/2 connection coalescing. This GEP proposes a
series of changes to the API to address these problems.

## Goals

* Take steps that will make it less likely for users to encounter these problems
* Warn when users have configuration that is prone to these issues
* Provide central source of documentation explaining both the problem and
  potential solutions

## Non-Goals

* Breaking or significantly disruptive changes to the existing API surface

## Introduction

Gateway API creates situations where clients might be able to send requests
through a Listener that, according to the Gateway’s configuration, is not
supposed to receive these requests. This can cause requests to be apparently
mis-routed.

The problem here is an inherent conflict between the API and the mechanics of
HTTPS. Gateway API uses the “hostname” field in the Listener to constrain both
the TLS certificate selection and the host header of requests. But when a server
presents a TLS certificate that is valid for multiple domains, a client is free
to reuse its TLS connection for requests sent to any of those domains (for
HTTP2, see RFC). The SNI hostname, which the client presents only with the
initial TLS handshake, doesn’t constrain the host header of the requests that
the client sends.

Gateway API deals with this situation imprecisely, stating:

    The Listener Hostname SHOULD match at both the TLS and HTTP protocol layers
    as described above. If an implementation does not ensure that both the SNI
    and Host header match the Listener hostname, it MUST clearly document that.

In practice we can end up with an implementation that misroutes requests when a
Gateway is configured using certificates that use multiple or wildcard SANs.

### Example

The following configuration ([from the Gateway API
documentation](../../guides/tls.md#wildcard-tls-listeners))
illustrates the problem:


```
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: wildcard-tls-gateway
spec:
  gatewayClassName: example
  listeners:
  - name: foo-https
    protocol: HTTPS
    port: 443
    hostname: foo.example.com
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: foo-example-com-cert  # SAN: foo.example.com
  - name: wildcard-https
    protocol: HTTPS
    port: 443
    hostname: "*.example.com"
    tls:
      certificateRefs:
      - kind: Secret
        group: ""
        name: wildcard-example-com-cert  # SAN: *.example.com
```


The Gateway API definition requires requests to `foo.example.com` to be
associated with the `foo-https` listener, on connections negotiated with
`foo-example-com-cert`.

Suppose a client sends a request to `bar.example.com`, specifying that as the
SNI hostname, and establishes a TLS connection attached to the `wildcard-https`
Listener. And then it sends a subsequent request to `foo.example.com`. The
client can correctly reuse its existing TLS connection for the second request,
because the `wildcard-example-com-cert` is valid also for `foo.example.com`. But
now the Gateway has a problem: Routing the request via the `wildcard-https`
Listener violates the intent of the configuration, and routing via the
`foo-https` Listener is inconsistent with the connection’s having been
negotiated with the other Listener’s certificate.

Mapping a request to a Listener matters if the Gateway configuration has
different
[HTTPRoutes](../../reference/spec.md#httproute)
bound to the different Listeners. It also matters if the Listeners have
different
[GatewayTlsConfigs](../../reference/spec.md#gatewaytlsconfig)
attached, for example if one Listener uses mutual TLS and the other does not.


### Interaction with Client Cert Validation

[GEP-91](../gep-91/index.md) introduced Client
Certificate Validation to Gateway Listeners as a new experimental concept. If an
implementation is unable to properly isolate HTTPS listeners, this could result
in this Client Cert Validation being bypassed. Before this feature can graduate
beyond experimental, we’ll need to resolve this underlying issue.

## Proposal

### A) Add Warning in Gateway Status
A new condition will be added to Gateways: `OverlappingTLSConfig`.
Implementations MUST add this condition to status when a Gateway is configured
with TLS configuration across multiple Listeners. Implementations MAY add this
condition to status when a Gateway is configured with overlapping TLS
certificates. Note that since this is a negative polarity condition, it would
only be populated when it is true.

### B) Modify API Spec to recommend sending 421s
The Gateway spec for `listener.hostname` will be updated to recommend returning
a 421 when this problem occurs.

### C) Top Level Gateway TLS Config for Client Cert Validation

A follow up discussion for GEP-91 will consider if Client Cert Validation should
be moved or copied to a new top level Gateway TLS config instead of
per-listener.

## Conformance Details

#### Feature Names

A) None, this will be required for any implementations that support HTTP +
Gateways.

B) `GatewayReturn421`

C) Will be covered in GEP-91

### Conformance tests

A) A new conformance test will be added to ensure that the new status condition
is populated when a Gateway is configured with overlapping TLS configuration.

B) A new conformance test will be added to ensure that implementations return a
421 when a connection is reused for a different listener with an overlapping
SNI.

C) Will be covered in GEP-91

## Alternatives

Discussed in more detail the [original
doc](https://docs.google.com/document/d/1g_TNN8eOaVDC3xesO9JFdvQbPFdSTHp1vb70TD3-Vrs/edit?tab=t.0#heading=h.qiz1tfw67tbp)
