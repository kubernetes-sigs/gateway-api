# Implementations

This document tracks downstream implementations of Gateway API and provides status and resource references for them.

Implementors of Gateway API are encouraged to update this document with status information about their implementations, the versions they cover, and documentation to help users get started.

## Implementation Status

- [Apache APISIX][1] (work in progress)
- [Contour][2] (alpha)
- [Emissary-Ingress (Ambassador API Gateway)][3] (alpha)
- [Gloo Edge 2.0][4] (work in progress)
- [Google Kubernetes Engine][5] (public preview)
- [HAProxy Ingress][6] (alpha)
- [Istio][7] (alpha)
- [Kong][8] (work in progress)
- [Traefik][9] (alpha)

[1]:#apisix
[2]:#contour
[3]:#emissary-ingress-ambassador-api-gateway
[4]:#gloo-edge
[5]:#google-kubernetes-engine
[6]:#haproxy-ingress
[7]:#istio
[8]:#kong
[9]:#traefik

## Project References

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### APISIX

[Apache APISIX][apisix] is a dynamic, real-time, high-performance API Gateway. APISIX provides rich traffic management features such as load balancing, dynamic upstream, canary release, circuit breaking, authentication, observability, and more.

APISIX is moving forward to implement the Gateway API `v1alpha2` version of the specification for its [Apache APISIX Ingress Controller][apisix-1], and status updates and documentation will be provided here as the work progresses.

[apisix]:https://apisix.apache.org/
[apisix-1]:https://github.com/apache/apisix-ingress-controller

### Contour

[Contour][contour] is an open source ingress controller for Kubernetes.

Contour currently implements much of Gateway API `v0.3.0`.

See the [Contour Gateway API Guide][contour-1] for information on how to deploy and use Contour's Gateway implementation.

_Note that not all of the `v0.3.0` specification is completed yet, [contributions welcome!][contour-2]._

[contour]:https://projectcontour.io
[latest]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/
[contour-1]:https://projectcontour.io/guides/gateway-api/
[contour-2]:https://github.com/projectcontour/contour/blob/main/CONTRIBUTING.md

### Emissary-Ingress (Ambassador API Gateway)

[Emissary-Ingress][emissary] (formerly known as Ambassador API Gateway) is an open source CNCF project that
provides an ingress controller and API gateway for Kubernetes built on top of [Envoy Proxy][envoy].
See [here][emissary-gateway-api] for more details on using the Gateway API with Emissary.

[emissary]:https://www.getambassador.io/docs/edge-stack
[envoy]:https://envoyproxy.io
[emissary-gateway-api]:https://www.getambassador.io/docs/edge-stack/latest/topics/using/gateway-api/

### Gloo Edge

Gloo Edge 2.0 is an Istio-native, fully-featured Envoy based API gateway that brings [Gloo Edge][glooedge] functionality to Istio. The [Solo.io][solo] team is actively working towards an implementation of the Gateway API.

[glooedge]:https://docs.solo.io/gloo-edge/
[solo]:https://www.solo.io

### Google Kubernetes Engine

[Google Kubernetes Engine (GKE)][gke] is a managed Kubernetes platform
offered by Google Cloud. GKE's implementation of the Gateway API is through the 
[GKE Gateway controller][gke-gateway] (currently in Preview) which provisions 
Google Cloud Load Balancers for Pods in GKE clusters. 

The GKE Gateway controller supports weighted traffic splitting, mirroring, 
advanced routing, multi-cluster load balancing and more. See the docs to deploy 
[private or public Gateways][gke-gateway-deploy] and also [multi-cluster Gateways][gke-multi-cluster-gateway]. 

[gke]:https://cloud.google.com/kubernetes-engine
[gke-gateway]:https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api
[gke-gateway-deploy]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways
[gke-multi-cluster-gateway]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-multi-cluster-gateways

### HAProxy Ingress

[HAProxy Ingress][h1] is a community driven ingress controller implementation for HAProxy.

HAProxy Ingress v0.13 partially supports the Gateway API's v1alpha1 specification. See the [controller's Gateway API documentation][h2] to get informed about conformance and roadmap.

[h1]:https://haproxy-ingress.github.io/
[h2]:https://haproxy-ingress.github.io/docs/configuration/gateway-api/

### Istio

[Istio][istio] is an open source [service mesh][mesh].

Istio is actively working towards an implementation of the Gateway API and you can use the [Istio Gateway API Documentation][istio-1] to deploy the Istio Gateway API in it's current state.

[istio]:https://istio.io
[mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### Kong

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

Kong is actively working towards an implementation of the Gateway API `v0.2.x` specification for it's [Kubernetes Ingress Controller][kic] and status updates and documentation will be provided here as the work progresses.

You can subscribe to [kong/kubernetes-ingress-controller/issues/692][kong-1] to track the implementation progress and [contribute][kong-2]!

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-1]:https://github.com/kong/kubernetes-ingress-controller/issues/692
[kong-2]:https://github.com/Kong/kubernetes-ingress-controller/blob/main/CONTRIBUTING.md

### Traefik

[Traefik][traefik] is an open source cloud-native application proxy.

Traefik currently supports the previous `v0.1.x` Gateway API specification, check the [Kubernetes Gateway Documentation][traefik-1] for information on how to deploy and use Traefik's Gateway implementation.

Traefik is currently working on implementing TCP, status updates and documentation will be provided here as the work progresses.

[traefik]:https://traefik.io
[traefik-1]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/
