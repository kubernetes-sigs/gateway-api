# Projectcontour Contour

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v1.28.1](https://github.com/projectcontour/contour/releases/tag/v1.28.1)|x|[v1.28.1 report](./v1.28.1-report.yaml)|
|experimental|[v1.28.2](https://github.com/projectcontour/contour/releases/tag/v1.28.2)|x|[v1.28.2 report](./v1.28.2-report.yaml)|
|experimental|[v1.28.3](https://github.com/projectcontour/contour/releases/tag/v1.28.3)|x|[v1.28.3 report](./v1.28.3-report.yaml)|
|experimental|[v1.28.4](https://github.com/projectcontour/contour/releases/tag/v1.28.4)|x|[v1.28.4 report](./v1.28.4-report.yaml)|
|experimental|[v1.28.5](https://github.com/projectcontour/contour/releases/tag/v1.28.5)|x|[v1.28.5 report](./experimental-v1.28.5-default-report.yaml)|
|experimental|[v1.28.6](https://github.com/projectcontour/contour/releases/tag/v1.28.6)|x|[v1.28.6 report](./experimental-v1.28.6-default-report.yaml)|
|experimental|[v1.29.0](https://github.com/projectcontour/contour/releases/tag/v1.29.0)|x|[v1.29.0 report](./v1.29.0-report.yaml)|
|experimental|[v1.29.1](https://github.com/projectcontour/contour/releases/tag/v1.29.1)|x|[v1.29.1 report](./experimental-v1.29.1-default-report.yaml)|
|experimental|[v1.29.2](https://github.com/projectcontour/contour/releases/tag/v1.29.2)|x|[v1.29.2 report](./experimental-v1.29.2-default-report.yaml)|

## Reproduce

### Prerequisites

Follow the Contour [contribution guide][0] documentation for setting up your local development environment, which includes ensuring `kubectl`, `docker`, `kinD`, and other tools are installed.

### Steps

1. Clone the Contour GitHub repository

   ```bash
   git clone https://github.com/projectcontour/contour && cd contour
   ```

2. Check out the desired version

   ```bash
   export VERSION=v<x.y.z>
   git checkout $VERSION
   ```

3. Run the conformance tests

   ```bash
   export CONTOUR_E2E_IMAGE="ghcr.io/projectcontour/contour:$VERSION"
   export GENERATE_GATEWAY_CONFORMANCE_REPORT="true"
   make setup-kind-cluster run-gateway-conformance cleanup-kind
   ```

   Note: you can omit the `cleanup-kind` target if you would prefer to keep the `kinD` cluster.

4. Check the produced report

   ```bash
   cat gateway-conformance-report/projectcontour-contour-*.yaml
   ```

   Note: you can set `GATEWAY_CONFORMANCE_REPORT_OUTDIR` before running the tests to customize the output location.

[0]: https://github.com/projectcontour/contour/blob/main/CONTRIBUTING.md#building-from-source
