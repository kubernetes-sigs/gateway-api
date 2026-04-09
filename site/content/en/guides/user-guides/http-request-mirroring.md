---
title: "HTTP request mirroring"
weight: 6
---

{{< details title="Extended Support Feature: HTTPRouteRequestMirror" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](/docs/concepts/conformance/).

{{< /details >}}
The [HTTPRoute resource](/reference/api-types/httproute/) can be used to mirror
requests to multiple backends. This is useful for testing new services with
production traffic.

Mirrored requests will only be sent to one single destination endpoint
within this backendRef, and responses from this backend MUST be ignored by
the Gateway.

Request mirroring is particularly useful in blue-green deployment. It can be
used to assess the impact on application performance without impacting
responses to clients in any way.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-filter-mirror
  labels:
    gateway: mirror-gateway
spec:
  parentRefs:
  - name: mirror-gateway
  hostnames:
  - mirror.example
  rules:
  - backendRefs:
    - name: foo-v1
      port: 8080
    filters:
    - type: RequestMirror
      requestMirror:
        backendRef:
          name: foo-v2
          port: 8080
```

In this example, all requests are forwarded to service `foo-v1` on port `8080`,
and they are also forwarded to service `foo-v2` on port `8080`, but responses
are only generated from service `foo-v1`.
