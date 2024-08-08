# Envoy Gateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
| experimental |[v1.1.0](https://github.com/envoyproxy/gateway/releases/tag/v1.1.0)| default |[link](./experimental-v1.1.0-default-report.yaml)|

## Reproduce

1. Clone the Envoy Gateway GitHub repository

   ```bash
   git clone https://github.com/envoyproxy/gateway.git && cd gateway
   ```

2. Check out the desired version

   ```bash
   export VERSION=v<x.y.z>
   git checkout $VERSION
   ```

3. Run the conformance tests

   ```bash
   CONFORMANCE_REPORT_PATH=conformance-report-k8s.yaml make experimental-conformance
   ```

4. Check the produced report

   ```bash
   cat ./conformance-report-k8s.yaml
   ```
