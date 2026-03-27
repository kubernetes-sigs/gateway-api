# GEP-696: GEP template

* Issue: [#4359](https://github.com/kubernetes-sigs/gateway-api/issues/4359)
* Status: Provisional

## TLDR

Right now Gateway API supports only full path or prefix rewrites, we want to extend it to regex-based path rewrites. This is already supported by [Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite), [NGINX](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#rewrite), and [HAProxy](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html#4.2-http-request%20replace-path); in this proposal we are closing the gap between Gateway API and current capabilities of the modern LBs.

## Goals

Close the regex-based path rewrites feature gap for Gateway API, i.e.:

 * Rewrite the path of a request based on a regular expression, regardless of initial match type
 * Substitute matching section(s) in the regular expression with predefined values

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

NGINX only replaces the first match of the patter using the rewrite directive, but you can get full substitution using Lua.

## API

This is a provisional GEP, so no specific API details, but at a high level there will be two fields: `pattern` and `substitution`.
`pattern` will be a regular expression.
**ALL** instances of `pattern` in the url path MUST be replaced with `substitution`.
The url path includes the leading slash.
If the resulting path is invalid (does not have a leading slash, etc), the proxy MUST return a 500 Internal Server Error.

To be consistent with [HTTP path match](https://gateway-api.sigs.k8s.io/reference/spec/#pathmatchtype), we let implementation define what flavor of regex they will support, and what features they might support.
For example, an implementation might disallow certain characters that they deem an injection risk.
Others might not allow capture groups in the `substitution`.

If the implementation deems `pattern` or `substitution` to be invalid (contains illegal characters, unmatched parenthesis, unsupported features),
it MUST not accept the HTTPRoute with a `Reason` of `invalid` and a descriptive `Message`.


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
