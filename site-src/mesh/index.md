
# Gateway API for Service Mesh

??? success "Standard Channel since v1.1.0"

    The [GAMMA initiative](/mesh/gamma) work for supporting service mesh use
    cases has been part of the Standard Channel since v1.1.0 and is considered
    GA. For more information refer to our [versioning guide](/concepts/versioning).

The "[GAMMA initiative](/mesh/gamma)" refers to the group that is defining how
Gateway API can be used for Service Mesh. To date, this group has been able to
define service mesh support in the Gateway API with relatively small changes.
The most significant change that GAMMA has introduced to date is that, when
configuring a service mesh, individual route resources (such as [HTTPRoute]) are
[associated directly with Service resources](#gateway-api-for-mesh).

This is primarily because there will typically only be one mesh active in the
cluster, so the [Gateway] and [GatewayClass] resources are not used when
working with a mesh. In turn, this leaves the Service resource as the most
universal binding point for routing information.

Since the Service resource is unfortunately complex, with several overloaded
or underspecified aspects, GAMMA has also found it critical to formally define
the [Service _frontend_ and _backend_ facets][service-facets]. In brief:

- The Service frontend is its name and cluster IP, and
- The Service backend is its collection of endpoint IPs.

This distinction helps the Gateway API to be exact about how routing within a
mesh functions, without requiring new resources that largely duplicate the
Service.

[GatewayClass]: /api-types/gatewayclass
[Gateway]: /api-types/gateway
[HTTPRoute]: /api-types/httproute
[TCPRoute]: /concepts/api-overview/#tcproute-and-udproute
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[service-mesh]:/concepts/glossary#service-mesh
[service-facets]:/concepts/service-facets

## Connecting routes and services <a name="gateway-api-for-mesh">

GAMMA specifies that individual Route resources attach directly to a Service,
representing configuration meant to be applied to _any traffic directed to the
Service_.

When one or more Routes are attached to a Service, **requests that do not match
at least one of the Routes will be rejected**. If no Routes are attached to a
Service, requests to the Service simply proceed per the mesh's default behavior
(usually resulting in the request being forwarded as if the mesh were not
present).

Which Routes attach to a given Service is controlled by the Routes themselves
(working with Kubernetes RBAC): the Route simply specifies a `parentRef` that is
a Service, rather than a Gateway.

```yaml
kind: HTTPRoute
metadata:
  name: smiley-route
  namespace: faces
spec:
  parentRefs:
    - name: smiley
      kind: Service
      group: core
      port: 80
  rules:
    ...
```

!!! note "Work in Progress"

    There is ongoing work around the relationship between producer
    routes and consumer routes.

The relationship between the Route's Namespace and the Service's Namespace is
important:

- Same Namespace <a name="producer-routes"></a>

    A Route in the same Namespace as its Service is called a [producer route],
    since it is typically created by the creator of the workload in order to
    define acceptable usage of the workload (for example, [Ana] would deploy
    both the workload and the Route). All requests from any client of the
    workload, from any Namespace, will be affected by this Route.

    The Route shown above is a producer route.

- Different Namespaces <a name="consumer-routes"></a>

    A Route in a different Namespace than its Service is called a [consumer
    route]. Typically, this is a Route meant to refine how a consumer of a given
    workload makes request of that workload (for example, configuring custom
    timeouts for that consumer's use of the workload). This Route will only
    affect requests from workloads in the same Namespace as the Route.

    For example, this HTTPRoute would cause all clients of the `smiley` workload
    in the `fast-clients` Namespace to have a 100ms timeout:

    ```yaml
    kind: HTTPRoute
    metadata:
      name: smiley-route
      namespace: fast-clients
    spec:
      parentRefs:
      - name: smiley
        namespace: faces
        kind: Service
        group: core
        port: 80
      rules:
        ...
        timeouts:
          request: 100ms
    ```

One important note about Routes bound to Services is that multiple Routes for
the same Service in a single Namespace - whether producer routes or consumer
routes - will be combined according to the Gateway API [Route merging rules]. As
such, it is not currently possible to define distinct consumer routes for
multiple consumers in the same Namespace.

For example, if the `blender` workload and the `mixer` workload both live in the
`foodprep` Namespace, and both call the `oven` workload using the same Service,
it is not currently possible for `blender` and `mixer` to use HTTPRoutes to set
different timeouts for their calls to the `oven` workload. `blender` and `mixer`
would need to be moved into separate Namespaces to allow this.

[Ana]:/concepts/roles-and-personas#ana
[producer route]:/concepts/glossary#producer-route
[consumer route]:/concepts/glossary#consumer-route
[service mesh]:/concepts/glossary#service-mesh
[Route merging rules]:/api-types/httproute#merging

## Request Flow

A typical [east/west] API request flow when a GAMMA-compliant mesh is in use
looks like:

1. A client workload makes a request to <http://foo.ns.service.cluster.local>.
2. The mesh data plane intercepts the request and identifies it as traffic for
   the Service `foo` in Namespace `ns`.
3. The data plane locates Routes associated with the `foo` Service, then:

    a. If there are no associated Routes, the request is always allowed, and the
       `foo` workload itself is considered the destination workload.

    b. If there are associated Routes and the request matches at least one of
       them, the `backendRefs` of the highest-priority matching Route are used
       to select the destination workload.

    c. If there are associated Routes, but the request matches none of them, the
       request is rejected.

6. The data plane routes the request on to the destination workload (most likely
   using [endpoint routing], but it is allowed to use [Service routing]).

[east/west]:/concepts/glossary#eastwest-traffic
[endpoint routing]:/concepts/glossary#endpoint-routing
[Service routing]:/concepts/glossary#service-routing
