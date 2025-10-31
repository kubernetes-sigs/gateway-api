# GKE (Google Kubernetes Engine) Gateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|standard|1.34.1-gke.1829001|gke-l7-regional-external-managed|[v1.34.1 rxlb report](./standard-1.43.1-rxlb-report.yaml)|

## Reproduce

GKE Gateway conformance report can be reproduced by the following steps.

1. create a GKE cluster with Gateway API enabled (the minimum cluster version that supports v1.4.0 CRD is `1.34.1-gke.1829001`)

```
gcloud container clusters create "${cluster_name}" --gateway-api=standard --location="${location} --cluster-version={$version}"
```

2. create a proxy-only subnet if using a regional Gateway following [guide](https://cloud.google.com/kubernetes-engine/docs/how-to/deploying-gateways#configure_a_proxy-only_subnet)

3. run the following command from within the [GKE Gateway repo](https://github.com/GoogleCloudPlatform/gke-gateway-api)

```
go test ./conformance -run TestConformance -v -timeout=3h -args \
    --gateway-class=gke-l7-regional-external-managed \
    --conformance-profiles=GATEWAY-HTTP \
    --organization=GKE \
    --project=gke-gateway \
    --url=https://cloud.google.com/kubernetes-engine/docs/concepts/gateway-api \
    --version=1.34.1-gke.1829001 \
    --contact=gke-gateway-dev@google.com \
    --skip-tests=HTTPRouteHostnameIntersection \
    --supported-features=Gateway,HTTPRoute,GatewayPort8080,HTTPRouteHostRewrite,HTTPRoutePathRedirect,HTTPRouteRequestMirror,HTTPRouteRequestPercentageMirror,HTTPRouteResponseHeaderModification,HTTPRouteSchemeRedirect \
    --report-output="/path/to/report"
```

Note: the repro result can be flaky in some cases because the conformance framework doesn't isolate test cases enough.
