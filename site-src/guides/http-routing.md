# HTTP routing

The [HTTPRoute resource](/api-types/httproute) allows you to match on HTTP traffic and
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
documentation](/api-types/httproute#merging).

In order to receive traffic from a [Gateway][gateway] an `HTTPRoute` resource
must be configured with `ParentRefs` which reference the parent gateway(s) that it
should be attached to. The following example shows how the combination
of `Gateway` and `HTTPRoute` would be configured to serve HTTP traffic:

```yaml
{% include 'standard/http-routing/gateway.yaml' %}
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
{% include 'standard/http-routing/foo-httproute.yaml' %}
```

Similarly, the `bar-route` HTTPRoute matches traffic for `bar.example.com`. All
traffic for this hostname will be evaluated against the routing rules. The most
specific match will take precedence which means that any traffic with the `env:
canary` header will be forwarded to `bar-svc-canary` and if the header is
missing or not `canary` then it'll be forwarded to `bar-svc`.

```yaml
{% include 'standard/http-routing/bar-httproute.yaml' %}
```

[gateway]: /references/spec/#gateway.networking.k8s.io/v1beta1.Gateway
[spec]: /references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteSpec
[svc]:https://kubernetes.io/docs/concepts/services-networking/service/
