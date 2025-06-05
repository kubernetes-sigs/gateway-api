# Linkerd

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [version-2.18](https://github.com/linkerd/linkerd2/releases/tag/version-2.18/) | default | [version-2.18 report](./version-2.18.yaml) |

### A note on Linkerd Versioning

The Linkerd project publishes and announces _versions_ that correspond to
specific project milestones and sets of new features. The current version is
Linkerd 2.18.

Linkerd versions are available in different types of _release artifacts_:

- _Edge releases_ are published on a weekly or near-weekly basis by the
  Linkerd open-source project. Their names are `edge-y.m.n`, where `y` is the
  two-digit year, `m` is the numeric month, and `n` is the number of the edge
  release in that month (e.g. `edge-25.5.1` is the first edge release in May
  of 2025).

  Each major version of Linkerd has a corresponding edge release, indicated by
  a `version-2.X` tag -- for example, Linkerd 2.18 corresponds to
  `edge-25.4.4`, and therefore the `version-2.18` tag and the `edge-25.4.4`
  tag are on the same commit.

- _Stable releases_ of Linkerd follow semantic versioning, and are published
  by the vendor community around Linkerd.

For more information on Linkerd versioning, see the Linkerd [Releases and
Versions] documentation.

Since Gateway API conformance tests _require_ semantic versioning for the
implementation version, the Linkerd project reports conformance using the
`version` tags. However, the reproduction instructions below reference the
corresponding `edge` tag to match the way the Linkerd CLI is published.

[Releases and Versions]: https://linkerd.io/releases/

## Reproduce

To reproduce a Linkerd conformance test report:

0. `cd` to the top level of this repository.

1. Create an empty cluster.

2. Install the Linkerd CLI:

    ```bash
    curl --proto '=https' --tlsv1.2 -sSfL \
         https://run.linkerd.io/install-edge \
         | env LINKERD2_VERSION=edge-25.4.4 sh
    ```

3. Install the Gateway API CRDs:

    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.3.0/standard-install.yaml
    ```

4. Install the Linkerd control plane:

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
         --organization Linkerd \
         --project Linkerd \
         --url https://github.com/linkerd/linkerd2 \
         --version version-2.18 \
         --contact https://github.com/linkerd/linkerd2/blob/main/MAINTAINERS.md \
         --report-output version-2.18.yaml
    ```
