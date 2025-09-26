# Cloud Provider KIND

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|standard|[v0.8.0-alpha.1]|default|[report](./standard-v0.8.0-alpha.1-default-report.yaml)|

## Reproduce

1. `[Install `cloud-provider-kind`](https://github.com/kubernetes-sigs/cloud-provider-kind/tree/v0.8.0-alpha.1?tab=readme-ov-file#install)

2. [Run a `KIND` cluster](https://kind.sigs.k8s.io/docs/user/quick-start/)

3. [Start the `cloud-provider-kind`](https://github.com/kubernetes-sigs/cloud-provider-kind/tree/v0.8.0-alpha.1?tab=readme-ov-file#gateway-api-support-alpha)

4. Run the conformance tests:

```sh
go test ./conformance -run TestConformance \
  --report-output /tmp/report.yaml \
  --organization=sigs.k8s.io \
  --project=cloud-provider-kind \
  --url=https://github.com/kubernetes-sigs/cloud-provider-kind \
  --version=v0.8.0-alpha.1 \
  --contact=https://github.com/kubernetes-sigs/cloud-provider-kind/issues/new \
  --gateway-class=cloud-provider-kind \
  --conformance-profiles=GATEWAY-HTTP \
  --supported-features=Gateway,HTTPRoute,ReferenceGrant
```