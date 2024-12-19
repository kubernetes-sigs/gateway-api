# Gloo Gateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [v1.18.0](https://github.com/solo-io/gloo/releases/tag/v1.18.0) | default | [Link](./experimental-v1.18.0-report.yaml) |

## Reproduce

> Note: [this is a mirror of the steps outlined in the conformance GHA workflow](https://github.com/solo-io/gloo/blob/main/.github/workflows/composite-actions/kube-gateway-api-conformance-tests/action.yaml).

### Prerequisites

In order to properly run the conformance tests, you need to have [KinD](https://github.com/kubernetes-sigs/kind)
and [Helm](https://github.com/helm/helm) available in your environment.
To run the conformance tests suite we will first create a KinD cluster, use Helm to deploy Gloo Gateway and finally
run the conformance tests.

### Steps

1. Clone the Gloo Gateway repository and checkout the release:

   ```sh
   git clone https://github.com/solo-io/gloo.git
   cd gloo
   export VERSION="v1.18.0"
   git checkout tags/$VERSION -b $VERSION
   ```

2. Create the KinD cluster with [MetalLB](https://metallb.io/):

   ```sh
   SKIP_DOCKER=true CONFORMANCE=true ci/kind/setup-kind.sh
   ```

3. Deploy Gloo Gateway Helm chart:
   ```sh
   helm upgrade -i --create-namespace -n gloo-system gloo gloo/gloo --version $VERSION --set kubeGateway.enabled=true
   ```

4. Run the conformance tests

   ```sh
   make conformance
   ```

5. View and verify the conformance report: `cat _test/conformance/$VERSION-report.yaml`
