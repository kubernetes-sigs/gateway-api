# Airlock Microgateway

## Table of contents

| API channel  | Implementation version                                               | Mode    | Report                                           |
|--------------|----------------------------------------------------------------------|---------|--------------------------------------------------|
| experimental | [v4.4.0](https://github.com/airlock/microgateway/releases/tag/4.4.0) | default | [link](./experimental-4.4.0-default-report.yaml) |

## Reproduce

The Airlock Microgateway conformance report can be reproduced by following the steps in the [Gateway API conformance guide](https://github.com/airlock/microgateway/tree/main/examples/gateway-api/conformance/conformance.md) on GitHub.

> [!NOTE]
> The `HTTPRouteWeight` test fires 10 concurrent request to 3 backends totaling in 500 requests to assert a distribution that matches the configured weight.
> Please be aware that this test exceeds the [5 req/sec rate-limit](https://docs.airlock.com/microgateway/latest/#data/1675772882054.html) enforced in the <!-- markdown-link-check-disable --> [community edition](https://www.airlock.com/en/secure-access-hub/components/microgateway/community-edition) <!-- markdown-link-check-enable -->, causing the test to fail.
> To successfully pass this test a <!-- markdown-link-check-disable --> [premium license](https://www.airlock.com/en/secure-access-hub/components/microgateway/premium-edition)  <!-- markdown-link-check-enable --> is required.
