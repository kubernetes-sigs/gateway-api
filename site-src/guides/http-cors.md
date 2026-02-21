# Cross-Origin Resource Sharing (CORS)

???+ info "Extended Support Feature: HTTPRouteCORS"
    This feature is part of extended support. For more information on support levels, refer to our [conformance guide](../concepts/conformance.md).

The [HTTPRoute resource](../api-types/httproute.md) can be used to configure
Cross-Origin Resource Sharing (CORS). CORS is a security feature that allows
or denies web applications running at one domain to make requests for resources
from a different domain.

The `CORS` filter in an `HTTPRouteRule` can be used to specify the CORS policy.

## Allowing requests from a specific origin

The following `HTTPRoute` allows requests from `https://app.example`:

```yaml
{% include 'standard/http-cors/httproute-specific-origin-no-creds.yaml' %}
```

Instead of specifying a list of specific origins, you can also specify a
single wildcard (`"*"`), which will allow any origin:

```yaml
{% include 'standard/http-cors/httproute-all-origins-no-creds.yaml' %}
```

It is also allowed to use semi-specified origins in the list,
where the wildcard appears after the scheme
and at the beginning of the hostname, e.g. `https://*.bar.com`:

```yaml
{% include 'standard/http-cors/httproute-origins-with-wildcards-no-creds.yaml' %}
```

## Allowing credentials

The `allowCredentials` field specifies whether the browser is allowed to
include credentials (such as cookies and HTTP authentication) in the CORS
request.

The following rule allows requests from `https://app.example` with
credentials:

```yaml
{% include 'standard/http-cors/httproute-credentials-true.yaml' %}
```

## Other CORS options

The `CORS` filter also allows you to specify other CORS options, such as:

- `allowMethods`: The HTTP methods that are allowed for CORS requests.
- `allowHeaders`: The HTTP headers that are allowed for CORS requests.
- `exposeHeaders`: The HTTP headers that are exposed to the client.
- `maxAge`: The maximum time in seconds that the browser should cache the
preflight response.

For `allowMethods`, `allowHeaders`, and `exposeHeaders`, it is also possible
to use a single wildcard (`"*"`) instead of a list of specific names.

A comprehensive example:

```yaml
{% include 'standard/http-cors/httproute-all-fields-set.yaml' %}
```
