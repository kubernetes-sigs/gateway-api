# AWS Load Balancer Controller

The AWS Load Balancer Controller manages AWS Elastic Load Balancers for a Kubernetes cluster. The controller provisions AWS Application Load Balancers (ALB) when you create a Kubernetes Ingress and AWS Network Load Balancers (NLB) when you create a Kubernetes Service of type LoadBalancer.

We are also releasing AWS Load Balancer Controller Gateway API support for both Layer 4 (L4) and Layer 7 (L7) routing. This highly anticipated feature enables customers to provision and manage AWS Network Load Balancers (NLBs) and Application Load Balancers (ALBs) directly from Kubernetes clusters using the extensible Gateway API, providing a modern alternative to traditional Ingress and Service resources.

## Table of contents

| API channel | Implementation version                                                                          | Mode | Report                                                   |
|-------------|-------------------------------------------------------------------------------------------------|------|----------------------------------------------------------|
| experimental | [v2.17.0](https://github.com/kubernetes-sigs/aws-load-balancer-controller/releases/tag/v2.17.0) | default | [v2.17.0 report](experimental-v2.17-default-report.yaml) |

## Reproduce

To reproduce the conformance test results for AWS Load Balancer Controller v2.17.0:

For detailed instructions, refer to the [AWS Load Balancer Controller - Conformance Test Instruction](https://github.com/kubernetes-sigs/aws-load-balancer-controller/blob/main/conformance/README.md).
