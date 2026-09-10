# WSO2 Gateway

## Table of contents

| API channel | Implementation version | Mode     | Report                                          |
|-------------|------------------------|-------------|-------------------------------------------------|
| standard    | [v1.2.0-alpha2](https://github.com/wso2/api-platform/releases/tag/gateway/v1.2.0-alpha2)        | default | [v1.2.0-alpha2 report](./v1.2.0-alpha2-report.yaml)  

## Steps to Reproduce

These steps build the WSO2 gateway images from source and run the Gateway API conformance suite against them, as required for report submission. 

Prerequisites: KinD, Helm, kubectl, Docker, `jq`, and a Go toolchain.

### 1. Clone the repository

```sh
git clone https://github.com/wso2/api-platform.git && cd api-platform && git checkout 827b87100a7d68c83861e96d64adacc1dc144574
```

### 2. Build the gateway images from source

Build the gateway-controller and gateway-runtime images:

```sh
cd gateway && make build
```

Then build the gateway-operator image:

```sh
cd ../kubernetes/gateway-operator && make docker-build
```

This will produce the relevant gateway components and operator images

### 3. Create the KinD cluster with MetalLB and load the images

```sh
cd ../conformance
./kind/setup-kind.sh          # macOS + Colima: use ./kind/setup-colima.sh instead
./load-images.sh              # loads the images built in step 2 into the cluster
```

MetalLB gives the operator-provisioned gateway-runtime LoadBalancer Service a routable address the suite can reach. On macOS, use `./kind/setup-colima.sh` instead of `./kind/setup-kind.sh`: see that script's header comments for the Colima host-reachability setup details.

### 4. Install the Gateway API CRDs, operator, and GatewayClass

```sh
./install-wso2-gateway.sh
```

### 5. Run the conformance suite

```sh
export IMPL_VERSION=v1.2.0-alpha2
./run-conformance.sh
```

### 6. View the report

```sh
cat wso2-api-platform-*-report.yaml
```
