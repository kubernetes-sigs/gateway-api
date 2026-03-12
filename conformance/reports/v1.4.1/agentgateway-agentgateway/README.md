# Agentgateway

## Table of Contents

|API channel|Implementation version|Mode|Report|
|-----------|----------------------|----|------|
|experimental|[v1.0.0-alpha.2](https://github.com/agentgateway/agentgateway/releases/tag/v1.0.0-alpha.2)|default|[report](./v1.0.0-alpha-report.yaml)|

## Reproduce

### Steps

1. Clone the agentgateway repository:

   ```sh
   git clone https://github.com/agentgateway/agentgateway.git && cd agentgateway && git checkout tags/v1.0.0-alpha.2
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

