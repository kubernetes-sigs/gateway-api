# Implementations

This document tracks downstream implementations of Gateway API and provides status and resource references for them.

Implementors of Gateway API are encouraged to update this document with status information about their implementations, the versions they cover, and documentation to help users get started.

## Implementation Status

|                | [Contour][1]       | [GCP][2]           | [Istio][3]         | [Kong][4]         | [Knative][5]       | [Traefik][6]       |
| :------------- | :----------------: | :----------------: | :----------------: | :---------------: | :----------------: | :----------------: |
| [v0.1.x][v0.1] | :heavy_minus_sign: | :heavy_minus_sign: | :heavy_check_mark: |:heavy_minus_sign: | :heavy_minus_sign: | :heavy_check_mark: |
| [v0.2.x][v0.2] | :heavy_check_mark: | :heavy_plus_sign:  | :heavy_plus_sign:  |:heavy_plus_sign:  | :heavy_minus_sign: | :heavy_plus_sign:  |

- :heavy_check_mark: available
- :heavy_plus_sign: in progress
- :heavy_minus_sign: not planned

[1]:#contour
[2]:#google-cloud-platform
[3]:#istio
[4]:#kong
[5]:#knative
[6]:#traefik

[v0.1]:https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.1.0
[v0.2]:https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.2.0

## Project References

In this section you will find specific links to blog posts, documentation and other Gateway API references for specific implementations.

### Contour

[Contour][contour] is an open source ingress controller for Kubernetes.

Contour currently implements the latest `v0.2.x` Gateway API Specification.

See the [Contour Gateway API Guide][contour-1] for information on how to deploy and use Contours's Gateway implementation.

[contour]:https://projectcontour.io
[latest]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/
[contour-1]:https://projectcontour.io/guides/gateway-api/

### Google Cloud Platform

The [Google Cloud Platform (GCP)][gcp] is a cloud computing platform and infrastructure provider.

GCP is actively working towards an implementation of the Gateway API `v0.2.x` specification, and status updates and documentation will be provided here as the work progresses.

[gcp]:https://cloud.google.com

### Istio

[Istio][istio] is an open source [service mesh][mesh] for Kubernetes clusters.

Istio currently supports the previous `v0.1.x` Gateway API specification and is working towards supporting the latest `v0.2.x` release.

Status updates and documentation for `v0.2.x` will be provided here as the work progresses.

See the [Istio Gateway API Documentation][istio-1] for information on how to deploy and use Istio's Gateway implementation.

[istio]:https://istio.io
[mesh]:https://istio.io/latest/docs/concepts/what-is-istio/#what-is-a-service-mesh
[istio-1]:https://istio.io/latest/docs/tasks/traffic-management/ingress/service-apis/

### Kong

[Kong][kong] is an open source API Gateway built for hybrid and multi-cloud environments.

Kong is actively working towards an implementation of the Gateway API `v0.2.x` specification for it's [Kubernetes Ingress Controller][kic] and status updates and documentation will be provided here as the work progresses.

You can subscribe to [kong/kubernetes-ingress-controller/issues/692][kong-1] to track the implementation progress and contribute.

[kong]:https://konghq.com
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-1]:https://github.com/kong/kubernetes-ingress-controller/issues/692

### Knative

[Knative][knative] is an open source Kubernetes-based platform to deploy and manage modern serverless workloads.

Knative is actively working towards an implementation of the Gateway API, status updates and documentation will be provided here as the work progresses.

[knative]:https://knative.dev/

### Traefik

[Traefik][traefik] is an open source cloud-native application proxy.

Traefik currently supports the previous `v0.1.x` Gateway API specification, check the [Kubernetes Gateway Documentation][traefik-1] for information on how to deploy and use Traefik's Gateway implementation.

Traefik is currently working on implementing `v0.2.x`, status updates and documentation will be provided here as the work progresses.

[traefik]:https://traefik.io
[traefik-1]:https://doc.traefik.io/traefik/routing/providers/kubernetes-gateway/
