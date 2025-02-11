# GEP-1282: Describing Backend Properties

* Issue: [#1282](https://github.com/kubernetes-sigs/gateway-api/issues/1282)
* Status: Declined


## Explanation for Declined status

This GEP is declined because, after spending a lot of time discussing it, we felt that it was too large, and had too many crosscutting concerns.

It's superseded for TLS Backend Properties by [GEP-1897](https://github.com/kubernetes-sigs/gateway-api/issues/1897).

## TLDR

End-user feedback of Gateway API has presented a trend: some types of configuration requested by users are more about defining functionality that describes _capabilities of the backend_ more than the _route you take to get to the backend_.

There is currently nowhere to put these capabilities in the API, so we need to find a place to put the information required to meet these use cases, with the simplest examples being backend TLS or websocket details.

The purpose of this GEP is to add APIs for capturing those backend capabilities to enable more complex routing composition.

## Goals

* Provide a method to store properties and capabilities for service backends as extended-only features
* The method must be structured (that is, stored in a typed schema), and extensible (the design must include ways in which we can add more properties and capabilities in the future without redesigning it).
* The method should have a roadmap where we can take what we learn from this design and bring it to the core Kubernetes resources later.
* Clear separation of concerns with PolicyAttachment for properties which require multiple levels of defaults/overrides (timeouts, retries, etc).


## Non-Goals

* We don’t want to require changes to v1.Service to start with, as the turnaround time for those changes can be years.
* Our solution should not use unstructured data like labels and annotations - these are very sticky and hard to get rid of once you start using them.
* v1.Service’s appProtocol field is not fit for purpose, because it is defined as accepting values either from the [IANA Service Name registry](https://www.iana.org/assignments/service-names-port-numbers/service-names-port-numbers.xhtmly), or domain-prefixed values and we need more flexibility than that.
* To be populated as we discuss more.


## Introduction

As implementations have started to build out HTTPRoute support, requests for some common capabilities have started coming up, particularly TLS re-encryption (encryption between Gateway and backend), Websockets tracking, and HTTP/2 support.

Evan Anderson opened a [discussion that generated a lot of interest](https://github.com/kubernetes-sigs/gateway-api/discussions/1244), and during that discussion, we brought up a few more points:
* Whatever we do to solve this in the HTTPRoute case may be applicable describing more general service properties, like identity, which could be useful for other Route use cases.
* This may be very useful for mesh/GAMMA use cases as well as more generally for adding arbitrary future capabilities to the backend service. An example is a CA cert for connecting to the backend - that’s tightly bound to the service, but there’s nowhere to put it at the moment.

It's also worth noting that the GAMMA discussion has made clear the distinction between the Kubernetes Service as a frontend (an identity and place to attach routing information to), and as a backend (as a grouping of endpoints). This GEP is concerned with backend capabilities, so it's concerned with the latter.

This initial form of this GEP is for the Gateway API community to agree on what problem we’re solving and why.

### Why build something?

We've got the following feature requests and discussions in the Gateway API repo:
- [#1244](https://github.com/kubernetes-sigs/gateway-api/discussions/1244) : Unclear how to specify upstream (webserver) HTTP protocol. This issue describes the problems that Evan had in trying to be able to define if a backend behind a HTTPRoute supports HTTP2 over cleartext or websockets. The question of how to tell the Gateway implementation that the backend needs TLS details for a proxy-based implementation to be able to connect it also came up in the discussion.
- [#1285](https://github.com/kubernetes-sigs/gateway-api/discussions/1285) has a more specific discussion about how different ingress implementations allow this to be configured today, whether that's with the Ingress resource or their own custom one. The great roundup that Candace did is reproduced in the next few bullet points.
  * Istio uses a [DestinationRule resource with ClientTLSSettings](https://istio.io/latest/docs/reference/config/networking/destination-rule/#ClientTLSSettings) to capture TLS details, and the DestinationRule resource also holds traffic policy information like load balancing algorithm, connection pool size, and so on.
  * Openshift’s Route resource allows the [configuration of reencryption](https://docs.openshift.com/container-platform/4.10/networking/routes/secured-routes.html#nw-ingress-creating-a-reencrypt-route-with-a-custom-certificate_secured-routes) specifically, along with custom certificate details.
  * Contour’s HTTPProxy captures TLS details using an Envoy client certificate, destination CA certificate, and optional SubjectName which sets what Envoy should expect to see from the backend service, all inside the HTTPProxy resource. It also requires either a Protocol field inside the HTTPProxy, or an annotation on the Service that tells Contour that the service expects TLS. This is [all documented](https://projectcontour.io/docs/v1.21.1/config/upstream-tls/), but I should note that Contour’s docs use the Envoy convention where a backend in Gateway parlance is called an Upstream (which may be confusing if you’re not used to it).
  * Linkerd uses a [Server resource](https://linkerd.io/2.11/reference/authorization-policy/#server) (which is functionally pretty similar to Service in that it associates a name with a Pod selector, but also has other details like if the service supports Proxy protocol), along with a ServerAuthorization resource that specifies some constructs that sit more at the service mesh level, including identity and access control.

In terms of other implementations existing use cases for features like this:
- For Contour, there are annotations to allow the configuration of the following on a backend Service:
  - Max connections
  - Max pending requests
  - Max requests
  - Max retries
  - Upstream protocol: This is what allows Contour to handle switching protocols for the backend service.
- In Istio, the `DestinationRule` resource allows the configuration of many settings like this:
  - Load balancer algorithms
  - Connection pool settings
  - Outlier Detection
  - Tunnel settings for non-HTTP workloads
- In Consul, [`ServiceDefaults`](https://www.consul.io/docs/connect/config-entries/service-defaults) allows specifying similar configuration:
  - Max connections
  - Max pending requests
  - Max concurrent requests
  - Protocol
  - Passive health check interval and max failures (outlier detection)
  - TLS SNI (we don't support configuring a specific CA at this granularity, but publishing a public key in a status field could be useful)
  
The properties we're talking about all share two things:
- They are tightly bound to the backend, rather than being important at the route level
- And the service owner (application developer in the Gateway API personas) should control these settings, not the owner of the Gateway implementation (the Cluster Operator)

A good example is the load balancing algorithms for backends, as distinct from load balancing for routes. Load balancing for routes is about choosing which route option the implementation should choose, while load balancing for backends is about which endpoint from the backend the implementation should send traffic to if that backend's route is chosen. To put this in GAMMA initiative terms, load balancing for Routes is a frontend concern that's addressed with config in the HTTPRoute (or other Route resource), while load balancing for backends is a backend concern.

It's likely that these will need to be set differently (for example, random load balancing at the route level for canarying between services, with Weighted Least Request at the backend level for the actual endpoints). This was discussed quite extensively in the [PR for GEP-992](https://github.com/kubernetes-sigs/gateway-api/pull/993), which ended up not going ahead. See [GEP-992](https://github.com/kubernetes-sigs/gateway-api/issues/992) for more details.

The existing resources we have (Service and HTTPRoute) don’t have space for this information currently, so we need to do *something*.

### What we're building

We’re looking to add specific, structured extension points somewhere in the resources used by the Gateway API to describe properties and capabilities of backend services.

Those specific, structured extension points need to be in a place where they can be owned by the person who owns the backend, since that could be different to the person who owns the Route. That is, whatever we choose should be something that extends a thing that is referred to by a `backendRef` in a Route, not inside the Route.

The initial list includes, but is not limited to:
* TLS information for connection from the Gateway to the backend service. Note that this doesn’t include any information used for service mesh encryption, just what a Gateway’s proxy would need to be able to connect to the backend service.
* Websocket protocol information for the backend service.
* Protocol disambiguation for “upgradeable” protocols like HTTP/2 and HTTP/1.1 which operate on the same port. (Websockets may be another case of this.)

Of course, from the section above, you can see that there are many other features and capabilities that can sit in at this level, which is why we want to ensure that this is a more general design.

## API

To be written later, once we have agreement on the “what” and the “why”. (This is the “how”).


## Alternatives

Again, we need to wait until we’re writing the “how”.

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
