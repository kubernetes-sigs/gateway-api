# HTTP request mirroring

???+ info "Extended Support Feature: HTTPRouteRequestMirror"
    This feature is part of extended support. For more information on release channels, refer to our [versioning guide](../concepts/versioning.md).

The [HTTPRoute resource](../api-types/httproute.md) can be used to mirror
requests to multiple backends. This is useful for testing new services with
production traffic.

Mirrored requests will only be sent to one single destination endpoint
within this backendRef, and responses from this backend MUST be ignored by
the Gateway.

Request mirroring is particularly useful in blue-green deployment. It can be
used to assess the impact on application performance without impacting
responses to clients in any way.

```yaml
{% include 'standard/http-request-mirroring/httproute-mirroring.yaml' %}
```

In this example, all requests are forwarded to service `foo-v1` on port `8080`,
and they are also forwarded to service `foo-v2` on port `8080`, but responses
are only generated from service `foo-v1`.
