# gwctl

gwctl is a tool that improves the usability of the Gateway API by providing a better way to view and manage policies ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713)). The aim is to make it available as a standalone binary, a kubectl plugin, and a library.

gwctl allows you to view all Gateway API policy types that are present in a cluster, as well as all "policy bindings" in a namespace (or across all namespaces). It also shows you the attached policies when you view any Gateway resource (like HTTPRoute, Gateway, GatewayClass, etc.)

gwctl uses the `gateway.networking.k8s.io/policy=true` label to identify Policy CRDs (https://gateway-api.sigs.k8s.io/geps/gep-713/#kubectl-plugin)

Please note that gwctl is still considered an [experimental feature of Gateway API](https://gateway-api.sigs.k8s.io/concepts/versioning/#release-channels-eg-experimental-standard). While we iterate on the early stages of this tool, bugs and incompatible changes will be more likely.


In the future, gwctl may be able to read status from the policy resource to determine if it has been applied correctly.

## Try it out!

```bash
# Clone the gwctl repository
git clone https://github.com/kubernetes-sigs/gateway-api.git

# Go to the gwctl directory
cd gateway-api

# Ensure vendor depedencies
go mod tidy
go mod vendor

# Build the gwctl binary
go build -o bin/gwctl gwctl/cmd/main.go

# Add binary to PATH
export PATH=./bin:${PATH}

# Start using!
gwctl help
```

## Examples
Here are some examples of how gwctl can be used:

```bash
# List all policies in the cluster. This will also give the resource they bind to.
gwctl get policies -A

# List all available policy types
gwctl get policycrds

# Describe all HTTPRoutes in namespace ns2
gwctl describe httproutes -n ns2

# Describe a single HTTPRoute in default namespace
gwctl describe httproutes demo-httproute-1

# Describe all Gateways across all namespaces.
gwctl describe gateways -A

# Describe a single GatewayClass
gwctl describe gatewayclasses foo-com-external-gateway-class
```

Here are some commands with their sample output:
```bash
❯ gwctl get policies -A
POLICYNAME                           POLICYKIND               TARGETNAME                      TARGETKIND
demo-timeout-policy-on-gatewayclass  TimeoutPolicy            foo-com-external-gateway-class  GatewayClass
demo-timeout-policy-on-namespace     TimeoutPolicy            default                         Namespace
demo-health-check-1                  HealthCheckPolicy        demo-gateway-1                  Gateway
demo-retry-policy-1                  RetryOnPolicy            demo-gateway-1                  Gateway
demo-retry-policy-2                  RetryOnPolicy            demo-httproute-2                HTTPRoute
demo-tls-min-version-policy-1        TLSMinimumVersionPolicy  demo-httproute-1                HTTPRoute
demo-tls-min-version-policy-2        TLSMinimumVersionPolicy  demo-gateway-2                  Gateway

❯ gwctl describe httproutes -n ns2
Name: demo-httproute-3
Namespace: ns2
Hostnames:
- example.com
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: demo-gateway-2
EffectivePolicies:
  ns2/demo-gateway-2:
    TLSMinimumVersionPolicy.baz.com:
      default:
        sampleField: hello


Name: demo-httproute-4
Namespace: ns2
Hostnames:
- demo.com
ParentRefs:
- group: gateway.networking.k8s.io
  kind: Gateway
  name: demo-gateway-1
  namespace: default
EffectivePolicies:
  default/demo-gateway-1:
    HealthCheckPolicy.foo.com:
      default:
        sampleField: hello
    RetryOnPolicy.foo.com:
      default:
        sampleField: hello
    TimeoutPolicy.bar.com:
      timeout1: parent
      timeout2: child
      timeout3: parent
      timeout4: child

❯ gwctl describe backends service/demo-svc
Kind: Service
Name: demo-svc
Namespace: default
EffectivePolicies:
  default/demo-gateway-1:
    HealthCheckPolicy.foo.com:
      default:
        sampleField: hello
    RetryOnPolicy.foo.com:
      default:
        sampleField: hello
    TLSMinimumVersionPolicy.baz.com: {}
    TimeoutPolicy.bar.com:
      timeout1: parent
      timeout2: child
      timeout3: parent
      timeout4: child
  ns2/demo-gateway-2:
    TLSMinimumVersionPolicy.baz.com:
      default:
        sampleField: hello
    TimeoutPolicy.bar.com:
      timeout1: child
      timeout2: child
      timeout3: child
      timeout4: child
```
