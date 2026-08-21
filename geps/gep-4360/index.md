---
title: "GEP-4360: Regex Path Rewrites"
---

* Issue: [#4359](https://github.com/kubernetes-sigs/gateway-api/issues/4359)
* Status: Provisional

## TLDR

Right now Gateway API supports only full path or prefix rewrites, we want to extend it to regex-based path rewrites.
This is already supported by [Envoy](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite),
[NGINX](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#rewrite),
and [HAProxy](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html#4.2-http-request%20replace-path);
in this proposal we are closing the gap between Gateway API and current capabilities of the modern LBs.

## Goals

Close the regex-based path rewrites feature gap for Gateway API, i.e.:

 * Rewrite the path of a request based on a regular expression, regardless of initial match type
 * Substitute matching section(s) in the regular expression with predefined values

## Non-Goals

  * Any sort of host rewriting

## Introduction/Overview

We would like to add an enhancement to the HTTPURLRewriteFilter that would allow the caller to specify path rewrite based on the provided pattern and substitution.
Right now Gateway API supports only full path or prefix rewrites, we want to extend it taking into account capabilities of the modern LBs.

## Purpose (Why and Who)

This is a highly requested feature. This is also supported by Envoy, NGINX, and HAProxy.

In this proposal we are closing the gap between Gateway API and current capabilities of the modern LBs.

## Implementation and Support

| Implementation | Support | Engine |
|----------------|------------|----------------|
| Envoy | [config.route.v3.RouteAction.regex_rewrite](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto#envoy-v3-api-field-config-route-v3-routeaction-regex-rewrite) | RE2 |
| HAProxy | [http-request replace-path](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html#4.2-http-request%20replace-path) | PCRE |
| NGINX | [ngx_http_rewrite_module.html#rewrite](https://nginx.org/en/docs/http/ngx_http_rewrite_module.html#rewrite) | PCRE |

NGINX only replaces the first match of the pattern using the rewrite directive, but you can get full substitution using Lua.

## API

This is a provisional GEP, so no specific API details, but at a high level there will be two fields: `pattern` and `substitution`.
`pattern` will be a [GEP-4359: Gateway API Regex](../gep-4359/index.md) (this is different from a path match of type RegularExpression).
**ALL** instances of `pattern` in the url path MUST be replaced with `substitution`.
