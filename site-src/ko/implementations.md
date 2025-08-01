# 구현체

이 문서는 게이트웨이 API의 다운스트림 구현 및 통합을 추적하고
이들에 대한 상태 및 리소스 참조를 제공한다.

게이트웨이 API의 구현자와 통합자들은 자신들의 구현체에 대한 상태 정보, 지원하는 버전, 그리고
사용자가 시작하는 데 도움이 되는 문서를, 이 문서에 업데이트하는 것이
권장된다.


!!! info "구현체 간 확장 지원 기능 비교"

    [프로젝트의 지원 기능을 빠르게 비교할 수 있는 표 확인](implementations/v1.3.md). 표는 코어 호환성 테스트를 통과한 게이트웨이 컨트롤러 구현체를 개요로 제시하며,  구현한 확장 호환성 기능에 초점을 맞춘다.

## 게이트웨이 컨트롤러 구현 상태 <a name="gateways"></a>

- [Acnodal EPIC][1]
- [Airlock Microgateway][34]
- [Amazon Elastic Kubernetes Service][23] (GA)
- [Apache APISIX][2] (beta)
- [Avi Kubernetes Operator][31]
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
- [kgateway][37] (GA)
- [Kong Ingress Controller][10] (GA)
- [Kong Gateway Operator][35] (GA)
* [Kubvernor][39](진행 중)
- [Kuma][11] (GA)
- [LiteSpeed Ingress Controller][19]
- [LoxiLB][36] (beta)
- [NGINX Gateway Fabric][12] (GA)
- [ngrok][33] (preview)
- [STUNner][21] (beta)
- [Traefik Proxy][13] (GA)
- [Tyk][29] (진행 중)
- [WSO2 APK][25] (GA)

## 서비스 메시 구현 상태 <a name="meshes"></a>

- [Google Cloud Service Mesh][38] (GA)
- [Istio][9] (GA)
- [Kuma][11] (GA)
- [Linkerd][28] (GA)

## 통합 <a name="integrations"></a>

- [Flagger][14] (public preview)
- [cert-manager][15] (alpha)
- [argo-rollouts][22] (alpha)
- [Knative][24] (alpha)
- [Kuadrant][26] (GA)

[1]:#acnodal-epic
[2]:#apisix
[3]:#contour
[4]:#emissary-ingress-ambassador-api-gateway
[5]:#gloo-gateway
[6]:#google-kubernetes-engine
[7]:#haproxy-ingress
[8]:#hashicorp-consul
[9]:#istio
[10]:#kong-kubernetes-ingress-controller
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
[34]:#airlock-microgateway
[35]:#kong-gateway-operator
[36]:#loxilb
[37]:#kgateway
[38]:#google-cloud-service-mesh
[39]:#kubvernor


[gamma]:mesh/index.md



## 구현체

이 섹션에서는 특정 구현체들에 대한 블로그 게시글, 문서 및 기타 게이트웨이 API 참조에 대한 구체적인 링크를 찾을 수 있다.

### Acnodal EPIC
[EPIC][epicdocs]는 쿠버네티스와 함께 설계되고 구축된 오픈 소스 외부 게이트웨이 플랫폼이다. 게이트웨이 클러스터, k8s 게이트웨이 컨트롤러, 독립형 리눅스 게이트웨이 컨트롤러 및 게이트웨이 서비스 매니저로 구성된다. 이들은 함께 클러스터 사용자에게 게이트웨이 서비스를 제공하는 플랫폼을 만든다. 각 게이트웨이는 워크로드 클러스터가 아닌 게이트웨이 클러스터에서 실행되는 여러 Envoy 인스턴스로 구성된다. 게이트웨이 서비스 매니저는 공용 및 사설 클러스터를 위한 Gateway-as-a-Service 인프라를 구현하고 비-k8s 엔드포인트를 통합하는 데 사용할 수 있는 간단한 사용자 관리 및 UI이다.

- [문서][epicdocs]
- [소스 저장소][epicsource]

[epicdocs]:https://www.epic-gateway.org/
[epicsource]:https://github.com/epic-gateway

### Airlock Microgateway
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.3.0-Airlock%20Microgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.3.0/airlock-microgateway)

[Airlock Microgateway][airlock-microgateway]는 쿠버네티스 환경에 최적화되고 RedHat OpenShift 인증을 받은 쿠버네티스 네이티브 WAAP(Web Application and API Protection, 이전 WAF) 솔루션이다.
현대적인 애플리케이션 보안이 개발 워크플로에 내장되어 DevSecOps 패러다임을 따른다.
Airlock Microgateway는 검증된 Airlock 보안 기능으로 애플리케이션과 마이크로서비스를 공격으로부터 보호하며, 높은 확장성도 제공한다.

#### 기능
- 알려진 공격(OWASP Top 10)으로부터 보호하는 거부 규칙, 헤더 필터링, JSON 파싱, OpenAPI 명세 강제 적용, GraphQL 스키마 검증과 같은 보안 기능을 갖춘 포괄적인 WAAP(이전 WAF)
- JWT 인증 또는 OIDC를 사용한 인증 강제를 가능하게 하는 ID 인식 프록시
- 요청 라우팅 규칙, TLS 종료 및 원격 IP 추출을 포함한 리버스 프록시 기능
- 허용 및 차단된 트래픽과 기타 메트릭에 대한 유용한 인사이트를 제공하는 사용하기 쉬운 Grafana 대시보드

#### 문서 및 링크
- [제품 문서][airlock-microgateway-documentation]
- [게이트웨이 상세 문서][airlock-microgateway-guide]
- 도움을 위해 [Airlock 커뮤니티 포럼][airlock-microgateway-community-support]과 [지원 프로세스][airlock-microgateway-premium-support]를 확인하자.

[airlock-microgateway]:https://www.airlock.com/en/secure-access-hub/components/microgateway
[airlock-microgateway-documentation]:https://docs.airlock.com/microgateway/latest
[airlock-microgateway-guide]:https://docs.airlock.com/microgateway/latest/?topic=MGW-00000142
[airlock-microgateway-community-support]:https://forum.airlock.com/
[airlock-microgateway-premium-support]:https://techzone.ergon.ch/support-process

### Amazon Elastic Kubernetes Service

[Amazon Elastic Kubernetes Service (EKS)][eks]는 자체 쿠버네티스 컨트롤 플레인이나 노드를 설치, 운영, 유지 관리할 필요 없이 AWS에서 쿠버네티스를 실행하는 데 사용할 수 있는 관리형 서비스이다. EKS는 [AWS 게이트웨이 API 컨트롤러][eks-gateway]를 통해 게이트웨이 API를 구현하며, 이 컨트롤러는 EKS 클러스터 내 게이트웨이 및 HTTPRoute를 위해 [Amazon VPC Lattice][vpc-lattice] 리소스를 프로비저닝한다.

[eks]:https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html
[eks-gateway]:https://github.com/aws/aws-application-networking-k8s
[vpc-lattice]:https://aws.amazon.com/vpc/lattice/

### APISIX

[Apache APISIX][apisix]는 동적이고 실시간이며 고성능인 API 게이트웨이이다. APISIX는 로드 밸런싱, 동적 업스트림, 카나리 릴리스, 서킷 브레이킹, 인증, 관찰 가능성 등과 같은 풍부한 트래픽 관리 기능을 제공한다.

APISIX는 현재 [Apache APISIX 인그레스 컨트롤러][apisix-1]에 대해 게이트웨이 API `v1beta1` 버전의 명세를 지원한다.

[apisix]:https://apisix.apache.org/
[apisix-1]:https://github.com/apache/apisix-ingress-controller

### Avi Kubernetes Operator

[Avi Kubernetes Operator (AKO)][ako]는 VMware AVI Advanced Load Balancer를 사용하여 L4-L7 로드 밸런싱을 제공한다.

AKO 버전 [v1.13.1]부터 게이트웨이 API 버전 v1.0.0이 지원된다. 게이트웨이 클래스, 게이트웨이 및 HTTPRoute 객체를 지원하는 게이트웨이 API 명세의 v1 버전을 구현한다.

AKO 게이트웨이 API를 배포하고 사용하는 문서는 [Avi 쿠버네티스 오퍼레이터 게이트웨이 API][ako-gw]에서 찾을 수 있다.

[ako]:https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-kubernetes-operator/AKO/avi-kubernetes-operator-1-13/avi-kubernetes-operator.html
[ako-gw]:https://techdocs.broadcom.com/us/en/vmware-security-load-balancing/avi-load-balancer/avi-kubernetes-operator/AKO/avi-kubernetes-operator-1-13/gateway-api/gateway-api-v1.html
[v1.13.1]:https://github.com/vmware/load-balancer-and-ingress-services-for-kubernetes

### Azure Application Gateway for Containers

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Partial%20Conformance%20v1.1.1-Azure%20Application%20Gateway%20for%20Containers-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/azure-application-gateway-for-containers)

[Application Gateway for Containers][azure-application-gateway-for-containers]는 Azure의 쿠버네티스 클러스터에서 실행되는 워크로드에 대한 동적 트래픽 관리 기능을 제공하는 관리형 애플리케이션(레이어 7) 로드 밸런싱 솔루션이다. ALB 컨트롤러를 배포하고 게이트웨이 API를 시작하려면 [빠른 시작 가이드][azure-application-gateway-for-containers-quickstart-controller]를 따른다.


[azure-application-gateway-for-containers]:https://aka.ms/appgwcontainers/docs
[azure-application-gateway-for-containers-quickstart-controller]:https://learn.microsoft.com/azure/application-gateway/for-containers/quickstart-deploy-application-gateway-for-containers-alb-controller

### Cilium

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-Cilium-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/cilium)

[Cilium][cilium]은 쿠버네티스 및 기타 네트워킹 환경을 위한 eBPF 기반 네트워킹, 관찰 가능성 및 보안 솔루션이다.
여기에는 [Cilium Service Mesh][cilium-service-mesh]가 포함되어 있으며,
이는 높은 효율을 가진 메시 데이터 플레인으로 [사이드카 없는 모드][cilium-sidecarless]에서 실행될 수 있어 성능을 크게 향상시키고,
사이드카로 인한 운영 복잡성을 피할 수 있다.
Cilium은 또한 사이드카 프록시 모델도 지원하여 사용자에게 선택권을 제공한다.
[Cilium 1.14][cilium114blog]부터 Cilium은 게이트웨이 API를 지원하며 v0.7.1에 대한 호환성을
통과한다.

Cilium은 오픈 소스이며 CNCF 졸업 프로젝트이다.

Cilium 서비스 메시에 대한 질문이 있다면 [Cilium Slack][cilium-slack]의 #service-mesh 채널에서 시작하는 것이 좋다.
개발 노력에 기여하려면 #development 채널을 확인하거나,
[주간 개발자 회의][cilium-meeting]에 참여하자.

[cilium]:https://cilium.io
[cilium-service-mesh]:https://docs.cilium.io/en/stable/gettingstarted/#service-mesh
[cilium-sidecarless]:https://isovalent.com/blog/post/cilium-service-mesh/
[cilium114blog]:https://isovalent.com/blog/post/cilium-release-114/
[cilium-slack]:https://cilium.io/slack
[cilium-meeting]:https://github.com/cilium/cilium#weekly-developer-meeting

### Contour

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Contour-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/projectcontour-contour)

[Contour][contour]는 쿠버네티스를 위한 CNCF 오픈 소스로 Envoy 기반 인그레스 컨트롤러이다.

Contour [v1.31.0][contour-release]은 게이트웨이 API v1.2.1을 구현한다.
모든 [표준 채널][contour-standard] v1 API 그룹 리소스(게이트웨이 클래스, 게이트웨이, HTTPRoute, 레퍼런스그랜트)와 대부분의 v1alpha2 API 그룹 리소스(TLSRoute, TCPRoute, GRPCRoute, 레퍼런스그랜트, BackendTLSPolicy)가 지원된다.
Contour의 구현은 v1.2.1 릴리스에 포함된 대부분의 코어 확장 게이트웨이 API 호환성 테스트를 통과한다.

Contour의 게이트웨이 API 구현을 배포하고 사용하는 방법에 대한 정보는 [Contour Gateway API Guide][contour-guide]를 확인하자.

Contour의 구현에 대한 도움과 지원을 받으려면, [이슈를 생성][contour-issue-new]하거나 [쿠버네티스 slack의 #contour 채널][contour-slack]에서 도움을 요청하자.

[contour]:https://projectcontour.io
[contour-release]:https://github.com/projectcontour/contour/releases/tag/v1.30.0
[contour-standard]:concepts/versioning.md#release-channels
[contour-guide]:https://projectcontour.io/docs/1.30/guides/gateway-api/
[contour-issue-new]:https://github.com/projectcontour/contour/issues/new/choose
[contour-slack]:https://kubernetes.slack.com/archives/C8XRH2R4J

### Easegress

[Easegress][easegress]는 클라우드 네이티브 트래픽 오케스트레이션 시스템이다.

이 시스템은 현대적인 고급 게이트웨이, 견고한 분산 클러스터, 유연한 트래픽 오케스트레이터, 또는 접근 가능한 서비스 메시로 기능할 수 있다.

Easegress는 현재 [게이트웨이 컨트롤러][easegress-gatewaycontroller]를 통해 게이트웨이 API `v1beta1` 버전의 명세를 지원한다.

[easegress]:https://megaease.com/easegress/
[easegress-gatewaycontroller]:https://github.com/megaease/easegress/blob/main/docs/04.Cloud-Native/4.2.Gateway-API.md

### Emissary-Ingress (Ambassador API Gateway)

[Emissary-Ingress][emissary] (이전 Ambassador API Gateway)는 [Envoy Proxy][envoy] 위에 구축된
쿠버네티스용 인그레스 컨트롤러와 API 게이트웨이를 제공하는 오픈 소스 CNCF 프로젝트이다.
Emissary와 함께 게이트웨이 API를 사용하는 자세한 내용은 [여기][emissary-gateway-api]를 참조하자.

[emissary]:https://www.getambassador.io/docs/edge-stack
[envoy]:https://envoyproxy.io
[emissary-gateway-api]:https://www.getambassador.io/docs/edge-stack/latest/topics/using/gateway-api/

### Envoy Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-EnvoyGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/envoy-gateway)

[Envoy Gateway][eg-home]는 Envoy 기반 애플리케이션 게이트웨이를 관리하기 위한 [Envoy][envoy-org] 하위 프로젝트이다.
지원되는 게이트웨이 API의 API와 필드는 [여기][eg-supported]에 설명되어 있다.
몇 가지 간단한 단계로 게이트웨이 API와 함께 Envoy Gateway를 실행하려면 [빠른 시작][eg-quickstart]을
사용하자.

[eg-home]:https://gateway.envoyproxy.io/
[envoy-org]:https://github.com/envoyproxy
[eg-supported]:https://gateway.envoyproxy.io/docs/tasks/quickstart/
[eg-quickstart]:https://gateway.envoyproxy.io/docs/tasks/quickstart

### Flomesh Service Mesh (FSM)

[Flomesh Service Mesh][fsm]는 쿠버네티스 동/서 및 북/남 트래픽 관리를 위한 커뮤니티 주도의 경량 서비스 메시이다. Flomesh는 레이어4 트래픽 관리를 위해 [ebpf](https://www.kernel.org/doc/html/latest/bpf/index.html)를, 레이어7 트래픽 관리에 [pipy](https://flomesh.io/pipy) 프록시를 사용한다. Flomesh는 로드 밸런서, 크로스 클러스터 서비스 등록/발견을 내장으로 제공하며, 멀티 클러스터 네트워킹을 지원한다. `Ingress`("인그레스 컨트롤러"로서)와 게이트웨이 API를 지원한다.

FSM의 게이트웨이 API 지원은 [Flomesh 게이트웨이 API][fgw] 위에 구축되며 현재 쿠버네티스 게이트웨이 API 버전 [v0.7.1](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.7.1)을 지원하고 `v0.8.0` 지원이 현재 진행 중이다.

- [FSM 쿠버네티스 게이트웨이 API 호환성 매트릭스](https://github.com/flomesh-io/fsm/blob/main/docs/gateway-api-compatibility.md)
- [FSM에서 게이트웨이 API 지원을 사용하는 방법](https://github.com/flomesh-io/fsm/blob/main/docs/tests/gateway-api/README.md)

[fsm]:https://github.com/flomesh-io/fsm
[flomesh]:https://flomesh.io
[fgw]:https://github.com/flomesh-io/fgw

### Gloo Gateway

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-GlooGateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/gloo-gateway)
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Partial%20Conformance%20v1.1.0-GlooGateway-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/gloo-gateway)

[Solo.io][solo]의 [Gloo 게이트웨이][gloogateway]는 기능이 풍부한 쿠버네티스 네이티브 인그레스 컨트롤러이자 차세대 API 게이트웨이이다.
Gloo 게이트웨이는 기존 컨트롤 플레인 구현에 게이트웨이 API의 완전한 기능과 커뮤니티 지원을 제공한다.

Gloo 게이트웨이 인그레스 컨트롤러는 `HTTPRouteServiceTypes`를 제외하고
v1.1.0 릴리스의 GATEWAY_HTTP 호환성 프로필에 대한 모든 코어 게이트웨이 API 호환성 테스트를 통과한다.

[gloogateway]:https://docs.solo.io/gateway/latest/
[solo]:https://www.solo.io

### Google Cloud Service Mesh


[Google Kubernetes Engine (GKE)][gke]는 구글 클라우드에서 제공하는
관리형 쿠버네티스 플랫폼이다.

GKE의 메시를 위한 게이트웨이 (GAMMA) 구현은 [클라우드 서비스 메시][cloud-service-mesh]를 통해 이루어진다.


구글 클라우드 서비스 메시는 [Envoy 기반 사이드카 메시][envoy-sidecar-mesh]와 [Proxyless-GRPC][proxyless-grpc-mesh] (GRPCRoute 사용)를 지원한다.


[gke]:https://cloud.google.com/kubernetes-engine
[cloud-service-mesh]:https://cloud.google.com/products/service-mesh
[envoy-sidecar-mesh]:https://cloud.google.com/service-mesh/docs/gateway/set-up-envoy-mesh
[proxyless-grpc-mesh]:https://cloud.google.com/service-mesh/docs/gateway/proxyless-grpc-mesh

### Google Kubernetes Engine

[![Conformance](https://img.shields.io/badge/Gateway_API_Partial_Conformance_v1.1.0-Google_Kubernetes_Engine-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.1.0/gke-gateway)

[Google 쿠버네티스 엔진 (GKE)][gke]은 구글 클라우드에서 제공하는
관리형 쿠버네티스 플랫폼이다.
GKE의 게이트웨이 API 구현은 GKE 클러스터의 파드를 위한 구글 클라우드 로드 밸런서를 프로비저닝하는
[GKE 게이트웨이 컨트롤러][gke-gateway]를 통해 이루어진다.

GKE 게이트웨이 컨트롤러는 가중치 트래픽 분할, 미러링, 고급 라우팅, 멀티 클러스터 로드 밸런싱 등을
지원한다.
[사설 또는 공용 게이트웨이][gke-gateway-deploy] 및 [멀티 클러스터 게이트웨이][gke-multi-cluster-gateway]를
배포하는 방법은 문서를 참조한다.

GKE 게이트웨이 컨트롤러는 `HTTPRouteHostnameIntersection`을 제외하고
v1.1.0 릴리스의 GATEWAY_HTTP 호환성 프로필에 대한 모든 코어 게이트웨이 API 호환성 테스트를 통과한다.

[gke]:https://cloud.google.com/kubernetes-engine
[gke-gateway]:https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api
[gke-gateway-deploy]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways
[gke-multi-cluster-gateway]:https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-multi-cluster-gateways

### HAProxy Ingress

[HAProxy 인그레스][h1]는 HAProxy를 위한 커뮤니티 주도 인그레스 컨트롤러 구현이다.

HAProxy 인그레스 v0.13은 게이트웨이 API의 v1alpha1 명세를 부분적으로 지원한다. 호환성과 로드맵에 대한 정보는 [컨트롤러의 게이트웨이 API 문서][h2]를 참조한다.

[h1]:https://haproxy-ingress.github.io/
[h2]:https://haproxy-ingress.github.io/docs/configuration/gateway-api/

### HAProxy Kubernetes Ingress Controller

HAProxy 쿠버네티스 인그레스 컨트롤러는 HAProxy Technologies에서 유지 관리하는 오픈 소스 프로젝트로, 쿠버네티스를 위한 빠르고 효율적인 트래픽 관리, 라우팅 및 관찰 가능성을 제공한다. 버전 1.10부터 게이트웨이 API에 대한 내장 지원을 제공한다. 동일한 인그레스 컨트롤러 배포로 인그레스 API와 게이트웨이 API를 모두 사용할 수 있다. 자세한 내용은 [문서][haproxytech-docs-gw]를 참조하자. [GitHub 저장소][haproxytech-github-gw]에서 지원되는 API 리소스에 대한 추가 정보도 찾을 수 있다.

[haproxytech-docs-gw]:https://www.haproxy.com/documentation/kubernetes-ingress/gateway-api/enable-gateway-api/
[haproxytech-github-gw]:https://github.com/haproxytech/kubernetes-ingress/blob/master/documentation/gateway-api.md

### HashiCorp Consul

[HashiCorp][hashicorp]의 [Consul][consul]은 멀티 클라우드 네트워킹을 위한 오픈 소스 컨트롤 플레인이다. 단일 Consul 배포로 베어 메탈, VM 및 컨테이너 환경에 걸쳐 확장될 수 있다.

Consul 서비스 메시는 모든 쿠버네티스 배포판에서 작동하고, 다중 클러스터를 연결을 지원하며, Consul CRD는 메시에서 트래픽 패턴과 권한을 관리하는 쿠버네티스 네이티브 워크플로를 제공한다. [Consul API 게이트웨이][consul-api-gw-docs]는 북/남 트래픽 관리를 위한 게이트웨이 API를 지원한다.

게이트웨이 API의 지원되는 버전과 기능에 대한 최신 정보는 [Consul API 게이트웨이 문서][consul-api-gw-docs]를 확인하길 바란다.

[consul]:https://consul.io
[consul-api-gw-docs]:https://www.consul.io/docs/api-gateway
[hashicorp]:https://www.hashicorp.com

### Istio

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Istio-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/istio-istio)

[Istio][istio]는 오픈 소스 [서비스 메시][istio-mesh] 및 게이트웨이 구현체이다.

Istio의 최소 설치만으로 클러스터 인그레스 트래픽 제어를 위한
쿠버네티스 게이트웨이 API 완전한 적합 구현을 사용할 수 있다.
서비스 메시 사용자를 위해,
Istio는 메시 내에서 [GAMMA 이니셔티브의][gamma] 게이트웨이 API
[동/서 트래픽 관리 지원][gamma]도 완전히 지원한다.

모든 [인그레스 작업][istio-1]과 여러 메시 내부 트래픽 관리 작업을 포함한 Istio 문서의 대부분은 이미 게이트웨이 API 또는 Istio 구성 API를 사용하여 트래픽을 구성하는 병렬 지침을 포함한다.
게이트웨이 API 또는 Istio 구성 API를 사용하여 트래픽을 구성한다.
Istio의 게이트웨이 API 구현에 대한 자세한 정보는 [게이트웨이 API task][istio-2]를 확인하자.

[istio]:https://istio.io
[istio-mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/
[istio-2]:https://istio.io/latest/docs/tasks/traffic-management/ingress/gateway-api/

### kgateway
[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-kgateway-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/kgateway)

[kgateway] 프로젝트는 기능이 풍부한 쿠버네티스 네이티브 인그레스 컨트롤러이자 차세대 API 게이트웨이이다.
우수한 HTTP 경험을 유지하는 데 중점을 두고 있으며, AI 및 MCP 게이트웨이와 같은 시나리오에서 고급 라우팅 기능을 확장하고, Istio와 같은 서비스 메쉬와 엠비언트 모드 및 사이드카 모드에서 상호 운용성을 지원한다.
이러한 초점은 많은 북/남 및 동/서 사용 사례에서 성능 효율적 방식인 합리적으로 분산된 Envoy 인스턴스 세트를 쉽게 구성할 수 있음을 의미한다.

Kgateway는 2.0 릴리스와 함께 일반적으로 사용 가능하다.

[kgateway]:https://kgateway.dev/docs


### Kong Kubernetes Ingress Controller

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Kong%20Ingress%20Controller-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/kong-kubernetes-ingress-controller)

[Kong][kong]은 하이브리드 및 멀티 클라우드 환경을 위해 구축된 오픈 소스 API 게이트웨이이다.

[Kong 쿠버네티스 인그레스 게이트웨이 (KIC)][kic]는 비관리형 게이트웨이를 구성하는 데 사용할 수 있다. 사용 정보는 [Gateway API Guide][kong-gw-guide]를 확인하자.

Kong 쿠버네티스 인그레스 컨트롤러에 대한 도움과 지원을 받으려면 [이슈를 생성][kic-issue-new]하거나 [토론][kic-disc-new]을 만들자. [쿠버네티스 slack의 #kong 채널][kong-slack]에서도 도움을 요청할 수 있다.

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-gw-guide]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/using-gateway-api/
[kic-issue-new]:https://github.com/Kong/kubernetes-ingress-controller/issues/new
[kic-disc-new]:https://github.com/Kong/kubernetes-ingress-controller/discussions/new
[kong-slack]:https://kubernetes.slack.com/archives/CDCA87FRD

### Kong Gateway Operator

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.0-Kong%20Gateway%20Operator-orange)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.0/kong-gateway-operator)

[Kong][kong]은 하이브리드 및 멀티 클라우드 환경을 위해 구축된 오픈 소스 API 게이트웨이이다.

[Kong 게이트웨이 오퍼레이터 (KGO)][kgo]는 관리형 게이트웨이를 구성하고, [Kong 쿠버네티스 인그레스 컨트롤러](#kong-kubernetes-ingress-controller)의 인스턴스를 오케스트레이션하는 데 사용할 수 있다.

Kong 쿠버네티스 인그레스 컨트롤러에 대한 도움과 지원을 받으려면 [이슈를 생성][kgo-issue-new]하거나 [토론][kgo-disc-new]을 만들자. [쿠버네티스 slack의 #kong 채널][kong-slack]에서도 도움을 요청할 수 있다.

[kgo]:https://docs.konghq.com/gateway-operator/latest/
[kgo-issue-new]:https://github.com/Kong/gateway-operator/issues/new
[kgo-disc-new]:https://github.com/Kong/gateway-operator/discussions/new


### Kubvernor
[Kubvernor][kubvernor]는 Rust 프로그래밍 언어로 구현된 오픈소스이자 고도로 실험적인 API 컨트롤러이다. 현재 Kubvernor는 Envoy Proxy를 지원하며, 다양한 게이트웨이(Envoy, Nginx, HAProxy 등)를 관리/배포할 수 있도록 가능한 한 일반적인 구조를 목표로 한다.

[kubvernor]:https://github.com/kubvernor/kubvernor

### Kuma

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.0.0-Kuma-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.0.0/kumahq-kuma)

[Kuma][kuma]는 오픈 소스 서비스 메시이다.

Kuma는 베타 안정성을 보장하며 Kuma 내장형, Envoy 기반 게이트웨이에 대한 게이트웨이 API 명세를 구현한다. 게이트웨이 API를 사용하여 Kuma 내장 게이트웨이를 설정하는 방법에 대한 정보는 [게이트웨이 API 문서][kuma-1]을 확인한다.

Kuma 2.3 이상은 메시 내에서 [GAMMA 이니셔티브의][gamma] 
게이트웨이 API [동/서 트래픽 관리 지원][gamma]을 지원한다.

[kuma]:https://kuma.io
[kuma-1]:https://kuma.io/docs/latest/using-mesh/managing-ingress-traffic/gateway-api/

### Linkerd

[Linkerd][linkerd]는 최초의 CNCF 졸업 [서비스 메시][linkerd-mesh]이다.
Envoy를 기반으로 하지 않은 유일한 주요 메쉬로,
대신 Rust로 특별히 설계된 마이크로 프록시를 활용해
Kubernetes에 보안, 가시성, 신뢰성을 제공하며 복잡성을 제거한다.

Linkerd 2.14 이상은 메시 내에서 [GAMMA 이니셔티브의][gamma]
게이트웨이 API [동/서 트래픽 관리 지원][gamma]을 지원한다.

[linkerd]:https://linkerd.io/
[linkerd-mesh]:https://buoyant.io/service-mesh-manifesto

### LiteSpeed Ingress Controller

[LiteSpeed 인그레스 컨트롤러](https://litespeedtech.com/products/litespeed-web-adc/features/litespeed-ingress-controller)는 LiteSpeed WebADC 컨트롤러를 사용하여 인그레스 컨트롤러 및 로드 밸런서로 동작하며, 쿠버네티스 클러스터 내의 트래픽을 관리한다. 이 컨틀로러는 게이트웨이, 게이트웨이 클래스, HTTPRoute, 레퍼런스그랜트를 포함한 게이트웨이 API의 코어 기능 전체와 cert-manager의 게이트웨이 기능을 구현하고 있다. 게이트웨이는 LiteSpeed Ingress Controller에 완전히 통합되어 있다.

- [제품 문서](https://docs.litespeedtech.com/cloud/kubernetes/).
- [게이트웨이 상세 문서](https://docs.litespeedtech.com/cloud/kubernetes/gateway).
- 전체 지원은 [LiteSpeed support 웹사이트](https://www.litespeedtech.com/support)에서 제공한다.

### LoxiLB

[kube-loxilb][kube-loxilb-gh]는 [LoxiLB's][loxilb-org]가 구현한 게이트웨이 API 및 쿠버네티스 서비스 로드 밸런서 명세 구현체로, 로드 밸런서 클래스, 고급 IPAM(공유 또는 전용) 등을 지원한다. kube-loxilb는 L4 서비스 로드 밸런서로서 [LoxiLB][loxilb-gh]를 인그레스(L7) 리소스를 위해 [loxilb-ingress][loxilb-ingress-gh]를 사용하여 게이트웨이 API 리소스를 관리한다.

간단한 단계로 게이트웨이 API와 함께 LoxiLB를 실행하려면 [빠른 시작 가이드][loxigw-guide]를 참고하자.

[loxilb-home]:https://loxilb.io/
[loxilb-org]:https://github.com/loxilb-io
[loxilb-gh]:https://github.com/loxilb-io/loxilb
[kube-loxilb-gh]:https://github.com/loxilb-io/kube-loxilb
[loxilb-ingress-gh]:https://github.com/loxilb-io/loxilb-ingress
[loxigw-guide]:https://docs.loxilb.io/latest/gw-api/

### NGINX Gateway Fabric

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-NGINX Gateway Fabric-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/nginx-nginx-gateway-fabric)

[NGINX 게이트웨이 패브릭][nginx-gateway-fabric]은 [NGINX][nginx]를 데이터 플레인으로 사용하는 게이트웨이 API의 구현체를 제공하는 오픈소스 프로젝트이다. 이 프로젝트의 목표는 쿠버네티스에서 실행되는 애플리케이션을 위한 HTTP 또는 TCP/UDP 로드 밸런서, 리버스 프록시 또는 API 게이트웨이를 구성하기 위해 코어 게이트웨이 API를 구현하는 것이다. [NGINX 문서][nginx-docs] 웹사이트에서 종합적인 NGINX 게이트웨이 패브릭 사용자 문서를 찾을 수 있다.

지원되는 게이트웨이 API 리소스 및 기능 목록은 [게이트웨이 API 호환성][nginx-compat] 문서를 확인하자.

NGINX 게이트웨이 패브릭에 대한 제안이 있거나 문제를 경험했다면 GitHub에서 [이슈를 생성][nginx-issue-new]하거나 [토론][nginx-disc-new]을 부탁한다. 또한 [NGINX 커뮤니티 포럼][nginx-forum]에서 도움도 요청할 수 있다.

[nginx-gateway-fabric]:https://github.com/nginx/nginx-gateway-fabric
[nginx]:https://nginx.org/
[nginx-docs]:https://docs.nginx.com/nginx-gateway-fabric/
[nginx-compat]:https://docs.nginx.com/nginx-gateway-fabric/overview/gateway-api-compatibility/
[nginx-issue-new]:https://github.com/nginx/nginx-gateway-fabric/issues/new
[nginx-disc-new]:https://github.com/nginx/nginx-gateway-fabric/discussions/new
[nginx-forum]:https://community.nginx.org/


### ngrok Kubernetes Operator

[ngrok 쿠버네티스 오퍼레이터][ngrok-k8s-operator]는 작년에 초기 지원을 추가한 이후로 게이트웨이 API의 전체 코어를 지원한다. 이것은 다음을 포함한다.

- 라우트: (HTTPRoute, TCPRoute, TLSRoute) 및 RouteMatches (Header, Path, 등)
- 필터: Header, Redirect, Rewrite 등
- 백엔드: 백엔드 Filters 및 가중치 기반 밸런싱
- 레퍼런스그랜트: 멀티 테넌트 클러스터 처리를 위한 RBAC
- 게이트웨이 API가 충분히 유연하지 않은 경우, extensionRef 또는 어노테이션으로 트래픽 정책 설정

자세한 내용은[docs][ngrok-k8s-gwapi-docs]를 참고하자. 기능 요청이나 버그 리포트는 [create an issue][ngrok-issue-new]을 통해 제출을 부탁한다. 또한 [Slack][ngrok-slack]에서 도움을 받을 수 있다.

[ngrok-k8s-operator]:https://github.com/ngrok/ngrok-operator
[ngrok]:https://ngrok.com
[ngrok-k8s-gwapi-docs]:https://ngrok.com/docs/k8s/
[ngrok-issue-new]: https://github.com/ngrok/ngrok-operator/issues/new/choose
[ngrok-slack]:https://ngrokcommunity.slack.com/channels/general

### STUNner

[STUNner][stunner]는 쿠버네티스용 오픈소스 클라우드 네이티브 WebRTC 미디어 게이트웨이이다. STUNner는 WebRTC 미디어 스트림을 쿠버네티스 클러스터로 원활하게 수신하기 위한 목적으로 설계되었으며, 간소화된 NAT 트래버설과 동적 미디어 라우팅을 제공한다. 동시에 STUNner는 대규모 실시간 통신 서비스에 대해 보안성과 모니터링 기능을 향상시킨다. STUNner의 데이터 플레인은 WebRTC 클라이언트를 위한 표준 규격의 TURN 서비스를 제공하며, 컨트롤 플레인은 게이트웨이 API의 일부를 지원한다.

현재 STUNner는 게이트웨이 API 명세의 `v1alpha2` 버전을 지원한다. WebRTC 미디어 수신을 위해 STUNner를 배포하고 사용하는 방법은 [설치 가이드][stunner-1]를 확인하자. STUNner와 관련된 모든 질문, 의견 및 버그 리포트는 [STUNner 프로젝트][stunner]로 보내주시기 바란다.

[stunner]:https://github.com/l7mp/stunner
[stunner-1]:https://github.com/l7mp/stunner/blob/main/doc/INSTALL.md

### Traefik Proxy

[![Conformance](https://img.shields.io/badge/Gateway%20API%20Conformance%20v1.2.1-Traefik Proxy-green)](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/reports/v1.2.1/traefik-traefik)

[Traefik Proxy][traefik-proxy]는 오픈 소스 클라우드 네이티브 애플리케이션 프록시이다.

Traefik 프록시는 현재 게이트웨이 API 명세의 `v1.2.1` 버전을 지원한다. 배포 및 사용 방법에 대한 자세한 정보는 [쿠버네티스 제공자 문서][traefik-proxy-gateway-api-doc]를 확인하자.
Traefik 프록시의 구현은 GRPCRoute와 같은 모든 HTTP 코어 및 일부 확장 호환성 테스트를 통과하며, 실험적 채널의 TCPRoute 및 TLSRoute 기능도 지원한다.

Traefik 프록시에 대한 도움과 지원을 받으려면, [이슈를 생성][traefik-proxy-issue-new]하거나 [Traefik Labs 커뮤니티 포럼][traefiklabs-community-forum]에서 도움을 요청하자.

[traefik-proxy]:https://traefik.io
[traefik-proxy-gateway-api-doc]:https://doc.traefik.io/traefik/v3.2/routing/providers/kubernetes-gateway/
[traefik-proxy-issue-new]:https://github.com/traefik/traefik/issues/new/choose
[traefiklabs-community-forum]:https://community.traefik.io/c/traefik/traefik-v3/21

### Tyk

[Tyk 게이트웨이][tyk-gateway]는 클라우드 네이티브 오픈소스 API 게이트웨이이다.

[Tyk.io][tyk] 팀은 게이트웨이 API 구현을 목표로 작업 중이며, 이 프로젝트의 진행 상황은 [여기][tyk-operator]에서 확인할 수 있다.

[tyk]:https://tyk.io
[tyk-gateway]:https://github.com/TykTechnologies/tyk
[tyk-operator]:https://github.com/TykTechnologies/tyk-operator

### WSO2 APK

[WSO2 APK][wso2-apk]는 쿠버네티스 환경을 위해 특별히 설계된 API 관리 솔루션으로, API 관리를 위한 통합성, 유연성, 확장성을 조직에 제공한다.

WSO2 APK는 게이트웨이 API를 구현하며, 게이트웨이 및 HTTPRoute 기능을 포함한다. 또한, 사용자 정의 리소스(CR)를 통해 레이트 리밋팅, 인증/인가, 분석/관찰 가능성을 지원한다.

게이트웨이 API의 지원 버전과 기능에 대한 최신 정보는 [APK 게이트웨이 문서][apk-doc]를 참고하자. 질문이 있거나 기여하고 싶다면 자유롭게 [이슈 또는 풀 리퀘스트][repo]를 생성할 수 있다. 또한 [Discord 채널][discord]에서 우리와 소통하고 토론에 참여할 수 있다.

[wso2-apk]:https://apk.docs.wso2.com/en/latest/
[apk-doc]:https://apk.docs.wso2.com/en/latest/catalogs/kubernetes-crds/
[repo]:https://github.com/wso2/apk
[discord]:https://discord.com/channels/955510916064092180/1113056079501332541

## 통합

이 섹션에서는 특정 통합을 위한 블로그 포스트, 문서 및 기타 게이트웨이 API 참조에 대한 구체적인 링크를 찾을 수 있다.

### Flagger

[Flagger][flagger]는 쿠버네티스에서 실행되는 애플리케이션의 릴리스 프로세스를 자동화하는 점진적 배포 도구이다.

Flagger는 게이트웨이 API를 사용하여 카나리 배포와 A/B 테스트를 자동화하는 데 사용할 수 있다. 게이트웨이 API의 `v1alpha2`와 `v1beta1` 명세를 모두 지원한다. 게이트웨이 API의 모든 구현과 함께 Flagger를 사용하려면 [이 튜토리얼][flagger-tutorial]을 참조한다.

[flagger]:https://flagger.app
[flagger-tutorial]:https://docs.flagger.app/tutorials/gatewayapi-progressive-delivery

### cert-manager

[cert-manager][cert-manager]는 클라우드 네이티브 환경에서 인증서 관리를 자동화하기 위한 도구이다.

cert-manager는 게이트웨이 리소스를 위한 TLS 인증서를 생성할 수 있다. 이는 게이트웨이에 어노테이션을 추가하여 구성된다. 현재 게이트웨이 API의 `v1alpha2` 명세를 지원한다. 사용해보려면 [cert-manager 문서][cert-manager-docs]를 참조한다.

[cert-manager]:https://cert-manager.io/
[cert-manager-docs]:https://cert-manager.io/docs/usage/gateway/

### Argo rollouts

[Argo Rollouts][argo-rollouts]는 쿠버네티스를 위한 점진적 배포 컨트롤러이다. 블루/그린 및 카나리와 같은 여러 고급 배포 방법을 지원한다. Argo Rollouts는 [플러그인][argo-rollouts-plugin]을 통해 게이트웨이 API를 지원한다.

[argo-rollouts]:https://argo-rollouts.readthedocs.io/en/stable/
[argo-rollouts-plugin]:https://github.com/argoproj-labs/rollouts-gatewayapi-trafficrouter-plugin/

### Knative

[Knative][knative]는 쿠버네티스 위에 구축된 서버리스 플랫폼이다. Knative Serving은 URL의 자동 관리, 리비전 간 트래픽 분할, 요청 기반 자동 스케일링(제로 스케일 포함), 자동 TLS 프로비저닝과 함께 상태 비저장 컨테이너를 실행하기 위한 간단한 API를 제공한다. Knative Serving은 플러그인 아키텍처를 통해 다중 HTTP 라우터를 지원하며, 이는 모든 Knative 기능이 지원되지 않아 현재 알파 단계에 있는 [게이트웨이 API 플러그인][knative-net-gateway-api]을 포함한다.

[knative]:https://knative.dev/
[knative-net-gateway-api]:https://github.com/knative-sandbox/net-gateway-api

### Kuadrant

[Kuadrant][kuadrant]는 다른 게이트웨이 API 제공자와 통합되고 정책 연결을 통해 정책을 제공하는 오픈 소스 멀티 클러스터 게이트웨이 API 컨트롤러이다.

Kuadrant는 게이트웨이를 중앙에서 정의하고 모든 게이트웨이에 적용되는 DNS, TLS, 인증 및 레이트 리밋팅과 같은 정책을 연결하기 위한 게이트웨이 API를 지원한다.

Kuadrant는 Istio와 Envoy Gateway를 기본 게이트웨이 API 제공자로 지원하며, 향후 다른 게이트웨이 제공자와도 작동할 계획이다.

Kuadrant의 구현에 대한 도움과 지원을 받으려면, 자유롭게 [이슈를 생성][kuadrant-issue-new]하거나 [쿠버네티스 slack의 #kuadrant 채널][kuadrant-slack]에서 도움을 요청하자.

[kuadrant]:https://kuadrant.io/
[kuadrant-issue-new]:https://github.com/Kuadrant/kuadrant-operator/issues/new
[kuadrant-slack]:https://kubernetes.slack.com/archives/C05J0D0V525

