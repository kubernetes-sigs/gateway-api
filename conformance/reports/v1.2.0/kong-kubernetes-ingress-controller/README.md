# Kong Kubernetes Ingress Controller

## Table of Contents

| API channel  | Implementation version                                                              | Mode        | Report                                                |
|--------------|-------------------------------------------------------------------------------------|-------------|-------------------------------------------------------|
| experimental | [v3.2.0-244-gea4944bb0](https://github.com/Kong/kubernetes-ingress-controller/commit/b5c61d20e325055cd4665b4c29aeea5288c20d2a) | expressions | [link](./experimental-v3.2.0-244-gea4944bb0-expressions-report.yaml) |

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
