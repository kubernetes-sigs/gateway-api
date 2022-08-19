# HTTP traffic splitting

The [HTTPRoute resource](/api-types/httproute) allows you to specify 
weights to shift traffic between different backends. This is useful for 
splitting traffic during rollouts, canarying changes, or for emergencies. 
The HTTPRoute`spec.rules.backendRefs` accepts a list of backends that a route 
rule will send traffic to. The relative weights of these backends define 
the split of traffic between them. The following YAML snippet shows how two 
Services are listed as backends for a single route rule. This route rule 
will split traffic 90% to `foo-v1` and 10% to `foo-v2`.

![Traffic splitting](/images/simple-split.png)

```yaml
{% include 'standard/traffic-splitting/simple-split.yaml' %}
```

`weight` indicates a proportional split of traffic (rather than percentage)
and so the sum of all the weights within a single route rule is the
denominator for all of the backends. `weight` is an optional parameter and if
not specified, defaults to 1. If only a single backend is specified for a
route rule it implicitly recieves 100% of the traffic, no matter what (if any)
weight is specified.

## Guide

This guide shows the deployment of two versions of a Service. Traffic splitting
is used to manage the gradual splitting of traffic from v1 to v2.

This example assumes that the following Gateway is deployed:

```yaml 
{% include 'standard/simple-gateway/gateway.yaml' %}
```

## Canary traffic rollout

At first, there may only be a single version of a Service that serves
production user traffic for `foo.example.com`. The following HTTPRoute has no
`weight` specified for `foo-v1`  or `foo-v2` so they will implicitly
recieve 100% of the traffic matched by each of their route rules. A canary
route rule is used (matching the header `traffic=test`) to send synthetic test
traffic before splitting any production user traffic to `foo-v2`.
[Routing precedence](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteRule)
ensures that all traffic with the matching host and header 
(the most specific match) will be sent to `foo-v2`.

![Traffic splitting](/images/traffic-splitting-1.png)


```yaml
{% include 'standard/traffic-splitting/traffic-split-1.yaml' %}
```

## Blue-green traffic rollout

After internal testing has validated succesful responses from `foo-v2`,
it's desirable to shift a small percentage of the traffic to the new Service
for gradual and more realistic testing. The HTTPRoute below adds `foo-v2`
as a backend along with weights. The weights add up to a total of 100 so
`foo-v1` recieves 90/100=90% of the traffic and `foo-v2` recieves
10/100=10% of the traffic.

![Traffic splitting](/images/traffic-splitting-2.png)


```yaml
{% include 'standard/traffic-splitting/traffic-split-2.yaml' %}
```

## Completing the rollout

Finally, if all signals are positive, it is time to fully shift traffic to
`foo-v2` and complete the rollout. The weight for `foo-v1` is set to
`0` so that it is configured to accept zero traffic. 

![Traffic splitting](/images/traffic-splitting-3.png)


```yaml
{% include 'standard/traffic-splitting/traffic-split-3.yaml' %}
```

At this point 100% of the traffic is being routed to `foo-v2` and the
rollout is complete. If for any reason `foo-v2` experiences errors, the
weights can be updated to quickly shift traffic back to `foo-v1`. Once
the rollout is deemed final, v1 can be fully decommissioned.
