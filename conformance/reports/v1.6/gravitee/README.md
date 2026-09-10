# Gravitee

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [version-4.10.3](https://github.com/gravitee-io/gravitee-kubernetes-operator/releases/tag/4.12.10) | default | [version-4.10.3 report](./standard-4.12.10-default-report.yaml) |


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
git clone --depth 1 --branch 4.12.10 https://github.com/gravitee-io/gravitee-kubernetes-operator.git
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
GATEWAY_API_MATCH_ACROSS_ROUTES=true make conformance
```

7. Print report

```bash
cat test/conformance/kubernetes.io/gateway-api/report/standard-4.12.10-default-report.yaml
```

