# GEP-3388: Retry Budgets

* Issue: [#3388](https://github.com/kubernetes-sigs/gateway-api/issues/3388)
* Status: Experimental

(See [status definitions](../overview.md#gep-states).)

## TLDR

To allow configuration of a "retry budget" across all endpoints of a destination service, preventing additional client-side retries when the percentage of the active request load consisting of retries reaches a certain threshold.

## Goals

* To allow specification of a retry ["budget"](https://finagle.github.io/blog/2016/02/08/retry-budgets/) to determine whether a request should be retried, and any shared configuration or interaction with configuration of a static retry limit within HTTPRoute.
* To allow specification of a percentage of active requests, or recently active requests, that should be able to be retried concurrently.
* To allow specification of a *minimum* number of retries that should be allowed per second or concurrently, such that the budget for retries never goes below this minimum value.
* To define a standard for retry budgets that reconciles the known differences in current retry budget functionality between Gateway API data plane implementations.

## Non-Goals

* To allow specifying a default retry budget policy across a namespace or attached to a specific gateway.
* To allow configuration of a back-off strategy or timeout window within the retry budget spec.
* To allow specifying inclusion of specific HTTP status codes and responses within the retry budget spec.
* To allow specification of more than one retry budget for a given service, or for specific subsets of its traffic.

## Introduction

Multiple data plane proxies offer optional configuration for budgeted retries, in order to create a dynamic limit on the amount of a service's active request load that is comprised of retries from across its clients. In the case of Linkerd, retry budgets are the default retry policy configuration for HTTP retries within the [ServiceProfile CRD](https://linkerd.io/2.12/reference/service-profiles/), with static max retries being a [fairly recent addition](https://linkerd.io/2024/08/13/announcing-linkerd-2.16/).

Configuring a limit for client retries is an important factor in building a resilient system, allowing requests to be successfully retried during periods of intermittent failure. But too many client-side retries can also exacerbate consistent failures and slow down recovery, quickly overwhelming a failing system and leading to cascading failures such as retry storms. Configuring a sane limit for max client-side retries is often challenging in complex systems. Allowing an application developer (Ana) to configure a dynamic "retry budget" reduces the risk of a high number of retries across clients. It allows a service to perform as expected in both times of high & low request load, as well as both during periods of intermittent & consistent failures.

While retry budget configuration has been a frequently discussed feature within the community, differences in the semantics between data plane implementations creates a challenge for a consensus on the correct location for the configuration. This proposal aims to determine where retry budgets should be defined within the Gateway API, and whether data plane proxies may need to be altered to accommodate the specification.

### Background on implementations

#### Envoy

Envoy offers retry budgets as a configurable circuit breaker threshold for concurrent retries to an upstream cluster, in favor of configuring a static max retry threshold. In Istio, Envoy circuit breaker thresholds are typically configured [within the DestinationRule CRD](https://istio.io/latest/docs/reference/config/networking/destination-rule/#ConnectionPoolSettings-HTTPSettings), which applies rules to clients of a service after routing has already occurred.

The optional [RetryBudget](https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#envoy-v3-api-msg-config-cluster-v3-circuitbreakers-thresholds-retrybudget) CircuitBreaker threshold can be configured with the following parameters:

* `budget_percent` Specifies the limit on concurrent retries as a percentage of the sum of active requests and active pending requests. For example, if there are 100 active requests and the budget_percent is set to 25, there may be 25 active retries. This parameter is optional. Defaults to 20%.

* `min_retry_concurrency` Specifies the minimum retry concurrency allowed for the retry budget. The limit on the number of active retries may never go below this number. This parameter is optional. Defaults to 3.

By default, Envoy uses a static threshold for retries. But when configured, Envoy's retry budget threshold overrides any other retry circuit breaker that has been configured.

#### linkerd2-proxy

The Linkerd implementation of retry budgets is configured alongside service route configuration, within the [ServiceProfile CRD](https://linkerd.io/2.12/reference/service-profiles/), limiting the number of total retries for a service as a percentage of the number of recent requests. In practice, this functions similarly to Envoy's retry budget implementation, as it is configured in a single location and measures the ratio of retry requests to original requests across all traffic destined for the service.

(Note that budgeted retries have become less commonly used since Linkerd added support for counted retries in [edge-24.7.5](https://github.com/linkerd/linkerd2/releases/tag/edge-24.7.5): ServiceProfile operates at the level of a backend workload, meaning that it cannot configure anything at the level of a route, but counted retries can be configured using annotations on Service, HTTPRoute, and GRPCRoute.)

For both counted retries and budgeted retries, the actual retry logic is implemented by the `linkerd2-proxy` making the request on behalf on an application workload. The receiving proxy is not aware of the retry configuration at all.

Linkerd's budgeted retries allow retrying an indefinite number of times, as long as the fraction of retries remains within the budget. Budgeted retries are supported only using Linkerd's native ServiceProfile CRD, which allows enabling retries, setting the retry budget (by default, 20% plus 10 "extra" retries per second), and configuring the window over which the fraction of retries to non-retries is calculated. The `retryBudget` field of the ServiceProfile spec can be configured with the following optional parameters:

* `retryRatio` Specifies a ratio of retry requests to original requests that is allowed. The default is 0.2, meaning that retries may add up to 20% to the request load.

* `minRetriesPerSecond` Specifies the minimum rate of retries per second that is allowed, so that retries are not prevented when the request load is very low. The default is 10.

* `ttl` A duration specifying how long requests are considered for when calculating the retry threshold. The default is 10s.

### Proposed Design

#### Retry Budget Policy Attachment

While current retry behavior is defined at the routing rule level within HTTPRoute, exposing retry budget configuration as a policy attachment offers some advantages:

* Users could define a single policy, targeting a service, that would dynamically configure a retry threshold based on the percentage of active requests across *all routes* destined for that service's backends.

* In both Envoy and Linkerd data plane implementations, a retry budget is configured once to match all endpoints of a service, regardless of the routing rule that the request matches on. A policy attachment will allow for a single configuration for a service's retry budget, as opposed to configuring the retry budget across multiple HTTPRoute objects (see [Alternatives](#httproute-retry-budget)).

* Being able to configure a dynamic threshold of retries at the service level, alongside a static max number of retries on the route level. In practice, application developers would then be allowed more granular control of which requests should be retried. For example, an application developer may not want to perform retries on a specific route where requests are not idempotent, and can disable retries for that route. By having a retry budget policy configured, retries from other routes will still benefit from the budgeted retries.

Configuring a retry budget through a Policy Attachment may produce some confusion from a UX perspective, as users will be able to configure retries in two different places (HTTPRoute for static retries, versus a policy attachment for a dynamic retry threshold). Though this is likely a fair trade-off.

Discrepancies in the semantics of retry budget behavior and configuration options between Envoy and Linkerd may require a change in either implementation to accommodate the Gateway API specification. While Envoy's `min_retry_concurrency` setting may behave similarly in practice to Linkerd's `minRetriesPerSecond`, they are not directly equivalent.

The implementation of a version of Linkerd's `ttl` parameter within Envoy might be a path towards reconciling the behavior of these implementations, as it could allow Envoy to express a `budget_percent` and minimum number of permissible retries over a period of time rather than by tracking active and pending connections. It is not currently clear which of these models is preferable, but being able to specify a budget as requests over a window of time seems like it might offer more predictable behavior.

## API

### Go

```golang
type BackendTrafficPolicy struct {
    // BackendTrafficPolicy defines the configuration for how traffic to a target backend should be handled.
    //
    // Support: Extended
    //
    // +optional
    // <gateway:experimental>
    //
    // Note: there is no Override or Default policy configuration.

    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of BackendTrafficPolicy.
    Spec BackendTrafficPolicySpec `json:"spec"`
    
    // Status defines the current state of BackendTrafficPolicy.
    Status PolicyStatus `json:"status,omitempty"`
}

type BackendTrafficPolicySpec struct {
  // TargetRef identifies an API object to apply policy to.
  // Currently, Backends (i.e. Service, ServiceImport, or any
  // implementation-specific backendRef) are the only valid API
  // target references.
  // +listType=map
  // +listMapKey=group
  // +listMapKey=kind
  // +listMapKey=name
  // +kubebuilder:validation:MinItems=1
  // +kubebuilder:validation:MaxItems=16
  TargetRefs []LocalPolicyTargetReference `json:"targetRefs"`

  // Retry defines the configuration for when to retry a request to a target backend.
  //
  // Implementations SHOULD retry on connection errors (disconnect, reset, timeout,
  // TCP failure) if a retry stanza is configured.
  //
  // Support: Extended
  //
  // +optional
  // <gateway:experimental>
  Retry *CommonRetryPolicy `json:"retry,omitempty"`

  // SessionPersistence defines and configures session persistence
  // for the backend.
  //
  // Support: Extended
  //
  // +optional
  SessionPersistence *SessionPersistence `json:"sessionPersistence,omitempty"`
}

// CommonRetryPolicy defines the configuration for when to retry a request.
//
type CommonRetryPolicy struct {
    // Support: Extended
    //
    // +optional
    BudgetPercent *Int `json:"budgetPercent,omitempty"`

    // Support: Extended
    //
    // +optional
    BudgetInterval *Duration `json:"budgetInterval,omitempty"`

    // Support: Extended
    //
    // +optional
    MinRetryRate *RequestRate `json:"minRetryRate,omitempty"`
}

// RequestRate expresses a rate of requests over a given period of time.
//
type RequestRate struct {
    // Support: Extended
    //
    // +optional
    Count *Int `json:"count,omitempty"`

    // Support: Extended
    //
    // +optional
    Interval *Duration `json:"interval,omitempty"`
}

// Duration is a string value representing a duration in time. The format is
// as specified in GEP-2257, a strict subset of the syntax parsed by Golang
// time.ParseDuration.
//
// +kubebuilder:validation:Pattern=`^([0-9]{1,5}(h|m|s|ms)){1,4}$`
type Duration string

### YAML

```yaml
apiVersion: gateway.networking.x-k8s.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: traffic-policy-example
spec:
  targetRefs:
    - group: ""
      kind: Service
      name: foo
  retry:
    budgetPercent: 20
    budgetInterval: 10s
    minRetryRate:
      count: 3
      interval: 1s
  sessionPersistence:
    ...
  status:
    ancestors:
    - ancestorRef:
        kind: Mesh
        namespace: istio-system
        name: istio
      controllerName: "istio.io/mesh-controller"
      conditions:
      - type: "Accepted"
        status: "False"
        reason: "Invalid"
        message: "BackendTrafficPolicy field sessionPersistence is not supported for Istio mesh traffic."
    - ancestorRef:
        kind: Gateway
        namespace: foo-ns
        name: foo-ingress
      controllerName: "istio.io/mesh-controller"
      conditions:
      - type: "Accepted"
        status: "False"
        reason: "Invalid"
        message: "BackendTrafficPolicy fields retry.budgetPercentage, retry.budgetInterval and retry.minRetryRate are not supported for Istio ingress gateways."
    ...
```

## Conformance Details

TODO

## Alternatives

### HTTPRoute Retry Budget

* The desired UX for retry budgets is to apply the policy at the service level, rather than individually across each route targeting the service. Placing the retry budget configuration within HTTPRoute would violate this requirement, as separate HTTPRoute objects could each have routing rules targeting the same destination service, and a single HTTPRoute object can target multiple destinations. To apply a retry budget to all routes targeting a service, a user would need to duplicate the configuration across multiple routing rules.

* If we wanted retry budgets to be configured on a per-route basis (as opposed to at the service level), it would require a change to be made in Envoy Route. And more than likely, similar changes would need to be made for Linkerd.

## Other considerations

* As there isn't anything inherently specific to HTTP requests in either known implementation, a retry budget policy on a target Service could likely be applicable to GRPCRoute as well as HTTPRoute requests.
* While retry budgets are commonly associated with service mesh uses cases to handle many distributed clients, a retry budget policy may also be desirable for north/south implementations of Gateway API to prioritize new inbound requests and minimize tail latency during periods of service instability.

## References

* <https://gateway-api.sigs.k8s.io/geps/gep-1731/>
* <https://finagle.github.io/blog/2016/02/08/retry-budgets/>
* <https://linkerd.io/2019/02/22/how-we-designed-retries-in-linkerd-2-2/>
* <https://linkerd.io/2.11/tasks/configuring-retries/>
* <https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/circuit_breaker.proto#config-cluster-v3-circuitbreakers-thresholds-retrybudget>
