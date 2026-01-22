# Agent Gateway (with kgateway)

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v2.2.0-alpha.1](https://github.com/kgateway-dev/kgateway/releases/tag/v2.2.0-alpha.1)|default|[report](./v2.2.0-report.yaml)|

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway && git checkout tags/v2.2.0-alpha.1
   ```

2. Bootstrap a KinD cluster with all the necessary components installed:

   ```sh
   make setup-base
   ```

3. Deploy the kgateway Helm charts:

   ```sh
   export VERSION="v2.2.0-alpha.1"

   helm upgrade -i --create-namespace --namespace kgateway-system --version $VERSION kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds

   helm upgrade -i --namespace kgateway-system --version $VERSION kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway
   ```

4. Run the conformance tests:

   ```sh
   make agw-conformance
   ```

5. View and verify the conformance report: `cat _test/conformance/*`

