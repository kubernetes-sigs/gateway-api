# Agentgateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v1.3.1](https://github.com/agentgateway/agentgateway/releases/tag/v1.3.1)|default|[report](./v1.3.1-report.yaml)|
|experimental|[v1.4.1](https://github.com/agentgateway/agentgateway/releases/tag/v1.4.1)|default|[report](./v1.4.1-report.yaml)|
|experimental|[v1.5.0](https://github.com/agentgateway/agentgateway/releases/tag/v1.5.0)|default|[report](./v1.5.0-report.yaml)|
|experimental|[v1.6.0](https://github.com/agentgateway/agentgateway/releases/tag/v1.6.0)|default|[report](./v1.6.0-report.yaml)|
|experimental|[v1.7.0](https://github.com/agentgateway/agentgateway/releases/tag/v1.7.0)|default|[report](./v1.7.0-report.yaml)|

## Reproduce

### Steps

1. Clone the agentgateway repository:

   ```sh
   git clone https://github.com/agentgateway/agentgateway.git && cd agentgateway && git checkout tags/v1.7.0
   ```

2. Bootstrap a KinD cluster with all the necessary components installed:

   ```sh
   ./controller/test/setup/setup-kind-ci.sh
   ```

3. Run the conformance tests

   ```sh
   make -C controller agw-conformance
   ```

4. View and verify the conformance report: `cat controller/_test/conformance/*`

