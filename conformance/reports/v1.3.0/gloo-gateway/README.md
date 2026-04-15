# Gloo Gateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | https://github.com/solo-io/gloo/releases/tag/v1.20.12 | default | [Link](./v1.20.12-report.yaml) |

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
   export VERSION="v1.20.12"
   git checkout $VERSION

   ```

   Alternatively, download and extract the [v1.20.12 release source](https://github.com/solo-io/gloo/releases/tag/v1.20.12).

1. Create the Kind cluster with MetalLB:

   ```sh
   SKIP_DOCKER="true" CONFORMANCE=true CLUSTER_NODE_VERSION="v1.34.0@sha256:7416a61b42b1662ca6ca89f02028ac133a309a2a30ba309614e8ec94d976dc5a" ci/kind/setup-kind.sh
   ```

1. Deploy Gloo Gateway Helm chart:
   ```sh
   helm upgrade -i --create-namespace -n gloo-system gloo gloo/gloo --version $VERSION --set kubeGateway.enabled=true
   ```

1. Run the conformance tests

   ```sh
   make conformance
   ```

1. View and verify the conformance report: `cat _test/conformance/$VERSION-report.yaml`
