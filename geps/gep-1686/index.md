# GEP-1686: Mesh conformance testing plan

- Issue: [#1686](https://github.com/kubernetes-sigs/gateway-api/issues/1686)
- Status: Standard

## TLDR

This testing plan specifies a new set of tests to define a "Mesh" [conformance profile](https://github.com/kubernetes-sigs/gateway-api/issues/1709).

## Goals

* Define a strategy for segmenting GAMMA tests from the existing conformance test suite
* Define a set of test scenarios to capture conformance with the GAMMA spec
* Rely on existing tests for non-GAMMA-specific Gateway API conformance

## Focus

Currently the GAMMA spec consists of two Gateway API GEPs [defining terminology and goals of Gateway API for service meshes](https://gateway-api.sigs.k8s.io/geps/gep-1324/)
and specifically [how route resources work in a service mesh context](https://gateway-api.sigs.k8s.io/geps/gep-1426/).
The goal of the initial conformance testing is to check the essential behavior as defined by GEP-1426, as it differs from the wider Gateway API spec. This GEP focuses on using a `Service` object as an `xRoute` `parentRef` to control how the GAMMA implementation directs traffic to the endpoints specified by the `Services` in `backendRefs` and how the traffic is filtered and modified.

## Conformance Profile

GAMMA intends to introduce a "Mesh" [conformance profile](https://gateway-api.sigs.k8s.io/geps/gep-1709/) to isolate tests specific to East/West functionality from both existing tests focused on North/South functionality and common Gateway API functionality shared by N/S and E/W implementations. A conformance profile is a set of tests that implementations can run to check their conformance to some subset of the Gateway API spec.

This appropach will enable service meshes to certify that an implementation follows the GAMMA spec without requiring a North/South implementation, and importantly avoid any expectation that North/South Gateway API implementations expand their scope to understand GAMMA and E/W traffic flows.

Leveraging existing tests for common functionality between N/S and E/W implementations will both ensure consistency across Gateway API implementations and help limit the maintence burden for the conformance testing suite.

### Support Levels

Using a conformance profile will enable granular conformance definitions for GAMMA, splitting functionality along the existing Gateway API [support levels](https://gateway-api.sigs.k8s.io/concepts/conformance/?h=conformance+levels#2-support-levels), with required functionality as Core, standardized but optional functionality as Extended, and Implementation-specific for things beyond the current or intended scope of GAMMA configuration. It is expected that some capabilities will begin as Implementation-specific and eventually migrate to Extended or Core conformance as GAMMA matures.

## Tests

Testing GAMMA implementations requires both a new suite of test cases as well as refactoring the existing test framework setup.

### Runner and Setup

The existing Gateway API conformance tests use a relatively simple implementation to send requests from outside a Kubernetes cluster to a gateway sitting at the edge, [capture the request and response](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/roundtripper/roundtripper.go), and [assert a match against an expected response](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/utils/http/http.go).

GAMMA conformance tests should still be based around a request/expected response suite, but requests will need to originate from _inside the cluster_, from either the same or different namespace as the target service. Adopting or developing tooling to enable this is being explored in [gateway-api#1340](https://github.com/kubernetes-sigs/gateway-api/issues/1340).

### Scenarios

All requests are sent from a client inside the same cluster/mesh and the same `Namespace`
as the `Service` under test.
Test scenarios are largely focused on the `backendRefs` and the
`Namespace` of an `xRoute` resource.

#### `Service` as `parentRef`

- Given a simple `HTTPRoute` with a single `backendRef`
  - With an explicit `port` in `parentRef`
    - Assert that only requests to this `Service` and `port` are directed to the
      backend
  - Without a `port` in `parentRef`
    - Assert that all requests to this `Service` are directed to the backend

#### Omitted `backendRefs`

- Given a simple `HTTPRoute` without `backendRefs`
  - Assert that requests are directed to the endpoints defined by the `Service`
    `parentRef` in its backend role

#### Only `Services` as frontends are affected

- Given a simple `HTTPRoute` with a single `backendRef`
  - Send requests directly the endpoints of the `parentRef` `Service`'s backend
  - Assert that traffic is not affected by the `HTTPRoute` resource

#### `Namespace`-dependent behavior, producer vs consumer

A producer `HTTPRoute` is in the same namespace as the `parentRef` `Service` (the
producer).

- Given a producer `HTTPRoute`
  - Assert that traffic from a client in the producer `Namespace` is routed by the
    `HTTPRoute`
  - Assert that traffic from a client in a different `Namespace` is routed by the
    `HTTPRoute`

A consumer `HTTPRoute` is in the same `Namespace` as the the request sender (the
consumer), a different `Namespace` as the `parentRef` `Service`.

- Given a consumer `HTTPRoute`
  - Assert that traffic from the consumer client is routed by the `HTTPRoute`
  - Assert that traffic from a client in a different `Namespace` is _not_ routed by the
    `HTTPRoute`

Consumer routes have priority over producer routes.

- Given both a consumer `HTTPRoute` and a producer `HTTPRoute`
  - Assert that traffic from the consumer client is routed by the consumer `HTTPRoute`
  - Assert that traffic from a client in a different `Namespace` is routed by
    the producer `HTTPRoute`

#### `xRoute`-specific

- Given multiple `xRoutes` of different types
  - Assert that routes take affect according to the specificity as defined in the spec
- Given an `HTTPRoute` without `matches`, all requests are received at the `Service` endpoints as if no `HTTPRoute` existed
- Given an `HTTPRoute` with `matches`, unmatched requests are dropped with a 404

#### Filters

Filters have the same effects on requests as any implementation. Gateway API conformance test framework can be
refactored to extract checks on filter behavior for use on both GAMMA and Gateway API tests.
