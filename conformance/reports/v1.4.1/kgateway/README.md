# kgateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [v2.2.0-beta.6](https://github.com/kgateway-dev/kgateway/releases/tag/v2.2.0-beta.6) | default | [Link](./v2.2.0-beta.6-report.yaml) |

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   export VERSION="v2.2.0-beta.6"
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway && git checkout tags/$VERSION
   ```

2. Bootstrap the environment, run kgateway, etc:

   ```sh
   make run
   ```

3. Run the conformance tests:

   ```sh
   make conformance
   ```

4. View and verify the conformance report: `cat _test/conformance/v2.2.0-beta.6-report.yaml`
