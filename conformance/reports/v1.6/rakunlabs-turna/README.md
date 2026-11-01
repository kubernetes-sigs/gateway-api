# Turna

[Turna](https://github.com/rakunlabs/turna) is an open source operations sidecar
and Gateway API implementation written in Go.

## Table of Contents

| API channel  | Implementation version                                              | Mode    | Report                                                       |
|--------------|---------------------------------------------------------------------|---------|--------------------------------------------------------------|
| experimental | [v0.12.5](https://github.com/rakunlabs/turna/releases/tag/v0.12.5) | default | [v0.12.5 report](./experimental-v0.12.5-default-report.yaml) |

## Reproduce

Requirements: Docker, [kind](https://kind.sigs.k8s.io/), kubectl and Go.

Clone the Turna repository and check out the release:

```shell
git clone https://github.com/rakunlabs/turna.git && cd turna
git checkout v0.12.5
```

Run the conformance suite:

```shell
WORK=/tmp/turna-conformance ./deploy/conformance/run.sh
```

The script creates the kind cluster `turna`, installs the Gateway API
experimental channel CRDs and MetalLB for LoadBalancer addresses, builds and
deploys Turna with `deploy/conformance/kustomize` (proxies are created in the
Gateway namespace, which the `GatewayInfrastructure` tests need), then runs the
`GATEWAY-HTTP`, `GATEWAY-GRPC`, `GATEWAY-TLS`, `GATEWAY-TCP` and `GATEWAY-UDP`
profiles with the features Turna reports as supported.

The report is written to `$WORK/report.yaml`.
