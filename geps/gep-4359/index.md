# GEP-696: GEP template

* Issue: [#4359](https://github.com/kubernetes-sigs/gateway-api/issues/4359)
* Status: Provisional

## TLDR

Right now Gateway API supports only full path or prefix rewrites, we want to extend it to regex-based path rewrites. This is already supported by [Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite), [NGINX](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#rewrite), and [HAProxy](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html#4.2-http-request%20replace-path); in this proposal we are closing the gap between Gateway API and current capabilities of the modern LBs.

## Goals

Close the feature gap for Gateway API.

## Introduction/Overview

We would like to add an enhancement to the HTTPURLRewriteFilter that would allow the caller to specify path rewrite based on the provided pattern and substitution. Right now Gateway API supports only full path or prefix rewrites, we want to extend it taking into account capabilities of the modern LBs.

## Purpose (Why and Who)

This is a highly requested feature. This is also supported by Envoy, NGINX, and HAProxy.

In this proposal we are closing the gap between Gateway API and current capabilities of the modern LBs.

## Implementation and Support

| Implementation | Support |
|----------------|------------|
| Envoy | [config.route.v3.RouteAction.regex_rewrite](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite) |
| HAProxy | [http-request replace-path](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html#4.2-http-request%20replace-path) |
| NGINX | [ngx_http_rewrite_module.html#rewrite](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#rewrite) |

## API

This GEP proposes the following API changes:

* Update google3/third_party/golang/sigs_k8s_io/gateway_api/v/v1/apis/v1/httproute_types.go by adding and new HTTPRegexModifier field

```go
type HTTPURLRewriteFilter struct {
	// Hostname is the value to be used to replace the Host header value during
	// forwarding.
	//
	// Support: Extended
	//
	// +optional
	Hostname *PreciseHostname `json:"hostname,omitempty"`

	// Path defines a path rewrite.
	//
	// Support: Extended
	//
	// +optional
	Path *HTTPPathModifier `json:"path,omitempty"`

	// RegexModifier defines a regex-based host and/or path rewrite.
	//
	// Support: Extended
	//
	// +optional
	RegexModifier *HTTPRegexModifier `json:"regexModifier,omitempty"`
}
```

* Update google3/third_party/golang/sigs_k8s_io/gateway_api/v/v1/apis/v1/httproute_types.go by adding a new HTTPRegexModifier struct

```go
type HTTPRegexModifier struct {
  // +optional
  // +kubebuilder:validation:Minimum=0
  // +kubebuilder:validation:Maximum=1024
	PathPattern *string `json:"pathPattern,omitempty"`
  // +optional
  // +kubebuilder:validation:Minimum=0
  // +kubebuilder:validation:Maximum=1024
	PathSubstitute *string `json:"pathSubstitute,omitempty"`
}
```

### Example

```
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: store-external
spec:
  parentRefs:
  - kind: Gateway
    name: external-http
  hostnames:
  - "*"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: "/"
    filters:
      - type: URLRewrite
        urlRewrite:
          regexModifier:
            pathPattern: "^/region/(?<region>[a-z]+)/bucket/(?<storage>[a-zA-Z0-9-]+)/(?<object>[a-z]+)\\.pdf$"
            pathSubstitution: "\\g<region>/bucket-\\g<storage>/\\g<object>.pdf"
    backendRefs:
    - name: store-v2
      port: 8080
```

## Conformance Details

These tests will excersie the regex-based path rewrites.

### Conformance test scenarios

#### Example test scenarios

A HTTPRoute with a URLRewrite filter should rewrite the path according to
the specification, routing traffic to the backend.

* A Regex Modifier with `pathPattern ="^/region/(?<region>[a-z]+)/bucket/(?<storage>[a-zA-Z0-9-]+)/(?<object>[a-z]+)\\.pdf$"` and `pathSubstitution ="\\g<region>/bucket-\\g<storage>/\\g<object>.pdf"` should route requests
  to `/region/eu/bucket/prod-storage/object.pdf` to `/eu/bucket-prod-storage/object.pdf` instead.
* A Regex Modifier with `pathPattern="^/service/([^/]+)(/.*)$"` and `pathSubstitution="\2/instance/\1"` should route requests `/service/foo/v1/api` to `/v1/api/instance/foo`.
* A Regex Modifier with `pathPattern="one"` and `pathSubstitution="two"` should route requests to `/xxx/one/yyy/one/zzz` to `/xxx/two/yyy/two/zzz`.
* A Regex Modifier with `pathPattern="^(.*?)one(.*)$"` and `pathSubstitution="\1two\2"` should route requests to `/xxx/one/yyy/one/zzz` to `/xxx/two/yyy/one/zzz`.
* A Regex Modifier with `pathPattern="(?i)/xxx/"` and `pathSubstitution="/yyy/"` should route requests to `/aaa/XxX/bbb` to `/aaa/yyy/bbb`.

## `Standard` Graduation Criteria

( This section outlines the criteria required for graduation to Standard. It MUST
contain at least the items in the template, but more MAY be added if necessary. )

( Required for Experimental status and above)

* At least one Feature Name must be listed.
* The `Conformance Details` must be filled out, with conformance test scenarios listed.
* Conformance tests must be implemented that test all the listed test scenarios.
* At least three (3) implementations must have submitted conformance reports that pass
  those conformance tests.
* At least six months must have passed from when the GEP moved to `Experimental`.


## Future extension

The future extensions we see for this GEP is regex-based host rewrites and regex-based host-to-path. We could extend `HTTPRegexModifier` with two additional fields, e.g. `hostPattern` and `hostSubstitution`.

### Example

```
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: store-external
spec:
  parentRefs:
  - kind: Gateway
    name: external-http
  hostnames:
  - "*"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: "/"
    filters:
      - type: URLRewrite
        urlRewrite:
          regexModifier:
            hostPattern: "^(?<apiversion>[a-z0-9]+)\\.domain\\.com$"
            hostSubstitution: "domain2.com"
            pathPattern: "^/region/(?<region>[a-z]+)/bucket/(?<storage>[a-zA-Z0-9-]+)/(?<object>[a-z]+)\\.pdf$"
            pathSubstitution: "\\g<region>/bucket-\\g<storage>/\\g<object>.pdf"
    backendRefs:
    - name: store-v2
      port: 8080
```

* A Regex Modifier with `hostPattern="^(?<apiversion>[a-z0-9]+)\\.domain\\.com$"`, `hostSubstitutuon="domain2.com"`, `pathPattern="^/(?<path>[a-zA-Z0-9/._]+)$"`, and `pathSubstitution="/backend/\\g<apiversion>/\\g<path>"` should route requests to `https://api1.domain.com/path1/path2` to `https://domain2.com/backend/api1/path1/path2`.