# GEP 2895: Query Parameter Filter

* Issue: [#2895](https://github.com/kubernetes-sigs/gateway-api/issues/2895)
* Status: Experimental

## TLDR

Just like modify header is useful, the same goes for query parameters. 
It's helpful to have a `HTTPQueryParamFilter` field in `HTTPRouteFilter` to set, 
add and remove a query parameter of the HTTP request before it is sent to the upstream target.

## Goals

* Provide a way to modify query parameters of an incoming request in a `HTTPRoute`.

## Introduction

The query parameters are an important part of the request URL. 
The developers can use query parameters to filter, sort or customize data of request body. 
Backend service can enable different function based on the query parameters. 
Query parameters are important information about search and track. 
Moreover, query parameter, headers and cookies are common techniques used in a canary release.

The `HTTPRouteFilter` API now supports filters `RequestHeaderModifier` and `ResponseHeaderModifier`. 
This GEP proposes adding support for modifying query parameters in a `HTTPRoute`.

## API

This GEP proposes to add a new field `HTTPQueryParamFilter` to `HTTPRouteFilter`.

The `HTTPQueryParamFilter` is considered an extended feature.

```golang
const (
    // HTTPRouteFilterQueryParamModifier can be used to set, add or remove a query
    // parameter from an HTTP request before it is sent to the upstream target.
    //
    // Support in HTTPRouteRule: Extended
    //
    // Support in HTTPBackendRef: Extended
    HTTPRouteFilterQueryParamModifier HTTPRouteFilterType = "QueryParamModifier"
)

// HTTPQueryParamFilter defines a filter that modifies HTTP query parameter.
// Only one action for a given query param name is permitted.
// Filters specifying multiple actions of the same or different type for any one
// query param name are invalid and will be rejected by CRD validation.
type HTTPQueryParamFilter struct {
    // Set overwrites the HTTP request with the given query param (name, value)
    // before the action. 
    // The request query parameter names are case-sensitive.
    // This must be an exact string match of query param name.
    // (See https://www.rfc-editor.org/rfc/rfc7230#section-2.7.3).
    //
    // Input:
    //   GET /foo?my-parameter=foo HTTP/1.1
    //
    // Config:
    //   set:
    //   - name: "my-parameter"
    //     value: "bar"
    //
    // Output:
    //   GET /foo?my-parameter=bar HTTP/1.1
    //
    // If the query parameter is not set in the request,
    // the Set action MUST be ignored by the Gateway.
    //
    // Input:
    //   GET /foo HTTP/1.1
    //
    // Config:
    //   set:
    //   - name: "my-parameter"
    //     value: "bar"
    //
    // Output:
    //   GET /foo HTTP/1.1
    //
    // +optional
    // +listType=map
    // +listMapKey=name
    // +kubebuilder:validation:MaxItems=16
    Set []HTTPHeader `json:"set,omitempty"`

    // Add adds the given query param(s) (name, value) to the HTTP request
    // before the action. Existing query params with the same name are not 
    // replaced, instead a new param with the same name is added.
    //
    // Input:
    //   GET /foo?my-parameter=foo HTTP/1.1
    //
    // Config:
    //   add:
    //   - name: "my-parameter"
    //     value: "bar"
    //
    // Output:
    //   GET /foo?my-parameter=foo&my-parameter=bar HTTP/1.1
    //
    // +optional
    // +listType=map
    // +listMapKey=name
    // +kubebuilder:validation:MaxItems=16
    Add []HTTPHeader `json:"add,omitempty"`

    // Remove the given query param(s) from the HTTP request before the action.
    // The value of Remove is a list of query param names. Note that the query
    // param names are case-sensitive (See
    // https://www.rfc-editor.org/rfc/rfc7230#section-2.7.3).
    //
    // Input:
    //   GET /foo?my-parameter1=foo&my-parameter2=bar&my-parameter3=baz HTTP/1.1
    //
    // Config:
    //   remove: ["my-parameter1", "my-parameter3"]
    //
    // Output:
    //   GET /foo?my-parameter2=bar HTTP/1.1
    //
    // +optional
    // +listType=set
    // +kubebuilder:validation:MaxItems=16
    Remove []string `json:"remove,omitempty"`
}
```

## Examples

The following example shows how a HTTPRoute modifies the query parameter of an HTTP request before it is sent to the upstream target.

It allows to add query parameter for only a certain canary backend, which can help in identifying certain users by the backend service. 
Based on the following http rule, query parameter "passtoken=$sign_passtoken_plain" will be added to the requests to be matched against the query parameter "gray=3", then the request will be routed to the canary service "http-route-canary:80".

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-query
spec:
  hostnames:
  - http.route.query.com
  - http.route.queries.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service
      name: http-route-production
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /
  - backendRefs:
    - kind: Service
      name: http-route-canary
      port: 80
    filters:
    - queryParamModifier:
        add:
        - name: passtoken
          value: $sign_passtoken_plain
      type: QueryParamModifier
    matches:
    - queryParams:
      - name: gray
        type: Exact
        value: 3

```

## Implementation-Specific Solutions

Some implementations already support query parameter modification.

### KONG supports this with a Request Transformer plugin

* The Request Transformer plugin for Kong allows simple transformation of requests before they reach the upstream server. These transformations can be used to add, append, remove, rename and replace of body, headers and querystring in incoming requests.

Below is an example that demonstrates a HTTP route adds the specified query parameter "new-param=some-value" to the requests "POST http://kong.test.org/test" before they are sent to the upstream target "my-service".

```yaml
services:
- name: my-service
  url:  http://kong.test.org
routes:
- name: my-route
  service: my-service
  methods:
  - POST
  paths:
    - /test
plugins:
- name: request-transformer
  route: my-route
  config:
    add:
      querystring:
      - new-param=some-value
```

### Traefik supports this with a Query Parameter Modification plugin

* This Traefik plugin allows users to modify the query parameters of an incoming request, by either adding new, deleting or modifying existing query parameters.

In the following example, HTTP route adds a new query parameter "authenticated=true" to the requests. 
Existing query params with the same name are not replaced, instead a new param with the same name is added. 

```
[http]
  [http.routers]
    [http.routers.router0]
      entryPoints = ["http"]
      service = "service-foo"
      rule = "Path(`/foo`)"
      middlewares = ["my-plugin"]

  [http.middlewares]
    [http.middlewares.my-plugin.plugin.dev]
      type = "add"
      paramName = "authenticated"
      newValue = "true"
```

### Tengine-ingress supports this with annotation nginx.ingress.kubernetes.io/canary-request-add-query

* The annotation nginx.ingress.kubernetes.io/canary-request-add-query adds a set of query parameters (key-value pair) to the end of URL. 
* The multiple query parameters are separated by the ampersand separator "&".
* If annotation canary-request-add-query has the same name as query parameter of an incoming request, instead a new param with the same name is added.

In the example below, if the value of the query parameter appid is "wx0ff419efbf920035", "wx4ad64dfd29b713a3" or "wxfb128531972f4bc0", HTTP route will add query parameter "passtoken=$sign_passtoken_plain" and "gray=on" to the request before it is sent to the upstream service "query-gray-service:80".

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-by-query: appid
    nginx.ingress.kubernetes.io/canary-by-query-value: wx0ff419efbf920035||wx4ad64dfd29b713a3||wxfb128531972f4bc0
    nginx.ingress.kubernetes.io/canary-request-add-query: passtoken=$sign_passtoken_plain&gray=on
  name: tengine-ingress-add-query
spec:
  rules:
  - host: tengine.query.net
    http:
      paths:
      - backend:
          service:
            name: query-gray-service
            port:
              number: 80
        path: /gray
        pathType: Prefix
```

## References

* [KONG `Request Transformer`](https://docs.konghq.com/hub/kong-inc/request-transformer/)
* [Traefik `Query Paramter Modification`](https://plugins.traefik.io/plugins/628c9f24ffc0cd18356a97bd/query-paramter-modification)
* [Tengine-ingress `ingress_routes`](https://tengine.taobao.org/document/ingress_routes.html)
* [RFC7230](https://www.rfc-editor.org/rfc/rfc7230)
