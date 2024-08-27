# Implementations

This document tracks downstream implementations and integrations of Gateway API
and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this
document with status information about their implementations, the versions they
cover, and documentation to help users get started.


!!! info "Compare extended supported features across implementations"

    [View a table to quickly compare supported features of projects](/implementations/v1.1). These outline Gateway controller implementations that have passed core conformance tests, and focus on extended conformance features that they have implemented.

## Gateway Controller Implementation Status <a name="gateways"></a>

- [Acnodal EPIC][1]
- [Amazon Elastic Kubernetes Service][23] (alpha)
- [Apache APISIX][2] (beta)
- [Avi Kubernetes Operator][31] (tech preview)
- [Azure Application Gateway for Containers][27] (GA)
- [Cilium][16] (beta)
- [Contour][3] (GA)
- [Easegress][30] (GA)
- [Emissary-Ingress (Ambassador API Gateway)][4] (alpha)
- [Envoy Gateway][18] (GA)
- [Flomesh Service Mesh][17] (beta)
- [Gloo Gateway][5] (GA)
- [Google Kubernetes Engine][6] (GA)
- [HAProxy Ingress][7] (alpha)
- [HAProxy Kubernetes Ingress Controller][32] (GA)
- [HashiCorp Consul][8]
- [Istio][9] (GA)
- [Kong][10] (GA)
- [Kuma][11] (GA)
- [LiteSpeed Ingress Controller][19]
- [NGINX Gateway Fabric][12] (GA)
- [ngrok][33] (preview)
- [STUNner][21] (beta)
- [Traefik Proxy][13] (GA)
- [Tyk][29] (work in progress)
- [WSO2 APK][25] (GA)

## Service Mesh Implementation Status <a name="meshes"></a>

- [Istio][9] (GA)
- [Kuma][11] (GA)
- [Linkerd][28] (experimental)

## Integrations <a name="integrations"></a>

- [Flagger][14] (public preview)
- [cert-manager][15] (alpha)
- [argo-rollouts][22] (alpha)
- [Knative][24] (alpha)
- [Kuadrant][26] (work in progress)

[1]:#acnodal-epic
[2]:#apisix
[3]:#contour
[4]:#emissary-ingress-ambassador-api-gateway
[5]:#gloo-gateway
[6]:#google-kubernetes-engine
[7]:#haproxy-ingress
[8]:#hashicorp-consul
[9]:#istio
[10]:#kong
[11]:#kuma
[12]:#nginx-gateway-fabric
[13]:#traefik-proxy
[14]:#flagger
[15]:#cert-manager
[16]:#cilium
[17]:#flomesh-service-mesh-fsm
[18]:#envoy-gateway
[19]:#litespeed-ingress-controller
[21]:#stunner
[22]:#argo-rollouts
[23]:#amazon-elastic-kubernetes-service
[24]:#knative
[25]:#wso2-apk
[26]:#kuadrant
[27]:#azure-application-gateway-for-containers
[28]:#linkerd
[29]:#tyk
[30]:#easegress
[31]:#avi-kubernetes-operator
[32]:#haproxy-kubernetes-ingress-controller
[33]:#ngrok-kubernetes-operator

[gamma]:/concepts/gamma/



## Implementations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### Acnodal EPIC
[EPIC][epicdocs] is an Open Source External Gateway platform designed and built with Kubernetes.  It consists of the Gateway Cluster, k8s Gateway controller, a stand alone Linux Gateway controller and the Gateway Service Manager.  Together they create a platform for providing Gateway services to cluster users.   Each gateway consists of multiple Envoy instances running on the gateway cluster not the workload clusters. The Gateway Service Manager is a simple user management and UI that can be used to implement Gateway-as-a-Service infrastructure for public and private clusters, and integrate non-k8s endpoints.

- [Documentation][epicdocs]
- [Source Repo][epicsource]

[epicdocs]:https://www.epic-gateway.org/
[epicsource]:https://github.com/epic-gateway

### Amazon Elastic Kubernetes Service

[Amazon Elastic Kubernetes Service (EKS)][eks] is a managed service that you can use to run Kubernetes on AWS without needing to install, operate, and maintain your own Kubernetes control plane or nodes. EKS's implementation of the Gateway API is through [AWS Gateway API Controller][eks-gateway] which provisions [Amazon VPC Lattice][vpc-lattice] Resources for gateway(s), HTTPRoute(s) in EKS clusters.

[eks]:https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html
[eks-gateway]:https://github.com/aws/aws-application-networking-k8s
[vpc-lattice]:https://aws.amazon.com/vpc/lattice/

### APISIX

[Apache APISIX][apisix] is a dynamic, real-time, high-performance API Gateway. APISIX provides rich traffic management features such as load balancing, dynamic upstream, canary release, circuit breaking, authentication, observability, and more.

APISIX currently supports Gateway API `v1beta1` version of the specification for its [Apache APISIX Ingress Controller][apisix-1].

[apisix]:https://apisix.apache.org/
[apisix-1]:https://github.com/apache/apisix-ingress-controller

### Avi Kubernetes Operator

[Avi Kubernetes Operator (AKO)][ako] provides L4-L7 load-balancing using VMware NSX Advanced Load Balancer.

Starting with AKO version [v1.11.1], Gateway API version v0.6.2 is supported. It implements v1beta1 version of Gateway API specification supporting GatewayClass, Gateway and HTTPRoute objects. AKO Gateway API is currently in Tech Preview.

Documentation to deploy and use AKO Gateway API can be found at [Avi Kubernetes Operator Gateway API][ako-gw].

[ako]:https://docs.vmware.com/en/VMware-NSX-Advanced-Load-Balancer/1.11/Avi-Kubernetes-Operator-Guide/GUID-1E86BD1A-899F-40C2-A931-29310B2B340F.html
[ako-gw]:https://docs.vmware.com/en/VMware-NSX-Advanced-Load-Balancer/1.11/Avi-Kubernetes-Operator-Guide/GUID-84BD68AB-B96F-425C-8323-3A249D6AC8B2.html
[v1.11.1]:https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

### Azure Application Gateway for Containers

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Partial%20Conformance%20v1.0.0-Azure%20Application%20Gateway%20for%20Containers-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/azure-application-gateway-for-containers)

[Application Gateway for Containers][azure-application-gateway-for-containers] is a managed application (layer 7) load balancing solution, providing dynamic traffic management capabilities for workloads running in a Kubernetes cluster in Azure. Follow the [quickstart guide][azure-application-gateway-for-containers-quickstart-controller] to deploy the ALB controller and get started with Gateway API.


[azure-application-gateway-for-containers]:https://aka.ms/appgwcontainers/docs
[azure-application-gateway-for-containers-quickstart-controller]:https://learn.microsoft.com/azure/application-gateway/for-containers/quickstart-deploy-application-gateway-for-containers-alb-controller

### Cilium

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-Cilium-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/cilium)

[Cilium][cilium] is an eBPF-based networking, observability and security
solution for Kubernetes and other networking environments. It includes [Cilium
Service Mesh][cilium-service-mesh], a highly efficient mesh data plane that can
be run in [sidecarless mode][cilium-sidecarless] to dramatically improve
performance, and avoid the operational complexity of sidecars. Cilium also
supports the sidecar proxy model, offering choice to users. As of [Cilium 1.14][cilium114blog],
Cilium supports Gateway API, passing conformance for v0.7.1.

Cilium is open source and is a CNCF Graduates project.

If you have questions about Cilium Service Mesh the #service-mesh channel on
[Cilium Slack][cilium-slack] is a good place to start. For contributing to the development
effort, check out the #development channel or join our [weekly developer meeting][cilium-meeting].

[cilium]:https://cilium.io
[cilium-service-mesh]:https://docs.cilium.io/en/stable/gettingstarted/#service-mesh
[cilium-sidecarless]:https://isovalent.com/blog/post/cilium-service-mesh/
[cilium114blog]:https://isovalent.com/blog/post/cilium-release-114/
[cilium-slack]:https://cilium.io/slack
[cilium-meeting]:https://github.com/cilium/cilium#weekly-developer-meeting

### Contour

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.1.0-Contour-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/projectcontour-contour)

[Contour][contour] is a CNCF open source Envoy-based ingress controller for Kubernetes.

Contour [v1.30.0][contour-release] implements Gateway API v1.1.0.
All [Standard channel][contour-standard] v1 API group resources (GatewayClass, Gateway, HTTPRoute, ReferenceGrant), plus most v1alpha2 API group resources (TLSRoute, TCPRoute, GRPCRoute, ReferenceGrant, and BackendTLSPolicy) are supported.
Contour's implementation passes all core and most extended Gateway API conformance tests included in the v1.1.0 release.

See the [Contour Gateway API Guide][contour-guide] for information on how to deploy and use Contour's Gateway API implementation.

For help and support with Contour's implementation, [create an issue][contour-issue-new] or ask for help in the [#contour channel on Kubernetes slack][contour-slack].

[contour]:https://projectcontour.io
[contour-release]:https://github.com/projectcontour/contour/releases/tag/v1.30.0
[contour-standard]:https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard
[contour-guide]:https://projectcontour.io/docs/1.30/guides/gateway-api/
[contour-issue-new]:https://github.com/projectcontour/contour/issues/new/choose
[contour-slack]:https://kubernetes.slack.com/archives/C8XRH2R4J

### Easegress

[Easegress][easegress] is a Cloud Native traffic orchestration system.

It can function as a sophisticated modern gateway, a robust distributed cluster, a flexible traffic orchestrator, or even an accessible service mesh.

Easegress currently supports Gateway API `v1beta1` version of the specification by [GatewayController][easegress-gatewaycontroller].

[easegress]:https://megaease.com/easegress/
[easegress-gatewaycontroller]:https://github.com/megaease/easegress/blob/main/docs/04.Cloud-Native/4.2.Gateway-API.md

### Emissary-Ingress (Ambassador API Gateway)

[Emissary-Ingress][emissary] (formerly known as Ambassador API Gateway) is an open source CNCF project that
provides an ingress controller and API gateway for Kubernetes built on top of [Envoy Proxy][envoy].
See [here][emissary-gateway-api] for more details on using the Gateway API with Emissary.

[emissary]:https://www.getambassador.io/docs/edge-stack
[envoy]:https://envoyproxy.io
[emissary-gateway-api]:https://www.getambassador.io/docs/edge-stack/latest/topics/using/gateway-api/

### Envoy Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-EnvoyGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/envoy-gateway)

[Envoy Gateway][eg-home] is an [Envoy][envoy-org] subproject for managing Envoy-based application gateways. The supported
APIs and fields of the Gateway API are outlined [here][eg-supported].
Use the [quickstart][eg-quickstart] to get Envoy Gateway running with Gateway API in a
few simple steps.

[eg-home]:https://gateway.envoyproxy.io/
[envoy-org]:https://github.com/envoyproxy
[eg-supported]:https://gateway.envoyproxy.io/docs/tasks/quickstart/
[eg-quickstart]:https://gateway.envoyproxy.io/docs/tasks/quickstart

### Flomesh Service Mesh (FSM)

[Flomesh Service Mesh][fsm] is a community driven lightweight service mesh for Kubernetes East-West and North-South traffic management. Flomesh uses [ebpf](https://www.kernel.org/doc/html/latest/bpf/index.html) for layer4 and [pipy](https://flomesh.io/pipy) proxy for layer7 traffic management. Flomesh comes bundled with a load balancer, cross-cluster service registration/discovery and it supports multi-cluster networking. It supports `Ingress` (and as such is an "Ingress controller") and Gateway API.

FSM support of Gateway API is built on top [Flomesh Gateway API](fgw) and it currently supports Kubernetes Gateway API version [v0.7.1](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.7.1) with support for `v0.8.0` currently in progress.

- [FSM Kubernetes Gateway API compatibility matrix](https://github.com/flomesh-io/fsm/blob/main/docs/gateway-api-compatibility.md)
- [How to use Gateway API support in FSM](https://github.com/flomesh-io/fsm/blob/main/docs/tests/gateway-api/README.md)

[fsm]:https://github.com/flomesh-io/fsm
[flomesh]:https://flomesh.io
[fgw]:https://github.com/flomesh-io/fgw

### Gloo Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-GlooGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/gloo-gateway)

[Gloo Gateway][gloogateway] by [Solo.io][solo] is a feature-rich, Kubernetes-native ingress controller and next-generation API gateway.
Gloo Gateway brings the full power and community support of Gateway API to its existing control-plane implementation.

[gloogateway]:https://docs.solo.io/gateway/latest/
[solo]:https://www.solo.io

### Google Kubernetes Engine

[![Conformance](https://img.shields.io/badge/Gateway_API_Partial_Conformance_v1.1.0-Google_Kubernetes_Engine-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/gke-gateway)

[Google Kubernetes Engine (GKE)][gke] is a managed Kubernetes platform offered
by Google Cloud. GKE's implementation of the Gateway API is through the [GKE
Gateway controller][gke-gateway] which provisions Google Cloud Load Balancers
for Pods in GKE clusters.

The GKE Gateway controller supports weighted traffic splitting, mirroring,
advanced routing, multi-cluster load balancing and more. See the docs to deploy
[private or public Gateways][gke-gateway-deploy] and also [multi-cluster
Gateways][gke-multi-cluster-gateway].

The GKE Gateway controller passes all the core Gateway API conformance tests in the
v1.1.0 release for the GATEWAY_HTTP conformance profile except `HTTPRouteHostnameIntersection`.

[gke]:https://cloud.google.com/kubernetes-engine
[gke-gateway]:https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api
[gke-gateway-deploy]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways
[gke-multi-cluster-gateway]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-multi-cluster-gateways

### HAProxy Ingress

[HAProxy Ingress][h1] is a community driven ingress controller implementation for HAProxy.

HAProxy Ingress v0.13 partially supports the Gateway API's v1alpha1 specification. See the [controller's Gateway API documentation][h2] to get informed about conformance and roadmap.

[h1]:https://haproxy-ingress.github.io/
[h2]:https://haproxy-ingress.github.io/docs/configuration/gateway-api/

### HAProxy Kubernetes Ingress Controller

HAProxy Kubernetes Ingress Controller is an open-source project maintained by HAProxy Technologies that provides fast and efficient traffic management, routing, and observability for Kubernetes. It has built-in support for the Gateway API since version 1.10. The same deployment of the ingress controller will allow you to use both the Ingress API and Gateway API. See the [documentation][haproxytech-docs-gw] for more details. In the [GitHub repository][haproxytech-github-gw], you will also find additional information about supported API resources.

[haproxytech-docs-gw]:https://www.haproxy.com/documentation/kubernetes-ingress/gateway-api/enable-gateway-api/
[haproxytech-github-gw]:https://github.com/haproxytech/kubernetes-ingress/blob/master/documentation/gateway-api.md

### HashiCorp Consul

[Consul][consul], by [HashiCorp][hashicorp], is an open source control plane for multi-cloud networking. A single Consul deployment can span bare metal, VM and container environments.

Consul service mesh works on any Kubernetes distribution, connects multiple clusters, and Consul CRDs provide a Kubernetes native workflow to manage traffic patterns and permissions in the mesh. [Consul API Gateway][consul-api-gw-doocs] supports Gateway API for managing North-South traffic.

Please see the [Consul API Gateway documentation][consul-api-gw-doocs] for current information on the supported version and features of the Gateway API.

[consul]:https://consul.io
[consul-api-gw-doocs]:https://www.consul.io/docs/api-gateway
[hashicorp]:https://www.hashicorp.com

### Istio

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.1.0-Istio-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/istio-istio)

[Istio][istio] is an open source [service mesh][istio-mesh] and gateway implementation.

A minimal install of Istio can be used to provide a fully compliant
implementation of the Kubernetes Gateway API for cluster ingress traffic
control. For service mesh users, Istio also fully supports the [GAMMA
initiative's][gamma] experimental Gateway API [support for east-west traffic
management][gamma] within the mesh.

Much of Istio's documentation, including all of the [ingress tasks][istio-1] and several mesh-internal traffic management tasks, already includes parallel instructions for
configuring traffic using either the Gateway API or the Istio configuration API.
Check out the [Gateway API task][istio-2] for more information about the Gateway API implementation in Istio.

[istio]:https://istio.io
[istio-mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/
[istio-2]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### Kong

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-Kong%20Ingress%20Controller-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/kong-kubernetes-ingress-controller)

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

Kong supports Gateway API in the [Kong Kubernetes Ingress Controller (KIC)][kic], see the [Gateway API Guide][kong-gw-guide] for usage information.

Kong also supports Gateway API in the [Kong Gateway Operator][kgo].

For help and support with Kong's implementations please feel free to [create an issue][kong-issue-new] or a [discussion][kong-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-gw-guide]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/using-gateway-api/
[kgo]:https://docs.konghq.com/gateway-operator/latest/
[kong-issue-new]:https://github.com/Kong/kubernetes-ingress-controller/issues/new
[kong-disc-new]:https://github.com/Kong/kubernetes-ingress-controller/discussions/new
[kong-slack]:https://kubernetes.slack.com/archives/CDCA87FRD

### Kuma

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-Kuma-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/kumahq-kuma)

[Kuma][kuma] is an open source service mesh.

Kuma implements the Gateway API specification for the Kuma built-in, Envoy-based Gateway with a beta stability guarantee. Check the [Gateway API Documentation][kuma-1] for information on how to setup a Kuma built-in gateway using the Gateway API.

Kuma 2.3 and later support the [GAMMA initiative's][gamma] experimental
Gateway API [support for east-west traffic management][gamma] within the mesh.

[kuma]:https://kuma.io
[kuma-1]:https://kuma.io/docs/latest/using-mesh/managing-ingress-traffic/gateway-api/

### Linkerd

[Linkerd][linkerd] is the first CNCF graduated [service mesh][linkerd-mesh].
It is the only major mesh not based on Envoy, instead relying on a
purpose-built Rust micro-proxy to bring security, observability, and
reliability to Kubernetes, without the complexity.

Linkerd 2.14 and later support the [GAMMA initiative's][gamma] experimental
Gateway API [support for east-west traffic management][gamma] within the mesh.

[linkerd]:https://linkerd.io/
[linkerd-mesh]:https://buoyant.io/service-mesh-manifesto

### LiteSpeed Ingress Controller

The [LiteSpeed Ingress Controller](https://litespeedtech.com/products/litespeed-web-adc/features/litespeed-ingress-controller) uses the LiteSpeed WebADC controller to operate as an Ingress Controller and Load Balancer to manage your traffic on your Kubernetes cluster.  It implements the full core Gateway API including Gateway, GatewayClass, HTTPRoute and ReferenceGrant and the Gateway functions of cert-manager.  Gateway is fully integrated into the LiteSpeed Ingress Controller.

- [Product documentation](https://docs.litespeedtech.com/cloud/kubernetes/).
- [Gateway specific documentation](https://docs.litespeedtech.com/cloud/kubernetes/gateway).
- Full support is available on the [LiteSpeed support web site](https://www.litespeedtech.com/support).

### NGINX Gateway Fabric

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.1.0-NGINX Gateway Fabric-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/nginxinc-nginx-gateway-fabric)

[NGINX Gateway Fabric][nginx-gateway-fabric] is an open-source project that provides an implementation of the Gateway API using [NGINX][nginx] as the data plane. The goal of this project is to implement the core Gateway API to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. You can find the comprehensive NGINX Gateway Fabric user documentation on the [NGINX Documentation][nginx-docs] website.

For a list of supported Gateway API resources and features, see the [Gateway API Compatibility][nginx-compat] doc.

If you have any suggestions or experience issues with NGINX Gateway Fabric, please [create an issue][nginx-issue-new] or a [discussion][nginx-disc-new] on GitHub. You can also ask for help in the [#nginx-gateway-fabric channel on NGINX slack][nginx-slack].

[nginx-gateway-fabric]:https://github.com/nginxinc/nginx-gateway-fabric
[nginx]:https://nginx.org/
[nginx-docs]:https://docs.nginx.com/nginx-gateway-fabric/
[nginx-compat]:https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-api-compatibility/
[nginx-issue-new]:https://github.com/nginxinc/nginx-gateway-fabric/issues/new
[nginx-disc-new]:https://github.com/nginxinc/nginx-gateway-fabric/discussions/new
[nginx-slack]:https://nginxcommunity.slack.com/channels/nginx-gateway-fabric


### ngrok Kubernetes Operator

[ngrok Kubernetes Operator][ngrok-k8s-operator] provides an implementation of the Gateway API that uses [ngrok's ingress-as-a-service][ngrok]. This project uses the Gateway API to support routing traffic from ngrok's global network to applications running on Kubernetes clusters. This easily adds the benefits of ngrok, like security, network policy, and a global presence with the simplicity of cloud service. The operator contains both a Gateway API implementation as well as a controller using Kubernetes Ingress. The Gateway API implementation is currently under development and supports only the Gateway, GatewayClass and HTTPRoute. As the TLSRoute and TCPRoute move from experimental to stable, they will also be implemented.

You can read our [docs][ngrok-k8s-gwapi-docs] for more information. If you have any feature requests or bug reports, please [create an issue][ngrok-issue-new]. You can also reach out for help on [Slack][ngrok-slack]

[ngrok-k8s-operator]:https://github.com/ngrok/kubernetes-ingress-controller
[ngrok]:https://ngrok.com
[ngrok-k8s-gwapi-docs]:https://ngrok.com/docs/k8s/
[ngrok-issue-new]: https://github.com/ngrok/kubernetes-ingress-controller/issues/new
[ngrok-slack]:https://ngrokcommunity.slack.com/channels/general

### STUNner

[STUNner][stunner] is an open source cloud-native WebRTC media gateway for Kubernetes. STUNner is purposed specifically to facilitate the seamless ingestion of WebRTC media streams into a Kubernetes cluster, with simplified NAT traversal and dynamic media routing. Meanwhile, STUNner provides improved security and monitoring for large-scale real-time communications services. The STUNner dataplane exposes a standards compliant TURN service to WebRTC clients, while the control plane supports a subset of the Gateway API.

STUNner currently supports version `v1alpha2` of the Gateway API specification. Check the [install guide][stunner-1] for information on how to deploy and use STUNner for WebRTC media ingestion. Please direct all questions, comments and bug-reports related to STUNner to the [STUNner project][stunner].

[stunner]:https://github.com/l7mp/stunner
[stunner-1]:https://github.com/l7mp/stunner/blob/main/doc/INSTALL.md

### Traefik Proxy

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.1.0-Traefik Proxy-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/traefik-traefik)

[Traefik Proxy][traefik-proxy] is an open source cloud-native application proxy.

Traefik Proxy currently supports version `v1.1.0` of the Gateway API specification, check the [Kubernetes Gateway Provider Documentation][traefik-proxy-gateway-api-doc] for more information on how to deploy and use it.
Traefik Proxy's implementation passes all HTTP core and some extended conformance tests, but also supports the TCPRoute and TLSRoute features from the Experimental channel.

For help and support with Traefik Proxy, [create an issue][traefik-proxy-issue-new] or ask for help in the [Traefik Labs Community Forum][traefiklabs-community-forum].

[traefik-proxy]:https://traefik.io
[traefik-proxy-gateway-api-doc]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/
[traefik-proxy-issue-new]:https://github.com/traefik/traefik/issues/new/choose
[traefiklabs-community-forum]:https://community.traefik.io/c/traefik/traefik-v3/21

### Tyk

[Tyk Gateway][tyk-gateway] is a cloud-native, open source, API Gateway.

The [Tyk.io][tyk] team is working towards an implementation of the Gateway API. You can track progress of this project [here][tyk-operator].

[tyk]:https://tyk.io
[tyk-gateway]:https://github.com/TykTechnologies/tyk
[tyk-operator]:https://github.com/TykTechnologies/tyk-operator

### WSO2 APK

[WSO2 APK][wso2-apk] is a purpose-built API management solution tailored for Kubernetes environments, delivering seamless integration, flexibility, and scalability to organizations in managing their APIs.

WSO2 APK implements the Gateway API, encompassing Gateway and HTTPRoute functionalities. Additionally, it provides support for rate limiting, authentication/authorization, and analytics/observability through the use of Custom Resources (CRs).

For up-to-date information on the supported version and features of the Gateway API, please refer to the [APK Gateway documentation][apk-doc]. If you have any questions or would like to contribute, feel free to create [issues or pull requests][repo]. Join our [Discord channel][discord] to connect with us and engage in discussions.

[wso2-apk]:https://apk.docs.wso2.com/en/latest/
[apk-doc]:https://apk.docs.wso2.com/en/latest/catalogs/kubernetes-crds/
[repo]:https://github.com/wso2/apk
[discord]:https://discord.com/channels/955510916064092180/1113056079501332541

## Integrations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific integrations.

### Flagger

[Flagger][flagger] is a progressive delivery tool that automates the release process for applications running on Kubernetes.

Flagger can be used to automate canary deployments and A/B testing using Gateway API. It supports both the `v1alpha2` and `v1beta1` spec of Gateway API. You can refer to [this tutorial][flagger-tutorial] to use Flagger with any implementation of Gateway API.

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

### Knative

[Knative][knative] is a serverless platform built on Kubernetes.  Knative Serving provides a simple API for running stateless containers with automatic management of URLs, traffic splitting between revisions, request-based autoscaling (including scale to zero), and automatic TLS provisioning.  Knative Serving supports multiple HTTP routers through a plugin architecture, including a [gateway API plugin][knative-net-gateway-api] which is currently in alpha as not all Knative features are supported.

[knative]:https://knative.dev/
[knative-net-gateway-api]:https://github.com/knative-sandbox/net-gateway-api

### Kuadrant

[Kuadrant][kuadrant] is an open source multi cluster Gateway API controller that integrates with and provides policies to other Gateway API providers.

Kuadrant supports Gateway API for defining gateways centrally and attaching policies such as DNS, TLS, Auth and Rate Limiting that apply to all gateway instances in a multi cluster environment.
Kuadrant works with Istio as the underlying gateway provider, with plans to work with other gateway providers such as Envoy Gateway.

For help and support with Kuadrant's implementation please feel free to [create an issue][kuadrant-issue-new] or ask for help in the [#kuadrant channel on Kubernetes slack][kuadrant-slack].

[kuadrant]:https://kuadrant.io/
[kuadrant-issue-new]:https://github.com/Kuadrant/multicluster-gateway-controller/issues/new
[kuadrant-slack]:https://kubernetes.slack.com/archives/C05J0D0V525
