# Introduction

Gateway API is an open source project managed by the [SIG-NETWORK][sig-network]
community. The project's goal is to evolve service networking APIs within the
Kubernetes ecosystem. Gateway API provide interfaces to expose Kubernetes
applications - Services, Ingress, and more.

### What is the goal of Gateway API?

Gateway API aims to improve service networking by providing expressive,
extensible, role-oriented interfaces that are implemented by many vendors and
have broad industry support.

Gateway API is a collection of API resources - `Service`, `GatewayClass`,
`Gateway`, `HTTPRoute`, `TCPRoute`, etc. Together these resources model a wide
variety of networking use-cases.

![Gateway API Model](./images/api-model.png)


How do Gateway API improve upon current standards like Ingress?

- **More expressive** - They express more core functionality for things like
header-based matching, traffic weighting, and other capabilities that were only
possible in Ingress through custom means.
- **More extensible** - They allow for custom resources to be linked at various
layers of the API. This allows for more granular customization at the
appropriate places within the API structure.
- **Role oriented** - They are separated into different API resources that map
to the common roles for running applications on Kubernetes.
- **Generic** - This isn't an improvement but rather something
that should stay the same. Just as Ingress is a universal specification with
[numerous implementations](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/),
Gateway API are designed to be a portable specification supported by many
implementations.

Some other notable capabilities include …

- **Shared Gateways** - They allow the sharing of load balancers and VIPs by
permitting independent Route resources to bind to the same Gateway. This allows
teams to share infrastructure safely without requiring direct coordination.
- **Typed backend references** - With typed backend references Routes can
reference Kubernetes Services, but also any kind of Kubernetes resource that is
designed to be a Gateway backend.
- **Cross-Namespace references** - Routes across different Namespaces can bind
to Gateways. This allows for shared networking infrastructure despite Namespace
segmentation for workloads.
- **Classes** - GatewayClasses formalize types of load balancing implementations.
These classes make it easy and explicit for users to understand what kind of
capabilities are available as a resource model itself.

[sig-network]: https://github.com/kubernetes/community/tree/master/sig-network

### Where to get started

To get started, please read through [API overview](api-overview.md). These
documents give the necessary background to understand the API and the use-cases
it targets. Once you have a good understanding of the API at a higher-level,
please follow one of our [guides](guides.md) to dive deeper into different parts
of the API.

For a complete API reference, please refer to:

- [API reference](spec.md)
- [Go docs for the package](https://pkg.go.dev/sigs.k8s.io/gateway-api/apis/v1alpha1)

### How to get involved

This project has many contributors, and we welcome anybody and everybody to get
involved. Join our weekly meetings, file issues, or ask questions in Slack. No
contribution is too small - even opinions matter!

- [Weekly meeting schedule](community.md#meetings)
- [Gateway API Slack](https://kubernetes.slack.com/messages/sig-network-gateway-api)
- [Enhancement requests](enhancement-requests.md)
- [Project owners](https://raw.githubusercontent.com/kubernetes-sigs/gateway-api/master/OWNERS)
