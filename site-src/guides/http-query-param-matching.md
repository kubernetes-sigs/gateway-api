# HTTP query parameter matching

???+ info "Extended Support Feature: HTTPRouteQueryParamMatching"
    This feature is part of extended support. For more information on release channels, refer to our [versioning guide](../concepts/versioning.md).

The [HTTPRoute resource](../api-types/httproute.md) can be used to match
requests based on query parameters. This guide shows how to use this
functionality.

## Matching requests based on a single query parameter

The following `HTTPRoute` splits traffic between two backends based on the
value of the `animal` query parameter:

```yaml
apiVersion: gateway.networking.k.io/v1
kind: HTTPRoute
metadata:
  name: query-param-matching
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - queryParams:
      - name: animal
        value: whale
    backendRefs:
    - name: infra-backend-v1
      port: 8080
  - matches:
    - queryParams:
      - name: animal
        value: dolphin
    backendRefs:
    - name: infra-backend-v2
      port: 8080
```

- A request to `/` with the query parameter `animal=whale` will be routed to `infra-backend-v1`.
- A request to `/` with the query parameter `animal=dolphin` will be routed to `infra-backend-v2`.

## Matching requests based on multiple query parameters

A rule can also match on multiple query parameters. The following rule routes
traffic to `infra-backend-v3` if the query parameters `animal=dolphin` AND
`color=blue` are present:

```yaml
  - matches:
    - queryParams:
      - name: animal
        value: dolphin
      - name: color
        value: blue
    backendRefs:
    - name: infra-backend-v3
      port: 8080
```

## ORing matches

If a rule has multiple `matches`, a request will be routed if it satisfies any
of them. The following rule routes traffic to `infra-backend-v3` if:

- The query parameters `animal=dolphin` AND `color=blue` are present.
- OR the query parameter `ANIMAL=Whale` is present.

```yaml
  - matches:
    - queryParams:
      - name: animal
        value: dolphin
      - name: color
        value: blue
    - queryParams:
      - name: ANIMAL
        value: Whale
    backendRefs:
    - name: infra-backend-v3
      port: 8080
```

## Combining with other match types

Query parameter matching can be combined with other match types like path and
header matching. The following rules demonstrate this:

```yaml
  - matches:
    - path:
        type: PathPrefix
        value: /path1
      queryParams:
      - name: animal
        value: whale
    backendRefs:
    - name: infra-backend-v1
      port: 8080
  - matches:
    - headers:
      - name: version
        value: one
      queryParams:
      - name: animal
        value: whale
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
      queryParams:
      - name: animal
        value: whale
    backendRefs:
    - name: infra-backend-v3
      port: 8080
```
