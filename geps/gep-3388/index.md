# GEP-3388: HTTPRoute Retry Budget

* Issue: [#3388](https://github.com/kubernetes-sigs/gateway-api/issues/3388)
* Status: Provisional

(See status definitions [here](/geps/overview/#gep-states).)

## TLDR

To allow budgeted retry configuration of a Gateway, in order to retry unsuccessful requests based on a percentage of it's
active request load, as opposed to a static max retry value.

## Goals

* To allow specification of a retry
  ["budget"](https://finagle.github.io/blog/2016/02/08/retry-budgets/) to
  determine whether a request should be retried, and any shared configuration
  or interaction with max count retry configuration.
* To allow specification of the percentage of active requests that should be able to be retried at the same time.
* To allow specification of the minimum number of retries that should be
  allowed per second or concurrently, such that the budget for retries never
  goes below this minimum value.
* To define a standard for retry budgets that reconciles the known
  differences in retry budget functionality between Gateway API implementations.

## Future Goals

## Non-Goals

## Introduction

Multiple data plane proxies offer optional configuration for budgeted retries,
either as a circuit breaker threshold for concurrent retries or as an
alternative for configuring a
static retry limit for client retries. In the case of Linkerd, retry budgets
are the default retry policy configuration for HTTP retries, with static max
retries being a fairly recent addition.

Configuring a limit for client retries is an important factor in building a
resilient system in order to
allow for requests to be successfully retried during periods of intermittent
failure. But too many client-side retries can also exacerbate consistent
failures and slow down recovery, quickly overwhelming a failing
system and leading to retry
storms. Configuring a sane
limit for max client-side retries is often challenging in complex
systems. Allowing an application developer (Ana) to instead configure a dynamic
"retry budget" prevents them from needing to decide on a static max retry value
that will perform as expected in both times of high & low request load, as well
as periods of intermittent or consistent failures.

While HTTPRoute retry budget configuration has been a frequently discussed
feature within the community, differences in semantics between different data
plane proxies
creates a challenge for a consensus on the correct location for the
configuration.

Envoy, for example, offers retry budgets as a configurable circuit breaker threshold
for concurrent retries to an upstream cluster. In Istio, Envoy circuit breaker
thresholds are typically configured [within the DestinationRule
CRD](https://istio.io/latest/docs/reference/config/networking/destination-rule/#ConnectionPoolSettings-HTTPSettings),
which
applies rules to clients of a service after routing has already occurred.
The linkerd implementation of
retry budgets is configured on specific routes, and instead limits the number
of total retry attempts as a percentage of original requests. This creates a
challenge for
defining where retry budget's should be configured within the Gateway API,
and how data plane proxies may need to be altered to accommodate the correct
path forward. If Istio were to implement Envoy's retry budget threshold also
at the per-route level in their API, retry budget
configuration would need to be introduced within [the VirtualService
CRD](https://istio.io/latest/docs/reference/config/networking/virtual-service/#HTTPRetry).
Envoy's retry budget threshold does not address overall retry attempts on the
client-side, though. A potential solution would be for Envoy to additionally
allow a budget for retry *attempts* as well as a concurrent retry threshold.

When configuring a retry budget on the route, you
subsequently need to define this value for each one. Defining a single
retry budget threshold for a destination is a simpler approach.

### Background on implementations


#### Envoy

Supports configuring a [RetryBudget](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-msg-config-cluster-v3-circuitbreakers-thresholds-retrybudget) with a following parameters in cluster CircuitBreaker thresholds.

* `budget_percent` Specifies the limit on concurrent retries as a percentage of the sum of active requests and active pending requests. For example, if there are 100 active requests and the budget_percent is set to 25, there may be 25 active retries. This parameter is optional. Defaults to 20%.

* `min_retry_concurrency` Specifies the minimum retry concurrency allowed for the retry budget. The limit on the number of active retries may never go below this number. This parameter is optional. Defaults to 3.

#### NGINX


#### HAProxy


#### Traefik

Supports configuration of a [Circuit Breaker](https://doc.traefik.io/traefik/middlewares/http/circuitbreaker/) which could possibly be used to implement budgeted retries. Each router gets its own instance of a given circuit breaker. One circuit breaker instance can be open while the other remains closed: their state is not shared. This is the expected behavior, we want you to be able to define what makes a service healthy without having to declare a circuit breaker for each route.

#### linkerd2-proxy

Linkerd supports [budgeted retries](https://linkerd.io/2.15/features/retries-and-timeouts/) and - as of [edge-24.7.5](https://github.com/linkerd/linkerd2/releases/tag/edge-24.7.5) - counted retries. In all cases, retries are implemented by the `linkerd2-proxy` making the request on behalf on an application workload.

Linkerd's budgeted retries allow retrying an indefinite number of times, as long as the fraction of retries remains within the budget. Budgeted retries are supported only using Linkerd's native ServiceProfile CRD, which allows enabling retries, setting the retry budget (by default, 20% plus 10 "extra" retries per second), and configuring the window over which the fraction of retries to non-retries is calculated.

## API

### Go


### YAML

## Conformance Details


## Alternatives

### Policy Attachment

## Other considerations

### What accommodations are needed for retry budget support?

Changing the retry stanza to a Kubernetes "tagged union" pattern with something like `mode: "budget"` to support mutually-exclusive distinct sibling fields is possible as a non-breaking change if omitting the `mode` field defaults to the currently proposed behavior (which could retroactively become something like `mode: count`).

## References

* <https://gateway-api.sigs.k8s.io/geps/gep-1731/>
* <https://linkerd.io/2019/02/22/how-we-designed-retries-in-linkerd-2-2/>
* <https://linkerd.io/2.11/tasks/configuring-retries/>
* <https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#config-cluster-v3-circuitbreakers-thresholds-retrybudget>
