# Projectcontour Contour

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v1.30.0](https://github.com/projectcontour/contour/releases/tag/v1.30.0)|x|[link](./experimental-v1.30.0-default-report.yaml)|

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
