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
| experimental | [v3.32.0](https://github.com/projectcalico/calico/releases/tag/v3.32.0) | default | [link](./experimental-v3.32.0-default-report.yaml)  |

The submitted report covers the GATEWAY-HTTP, GATEWAY-GRPC, and
GATEWAY-TLS profiles. All core and extended tests in those profiles
passed (Failed: 0, Skipped: 0).

## Reproduce

1. Clone the Calico repository:

   ```bash
   git clone https://github.com/projectcalico/calico && cd calico
   ```

2. Run the conformance suite:

   ```bash
   make e2e-test-gateway-conformance \
     GATEWAY_CONFORMANCE_VERSION=v3.32.0 \
     GATEWAY_CONFORMANCE_CONTACT=https://www.tigera.io/contact/
   ```

3. Inspect the produced report:

   ```bash
   cat report/gateway-conformance-report.yaml
   ```

> **Note**: The conformance runner landed on `master` after the v3.32.0
> tag was cut. The Envoy Gateway version (`v1.7.0`) and Gateway API
> channel (`experimental`, `v1.4.1`) bundled by tigera-operator on
> `master` match what ships in the v3.32.0 release; the version metadata
> in the report is pinned via `GATEWAY_CONFORMANCE_VERSION`.

## Contact

Questions and issues are tracked in the [Calico project][calico]. For
commercial and maintainer inquiries, see the [Tigera contact form][tigera-contact].

[calico]: https://github.com/projectcalico/calico
[calico-operator]: https://github.com/tigera/operator
[envoy-gateway]: https://gateway.envoyproxy.io
[tigera-contact]: https://www.tigera.io/contact/
