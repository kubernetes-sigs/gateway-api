# Higress

[Higress](https://higress.ai/) is a cloud-native API gateway built on Istio and
Envoy and is a CNCF project.

## Table of Contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
| standard | [v2.2.4](https://github.com/higress-group/higress/releases/tag/v2.2.4) | default | [v2.2.4 report](./standard-v2.2.4-default-report.yaml) |

## Reproduce

Check out the Higress v2.2.4 release, create a Kubernetes cluster, and install
the Gateway API v1.6.0 standard CRDs. Install `helm/core` with the v2.2.4
controller, Pilot, and Gateway images, with the Gateway API deployment
controller enabled.

Run the official Gateway API v1.6.0 conformance suite through the Higress
wrapper:

```sh
GATEWAY_CLASS=higress \
GATEWAY_API_VERSION=v1.6.0 \
GATEWAY_CONFORMANCE_TEST_DIR=test/gateway/v1.6 \
GATEWAY_CONFORMANCE_SUPPORTED_FEATURES=Gateway,HTTPRoute,ReferenceGrant \
GATEWAY_CONFORMANCE_PROFILE=GATEWAY-HTTP \
GATEWAY_CONFORMANCE_REPORT=out/gateway-api-v1.6.0-report.yaml \
GATEWAY_CONFORMANCE_CONTACT=https://github.com/higress-group/higress/issues \
HIGRESS_CONFORMANCE_VERSION=v2.2.4 \
GATEWAY_CONFORMANCE_ALLOW_CRDS_MISMATCH=false \
GATEWAY_CONFORMANCE_SUPPORTS_TEST_CLEANUP=true \
GATEWAY_CONFORMANCE_CLEANUP_TEST_RESOURCES=true \
tools/hack/run-gateway-api-conformance.sh
```

The test uses the release images referenced by the Higress v2.2.4 Helm chart.
