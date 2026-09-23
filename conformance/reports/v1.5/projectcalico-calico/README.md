# Calico

[Calico][calico] is an open-source networking and security solution for
Kubernetes and other cloud-native environments. Calico's Gateway API
implementation is built on the [tigera-operator][calico-operator] and
[Envoy Gateway][envoy-gateway]: the operator reconciles a `GatewayAPI`
custom resource, provisions an Envoy Gateway control plane, and creates
a default `tigera-gateway-class` GatewayClass on the cluster. The
underlying Envoy data plane is integrated with Calico's native
networking and policy.

## Table of Contents

| API channel  | Implementation version                                                  | Mode    | Report                                              |
|--------------|-------------------------------------------------------------------------|---------|-----------------------------------------------------|
| experimental | [v3.32.2](https://github.com/projectcalico/calico/releases/tag/v3.32.2) | default | [link](./experimental-v3.32.2-default-report.yaml)  |

The report has three profiles: GATEWAY-HTTP, GATEWAY-GRPC and
GATEWAY-TLS. Calico v3.32.2 ships tigera-operator v1.42.6 and Envoy
Gateway v1.8. Both use Gateway API v1.5.1.

## Reproduce

The v3.32.2 tag does not have the test runner. It came later. So step 2
copies it in from `release-v3.33`, which uses the same Gateway API
v1.5.1. Everything under test is built from the v3.32.2 tag.

1. Clone Calico at the release:

   ```bash
   git clone --branch v3.32.2 https://github.com/projectcalico/calico && cd calico
   ```

2. Add the runner:

   ```bash
   git fetch origin release-v3.33
   git checkout FETCH_HEAD -- e2e/cmd/gateway hack/test/kind/gateway-setup.sh
   go get sigs.k8s.io/gateway-api@v1.5.1 sigs.k8s.io/gateway-api/conformance@v1.5.1
   go test ./e2e/cmd/gateway -c -o e2e/bin/gateway/e2e.test
   ```

3. Make a kind cluster with the v3.32.2 images and operator v1.42.6:

   ```bash
   make kind-up OPERATOR_BRANCH=v1.42.6
   ```

   Then build the three `third_party/envoy-*` images from the tag and load
   them into kind as `calico/envoy-*:test-build`.

   Two test-setup fixes are needed on this older tag. First, `kind-up`
   scales `tigera-operator` to 0 at the end, so scale it back to 1. Second,
   swap its MetalLB v0.9.5 for the v0.14.9 L2 manifest in `release-v3.33`
   (`hack/test/kind/infra/metallb.yaml`), and set the kube-controllers
   load balancer to `RequestedServicesOnly`. MetalLB only hands out
   LoadBalancer IPs on the kind network. Operator v1.42.6 puts all Envoy
   Services in `tigera-gateway`, so add that namespace to the
   `gateway-conformance` pool's `serviceAllocation.namespaces`.

4. Set up the gateway, then run the suite:

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

## Contact

File issues in the [Calico project][calico]. For other questions, use the
[Tigera contact form][tigera-contact].

[calico]: https://github.com/projectcalico/calico
[calico-operator]: https://github.com/tigera/operator
[envoy-gateway]: https://gateway.envoyproxy.io
[tigera-contact]: https://www.tigera.io/contact/
