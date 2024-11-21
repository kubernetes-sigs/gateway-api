# Kong Gateway Operator

## Table of Contents

| API channel  | Implementation version                                                              | Mode        | Report                                                |
|--------------|-------------------------------------------------------------------------------------|-------------|-------------------------------------------------------|
| standard | [v1.4.0](https://github.com/Kong/gateway-operator/releases/tag/v1.4.0) | expressions | [link](./standard-v1.4.0-expressions-report.yaml) |

## Reproduce

### Prerequisites

In order to properly run the conformance tests, you need to have [KinD](https://github.com/kubernetes-sigs/kind)
and [Helm](https://github.com/helm/helm) installed on your local machine, as the
test suite will create a fresh KinD cluster and will use Helm to deploy some additional
components.

### Steps

1. Clone the Kong Gateway Operator GitHub repository

   ```bash
   git clone https://github.com/Kong/gateway-operator.git && cd gateway-operator
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
   cat ./standard-${VERSION}-expressions-report.yaml
   ```
