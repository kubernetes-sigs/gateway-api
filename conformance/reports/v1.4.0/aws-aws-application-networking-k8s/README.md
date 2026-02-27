# Amazon VPC Lattice Gateway API Controller

The [Amazon VPC Lattice Gateway API Controller](https://github.com/aws/aws-application-networking-k8s) is an implementation of the Kubernetes Gateway API that orchestrates AWS VPC Lattice resources using Kubernetes Custom Resource Definitions like Gateway and HTTPRoute.

## Table of contents

| API channel  | Implementation version | Mode    | Report |
|--------------|------------------------|---------|--------|
| experimental | [v2.0.1](https://github.com/aws/aws-application-networking-k8s/releases/tag/v2.0.1) | default | [v2.0.1 report](./experimental-v2.0.1-default-report.yaml) |

## Reproduce

1. Create an EKS cluster with VPC Lattice configured in the same VPC.
2. Install the Amazon VPC Lattice Gateway API Controller v2.0.1.
3. Follow the conformance test instructions in the [controller documentation](https://github.com/aws/aws-application-networking-k8s/blob/main/docs/conformance-test.md).
