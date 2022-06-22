# Implementations

This document tracks downstream implementations and integrations of Gateway API and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this document with status information about their implementations, the versions they cover, and documentation to help users get started.

## Implementation Status


- [Acnodal EPIC][1] (public preview)
- [Apache APISIX][2] (work in progress)
- [Contour][3] (alpha)
- [Emissary-Ingress (Ambassador API Gateway)][4] (alpha)
- [Gloo Edge 2.0][5] (work in progress)
- [Google Kubernetes Engine][6] (public preview)
- [HAProxy Ingress][7] (alpha)
- [HashiCorp Consul][8]
- [Istio][9] (alpha)
- [Kong][10] (alpha)
- [Kuma][11] (alpha)
- [NGINX Kubernetes Gateway][12] (pre-alpha)
- [Traefik][13] (alpha)

## Integration Status
- [Flagger][14] (public preview)
- [cert-manager][15] (alpha)

[1]:#acnodal-epic
[2]:#apisix
[3]:#contour
[4]:#emissary-ingress-ambassador-api-gateway
[5]:#gloo-edge
[6]:#google-kubernetes-engine
[7]:#haproxy-ingress
[8]:#hashicorp-consul
[9]:#istio
[10]:#kong
[11]:#kuma
[12]:#nginx-kubernetes-gateway
[13]:#traefik
[14]:#flagger
[15]:#cert-manager

## Implementations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### Acnodal EPIC
[EPIC][epic] is a Managed Application & API Gateway Service.  The epic-controller installed in the cluster implements v1alpha2 and currently supports HTTPRoute.  Defining Gateways and Routes creates a Gateway in the EPIC Service consisting of Envoy instances allocating public IP addresses and DNS for clients, and configures transport that sends request directly upstream to application endpoints in the cluster.  EPIC is in public preview.
 
Documentation can be found at [EPIC Application & API Gateway Service][epic]
 
[epic]:https://www.epick8sgw.io

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

### HashiCorp Consul

[Consul][consul], by [HashiCorp][hashicorp], is an open source control plane for multi-cloud networking. A single Consul deployment can span bare metal, VM and container environments.  

Consul service mesh works on any Kubernetes distribution, connects multiple clusters, and Consul CRDs provide a Kubernetes native workflow to manage traffic patterns and permissions in the mesh. [Consul API Gateway][consul-api-gw-doocs] supports Gatewway API for managing North-South traffic.

Please see the [Consul API Gateway documentation][consul-api-gw-doocs] for current infomation on the supported version and features of the Gateway API.

[consul]:https://consul.io
[consul-api-gw-doocs]:https://www.consul.io/docs/api-gateway
[hashicorp]:https://www.hashicorp.com

### Istio

[Istio][istio] is an open source [service mesh][mesh].

Istio is actively working towards an implementation of the Gateway API and you can use the [Istio Gateway API Documentation][istio-1] to deploy the Istio Gateway API in it's current state.

[istio]:https://istio.io
[mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### Kong

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

Kong supports Gateway API in the [Kong Kubernetes Ingress Controller (KIC)][kic], see the [Gateway API Guide][kong-gw-guide] for usage information.

Kong also supports Gateway API in the [Kong Gateway Operator][kgo].

For help and support with Kong's implementations please feel free to [create an issue][kong-issue-new] or a [discussion][kong-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-gw-guide]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/using-gateway-api/
[kgo]:https://github.com/kong/gateway-operator
[kong-issue-new]:https://github.com/Kong/kubernetes-ingress-controller/issues/new
[kong-disc-new]:https://github.com/Kong/kubernetes-ingress-controller/discussions/new
[kong-slack]:https://kubernetes.slack.com/archives/CDCA87FRD

### Kuma

[Kuma][kuma] is an open source service mesh.

Kuma is actively working on an implementation of Gateway API specification for the Kuma builtin Gateway. Check the [Gateway API Documentation][kuma-1] for information on how to setup a Kuma builtin gateway using the Gateway API.

[kuma]:https://kuma.io
[kuma-1]:https://kuma.io/docs/latest/explore/gateway-api/

### NGINX Kubernetes Gateway

[NGINX Kubernetes Gateway][nginx-kubernetes-gateway] is an open-source project that provides an implementation of the Gateway API using [NGINX][nginx] as the data plane. The goal of this project is to implement the core Gateway APIs -- Gateway, GatewayClass, HTTPRoute, TCPRoute, TLSRoute, and UDPRoute -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

[nginx-kubernetes-gateway]:https://github.com/nginxinc/nginx-kubernetes-gateway
[nginx]:https://nginx.org/

### Traefik

[Traefik][traefik] is an open source cloud-native application proxy.

Traefik currently supports version `v1alpha2` (`v0.4.x`) of the Gateway API specification, check the [Kubernetes Gateway Documentation][traefik-1] for information on how to deploy and use Traefik's Gateway implementation.

Traefik is currently working on implementing UDP, and ReferenceGrant. Status updates and documentation will be provided here as the work progresses.

[traefik]:https://traefik.io
[traefik-1]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/

## Integrations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific integrations.

### Flagger

[Flagger][flagger] is a progressive delivery tool that automates the release process for applications running on Kubernetes.

Flagger can be used to automate canary deployments and A/B testing using Gateway API. It currently supports the `v1alpha2` spec of Gateway API. You can refer to [this tutorial][flagger-tutorial] to use Flagger with any implementation of Gateway API.

[flagger]:https://flagger.app
[flagger-tutorial]:https://docs.flagger.app/tutorials/gatewayapi-progressive-delivery

### cert-manager

[cert-manager][cert-manager] is a tool to automate certificate management in cloud native environments.

cert-manager can generate TLS certificates for Gateway resources. This is configured by adding annotations to a Gateway. It currently supports the `v1alpha2` spec of Gateway API. You can refer to the [cert-manager docs][cert-manager-docs] to try it out.

[cert-manager]:https://cert-manager.io/
[cert-manager-docs]:https://cert-manager.io/docs/usage/gateway/
