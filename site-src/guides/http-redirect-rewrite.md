# HTTP path redirects and rewrites

[HTTPRoute resources](../api-types/httproute.md) can issue redirects to
clients or rewrite paths sent upstream using
[filters](../api-types/httproute.md#filters-optional). This guide shows how
to use these features.

Note that redirect and rewrite filters are mutually incompatible. Rules cannot
use both filter types at once.

## Redirects

Redirects return HTTP 3XX responses to a client, instructing it to retrieve a
different resource. [`RequestRedirect` rule
filters](../reference/spec.md#gateway.networking.k8s.io/v1.HTTPRequestRedirectFilter)
instruct Gateways to emit a redirect response to requests matching a filtered
HTTPRoute rule.

### Supported Status Codes

Gateway API supports the following HTTP redirect status codes:

- **301 (Moved Permanently)**: Indicates that the resource has permanently moved to a new location. Search engines and clients will update their references to use the new URL. Use this for permanent redirects like HTTP to HTTPS upgrades or permanent URL changes.

- **302 (Found)**: Indicates that the resource is temporarily available at a different location. This is the default status code if none is specified. Use this for temporary redirects where the original URL may be valid again in the future.

- **303 (See Other)**: Indicates that the response to the request can be found at a different URL using a GET method. This is commonly used after POST requests to redirect to a confirmation page and prevent duplicate form submissions.

- **307 (Temporary Redirect)**: Similar to 302, but guarantees that the HTTP method will not change when following the redirect. Use this when you need to preserve the original HTTP method (POST, PUT, etc.) in the redirect.

- **308 (Permanent Redirect)**: Similar to 301, but guarantees that the HTTP method will not change when following the redirect. Use this for permanent redirects where the HTTP method must be preserved.

Redirect filters can substitute various URL components independently. For
example, to issue a permanent redirect (301) from HTTP to HTTPS, configure
`requestRedirect.statusCode=301` and `requestRedirect.scheme="https"`:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-http.yaml' %}
```

Redirects change configured URL components to match the redirect configuration
while preserving other components from the original request URL. In this
example, the request `GET http://redirect.example/cinnamon` will result in a
301 response with a `location: https://redirect.example/cinnamon` header. The
hostname (`redirect.example`), path (`/cinnamon`), and port (implicit) remain
unchanged.

### Method-Preserving Redirects

When you need to ensure that the HTTP method is preserved during a redirect, use status codes 307 or 308:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-307.yaml' %}
```

For permanent redirects that must preserve the HTTP method, use status code 308:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-308.yaml' %}
```

### POST-Redirect-GET Pattern

For implementing the POST-Redirect-GET pattern, use status code 303 to redirect POST requests to a GET endpoint:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-303.yaml' %}
```

### HTTP-to-HTTPS redirects

To redirect HTTP traffic to HTTPS, you need to have a Gateway with both HTTP
and HTTPS listeners.

```yaml
{% include 'standard/http-redirect-rewrite/gateway-redirect-http-https.yaml' %}
```
There are multiple ways to secure a Gateway. In this example, it is secured
using a Kubernetes Secret(`redirect-example` in the `certificateRefs` section).

You need a HTTPRoute that attaches to the HTTP listener and does the redirect
to HTTPS. Here we set `sectionName` to be `http` so it only selects the
listener named `http`.

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-http.yaml' %}
```

You also need a HTTPRoute that attaches to the HTTPS listener that forwards
HTTPS traffic to application backends.

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-https.yaml' %}
```

### Path redirects

Path redirects use an HTTP Path Modifier to replace either entire paths or path
prefixes. For example, the HTTPRoute below will issue a 302 redirect to all
`redirect.example` requests whose path begins with `/cayenne` to `/paprika`.
Note that you can use any of the supported status codes (301, 302, 303, 307, 308)
depending on your specific requirements:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-full.yaml' %}
```

Both requests to
`https://redirect.example/cayenne/pinch` and
`https://redirect.example/cayenne/teaspoon` will receive a redirect with a
`location: https://redirect.example/paprika`.

The other path redirect type, `ReplacePrefixMatch`, replaces only the path
portion matching `matches.path.value`. Changing the filter in the above to:

```yaml
{% include 'standard/http-redirect-rewrite/httproute-redirect-prefix.yaml' %}
```

will result in redirects with `location:
https://redirect.example/paprika/pinch` and `location:
https://redirect.example/paprika/teaspoon` response headers.

## Rewrites

Rewrites modify components of a client request before proxying it upstream. A
[`URLRewrite`
filter](../reference/spec.md#gateway.networking.k8s.io/v1.HTTPURLRewriteFilter)
can change the upstream request hostname and/or path. For example, the
following HTTPRoute will accept a request for
`https://rewrite.example/cardamom` and send it upstream to `example-svc` with
`host: elsewhere.example` in request headers instead of `host:
rewrite.example`.

```yaml
{% include 'standard/http-redirect-rewrite/httproute-rewrite.yaml' %}
```

Path rewrites also make use of HTTP Path Modifiers. The HTTPRoute below
will take request for `https://rewrite.example/cardamom/smidgen` and proxy a
request to `https://elsewhere.example/fennel` upstream to `example-svc`.

```yaml
{% include 'standard/http-redirect-rewrite/httproute-rewrite-full-path.yaml' %}
```

Instead using `type: ReplacePrefixMatch` and `replacePrefixMatch: /fennel` will
request `https://elsewhere.example/fennel/smidgen` upstream.

```yaml
{% include 'standard/http-redirect-rewrite/httproute-rewrite-prefix-path.yaml' %}
```
