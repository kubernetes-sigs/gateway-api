# GEP-3388: HTTPRoute Retry Budget

* Issue: [#3388](https://github.com/kubernetes-sigs/gateway-api/issues/3388)
* Status: Provisional

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

To allow configuration of a "retry budget" in HTTPRoute, to limit the rate of client-side retries based on a percentage of the active request load across all endpoints of a destination service.

## Goals

* To allow specification of a retry ["budget"](https://finagle.github.io/blog/2016/02/08/retry-budgets/) to determine whether a request should be retried, and any shared configuration or interaction with configuration of a static retry limit within HTTPRoute.
* To allow specification of a percentage of active requests, or recently active requests, that should be able to be retried concurrently.
* To allow specification of a *minimum* number of retries that should be allowed per second or concurrently, such that the budget for retries never goes below this minimum value.
* To define a standard for retry budgets that reconciles the known differences in current retry budget functionality between Gateway API data plane implementations.

## Non-Goals

* To allow specifying a default retry budget policy across a namespace or attached to a specific gateway.
* To allow configuration of a back-off strategy or timeout window within the retry budget spec.
* To allow specifying inclusion of specific HTTP status codes and responses within the retry budget spec.
* To allow specification of more than one retry budget for a given service, for specific subsets of its traffic.


## Introduction

Multiple data plane proxies offer optional configuration for budgeted retries, in order to create a dynamic limit on the amount of a service's active request that is being retried across its clients. In the case of Linkerd, retry budgets are the default retry policy configuration for HTTP retries within the [ServiceProfile CRD](https://linkerd.io/2.12/reference/service-profiles/), with static max retries being a [fairly recent addition](https://linkerd.io/2024/08/13/announcing-linkerd-2.16/).

Configuring a limit for client retries is an important factor in building a resilient system, allowing requests to be successfully retried during periods of intermittent failure. But too many client-side retries can also exacerbate consistent failures and slow down recovery, quickly overwhelming a failing system and leading to cascading failures such as retry storms. Configuring a sane limit for max client-side retries is often challenging in complex systems. Allowing an application developer (Ana) to configure a dynamic "retry budget", reducing the risk of a high number of retries across clients, allows a service to perform as expected in both times of high & low request load, as well as both during periods of intermittent & consistent failures.

While HTTPRoute retry budget configuration has been a frequently discussed feature within the community, differences in semantics between different data plane proxies creates a challenge for a consensus on the correct location for the configuration.

Envoy, for example, offers retry budgets as a configurable circuit breaker threshold for concurrent retries to an upstream cluster, in favor of configuring a static active retry threshold. In Istio, Envoy circuit breaker thresholds are typically configured [within the DestinationRule CRD](https://istio.io/latest/docs/reference/config/networking/destination-rule/#ConnectionPoolSettings-HTTPSettings), which applies rules to clients of a service after routing has already occurred. The linkerd implementation of retry budgets is configured alongside service route configuration, within the [ServiceProfile CRD](https://linkerd.io/2.12/reference/service-profiles/), limiting the number of total retries for a service as a percentage of the number of recent requests. This proposal aims to determine where retry budget's should be defined within the Gateway API, and whether data plane proxies may need to be altered to accommodate the specification.

### Background on implementations

#### Envoy

Supports configuring a [RetryBudget](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-msg-config-cluster-v3-circuitbreakers-thresholds-retrybudget) CircuitBreaker threshold across a group of upstream endpoints, with the following parameters.

* `budget_percent` Specifies the limit on concurrent retries as a percentage of the sum of active requests and active pending requests. For example, if there are 100 active requests and the budget_percent is set to 25, there may be 25 active retries. This parameter is optional. Defaults to 20%.

* `min_retry_concurrency` Specifies the minimum retry concurrency allowed for the retry budget. The limit on the number of active retries may never go below this number. This parameter is optional. Defaults to 3.

#### linkerd2-proxy

Linkerd supports [budgeted retries](https://linkerd.io/2.15/features/retries-and-timeouts/), the default way to specify retries to a service, and - as of [edge-24.7.5](https://github.com/linkerd/linkerd2/releases/tag/edge-24.7.5) - counted retries. In all cases, retries are implemented by the `linkerd2-proxy` making the request on behalf on an application workload.

Linkerd's budgeted retries allow retrying an indefinite number of times, as long as the fraction of retries remains within the budget. Budgeted retries are supported only using Linkerd's native ServiceProfile CRD, which allows enabling retries, setting the retry budget (by default, 20% plus 10 "extra" retries per second), and configuring the window over which the fraction of retries to non-retries is calculated.

## API

### Go

TODO

### YAML

TODO

## Conformance Details

TODO

## Alternatives

### Policy Attachment

TODO

## Other considerations

TODO

### What accommodations are needed for retry budget support?

Changing the retry stanza to a Kubernetes "tagged union" pattern with something like `mode: "budget"` to support mutually-exclusive distinct sibling fields is possible as a non-breaking change if omitting the `mode` field defaults to the currently proposed behavior (which could retroactively become something like `mode: count`).

## References

* <https://gateway-api.sigs.k8s.io/geps/gep-1731/>
* <https://linkerd.io/2019/02/22/how-we-designed-retries-in-linkerd-2-2/>
* <https://linkerd.io/2.11/tasks/configuring-retries/>
* <https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#config-cluster-v3-circuitbreakers-thresholds-retrybudget>
