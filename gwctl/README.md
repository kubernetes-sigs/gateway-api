# gwctl

gwctl is a tool that improves the usability of the Gateway API by providing a better way to view and manage policies ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713)). The aim is to make it available as a standalone binary, a [kubectl plugin](https://gateway-api.sigs.k8s.io/geps/gep-713/#kubectl-plugin-or-command-line-tool), and a library.

gwctl allows you to view all Gateway API policy types that are present in a cluster, as well as all "policy bindings" in a namespace (or across all namespaces). It also shows you the attached policies when you view any Gateway resource (like HTTPRoute, Gateway, GatewayClass, etc.)

gwctl uses the `gateway.networking.k8s.io/policy=true` label to identify Policy CRDs.

> [!NOTE]
> gwctl is still considered an [experimental feature of the Gateway API](https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard). While we iterate on the early stages of this tool, bugs and incompatible changes will be more likely.

In the future, gwctl may be able to read status from the policy resource to determine if it has been applied correctly.

## Installation

1. Before you install gwctl, ensure that your system meets the following requirements:
   1. Install Git: Make sure Git is installed on your system to clone the project repository.
   2. Install Go. Make sure the Go language is installed on your system. You can download it from the [official website](https://golang.org/dl/) and follow the installation instructions.

2. Clone the project repository:
   
   ```bash
   git clone https://github.com/kubernetes-sigs/gateway-api.git && cd gateway-api/gwctl
   ```

3. Build the project:
   
   ```bash
   make build
   ```

4. Add binary to `PATH`:
   
   ```bash
   export PATH="./bin:${PATH}"
   ```

5. Run gwctl:
   
   ```shell
   gwctl help
   ```

## Usage

The examples below demonstrate how gwctl can be used.

List all policies in the cluster. This will also show the resource they bind to:

```bash
gwctl get policies -A
```

```
POLICYNAME                      POLICYKIND               TARGETNAME                      TARGETKIND
timeout-policy-on-gatewayclass  TimeoutPolicy            foo-com-external-gateway-class  GatewayClass
timeout-policy-on-namespace     TimeoutPolicy            default                         Namespace
health-check-1                  HealthCheckPolicy        gateway-1                       Gateway
retry-policy-1                  RetryOnPolicy            gateway-1                       Gateway
retry-policy-2                  RetryOnPolicy            httproute-2                     HTTPRoute
tls-min-version-policy-1        TLSMinimumVersionPolicy  httproute-1                     HTTPRoute
tls-min-version-policy-2        TLSMinimumVersionPolicy  gateway-2                       Gateway
```

List all available policy types:

```bash
gwctl get policycrds
```

```
NAME                                          GROUP                      KIND              POLICY TYPE  SCOPE
backendtlspolicies.gateway.networking.k8s.io  gateway.networking.k8s.io  BackendTLSPolicy  Direct       Namespaced
```

Describe all HTTPRoutes in namespace `prod`:

```bash
gwctl describe httproutes -n prod
```

```
Name: httproute-3
Namespace: prod
Hostnames:
- example.com
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: gateway-2
EffectivePolicies:
  prod/gateway-2:
    TLSMinimumVersionPolicy.baz.com:
      default:
        sampleField: sample


Name: httproute-4
Namespace: prod
Hostnames:
- demo.com
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: gateway-1
  namespace: default
EffectivePolicies:
  default/gateway-1:
    HealthCheckPolicy.foo.com:
      default:
        sampleField: hello
    RetryOnPolicy.foo.com:
      default:
        sampleField: sample
    TimeoutPolicy.bar.com:
      timeout1: parent
      timeout2: child
      timeout3: parent
      timeout4: child
```

Describe a single HTTPRoute in default namespace:

```shell
gwctl describe httproutes httproute-1
```

```
Name: httproute-1
Namespace: dev
Hostnames:
- example.com
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: gateway-1
EffectivePolicies:
  dev/gateway-2:
    TLSMinimumVersionPolicy.baz.com:
      default:
        sampleField: sample
```

Describe all Gateways across all namespaces:

```shell
gwctl describe gateways -A
```

```
Name: gateway-1
Namespace: default
GatewayClass: foo-com-external-gateway-class

Name: gateway-2
Namespace: prod
GatewayClass: foo-com-external-gateway-class

Name: gateway-3
Namespace: dev
GatewayClass: foo-com-external-gateway-class
```

> [!TIP]
> You can use the `--help` or the `-h` flag for a usage guide for any subcommand.

## Get Involved

This project will be discussed in the same Slack channel and community meetings as the rest of the Gateway API subproject. For more information, refer to the [Gateway API Community](https://gateway-api.sigs.k8s.io/contributing/) page.

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).

[owners]: https://git.k8s.io/community/contributors/guide/owners.md
[Creative Commons 4.0]: https://git.k8s.io/website/LICENSE
