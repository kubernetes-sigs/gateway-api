# Kong Kubernetes Ingress Controller

## Table of Contents

| API channel  | Implementation version                                                              | Mode        | Report                                                |
|--------------|-------------------------------------------------------------------------------------|-------------|-------------------------------------------------------|
| experimental | [v3.4.0](https://github.com/Kong/kubernetes-ingress-controller/releases/tag/v3.4.0) | expressions | [v3.4.0 expressions report](./experimental-v3.4.0-expressions-report.yaml) |
| experimental | [v3.4.0](https://github.com/Kong/kubernetes-ingress-controller/releases/tag/v3.4.0) | traditional_compatible | [v3.4.0 traditional compatible report](./experimental-v3.4.0-traditional_compatible-report.yaml) |

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
   TEST_KONG_ROUTER_FLAVOR=<traditional_compatible|expressions> make test.conformance
   ```

4. Check the produced report

   ```bash
   cat ./kong-kubernetes-ingress-controller.yaml
   ```
