# Routes in multiple namespaces

With the Service APIs, a single Gateway can target routes across multiple
namespaces. 

This guide assumes that you have installed Service APIs CRDs and a conformant
controller.

In the following example:

- `acme-lb` GatewayClass: The GatewayClass responsible for satisfying Gateway
  and Route resources.
- `multi-ns-gateaway` Gateway: The Gateway is configured with a single listener
  on port 80 which selects routes that have the label `product: baz` in any
  namespace.  Notice how the `routes.namespaces.from` field in the listener is
  set to `All`.
- `service-apis-example-ns1` and `service-apis-example-ns2` Namespaces: These
  are the namespaces in which route resources are instantiated.
- `http-app-1` and `http-app-2` HTTPRoutes: These are two resources that are
  installed in separate namespaces. These routes will be bound to Gateway
  `multi-ns-gateway` for the following reasons:
    - Both have the `product: baz` label on them.
    - `http-app-1` HTTPRoute has `spec.gateways.allow` set to `All`.  The route
      owner has opted to allow **all** Gateways in the cluster to bind to this
      Route.
    - `http-app-2` HTTPRoute has `spec.gateways.allow` set to `FromList` and
      contains a reference to the `multi-ns-gateway` in `default` namespace.
      This means that only the specified Gateway resource can bind to this
      route.  Additional Gateways may be added to this list to allow them to
      bind to this route.

```
{% include 'routes-in-multiple-namespaces.yaml' %}
```

Please note that this guide illustrates this feature for HTTPRoute resource
only as an example. The same can be accomplished with other route types as
well.
