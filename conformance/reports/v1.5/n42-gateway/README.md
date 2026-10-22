# N42 Gateway

## Table of contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
| experimental | [v0.17.0-alpha.3](https://github.com/n42-gateway/n42-gateway/releases/tag/v0.17.0-alpha.3) | default | [v0.17.0-alpha.3 report](./experimental-v0.17.0-alpha.3-default-report.yaml) |

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

1. Deploy N42 Gateway

    ```bash
    helm upgrade --install --create-namespace --namespace n42-gateway-system \
      n42 https://github.com/n42-gateway/charts/releases/download/0.17.0-alpha.3/n42-gateway-0.17.0-alpha.3.tgz \
      --set controller.kind=DaemonSet \
      --set controller.daemonset.useHostPort=true \
      --set controller.service.type=ClusterIP \
      --set controller.extraArgs.publish-address=127.0.0.1 \
      --set controller.gatewayClassResource.enabled=true
    ```
    > Set only `gatewayClassResource` value above to expose controller via LoadBalancer service type instead.

1. Run the conformance tests

    ```bash
    git clone --branch v1.5.1 --depth 1 https://github.com/kubernetes-sigs/gateway-api
    cd gateway-api
    ```
    
    ```bash
    go test ./conformance -run TestConformance -v -timeout=1h -args \
      --project=n42-gateway \
      --organization="N42 Gateway" \
      --url=https://n42-gateway.github.io \
      --version=v0.17.0-alpha.3 \
      --contact=https://kubernetes.slack.com/channels/n42-gateway \
      --gateway-class=n42 \
      --supported-features=Gateway,HTTPRoute,GatewayAddressEmpty,GatewayPort8080,HTTPRouteBackendProtocolWebSocket,HTTPRouteCORS,HTTPRouteDestinationPortMatching,HTTPRouteNamedRouteRule,HTTPRouteResponseHeaderModification,HTTPRouteSchemeRedirect \
      --conformance-profiles=GATEWAY-HTTP \
      --report-output=${PWD}$/experimental-report.yaml
    ```
