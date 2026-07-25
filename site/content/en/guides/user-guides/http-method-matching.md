---
title: "HTTP method matching"
weight: 8
---

{{< details title="Extended Support Feature: HTTPRouteMethodMatching" open="true" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](/docs/concepts/conformance/).

{{< /details >}}

The [HTTPRoute resource](/reference/api-types/httproute/) can be used to match
requests based on the HTTP method. This guide shows how to use this
functionality.

## Matching requests based on the HTTP method

The following `HTTPRoute` splits traffic between two backends based on the
HTTP method of the request:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: method-matching
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - method: POST
    backendRefs:
    - name: infra-backend-v1
      port: 8080
  - matches:
    - method: GET
    backendRefs:
    - name: infra-backend-v2
      port: 8080
```

- A `POST` request to `/` will be routed to `infra-backend-v1`.
- A `GET` request to `/` will be routed to `infra-backend-v2`.

## Combining with other match types

Method matching can be combined with other match types like path and header
matching. The following rules demonstrate this:

```yaml
  # Combinations with core match types.
  - matches:
    - path:
        type: PathPrefix
        value: /path1
      method: GET
    backendRefs:
    - name: infra-backend-v1
      port: 8080
  - matches:
    - headers:
      - name: version
        value: one
      method: PUT
    backendRefs:
    - name: infra-backend-v2
      port: 8080
  - matches:
    - path:
        type: PathPrefix
        value: /path2
      headers:
      - name: version
        value: two
      method: POST
    backendRefs:
    - name: infra-backend-v3
      port: 8080
```

## Matching multiple methods in a single match

{{< details title="Experimental Support Feature: HTTPRouteMultipleMethodMatching" open="true" >}}
The `methods` field is part of experimental support. For more information on
support levels, refer to our [conformance guide](/docs/concepts/conformance/).
{{< /details >}}

When the `methods` field is used, you can list multiple HTTP methods in one
`HTTPRouteMatch`. A request matches if its method is any value in `methods`.
When `methods` is specified, `method` must also be specified and must equal
`methods[0]`. This differs from [ORing matches](#oring-matches) below, where
multiple `matches` entries are combined with OR across separate match blocks.

```yaml
  - matches:
    - path:
        type: PathPrefix
        value: /api/
      method: GET
      methods:
      - GET
      - HEAD
      - POST
      - PUT
      - DELETE
      - PATCH
      - OPTIONS
    backendRefs:
    - name: infra-backend-v1
      port: 8080
```

- A request to `/api/` with method `GET`, `HEAD`, `POST`, `PUT`, `DELETE`,
  `PATCH`, or `OPTIONS` will be routed to `infra-backend-v1`.
- A request with method `TRACE` will not match this rule.

## ORing matches

If a rule has multiple `matches`, a request will be routed if it satisfies any
of them. The following rule routes traffic to `infra-backend-v1` if:

- The request is a `PATCH` to `/path3`.
- OR the request is a `DELETE` to `/path4` with the `version: three` header.

```yaml
  # Match of the form (cond1 AND cond2) OR (cond3 AND cond4 AND cond5)
  - matches:
    - path:
        type: PathPrefix
        value: /path3
      method: PATCH
    - path:
        type: PathPrefix
        value: /path4
      headers:
      - name: version
        value: three
      method: DELETE
    backendRefs:
    - name: infra-backend-v1
      port: 8080
```
