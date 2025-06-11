# Linkerd

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| standard     | [version-2.18](https://github.com/linkerd/linkerd2/releases/tag/version-2.18/) | default | [version-2.18 report](./standard-2.18-default-report.yaml) |

## Notes

This report uses the v1.3.0 Gateway API CRDs, but was run using the tests on
the `main` branch at commit `6cd1558a9e`, in order to take advantage of more
effective tests for the `MESH` conformance profile that landed after v1.3.0
was cut.

### Linkerd Versioning

The Linkerd project publishes and announces _versions_ that correspond to
specific project milestones and sets of new features. This report is for
Linkerd 2.18.

Linkerd versions are available in different types of _release artifacts_:

- _Edge releases_ are published on a weekly or near-weekly basis by the
  Linkerd open-source project. Their names are `edge-y.m.n`, where `y` is the
  two-digit year, `m` is the numeric month, and `n` is the number of the edge
  release in that month (e.g. `edge-25.5.1` is the first edge release in May
  of 2025).

  Each major version of Linkerd has a corresponding edge release, indicated by
  a `version-2.X` tag. For example, Linkerd 2.18 corresponds to `edge-25.4.4`,
  and therefore the `version-2.18` tag and the `edge-25.4.4` tag are on the
  same commit.

- _Stable releases_ of Linkerd follow semantic versioning, and are published
  by the vendor community around Linkerd.

For more information on Linkerd versioning, see the Linkerd [Releases and
Versions] documentation.

Since Gateway API conformance tests require semantic versioning for the
implementation version, the Linkerd project reports conformance using the
`version` tags. However, the `run_conformance.sh` script referenced below
installs the corresponding `edge` tag, because the Linkerd CLI is actually
published using the `edge` tag.

[Releases and Versions]: https://linkerd.io/releases/

## Reproduce

To reproduce a Linkerd conformance test report:

0. `cd` to the top level of this repository.

1. Create an empty cluster.

2. Run `bash conformance/reports/v1.3.0/linkerd-linkerd/run-conformance.sh`.

   You can set `LINKERD_VERSION`, `LINKERD_EDGE_VERSION`,
   `GATEWAY_API_CHANNEL`, and `GATEWAY_API_VERSION` if you want to try
   different versions of things. (Note that if you set `GATEWAY_API_VERSION`,
   you'll need to be on a matching Gateway API branch.)

3. The conformance report will be written to the
   `conformance/reports/v1.3.0/linkerd-linkerd/` directory.
