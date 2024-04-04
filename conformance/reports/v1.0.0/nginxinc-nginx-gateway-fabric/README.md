# Nginxinc NGINX Gateway Fabric

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|x|[v1.1.0](https://github.com/nginxinc/nginx-gateway-fabric/releases/tag/v1.1.0)|x|[link](./v1.1.0-report.yaml)|
|x|[v1.2.0](https://github.com/nginxinc/nginx-gateway-fabric/releases/tag/v1.2.0)|x|[link](./v1.2.0-report.yaml)|

## Reproduce

To reproduce results, clone the NGF repository:

```shell
git clone https://github.com/nginxinc/nginx-gateway-fabric.git && cd nginx-gateway-fabric/conformance
```

Follow the steps in the [Conformance README](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/conformance/README.md). If you are running tests on the `edge` version, then you don't need to build any images. Otherwise, you'll need to check out the specific release tag that you want to test, and then build and load the images onto your cluster, per the steps in the README.

After running, see the conformance report:

```shell
cat conformance-profile.yaml
```
