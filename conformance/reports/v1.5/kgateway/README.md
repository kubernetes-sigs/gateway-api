# kgateway

## Table of contents

| API channel  | Implementation version                                                     | Mode    | Report                                              |
|--------------|----------------------------------------------------------------------------|---------|-----------------------------------------------------|
| experimental | [v2.3.0-beta.3](https://github.com/kgateway-dev/kgateway/releases/tag/v2.3.0-beta.3) | default | [Link](./v2.3.0-beta.3-report.yaml) |

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   export VERSION="v2.3.0-beta.3"
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway && git checkout b98e8e36d80afe0ce05b7872b131264e881e0fc5
   ```

2. Bootstrap a KinD cluster with all the necessary components installed:

   ```sh
   make setup-base
   ```

3. Deploy the published kgateway Helm charts:

   ```sh
   helm upgrade -i --create-namespace --namespace kgateway-system --version $VERSION kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds

   helm upgrade -i --namespace kgateway-system --version $VERSION kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway
   ```

4. Run the conformance tests:

   ```sh
   make conformance
   ```

5. View and verify the conformance report: `cat _test/conformance/$VERSION-report.yaml`
