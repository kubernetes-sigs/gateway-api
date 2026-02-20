# Cross-Origin Resource Sharing (CORS)

The [HTTPRoute resource](../api-types/httproute.md) can be used to configure
Cross-Origin Resource Sharing (CORS). CORS is a security feature that allows
or denies web applications running at one domain to make requests for resources
from a different domain.

The `CORS` filter in an `HTTPRouteRule` can be used to specify the CORS policy.

## Allowing requests from a specific origin

The following `HTTPRoute` allows requests from `https://app.example`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cors-allow-credentials
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /cors-behavior-creds-false
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    filters:
    - cors:
        allowOrigins:
        - https://app.example
        allowCredentials: false
      type: CORS
```

Instead of specifying a list of specific origins, you can also specify a
single wildcard (`"*"`), which will allow any origin:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cors-allow-credentials
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /cors-behavior-creds-false
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    filters:
    - cors:
        allowOrigins:
        - "*"
        allowCredentials: false
      type: CORS
```

It is also allowed to use semi-specified origins in the list,
where the wildcard appears after the scheme
and at the beginning of the hostname, e.g. `https://*.bar.com`:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cors-allow-credentials
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /cors-behavior-creds-false
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    filters:
    - cors:
        allowOrigins:
        - https://www.baz.com
        - https://*.bar.com
        - https://*.foo.com
        allowCredentials: false
      type: CORS
```

## Allowing credentials

The `allowCredentials` field specifies whether the browser is allowed to
include credentials (such as cookies and HTTP authentication) in the CORS
request.

The following rule allows requests from `https://app.example` with
credentials:

```yaml
  - matches:
    - path:
        type: PathPrefix
        value: /cors-behavior-creds-true
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    filters:
    - cors:
        allowOrigins:
        - https://app.example
        allowCredentials: true
      type: CORS
```

## Other CORS options

The `CORS` filter also allows you to specify other CORS options, such as:

- `allowMethods`: The HTTP methods that are allowed for CORS requests.
- `allowHeaders`: The HTTP headers that are allowed for CORS requests.
- `exposeHeaders`: The HTTP headers that are exposed to the client.
- `maxAge`: The maximum time in seconds that the browser should cache the
preflight response.

For `allowMethods`, `allowHeaders`, and `exposeHeaders`, it is also possible
to use a single wildcard (`"*"`) instead of a list of specific names.

A comprehensive example:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: cors-allow-credentials
  namespace: gateway-conformance-infra
spec:
  parentRefs:
  - name: same-namespace
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /cors-behavior-creds-true
    backendRefs:
    - name: infra-backend-v1
      port: 8080
    filters:
    - cors:
        allowOrigins:
        - "https://www.foo.com"
        - "https://*.bar.com"
        allowMethods:
        - GET
        - OPTIONS
        allowHeaders:
        - "*"
        exposeHeaders:
        - "x-header-3"
        - "x-header-4"
        allowCredentials: true
        maxAge: 3600
      type: CORS
```
