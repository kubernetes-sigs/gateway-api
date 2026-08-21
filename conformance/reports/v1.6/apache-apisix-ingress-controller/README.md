# Apache APISIX Ingress Controller

[Apache APISIX Ingress Controller](https://github.com/apache/apisix-ingress-controller)
configures [Apache APISIX](https://apisix.apache.org/) from Kubernetes
resources, and implements Gateway API alongside its own CRDs and Ingress.

## Table of Contents

| API channel  | Implementation version                                                              | Mode              | Report                                                                    |
|--------------|-------------------------------------------------------------------------------------|-------------------|---------------------------------------------------------------------------|
| experimental | [2.2.0](https://github.com/apache/apisix-ingress-controller/releases/tag/2.2.0)      | default           | [2.2.0 report](./experimental-2.2.0-default-report.yaml)                  |
| experimental | [2.2.0](https://github.com/apache/apisix-ingress-controller/releases/tag/2.2.0)      | apisix-standalone | [2.2.0 report](./experimental-2.2.0-apisix-standalone-report.yaml)        |

## Overview

The controller drives one shared APISIX data plane rather than provisioning one
per Gateway, and it supports two ways of configuring it. In the `default` mode
APISIX keeps its configuration in etcd and the controller writes through the
APISIX Admin API. In the `apisix-standalone` mode APISIX runs without etcd and
the controller pushes the whole configuration to it, which is the
[API-driven mode](https://apisix.apache.org/docs/apisix/deployment-modes/) APISIX
documents. The two modes run the same suite and are reported separately.

## Prerequisites

- [docker](https://docs.docker.com/get-started/get-docker/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [kind](https://github.com/kubernetes-sigs/kind)
- [go](https://go.dev/learn/)

Any cluster works as long as it supports LoadBalancer Services: the suite
reaches the gateway through the data plane Service's external address, and the
controller publishes that address in every Gateway's `status.addresses`. Steps 2
and 3 below only give a local kind cluster that capability.

## Reproduce

1. Clone the repository and check out the release

   ```shell
   git clone https://github.com/apache/apisix-ingress-controller.git && cd apisix-ingress-controller
   git checkout 2.2.0
   ```

   Checking out the release tag is what selects the published images for it and
   makes the report name that release. `make conformance-images` prints the
   three images the run deploys.

2. Create the cluster

   ```shell
   make kind-up
   ```

3. Run a local LoadBalancer provider

   ```shell
   make kind-lb
   ```

   This runs [cloud-provider-kind](https://kind.sigs.k8s.io/docs/user/loadbalancer)
   in the background. Skip it on a cluster that already has LoadBalancer
   support.

4. Install the Gateway API and the controller's CRDs

   ```shell
   make install
   ```

5. Run the suite

   ```shell
   make conformance-test
   ```

   For the `apisix-standalone` mode:

   ```shell
   PROVIDER_TYPE=apisix-standalone make conformance-test CONFORMANCE_MODE=apisix-standalone
   ```

6. Read the report

   ```shell
   cat "$(make -s conformance-report-path)"
   ```

   Pass the same `CONFORMANCE_MODE` to get the path of a standalone run. The
   file is named `<channel>-<version>-<mode>-report.yaml`, which is the name
   this repository expects, so it is submitted as produced.

The reports here came from the `APISIX Conformance Test` workflow run on the
2.2.0 tag, which runs exactly these steps against the published 2.2.0 images.

## Skipped tests

`test/conformance/conformance_test.go` lists every skip with its reason. In
short: four TLSRoute tests pin `mode: Passthrough`, which APISIX does not do
because it terminates TLS and matches stream routes by SNI;
`HTTPRouteMultipleGateways` needs a route served independently from each parent,
which one shared data plane does not do; `HTTPRouteHTTPSListener` and three
`HTTPRouteInvalidBackendRef*` tests are gaps tracked for follow-up.
