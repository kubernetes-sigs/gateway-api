# GEP 2891: HTTP Cookie Match

* Issue: [#2891](https://github.com/kubernetes-sigs/gateway-api/issues/2891)
* Status: Experimental

## TLDR

Just like HTTP route based on header and query parameter is common, it’d be helpful to have a `HTTPCookieMatch` field in `HTTPRouteMatch` which would let gateway forwards request based on cookie to the certain backends.

## Goals

* Support cookie matching in a `HTTPRoute`
* Add a new match type `List`

## Introduction

Cookie is an essential part of the HTTP request. The Cookie header has multiple cookie-pair which contains the cookie-name and cookie-value.
```
cookie-header = "Cookie:" OWS cookie-string OWS
cookie-string = cookie-pair *( ";" SP cookie-pair )
cookie-pair   = cookie-name "=" cookie-value
cookie-name   = token
cookie-value  = *cookie-octet / ( DQUOTE *cookie-octet DQUOTE )
cookie-octet  = %x21 / %x23-2B / %x2D-3A / %x3C-5B / %x5D-7E
                  ; US-ASCII characters excluding CTLs,
                  ; whitespace DQUOTE, comma, semicolon,
                  ; and backslash
token         = <token, defined in [[RFC2616], Section 2.2](https://www.rfc-editor.org/rfc/rfc2616#section-2.2)>
```

They are used to maintain state and identify specific users. Moreover, cookies, headers and query parameters are common techniques used in a canary release.
Currently `HTTPRouteMatch` API supports the following condition: path, method, header and query parameter. This GEP proposes adding support for cookie matching in a `HTTPRoute`.

## API

This GEP proposes to add a new field `HTTPCookieMatch` to `HTTPRouteMatch`. 
Moreover, a new match type `List` is added. Matches if the value of the cookie with name field is present in a list of strings. This match type `List` can be applied to serving `HTTPHeaderMatch` and `HTTPQueryParamMatch` as well. 

The `HTTPCookieMatch` and `List` are considered an extended feature.

```golang
// CookieMatchType specifies the semantics of how HTTP cookie values should be
// compared. Valid CookieMatchType values, along with their conformance levels, are:
//
// * "Exact" - Core
// * "List" - Extended
// * "RegularExpression" - Implementation Specific
//
// * "Exact" matching exact string
// * "List" matching string in a list of strings
//
// Note that values may be added to this enum, implementations
// must ensure that unknown values will not cause a crash.
//
// Unknown values here must result in the implementation setting the
// Accepted Condition for the Route to `status: False`, with a
// Reason of `UnsupportedValue`.
//
// +kubebuilder:validation:Enum=Exact;RegularExpression
type CookieMatchType string

// CookieMatchType constants.
const (
	CookieMatchExact             CookieMatchType = "Exact"
	CookieMatchList              CookieMatchType = "List"
	CookieMatchRegularExpression CookieMatchType = "RegularExpression"
)

// HTTPCookieMatch describes how to select a HTTP route by matching HTTP request
// cookies.
type HTTPCookieMatch struct {
	// Type specifies how to match against the value of the cookie.
	//
	// Support: Core (Exact)
	//
	// Support: Extended (List)
	//
	// Support: Implementation-specific (RegularExpression)
	//
	// Since RegularExpression CookieMatchType has implementation-specific
	// conformance, implementations can support POSIX, PCRE or any other dialects
	// of regular expressions. Please read the implementation's documentation to
	// determine the supported dialect.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *CookieMatchType `json:"type,omitempty"`

	// Name is the cookie-name of the cookie-pair in the HTTP Cookie header to be matched.
	// The cookie names are case-sensitive. This must be an exact string match. (See
	// https://www.rfc-editor.org/rfc/rfc6265)
	//
	// cookie-header = "Cookie:" OWS cookie-string OWS
	// cookie-string = cookie-pair *( ";" SP cookie-pair )
	// cookie-pair   = cookie-name "=" cookie-value
	// cookie-name   = token
	// token         = <token, defined in [RFC2616], Section 2.2>
	//
	// If the cookie-name is empty, ignore this HTTPCookieMatch entirely.
	//
	// If multiple entries specify equivalent cookie names, only the first
	// entry with an equivalent name MUST be considered for a match. Subsequent
	// entries with an equivalent cookie name MUST be ignored. Due to the
	// case-sensitive of cookie names, "foo" and "Foo" are considered different
	// cookie name.
	Name HTTPHeaderName `json:"name"`

	// Values is the cookie-value of the cookie-pair in the HTTP Cookie header to be matched.
	// Matches if the value of the cookie with name field is present in the HTTP Cookie header.
	// The cookie-value is always case-sensitive. This must be an exact string match.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Value string `json:"value"`

	// Values are the cookie-value list of the cookie-pair in the HTTP Cookie header to be matched.
	// Matches if the value of the cookie with name field is present in the list.
	//
	// +optional
	// +listType=set
	// +kubebuilder:validation:MaxItems=16
	Values []string `json:"values"`
}
```

## Examples

The following example shows how the HTTPRoute matches request on cookie and forward it to different backend service.

* Backend service wants to select specific users for measuring the effectiveness of advertising campaigns. With the following HTTPRoute, the http requests with cookie name "unb" and cookie value in a list of strings (i.e., 2426168118, 2208203664638, 2797880990, 70772956, 2215140160618) will be routed to the service "http-route-canary-campaign:7001".

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-cookie
spec:
  hostnames:
  - http.route.cookies.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service 
      name: http-route-production
      port: 7001
    matches:
    - path:
        type: PathPrefix
        value: /
  - backendRefs:
    - kind: Service
      name: http-route-canary-campaign
      port: 7001
    matches:
    - cookies:
      - name: unb
        type: List
        values: 
        - 2426168118
        - 2208203664638
        - 2797880990
        - 70772956
        - 2215140160618
```

* This HTTPRoute directs incoming HTTP requests with cookie "gray=true" to the canary service "http-site-canary:80".

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: http-route-cookie
spec:
  hostnames:
  - http.site.cookie.com
  - http.site.cookies.com
  parentRefs:
  - group: gateway.networking.k8s.io
    kind: Gateway
    name: http-gateway
  rules:
  - backendRefs:
    - kind: Service
      name: http-site-production
      port: 80 
    matches:
    - path:
        type: PathPrefix
        value: /
  - backendRefs:
    - kind: Service
      name: http-site-canary
      port: 80
    matches:
    - cookies:
      - name: gray
        type: Exact
        value: true
```

## Prior Art
Some implementations already support HTTP cookie match.

### Ingress-nginx supports this with an annotation nginx.ingress.kubernetes.io/canary-by-cookie

* The cookie to use for notifying the Ingress to route the request to the service specified in the Canary Ingress. When the cookie value of this annotation is set to always, it will be routed to the canary. When the cookie value of the annotation is set to never, it will never be routed to the canary.

For example, if the Cookie header has a cookie-pair "cookie-test=always", the request will be forwarded to the backend service "cookie-test-service:80".

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-by-cookie: "cookie-test"
  name: ingress-canary-cookie-test
spec:
  ingressClassName: nginx-default
  rules:
    - host: canary.cookie.com
      http:
        paths:
          - path: /
            pathType: Exact
            backend:
              service:
                name: cookie-test-service
                port:
                  number: 80
```

### Nginx-ingress-controller supports this with CRDs

* The Match section of VirtualServer or VirtualServerRoute defines a match between conditions and an action or splits. The Condition defines a condition in a match. The Condition includes header, cookie, argument or variable.

In the example below, NGINX routes requests with the path "/coffee" to different upstreams based on the value of the cookie "user":

user=john -> coffee-future
user=bob -> coffee-deprecated

If the cookie is not set or not equal to either "john" or "bob", NGINX routes to "coffee-stable".

```yaml
path: /coffee
matches:
- conditions:
  - cookie: user
    value: john
  action:
    pass: coffee-future
- conditions:
  - cookie: user
    value: bob
  action:
    pass: coffee-deprecated
action:
  pass: coffee-stable
```

### Tengine-ingress supports this with annotations nginx.ingress.kubernetes.io/canary-by-cookie and nginx.ingress.kubernetes.io/canary-by-cookie-value

* The annotation nginx.ingress.kubernetes.io/canary-by-cookie sets the cookie name.
* The annotation nginx.ingress.kubernetes.io/canary-by-cookie-value sets the multiple cookie values.

In the following example, if the value of the cookie user is "mike" or "bob", the requests with the path "/gray" will be forwarded to the service "cookie-net-service:80".

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-by-cookie: user
    nginx.ingress.kubernetes.io/canary-by-cookie-value: mike||bob
  name: tengine-ingress-cookie-value
spec:
  rules:
  - host: tengine.cookie.net
    http:
      paths:
      - backend:
          service:
            name: cookie-net-service
            port:
              number: 80
        path: /gray
        pathType: ImplementationSpecific
```

## References
* [Ingress-nginx `canary`](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#canary)
* [Kubernetes-ingress `Condition`](https://docs.nginx.com/nginx-ingress-controller/configuration/virtualserver-and-virtualserverroute-resources/#condition)
* [Tengine-ingress / `ingress_routes`](https://tengine.taobao.org/document/ingress_routes.html)
* [RFC6265](https://www.rfc-editor.org/rfc/rfc6265)
