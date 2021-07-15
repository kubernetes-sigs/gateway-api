# Cross-Namespace routing

The Gateway API is designed with the ability to route across Namespace boundaries. This is useful when there is more than one user or team sharing the same underlying networking infrastructure. Gateways and Routes can be deployed into different Namespaces and select each other across Namespace boundaries. This allows differing user access and roles (RBAC) to be applied to different Namespaces, effectively controlling which parts of the routing configuration are controlled by which teams. This example shows how two teams, who have no direct coordination, can bind to the same Gateway independently from different Namespaces. The ability for Routes to use Gateways across Namespace boundaries is goverend by [_Route binding_](/guides/multiple-ns#cross-namespace-route-binding).

In this example, there are two independent teams, store and site, operating in the same Kubernetes cluster, in the `store` and `site` Namespaces. These are their requirements:

- The site team has two applications, home and login, that are running behind `foo.example.com`. They have decided to deploy independent HTTPRoutes for each application so that the routing can be configured indendently for each app. By having an independent HTTPRoute for each app, it allows them to update the canary weights for the login app without having to touch any routing configuration for home.
- The store team has a single Service called `store` that they have deployed in the `store` Namespace. 
- The Foobar Corporation operates behind the `foo.example.com` domain so they would like to host all applications on the same Gateway resource. This is controlled by a central infrastructure team, operating in the `infra` Namespace.
- Lastly, the security team controls the certificate for `foo.example.com`. By managing this certificate through the single shared Gateway they are able to centrally control security without directly involving application teams.

The logical relationship between the Gateway API resources looks like this:

![Cross-Namespace routing](/images/cross-namespace-routing.svg)

## Cross-namespace Route binding

[Route binding](/concepts/api-overview/#route-binding) is an important concept that dictates how routes and Gateways select eachother for binding. It is especially relevant when there are multiple Gateways and multiple Namespaces in a cluster. Route binding is based on the principle of bi-directional selection - this means that each resource, the Gateway and the Route, have independent control to determine which resources they permit binding with. This allows Route-owners and Gateway-owners to be independent actors. Route-owners can specify that they will bind with all Gateways in the cluster, or only Gateways from a specific Namespace, with a specific label selector, or an individual Gateway. Similarly, Gateways provide the same level of control. This allows a cluster to be more self-governed, which requires less central administration to ensure that Routes are not over-exposed.

## Resource Deployment

The infrastructure team deploys the `shared-gateway` Gateway into the `infra` Namespace. 

```yaml
{% include 'cross-namespace-routing/gateway.yaml' %}
```

A couple notes about this Gateway:

- It is matching for the `foo.example.com` domain. This is configured on the Gateway so that each HTTPRoute does not also have to configure hostname matching, since they are all using the same domain.
- The Gateway is configured for HTTPS and references the `foo-example-com` Secret. This allows the certificate to be managed centrally for all applications which are using this Gateway.
- `routes.Namespaces.from: All` is the configuration to allow routes from any Namespace to use this Gateway. This impacts how routes are able to bind to this Gateway.

The store team deploys their route for the `store` Service in the `store` Namespace:

```yaml
{% include 'cross-namespace-routing/store-route.yaml' %}
```

This Route has straightforward routng logic as it just matches for `/store` traffic which it sends to the `store` Service. The following snippet of the [`gateways` field](/references/spec/#networking.x-k8s.io/v1alpha1.RouteGateways) controlls which Gateways this Route can bind to:

```yaml
  gateways:
    allow: FromList
    gatewayRefs:
    - name: shared-gateway
      namespace: infra
```

`gateways.allow` can be configured for Gateways in the same Namespace as the Route (the default), all Gateways, or a list of specific Gateways. In this example the store and site teams decide to reference a specific Gateway. This is the least permissive choice which ensures that other Gateways in the cluster (perhaps created in the future at some point) will not bind with these Routes. If cluster administrators have full control over how Gateways are deployed in a cluster then a more permissive binding option could be configured on Routes. The less permissive the Gateway selection is, the less that application owners need to know about which Gateways are deployed. 

The site team now deploys Routes for their applications. They deploy two HTTPRoutes into the `site` Namespace:

- `httproute/home` which acts as a default routing rule, matching for all traffic to `foo.example.com/*` that is not matched by an existing routing rule and routing it to `service/home`
- `httproute/login` which routes traffic for `foo.example.com/login` to `service/login-v1` and `service/login-v2` according to the canary traffic weights 

Both of these Routes use the same Gateway binding configuration which specifies `gateway/shared-gateway` in the `infra` Namespace as the only Gateway that these Routes can bind with.

```yaml
{% include 'cross-namespace-routing/site-route.yaml' %}
```

After these three Routes are deployed, they will all be bound to `gateway/shared-gateway`. The Gateway merges its bound Routes into a single flat list of routing rules. [Routing precedence](/references/spec/#networking.x-k8s.io/v1alpha1.HTTPRouteRule) within an HTTPRoute is determined by most specific match and direct conflicts between Routes would result in an error and prevent Routes from binding to that Gateway.

Thanks to cross-Namespace routing, the Foobar Corporation can distribute ownership of their infrastructure more evenly, while still retaining centralized control. This gives them the best of both worlds, all delivered through declarative and open source APIs.