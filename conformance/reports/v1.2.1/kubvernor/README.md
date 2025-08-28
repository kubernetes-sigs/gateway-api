# Kubvernor

## Table of Contents

| API channel   | Implementation version                                             | Mode          | Report                                             |
|:-------------:|:------------------------------------------------------------------:|:-------------:|:--------------------------------------------------:|
|standard       |[v0.1.0](https://github.com/kubvernor/kubvernor/releases/tag/0.1.0) |default        |[Report](./kubvernor-conformance-output-1.2.1.yaml) |




## Reproduce

0. Install Docker and Kind

1. Clone the Kubvernor GitHub repository

   ```bash
   git clone https://github.com/kubvernor/kubvernor && cd kubvernor
   ```

2. Deploy your cluster

   ```bash
   curl --proto '=https' --tlsv1.2 -sSf https://github.com/kubernetes-sigs/gateway-api/blob/main/hack/implementations/common/create-cluster.sh | sh

   ```

3. Compile and run Kubvernor
   
   ```bash
   # Install Rust
   curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
   
   # Start Kubvernor 
   export CONTROL_PLANE_IP=<IP>
   ./run_kubvernor.sh 
   
   ```

4. Run conformance tests

   ```bash
   ./run_conformance_tests.sh
   ```
   
5. Check the results

   ```bash
   cat conformance/kubvernor-conformance-output-1.2.1.yaml
   ```
