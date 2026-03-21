# HAProxy Ingress

## Table of contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
| experimental | [v0.17.0-alpha.1](https://github.com/jcmoraisjr/haproxy-ingress/releases/tag/v0.17.0-alpha.1) | default | [v0.17.0-alpha.1 report](./experimental-v0.17.0-alpha.1-default-report.yaml) |

## Reproduce

The following steps reproduce the HAProxy Ingress conformance test report.

1. Create a kind cluster

    ```yaml
    # kind.yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    nodes:
    - role: control-plane
      extraPortMappings:
      - containerPort: 80
        hostPort: 80
      - containerPort: 443
        hostPort: 443
    ```

    ```bash
    kind create cluster --config kind.yaml
    ```

1. Deploy Gateway API CRDs

    ```bash
    kubectl create -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.5.1/experimental-install.yaml
    ```

1. Deploy HAProxy Ingress

    ```bash
    helm upgrade --install --create-namespace --namespace ingress-controller \
      haproxy-ingress https://github.com/haproxy-ingress/charts/releases/download/0.17.0-alpha.1/haproxy-ingress-0.17.0-alpha.1.tgz \
      --set controller.kind=DaemonSet \
      --set controller.daemonset.useHostPort=true \
      --set controller.service.type=ClusterIP \
      --set controller.extraArgs.publish-address=127.0.0.1 \
      --set controller.gatewayClassResource.enabled=true
    ```
    > Set only `gatewayClassResource` value above to expose controller via LoadBalancer service type instead.

1. Run the conformance tests

    ```bash
    git clone https://github.com/kubernetes-sigs/gateway-api
    cd gateway-api
    git checkout v1.5.1
    ```
    
    ```bash
    go test ./conformance -run TestConformance -v -timeout=1h -args \
      --project=haproxy-ingress \
      --organization="HAProxy Ingress" \
      --url=https://haproxy-ingress.github.io \
      --version=v0.17.0-alpha.1 \
      --contact=https://kubernetes.slack.com/channels/haproxy-ingress \
      --gateway-class=haproxy \
      --supported-features=Gateway,HTTPRoute,GatewayAddressEmpty,GatewayPort8080,HTTPRouteBackendProtocolWebSocket,HTTPRouteCORS,HTTPRouteDestinationPortMatching,HTTPRouteNamedRouteRule,HTTPRouteResponseHeaderModification,HTTPRouteSchemeRedirect \
      --conformance-profiles=GATEWAY-HTTP \
      --report-output=/tmp/experimental-report.yaml
    ```
