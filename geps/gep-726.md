# GEP-726: Add Path Redirects and Rewrites

* Issue: [#726](https://github.com/kubernetes-sigs/gateway-api/issues/726)
* Status: Standard

## TLDR

This GEP proposes adding support for path redirects and rewrites in addition to
host rewrites. This would augment the existing host redirection capabilities.

## Goals

* Implement path redirects.
* Implement the most portable and simple forms of path rewrites.
* Describe how more advanced rewrite and redirect and redirect capabilities
  could be added in the future.

## API

Although many implementations support very advanced rewrite and redirect
capabilities, the following are the most [portable](#portability) concepts that
are not already supported by the Gateway API:

* Path redirects
* Path prefix redirects
* Path prefix rewrites
* Host rewrites

Although regular expression based redirects and rewrites are commonly supported,
there is significantly more variation in both if and how they are implemented.
Given the wide support for this concept, it is important to design the API in a
way that would make it easy to add this capability in the future.

### Path Modifiers

Both redirects and rewrites would share the same `PathModifier` types:

```go
// HTTPPathModifierType defines the type of path redirect.
type HTTPPathModifierType string

const (
  // This type of modifier indicates that the complete path will be replaced by
  // the path redirect value.
  AbsoluteHTTPPathModifier HTTPPathModifierType = "Absolute"

  // This type of modifier indicates that any prefix path matches will be
  // replaced by the substitution value. For example, a path with a prefix match
  // of "/foo" and a ReplacePrefixMatch substitution of "/bar" will have the
  // "/foo" prefix replaced with "/bar" in matching requests.
  PrefixMatchHTTPPathModifier HTTPPathModifierType = "ReplacePrefixMatch"
)

// HTTPPathModifier defines configuration for path modifiers.
type HTTPPathModifier struct {
  // Type defines the type of path modifier.
  //
  // +kubebuilder:validation:Enum=Absolute;ReplacePrefixMatch
  Type HTTPPathModifierType `json:"type"`

  // Substitution defines the HTTP path value to substitute. An empty value ("")
  // indicates that the portion of the path to be changed should be removed from
  // the resulting path. For example, a request to "/foo/bar" with a prefix
  // match of "/foo" would be modified to "/bar".
  //
  // +kubebuilder:validation:MaxLength=1024
  Substitution string `json:"substitution"`
}
```

### Redirects

The existing `RequestRedirect` filter can be expanded to support path redirects.
In the following example, a request to `/foo/abc` would be redirected to
`/bar/abc`.

```yaml
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: http-filter-1
spec:
  rules:
    - matches:
      - path:
          type: Prefix
          value: /foo
      filters:
      - type: RequestRedirect
        requestRedirect:
          hostname: foo.com
          path:
            type: ReplacePrefixMatch
            value: /bar
```

This would be represented with the following API addition to the existing
HTTPRequestRedirect filter:
```go
// HTTPRequestRedirect defines a filter that redirects a request. At most one of
// these filters may be used on a Route rule. This may not be used on the same
// Route rule as a HTTPRequestRewrite filter.
//
// Support: Extended
type HTTPRequestRedirect struct {
  // Path defines a path redirect.
  //
  // Support: Extended
  //
  // +optional
  Path *HTTPPathModifier `json:"path,omitempty"`
  // ...
}
```

### Rewrites

A new `URLRewrite` filter can be added to support rewrites. In the following
example, a request to `example.com/foo/abc` would be rewritten to
`example.net/bar/abc`.

```yaml
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: http-filter-1
spec:
  hostnames:
  - example.com
  rules:
    - matches:
      - path:
          type: Prefix
          value: /foo
      filters:
      - type: URLRewrite
        requestRewrite:
          hostname: example.net
          path:
            type: ReplacePrefixMatch
            substitution: /bar
```

This would be represent with the following API additions:
```go
// HTTPURLRewrite defines a filter that modifies a request during forwarding.
// At most one of these filters may be used on a Route rule. This may not be
// used on the same Route rule as a HTTPRequestRedirect filter.
//
// Support: Extended
type HTTPURLRewrite struct {
  // Hostname is the value to be used to replace the Host header value during
  // forwarding.
  //
  // Support: Extended
  //
  // +optional
  // +kubebuilder:validation:MaxLength=255
  Hostname *string `json:"hostname,omitempty"`

  // Path defines a path rewrite.
  //
  // Support: Extended
  //
  // +optional
  Path *HTTPPathModifier `json:"path,omitempty"`
}
```

Note: `RequestRewrite` was originally considered as a name for this filter.
`URLRewrite` was chosen as it more clearly represented the capabilities of the
filter and would not be confused with header or query param modification.

## Portability

When considering what should be possible in the API, it's worth evaluating what
common tooling is capable of. This is by no means a complete list, but this
provides a high level overview of how this is configured across different
implementations.

Although not all of these implementations directly support prefix rewrites or
redirects, the ones that don't include regular expression support which can be
used to implement prefix rewrites and redirects.

Note: This section intentionally excludes the redirect capabilities already
contained in the API.

### Envoy
Envoy supports the following relevant capabilities
([reference](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/route/v3/route_components.proto)):

* path_redirect (redirect only)
* prefix_rewrite (redirect and forwarding)
* regex_rewrite (redirect and forwarding)
* host_rewrite_literal (forwarding only)
* strip_query (redirect only)

Note that path rewrite relies on the prefix match for the route, there is not
a way to differentiate between the prefix used for matching and rewriting.

### Google Cloud
Google Cloud URL Maps support the following relevant capabilities
([reference](https://cloud.google.com/compute/docs/reference/rest/v1/urlMaps)):

* pathPrefixRewrite (forwarding only)
* hostRewrite (forwarding only)
* pathRedirect (redirect only)
* prefixRedirect (redirect only)
* stripQuery (redirect only)

Note that path rewrite relies on the prefix match for the route, there is not
a way to differentiate between the prefix used for matching and rewriting.

### HAProxy
HAProxy supports the following relevant capabilities
([reference](https://cbonte.github.io/haproxy-dconv/2.5/configuration.html)):

* http-request set-path (advanced path rewrite capabilities)
* http-request replace-path (rewrites entire path)
* http-request replace-pathq (rewrites entire path + query string)
* http-request replace-uri (URI rewrite based on input regex)
* redirect location (advanced redirect capabilities)

### NGINX
The NGINX rewrite module contains the following relevant capabilities
([reference](http://nginx.org/en/docs/http/ngx_http_rewrite_module.html)):

* PCRE regex based rewrites
* Rewrite directive can be used during forwarding or redirects
* Rewrite directive can affect host, path, or both
* Rewrite directive can be chained

## Future Extension
There are two relatively common types of path rewrite/redirect that are not
covered by this proposal:

1. Replace a path prefix separate from the match
2. Replace with a Regular Expression substitution

Both of the following can be represented by adding a new field new types. For
example, this config would result in a request to `/foo/baz` to be rewritten to
`/bar/baz`:

```yaml
filters:
- type: RequestRewrite
  requestRewrite:
    path:
      type: ReplacePrefix
      pattern: /foo
      substitution: /bar
```

Similarly, this config would result in a request to `/foo/bar/baz` being
rewritten to `/foo/other/baz`.
```yaml
filters:
- type: RequestRewrite
  requestRewrite:
    path:
      type: RegularExpression
      pattern: /foo/(.*)/baz
      substitution: other
```

Although both of the above are natural extensions of the API, they are not quite
as broadly supported. For that reason, this GEP proposes omitting these types
from the initial implementation.

## Alternatives

### 1. Generic Path Match Replacement
Instead of the `ReplacePrefixMatch` option proposed above, we could have a
`ReplacePathMatch` option. This would provide significantly more flexibility and
room for growth than prefix replacement.

Unfortunately it would be difficult to represent conformance and support levels.
It also would have limited value. Replacing "Exact" match types would be nearly
identical to the "Absolute" match type, and replacing "RegularExpression" match
types would likely not yield the desired result. In most cases, RegEx rewrites
are implemented separately from RegEx path matching. So a user may want to match
all paths matching one RegEx, but use a separate RegEx + substitution value for
rewrites.

It is theoretically possible that future patch match types could be useful as a
rewrite source, but the common proxies described above seem to be limited to the
rewrite types described above.

### 2. Top Level Rewrite Fields
Although a small difference, we could restructure how the path rewrites and
redirects were configured. One example would be adding top level fields in the
filters for each kind of path rewrite or redirect. That would result in a change
like this:

**Before:**
```yaml
requestRewrite:
  hostname: foo.com
  path:
    type: Prefix
    substitution: /bar
```

**After:**
```yaml
requestRewrite:
  hostname: foo.com
  pathPrefix: /bar
```

Although simpler for the initial use cases, it may become more difficult to
maintain and validate as additional types of rewrites and redirects were added.


## References

Issues:

- [#200: Add support for configurable HTTP redirects](https://github.com/kubernetes-sigs/gateway-api/issues/200)
- [#678: Add support for HTTP rewrites](https://github.com/kubernetes-sigs/gateway-api/issues/678)
