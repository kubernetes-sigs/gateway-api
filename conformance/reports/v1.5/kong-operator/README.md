# Kong Operator

## Table of Contents

| API channel  | Implementation version                                              | Mode                   | Report                                                                                                          |
|--------------|---------------------------------------------------------------------|------------------------|-----------------------------------------------------------------------------------------------------------------|
| experimental | [v2.2.0](https://github.com/Kong/kong-operator/releases/tag/v2.2.0) | expressions            | [v2.2.0 expressions report](./experimental-v2.2.0-expressions-standard-report.yaml)                             |
| experimental | [v2.2.0](https://github.com/Kong/kong-operator/releases/tag/v2.2.0) | traditional_compatible | [v2.2.0 traditional compatible standard report](./experimental-v2.2.0-traditional_compatible-standard-report.yaml) |
| experimental | [v2.2.0](https://github.com/Kong/kong-operator/releases/tag/v2.2.0) | traditional_compatible | [v2.2.0 traditional compatible hybrid report](./experimental-v2.2.0-traditional_compatible-hybrid-report.yaml)    |

## Reproduce

### Prerequisites

In order to properly run the conformance tests, you need to have the following
tools installed on your local machine:
- [KinD](https://github.com/kubernetes-sigs/kind)
- [Helm](https://github.com/helm/helm)
- [mise](https://github.com/jdx/mise)

The test suite will create a fresh KinD cluster and will use Helm to deploy some additional
components.

### Steps

1. Clone the Kong Operator GitHub repository

   ```bash
   git clone https://github.com/kong/kong-operator.git && cd kong-operator
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
   cat ./*report.yaml
   ```
