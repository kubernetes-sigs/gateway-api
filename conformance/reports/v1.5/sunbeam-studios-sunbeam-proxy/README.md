# Sunbeam Proxy

Sunbeam Proxy is a cloud-native reverse proxy with adaptive ML threat detection that implements the Kubernetes Gateway API v1.5.1 control plane and data plane.

## Table of contents

| API channel  | Implementation version                                              | Mode    | Report                                                          |
|--------------|---------------------------------------------------------------------|---------|-----------------------------------------------------------------|
| experimental | [v0.2.0](https://github.com/sunbeamdotpt/proxy/releases/tag/v0.2.0) | default | [link](./experimental-v0.2.0-default-report.yaml)               |

## Overview

This report covers the Gateway API conformance test results for Sunbeam Proxy v0.2.0. The tests were run against the published multi-arch container image
`ghcr.io/sunbeamdotpt/proxy:v0.2.0` using the upstream Gateway API v1.5.1 experimental conformance suite.

The following conformance profiles were exercised:

- `GATEWAY-HTTP`
- `GATEWAY-GRPC`
- `GATEWAY-TLS`

Service mesh profiles are not supported and are skipped.

TCPRoute & UDPRoute are fully supported, however there are no conformance tests for them.

## Reproduce

1. Clone the Sunbeam Proxy repository:

   ```bash
   git clone https://github.com/sunbeamdotpt/proxy.git
   cd proxy
   ```

2. Check out the release tag:

   ```bash
   git checkout v0.2.0
   ```

3. Ensure you have a Kubernetes cluster with the Gateway API v1.5.1 experimental CRDs installed and a `GatewayClass` named `sunbeam` whose `controllerName` is `sunbeam.pt/sunbeam-proxy`. The included helper script expects a Multipass VM named `sunbeam-proxy-dev` running k3s and a kubeconfig at `/tmp/k3s.yaml`.

   To create the VM from the project's cloud-init, run:

   ```bash
   multipass launch --name sunbeam-proxy-dev \
     --cpus 2 \
     --memory 8G \
     --disk 25G \
     --cloud-init cloud-init-multipass-dev.yaml \
     lts
   ```

   Then copy the kubeconfig from the VM:

   ```bash
   multipass exec sunbeam-proxy-dev -- sudo cat /etc/rancher/k3s/k3s.yaml \
     | sed 's/127.0.0.1/<vm-ip>/g' > /tmp/k3s.yaml
   ```

   Replace `<vm-ip>` with the VM's IP from `multipass info sunbeam-proxy-dev`.

4. Run the conformance suite against the published image:

   ```bash
   ./scripts/conformance.sh run -p
   ```

   The `-p` flag pulls `ghcr.io/sunbeamdotpt/proxy:v0.2.0` from the registry instead of building a local image.

5. Inspect the generated report:

   ```bash
   cat target/conformance-report.yaml
   ```
