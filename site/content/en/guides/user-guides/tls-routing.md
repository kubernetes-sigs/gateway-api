---
title: "TLS routing"
weight: 13
---

The [TLSRoute resource](/reference/api-types/tlsroute/) allows you to match on TLS
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
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: example-gateway
spec:
  gatewayClassName: example-gateway-class
  listeners:
  - name: tls
    protocol: TLS
    port: 443
    tls:
      mode: Passthrough
  - name: tls-terminate
    protocol: TLS
    port: 8443
    tls:
      mode: Terminate
      certificateRefs:
      - name: tls-terminate-certificate
```

A TLSRoute can match against a [single set of hostnames][spec]. For details on
hostname intersection with Listeners and routing behavior, see [Hostnames in
Gateway API](/docs/concepts/hostnames/). Since `foo.example.com` and
`bar.example.com` are separate hosts with different routing requirements, each
is deployed as its own TLSRoute - `foo-route` and `bar-route`.

The following `foo-route` TLSRoute will match any traffic for `foo.example.com`
and apply its routing rules to forward the traffic to the configured backend.
Since it is attached to a listener that is configured in `Passthrough` mode, the
Gateway will pass the encrypted TCP stream directly to the backend:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: foo-route
spec:
  parentRefs:
  - name: example-gateway
    sectionName: tls
  hostnames:
  - "foo.example.com"
  rules:
  - backendRefs:
    - name: foo-svc
      port: 443
```

Similarly, the `bar-route` TLSRoute matches traffic for `bar.example.com`.
However, since it is attached to a listener that is configured in `Terminate`
mode, the Gateway will first terminate the TLS stream using the certificate
specified on the listener, and then pass the resulting unencrypted TCP stream to
the backend:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: TLSRoute
metadata:
  name: bar-route
spec:
  parentRefs:
  - name: example-gateway
    sectionName: tls-terminate
  hostnames:
  - "bar.example.com"
  rules:
  - backendRefs:
    - name: bar-svc
      port: 8080
```

[gateway]: /reference/api-spec/main/spec/#gateway
[spec]: /reference/api-spec/main/spec/#tlsroutespec
