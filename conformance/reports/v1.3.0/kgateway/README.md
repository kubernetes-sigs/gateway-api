# kgateway

## Table of contents

| API channel  | Implementation version                                                        | Mode    | Report                                                    |
|--------------|-------------------------------------------------------------------------------|---------|-----------------------------------------------------------|
| experimental | [main](https://github.com/kgateway-dev/kgateway) | default | [Link](./experimental-main-report.yaml) |

## Reproduce

### Steps

1. Clone the kgateway repository:

   ```sh
   git clone https://github.com/kgateway-dev/kgateway.git && cd kgateway
   ```

2. Bootstrap a KinD cluster:

   ```sh
   make run
   ```

3. Run the conformance tests:

   ```sh
   make conformance
   ```

4. View and verify the conformance report: `cat _test/conformance/1.0.1-dev-report.yaml`
