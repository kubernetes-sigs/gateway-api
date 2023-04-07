## What is the Gateway API?

Gateway API is an open source project managed by the [SIG-NETWORK][sig-network]
community. It is a collection of resources that model service networking 
in Kubernetes. These resources - `GatewayClass`, `Gateway`, `HTTPRoute`, 
`TCPRoute`, `Service`, etc - aim to evolve Kubernetes service networking through 
expressive, extensible, and role-oriented interfaces that are implemented by 
many vendors and have broad industry support. 

![Gateway API Model](./images/api-model.png)

## Getting started

Whether you are a user interested in using the Gateway API or an implementer 
interested in conforming to the API, the following resources will help give 
you the necessary background:

- [API overview](/concepts/api-overview)
- [User guides](/guides)
- [Gateway controller implementations](/implementations)
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
and constraints set by cluster operators. The following example shows how this
works in practice.

A cluster operator creates a [Gateway](/api-types/gateway) resource derived from a
[GatewayClass](/api-types/gatewayclass). This Gateway deploys or configures the
underlying network resources that it represents. Through the
[Route Attachment Process](/concepts/api-overview#attaching-routes-to-gateways)
between the Gateway and Routes, the cluster operator and specific teams must
agree on what can attach to this Gateway and expose their applications through
it. Centralized policies [such as TLS](/guides/tls#downstream-tls) can
be enforced on the Gateway by the cluster operator. Meanwhile, the store and site
teams run [in their own Namespaces](/guides/multiple-ns), but attach their
Routes to the same shared Gateway, allowing them to independently control
their [routing logic](/guides/http-routing). This separation of concerns
allows the store team to manage their own
[traffic splitting rollout](/guides/traffic-splitting) while
leaving centralized policies and control to the cluster operators.

![Gateway API Roles](./images/gateway-roles.png)

This flexibility allows the API to adapt to vastly different
organizational models and implementations while remaining a portable and
standard API.


## Who is working on Gateway?

The Gateway API is a
[SIG-Network](https://github.com/kubernetes/community/tree/master/sig-network)
project being built to improve and standardize service networking in
Kubernetes. Current and in-progress implementations include Contour,
Emissary-Ingress (Ambassador API Gateway), Google Kubernetes Engine (GKE), Istio,
Kong, and Traefik. Check out the [implementations
reference](implementations.md) to see the latest projects &
products that support Gateway. If you are interested in contributing to or
building an implementation using the Gateway API then don’t hesitate to [get
involved!](/contributing/community)

[sig-network]: https://github.com/kubernetes/community/tree/master/sig-network

