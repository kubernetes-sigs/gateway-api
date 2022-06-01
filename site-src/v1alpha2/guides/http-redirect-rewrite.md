# HTTP path redirects and rewrites

[HTTPRoute resources](/v1alpha2/api-types/httproute) can issue redirects to
clients or rewrite paths sent upstream using
[filters](/v1alpha2/api-types/httproute#filters-optional). This guide shows how
to use these features.

Note that redirect and rewrite filters are mutually incompatible. Rules cannot
use both filter types at once.

## Redirects

Redirects return HTTP 3XX responses to a client, instructing it to retrive a
different resource. [`RequestRedirect` rule
filters](v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.HTTPRequestRedirectFilter)
instruct Gateways to emit a redirect response to requests matching a filtered
HTTPRoute rule.

Redirect filters can substitute various URL components independently. For
example, to issue a permanent redirect (301) from HTTP to HTTPS, configure
`requestRedirect.statusCode=301` and `requestRedirect.scheme="https"`:

```yaml
{% include 'v1alpha2/http-redirect-rewrite/httproute-redirect-https.yaml' %}
```

Redirects change configured URL components to match the redirect configuration
while preserving other components from the original request URL. In this
example, the request `GET http://redirect.example/cinammon` will result in a
301 response with a `location: https://redirect.example/cinnamon` header. The
hostname (`redirect.example`), path (`/cinnamon`), and port (implicit) remain
unchanged.

<!---
The above raises some questions. Do redirects apply unconditionally? If my
request is already HTTPS, will the above create a redirect loop? If redirects
are unconditional (i.e. they _do not_ nothing and proxy upstream if the request
already uses HTTPS), how should you properly write configuration elsewhere to
ensure this only applies to HTTP requests?
-->

### Path redirects

Path redirects use an HTTP Path Modifier to replace either entire paths or path
prefixes. For example, the HTTPRoute below will issue a 302 redirect to all
`redirect.example` requests whose path begins with `/cayenne` to `/paprika`:

```yaml
{% include 'v1alpha2/http-redirect-rewrite/httproute-redirect-full.yaml' %}
```

Both requests to
`https://redirect.example/cayenne/pinch` and
`https://redirect.example/cayenne/teaspoon` will receive a redirect with a
`location: https://redirect.example/paprika`.

The other path redirect type, `ReplacePrefixMatch`, replaces only the path
portion matching `matches.path.value`. Changing the filter in the above to:

```yaml
{% include 'v1alpha2/http-redirect-rewrite/httproute-redirect-prefix.yaml' %}
```

will result in redirects with `location:
https://redirect.example/paprika/pinch` and `location:
https://redirect.example/paprika/teaspoon` response headers.

<!---
Do these behave at all differently based on matches.path.type? I expect no, and
that implementations should just always replace only the matching path prefix
or the entire path. Using ReplacePrefixMatch on Exact is the same as
ReplaceFullPath. RegularExpression looks a bit ambiguous, since the
PathMatchType looks unclear on whether they're handled as prefixes or not.
-->

## Rewrites

Rewrites modify components of a client request before proxying it upstream. A
[`URLRewrite`
filter](/v1alpha2/references/spec/#gateway.networking.k8s.io/v1alpha2.HTTPURLRewriteFilter)
can change the upstream request hostname and/or path. For example, the
following HTTPRoute will accept a request for
`https://rewrite.example/cardamom` and send it upstream to `example-svc` with
`host: elsewhere.example` in request headers instead of `host:
rewrite.example`.

```yaml
{% include 'v1alpha2/http-redirect-rewrite/httproute-rewrite.yaml' %}
```

Path rewrites also make use of HTTP Path Modifiers. The HTTPRoute below
will take request for `https://rewrite.example/cardamom/smidgen` and proxy a
request to `https://elsewhere.example/fennel` upstream to `example-svc`.
Instead using `type: ReplacePrefixMatch` and `replacePrefixMatch: /fennel` will
request `https://elsewhere.example/fennel/smidgen` upstream.

```yaml
{% include 'v1alpha2/http-redirect-rewrite/httproute-rewrite.yaml' %}
```
