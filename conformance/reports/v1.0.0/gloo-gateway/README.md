# Gloo Gateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|x|[v1.17.4](https://github.com/solo-io/gloo/releases/tag/v1.17.4)|x|[v1.17.4 report](./v1.17.4-report.yaml)|

## Reproduce

> Note: [this is a mirror of the steps outlined in the conformance GHA workflow](https://github.com/solo-io/gloo/blob/main/.github/workflows/composite-actions/kube-gateway-api-conformance-tests/action.yaml).

1. Checkout the repository: `git clone https://github.com/solo-io/gloo && cd gloo`
2. Export the VERSION environment variable: `export VERSION="1.17.4"`
3. Create a kind cluster: `SKIP_DOCKER="true" CONFORMANCE="true" IMAGE_VARIANT="standard" CLUSTER_NODE_VERSION="v1.29.2@sha256:51a1434a5397193442f0be2a297b488b6c919ce8a3931be0ce822606ea5ca245" ./ci/kind/setup-kind.sh`
4. Install the Gloo Gateway helm chart with the Gateway API extension enabled: `helm upgrade -i --create-namespace -n gloo-system gloo gloo/gloo --version $VERSION --set kubeGateway.enabled=true`
5. Run the conformance suite locally: `make conformance-experimental`
6. Verify the conformance report: `cat _test/conformance/1.17.4-report.yaml`
