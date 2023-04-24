# HTTP Header Modifiers

[HTTPRoute resources](/api-types/httproute) can issue modify the headers of HTTP requests, as well as their responses from clients. 
There are two types of [filters](/api-types/httproute#filters-optional): `RequestHeaderModifier` and `ResponseHeaderModifier`.

This guide shows how to use these features.

Note that these features are compatible. HTTP headers of the incoming requests and the headers of their responses can both be modified using a single [HTTPRoute resource](/api-types/httproute).

## HTTP Request Header Modifier

HTTP header modification is the process of adding, removing, or modifying HTTP headers in incoming requests. 

To configure HTTP header modification, you define a Gateway object with one or more HTTP filters. Each filter specifies a specific modification to make to incoming requests, such as adding a custom header or modifying an existing header.

To add a header, use a filter of the type `RequestHeaderModifier`, with the `add` action and the name and value of the header:

```yaml
{% include 'standard/http-request-header-add.yaml' %}
```

The HTTP header will be updated accordingly. For example, when making a `curl` request processed by the Gateway with the filter configuration above, the request headers would:

```
Request Headers:
        accept=*/*  
        my-header-name=my-header-value  
        user-agent=curl/7.81.0  
        x-forwarded-proto=http  
        x-request-id=5ea1a402-f847-4fd5-b165-5a4324a2ffaa
```

To edit an existing header, use the `set` action and 

```
    filters:
    - type: RequestHeaderModifier
      requestHeaderModifier:
        set: 
          - name: my-header-name
            value: my-new-header-value
```

Headers can also be removed, by using the `remove` keyword and an array of headers to remove. 

```yaml
{% include 'standard/http-request-header-remove.yaml' %}
```

Using the example above would remove the `x-request-id` header from the HTTP request:

```
Request Headers:
        accept=*/*  
        host=172.18.255.200  
        user-agent=curl/7.81.0  
        x-forwarded-proto=http 
```

Finally, setting the header value would require the operator to use the `set` action and the name of the header to be modified, alongside its new value: 


```yaml
{% include 'standard/http-request-header-set.yaml' %}
```

### HTTP Response Header Modifier

!!! info "Experimental Channel"

    The `Path` field described below is currently only included in the
    "Experimental" channel of Gateway API. Starting in v0.7.0, this
    feature will graduate to the "Standard" channel.

Just like editing request headers can be useful, the same goes for response headers. For example, it allows teams to add/remove cookies for only a certain backend, which can help in identifying certain users that were redirected to that backend previously.

Another potential use case could be when you have a frontend that needs to know whether it’s talking to a stable or a beta version of the backend server, in order to render different UI or adapt its response parsing accordingly.

Modifying the HTTP header response leverages a very similar syntax to the one used to modify the original request, albeit with a different filter (`ResponseHeaderModifier`).

Headers can be added, edited and removed. Multiple headers can be added, as shown in this example below:

```yaml
{% include 'experimental/http-response-header.yaml' %}
```
