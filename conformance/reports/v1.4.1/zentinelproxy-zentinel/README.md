# Zentinel

[Zentinel](https://github.com/zentinelproxy/zentinel) is a security-first reverse proxy built on Cloudflare's Pingora framework. It emphasizes predictability, transparency, and operational simplicity.

## Table of Contents

| API channel | Implementation version                                                        | Mode    | Report                                                         |
|-------------|-------------------------------------------------------------------------------|---------|----------------------------------------------------------------|
| standard    | [v0.6.1](https://github.com/zentinelproxy/zentinel/releases/tag/v0.6.1)      | default | [v0.6.1 report](./standard-v0.6.1-default-report.yaml)        |

## Reproduce

Clone the Zentinel repository and run the conformance test script:

```shell
git clone https://github.com/zentinelproxy/zentinel.git && cd zentinel
./scripts/conformance-test.sh --report
```

Prerequisites: Docker, kind, kubectl, helm, Go 1.22+.

The script creates a kind cluster, builds the gateway controller and proxy images, installs Gateway API CRDs (v1.4.1), deploys Zentinel via Helm, and runs the official conformance suite. The report is written to `conformance/reports/standard-v0.6.1-default-report.yaml`.
