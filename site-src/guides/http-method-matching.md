# HTTP method matching

???+ info "Extended Support Feature: HTTPRouteMethodMatching"
    This feature is part of extended support. For more information on release channels, refer to our [versioning guide](../concepts/versioning.md).

The [HTTPRoute resource](../api-types/httproute.md) can be used to match
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
