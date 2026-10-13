# Portus

[Portus](https://portus-gateway.dev) is a Kubernetes gateway written in Rust that carries HTTP, LLM and MCP traffic on one data plane.

## Table of Contents

| API channel  | Implementation version                                                     | Mode    | Report                                                       |
|--------------|----------------------------------------------------------------------------|---------|--------------------------------------------------------------|
| experimental | [v0.2.12](https://github.com/Portus-Gateway/Portus/releases/tag/v0.2.12)   | default | [v0.2.12 report](./experimental-v0.2.12-default-report.yaml) |

## Reproduce

The suite runs in-cluster, because each Gateway's address is the ClusterIP of its own data plane Service.
You need Docker, [k3d](https://k3d.io), kubectl, Helm 3.8+ and Go 1.26.

Clone the repository and create a cluster with the experimental Gateway API CRDs:

```shell
git clone https://github.com/Portus-Gateway/Portus.git && cd Portus
make k3d-up gateway-api-crds
```

Install the released chart:

```shell
helm install portus oci://ghcr.io/portus-gateway/charts/portus-gateway --version 0.2.12 \
  --namespace portus --create-namespace \
  --set dataplane.service.type=ClusterIP --set dataplane.replicasPerGateway=1 --wait
```

Build the conformance runner and run all five profiles (HTTP, GRPC, TLS, TCP, UDP):

```shell
make conformance-image
make conformance-run CONFORMANCE_IMPL_VERSION=v0.2.12
```

The report is written to `tests/conformance/conformance-report.yaml`.
