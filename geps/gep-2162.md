# GEP-2162: Supported features in GatewayClass Status

* Issue: [#2162](https://github.com/kubernetes-sigs/gateway-api/issues/2162)
* Status: Experimental

## TLDR

This GEP proposes to enhance the [GatewayClassStatus](https://github.com/kubernetes-sigs/gateway-api/blob/f2cd9bb92b4ff392416c40d6148ff7f76b30e649/apis/v1beta1/gatewayclass_types.go#L185) to include a list of Gateway API features supported by the installed GatewayClass. 

## Goals

* Improve UX by enabling users to easily see what features the implementation (GatewayClass) support.

* Standardize features and conformance tests names.

* Automatically run conformance tests based on the supported features populated in GatewayClass status.

* Provide foundation for tools to block or warn when unsupported features are used.


## Non-Goals

* Validate correctness of supported features published by the implementation.
    Meaning we don't intend to verify whether the supported features reported by
    the implementation are indeed supported.

    However, the supported features in the status of the GatewayClass should
    make it very easy for any individual to run conformance tests against the
    GatewayClass using our conformance tooling.

## Introduction

The current [GatewayClassStatus](https://github.com/kubernetes-sigs/gateway-api/blob/f2cd9bb92b4ff392416c40d6148ff7f76b30e649/apis/v1beta1/gatewayclass_types.go#L185) is only used to store conditions the controller publishes.

Partnered with the [Conformance Profiles](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-1709.md) work, we want to:

1. Improve UX by enabling users to easily see what features the implementation(GatewayClass) support.
1. Standardize features and conformance tests names.
1. Automatically run conformance tests based on the supported features populated in GatewayClass status.
1. Potentially build tooling to block or warn when unsupported features are used (more under [Future Work](#future-work)).

This doc proposes to enhance the GatewayClassStatus API so implementations could publish a list of features they support/don't support.

Implementations **must** publish the supported features before Accepting the GatewayClass, or in the same operation.

Implementations are free to decide how they manage this information. A common approach could be to maintain static lists of supported features or using predefined sets.

Note: implementations must keep the published list sorted in ascending alphabetical order.

## API

This GEP proposes API changes describes as follow:

* Update the `GatewayClassStatus` struct to include a string-represented list of `SupportedFeatures`.


```go
// GatewayClassStatus is the current status for the GatewayClass.
type GatewayClassStatus struct {
    // Conditions is the current status from the controller for
    // this GatewayClass.
    //
    // Controllers should prefer to publish conditions using values
    // of GatewayClassConditionType for the type of each Condition.
    //
    // +optional
    // +listType=map
    // +listMapKey=type
    // +kubebuilder:validation:MaxItems=8
    Conditions []metav1.Condition `json:"conditions,omitempty"`

    // SupportedFeatures is the features the GatewayClass support.
    // <gateway:experimental>
    // +kubebuilder:validation:MaxItems=64
    SupportedFeatures []string `json:"supportedFeatures,omitempty"`
}
```

## Understanding SupportedFeatures field

Its important to define how we read the list of `SupportedFeatures` we report.

We have no supported features for core features. If an implementation reports a resource name e.g `HTTPRoute` as a supportedFeature it means it supports all its core features.
In other words, supporting the resource's core features is a requirement for the implementation to say that it supports the resource.

For Extended/Implementation-specific features we have the supported features names.

An example of a GatewayClass Status with the SupportedFeatures reported would look like:

```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: GatewayClass
...
status:
  conditions:
  - lastTransitionTime: "2022-11-16T10:33:06Z"
    message: Handled by XXX controller
    observedGeneration: 1
    reason: Accepted
    status: "True"
    type: Accepted
  supportedFeatures:
    - HTTPRoute
    - HTTPRouteHostRewrite
    - HTTPRoutePortRedirect
    - HTTPRouteQueryParamMatching

```
## Standardize features and conformance tests names

Before we add the supported features into our API, it is necessary to establish standardized naming and formatting conventions.

### Formatting Proposal

#### Feature Names

Every feature should:

1. Start with the resource name. i.e HTTPRouteXXX
2. Follow the PascalCase convention. Note that the resource name in the string should come as is and not be converted to PascalCase, i.e HTTPRoutePortRedirect and not HttpRoutePortRedirect.
3. Not exceed 128 characters.
4. Contain only letters and numbers

#### Conformance test names

Conformance tests file names should try to follow the the `pascal-case-name.go` format.
For example for `HTTPRoutePortRedirect` - the test file would be `httproute-port-redirect.go`.

We should treat this guidance as "best effort" because we might have test files that check the combination of several features and can't follow the same format.

In any case, the conformance tests file names should be meaningful and easy to understand.


## Followups

Before we make the changes we need to;

1. Change the names of the supported features and conformance tests that don't conform with the formatting rules.


## Alternatives

### Re-using ConformanceProfiles structs

We could use the same structs as we do in conformance profile object, more specifically, the [ProfileReport](https://github.com/kubernetes-sigs/gateway-api/blob/main/conformance/apis/v1alpha1/profilereport.go#LL24C6-L24C19) struct.

Though it would be nice to have only one place to update, these structs seems to include much more data relevant to the conformance report but not for our use case. 

That said, conformance profiles are still at experimental stage, we could explore the option to create a shared struct that will be used both for the conformance reports and for the GatewayClass status.

### Instruct users to read from the future conformance profiles report

The current plan for conformance profiles is to also include centralized reporting. (more info in [gep-1709](https://github.com/kubernetes-sigs/gateway-api/blob/main/geps/gep-1709.md))
We could wait for this to be implemented and instruct users to read from that source to determine what features their installed GatewayClass support.

However, having the supported features published in the GatewayClass Status adds the following values:

* We could build a mechanism or a tool to block or warn when unsupported features are used.
* Users will be able to select the GatewayClass that suits their needs without having to refer to documentation or conformance reports.

This does not cover a future piece of work we want to implement which is to warn/block users from applying a Gateway API object if the installed GWC doesn't support it. (originally suggested in [#1804](https://github.com/kubernetes-sigs/gateway-api/issues/1804)). 


## References

[discussion #2108](https://github.com/kubernetes-sigs/gateway-api/discussions/2108)
[#1804](https://github.com/kubernetes-sigs/gateway-api/issues/1804)

## Future Work

### Research the development of an unsupported feature warning/blocking mechanism
Once the GatewayClass features support are is published into the status we could look into;

1. Using the supported features in the webhook to validate or block attempts to apply manifests with unsupported features.

    * Developing such mechanism looks like it would have to include cross-resource validation. (checking the GatewayClass while trying to apply a Route for example). This comes with a lot of caveats and we will need consider it carefully.

2. Build tooling to check and warn when unsupported features are used.

### Add Gateway API Version field to the GatewayClass Status

We got some feedback that it will be useful to indicate what what Gateway API version the implementation supports. So when we have supported features published in the GatewayClass Status, users will also be able to understand that those are the supported features for a specific Gateway API version.

This work is likely to require its own small GEP but ideally what this field would mean is that an implementation supports Max(vX.X). 

The value of it is to provide a better user experience and also more foundation for tools to be able to warn for example when a GatewayClass and CRDs have mismatched versions.

### Add a table with feature name and description to document what each feature means

Create a comprehensive table detailing feature names and their corresponding descriptions, providing a clear understanding of each feature's purpose and functionality.