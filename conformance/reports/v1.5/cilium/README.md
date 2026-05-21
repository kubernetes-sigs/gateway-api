# Cilium

## Table of Contents

| API channel  | Implementation version                    | Mode    | Report                                                 |
|--------------|-------------------------------------------|---------|--------------------------------------------------------|
| experimental | [main](https://github.com/cilium/cilium/) | default | [v1.20.0-pre.2 report](./experimental-v1.20.0-pre.2-default-report.yaml) |

## Reproduce

Cilium conformance tests can be reproduced by the following steps from within the [Cilium repo](https://github.com/cilium/cilium).

1. Build a Kind cluster, and ensure Cilium is working. Cilium will install the checked-out version from the Cilium repo for you if you use the make target:

```sh
WAIT_DURATION=120s make kind-servicemesh-install-cilium-fast
```
(The WAIT_DURATION is there because pulling the images and building all the BPF code can take a while the first time.)

2. Run the conformance tests using the make target:

```sh
make gateway-api-conformance
```

This will run the conformance test using the currently-configured kubeconfig - so it will also work against a non-Kind cluster,
as long as the cluster has:

* Cilium installed with Gateway API enabled
* Loadbalancer Service support
