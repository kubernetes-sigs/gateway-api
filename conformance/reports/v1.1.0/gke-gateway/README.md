# GKE (Google Kubernetes Engine) Gateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|standard|1.30.3-gke.1211000|gke-l7-global-external-managed|[link](./standard-1.30.3-gxlb-report.yaml)|
|standard|1.30.3-gke.1211000|gke-l7-regional-external-managed|[link](./standard-1.30.3-rxlb-report.yaml)|
|standard|1.30.3-gke.1211000|gke-l7-rilb|[link](./standard-1.30.3-rilb-report.yaml)|

## Reproduce

GKE Gateway conformance report can be reproduced by the following steps.

1. create a GKE cluster with Gateway API enabled

```
gcloud container clusters create "${cluster_name}" --gateway-api=standard --location="${location}"
```

2. create a proxy-only subnet if using a regional Gateway following [guide](https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways#configure_a_proxy-only_subnet)

3. run the following command from within the [GKE Gateway repo](https://github.com/GoogleCloudPlatform/gke-gateway-api)

```
go test ./conformance -run TestConformance -v -timeout=3h -args \
    --gateway-class=gke-l7-global-external-managed \
    --conformance-profiles=GATEWAY-HTTP \
    --organization=GKE \
    --project=gke-gateway \
    --url=https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api \
    --version=1.30.3-gke.1211000 \
    --contact=gke-gateway-dev@google.com \
    --report-output="/path/to/report"
```

or run a single conformance test case

```
go test ./conformance -run TestConformance -v -args \
    --gateway-class=gke-l7-global-external-managed \
    --run-test=HTTPRouteRequestMirror
```

Note: the repro result can be flaky in some cases because the conformance framework doesn't isolate test cases enough.
See this [Issue](https://github.com/kubernetes-sigs/gateway-api/issues/3233) for more details.
The flakiness will be eliminated after this [PR](https://github.com/kubernetes-sigs/gateway-api/pull/3243) is merged and appropriate test isolation timeout is configured.
