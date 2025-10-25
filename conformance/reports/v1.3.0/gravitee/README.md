# Gravitee

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [version-4.8.5](https://github.com/gravitee-io/gravitee-kubernetes-operator/releases/tag/4.8.5) | default | [version-4.8.5 report](./standard-4.8.5-default-report.yaml) |

> The Gravitee Kubernetes Operator provides partial conformance for Gateway - HTTP features in version 4.8.5. It does not support matching rules across routes or defining services of a type other than Kubernetes Core v1 services. These features will be introduced in a future release.

## Prerequisites

The following binaries are assumed to be installed on your device
  
  - [docker](https://docs.docker.com/get-started/get-docker/)
  - [kubectl](https://kubernetes.io/docs/tasks/tools/)
  - [kind](https://github.com/kubernetes-sigs/kind)
  - [go](https://go.dev/learn/)

The reproducer has been tested on macOS and Linux only.

## Reproducer

1. Clone the Gravitee Kubernetes Operator repository

```bash
git clone --depth 1 --branch 4.8.x https://github.com/gravitee-io/gravitee-kubernetes-operator.git && cd gravitee-kubernetes-operator
```

2. Start the Kubernetes cluster

```bash
make start-conformance-cluster
```

3. Run a local Load Balancer Service

> The make target runs [cloud-provider-kind](https://kind.sigs.k8s.io/docs/user/loadbalancer). If you are reproducing on a macOS device, the binary requires `sudo` privileges and you will be prompted for a password. For Linux devices, cloud-provider-kind will be run using Docker compose.

```bash
make cloud-lb
```

4. Run the operator

```bash
make run
```

5. Install the Gravitee GatewayClass

```bash
kubectl apply -f ./test/conformance/gateway-class-parameters.report.yaml -f ./test/conformance/gateway-class.yaml
```

6. Run the conformance tests

```bash
make conformance
```

7. Print report

```bash
cat test/conformance/kubernetes.io/gateway-api/report/standard-4.8.5-default-report.yaml
```

