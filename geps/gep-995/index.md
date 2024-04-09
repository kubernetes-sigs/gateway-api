# GEP-995: Named route rules

* Issue: [#995](https://github.com/kubernetes-sigs/gateway-api/issues/995)
* Status: Provisional

## TLDR

Add a new optional `name` field to the route rule types ([GRPCRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteRule), [HTTPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule), [TCPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRouteRule), [TLSRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRouteRule) and [UDPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRouteRule)) to support referencing individual rules by name.

## Goals

* Support referencing individual route rules by name from other resources, such as from metaresources ([GEP-2648](https://gateway-api.sigs.k8s.io/geps/gep-2648/#apply-policies-to-sections-of-a-resource).)
* Support referencing individual route rules by name from condition messages propagated in the status stanza of route resources as suggested in https://github.com/kubernetes-sigs/gateway-api/issues/1696#issuecomment-1666258188.
* Support referencing individual route rules by name at other observability and networking tools that are part of the ecosystem based on Gateway API.
* Provide a rather intuitive API for users of Kubernetes who are familiar with the same pattern employed already by other kinds of resources where lists of complex elements can be declared – e.g. service [ports](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/#ServiceSpec), pod [containers](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#containers) and pod [volumes](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#volumes).
* Provide a guide to the implementations about the expected behavior in cases where the name of the route rule is missing (empty value or `nil`.)

## Non-Goals

* Mandate the `name` field to be a require field.
* Limit the usage of the route rule name value for the implementations, such as exclusively for the `targetRef` section of policies (metaresources.)
* Define a patch strategy for the route objects based on rule `name`.

## Introduction

Some kinds of Gateway API types are complex types that support specifying lists of yet other complex object details within them. Examples include the [`GatewaySpec`](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.GatewaySpec) type, the [`HTTPRouteSpec`](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteSpec) type, as well as other kinds of route specification types. Specifically, `Gateway` objects can declare multiple complex listener details (`spec.listeners`); similarly, `HTTPRoute` objects may contain multiple complex routing rule details (`spec.rules`).

Even with a limited number of elements declared within those lists of resource specification details, without a field that works as a unique identifier of each element (e.g., a `name` field), referring individual ones can often lead to implementations that are inconsistent, complex, and error-prone. This is an issue for any kind of referencing pattern, including for Policy Attachment, status reporting, event logging, etc.

Referencing list elements without a unique identifier is also prone to execution errors, either when relying on how the elements are sorted in the list (i.e., based on the index) or on partial or total repetition of values of the referents. The order of elements within a list may change without necessarily any semantic reason. Complex elements can sometimes differ only subtly from each other, thus easily being overlooked when making the reference and resulting in a higher chance of typos and/or references that are possibly ambiguous or broken. In both cases, such references are fragile and can result in unexpected errors.

For the `Gateway` resource, problems above were addressed/mitigated by adding a `name` field to the [`Listener`](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.Listener) type ([#724](https://github.com/kubernetes-sigs/gateway-api/issues/).) Listener names are required and must be unique of each listener declared in a gateway. This allowed for more explicit route and policy attachment relying on _sectionName_, as well as it opened for better implementation of status reporting and log recording of events related to specific gateway listeners.

In general, declaring explicit names for complex list elements is a common pattern in Kubernetes, observed in several other APIs. Examples include [containers](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#containers) and [volumes](https://kubernetes.io/docs/reference/kubernetes-api/workload-resources/pod-v1/#volumes) of a Pod, [ports](https://kubernetes.io/docs/reference/kubernetes-api/service-resources/service-v1/#ServiceSpec) of a Service, and many others.

This GEP aims to rollout the same pattern of declarative `name` fields of these examples to the Gateway API route rule types.

## API

This GEP proposes to add a new optional `name` field to the [GRPCRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.GRPCRouteRule), [HTTPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1.HTTPRouteRule), [TCPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TCPRouteRule), [TLSRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.TLSRouteRule) and [UDPRouteRule](https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io/v1alpha2.UDPRouteRule) types.

### Format

The name of a route rule, if present, MUST be a string that begins with a single lowercase letter (`a-z`) and is followed by any number of characters, from 0 to a maximum of 252, that is either a letter, a digit, or any of the following symbols `-`, `_`. The value of the `name` field MUST be treated as case-sensitive.

A simple regular expression that can be used to test a string conforms with the format of the `name` field for route rules, compatible with Golang's [`regexp`](https://pkg.go.dev/regexp) package, is: `^[a-z][-_a-zA-Z0-9]{0,252}$`. (Provided as example)

A set of [kubebuilder](https://book.kubebuilder.io/reference/markers/crd-validation) annotations for the validation of patterns compatible with the `name` field, to be used in CRDs, is: (Provided as example)

```yaml
// +kubebuilder:validation:Pattern=`^[a-z][-_a-zA-Z0-9]*$`
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
```

This format for the `name` field of route rules differs from the pattern adopted for the [`SectionName`](https://github.com/kubernetes-sigs/gateway-api/blob/f544a46ef92b7f234ee3e7bf50da35b05f862c35/apis/v1/shared_types.go#L624C10-L624C10) type, which was thought for specifying mainly DNS subdomain names ([RFC 1123](https://www.rfc-editor.org/rfc/rfc1123)), due to its use in the gateway listeners originally.

### Volition

To preserve backward compatibility with previous version of the affected APIs, the `name` field for route rules should be introduced in the API as optional – i.e., end-user are not forced to add it to their existing or new route objects.

Implementations MAY recomend the usage of the `name` field for enabling specific features, such as for supporting policy attachment targetting individual route rules, and more assertive log messages and/or status reporting that include on the name of the rule. However, because as by API design the presence of the field is optional, implementations MUST take into account that a value may sometimes not be available. For such cases, implementations are free to decide whether to provide the feature depending the `name` field, if the feature is not required for Core compliance, or to enable the feature relying on another method of referencing of choice.

### Default value

Implementations MUST NOT assume a default value for the `name` field for route rules when not specified by the end-user.

Please refer to the [Volition](#volition) subsection for alternatives if `name` field is missing.

### Mutability

Implementations MAY prevent end-users from updating the name of a route rule. If updates are allowed the semantics and behaviour will depend on the underlying implementation.

## Alternatives

### Reference route rules by index
- Consistency can be easily broken by adding/removing new rules
- Less descriptive compared to a name

### Reference route rules by matching field values of rule
- Complexity associated with the API types of the referent bubbles up to the level of the reference
- Consistency can be easily broken due to duplication of values between reference and referent
- Does not solve for easy referencing in debugging, logging, status messages – i.e. some verbosity required to communicate what route rule the events refers to

### Break down route objects into smaller ones (with less rules)
- Could lead to an explosion of route objects containing a single rule each, thus defeating the purpose of supporting lists of rules within route objects in the first place – though large routes with many rules are also generally discouraged (https://github.com/kubernetes-sigs/gateway-api/issues/1696#issuecomment-1679804122)

### Use filters to extend behavior of specific route rules
- Does not solve for easy referencing in debugging, logging, status messages – i.e. some verbosity required to communicate what route rule the events refers to
- No support for third-parties
