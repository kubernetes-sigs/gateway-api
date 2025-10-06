# GEP-3614: Firewall

* Issue: [#3614](https://github.com/kubernetes-sigs/gateway-api/issues/3614)
* Status: Provisional

## TLDR

The ability to attach firewall rules for ingress L3, L4 and L7 Gateway traffic.

## Motivation

`Gateways` are commonly exposed to the internet, which puts them as risk of
attack. Internal networks can become compromised as well. We should provide
tooling, documentation and best-practices for users to restrict and control
access to their `Gateways`.

### Definitions

* "Firewall Engine" - A processor of request payloads and applies rulesets to
  the contents to identify malicious, anomalous or otherwise unwanted traffic.
  These are generally at the front of the request path, and may be attached to a
  `Gateway` as a sidecar, integrated natively as part of the `Gateway`, or
  deployed in front of the `Gateway` as part of the networking path.

### User Stories

* As an application developer, I want to allow specific IPs to access my
  application.
* As an application developer, I want to block or allow requests based on
  headers; e.g. allow or deny specific user-agents.
* As a gateway operator I want to be able to identify and block and log
  malformed HTTP requests before they reach backend applications.
* As a gateway operator I want to be able to provide my own signature-based
  detection rulesets to spot patterns of known malicious traffic and block and
  log them, updating those rules dynamically over time.
* As a gateway operator I want to attach complete rulesets maintained by
  upstream standards bodies to block well known common threats and dynamically
  update for new threats over time.
* As a gateway operator I want to detect anomalies in traffic (which may or
  may not be conclusively malicious) and log the requests with the option to
  block them as well.
* As a cluster operator I want to be able to block traffic to gateways from
  specific geographical regions, or only allow specific regions.
* As a cluster operator I want to be able to rate limit traffic to gateways to
  avoid overuse and abuse that could decrease stability and/or spike costs.
* As a compliance officer I want to mutate (or block) and log responses that may
  contain personally identifiable information (PII).
* As a gateway operator and in the context of a request I want information about
   the request (e.g. headers) to be defined as triggers for changes in the
   subsequent rules that apply to the request (modifying, or even disabling those
   rules based on the trigger).
## Goals

* Enable attaching firewall engines to a `Gateway`
* Enable `Gateway`-level firewall rule enforcement
* Enable `HTTPRoute`-level firewall rule enforcement
* Enable processing of both requests _and_ responses
* Provide documentation and best practices for implementations which describe
  how firewall engines and rules can best be integrated into a Gateway API
  implementation.

## Non-Goals

* Building a firewall implementation
* Mesh-level support

## API

**TODO**: First PR will not include any implementation details, in favor of
building consensus on the motivation, goals and non-goals first. _"How?"_ we
implement shall be left open-ended until _"What?"_ and _"Why?"_ are solid.

## Alternatives Considered

### NetworkPolicy

When discussing this originally the obvious question whether `NetworkPolicy`
is sufficient, or should have some role in this, was asked. We do not consider
it sufficient to resolve the goals unto itself. For the purposes of this GEP,
we consider `NetworkPolicy` as an implementation detail at most: implementations
_may_ choose how they enforce firewall rules, whether some of that is
implemented with `NetworkPolicy` under the hood or not is up to them.

## References

* [GEP-1767: CORS](https://github.com/kubernetes-sigs/gateway-api/issues/1767)

