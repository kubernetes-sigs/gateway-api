# Buoyant Enterprise for Linkerd

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [enterprise-2.18](https://docs.buoyant.io/buoyant-enterprise-linkerd/latest/overview//) | default | [enterprise-2.18 report](./standard-2.18-default-report.yaml) |

## Reproduce

To reproduce a Buoyant Enterprise for Linkerd conformance test report:

0. `cd` to the top level of this repository.

1. Create an empty cluster.

2. Run `bash conformance/reports/v1.3.0/buoyant-enterprise-for-linkerd/run-conformance.sh`.

   You can set `LINKERD_VERSION`, `GATEWAY_API_CHANNEL`, and
   `GATEWAY_API_VERSION` if you want to try different versions of things.
   (Note that if you set `GATEWAY_API_VERSION`, you'll need to be on a
   matching Gateway API branch.)

3. The conformance report will be written to the
   `conformance/reports/v1.3.0/buoyant-enterprise-for-linkerd/` directory.
