# Envoy Gateway

## Table of Contents

| API channel  | Implementation version                                              | Mode                         | Report                                                           |
|--------------|---------------------------------------------------------------------|------------------------------|------------------------------------------------------------------|
| experimental | [v1.5.0](https://github.com/envoyproxy/gateway/releases/tag/v1.5.0) | ControllerNamespace(default) | [link](./experimental-v1.5.0-default-report.yaml)                |
| experimental | [v1.5.0](https://github.com/envoyproxy/gateway/releases/tag/v1.5.0) | GatewayNamespace             | [link](./experimental-v1.5.0-gateway-namespace-mode-report.yaml) |


## Overview

Envoy Gateway supports different deployment [modes](https://gateway.envoyproxy.io/docs/tasks/operations/deployment-mode/#supported-modes),
including a controller namespace mode(the default one) and a [gateway namespace mode](https://gateway.envoyproxy.io/docs/tasks/operations/deployment-mode/#gateway-namespace-mode).
The conformance tests are run against both modes to ensure compatibility and functionality.

## Reproduce

1. Clone the Envoy Gateway GitHub repository

   ```bash
   git clone https://github.com/envoyproxy/gateway.git && cd gateway
   ```

2. Check out the desired version

   ```bash
   export VERSION=v<x.y.z>
   git checkout $VERSION
   ```

3. Run the conformance tests

    ```bash
   KUBE_DEPLOY_PROFILE=default CONFORMANCE_REPORT_PATH=conformance-report-k8s.yaml make experimental-conformance
   ```
   or 

   ```bash
   KUBE_DEPLOY_PROFILE=gateway-namespace-mode CONFORMANCE_REPORT_PATH=conformance-report-k8s.yaml make experimental-conformance
   ```

4. Check the produced report

   ```bash
   cat ./conformance-report-k8s.yaml
   ```
