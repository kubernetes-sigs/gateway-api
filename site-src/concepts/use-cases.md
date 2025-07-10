# Use Cases

Gateway API covers a _very_ wide range of use cases (which is both a
strength and a weakness!). This page is emphatically _not_ meant to be an
exhaustive list of these use cases: rather, it is meant to provide some
examples that can be helpful to demonstrate how the API can be used.

In all cases, it's very important to bear in mind the [roles and personas]
used in Gateway API. The use cases presented here are deliberately
described in terms of [Ana], [Chihiro], and [Ian]: they are the ones for whom
the API must be usable. (It's also important to remember that even though
these roles might be filled by the same person, especially in smaller
organizations, they all have distinct concerns that we need to consider
separately.)

[roles and personas]:roles-and-personas.md
[Ana]:roles-and-personas.md#key-roles-and-personas
[Chihiro]:roles-and-personas.md#key-roles-and-personas
[Ian]:roles-and-personas.md#key-roles-and-personas

## The Use Cases

- [Use Cases](#use-cases)
  - [The Use Cases](#the-use-cases)
  - [Basic north/south use case](#basic-northsouth-use-case)
  - [Multiple applications behind a single Gateway](#multiple-applications-behind-a-single-gateway)
  - [Basic east/west use case](#basic-eastwest-use-case)
  - [Gateway and mesh use case](#gateway-and-mesh-use-case)

[role and personas]:roles-and-personas.md

## Basic [north/south] use case

??? success "Standard Channel since v0.8.0"

    The [north/south] use case is fully supported by the Standard Channel
    since `v0.8.0`. For more information on release
    channels, refer to our [versioning guide](versioning.md).

Ana has created a microservice application which she wants to run in
Kubernetes. Her application will be used by clients outside the cluster, and
while Ana has created the application, setting up the cluster is not in her
wheelhouse.

1. Ana goes to Chihiro to ask them to set up a cluster. Ana tells Chihiro that
   her clients will expect her APIs to be available using URLs rooted at
   `https://ana.application.com/`.

2. Chihiro goes to Ian and requests a cluster.

3. Ian provisions a cluster running a gateway controller with a [GatewayClass]
   resource named `basic-gateway-class`. The gateway controller manages the
   infrastructure associated with routing traffic from outside the cluster to
   inside the cluster.

4. Ian gives Chihiro credentials to the new cluster, and tells Chihiro that
   they can use GatewayClass `basic-gateway-class` to set things up.

5. Chihiro applies a [Gateway] named `ana-gateway` to the cluster, telling it
   to listen on port 443 for TLS traffic, and providing it a TLS certificate
   with a Subject CN of `ana.application.com`. They associate this Gateway with the `basic-gateway-class` GatewayClass.

6. The gateway controller that Ian provisioned in step 3 allocates a load
   balancer and an IP address for `ana-gateway`, provisions data-plane
   components that can route requests arriving at the load balancer on port
   443, and starts watching for routing resources associated with
   `ana-gateway`.

7. Chihiro gets the IP address of `ana-gateway` and creates a DNS record
   outside the cluster for `ana.application.com` to match.

8. Chihiro tells Ana that she's good to go, using the Gateway named
   `ana-gateway`.

9. Ana writes and applies [HTTPRoute] resources to configure which URL paths
   are allowed and which microservices should handle them. She associates
   these HTTPRoutes with Gateway `ana-gateway` using the Gateway API [Route
   Attachment Process].

10. At this point, when requests arrive at the load balancer, they are routed
    to Ana's application according to her routing specification.

This allows Chihiro to enforce centralized policies [such as
TLS](../guides/tls.md#downstream-tls) at the Gateway, while simultaneously allowing
Ana and her colleagues control over the application's [routing
logic](../guides/http-routing.md) and rollout plans (e.g. [traffic splitting
rollouts](../guides/traffic-splitting.md)).

[north/south]:glossary.md#northsouth-traffic

## Multiple applications behind a single Gateway

??? success "Standard Channel since v0.8.0"

    The [north/south] use case is fully supported by the Standard Channel
    since `v0.8.0`. For more information on release
    channels, refer to our [versioning guide](versioning.md).

This is remarkably similar to the [basic north/south use
case](#basic-northsouth-use-case), but there are multiple application teams:
Ana and her team are managing a storefront application in the `store`
Namespace, while Allison and her team are managing a website in the `site`
Namespace.

- Ian and Chihiro work together to provide a cluster, `GatewayClass`, and
  `Gateway`, as above.

- Ana and Allison independently deploy workloads and HTTPRoutes bound to the
  same `Gateway` resource.

Again, this separation of concerns allows Chihiro to enforce centralized
policies [such as TLS](../guides/tls.md#downstream-tls) can be enforced at the
Gateway. Meanwhile, Ana and Allison run their applications [in their own
Namespaces](../guides/multiple-ns.md), but attach their Routes to the same shared
Gateway, allowing them to independently control their [routing
logic](../guides/http-routing.md), [traffic splitting
rollout](../guides/traffic-splitting.md), etc., while not worrying about the things
that Chihiro and Ian are handling.

[HTTPRoute]:../api-types/httproute.md
[GatewayClass]:../api-types/gatewayclass.md
[Gateway]:../api-types/gateway.md
[Route Attachment Process]:api-overview.md#attaching-routes-to-gateways

![Gateway API Roles](../images/gateway-roles.png)

## Basic [east/west] use case

In this scenario, Ana has built a workload which is already running in a
cluster with a [GAMMA]-compliant [service mesh]. She wants to use the mesh to
protect her workload by rejecting calls to her workload with incorrect
URL paths, and by enforcing timeouts whenever anyone makes a request of her
workload.

- Chihiro and Ian have already provided a cluster with a running service mesh.
  Ana doesn't need to make any requests of them.

- Ana writes an HTTPRoute that defines acceptable routes and timeouts and has
  a `parentRef` of her workload's Service.

- Ana applies her HTTPRoute in the same Namespace as her workload.

- The mesh automatically starts enforcing the routing policy described by
  Ana's HTTPRoute.

In this case, the separation of concerns across roles allows Ana to take
advantage of the service mesh, with custom routing logic, without any
bottlenecks in requests to Chihiro or Ian.

[east/west]:glossary.md#eastwest-traffic
[GAMMA]:../mesh/index.md
[service mesh]:glossary.md#service-mesh

## Gateway and mesh use case

This is effectively a combination of the [multiple applications behind a
single Gateway](#multiple-applications-behind-a-single-gateway) and [basic
east/west](#basic-eastwest-use-case) use cases:

- Chihiro and Ian will provision a cluster, a [GatewayClass], and a [Gateway].

- Ana and Allison will deploy their applications in the appropriate
  Namespaces.

- Ana and Allison will then apply HTTPRoute resources as appropriate.

There are two very important changes in this scenario, though, since a mesh is
involved:

1. If Chihiro has deployed a [gateway controller] that defaults to [Service
   routing], they will probably need to reconfigure it for [endpoint routing].
   (This is an ongoing area of work for [GAMMA], but the expectation is that
   endpoint routing will be recommended.)

2. Ana and/or Allison will need to bind HTTPRoutes to their respective
   workloads' Services to configure mesh routing logic. These could be
   distinct HTTPRoutes solely for the mesh, or they could apply single
   HTTPRoutes that bind to both the Gateway and a Service.

As always, the ultimate point of separating concerns in this way is that it
permits Chihiro to enforce centralized policies [such as
TLS](../guides/tls.md#downstream-tls) at the Gateway, while allowing Ana and
Allison to retain independent control of [routing
logic](../guides/http-routing.md), [traffic splitting
rollout](../guides/traffic-splitting.md), etc., both for [north/south] and for
[east/west] routing.




[gateway controller]:glossary.md#gateway-controller
[Service routing]:glossary.md#service-routing
[endpoint routing]:glossary.md#endpoint-routing
