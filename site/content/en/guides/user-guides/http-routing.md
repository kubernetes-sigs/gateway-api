---
title: "HTTP routing"
weight: 2
---

The [HTTPRoute resource](/reference/api-types/httproute/) allows you to match on HTTP traffic and
direct it to Kubernetes backends. This guide shows how the HTTPRoute matches
traffic on host, header, and path fields and forwards it to different
Kubernetes Services.

The following diagram describes a required traffic flow across three different
Services:

- Traffic to `foo.example.com/login` is forwarded to `foo-svc`
- Traffic to `bar.example.com/*` with a `env: canary` header is forwarded
to `bar-svc-canary`
- Traffic to `bar.example.com/*` without the header is forwarded to `bar-svc`

![HTTP Routing](/images/http-routing.png)

The dotted lines show the Gateway resources deployed to configure this routing
behavior. There are two HTTPRoute resources that create routing rules on the
same `prod-web` Gateway. This illustrates how more than one Route can bind to a
Gateway which allows Routes to merge on a Gateway as long as they don't
conflict. For more information on Route merging, refer to the [HTTPRoute
documentation](/reference/api-types/httproute/#merging).

In order to receive traffic from a [Gateway][gateway] an `HTTPRoute` resource
must be configured with `ParentRefs` which reference the parent gateway(s) that it
should be attached to. The following example shows how the combination
of `Gateway` and `HTTPRoute` would be configured to serve HTTP traffic:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: example-gateway
spec:
  gatewayClassName: example-gateway-class
  listeners:
  - name: http
    protocol: HTTP
    port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: example-route
spec:
  parentRefs:
  - name: example-gateway
  hostnames:
  - "example.com"
  rules:
  - backendRefs:
    - name: example-svc
      port: 80
```

An HTTPRoute can match against a [single set of hostnames][spec].
These hostnames are matched before any other matching within the HTTPRoute takes
place. Since `foo.example.com` and `bar.example.com` are separate hosts with
different routing requirements, each is deployed as its own HTTPRoute -
`foo-route` and `bar-route`.

The following `foo-route` will match any traffic for `foo.example.com` and apply
its routing rules to forward the traffic to the correct backend. Since there is
only one match specified, only `foo.example.com/login/*` traffic will be
forwarded. Traffic to any other paths that do not begin with `/login` will not
be matched by this Route.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: foo-route
spec:
  parentRefs:
  - name: example-gateway
  hostnames:
  - "foo.example.com"
  rules:
  - matches:
    - path:
        type: PathPrefix
        value: /login
    backendRefs:
    - name: foo-svc
      port: 8080

```

Similarly, the `bar-route` HTTPRoute matches traffic for `bar.example.com`. All
traffic for this hostname will be evaluated against the routing rules. The most
specific match will take precedence which means that any traffic with the `env:
canary` header will be forwarded to `bar-svc-canary` and if the header is
missing or not `canary` then it'll be forwarded to `bar-svc`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: bar-route
spec:
  parentRefs:
  - name: example-gateway
  hostnames:
  - "bar.example.com"
  rules:
  - matches:
    - headers:
      - type: Exact
        name: env
        value: canary
    backendRefs:
    - name: bar-svc-canary
      port: 8080
  - backendRefs:
    - name: bar-svc
      port: 8080
```

[gateway]: /reference/api-spec/main/spec/#gateway
[spec]: /reference/api-spec/main/spec/#httproutespec
[svc]: https://kubernetes.io/docs/concepts/services-networking/service/
