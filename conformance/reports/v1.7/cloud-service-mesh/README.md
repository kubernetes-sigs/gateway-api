# Google Cloud Service Mesh

## Table of Contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
| standard | [1.37.1-gke.2097000](https://cloud.google.com/kubernetes-engine/docs/release-notes-rapid) | default | [1.37.1-gke.2097000 report](./standard-1.37.1-gke.2097000-default-report.yaml) |

## Reproduce

The Cloud Service Mesh conformance report can be reproduced with the following steps.

1. Prepare a Google Cloud project (enable required APIs, enable the fleet `servicemesh` feature, and grant Workload Identity access to Traffic Director) following [Prepare Cloud Service Mesh with the Gateway API](https://cloud.google.com/service-mesh/docs/gateway/prepare-gateway).

2. Create a GKE cluster without `--gateway-api` first so the Gateway API `v1.7.0` CRDs can be installed before enabling the GKE Gateway API add-on (GKE `1.37.1-gke.2097000` bundles `v1.6.2` CRDs by default):

```sh
gcloud container clusters create "${CLUSTER_NAME}" \
    --project="${PROJECT_ID}" \
    --location="${LOCATION}" \
    --release-channel=rapid \
    --cluster-version=1.37.1-gke.2097000 \
    --machine-type=e2-standard-4 \
    --num-nodes=3 \
    --workload-pool="${PROJECT_ID}.svc.id.goog" \
    --enable-fleet
```

3. Install the Gateway API `v1.7.0` standard channel CRDs, configure Cloud Service Mesh on the cluster's fleet membership to use the Gateway API, and enable `--gateway-api=standard` on the cluster (GKE keeps the newer `v1.7.0` CRDs):

```sh
kubectl apply --server-side \
    -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.7.0/standard-install.yaml

gcloud container fleet mesh update \
    --project="${PROJECT_ID}" \
    --location="${FLEET_LOCATION}" \
    --memberships="${CLUSTER_NAME}" \
    --config-api=gateway

gcloud container clusters update "${CLUSTER_NAME}" \
    --project="${PROJECT_ID}" \
    --location="${LOCATION}" \
    --gateway-api=standard
```

Wait until `kubectl get controlplanerevision csm-managed-gateway -n istio-system` reports `Reconciled=True`.

4. Run the conformance suite from a checkout of the [Kubernetes Gateway API repository](https://github.com/kubernetes-sigs/gateway-api) at tag `v1.7.0`:

```sh
git checkout v1.7.0

go test ./conformance -run TestConformance -v -timeout=360m -args \
    --conformance-profiles=MESH-HTTP \
    --supported-features=Mesh,MeshHTTPRouteBackendRequestHeaderModification,MeshHTTPRouteQueryParamMatching,MeshHTTPRouteRedirectPath,MeshHTTPRouteRedirectPort,MeshHTTPRouteSchemeRedirect,MeshHTTPRouteNamedRouteRule \
    --namespace-labels=mesh.cloud.google.com/csm-injection=sidecar \
    --skip-tests=MeshHTTPRouteMatching \
    --cleanup-base-resources=true \
    --debug \
    --organization="Google Cloud" \
    --project="Cloud Service Mesh" \
    --url=https://cloud.google.com/service-mesh/docs/gateway/overview \
    --version=1.37.1-gke.2097000 \
    --contact=csm-gateway-api@google.com \
    --report-output="./standard-1.37.1-gke.2097000-default-report.yaml" \
    "--timeout-config-overrides=DefaultTestTimeout:900;HTTPRouteMustHaveCondition:300;LatestObservedGenerationSet:300;MaxTimeToConsistency:1200;NamespacesMustBeReady:600;RequestTimeout:20;RouteMustHaveParents:600"
```

Note: Managed Cloud Service Mesh takes several minutes to program new routes, so `--timeout-config-overrides` sets `MaxTimeToConsistency` to 1200 seconds. `MeshHTTPRouteWeight` can occasionally fail on a warm mesh if its initial reachability check passes before the weighted split route finishes programming.
