# Airlock Microgateway

## Table of contents

| API channel  | Implementation version                                               | Mode    | Report                                           |
|--------------|----------------------------------------------------------------------|---------|--------------------------------------------------|
| experimental | [v4.4.0](https://github.com/airlock/microgateway/releases/tag/4.4.0) | default | [link](./experimental-4.4.0-default-report.yaml) |
| standard     | [v4.5.0](https://github.com/airlock/microgateway/releases/tag/4.5.0) | default | [link](./standard-4.5.0-default-report.yaml)     |
| standard    | [v4.6.0](https://github.com/airlock/microgateway/releases/tag/4.6.0) | default | [link](./standard-4.6.0-default-report.yaml) |
| standard    | [v4.7.0](https://github.com/airlock/microgateway/releases/tag/4.7.0) | default | [link](./standard-4.7.0-default-report.yaml) |

## Reproduce

The Airlock Microgateway conformance report can be reproduced by following the steps in the [Gateway API conformance guide](https://github.com/airlock/microgateway/tree/main/gateway-api/conformance/conformance.md) on GitHub.

> [!NOTE]
> The `HTTPRouteWeight` test fires 10 concurrent request to 3 backends totaling in 500 requests to assert a distribution that matches the configured weight.
> Please be aware that this test exceeds the [5 req/sec rate-limit](https://docs.airlock.com/microgateway/latest/?topic=MGW-00000056) enforced in the <!-- markdown-link-check-disable --> [community edition](https://www.airlock.com/en/secure-access-hub/components/microgateway/community-edition) <!-- markdown-link-check-enable -->, causing the test to fail.
> To successfully pass this test a <!-- markdown-link-check-disable --> [premium license](https://www.airlock.com/en/secure-access-hub/components/microgateway/premium-edition)  <!-- markdown-link-check-enable --> is required.
> 
> The Airlock Microgateway drops all request headers except for a well-known built-in standard and tracing headers list (e.g., Accept, Cookie, X-CSRF-TOKEN) to reduce the attack surface.
> Therefore, to run the conformance tests, a `ContentSecurityPolicy` with a `HeaderRewrites` (see [`conformance-report.yaml`](https://github.com/airlock/microgateway/tree/main/gateway-api/conformance/manifests/conformance-report.yaml)) is required to disable request header filtering for all `HTTPRoute` tests relying on the `MakeRequestAndExpectEventuallyConsistentResponse` assertion.
> Regardless of whether request header filtering is enabled or disabled, header-based routing works as specified in the Gateway API, as the headers are only filtered before the request is forwarded to the upstream.
