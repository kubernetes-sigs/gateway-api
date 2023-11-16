# GEP-995: Named route rules

* Issue: [#995](https://github.com/kubernetes-sigs/gateway-api/issues/995)
* Status: Provisional

## TLDR

Add a new `name` field to the route rule types (HTTPRouteRule, GRPCRouteRule) to support referencing individual rules by name.

## Goals

* Support referencing individual route rules by name from other resources, such as from metaresource ([GEP-713](https://gateway-api.sigs.k8s.io/geps/gep-713/).)
* Support referencing individual route rules by name from condition messages propagated in the status stanza of route resources.
* Support referencing individual route rules by name at other observability and networking tools that are part of the ecosystem based on Gateway API.
* Provide a rather intuitive API for users of Kubernetes who are familiar with the same pattern employed already by other kinds of resources where lists of complex objects can be declared – e.g. service [ports](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/#ServiceSpec), pod [containers](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#containers) and pod [volumes](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#volumes).
* Provide a guide to the implementations about the expected behavior in cases where the name of the route rule is missing (empty value or `nil`.)

## Non-Goals

* Mandate the `name` field to be a require field
* Limit the usage of the route rule name value for the implementations, such as exclusively for the `targetRef` section of policies (metaresources.)
* Define a patch strategy for the route objects based on rule `name`.

## Introduction

(Can link to external doc -- but we should bias towards copying
the content into the GEP as online documents are easier to lose
-- e.g. owner messes up the permissions, accidental deletion)

## API

(... details, can point to PR with changes)

## Conformance Details

(This section describes the names to be used for the feature or
features in conformance tests and profiles.

These should be `CamelCase` names that specify the feature as
precisely as possible, and are particularly important for
Extended features, since they may be surfaced to users.)

## Alternatives

### Reference route rules by index
- Consistency can be easily broken by adding/removing new rules
- Less descriptive compared to a name

### Reference route rules by matching field values of rule
- Consistency can be easily broken due to duplication of values between reference and referent
- Does not solve for easy referencing in debugging, logging, status messages – i.e. some verbosity required to communicate what route rule the events refers to

### Break down route objects into smaller ones (with less rules)
- Could lead to an explosion of route objects containing no more than one rule each, thus defeating the purpose of supporting lists of rules within route objects in the first place

### Use filters to extend behavior of specific route rules
- Does not solve for easy referencing in debugging, logging, status messages – i.e. some verbosity required to communicate what route rule the events refers to
- No support for third-parties

## References

(Add any additional document links. Again, we should try to avoid
too much content not in version control to avoid broken links)
