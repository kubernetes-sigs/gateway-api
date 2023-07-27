## What is the Gateway API?

Gateway API is an open source project managed by the
[SIG-NETWORK][sig-network] community. It is an API (collection of resources)
that model service networking in Kubernetes. These resources -
[`GatewayClass`][GatewayClass], [`Gateway`][Gateway],
[`HTTPRoute`][HTTPRoute], [`TCPRoute`][TCPRoute], etc., as well as the
Kubernetes [`Service`][Service] resource - aim to evolve Kubernetes service
networking through expressive, extensible, and role-oriented interfaces that
are implemented by many vendors and have broad industry support.

[GatewayClass]: /api-types/gatewayclass
[Gateway]: /api-types/gateway
[HTTPRoute]: /api-types/httproute
[TCPRoute]: /api-types/tcproute
[Service]: https://kubernetes.io/docs/concepts/services-networking/service/

![Gateway API Model](./images/api-model.png)

The Gateway API was originally designed to manage traffic from clients outside
the cluster to services inside the cluster -- the _ingress_ or
[_north/south_][north/south traffic] case. Over time, interest from service
mesh users prompted the creation of the [GAMMA initiative][gamma] to define
how the Gateway API could also be used for inter-service or [_east/west_
traffic][east/west traffic] within the same cluster.

## Gateway API for Ingress <a name="for-ingress"></a>

When using the Gateway API to manage ingress traffic, the [Gateway] resource
defines a point of access at which traffic from outside clients can be routed
into the cluster ([north/south traffic]). (If you're familiar with the older
[Ingress API], you can think of the Gateway API as analogous to a
more-expressive next-generation version of that API.)

Each Gateway is associated with a [GatewayClass], which describes the actual
kind of [gateway controller] that will handle traffic for the Gateway;
individual routing resources (such as [HTTPRoute]) are then [associated with
the Gateway resources][gateway-attachment]. Separating these different
concerns into distinct resources is a critical part of the role-oriented
nature of the Gateway API, as well as allowing for multiple kinds of gateway
controllers (represented by GatewayClass resources), each with multiple
instances (represented by Gateway resources), in the same cluster.

[Ingress API]:https://kubernetes.io/docs/concepts/services-networking/ingress/
[north/south traffic]:/concepts/glossary#northsouth-traffic
[east/west traffic]:/concepts/glossary#eastwest-traffic
[gateway controller]:/concepts/glossary#gateway-controller
[gateway-attachment]:/concepts/api-overview#attaching-routes-to-gateways

## Gateway API for Service Mesh (the [GAMMA initiative][gamma]) <a name="for-service-mesh"></a>

Things are a bit different when using the Gateway API to manage a [service
mesh][service-mesh]. Since there will usually only be one mesh active in the
cluster, the [Gateway] and [GatewayClass] resources are not used; instead,
individual route resources (such as [HTTPRoute]) are [associated directly with
Service resources][mesh-attachment], permitting the mesh to manage traffic
from any traffic directed to that Service while preserving the role-oriented
nature of the Gateway API.

This use case is still rather new, and should be expected to evolve fairly
quickly. One particular area that has rapidly become critical for GAMMA is the
definition of the different [facets of the Service resource][service-facets].

[gamma]:/contributing/gamma/
[service-mesh]:/concepts/glossary#service-mesh
[service-facets]:/concepts/service-facets
[mesh-attachment]:/concepts/api-overview#attaching-routes-to-services

## Getting started

Whether you are a user interested in using the Gateway API or an implementer
interested in conforming to the API, the following resources will help give
you the necessary background:

- [API overview](/concepts/api-overview)
- [User guides](/guides)
- [Gateway controller implementations](/implementations#gateways)
- [Service Mesh implementations](/implementations#meshes)
- [API reference spec](/references/spec)
- [Community links](/contributing/community) and [developer guide](/contributing/devguide)

## Gateway API concepts
The following design goals drive the concepts of the Gateway API. These
demonstrate how Gateway aims to improve upon current standards like Ingress.

- **Role-oriented** - Gateway is composed of API resources which model
organizational roles that use and configure Kubernetes service networking.
- **Portable** - This isn't an improvement but rather something
that should stay the same. Just as Ingress is a universal specification with
[numerous implementations](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/),
Gateway API is designed to be a portable specification supported by many
implementations.
- **Expressive** - Gateway API resources support core functionality for things
like header-based matching, traffic weighting, and other capabilities that
were only possible in Ingress through custom annotations.
- **Extensible** - Gateway API allows for custom resources to be linked at
various layers of the API. This makes granular customization possible at the
appropriate places within the API structure.

Some other notable capabilities include:

- **GatewayClasses** - GatewayClasses formalize types of load balancing
implementations. These classes make it easy and explicit for users to
understand what kind of capabilities are available via the Kubernetes resource
model.
- **Shared Gateways and cross-Namespace support** - They allow the sharing of
load balancers and VIPs by permitting independent Route resources to attach to
the same Gateway. This allows teams (even across Namespaces) to share
infrastructure safely without direct coordination.
- **Typed Routes and typed backends** - The Gateway API supports typed Route
resources and also different types of backends. This allows the API to be
flexible in supporting various protocols (like HTTP and gRPC) and
various backend targets (like Kubernetes Services, storage buckets, or
functions).
- **Service mesh support** with the GAMMA initiative - The Gateway API
supports associating routing resources with Service resources, to configure
service meshes as well as ingress controllers.

## Why does a role-oriented API matter?

Whether it’s roads, power, data centers, or Kubernetes clusters,
infrastructure is built to be shared. However, shared infrastructure raises a
common challenge - how to provide flexibility to users of the infrastructure
while maintaining control by owners of the infrastructure?

The Gateway API accomplishes this through a role-oriented design for
Kubernetes service networking that strikes a balance between distributed
flexibility and centralized control. It allows shared network infrastructure
(hardware load balancers, cloud networking, cluster-hosted proxies etc) to be
used by many different and non-coordinating teams, all bound by the policies
and constraints set by cluster operators.

The roles used for the Gateway API's design are defined by three personas:

### Personas

- **Ian** (he/him) is an _infrastructure provider_. His role is the care and
  feeding of a set of infrastructure that permits multiple isolated clusters
  to serve multiple tenants. He is not beholden to any single tenant; rather,
  he worries about all of them collectively.

- **Charlie** (they/them) is a _cluster operator_. Their role is to manage a
  single cluster, ensuring that it meets the needs of its several users.
  Again, Charlie is beholden to no single user of their cluster; they need to
  make sure that the cluster serves all of them as needed.

- **Ana** (she/her) is an _application developer_. Ana is in a unique position
  among the Gateway API roles: her focus is on the business needs her
  application is meant to serve, _not_ Kubernetes or the Gateway API. In fact,
  Ana is likely to view the Gateway API and Kubernetes as pure friction
  getting in her way to get things done.

(These three are discussed in more detail in [Roles and
Personas](/concepts/roles-and-personas).)

It should be clear that while Ana, Charlie, and Ian do not necessarily see
eye-to-eye about everything, they need to work together to keep things running
smoothly. This is the core challenge of the Gateway API in a nutshell.

### Use Cases

The [example use cases][use-cases] show this role-oriented model at work. Its
flexibility allows the API to adapt to vastly different organizational models
and implementations while remaining a portable and standard API.

The use cases presented are deliberately cast in terms of the roles presented
above. Ultimately the Gateway API is meant for use by humans, which means that
it must fit the uses to which each of Ana, Charlie, and Ian will put it.

[use-cases]:/concepts/use-cases

## What's the difference between Gateway API and an API Gateway?

An API Gateway is a general concept that describes anything that exposes
capabilities of a backend service, while providing extra capabilities for
traffic routing and manipulation, such as load balancing, request and response
transformation, and sometimes more advanced features like authentication and
authorization, rate limiting, and circuit breaking.

The Gateway API is an interface, or set of resources, that model service
networking in Kubernetes. One of the main resources is a `Gateway`, which
declares the Gateway type (or class) to instantiate and its configuration. As
a Gateway Provider, you can implement the Gateway API to model Kubernetes
service networking in an expressive, extensible, and role-oriented way.

Most Gateway API implementations are API Gateways to some extent, but not all
API Gateways are Gateway API implementations.

## Who is working on Gateway API?

The Gateway API is a
[SIG-Network](https://github.com/kubernetes/community/tree/master/sig-network)
project being built to improve and standardize service networking in
Kubernetes. Current and in-progress implementations include Contour,
Emissary-ingress (Ambassador API Gateway), Google Kubernetes Engine (GKE),
Istio, Kong, Linkerd, and Traefik. Check out the [implementations
reference](implementations.md) to see the latest projects & products that
support Gateway. If you are interested in contributing to or building an
implementation using the Gateway API then don’t hesitate to [get
involved!](/contributing/community)

[sig-network]: https://github.com/kubernetes/community/tree/master/sig-network

