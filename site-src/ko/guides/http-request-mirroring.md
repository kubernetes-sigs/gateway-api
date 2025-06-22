# HTTP request mirroring

??? example "Extended Support Feature"

    As of v1.0.0, the Request Mirroring feature is an Extended feature, and
    requires implementations to support the `HTTPRouteRequestMirror` feature.

The [HTTPRoute resource](../api-types/httproute.md) allows you to mirror HTTP
requests to another backend using
[filters](../api-types/httproute.md#filters-optional). This guide shows how to use
this feature.

Mirrored requests will must only be sent to one single destination endpoint
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
