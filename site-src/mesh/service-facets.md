# The Different Facets of a Service

The Kubernetes [Service] resource is considerably more complex than people
often realize. When you create a Service, typically the cluster machinery will:

- Allocate a cluster-wide IP address for the Service itself (its _cluster IP_);
- Allocate a DNS name for the Service, resolving to the cluster IP (its _DNS name_);
- Collect the separate cluster-wide IP addresses assigned to each Pod matched
  by the Service's selector (the _endpoint IPs_) into the Service's Endpoints
  or EndpointSlices.
- Configure the network such that traffic to the cluster IP will be
  load-balanced across all the endpoint IPs.

Unfortunately, these implementation details become very important when
considering how Gateway API can work for service meshes!

In [GAMMA initiative][gamma] work, it has become useful to consider Services
as comprising two separate _facets_:

- The **frontend** of the Service is the combination of the cluster IP and
  its DNS name.

- The **backend** of the Service is the collection of endpoint IPs. (The Pods
  are not part of the Service backend, but they are of course strongly
  associated with the endpoint IPs.)

The distinction between the facets is critical because the
[gateway](/api-types/gateway/) and the [mesh](/mesh) each need to decide whether
a request that mentions a given Service should be directed to the Service's
frontend or its backend:

- Directing the request to the Service's frontend (_Service routing_) leaves
  the decision of which endpoint IP to use to the underlying network
  infrastructure (which might be `kube-proxy`, a service mesh, or something
  else).

- Directing the request to the Service's backend (_endpoint routing_) is
  often necessary to enable more advanced load balancing decisions (for
  example, a gateway implementing sticky sessions).

While Service routing may be the most direct fit for [Ana]'s sense of routing,
endpoint routing can be more predictable when using Gateway API for both
[north/south] and [east/west traffic]. The [GAMMA initiative][gamma] is working to
formalize guidance for this use case.

[Service]: https://kubernetes.io/docs/concepts/services-networking/service/
[north/south]:/concepts/glossary#northsouth-traffic
[east/west traffic]:/concepts/glossary#eastwest-traffic
[gamma]:/concepts/gamma/
[Ana]:/concepts/roles-and-personas#ana
