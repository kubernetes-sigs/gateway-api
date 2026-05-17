# Varnish Gateway

[Varnish Gateway](https://gateway.varnish.org) is a Kubernetes [Gateway API](https://gateway-api.sigs.k8s.io/)
implementation that uses [Varnish](https://varnish-cache.org/) as its data plane. It is developed and maintained
by [Varnish Software AS](https://www.varnish-software.com/).

## Table of contents

| API channel | Implementation version | Mode    | Report                                                                |
|-------------|------------------------|---------|-----------------------------------------------------------------------|
| standard    | [v0.20.0](https://github.com/varnish/gateway/releases/tag/v0.20.0) | default | [v0.20.0 report](./standard-v0.20.0-default-report.yaml) |

## Reproduce

The following steps reproduce the Varnish Gateway conformance test report.

1. Create a kind cluster and install the Gateway API standard CRDs:

    ```bash
    git clone https://github.com/varnish/gateway
    cd gateway
    git checkout v0.20.0
    make kind-create
    ```

1. Build the operator and chaperone images, load them into the kind cluster, and deploy:

    ```bash
    make docker
    make kind-load
    make kind-deploy
    ```

1. Run the conformance tests and generate the report:

    ```bash
    make test-conformance-report
    ```

    The report will be written to `dist/conformance-report.yaml`.

Alternatively, the full cycle (kind cluster create, build, deploy, test, teardown) can be run with:

```bash
make test-conformance-kind
```
