# Running Conformance Tests for Envoy Gateway

## Prerequisites

Make sure you have these tools installed

* [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
* [helm](https://helm.sh/docs/intro/install/)

## (Optional) Step 1: Create kind cluster

* Skip this step if you have an existing Kubernetes you would like to use.

```shell script
../common/create-cluster.sh
```

## Step 2: Install Envoy Gateway (`v0.0.0-latest`)

```shell script
helm install eg oci://docker.io/envoyproxy/gateway-helm --version v0.0.0-latest -n envoy-gateway-system --create-namespace
```

## Step 3: Install the GatewayClass resource

```shell script
cat <<EOF | kubectl apply -f -
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1
metadata:
  name: envoy-gateway
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
EOF
```

## Step 4: Run a specific conformance test

```shell script
go test -v ./conformance \
--run TestConformance/TLSRouteSimpleSameNamespace \
--gateway-class=envoy-gateway --supported-features=Gateway,TLSRoute \
--allow-crds-mismatch
```

## (Optional) Step 4: Delete kind cluster

```shell script
kind delete cluster --name envoy-gateway
```
