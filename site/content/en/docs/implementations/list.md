---
title: "Implementations"
linkTitle: "List"
weight: 1
---

This document tracks downstream implementations and integrations of Gateway API
and provides status and resource references for them.

Implementors and integrators of Gateway API are encouraged to update this
document with status information about their implementations, the versions they
cover, and documentation to help users get started. This status information should
be no longer than a few paragraphs.

## Conformance levels

There are two levels of Gateway API conformance:

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

## Implementation profiles

Implementations may also support two types of traffic:

* **Gateway** controllers reconcile the Gateway resource and are intended to
handle north-south traffic, mainly concerned with coming from outside the
cluster to inside.
* **Mesh** controllers reconcile Service resources with HTTPRoutes attached
and are intended to handle east-west traffic, within the same cluster or
set of clusters.

Each parent resource has a set of conformance tests associated with it, that lay out
the expected behavior for implementations to be conformant (as above).

Implementations may also handle both parent resources.

## Integrations

Also listed on this page are **integrations**, which are other software
projects that are able to make use of Gateway API resources to perform
other functions (like managing DNS or creating certificates).

{{% alert color="primary" %}}
This page contains links to third party projects that provide functionality
required for Gateway API to work. The Gateway API project authors aren't
responsible for these projects, which are listed alphabetically within their
class.

{{% /alert %}}

{{% alert color="info" title="Compare extended supported features across implementations" %}}

[View a table to quickly compare supported features of projects](/docs/implementations/versions/v1.6/). These outline Gateway controller implementations that have passed core conformance tests, and focus on extended conformance features that they have implemented. These tables will be generated and uploaded to the site once at least 3 implementations have uploaded their conformance reports under the [conformance reports](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports).

{{% /alert %}}

## Gateway Controller Implementation Status <a name="gateways"></a>

### Conformant
- [Agentgateway](#agentgateway)
- [Airlock Microgateway](#airlock-microgateway)
- [Cilium](#cilium)
- [Envoy Gateway](#envoy-gateway)
- [Google Kubernetes Engine](#google-kubernetes-engine)
- [Gravitee Kubernetes Operator](#gravitee-kubernetes-operator)
- [HAProxy Ingress](#haproxy-ingress)
- [Higress](#higress)
- [Istio](#istio)
- [Kong Operator](#kong-operator)
- [NGINX Gateway Fabric](#nginx-gateway-fabric)
- [Sunbeam Proxy](#sunbeam-proxy)
- [Traefik Proxy](#traefik-proxy)
- [Varnish Gateway](#varnish-gateway)
- [WSO2 Gateway](#wso2-gateway)
- [kgateway](#kgateway)


### Partially Conformant
- [AWS Load Balancer Controller](#aws-load-balancer-controller)
- [Amazon Elastic Kubernetes Service](#amazon-elastic-kubernetes-service)
- [Calico](#calico)
- [Gloo Gateway](#gloo-gateway)


## Service Mesh Implementation Status <a name="meshes"></a>

### Conformant
- [Cilium](#cilium)
- [Istio](#istio)




## Integrations <a name="integrations"></a>


- [Argo Rollouts](#argo-rollouts)
- [cert-manager](#cert-manager)
- [Flagger](#flagger)
- [Knative](#knative)
- [OpenKruise Rollouts](#openkruise-rollouts)


[gamma]: /docs/mesh/



## Implementations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.



### AWS Load Balancer Controller

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-AWS+Load+Balancer+Controller-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/aws-load-balancer-controller/conformance-report.yaml)

[AWS Load Balancer Controller][aws-lbc] manages AWS Elastic Load Balancers for Kubernetes clusters. The controller provisions AWS Application Load Balancers (ALB) when you create a Kubernetes Ingress and AWS Network Load Balancers (NLB) when you create a Kubernetes Service of type LoadBalancer.

Gateway API support is GA for both Layer 4 (L4) and Layer 7 (L7) routing, enabling customers to provision and manage AWS NLBs and ALBs directly from Kubernetes clusters using the extensible Gateway API.

See the [AWS Load Balancer Controller documentation][aws-lbc-docs] for information on how to deploy and use the Gateway API implementation.

[aws-lbc]:https://github.com/kubernetes-sigs/aws-load-balancer-controller
[aws-lbc-docs]:https://kubernetes-sigs.github.io/aws-load-balancer-controller/


### Agentgateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.0-Agentgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/agentgateway-agentgateway/v1.3.1-report.yaml)

[Agentgateway](https://agentgateway.dev/) is an open source Gateway API implementation hosted as a part of the Linux Foundation, focusing on AI use cases, including LLM consumption, LLM serving, agent-to-agent ([A2A](https://a2aproject.github.io/A2A/latest/)), agent-to-tool ([MCP](https://modelcontextprotocol.io/introduction)), as well as traditional TCP/HTTP traffic serving.
It is the first and only proxy designed specifically for the Kubernetes Gateway API, powered by a high performance and scalable Rust dataplane implementation.


### Airlock Microgateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.0-Airlock+Microgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/airlock-microgateway/experimental-5.1.0-default-report.yaml)

[Airlock Microgateway][airlock-microgateway] is a Kubernetes-native security solution that extends the routing capabilities of the Gateway API with WAAP (Web Application and API Protection) features and identity-aware proxying.
It filters requests using deny rules against known attacks (OWASP Top 10), along with header filtering, JSON parsing, OpenAPI specification enforcement, GraphQL schema validation, and antivirus scanning via ICAP.
Authentication can be enforced via client certificates, JWT, or OIDC with step-up authentication for MFA, with support for OAuth 2.0 Token Introspection and Token Exchange. Airlock Microgateway is certified for Red Hat OpenShift, and built-in Grafana dashboards provide real-time reporting on system health, traffic and threats.


### Amazon Elastic Kubernetes Service

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.0-Amazon+Elastic+Kubernetes+Service-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.0/aws-aws-application-networking-k8s/experimental-v2.0.1-default-report.yaml)

[Amazon Elastic Kubernetes Service (EKS)][eks] is a managed service that you can use to run Kubernetes on AWS without needing to install, operate, and maintain your own Kubernetes control plane or nodes. EKS's implementation of the Gateway API is through [AWS Gateway API Controller][eks-gateway] which provisions [Amazon VPC Lattice][vpc-lattice] Resources for gateway(s), HTTPRoute(s) in EKS clusters.

[eks]:https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html
[eks-gateway]:https://github.com/aws/aws-application-networking-k8s
[vpc-lattice]:https://aws.amazon.com/vpc/lattice/


### Calico

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.1-Calico-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.1/projectcalico-calico/experimental-v3.32.0-default-report.yaml)

[Calico][calico] is an open-source networking and security solution for
Kubernetes and other cloud-native environments. Calico's Gateway API
implementation is built on the [tigera-operator][calico-operator] and
[Envoy Gateway][envoy-gateway]: the operator reconciles a `GatewayAPI`
custom resource, provisions an Envoy Gateway control plane, and creates
a default `tigera-gateway-class` GatewayClass on the cluster.

Questions and contributions are welcome on [GitHub][calico]. For
maintainer inquiries, see the [Tigera contact form][tigera-contact].

[calico]: https://github.com/projectcalico/calico
[calico-operator]: https://github.com/tigera/operator
[envoy-gateway]: https://gateway.envoyproxy.io
[tigera-contact]: https://www.tigera.io/contact/


### Cilium

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-Cilium-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/cilium/experimental-v1.20.0-default-report.yaml)

Cilium][cilium] is an eBPF-based networking, observability and security
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

[cilium]: https://cilium.io
[cilium-service-mesh]: https://docs.cilium.io/en/stable/gettingstarted/#service-mesh
[cilium-sidecarless]: https://isovalent.com/blog/post/cilium-service-mesh/
[cilium118blog]: https://isovalent.com/blog/post/cilium-1-18/#service-mesh-gateway-api
[cilium-slack]: https://slack.cilium.io
[cilium-meeting]: https://github.com/cilium/cilium#weekly-developer-meeting


### Envoy Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-Envoy+Gateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/envoy-gateway/experimental-v1.9.0-default-report.yaml)

[Envoy Gateway][eg-home] is an [Envoy][envoy-org] subproject for managing Envoy-based application gateways. The supported
APIs and fields of the Gateway API are outlined [here][eg-supported].
Use the [quickstart][eg-quickstart] to get Envoy Gateway running with Gateway API in a
few simple steps.

[eg-home]:https://gateway.envoyproxy.io/
[envoy-org]:https://github.com/envoyproxy
[eg-supported]:https://gateway.envoyproxy.io/docs/tasks/quickstart/
[eg-quickstart]:https://gateway.envoyproxy.io/docs/tasks/quickstart


### Gloo Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.4.1-Gloo+Gateway-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.4.1/gloo-gateway/v1.21.3-report.yaml)

[Gloo Gateway][gloogateway] by [Solo.io][solo] is a feature-rich, Kubernetes-native ingress controller and next-generation API gateway.
Gloo Gateway brings the full power and community support of Gateway API to its existing control-plane implementation.

[gloogateway]: https://docs.solo.io/gateway/latest/
[solo]: https://www.solo.io


### Google Kubernetes Engine

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.0-Google+Kubernetes+Engine-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/gke-gateway/v1.6.0-gke-report.yaml)

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


### Gravitee Kubernetes Operator

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-Gravitee+Kubernetes+Operator-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/gravitee/standard-4.12.10-default-report.yaml)

The [Gravitee Kubernetes Operator](https://documentation.gravitee.io/gravitee-kubernetes-operator-gko) (GKO) lets you manage [Gravitee](https://www.gravitee.io/) APIs, applications, and other assets in a Kubernetes-native and declarative way.

For support, feedback, or to engage in a discussion about the Gravitee Kubernetes Operator, please feel free to submit an [issue](https://github.com/gravitee-io/issues/issues) or visit our community [forum](https://community.gravitee.io/c/support/gravitee-kubernetes-operator-gko/26).


### HAProxy Ingress

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-HAProxy+Ingress-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5/haproxy-ingress/experimental-v0.17.0-alpha.1-default-report.yaml)

[HAProxy Ingress][h1] is a community driven ingress controller implementation for HAProxy.

HAProxy Ingress is a conformant Gateway API implementation since `v0.17`. It implements all core features from the standard channel, as well as TLSRoute and TCPRoute APIs from the experimental channel.

[h1]:https://haproxy-ingress.github.io/


### Higress

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.0-Higress-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/higress-group-higress/standard-v2.2.4-default-report.yaml)

[Higress](https://higress.ai/) is a cloud-native API gateway built on Istio and
Envoy. It provides Kubernetes Gateway API and Ingress support alongside API
management and AI gateway capabilities. Higress is an open source
[CNCF project](https://www.cncf.io/projects/higress/).

Source code, documentation, and issue tracking are available in the
[Higress repository](https://github.com/higress-group/higress).


### Istio

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-Istio-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5/istio-istio/1.30.0-default-report.yaml)

[Istio][istio] is an open source [service mesh][istio-mesh] and gateway implementation.

A minimal install of Istio can be used to provide a fully compliant
implementation of the Kubernetes Gateway API for cluster ingress traffic
control. For service mesh users, Istio also fully supports the [GAMMA
initiative's][gamma] Gateway API [support for east-west traffic
management][gamma] within the mesh.

Much of Istio's documentation, including all of the [ingress tasks][istio-1] and several mesh-internal traffic management tasks, already includes parallel instructions for
configuring traffic using either the Gateway API or the Istio configuration API.
Check out the [Gateway API task][istio-2] for more information about the Gateway API implementation in Istio.

[istio]: https://istio.io
[istio-mesh]: https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]: https://istio.io/latest/docs/tasks/traffic-management/ingress/
[istio-2]: https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/


### Kong Operator

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-Kong+Operator-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/kong-operator/experimental-v2.3.0-rc.3-expressions-standard-report.yaml)

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

The [Kong Operator][kong-operator-docs] can be used to configure managed Gateways and orchestrate instances of Kong Kubernetes Ingress Controllers.

For help and support with Kong Operator please feel free to [create an issue][kong-operator-issue-new] or a [discussion][kong-operator-disc-new]. You can also ask for help in the [#kong channel on Kubernetes slack][kong-slack].

[kong-operator-docs]: https://developer.konghq.com/operator/
[kong-operator-issue-new]: https://github.com/Kong/kong-operator/issues/new
[kong-operator-disc-new]: https://github.com/Kong/kong-operator/discussions/new
[kong]: https://konghq.com
[kong-slack]: https://kubernetes.slack.com/archives/CDCA87FRD


### NGINX Gateway Fabric

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-NGINX+Gateway+Fabric-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/nginx-nginx-gateway-fabric/experimental-2.7.0-pre-release-default-report.yaml)

[NGINX Gateway Fabric][nginx-gateway-fabric] is an open-source project that provides an implementation of the Gateway API using [NGINX][nginx] as the data plane. The goal of this project is to implement the core Gateway API to configure an HTTP or TCP/UDP load balancer, reverse-proxy, or API gateway for applications running on Kubernetes. You can find the comprehensive NGINX Gateway Fabric user documentation on the [NGINX Documentation][nginx-docs] website.

For a list of supported Gateway API resources and features, see the [Gateway API Compatibility][nginx-compat] doc.

If you have any suggestions or experience issues with NGINX Gateway Fabric, please [create an issue][nginx-issue-new] or a [discussion][nginx-disc-new] on GitHub. You can also ask for help in the [NGINX Community Forum][nginx-forum].

[nginx-gateway-fabric]: https://github.com/nginx/nginx-gateway-fabric
[nginx]: https://nginx.org/
[nginx-docs]: https://docs.nginx.com/nginx-gateway-fabric/
[nginx-compat]: https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-api-compatibility/
[nginx-issue-new]: https://github.com/nginx/nginx-gateway-fabric/issues/new
[nginx-disc-new]: https://github.com/nginx/nginx-gateway-fabric/discussions/new
[nginx-forum]: https://community.nginx.org/


### Sunbeam Proxy

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-Sunbeam+Proxy-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5/sunbeam-studios-sunbeam-proxy/experimental-v0.2.0-default-report.yaml)

[Sunbeam Proxy][sunbeam-proxy] is a cloud-native reverse proxy with adaptive machine learning threat detection which implements the Kubernetes Gateway API control plane and data plane.

[sunbeam-proxy]: https://github.com/sunbeamdotpt/proxy


### Traefik Proxy

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-Traefik+Proxy-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/traefik-traefik/experimental-v3.7.10-default-report.yaml)

Traefik Proxy supports the Gateway API specification, check the [Kubernetes Gateway Provider Documentation][traefik-proxy-gateway-api-doc] for more information on how to deploy and use it.

For help and support with Traefik Proxy, [create an issue][traefik-proxy-issue-new] or ask for help in the [Traefik Labs Community Forum][traefiklabs-community-forum].

[traefik-proxy]:https://traefik.io
[traefik-proxy-gateway-api-doc]:https://doc.traefik.io/traefik/v3.7/reference/install-configuration/providers/kubernetes/kubernetes-gateway
[traefik-proxy-issue-new]:https://github.com/traefik/traefik/issues/new/choose
[traefiklabs-community-forum]:https://community.traefik.io/c/traefik/traefik-v3/21


### Varnish Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.0-Varnish+Gateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5/varnish-software-varnish-gateway/standard-v0.20.0-default-report.yaml)

[Varnish Gateway][varnish-gateway] is an open source Kubernetes Gateway API implementation
developed by [Varnish Software AS][varnish-software], using [Varnish][varnish] as its data plane.

In addition to Gateway API resources, Varnish Gateway exposes a `VarnishCachePolicy` policy
attachment for fine-grained caching control (TTL, grace, request coalescing, cache key
customization, bypass conditions) at the Gateway, HTTPRoute, or rule level.

[varnish-gateway]:https://gateway.varnish.org
[varnish-software]:https://www.varnish-software.com/
[varnish]:https://www.varnish.org/


### WSO2 Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.5.1-WSO2+Gateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.5/wso2-gateway/v1.2.0-alpha2-report.yaml)

[WSO2 Gateway](https://wso2.com/api-platform/docs/) is an AI-ready, GitOps-driven platform for building, securing, and governing APIs across cloud, hybrid, and on-premises environments. It represents the next generation of the [WSO2 Kubernetes Gateway (APK)](https://apk.docs.wso2.com/en/latest/), delivering a unified Kubernetes-native gateway experience with deeper platform integration and modern GitOps workflows.

For supported Gateway API resources and features, see the [Gateway API support guide](https://wso2.com/api-platform/docs/api-gateway/next/deployment/deployment-modes/kubernetes/gateway-operator/#kubernetes-gateway-api-path). For questions and contributions, visit [GitHub](https://github.com/wso2/api-platform).


### kgateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.6.1-kgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.6/kgateway/v2.4.0-rc.1-report.yaml)

The [kgateway] project is a feature-rich, Kubernetes-native ingress controller and next-generation API gateway.
It is focused on maintaining a great HTTP experience, extending features for advanced routing in scenarios such as AI and MCP gateways, and interoperating with a service mesh such as Istio in both ambient and sidecar modes.
This focus means that you can easily configure a set of Envoy instances that are reasonably distributed in a performant way across many north-south and east-west use cases.

Kgateway is generally available with its 2.0 release.

[kgateway]:https://kgateway.dev/docs




## Integrations

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific integrations.



### Argo Rollouts

[Argo Rollouts][argo-rollouts] is a progressive delivery controller for Kubernetes. It supports several advanced deployment methods such as blue/green and canaries. Argo Rollouts supports the Gateway API via [a plugin][argo-rollouts-plugin].

[argo-rollouts]: https://argo-rollouts.readthedocs.io/en/stable/
[argo-rollouts-plugin]: https://github.com/argoproj-labs/rollouts-gatewayapi-trafficrouter-plugin/



### cert-manager

[cert-manager][cert-manager] is a tool to automate certificate management in cloud native environments.

cert-manager can generate TLS certificates for Gateway resources. This is configured by adding annotations to a Gateway. It currently supports the `v1` spec of Gateway API. You can refer to the [cert-manager docs][cert-manager-docs] to try it out.

[cert-manager]: https://cert-manager.io/
[cert-manager-docs]: https://cert-manager.io/docs/usage/gateway/



### Flagger

[Flagger][flagger] is a progressive delivery tool that automates the release process for applications running on Kubernetes.

Flagger can be used to automate canary deployments and A/B testing using Gateway API. You can refer to [this tutorial][flagger-tutorial] to use Flagger with any implementation of Gateway API.

[flagger]: https://flagger.app
[flagger-tutorial]: https://docs.flagger.app/tutorials/gatewayapi-progressive-delivery



### Knative

[Knative][knative] is a serverless platform built on Kubernetes.  Knative Serving provides a simple API for running stateless containers with automatic management of URLs, traffic splitting between revisions, request-based autoscaling (including scale to zero), and automatic TLS provisioning.  Knative Serving supports multiple HTTP routers through a plugin architecture, including a [gateway API plugin][knative-net-gateway-api] which is currently in alpha as not all Knative features are supported.

[knative]: https://knative.dev/
[knative-net-gateway-api]: https://github.com/knative-sandbox/net-gateway-api



### OpenKruise Rollouts

[OpenKruise Rollouts][kruise-rollouts] is a plugin-n-play progressive delivery controller for Kubernetes. It supports several advanced deployment methods such as blue/green and canaries. OpenKruise Rollouts has built-in support for the Gateway API.

[kruise-rollouts]: https://openkruise.io/rollouts/introduction





## Adding new entries

This page is automatically generated; please do not make PRs to this page.

### For implementations

Implementations wanting to add themselves must:

* Add a conformance report that's at least partially conformant to `conformance/reports`.
* Add an ImplementationDetails YAML file to `conformance/list/implementations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make generate` in the top level of the repository, and
the implementations list generation code will update the page for you. Include
the updated page in your PR.

This process replaces an older, mainatainer-performed process.

### For integrations

Integrations wanting to add themselves must:

* Add an ImplementationDetails YAML file to `conformance/list/integrations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make generate` in the top level of the repository, and
the implementations list generation code will update the page for you. Include
the updated page in your PR.
