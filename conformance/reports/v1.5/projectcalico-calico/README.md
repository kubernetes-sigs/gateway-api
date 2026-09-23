# Calico

## Table of Contents

| API channel  | Implementation version                                                  | Mode    | Report                                              |
|--------------|-------------------------------------------------------------------------|---------|-----------------------------------------------------|
| experimental | [v3.32.2](https://github.com/projectcalico/calico/releases/tag/v3.32.2) | default | [link](./experimental-v3.32.2-default-report.yaml)  |

## Reproduce

The conformance runner is not in the v3.32.2 tag, so step 2 takes it from
`release-v3.33` (also Gateway API v1.5.1). Everything under test is built
from v3.32.2.

1. Clone Calico:

   ```bash
   git clone --branch v3.32.2 https://github.com/projectcalico/calico && cd calico
   ```

2. Add the runner:

   ```bash
   git fetch origin release-v3.33
   git checkout FETCH_HEAD -- e2e/cmd/gateway hack/test/kind/gateway-setup.sh
   # In e2e/cmd/gateway/e2e_test.go, move BackendTLSPolicySANValidation from
   # the curated SkipTests to skipFeatures, so it is reported as unsupported
   # (https://github.com/envoyproxy/gateway/issues/9687).
   go get sigs.k8s.io/gateway-api@v1.5.1 sigs.k8s.io/gateway-api/conformance@v1.5.1
   go test ./e2e/cmd/gateway -c -o e2e/bin/gateway/e2e.test
   ```

3. Create the kind cluster, then build the three `third_party/envoy-*`
   images and load them into kind as `calico/envoy-*:test-build`:

   ```bash
   make kind-up OPERATOR_BRANCH=v1.42.6
   ```

4. Fix up the cluster for this tag:

   - Scale `tigera-operator` back to 1 replica.
   - Replace MetalLB with `hack/test/kind/infra/metallb.yaml` from
     `release-v3.33`, and set the kube-controllers load balancer to
     `RequestedServicesOnly`.
   - Add `tigera-gateway` to the `gateway-conformance` pool's
     `serviceAllocation.namespaces`.

5. Set up the gateway and run the tests:

   ```bash
   export KUBECONFIG=hack/test/kind/kind-kubeconfig.yaml
   GATEWAY_API_CR=e2e/cmd/gateway/manifests/gatewayapi.yaml \
   GATEWAY_ENVOY_PROXY=e2e/cmd/gateway/manifests/envoyproxy.yaml \
   GATEWAY_METALLB_POOL=e2e/cmd/gateway/manifests/metallb-pool.yaml \
   GATEWAY_CLASS_NAME=tigera-gateway-class \
     hack/test/kind/gateway-setup.sh
   ./e2e/bin/gateway/e2e.test -gateway-class=tigera-gateway-class \
     -curated=envoy-gateway -mode=default \
     -organization=projectcalico -project=calico \
     -url=https://github.com/projectcalico/calico \
     -contact=https://www.tigera.io/contact/ -version=v3.32.2 \
     -report-output=report/gateway-conformance-report.yaml
   ```
