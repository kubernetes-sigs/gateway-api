# GEP-4359: Regex Rewrites

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
 * Define a common regex syntax that all implementations should support
 * Define capture group references in the substitution string

## Non-Goals

  * Any sort of host rewriting
  * Limit regex features to a common subset.

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
`pattern` will be a regular expression (this is different from a path match of type RegularExpression).
**ALL** instances of `pattern` in the url path MUST be replaced with `substitution`.
See below for more details on the regex flavor and substitution syntax.
The URL path INCLUDES the leading slash when matching against the pattern.
We do not define what happens if the path is invalid after the substitution (e.g. missing leading slash)

If the implementation deems `pattern` or `substitution` to be invalid (contains illegal characters, unmatched parenthesis, unsupported features),
the implementation MUST consider the rule to be invalid.


Here are some examples

| `pattern`                   | `substitution` |  Input path  | Output path  |
| ------------                | -------------- | ------------ | ------------ |
| `a`                         | `c`            | `/aba`       | `/cbc`       |
| `^/a`                       | `/c`           | `/aba`       | `/cba`       |
| `^/a`                       | `/c`           | `/aba`       | `/cba`       |
| `^/a`                       | `c`            | `/aba`       |  Undefined   |
| `a$`                        | `c`            | `/aba`        | `/abc`       |
| `^/([A-Za-z]+)/([A-Za-z]+)` | `/\2/\1`       | `/my/path`   | `/path/my`   |

### Regex Flavor

Because regex flavors differ in features, we define a common denominator to ensure portability.
Each implementation's regex flavor MUST be a superset of [IEEE POSIX ERE](https://pubs.opengroup.org/onlinepubs/9699919799/basedefs/V1_chap09.html), with a few exceptions:

* We do not define whether matches are leftmost-longest, leftmost-first, or something else.
* Collating symbols. For example, in some locales, the string `ch` is considered to be one character, and `[[.ch.]]` will match it.
* Equivalence classes. For example, in some locales `a`, `A`, and other unicode variations of `a` are considered to be equivalent, and `[[=a=]]` will match all of them.
* Character classes. For example, `[:alpha:]`.
* Any other locale-specific behavior will assume the [C/POSIX locale](https://pubs.opengroup.org/onlinepubs/7908799/xbd/locale.html) (e.g. character ordering).

IEEE POSIX ERE is a good common denominator because
* The set of supported features is small (e.g. no backreferences)
* Broadly compatible across regex engines (RE2, rust-lang/regex, PCRE), especially because we don't have to worry about unicode, line breaks, and control sequences.

Although there is no single source saying POSIX ERE with above exceptions is supported by all three engines, we have verified that the features we need are supported by all three engines, it certainly seems to be the case:
* They all support the special regex characters (e.g. `*`, `+`, `?`, `|`, `()`, `[]`, `^`, `$`).
* Special regex characters can be escaped with `\` in all three engines. 
* Special regex characters lose their special meaning when they are inside a character class (e.g. `[*]` matches `*`).
* The precedence of operators is the same (though there doesn't seem to be a neat chart for this in PCRE)

For the substitution string, implementations MUST allow `\1`, `\2`, etc to reference capturing groups in the pattern.
The ordering of the capturing groups MUST be determined by the order of the opening parentheses in the pattern.

