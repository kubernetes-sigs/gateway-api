# Implementations

This document tracks downstream implementations and integrations of Gateway API and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this document with status information about their implementations, the versions they cover, and documentation to help users get started.

## Implementation Status

- [Acnodal EPIC][1] (public preview)
- [Amazon Elastic Kubernetes Service][23] (alpha)
- [Apache APISIX][2] (alpha)
- [BIG-IP Kubernetes Gateway][20] (beta)
- [Cilium][16] (beta)
- [Contour][3] (beta)
- [Emissary-Ingress (Ambassador API Gateway)][4] (alpha)
- [Envoy Gateway][18] (alpha)
- [Flomesh Service Mesh][17] (work in progress)
- [Gloo Edge 2.0][5] (work in progress)
- [Google Kubernetes Engine][6] (GA)
- [HAProxy Ingress][7] (alpha)
- [HashiCorp Consul][8]
- [Istio][9] (beta)
- [Kong][10] (beta)
- [Kuma][11] (alpha)
- [LiteSpeed Ingress Controller][19]
- [NGINX Kubernetes Gateway][12]
- [STUNner][21] (beta)
- [Traefik][13] (alpha)

## Integration Status
- [Flagger][14] (public preview)
- [cert-manager][15] (alpha)
- [argo-rollouts][22] (alpha)

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
[16]:#cilium
[17]:#flomesh-service-mesh-fsm
[18]:#envoy-gateway
[19]:#litespeed-ingress-controller
[20]:#big-ip-kubernetes-gateway
[21]:#stunner
[22]:#argo-rollouts
[23]:#amazon-elastic-kubernetes-service

## Implementations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### Acnodal EPIC
[EPIC][epic] is a Managed Application & API Gateway Service.  The epic-controller installed in the cluster implements v1alpha2 and currently supports HTTPRoute.  Defining Gateways and Routes creates a Gateway in the EPIC Service consisting of Envoy instances allocating public IP addresses and DNS for clients, and configures transport that sends request directly upstream to application endpoints in the cluster.  EPIC is in public preview.

Documentation can be found at [EPIC Application & API Gateway Service][epic]

[epic]:https://www.epick8sgw.io

### Amazon Elastic Kubernetes Service

[Amazon Elastic Kubernetes Service (EKS)][eks] is a managed service that you can use to run Kubernetes on AWS without needing to install, operate, and maintain your own Kubernetes control plane or nodes. EKS's implementation of the Gateway API is through [AWS Gateway API Controller][eks-gateway] which provisions [Amazon VPC Lattice][vpc-lattice] Resources for gateway(s), HTTPRoute(s) in EKS clusters.

[eks]:https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html
[eks-gateway]:https://github.com/aws/aws-application-networking-k8s
[vpc-lattice]:https://aws.amazon.com/vpc/lattice/

### APISIX

[Apache APISIX][apisix] is a dynamic, real-time, high-performance API Gateway. APISIX provides rich traffic management features such as load balancing, dynamic upstream, canary release, circuit breaking, authentication, observability, and more.

APISIX currently supports Gateway API `v1alpha2` version of the specification for its [Apache APISIX Ingress Controller][apisix-1].

[apisix]:https://apisix.apache.org/
[apisix-1]:https://github.com/apache/apisix-ingress-controller

### BIG-IP Kubernetes Gateway

[BIG-IP Kubernetes Gateway][big-ip-kubernetes-gateway] is an open-source project that provides an implementation of the Gateway API using [F5 BIG-IP][f5bigip] as the data plane. It provides enterprises with high-performance Gateway API implementation.

We are actively supporting various features of the Gateway API. For compatibility with the features of the Gateway API, please refer to [here][bigipgwfeatures]. For any questions about this project, welcome to create [Issues][bigipgwissues] or [PR][bigipgwpr]. Also, you are welcome to connect with us in the [slack channel][bigipgwslacklink].


[big-ip-kubernetes-gateway]:https://gateway-api.f5se.io/
[f5bigip]:https://f5.com
[bigipgwfeatures]:https://github.com/f5devcentral/bigip-kubernetes-gateway/blob/master/docs/gateway-api-compatibility.md
[bigipgwissues]:https://github.com/f5devcentral/bigip-kubernetes-gateway/issues
[bigipgwpr]:https://github.com/f5devcentral/bigip-kubernetes-gateway/pulls
[bigipgwslacklink]: https://gateway-api.f5se.io/Support-and-contact/

### Cilium

[Cilium][cilium] is an eBPF-based networking, observability and security
solution for Kubernetes and other networking environments. It includes [Cilium
Service Mesh][cilium-service-mesh], a highly efficient mesh data plane that can
be run in [sidecarless mode][cilium-sidecarless] to dramatically improve
performance, and avoid the operational complexity of sidecars. Cilium also
supports the sidecar proxy model, offering choice to users. As of [Cilium 1.13][cilium113blog],
Cilium supports Gateway API, passing conformance for v0.5.1.

Cilium is open source and is a CNCF incubation project. 

If you have questions about Cilium Service Mesh the #service-mesh channel on
[Cilium Slack][cilium-slack] is a good place to start. For contributing to the development
effort, check out the #development channel or join our [weekly developer meeting][cilium-meeting].

[cilium]:https://cilium.io
[cilium-service-mesh]:https://docs.cilium.io/en/stable/gettingstarted/#service-mesh
[cilium-sidecarless]:https://isovalent.com/blog/post/cilium-service-mesh/
[cilium113blog]:https://isovalent.com/blog/post/cilium-release-113/
[cilium-slack]:https://cilium.io/slack
[cilium-meeting]:https://github.com/cilium/cilium#weekly-developer-meeting

### Contour

[Contour][contour] is a CNCF open source Envoy-based ingress controller for Kubernetes.

Contour implements Gateway API v0.5.1, supporting the v1alpha2 and v1beta1 API versions.
All [Standard channel][contour-standard] resources (GatewayClass, Gateway, HTTPRoute), plus ReferenceGrant and TLSRoute, are supported.
Contour's implementation passes all Gateway API conformance tests included in the v0.5.1 release.

See the [Contour Gateway API Guide][contour-guide] for information on how to deploy and use Contour's Gateway API implementation.

For help and support with Contour's implementation, [create an issue][contour-issue-new] or ask for help in the [#contour channel on Kubernetes slack][contour-slack].

_Some "extended" functionality is not implemented yet, [contributions welcome!][contour-contrib]._

[contour]:https://projectcontour.io
[contour-standard]:https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard
[contour-guide]:https://projectcontour.io/guides/gateway-api/
[contour-issue-new]:https://github.com/projectcontour/contour/issues/new/choose
[contour-slack]:https://kubernetes.slack.com/archives/C8XRH2R4J
[contour-contrib]:https://github.com/projectcontour/contour/blob/main/CONTRIBUTING.md

### Emissary-Ingress (Ambassador API Gateway)

[Emissary-Ingress][emissary] (formerly known as Ambassador API Gateway) is an open source CNCF project that
provides an ingress controller and API gateway for Kubernetes built on top of [Envoy Proxy][envoy].
See [here][emissary-gateway-api] for more details on using the Gateway API with Emissary.

[emissary]:https://www.getambassador.io/docs/edge-stack
[envoy]:https://envoyproxy.io
[emissary-gateway-api]:https://www.getambassador.io/docs/edge-stack/latest/topics/using/gateway-api/

### Envoy Gateway

[Envoy Gateway][eg-home] is an [Envoy][envoy-org] subproject for managing Envoy-based application gateways. The supported
APIs and fields of the Gateway API are outlined [here][eg-supported]. 
Use the [quickstart][eg-quickstart] to get Envoy Gateway running with Gateway API in a
few simple steps.

[eg-home]:https://gateway.envoyproxy.io/
[envoy-org]:https://github.com/envoyproxy
[eg-supported]: https://gateway.envoyproxy.io/v0.4.0/design/gatewayapi-support.html
[eg-quickstart]:https://gateway.envoyproxy.io/v0.4.0/user/quickstart.html

### Flomesh Service Mesh (FSM)

[Flomesh Service Mesh][fsm] is a community driven Kubernetes North-South traffic manager, and provides an implementation of Ingress controller, Gateway API, Load Balancer, and cross-cluster service registration and service discovery.

The [Flomesh.io][flomesh] team is actively working towards an implementation of the Gateway API. You can track progress of this implementation [here](https://github.com/flomesh-io/fsm/issues/18).

[fsm]:https://github.com/flomesh-io/fsm
[flomesh]:https://flomesh.io

### Gloo Edge

Gloo Edge 2.0 is an Istio-native, fully-featured Envoy based API gateway that brings [Gloo Edge][glooedge] functionality to Istio. The [Solo.io][solo] team is actively working towards an implementation of the Gateway API.

[glooedge]:https://docs.solo.io/gloo-edge/
[solo]:https://www.solo.io

### Google Kubernetes Engine

[Google Kubernetes Engine (GKE)][gke] is a managed Kubernetes platform offered
by Google Cloud. GKE's implementation of the Gateway API is through the [GKE
Gateway controller][gke-gateway] which provisions Google Cloud Load Balancers
for Pods in GKE clusters.

The GKE Gateway controller supports weighted traffic splitting, mirroring,
advanced routing, multi-cluster load balancing and more. See the docs to deploy
[private or public Gateways][gke-gateway-deploy] and also [multi-cluster
Gateways][gke-multi-cluster-gateway].

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

[Istio][istio] is an open source [service mesh][mesh] and gateway implementation.

A light-weight minimal install of Istio can be used to provide a Beta-quality implementation of the Kubernetes Gateway API for cluster ingress traffic control. For service mesh users,
the Istio implementation also lets you start trying out the experimental Gateway API [support for east-west traffic management][gamma] within the mesh.

Much of Istio's documentation, including all of the [ingress tasks][istio-1] and several mesh-internal traffic management tasks, already includes parallel instructions for
configuring traffic using either the Gateway API or the Istio configuration API.
Check out the [Gateway API task][istio-2] for more information about the Gateway API implementation in Istio.

[istio]:https://istio.io
[mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[gamma]:https://gateway-api.sigs.k8s.io/contributing/gamma/
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/
[istio-2]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### Kong

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

Kong supports Gateway API in the [Kong Kubernetes Ingress Controller (KIC)][kic], see the [Gateway API Guide][kong-gw-guide] for usage information.

Kong also supports Gateway API in the [Kong Gateway Operator][kgo].

For help and support with Kong's implementations please feel free to [create an issue][kong-issue-new] or a [discussion][kong-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-gw-guide]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/using-gateway-api/
[kgo]:https://github.com/kong/gateway-operator-docs
[kong-issue-new]:https://github.com/Kong/kubernetes-ingress-controller/issues/new
[kong-disc-new]:https://github.com/Kong/kubernetes-ingress-controller/discussions/new
[kong-slack]:https://kubernetes.slack.com/archives/CDCA87FRD

### Kuma

[Kuma][kuma] is an open source service mesh.

Kuma is actively working on an implementation of Gateway API specification for the Kuma builtin Gateway. Check the [Gateway API Documentation][kuma-1] for information on how to setup a Kuma builtin gateway using the Gateway API.

[kuma]:https://kuma.io
[kuma-1]:https://kuma.io/docs/latest/explore/gateway-api/

### LiteSpeed Ingress Controller

The [LiteSpeed Ingress Controller](https://litespeedtech.com/products/litespeed-web-adc/features/litespeed-ingress-controller) uses the LiteSpeed WebADC controller to operate as an Ingress Controller and Load Balancer to manage your traffic on your Kubernetes cluster.  It implements the full core Gateway API including Gateway, GatewayClass, HTTPRoute and ReferenceGrant and the Gateway functions of cert-manager.  Gateway is fully integrated into the LiteSpeed Ingress Controller.

- [Product documentation](https://docs.litespeedtech.com/cloud/kubernetes/).
- [Gateway specific documentation](https://docs.litespeedtech.com/cloud/kubernetes/gateway).
- Full support is available on the [LiteSpeed support web site](https://www.litespeedtech.com/support).

### NGINX Kubernetes Gateway

[NGINX Kubernetes Gateway][nginx-kubernetes-gateway] is an open-source project that provides an implementation of the Gateway API using [NGINX][nginx] as the data plane. The goal of this project is to implement the core Gateway API -- Gateway, GatewayClass, HTTPRoute, TCPRoute, TLSRoute, and UDPRoute -- to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. NGINX Kubernetes Gateway is currently under development and supports a subset of the Gateway API.

If you have any suggestions or experience issues with NGINX Kubernetes Gateway, please [create an issue][nginx-issue-new] or a [discussion][nginx-disc-new] on GitHub. You can also ask for help in the [#nginx-kubernetes-gateway channel on NGINX slack][nginx-slack].

[nginx-kubernetes-gateway]:https://github.com/nginxinc/nginx-kubernetes-gateway
[nginx]:https://nginx.org/
[nginx-issue-new]:https://github.com/nginxinc/nginx-kubernetes-gateway/issues/new
[nginx-disc-new]:https://github.com/nginxinc/nginx-kubernetes-gateway/discussions/new
[nginx-slack]:https://nginxcommunity.slack.com/channels/nginx-kubernetes-gateway

### STUNner

[STUNner][stunner] is an open source cloud-native WebRTC media gateway for Kubernetes. STUNner is purposed specifically to facilitate the seamless ingestion of WebRTC media streams into a Kubernetes cluster, with simplified NAT traversal and dynamic media routing. Meanwhile, STUNner provides improved security and monitoring for large-scale real-time communications services. The STUNner dataplane exposes a standards compliant TURN service to WebRTC clients, while the control plane supports a subset of the Gateway API.

STUNner currently supports version `v1alpha2` of the Gateway API specification. Check the [install guide][stunner-1] for information on how to deploy and use STUNner for WebRTC media ingestion. Please direct all questions, comments and bug-reports related to STUNner to the [STUNner project][stunner].

[stunner]:https://github.com/l7mp/stunner
[stunner-1]:https://github.com/l7mp/stunner/blob/main/doc/INSTALL.md

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

### Argo rollouts

[Argo Rollouts][argo-rollouts] is a progressive delivery controller for Kubernetes. It supports several advanced deployment methods such as blue/green and canaries. Argo Rollouts supports the Gateway API via [a plugin][argo-rollouts-plugin].

[argo-rollouts]:https://argo-rollouts.readthedocs.io/en/stable/
[argo-rollouts-plugin]:https://github.com/argoproj-labs/rollouts-gatewayapi-trafficrouter-plugin/
