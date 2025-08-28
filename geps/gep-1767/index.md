# GEP 1767: CORS Filter

* Issue: [#1767](https://github.com/kubernetes-sigs/gateway-api/issues/1767)
* Status: Experimental

## TLDR
Cross-origin resource sharing (CORS) is an HTTP-header based mechanism that allows a web page to access restricted resources from a server on an origin (domain, scheme, or port) different than the domain that served the web page.
It's helpful to have a `HTTPCorsFilter` field in `HTTPRouteFilter` to handle the cross-origin requests before the response is sent to the client.

## Goals

* Support CORS filter in a `HTTPRoute`

## Introduction
The CORS protocol is the current specification to support secure cross-origin requests and data transfers between clients and servers.

A CORS request is an HTTP request that includes an `Origin` header. 
An origin consists of three parts: the scheme, host and port. Two URLs have the same origin if they have the same scheme, host, and port.
All of the following URLs have the same origin.
```text
   http://example.com/
   http://example.com:80/
   http://example.com/path/file
```

Each of the following URLs has a different origin from the others.
```text
   http://example.com/
   http://example.com:8080/
   http://www.example.com/
   https://example.com:80/
   https://example.com/
   http://example.org/
   http://ietf.org/
```

Before the actual cross-origin requests, clients will initiate an extra "preflight" request to determine whether the server will permit the actual requests. 
The CORS "preflight" request uses `OPTIONS` as method and includes the following headers:
    `Origin` request header indicates where a request originates from.
    `Access-Control-Request-Method` request header lets the server know which HTTP method will be used when the actual cross-origin request is made.
    `Access-Control-Request-Headers` is an optional request header indicates which headers the actual cross-origin request might use.

The server response for the CORS "preflight" request includes the following headers:
    `Access-Control-Allow-Origin` response header indicates whether the response can be shared with requested resource from the given `Origin`.
    `Access-Control-Allow-Methods` response header specifies one or more HTTP methods are accepted by the server when accessing the requested resource. 
    `Access-Control-Allow-Headers` response header indicates which HTTP headers can be used during the actual cross-origin request.
    
The `Access-Control-Max-Age` optional response header indicates how long (in seconds) the information provided by the headers `Access-Control-Allow-Methods` and `Access-Control-Allow-Headers` can be cached by client. The default value for `Access-Control-Max-Age` is 5 seconds. Until the time specified by `Access-Control-Max-Age` elapses, the client doesn't have to send another "preflight" request.

The optional response header `Access-Control-Expose-Headers` controls which HTTP response headers are exposed to clients for the actual cross-origin request. 

If the server specifies the response header `Access-Control-Allow-Credentials: true`, the actual cross-origin request will be able to use credentials for getting sensitive resources. 
Credentials are cookies, TLS client certificates, or authentication headers containing a username and password.

After the server has permitted the CORS "preflight" request, the client will be able to send actual cross-origin request.
If the server doesn't want to allow cross-origin access, it will omit the CORS headers to the client.
Therefore, the client doesn't attempt the actual cross-origin request.

In a simple cross-origin interaction, the client sends the request and cross-origin headers at the same time. These are usually GET data requests and are considered low-risk.

For example, a client sends a GET request with cross-origin header `Origin`.
```
GET /resource/foo HTTP/1.1
Host: http.route.cors.com
Origin: https://foo.example
```

The server sets response header Access-Control-Allow-Origin with "*", which means that the requested resource can be accessed from the any `Origin`.
```
HTTP/1.1 200 OK
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, HEAD, POST
Access-Control-Allow-Headers: Accept,Accept-Language,Content-Language,Content-Type,Range
```

Some HTTP requests are considered complex and require server confirmation before the actual cross-origin request is sent. Before the actual cross-origin requests, clients will initiate an extra "preflight" request to determine whether that the server will permit the actual requests.

For example, a client sends a cross-origin "preflight" request for asking a server whether it would allow a PUT request before the actual cross-origin request is sent.
```
OPTIONS /resource/foo HTTP/1.1
Host: http.route.cors.com
Origin: https://foo.example
Access-Control-Request-Method: PUT
```

If the "preflight" request is denied, the requested resource will end up not being shared.
The server returns 200 OK but doesn't set the cross-origin response headers.
Therefore, the client doesn't attempt the actual cross-origin request.
```
HTTP/1.1 200 OK
Content-Type: text/plain charset=UTF-8
Content-Length: 0
```

If the server allows it, it will respond with an OK status (i.e., 204 or 200) and the following response headers.
```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://foo.example
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization
Access-Control-Expose-Headers: Content-Security-Policy
Access-Control-Max-Age: 1728000
Content-Type: text/plain charset=UTF-8
Content-Length: 0
```

Then, the client will be able to send actual cross-origin request.
```
PUT /resource/foo HTTP/1.1
Host: http.route.cors.com
Keep-Alive: timeout=5, max=1000
Origin: https://foo.example
Authorization: Basic YWxhZGRpbjpvcGVuc2VzYW1l
```

At last, the cross-origin response headers will be added by the server to the response.
```
Access-Control-Allow-Origin: https://foo.example
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization
Access-Control-Expose-Headers: Content-Security-Policy
```

## API
This GEP proposes to add a new field `HTTPCORSFilter` to `HTTPRouteFilter`.
If `HTTPCORSFilter` is set, then the gateway will generate the response of the "preflight" requests and send back it to the client directly.
For the actual cross-origin request, the gateway will add CORS headers to the response before it is sent to the client.

```golang
const (
    // HTTPRouteFilterCORS can be used to add CORS headers to an 
    // HTTP response before it is sent to the client.
    //
    // Support in HTTPRouteRule: Extended
    //
    // Support in HTTPBackendRef: Extended
    HTTPRouteFilterCORS HTTPRouteFilterType = "CORS"
)

type HTTPRouteFilter struct {
    // CORS defines a schema for a filter that responds to the
    // cross-origin request based on HTTP response header.
    // 
    // Support: Extended
    //
    // +optional
    CORS *HTTPCORSFilter `json:"cors,omitempty"`
}

// HTTPCORSFilter defines a filter that responds to the
// cross-origin request based on HTTP header.
type HTTPCORSFilter struct {
    // AllowOrigins indicates whether the response can be shared with 
    // requested resource from the given `Origin`.
    // 
    // The `Origin` consists of a scheme and a host, with an optional 
    // port, and takes the form `<scheme>://<host>(:<port>)`.
    //
    // Valid values for scheme are: `http` and `https`.
    //
    // Valid values for port are any integer between 1 and 65535 
    // (the list of available TCP/UDP ports). Note that, if not included, 
    // port `80` is assumed for `http` scheme origins, and port `443` 
    // is assumed for `https` origins. This may affect origin matching.
    //
    // The host part of the origin may contain the wildcard character `*`.
    // These wildcard characters behave as follows:
    //
    // * `*` is a greedy match to the _left_, including any number of 
    //   DNS labels to the left of its position. This also means that 
    //   `*` will include any number of period `.` characters to the 
    //   left of its position.
    // * A wildcard by itself matches all hosts.
    //
    // An origin value that includes _only_ the `*` character 
    // indicates requests from all `Origin`s are allowed.
    //
    // When the `AllowOrigins` field is configured with multiple 
    // origins, it means the server supports clients from multiple 
    // origins. If the request `Origin` matches the configured 
    // allowed origins, the gateway must return the given `Origin` 
    // and sets value of the header `Access-Control-Allow-Origin` 
    // same as the `Origin` header provided by the client.
    //
    // The status code of a successful response to a "preflight" 
    // request is always an OK status (i.e., 204 or 200). 
    //
    // Input:
    //   Origin: https://foo.example
    //
    // Config:
    //   allowOrigins: ["https://foo.example", "http://foo.example", "https://test.example", "http://test.example"]
    //
    // Output:
    //   Access-Control-Allow-Origin: https://foo.example
    //
    // If the request `Origin` does not match the configured allowed origins, 
    // the gateway returns 204/200 response but doesn't set the relevant 
    // cross-origin response headers. Alternatively, the gateway responds with 
    // 403 status to the "preflight" request is denied, coupled with omitting 
    // the CORS headers. The cross-origin request fails on the client side.
    // Therefore, the client doesn't attempt the actual cross-origin request.
    //
    // Input:
    //   Origin: https://foo.example
    //
    // Config:
    //   allowOrigins: ["https://test.example", "http://test.example"]
    //
    // Output:
    //
    // The `Access-Control-Allow-Origin` response header can only use `*` 
    // wildcard as value when the `AllowCredentials` field is false.
    //
    // Input:
    //   Origin: https://foo.example
    //
    // Config:
    //   allowOrigins: ["*"]
    //
    // Output:
    //   Access-Control-Allow-Origin: *
    //
    // When the `AllowCredentials` field is true and `AllowOrigins`
    // field specified with the `*` wildcard, the gateway must return a 
    // single origin in the value of the `Access-Control-Allow-Origin` 
    // response header, instead of specifying the `*` wildcard. The value 
    // of the header `Access-Control-Allow-Origin` is same as the `Origin` 
    // header provided by the client.
    //
    // Input:
    //   Origin: https://foo.example
    //
    // Config:
    //   allowOrigins: ["*"]
    //   allowCredentials: true
    //
    // Output:
    //   Access-Control-Allow-Origin: https://foo.example
    //   Access-Control-Allow-Credentials: true
    //
    // Support: Extended
    // +listType=set
    // +kubebuilder:validation:MaxItems=64
    AllowOrigins []string `json:"allowOrigins,omitempty"`
 
    // AllowCredentials indicates whether the actual cross-origin request 
    // allows to include credentials.
    //
    // When set to true, the gateway will include the `Access-Control-Allow-Credentials`
    // response header with value true (case-sensitive).
    //
    // Input:
    //   Origin: https://foo.example
    //
    // Config:
    //   allowCredentials: true
    //
    // Output:
    //   Access-Control-Allow-Origin: https://foo.example
    //   Access-Control-Allow-Credentials: true
    //
    // When set to false, the gateway will omit the header
    // `Access-Control-Allow-Credentials` entirely (this is the standard CORS
    // behavior).
    //
    // Support: Extended
    AllowCredentials *bool `json:"allowCredentials,omitempty"`

    // AllowMethods indicates which HTTP methods are supported 
    // for accessing the requested resource.
    //
    // Valid values are any method defined by RFC9110, along with the special 
    // value `*`, which represents all HTTP methods are allowed.
    //
    // Method names are case-sensitive, so these values are also case-sensitive.
    // (See https://www.rfc-editor.org/rfc/rfc2616#section-5.1.1)
    //
    // Multiple method names in the value of the `Access-Control-Allow-Methods` 
    // response header are separated by a comma (",").
    //
    // A CORS-safelisted method is a method that is `GET`, `HEAD`, or `POST`.
    // (See https://fetch.spec.whatwg.org/#cors-safelisted-method)
    // The CORS-safelisted methods are always allowed, regardless of whether 
    // they are specified in the `AllowMethods` field.
    //
    // When the `AllowMethods` field is configured with one or more methods, 
    // the gateway must return the `Access-Control-Allow-Methods` response 
    // header which value is present in the `AllowMethods` field.
    //
    // If the HTTP method of the `Access-Control-Request-Method` request header 
    // is not included in the list of methods specified by the response header 
    // `Access-Control-Allow-Methods`, it will present an error on the client 
    // side.
    // 
    // Input:
    //   Access-Control-Request-Method: PUT
    //
    // Config:
    //   allowMethods: ["GET", "POST", "DELETE", "PATCH", "OPTIONS"]
    //
    // Output:
    //   Access-Control-Allow-Methods: GET, POST, DELETE, PATCH, OPTIONS
    //
    // The `Access-Control-Allow-Methods` response header can only use `*` 
    // wildcard as value when the `AllowCredentials` field is false.
    //
    // Input:
    //   Access-Control-Request-Method: PUT
    //
    // Config:
    //   allowMethods: ["*"]
    //
    // Output:
    //   Access-Control-Allow-Methods: *
    //
    // When the `AllowCredentials` field is true and the `AllowMethods`
    // field specified with the `*` wildcard, the gateway must specify one 
    // HTTP method in the value of the Access-Control-Allow-Methods response 
    // header. The value of the header `Access-Control-Allow-Methods` is same 
    // as the `Access-Control-Request-Method` header provided by the client. 
    // If the header `Access-Control-Request-Method` is not included in the 
    // request, the gateway will omit the `Access-Control-Allow-Methods` 
    // response header, instead of specifying the `*` wildcard. A Gateway 
    // implementation may choose to add implementation-specific default 
    // methods.
    //
    // Input:
    //   Access-Control-Request-Method: PUT
    //
    // Config:
    //   allowMethods: ["*"]
    //   allowCredentials: true
    //
    // Output:
    //   Access-Control-Allow-Methods: PUT
    //   Access-Control-Allow-Credentials: true
    //
    // Support: Extended
    //
    // +listType=set
    // +kubebuilder:validation:MaxItems=16
    AllowMethods []HTTPMethod `json:"allowMethods,omitempty"`
 
    // AllowHeaders indicates which HTTP request headers are supported 
    // for accessing the requested resource.
    //
    // Header names are not case-sensitive.
    //
    // Multiple header names in the value of the `Access-Control-Allow-Headers` 
    // response header are separated by a comma (",").
    //
    // When the `AllowHeaders` field is configured with one or more headers, 
    // the gateway must return the `Access-Control-Allow-Headers` response 
    // header which value is present in the `AllowHeaders` field.
    //
    // If any header name in the `Access-Control-Request-Headers` request header 
    // is not included in the list of header names specified by the response header 
    // `Access-Control-Allow-Headers`, it will present an error on the client side.
    //
    // If any header name in the `Access-Control-Allow-Headers` response header does 
    // not recognize by the client, it will also occur an error on the client side.
    //
    // Input:
    //   Access-Control-Request-Headers: Cache-Control, Content-Type
    //
    // Config:
    //   allowHeaders: ["DNT", "Keep-Alive", "User-Agent", "X-Requested-With", "If-Modified-Since", "Cache-Control", "Content-Type", "Range", "Authorization"]
    //
    // Output:
    //   Access-Control-Allow-Headers: DNT, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Range, Authorization
    //
    // A wildcard indicates that the requests with all HTTP headers are allowed.
    // The `Access-Control-Allow-Headers` response header can only use `*` wildcard 
    // as value when the `AllowCredentials` field is false.
    //
    // Input:
    //   Access-Control-Request-Headers: Content-Type, Cache-Control
    //
    // Config:
    //   allowHeaders: ["*"]
    //
    // Output:
    //   Access-Control-Allow-Headers: *
    //
    // When the `AllowCredentials` field is true and the `AllowHeaders` field
    // is specified with the `*` wildcard, the gateway must specify one or more
    // HTTP headers in the value of the `Access-Control-Allow-Headers` response 
    // header. The value of the header `Access-Control-Allow-Headers` is same as 
    // the `Access-Control-Request-Headers` header provided by the client. If 
    // the header `Access-Control-Request-Headers` is not included in the request, 
    // the gateway will omit the `Access-Control-Allow-Headers` response header, 
    // instead of specifying the `*` wildcard. A Gateway implementation may choose 
    // to add implementation-specific default headers.
    //
    // Input:
    //   Access-Control-Request-Headers: Content-Type, Cache-Control
    //
    // Config:
    //   allowHeaders: ["*"]
    //   allowCredentials: true
    //
    // Output:
    //   Access-Control-Allow-Headers: Content-Type, Cache-Control 
    //   Access-Control-Allow-Credentials: true
    //
    // Support: Extended
    //
    // +listType=set
    // +kubebuilder:validation:MaxItems=64
    AllowHeaders []string `json:"allowHeaders,omitempty"`

    // ExposeHeaders indicates which HTTP response headers can be exposed 
    // to client-side scripts in response to a cross-origin request.
    //
    // A CORS-safelisted response header is an HTTP header in a CORS response 
    // that it is considered safe to expose to the client scripts. 
    // The CORS-safelisted response headers include the following headers:
    // `Cache-Control`
    // `Content-Language`
    // `Content-Length`
    // `Content-Type`
    // `Expires`
    // `Last-Modified`
    // `Pragma`
    // (See https://fetch.spec.whatwg.org/#cors-safelisted-response-header-name)
    // The CORS-safelisted response headers are exposed to client by default.
    //
    // When an HTTP header name is specified using the `ExposeHeaders` field, this 
    // additional header will be exposed as part of the response to the client.
    //
    // Header names are not case-sensitive.
    //
    // Multiple header names in the value of the `Access-Control-Expose-Headers` 
    // response header are separated by a comma (",").
    //
    // Config:
    //   exposeHeaders: ["Content-Security-Policy", "Content-Encoding"]
    //
    // Output:
    //   Access-Control-Expose-Headers: Content-Security-Policy, Content-Encoding
    //
    // A wildcard indicates that the responses with all HTTP headers are exposed 
    // to clients. The `Access-Control-Expose-Headers` response header can only use 
    // `*` wildcard as value when the `AllowCredentials` field is false.
    //
    // Config:
    //   exposeHeaders: ["*"]
    //
    // Output:
    //   Access-Control-Expose-Headers: *
    //
    // Support: Extended
    //
    // +optional
    // +listType=set
    // +kubebuilder:validation:MaxItems=64
    ExposeHeaders []string `json:"exposeHeaders,omitempty"`

    // MaxAge indicates the duration (in seconds) for the client to cache 
    // the results of a "preflight" request.
    //
    // The information provided by the `Access-Control-Allow-Methods` and 
    // `Access-Control-Allow-Headers` response headers can be cached by the 
    // client until the time specified by `Access-Control-Max-Age` elapses.
    //
    // The default value of `Access-Control-Max-Age` response header is 
    // 5 (seconds). 
    //
    // When the `MaxAge` field is unspecified, the gateway sets the response 
    // header "Access-Control-Max-Age: 5" by default.
    //
    // Config:
    //   maxAge: 1728000
    //
    // Output:
    //   Access-Control-Max-Age: 1728000
    //
    // Support: Extended
    //
    // +optional
    // +kubebuilder:default=5
    // +kubebuilder:validation:Minimum=1
    MaxAge int32 `json:"maxAge,omitempty"`
}
```

## Examples

The following example shows how a HTTPRoute supports secure cross-origin requests and data transfers between clients and servers.

### Simple cross-origin interaction

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-cors
spec:
  hostnames:
  - http.route.cors.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service
      name: http-route-cors
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /resource/foo
    filters:
    - cors:
        allowOrigins:
        - *
        allowMethods: 
        - GET
        - HEAD
        - POST
        allowHeaders: 
        - Accept
        - Accept-Language
        - Content-Language
        - Content-Type
        - Range
      type: CORS
```

A client sends a GET request with cross-origin header `Origin`.
```
GET /resource/foo HTTP/1.1
Host: http.route.cors.com
Origin: https://foo.example

```

The cross-origin response headers will be added by the gateway to the response based on the above HTTPRoute.
The gateway returns an Access-Control-Allow-Origin header with "*", which means that the requested resource can be accessed from the any `Origin`.
```
HTTP/1.1 200 OK
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, HEAD, POST
Access-Control-Allow-Headers: Accept,Accept-Language,Content-Language,Content-Type,Range
```

###  Complex cross-origin interaction

Some HTTP requests are considered complex and require server confirmation before the actual cross-origin request is sent. Before the actual cross-origin requests, clients will initiate an extra "preflight" request to determine whether that the server will permit the actual requests.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-cors
spec:
  hostnames:
  - http.route.cors.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service
      name: http-route-cors
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /resource/foo
    filters:
    - cors:
        allowOrigins:
        - https://foo.example
        - http://foo.example
        allowCredentials: true
        allowMethods: 
        - GET
        - PUT
        - POST
        - DELETE
        - PATCH
        - OPTIONS
        allowHeaders: 
        - DNT
        - X-CustomHeader
        - Keep-Alive
        - User-Agent
        - X-Requested-With
        - If-Modified-Since
        - Cache-Control
        - Content-Type
        - Authorization
        exposeHeaders: 
        - Content-Security-Policy
        maxAge: 1728000
      type: CORS
```

A client sends a cross-origin "preflight" request for asking a server whether it would allow a PUT request before the actual cross-origin request is sent.
```
OPTIONS /resource/foo HTTP/1.1
Host: http.route.cors.com
Origin: https://foo.example
Access-Control-Request-Method: PUT
```

The status code of a successful response to a "preflight" request is an OK status (i.e., 204 or 200). 
Based on the above HTTPRoute, a successful "preflight" response will be generated by the gateway. 
Moreover, the gateway will send the "preflight" response to the client directly.

```
HTTP/1.1 204 No Content
Access-Control-Allow-Origin: https://foo.example
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization
Access-Control-Expose-Headers: Content-Security-Policy
Access-Control-Max-Age: 1728000
Content-Type: text/plain charset=UTF-8
Content-Length: 0
```

Then, the client will be able to send actual cross-origin request.
```
PUT /resource/foo HTTP/1.1
Host: http.route.cors.com
Keep-Alive: timeout=5, max=1000
Origin: https://foo.example
Authorization: Basic YWxhZGRpbjpvcGVuc2VzYW1l
```

At last, the cross-origin response headers will be added by the gateway to the response based on the HTTPRoute.
```
Access-Control-Allow-Origin: https://foo.example
Access-Control-Allow-Credentials: true
Access-Control-Allow-Methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
Access-Control-Allow-Headers: DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,Authorization
Access-Control-Expose-Headers: Content-Security-Policy
```

###  Disabling credentials

To disable credentials for cross-origin requests, simply don't set the
`allowCredentials` field at all. If you prefer to be explicit, you can
set it to `false`, although this will generally not be necessary:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-cors-no-credentials
spec:
  hostnames:
  - http.route.cors.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service
      name: http-route-cors
      port: 80
    matches:
    - path:
        type: PathPrefix
        value: /resource/bar
    filters:
    - cors:
        allowOrigins:
        - https://foo.example
        allowCredentials: false
        allowMethods: 
        - GET
        - POST
      type: CORS
```

Omitting the field, and setting it to `false` both mean `false`. In this
configuration the gateway will _not_ include the
`Access-Control-Allow-Credentials` header in responses.

## Prior Art
Some implementations already support CORS.

### Ingress-NGINX supports CORS with the following annotations:

* nginx.ingress.kubernetes.io/enable-cors enables CORS in an ingress rule. The default value is false.
* nginx.ingress.kubernetes.io/cors-allow-methods controls which HTTP methods are accepted by the server.
* nginx.ingress.kubernetes.io/cors-allow-headers controls which HTTP headers are accepted by the server.
* nginx.ingress.kubernetes.io/cors-expose-headers controls which headers are exposed to response.
* nginx.ingress.kubernetes.io/cors-allow-origin controls what's the allowed `Origin` for the requested resource.
* nginx.ingress.kubernetes.io/cors-allow-credentials controls whether the actual cross-origin request allows to include credentials. The default value is true.
* nginx.ingress.kubernetes.io/cors-max-age controls how long "preflight" requests can be cached by client. The default value is 1728000 seconds (i.e., 20 days).

For example, the following rule restricts cross-origin requests to the requested resource originating from https://foo.example using HTTP GET/PUT/POST/DELETE/PATCH/OPTIONS and sets Access-Control-Allow-Credentials header to true by default. HTTP headers "Origin,No-Cache,X-Requested-With,If-Modified-Since,Pragma,Last-Modified,Cache-Control,Expires,Content-Type,X-E4M-With,userId,token,authorization,groot-jwt,x-mokelay-custom-header,EagleEye-UserData,EagleEye-TraceId,EagleEye-RpcId,x-xsrf-token,cn-gw-custom-headers,X-XSRF-TOKEN,DNT,X-CustomHeader,Keep-Alive,User-Agent" can be used during the actual cross-origin request. The default value of header Access-Control-Max-Age is 20 days.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/enable-cors: "true"
    nginx.ingress.kubernetes.io/cors-allow-origin: 'https://foo.example'
    nginx.ingress.kubernetes.io/cors-allow-methods: GET, PUT, POST, DELETE, PATCH, OPTIONS
    nginx.ingress.kubernetes.io/cors-allow-headers: Origin,No-Cache,X-Requested-With,If-Modified-Since,Pragma,Last-Modified,Cache-Control,Expires,Content-Type,X-E4M-With,userId,token,authorization,groot-jwt,x-mokelay-custom-header,EagleEye-UserData,EagleEye-TraceId,EagleEye-RpcId,x-xsrf-token,cn-gw-custom-headers,X-XSRF-TOKEN,DNT,X-CustomHeader,Keep-Alive,User-Agent
  name: ingress-cors-test
spec:
  ingressClassName: nginx-default
  rules:
  - host: foo.cors.com
    http:
      paths:
      - path: /
        pathType: Exact
        backend:
          service:
            name: cors-test-service
            port:
              number: 80
```

### NGINX Ingress Controller supports CORS with CRDs VirtualServer and VirtualServerRoute resources:

* Action.Proxy.ResponseHeaders field of CRDs VirtualServer and VirtualServerRoute modifies the headers of the response to the client.

```yaml
responseHeaders:
  description: ProxyResponseHeaders defines the response 
               headers manipulation in an ActionProxy.
  properties:
    add:
      items:
        description: AddHeader defines an HTTP Header 
                     with an optional Always field to use with
                     the add_header NGINX directive.
        properties:
          always:
            type: boolean
          name:
            type: string
          value:
            type: string
        type: object
      type: array
```

* The name field is the name of the header. 
* The value field is the value of the header.
* If the always field is true, the header will be added regardless of the response code.

The following example shows how the VirtualServer and VirtualServerRoute resources supports enabling and configuring CORS. 
The rule restricts cross-origin requests to the requested resource originating from https://foo.example using HTTP GET/POST/PATCH, and sets the  Access-Control-Allow-Credentials header to true.
Moreover, HTTP headers "Origin,No-Cache,X-Requested-With,If-Modified-Since,Pragma" can be used during the actual cross-origin request. 
At last, it sets an expiry period of 48 hours for the client to cache the results of a "preflight" request.

```yaml
responseHeaders:
  add:
    - name: Access-Control-Allow-Origin
      value: "https://foo.example"
    - name: Access-Control-Allow-Credentials
      value: "true"
    - name: Access-Control-Allow-Methods
      value: "GET, POST, PATCH"
    - name: Access-Control-Allow-Headers
      value: "Origin,No-Cache,X-Requested-With,If-Modified-Since,Pragma"
    - name: Access-Control-Max-Age
      value: "172800"
```

### Istio supports CORS with CRD VirtualService:

* allowOrigins: string patterns that match allowed origins.
* allowMethods: a list of HTTP methods allowed to access the resource.
* allowHeaders: a list of HTTP headers that can be used when requesting the resource.
* exposeHeaders: a list of HTTP headers that the browsers are allowed to access.
* maxAge: how long the results of a preflight request can be cached.
* allowCredentials: whether the caller is allowed to send the actual request (not the preflight) using credentials.
* unmatchedPreflights: whether preflight requests not matching the configured allowed origin shouldn’t be forwarded to the upstream. 

```yaml
http:
  description: An ordered list of route rules for HTTP traffic.
  items:
    properties:
      corsPolicy:
        description: Cross-Origin Resource Sharing policy (CORS).
        properties:
          allowCredentials:
            description: Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials.
            nullable: true
            type: boolean
          allowHeaders:
            description: List of HTTP headers that can be used when requesting the resource.
            items:
              type: string
            type: array
          allowMethods:
            description: List of HTTP methods allowed to access the resource.
            items:
              type: string
            type: array
          allowOrigins:
            description: String patterns that match allowed origins.
            items:
              oneOf:
              - not:
                  anyOf:
                  - required:
                    - exact
                  - required:
                    - prefix
                  - required:
                    - regex
              - required:
                - exact
              - required:
                - prefix
              - required:
                - regex
              properties:
                exact:
                  type: string
                prefix:
                  type: string
                regex:
                  description: '[RE2 style regex-based match](https://github.com/google/re2/wiki/Syntax).'
                  type: string
              type: object
            type: array
          exposeHeaders:
            description: A list of HTTP headers that the browsers are allowed to access.
            items:
              type: string
            type: array
          maxAge:
            description: Specifies how long the results of a preflight request can be cached.
            type: string
            x-kubernetes-validations:
            - message: must be a valid duration greater than 1ms
              rule: duration(self) >= duration('1ms')
          unmatchedPreflights:
            description: |-
              Indicates whether preflight requests not matching the configured allowed origin shouldn't be forwarded to the upstream.
              
              Valid Options: FORWARD, IGNORE
            enum:
            - UNSPECIFIED
            - FORWARD
            - IGNORE
            type: string
        type: object
```

In the following example, the rule restricts cross-origin requests to the requested resource originating from https://foo.example and http://foo.example using HTTP GET/HEAD/POST. It sets the Access-Control-Allow-Credentials header to false. The credentials will not be allowed in the actual cross-origin requests.
Moreover, only the HTTP header "X-Foo-Example" can be used during the actual cross-origin request. At last, it sets an expiry period of 1 day for the browser to cache the results of a "preflight" request.

```yaml
apiVersion: networking.istio.io/v1
kind: VirtualService
metadata:
  name: http-route-cors
spec:
  hosts:
  - http.route.cors.com
  http:
  - route:
    - destination:
        host: cors.svc.cluster.local
        subset: v1
    corsPolicy:
      allowOrigins:
      - exact: https://foo.example
      - exact: http://foo.example
      allowMethods:
      - GET
      - HEAD
      - POST
      allowCredentials: false
      allowHeaders:
      - X-Foo-Example
      maxAge: "24h"
```

### Traefik supports CORS with CRD Middleware:

* CORS Headers of HTTP Headers middleware manages the cross-origin response headers.

```yaml
headers:
description: |-
  Headers holds the headers middleware configuration.
  This middleware manages the requests and responses headers.
  More info: https://doc.traefik.io/traefik/v3.2/middlewares/http/headers/#customrequestheaders
properties:
  accessControlAllowCredentials:
    description: AccessControlAllowCredentials defines whether the
                 request can include user credentials.
    type: boolean
  accessControlAllowHeaders:
    description: AccessControlAllowHeaders defines the Access-Control-Allow-Headers
                 values sent in preflight response.
    items:
      type: string
    type: array
  accessControlAllowMethods:
    description: AccessControlAllowMethods defines the Access-Control-Allow-Methods
                 values sent in preflight response.
    items:
      type: string
    type: array
  accessControlAllowOriginList:
    description: AccessControlAllowOriginList is a list of allowable
                 origins. Can also be a wildcard origin "*".
    items:
      type: string
    type: array
  accessControlAllowOriginListRegex:
    description: AccessControlAllowOriginListRegex is a list of allowable
                 origins written following the Regular Expression syntax (https://golang.org/pkg/regexp/).
    items:
      type: string
    type: array
  accessControlExposeHeaders:
    description: AccessControlExposeHeaders defines the Access-Control-Expose-Headers
                 values sent in preflight response.
    items:
      type: string
    type: array
  accessControlMaxAge:
    description: AccessControlMaxAge defines the time that a preflight
                 request may be cached.
    format: int64
    type: integer
  addVaryHeader:
    description: AddVaryHeader defines whether the Vary header is
                 automatically added/updated when the AccessControlAllowOriginList
                 is set.
    type: boolean
```

* accessControlAllowCredentials: whether the request can include user credentials.
* accessControlAllowHeaders: a list of headers can be used during cross-origin requests.
* accessControlAllowMethods: a list of methods can be used during cross-origin requests.
* accessControlAllowOriginList: a list of allowable origins.
* accessControlAllowOriginListRegex: a list of allowable origins written following the Regular Expression syntax.
* accessControlExposeHeaders: which headers are safe to expose to the api of a CORS API specification.
* accessControlMaxAge: how many seconds a preflight request can be cached.
* addVaryHeader: whether the Vary header should be added or modified to demonstrate that server responses can differ based on the value of the origin header.

In the example below, the rule restricts cross-origin requests to the requested resource originating from https://foo.example and http://foo.example using HTTP GET/OPTIONS/PUT. It specifies the response header Access-Control-Allow-Credentials to true, the actual cross-origin request will be able to use credentials for getting sensitive resources. Moreover, all HTTP headers can be used during the actual cross-origin request. At last, it sets an expiry period of 100 seconds for the browser to cache the results of a "preflight" request.

```yaml
apiVersion: traefik.io/v1alpha1
kind: Middleware
metadata:
  name: cors-header
spec:
  headers:
    accessControlAllowMethods:
      - "GET"
      - "OPTIONS"
      - "PUT"
    accessControlAllowHeaders:
      - "*"
    accessControlAllowOriginList:
      - "https://foo.example"
      - "http://foo.example"
     accessControlAllowCredentials: true
    accessControlMaxAge: 100
    addVaryHeader: true
```

## Alternatives Considered
### Use top level field in HTTPRouteRule to implement CORS
As suggested in [this comment](https://github.com/kubernetes-sigs/gateway-api/pull/3435/#discussion_r1874186947), a new field `CORS` is added to `HTTPRouteRule`, which allows for handling the cross-origin requests before the response is sent to the client. This is similar to the `Timeouts` field in `HTTPRouteRule`. 

```golang
type HTTPRouteRule struct {
  Name *SectionName `json:"name,omitempty"`

  Matches []HTTPRouteMatch `json:"matches,omitempty"`

  Filters []HTTPRouteFilter `json:"filters,omitempty"`

  BackendRefs []HTTPBackendRef `json:"backendRefs,omitempty"`

  Timeouts *HTTPRouteTimeouts `json:"timeouts,omitempty"`

  Retry *HTTPRouteRetry `json:"retry,omitempty"`

  SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`

  // CORS defines the CORS rules that respond to the 
  // cross-origin request based on HTTP response header.
  //
  // Support: Extended
  //
  // +optional
  CORS *HTTPCORS `json:"cors,omitempty"`
}
```

If a filter logically as a 
```rust
fn run_filter(req: HTTPRequest) -> FilterResponse;
enum FilterResponse {
  Request(HTTPRequest),
  Response(HTTPResponse)
```

A `Timeouts` doesn't really meet that, but `CORS` does:
```rust
fn run_filter(req: HTTPRequest) -> FilterResponse {
  if req.method == origin { return Response(cors_response()) }
    return Request(req)
}
```

Moreover, CORS is a HTTP feature based on HTTP-header. This fits as a filter.

## References
* [RFC2616](https://www.rfc-editor.org/rfc/rfc2616)
* [RFC6454](https://www.rfc-editor.org/rfc/rfc6454)
* [RFC7230](https://www.rfc-editor.org/rfc/rfc7230)
* [RFC9110](https://www.rfc-editor.org/rfc/rfc9110)
* [RFC9111](https://www.rfc-editor.org/rfc/rfc9111)
* [Fetch Living Standard](https://fetch.spec.whatwg.org)
* [Ingress-NGINX `Enable CORS`](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#enable-cors)
* [NGINX Ingress Controller `ResponseHeaders`](https://docs.nginx.com/nginx-ingress-controller/configuration/virtualserver-and-virtualserverroute-resources/#actionproxyresponseheaders)
* [Istio `CorsPolicy`](https://istio.io/latest/docs/reference/config/networking/virtual-service/#CorsPolicy)
* [Traefik `CORS Headers`](https://doc.traefik.io/traefik/middlewares/http/headers/#cors-headers)
