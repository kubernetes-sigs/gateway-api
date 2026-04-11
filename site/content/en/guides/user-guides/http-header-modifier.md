---
title: "HTTP Header Modifiers"
linkTitle: "HTTP header modifier"
weight: 4
---

[HTTPRoute resources](/reference/api-types/httproute/) can modify the headers of HTTP requests and the HTTP responses from clients.
There are two types of [filters](/reference/api-types/httproute/#filters-optional) available to meet these requirements: `RequestHeaderModifier` and `ResponseHeaderModifier`.

This guide shows how to use these features.

Note that these features are compatible. HTTP headers of the incoming requests and the headers of their responses can both be modified using a single [HTTPRoute resource](/reference/api-types/httproute/).

## HTTP Request Header Modifier

{{< details title="Extended Support Feature: HTTPRouteBackendRequestHeaderModification" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](/docs/concepts/conformance/).

{{< /details >}}
HTTP header modification is the process of adding, removing, or modifying HTTP headers in incoming requests.

To configure HTTP header modification, define a Gateway object with one or more HTTP filters. Each filter specifies a specific modification to make to incoming requests, such as adding a custom header or modifying an existing header.

To add a header to an HTTP request, use a filter of the type `RequestHeaderModifier`, with the `add` action and the name and value of the header:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: header-http-echo
spec:
  parentRefs:
    - name: acme-gw
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: /add-a-request-header
      filters:
        - type: RequestHeaderModifier
          requestHeaderModifier:
            add:
              - name: my-header-name
                value: my-header-value
      backendRefs:
        - name: echo
          port: 8080
```

To edit an existing header, use the `set` action and specify the value of the header to be modified and the new header value to be set.

```yaml
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set:
          - name: my-header-name
            value: my-new-header-value
```

Headers can also be removed, by using the `remove` keyword and a list of header names.

```yaml
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        remove: ["x-request-id"]
```

Using the example above would remove the `x-request-id` header from the HTTP request.

### HTTP Response Header Modifier

{{< details title="Extended Support Feature: HTTPRouteResponseHeaderModification" >}}
This feature is part of extended support. For more information on support levels, refer to our [conformance guide](/docs/concepts/conformance/).

{{< /details >}}
Just like editing request headers can be useful, the same goes for response headers. For example, it allows teams to add/remove cookies for only a certain backend, which can help in identifying certain users that were redirected to that backend previously.

Another potential use case could be when you have a frontend that needs to know whether it’s talking to a stable or a beta version of the backend server, in order to render different UI or adapt its response parsing accordingly.

Modifying the HTTP header response leverages a very similar syntax to the one used to modify the original request, albeit with a different filter (`ResponseHeaderModifier`).

Headers can be added, edited and removed. Multiple headers can be added, as shown in this example below:

```yaml
    filters:
    - type: ResponseHeaderModifier
      responseHeaderModifier:
        add:
        - name: X-Header-Add-1
          value: header-add-1
        - name: X-Header-Add-2
          value: header-add-2
        - name: X-Header-Add-3
          value: header-add-3
```
