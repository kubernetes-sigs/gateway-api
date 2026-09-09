# Buoyant Enterprise for Linkerd

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [enterprise-2.20.2](https://docs.buoyant.io/buoyant-enterprise-linkerd/2.20/overview//) | default | [enterprise-2.20.2 report](./standard-2.20.2-default-report.yaml) |

## Notes

This report uses the v1.6.2 Gateway API CRDs, but was run using the tests
on the `main` branch at commit `629880ed4c`. This is because [PR#4934]
has a needed fix for mesh-only conformance -- but that PR failed its
cherry-pick into release-1.6, so it isn't included in the v1.6.2
conformance tests.

[PR#4934]: https://github.com/kubernetes-sigs/gateway-api/pull/4934

## Reproduce

To reproduce a Buoyant Enterprise for Linkerd conformance test report:

0. `cd` to the top level of this repository.

1. Create an empty cluster.

2. Run `bash conformance/reports/v1.6/buoyant-enterprise-for-linkerd/run-conformance.sh`.

   You can set `LINKERD_VERSION`, `GATEWAY_API_CHANNEL`, and
   `GATEWAY_API_VERSION` if you want to try different versions of things.
   (Note that if you set `GATEWAY_API_VERSION`, you'll need to be on a
   matching Gateway API branch.)

3. The conformance report will be written to the
   `conformance/reports/v1.6/buoyant-enterprise-for-linkerd/` directory.
