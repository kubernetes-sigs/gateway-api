# kgateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [v2.0.0](https://github.com/kgateway-dev/kgateway/releases/tag/v2.0.0) | default | [Link](./v2.0.0-report.yaml) |

## Reproduce

> Note: [this is a mirror of the steps outlined in the conformance GHA workflow](https://github.com/kgateway-dev/kgateway/blob/v2.0.0/.github/actions/kube-gateway-api-conformance-tests/action.yaml).

### Prerequisites

In order to properly run the conformance tests, you need to have [KinD](https://github.com/kubernetes-sigs/kind)
and [Helm](https://github.com/helm/helm) available in your environment.
To run the conformance tests suite we will first create a KinD cluster, use Helm to deploy kgateway and finally
run the conformance tests.

### Steps

1. Clone the kgateway repository and checkout the release:

   ```sh
   git clone https://github.com/kgateway-dev/kgateway.git
   cd kgateway
   export VERSION="v2.0.0"
   git checkout tags/$VERSION -b $VERSION
   ```

2. Create the KinD cluster with [MetalLB](https://metallb.io/):

   ```sh
   SKIP_DOCKER=true CONFORMANCE=true ci/kind/setup-kind.sh
   ```

3. Deploy kgateway Helm charts:
   ```sh
   helm upgrade -i --create-namespace --namespace kgateway-system --version $VERSION kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds

    helm upgrade -i --namespace kgateway-system --version $VERSION kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway
   ```

4. Run the conformance tests

   ```sh
   make conformance
   ```

5. View and verify the conformance report: `cat _test/conformance/$VERSION-report.yaml`