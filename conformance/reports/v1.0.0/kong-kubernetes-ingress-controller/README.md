# Kong Kubernetes Ingress Controller

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|x|[v3.0.2](https://github.com/Kong/kubernetes-ingress-controller/releases/tag/v3.0.2)|x|[link](./v3.0.2-report.yaml)|
|x|[v3.1.1](https://github.com/Kong/kubernetes-ingress-controller/releases/tag/v3.1.1)|x|[link](./v3.1.1-report.yaml)|

## Reproduce

### Prerequisites

In order to properly run the conformance tests, you need to have [KinD](https://github.com/kubernetes-sigs/kind)
and [Helm](https://github.com/helm/helm) installed on your local machine, as the
test suite will create a fresh KinD cluster and will use Helm to deploy some additional
components.

### Steps

1. Clone the Kong Ingress Controller GitHub repository

   ```bash
   git clone https://github.com/Kong/kubernetes-ingress-controller.git && cd kubernetes-ingress-controller
   ```

2. Check out the desired version

   ```bash
   export VERSION=v<x.y.z>
   git checkout $VERSION
   ```

3. Run the conformance tests

   ```bash
   KONG_TEST_EXPRESSION_ROUTES=true make test.conformance
   ```

4. Check the produced report

   ```bash
   cat ./kong-kubernetes-ingress-controller.yaml
   ```
