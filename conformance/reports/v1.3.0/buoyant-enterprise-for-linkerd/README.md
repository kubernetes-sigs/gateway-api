# Buoyant Enterprise for Linkerd

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [enterprise-2.18](https://docs.buoyant.io/buoyant-enterprise-linkerd/latest/overview//) | default | [enterprise-2.18 report](./enterprise-2.18.yaml) |

## Reproduce

To reproduce a Buoyant Enterprise for Linkerd conformance test report:

0. `cd` to the top level of this repository.

1. Create an empty cluster.

2. Install the Linkerd CLI:

    ```bash
    curl --proto '=https' --tlsv1.2 -sSfL \
         https://enterprise.buoyant.io/install \
         | env LINKERD2_VERSION=enterprise-2.18 sh
    ```

3. Install the Gateway API CRDs:

    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
    ```

4. Install the Buoyant Enterprise for Linkerd control plane:

    ```bash
    linkerd install --crds | kubectl apply -f -
    linkerd install | kubectl apply -f -
    linkerd check
    ```

5. Run the conformance tests:

    ```bash
    go test \
       -p 4 \
       ./conformance \
       -run TestConformance \
       -args \
         --conformance-profiles MESH-HTTP,MESH-GRPC \
         --namespace-annotations=linkerd.io/inject=enabled \
         --exempt-features=Gateway,ReferenceGrant \
         --organization Buoyant \
         --project "Buoyant Enterprise for Linkerd" \
         --url https://buoyant.io/ \
         --version enterprise-2.18 \
         --contact "gateway-api@buoyant.io" \
         --report-output bel-2.18.yaml
    ```
