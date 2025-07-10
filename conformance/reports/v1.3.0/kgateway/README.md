# kgateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [main](https://github.com/kgateway-dev/kgateway) | default | [Link](./v2.1.0-main-report.yaml) |

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway
   ```

2. Override the version Makefile variable:

   > Note: The main branch defaults to version `1.0.1-dev` for Helm chart validation purposes. For conformance testing,
   > we need to override this with a more descriptive version that reflects the main branch:

   ```sh
   export VERSION="v2.1.0-main"
   ```

3. Bootstrap a KinD cluster:

   ```sh
   make run
   ```

4. Run the conformance tests:

   ```sh
   make conformance
   ```

5. View and verify the conformance report: `cat _test/conformance/v2.1.0-main-report.yaml`
