# Implementations

This document tracks downstream implementations and integrations of Gateway API
and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this
document with status information about their implementations, the versions they
cover, and documentation to help users get started. This status information should
be no longer than a few paragraphs.

## Conformance levels

There are three levels of Gateway API conformance:

### Conformant implementations

These implementations have submitted at least one conformance report that has passes for:

  * All core conformance tests for at least one combination of Route type and
    Profile
  * All claimed Extended features

for one of the two (2) most recent Gateway API releases.

So, it's conformant to support Mesh + HTTPRoute, or Gateway + HTTPRoute, or
Gateway + TLSRoute, or Gateway + Mesh + HTTPRoute, plus any extended features
the implementation claims. But implementations _must_ support at least one
Profile and one Route type in that profile, and must pass all Core conformance
tests for that Profile and Route type in addition to all claimed Extended
features.

### Partially Conformant implementations

These implementations are aiming for full conformance but are not currently
achieving it. They have submitted at least one conformance report passing some
of the tests to be Conformant (as above) for one of the three (3) most recent
Gateway API releases. Note that the requirements to be considered "partially
conformant" may be tightened in a future release of Gateway API.

### Stale implementations

These implementations may not be being actively developed and will be removed
from this page on the next page review unless they submit a conformance report
moving them to one of the other categories.

Page reviews are performed at least one month after every Gateway API release,
with the first being performed after the release of Gateway API v1.3, in late
June 2025. Following the Gateway API v1.5 review process, due in mid-2026,
stale implementations will no longer be listed.

## Implementation profiles

Implementations also generally fall into two categories, which are called
_profiles_:

* **Gateway** controllers reconcile the Gateway resource and are intended to
handle north-south traffic, mainly concerned with coming from outside the
cluster to inside.
* **Mesh** controllers reconcile Service resources with HTTPRoutes attached
and are intended to handle east-west traffic, within the same cluster or
set of clusters.

Each profile has a set of conformance tests associated with it, that lay out
the expected behavior for implementations to be conformant (as above).

Implementations may also fit both profiles.

## Integrations

Also listed on this page are **integrations**, which are other software
projects that are able to make use of Gateway API resources to perform
other functions (like managing DNS or creating certificates).

!!! note
    This page contains links to third party projects that provide functionality
    required for Gateway API to work. The Gateway API project authors aren't
    responsible for these projects, which are listed alphabetically within their
    class.

!!! info "Compare extended supported features across implementations"

    [View a table to quickly compare supported features of projects](implementations/v1.4.md). These outline Gateway controller implementations that have passed core conformance tests, and focus on extended conformance features that they have implemented. These tables will be generated and uploaded to the site once at least 3 implementations have uploaded their conformance reports under the [conformance reports](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports).

## Gateway Controller Implementation Status <a name="gateways"></a>

### Conformant
- [Agentgateway][40]
- [Airlock Microgateway][34]
- [Cilium][16]
- [Google Kubernetes Engine][6] (GA)
- [HAProxy Ingress][7]
- [Istio][9] (GA)
- [kgateway][37] (GA)
- [NGINX Gateway Fabric][12] (GA)
- [Traefik Proxy][13] (GA)

### Partially Conformant

- [Amazon Elastic Kubernetes Service][23] (GA)
- [AWS Load Balancer Controller][44] (GA)
- [Contour][3] (GA)
- [Envoy Gateway][18] (GA)
- [Gravitee Kubernetes Operator][42] (GA)
- [Kong Gateway Operator][35] (GA)

### Stale

- [Azure Application Gateway for Containers][27] (GA)
- [Gloo Gateway][5] (GA)
- [Kong Ingress Controller][10] (GA)
- [Kubvernor][39](work in progress)

## Service Mesh Implementation Status <a name="meshes"></a>

### Conformant

- [Istio][9] (GA)
- [Cilium][16] (GA)

### Partially Conformant

- [Alibaba Cloud Service Mesh][43] (GA)
- [Linkerd][28] (GA)

## Integrations <a name="integrations"></a>

- [Flagger][14] (public preview)
- [cert-manager][15] (alpha)
- [argo-rollouts][22] (alpha)
- [Knative][24] (alpha)
- [Kuadrant][26] (GA)
- [kruise-rollouts][41] (alpha)

[3]:#contour
[5]:#gloo-gateway
[6]:#google-kubernetes-engine
[7]:#haproxy-ingress
[9]:#istio
[10]:#kong-kubernetes-ingress-controller
[12]:#nginx-gateway-fabric
[13]:#traefik-proxy
[14]:#flagger
[15]:#cert-manager
[16]:#cilium
[18]:#envoy-gateway
[22]:#argo-rollouts
[23]:#amazon-elastic-kubernetes-service
[24]:#knative
[26]:#kuadrant
[27]:#azure-application-gateway-for-containers
[28]:#linkerd
[33]:#ngrok-kubernetes-operator
[34]:#airlock-microgateway
[35]:#kong-gateway-operator
[37]:#kgateway
[39]:#kubvernor
[40]:#agentgateway
[41]:#kruise-rollouts
[42]:#gravitee-kubernetes-operator
[43]:#alibaba-cloud-service-mesh
[44]:#aws-load-balancer-controller


[gamma]:mesh/index.md



## Implementations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### Agentgateway
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.0-Agentgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5.0/agentgateway-agentgateway)

[Agentgateway](https://agentgateway.dev/) is an open source Gateway API implementation hosted as a part of the Linux Foundation, focusing on AI use cases, including LLM consumption, LLM serving, agent-to-agent ([A2A](https://a2aproject.github.io/A2A/latest/)), agent-to-tool ([MCP](https://modelcontextprotocol.io/introduction)), as well as traditional TCP/HTTP traffic serving.
It is the first and only proxy designed specifically for the Kubernetes Gateway API, powered by a high performance and scalable Rust dataplane implementation.

### Airlock Microgateway
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-Airlock%20Microgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5.1/airlock-microgateway)

[Airlock Microgateway][airlock-microgateway] is a Kubernetes native WAAP (Web Application and API Protection) solution optimized for Kubernetes environments and certified for Red Hat OpenShift.
Modern application security is embedded in the development workflow and follows DevSecOps paradigms.
Airlock Microgateway protects your applications and microservices with the tried-and-tested Airlock security features against attacks, while also providing a high degree of scalability.

#### Features
- Comprehensive set of security features, including deny rules to protect against known attacks (OWASP Top 10), header filtering, JSON parsing, OpenAPI specification enforcement, GraphQL schema validation, and antivirus scanning with ICAP
- Identity aware proxy which makes it possible to enforce authentication using client certificate based authentication, JWT authentication or OIDC with step-up authentication to realize multi factor authentication (MFA). Provides OAuth 2.0 Token Introspection and Token Exchange for continuous validation and secure delegation across services
- Reverse proxy functionality with request routing rules, TLS termination, and remote IP extraction
- Easy-to-use Grafana dashboards which provide valuable insights in allowed and blocked traffic and other metrics

#### Documentation and links
- [Product documentation][airlock-microgateway-documentation]
- [Gateway specific documentation][airlock-microgateway-guide]
- Check our [Airlock community forum][airlock-microgateway-community-support] and [support process][airlock-microgateway-premium-support] for support.

[airlock-microgateway]:https://www.airlock.com/en/secure-access-hub/components/microgateway
[airlock-microgateway-documentation]:https://docs.airlock.com/microgateway/latest
[airlock-microgateway-guide]:https://docs.airlock.com/microgateway/latest/?topic=MGW-00000142
[airlock-microgateway-community-support]:https://forum.airlock.com/
[airlock-microgateway-premium-support]:https://techzone.ergon.ch/support-process

### Alibaba Cloud Service Mesh

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.3.0-Alibaba%20Cloud%20Service%20Mesh-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.3.0/alibaba-cloud-servicemesh)

[Alibaba Cloud Service Mesh (ASM)][asm] provides a fully managed service mesh platform that is compatible with the community Istio. It simplifies service governance, including traffic routing and split management between service calls, authentication security for inter-service communication, and mesh observability capabilities, thereby greatly reducing the workload of development and operations.

[asm]:https://www.alibabacloud.com/help/en/asm/product-overview/what-is-asm

### Amazon Elastic Kubernetes Service

[Amazon Elastic Kubernetes Service (EKS)][eks] is a managed service that you can use to run Kubernetes on AWS without needing to install, operate, and maintain your own Kubernetes control plane or nodes. EKS's implementation of the Gateway API is through [AWS Gateway API Controller][eks-gateway] which provisions [Amazon VPC Lattice][vpc-lattice] Resources for gateway(s), HTTPRoute(s) in EKS clusters.

[eks]:https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html
[eks-gateway]:https://github.com/aws/aws-application-networking-k8s
[vpc-lattice]:https://aws.amazon.com/vpc/lattice/

### AWS Load Balancer Controller

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Partial%20Conformance%20v1.3.0-AWS%20Load%20Balancer%20Controller-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.3.0/aws-load-balancer-controller)

[AWS Load Balancer Controller][aws-lbc] manages AWS Elastic Load Balancers for Kubernetes clusters. The controller provisions AWS Application Load Balancers (ALB) when you create a Kubernetes Ingress and AWS Network Load Balancers (NLB) when you create a Kubernetes Service of type LoadBalancer.

Gateway API support is GA for both Layer 4 (L4) and Layer 7 (L7) routing, enabling customers to provision and manage AWS NLBs and ALBs directly from Kubernetes clusters using the extensible Gateway API.

See the [AWS Load Balancer Controller documentation][aws-lbc-docs] for information on how to deploy and use the Gateway API implementation.

[aws-lbc]:https://github.com/kubernetes-sigs/aws-load-balancer-controller
[aws-lbc-docs]:https://kubernetes-sigs.github.io/aws-load-balancer-controller/

### Azure Application Gateway for Containers

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Azure%20Application%20Gateway%20for%20Containers-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/azure-application-gateway-for-containers)

[Application Gateway for Containers][azure-application-gateway-for-containers] is a managed application (layer 7) load balancing solution, providing dynamic traffic management capabilities for workloads running in a Kubernetes cluster in Azure. Follow the [quickstart guide][azure-application-gateway-for-containers-quickstart-controller] to deploy the ALB controller and get started with Gateway API.


[azure-application-gateway-for-containers]:https://aka.ms/appgwcontainers/docs
[azure-application-gateway-for-containers-quickstart-controller]:https://learn.microsoft.com/azure/application-gateway/for-containers/quickstart-deploy-application-gateway-for-containers-alb-controller

### Cilium

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.0-Cilium-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.0/cilium)

[Cilium][cilium] is an eBPF-based networking, observability and security
solution for Kubernetes and other networking environments. It includes [Cilium
Service Mesh][cilium-service-mesh], a highly efficient mesh data plane that can
be run in [sidecarless mode][cilium-sidecarless] to dramatically improve
performance, and avoid the operational complexity of sidecars. Cilium also
supports the sidecar proxy model, offering choice to users.
Cilium supports Gateway API, passing conformance for v1.4.0 as of Cilium 1.19

Cilium is open source and is a CNCF Graduated project.

If you have questions about Cilium Service Mesh the #service-mesh channel on
[Cilium Slack][cilium-slack] is a good place to start. For contributing to the development
effort, check out the #development channel or join our [weekly developer meeting][cilium-meeting].

[cilium]:https://cilium.io
[cilium-service-mesh]:https://docs.cilium.io/en/stable/gettingstarted/#service-mesh
[cilium-sidecarless]:https://isovalent.com/blog/post/cilium-service-mesh/
[cilium118blog]:https://isovalent.com/blog/post/cilium-1-18/#service-mesh-gateway-api
[cilium-slack]:https://slack.cilium.io
[cilium-meeting]:https://github.com/cilium/cilium#weekly-developer-meeting

### Contour

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.3.0-Contour-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.3.0/projectcontour-contour)

[Contour][contour] is a CNCF open source Envoy-based ingress controller for Kubernetes.

Contour [v1.33.0][contour-release] implements Gateway API v1.3.0.
All [Standard channel][contour-standard] v1 API group resources (GatewayClass, Gateway, HTTPRoute, ReferenceGrant), plus most v1alpha2 API group resources (TLSRoute, TCPRoute, GRPCRoute, ReferenceGrant, and BackendTLSPolicy) are supported.
Contour's implementation passes most core extended Gateway API conformance tests included in the v1.3.0 release.

See the [Contour Gateway API Guide][contour-guide] for information on how to deploy and use Contour's Gateway API implementation.

For help and support with Contour's implementation, [create an issue][contour-issue-new] or ask for help in the [#contour channel on Kubernetes slack][contour-slack].

[contour]:https://projectcontour.io
[contour-release]:https://github.com/projectcontour/contour/releases/tag/v1.33.0
[contour-standard]:concepts/versioning.md#release-channels
[contour-guide]:https://projectcontour.io/docs/1.33/guides/gateway-api/
[contour-issue-new]:https://github.com/projectcontour/contour/issues/new/choose
[contour-slack]:https://kubernetes.slack.com/archives/C8XRH2R4J

### Envoy Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.0-EnvoyGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.0/envoy-gateway)

[Envoy Gateway][eg-home] is an [Envoy][envoy-org] subproject for managing Envoy-based application gateways. The supported
APIs and fields of the Gateway API are outlined [here][eg-supported].
Use the [quickstart][eg-quickstart] to get Envoy Gateway running with Gateway API in a
few simple steps.

[eg-home]:https://gateway.envoyproxy.io/
[envoy-org]:https://github.com/envoyproxy
[eg-supported]:https://gateway.envoyproxy.io/docs/tasks/quickstart/
[eg-quickstart]:https://gateway.envoyproxy.io/docs/tasks/quickstart

### Gloo Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-GlooGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/gloo-gateway)
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Partial%20Conformance%20v1.1.0-GlooGateway-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/gloo-gateway)

[Gloo Gateway][gloogateway] by [Solo.io][solo] is a feature-rich, Kubernetes-native ingress controller and next-generation API gateway.
Gloo Gateway brings the full power and community support of Gateway API to its existing control-plane implementation.

The Gloo Gateway ingress controller passes all the core Gateway API conformance tests in the v1.1.0 release for the GATEWAY_HTTP conformance
profile except `HTTPRouteServiceTypes`.

[gloogateway]:https://docs.solo.io/gateway/latest/
[solo]:https://www.solo.io

### Google Cloud Service Mesh


[Google Kubernetes Engine (GKE)][gke] is a managed Kubernetes platform offered
by Google Cloud.

GKE's implementation of Gateway For Mesh (GAMMA) is through the [Cloud Service Mesh][cloud-service-mesh].


Google Cloud Service Mesh supports [Envoy-based sidecar mesh][envoy-sidecar-mesh] and [Proxyless-GRPC][proxyless-grpc-mesh] (using GRPCRoute).


[gke]:https://cloud.google.com/kubernetes-engine
[cloud-service-mesh]:https://cloud.google.com/products/service-mesh
[envoy-sidecar-mesh]:https://cloud.google.com/service-mesh/docs/gateway/set-up-envoy-mesh
[proxyless-grpc-mesh]:https://cloud.google.com/service-mesh/docs/gateway/proxyless-grpc-mesh

### Google Kubernetes Engine

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.0-Google_Kubernetes_Engine-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5.0/gke-gateway)

[Google Kubernetes Engine (GKE)][gke] is a managed Kubernetes platform offered
by Google Cloud. GKE's implementation of the Gateway API is through the [GKE
Gateway controller][gke-gateway] which provisions Google Cloud Load Balancers
for Pods in GKE clusters.

The GKE Gateway controller supports weighted traffic splitting, mirroring,
advanced routing, multi-cluster load balancing and more. See the docs to deploy
[private or public Gateways][gke-gateway-deploy] and also [multi-cluster
Gateways][gke-multi-cluster-gateway].

The GKE Gateway controller passes all the core Gateway API conformance tests in the
v1.5.0 release for the GATEWAY_HTTP conformance profile.

[gke]:https://cloud.google.com/kubernetes-engine
[gke-gateway]:https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api
[gke-gateway-deploy]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways
[gke-multi-cluster-gateway]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-multi-cluster-gateways

### HAProxy Ingress

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-HAProxy%20Ingress-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5.1/haproxy-ingress)

[HAProxy Ingress][h1] is a community driven ingress controller implementation for HAProxy.

HAProxy Ingress is a conformant Gateway API implementation since `v0.17`. It implements all core features from the standard channel, as well as TLSRoute and TCPRoute APIs from the experimental channel.

[h1]:https://haproxy-ingress.github.io/

### Istio

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.0-Istio-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.0/istio-istio)

[Istio][istio] is an open source [service mesh][istio-mesh] and gateway implementation.

A minimal install of Istio can be used to provide a fully compliant
implementation of the Kubernetes Gateway API for cluster ingress traffic
control. For service mesh users, Istio also fully supports the [GAMMA
initiative's][gamma] Gateway API [support for east-west traffic
management][gamma] within the mesh.

Much of Istio's documentation, including all of the [ingress tasks][istio-1] and several mesh-internal traffic management tasks, already includes parallel instructions for
configuring traffic using either the Gateway API or the Istio configuration API.
Check out the [Gateway API task][istio-2] for more information about the Gateway API implementation in Istio.

[istio]:https://istio.io
[istio-mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/
[istio-2]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### kgateway
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.0-kgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.0/kgateway)

The [kgateway] project is a feature-rich, Kubernetes-native ingress controller and next-generation API gateway.
It is focused on maintaining a great HTTP experience, extending features for advanced routing in scenarios such as AI and MCP gateways, and interoperating with a service mesh such as Istio in both ambient and sidecar modes.
This focus means that you can easily configure a set of Envoy instances that are reasonably distributed in a performant way across many north-south and east-west use cases.

Kgateway is generally available with its 2.0 release.

[kgateway]:https://kgateway.dev/docs

### Kong Kubernetes Ingress Controller

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Kong%20Ingress%20Controller-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/kong-kubernetes-ingress-controller)

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

The [Kong Kubernetes Ingress Controller (KIC)][kic] can be used to configure unmanaged Gateways. See the [Gateway API Guide][kong-gw-guide] for usage information.. See the [Gateway API Guide][kong-gw-guide] for usage information.

For help and support with Kong Kubernetes Ingress Controller please feel free to [create an issue][kic-issue-new] or a [discussion][kic-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-gw-guide]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/using-gateway-api/
[kic-issue-new]:https://github.com/Kong/kubernetes-ingress-controller/issues/new
[kic-disc-new]:https://github.com/Kong/kubernetes-ingress-controller/discussions/new
[kong-slack]:https://kubernetes.slack.com/archives/CDCA87FRD

### Kong Gateway Operator

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.0-Kong%20Gateway%20Operator-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.0/kong-gateway-operator)

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

The [Kong Gateway operator (KGO)][kgo] can be used to configure managed Gateways and orchestrate instances of [Kong Kubernetes Ingress Controllers](#kong-kubernetes-ingress-controller).

For help and support with Kong Gateway operator please feel free to [create an issue][kgo-issue-new] or a [discussion][kgo-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kgo]:https://docs.konghq.com/gateway-operator/latest/
[kgo-issue-new]:https://github.com/Kong/gateway-operator/issues/new
[kgo-disc-new]:https://github.com/Kong/gateway-operator/discussions/new


### Kubvernor
[Kubvernor][kubvernor] is an open-source, highly experimental implementation of API controller in Rust programming language. Currently, Kubvernor supports Envoy Proxy. The project aims to be as generic as possible so Kubvernor can be used to manage/deploy different gateways (Envoy, Nginx, HAProxy, etc.).

[kubvernor]:https://github.com/kubvernor/kubvernor

### Linkerd

[Linkerd][linkerd] is the first CNCF graduated [service mesh][linkerd-mesh].
It is the only major mesh not based on Envoy, instead relying on a
purpose-built Rust micro-proxy to bring security, observability, and
reliability to Kubernetes, without the complexity.

Linkerd 2.14 and later support the [GAMMA initiative's][gamma]
Gateway API [support for east-west traffic management][gamma] within the mesh.

[linkerd]:https://linkerd.io/
[linkerd-mesh]:https://buoyant.io/service-mesh-manifesto

### NGINX Gateway Fabric

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.1-NGINX Gateway Fabric-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.1/nginx-nginx-gateway-fabric)

[NGINX Gateway Fabric][nginx-gateway-fabric] is an open-source project that provides an implementation of the Gateway API using [NGINX][nginx] as the data plane. The goal of this project is to implement the core Gateway API to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. You can find the comprehensive NGINX Gateway Fabric user documentation on the [NGINX Documentation][nginx-docs] website.

For a list of supported Gateway API resources and features, see the [Gateway API Compatibility][nginx-compat] doc.

If you have any suggestions or experience issues with NGINX Gateway Fabric, please [create an issue][nginx-issue-new] or a [discussion][nginx-disc-new] on GitHub. You can also ask for help in the [NGINX Community Forum][nginx-forum].

[nginx-gateway-fabric]:https://github.com/nginx/nginx-gateway-fabric
[nginx]:https://nginx.org/
[nginx-docs]:https://docs.nginx.com/nginx-gateway-fabric/
[nginx-compat]:https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-api-compatibility/
[nginx-issue-new]:https://github.com/nginx/nginx-gateway-fabric/issues/new
[nginx-disc-new]:https://github.com/nginx/nginx-gateway-fabric/discussions/new
[nginx-forum]:https://community.nginx.org/


### Traefik Proxy

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-Traefik Proxy-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5.1/traefik-traefik)

[Traefik Proxy][traefik-proxy] is an open source cloud-native application proxy.

Traefik Proxy currently supports version `v1.5.1` of the Gateway API specification, check the [Kubernetes Gateway Provider Documentation][traefik-proxy-gateway-api-doc] for more information on how to deploy and use it.
Traefik Proxy fully supports all HTTPRoute core and some extended features, like BackendTLSPolicy, GRPCRoute, and TLSRoute resources from the Standard channel, as well as TCPRoute from the Experimental channel.

For help and support with Traefik Proxy, [create an issue][traefik-proxy-issue-new] or ask for help in the [Traefik Labs Community Forum][traefiklabs-community-forum].

[traefik-proxy]:https://traefik.io
[traefik-proxy-gateway-api-doc]:https://doc.traefik.io/traefik/v3.7/reference/install-configuration/providers/kubernetes/kubernetes-gateway
[traefik-proxy-issue-new]:https://github.com/traefik/traefik/issues/new/choose
[traefiklabs-community-forum]:https://community.traefik.io/c/traefik/traefik-v3/21

## Integrations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific integrations.

### Flagger

[Flagger][flagger] is a progressive delivery tool that automates the release process for applications running on Kubernetes.

Flagger can be used to automate canary deployments and A/B testing using Gateway API. It supports both the `v1alpha2` and `v1beta1` spec of Gateway API. You can refer to [this tutorial][flagger-tutorial] to use Flagger with any implementation of Gateway API.

[flagger]:https://flagger.app
[flagger-tutorial]:https://docs.flagger.app/tutorials/gatewayapi-progressive-delivery

### cert-manager

[cert-manager][cert-manager] is a tool to automate certificate management in cloud native environments.

cert-manager can generate TLS certificates for Gateway resources. This is configured by adding annotations to a Gateway. It currently supports the `v1` spec of Gateway API. You can refer to the [cert-manager docs][cert-manager-docs] to try it out.

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

[Kuadrant][kuadrant] is an open source multi cluster Gateway API controller that integrates with and provides policies via policy attachment to other Gateway API providers.

Kuadrant supports Gateway API for defining gateways centrally and attaching policies such as DNS, TLS, Auth and Rate Limiting that apply to all of your Gateways.

Kuadrant works with both Istio and Envoy Gateway as underlying Gateway API providers, with plans to work with other gateway providers in future.

For help and support with Kuadrant's implementation please feel free to [create an issue][kuadrant-issue-new] or ask for help in the [#kuadrant channel on Kubernetes slack][kuadrant-slack].

[kuadrant]:https://kuadrant.io/
[kuadrant-issue-new]:https://github.com/Kuadrant/kuadrant-operator/issues/new
[kuadrant-slack]:https://kubernetes.slack.com/archives/C05J0D0V525

### OpenKruise Rollouts
[OpenKruise Rollouts][kruise-rollouts] is a plugin-n-play progressive delivery controller for Kubernetes. It supports several advanced deployment methods such as blue/green and canaries. OpenKruise Rollouts has built-in support for the Gateway API.

[kruise-rollouts]:https://openkruise.io/rollouts/introduction

## Adding new entries

Implementations are free to make a PR to add their entry to this page; however,
in order to meet the requirements for being Partially Conformant or Conformant,
the implementation must have had a conformance report submission PR merged.

Part of the review process for new additions to this page is that a maintainer
will check the conformance level and verify the state.

## Page Review Policy

This page is intended to showcase actively developed and conformant implementations
of Gateway API, and so is subject to regular reviews.

These reviews are performed at least one month after every Gateway API release
(starting with the Gateway API v1.3 release).

As part of the review, a maintainer will check:

* which implementations are **Conformant** - as defined above in this document.
* which implementations are **Partially Conformant**, as defined above in this
  document.

If the maintainer performing the review finds that there are implementations
that no longer satisfy the criteria for Partially Conformant or Conformant, or
finds implementations that are in the "Stale" state, then that maintainer will:

* Inform the other maintainers and get their agreement on the list of stale and
to-be-removed implementations
* Open a draft PR with the changes to this page.
* Post on the #sig-network-gateway-api channel informing the maintainers of
implementations that are no longer at least partially conformant should contact
the Gateway API maintainers to discuss the implementation's status. This period
is called the "**right-of-reply**" period, is at least two weeks long, and functions
as a lazy consensus period.
* Any implementations that do not respond within the right-of-reply period will be
downgraded in status, either by being moved to "Stale", or being removed
from this page if they are already "Stale".

Page review timeline, starting with the v1.4 Page Review:

* Gateway API v1.4 release Page Review (at least one month after the actual
  release): a maintainer will move anyone who hasn't submitted a conformance
  report using the rules above to "Stale". They will also contact anyone who
  moves to Stale to inform them about this rule change.
* Gateway API v1.5 release Page Review (at least one month after the actual
  release): A maintainer will perform the Page Review process again, removing
  any implementations that are still Stale (after a right-of-reply period).
  **You are here**
* Gateway API v1.6 release Page Review (at least one month after the actual
  release): We will remove the Stale category, and implementation maintainers
  will need to be at least partially conformant on each review, or during the
  right-of-reply period, or be removed from the implementations page.

This means that, after the Gateway API v1.6 release, implementations cannot be
added to this page unless they have submitted at least a Partially Conformant
conformance report.
