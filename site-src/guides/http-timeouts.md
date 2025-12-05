# HTTP timeouts

???+ info "Extended Support Feature: HTTPRouteRequestTimeout"
    This feature is part of extended support. For more information on release channels, refer to our [versioning guide](../concepts/versioning.md).

The [HTTPRoute resource](../api-types/httproute.md) can be used to configure
timeouts for HTTP requests. This is useful for preventing long-running requests
from consuming resources and for providing a better user experience.

The `timeouts` field in an `HTTPRouteRule` can be used to specify a request
timeout.

## Setting a request timeout

The following `HTTPRoute` sets a request timeout of 500 milliseconds for all
requests with a path prefix of `/request-timeout`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: request-timeout
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /request-timeout
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    timeouts:
      request: 500ms
```

If a request to this path takes longer than 500 milliseconds, the gateway will
return a timeout error.

## Disabling the request timeout

To disable the request timeout, set the `request` field to `"0s"`:

```yaml
  - matches:
    - path:
        type: PathPrefix
        value: /disable-request-timeout
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    timeouts:
      request: "0s"
```
