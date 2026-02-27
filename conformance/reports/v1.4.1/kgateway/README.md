# kgateway

## Table of contents

| API channel  | Implementation version                                                           | Mode    | Report                                                    |
|--------------|----------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [v2.2.0-rc.2](https://github.com/kgateway-dev/kgateway/releases/tag/v2.2.0-rc.2) | default | [Link](./v2.2.0-rc.2-report.yaml) |

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   export VERSION="v2.2.0-rc.2"
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway && git checkout 8d9b57e08dc8ce2fffc3aa417f507ec45a9e01b6
   ```

2. Bootstrap the environment, run kgateway, etc:

   ```sh
   make run
   ```

3. Run the conformance tests:

   ```sh
   make conformance
   ```

4. View and verify the conformance report: `cat _test/conformance/$VERSION-report.yaml`
