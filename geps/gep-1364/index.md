# GEP-1364: Status and Conditions Update

* Issue: [#1364](https://github.com/kubernetes-sigs/gateway-api/issues/1364)
* Status: Standard

## TLDR

The status, particularly the Conditions, across the whole Gateway API have very much
grown organically, and so have many inconsistencies and odd behaviors.
This GEP covers doing a review and consolidation to make Condition behavior consistent
across the whole API.

## Goals

* Update Conditions design to be consistent across Gateway API resources
* Provide a model and guidelines for Conditions for future new resources
* Specify changes to conformance required for Condition updates

## Non-Goals

* Define the full set of Conditions that will ever be used with Gateway API

## Introduction

Gateway API currently has a lot of issues related to status, especially that
status is inconsistent ([#1111][1111]), that names are hard to understand ([#1110][1110]),
and that Reasons aren't explained properly ([#1362][1362]).

As the API has grown, the way we talk about resources has changed a lot, and some of the
status design hasn't been updated since resources were created.

So, for example, we have GatewayClass with `Accepted`, Gateway with `Scheduled`,
the Gateway Listeners with `Detached` (which you want to be `false`, unlike the previous
two), and then Gateways and Gateway Listeners have `Ready`, but Route doesn't (and which
also you want to be `true`).

This document lays out large-scale changes to the way that we talk about resources,
and the Conditions to match them. This means that there will be an unavoidable break
in what constitutes a healthy or unhealthy resource, and there will be changes
required for all implementations to be conformant with the release that includes
these changes.

The constants that mark the deprecated types will be also marked as deprecated,
and will no longer be tested as part of conformance. They'll still be present,
and will work, but they won't be part of the spec anymore. This should give
implementations and users a release to transition to the new design (in UX terms).
This grace period should be one release (so, the constants will be removed in
v0.7.0.)

This level of change is not optimal, and the intent is to make this a one-off change
that can be built upon for future resources - since there are definitely more resources
on the way.

## Background: Kubernetes API conventions and prior art on Conditions

Because this GEP is mainly concerned with updating the Conditions we are setting in
Gateway API resources' `status`, it's worth reviewing some important points about
Conditions. (This information is mainly taken from the [Typical status properties][typstatus]
section of the API conventions document.)

1. Conditions are a standard type used to represent arbitrary higher-level status from
a controller.
2. They are a listMapType, a list that is enforced by the apiserver to have only
one entry of each item, using the `type` field as a key. (So, this is effectively
a map that looks like a list in YAML form).
3. Each has a number of fields, the most important of which for this discussion
are `type`, `status`, `reason`, and `observedGeneration`.

    * `type` is a string value indicating the Condition type. `Accepted`, `Scheduled`,
    and `Ready` are current examples.
    * `status` indicates the state of the condition, and can be one of three values,
    `true`, `false`, or `unknown`. Unknown in particular is important, because it
    means that the controller is unable to determine the status for some reason.
    (Also notable is that "" is also valid, and must be treated as `Unknown`.
    Controllers must not set the value to "", but consumers should accept it
    as meaning the same thing as `Unknown`.)
    * `reason` is a CamelCase string that is a brief description of the reason why
    the `status` is set the way it is.
    * `observedGeneration` is an optional field that sets what the `metadata.generation`
    field was when the controller last saw a resource. Note that this is optional
    _in the struct_, but is required for Gateway API conditions. This will be
    enforced in the conformance tests in the future.

4. Conditions should describe the _current state_ of the resource at observation
time, which means that they should be an adjective (like `Ready`), or a past-tense
verb (like `Accepted`). This one in particular is documented pretty closely in the
[Typical status properties][typstatus] section of the guidelines.
5. Conditions should be applied to a resource the first time the controller sees
the resource. This seems to imply that _all conditions should be present on every
resource owned by a controller_, but the rest of the conventions don't make this
clear, and it is often not complied with.
6. It's helpful to have a top-level condition which summarizes more detailed conditions.
The guidelines suggest using either `Ready` for long-running processes, or `Succeeded`
for bounded execution.

From these guidelines, we can see that Conditions can be either _positive polarity_
(healthy resources have them as `status: true`) or _negative polarity_ (healthy
resources have them as `status: false`). `Ready` is an example of a positive polarity
condition, and conditions like `Conflicted` from Listener or `NetworkUnavailable`,
`MemoryPressure`, or `DiskPressure` from the Node resource are examples of
negative-polarity conditions.

There is also some extra context that's not in the API conventions doc:

SIG-API Machinery has been reluctant to add fields that would aid in machine-parsing
of Conditions, especially fields that would indicate the polarity, because they
are intended more for human consumption than machine consumption. Probably the best
example of this was in the PR [#4521](https://github.com/kubernetes/community/pull/4521#issuecomment-64894206).

This means that there's no guidance from upstream about condition polarity. We'll
discuss this more when we talk about new conditions.

The guidance about Conditions being added as soon as a controller sees a resource
is a bit unclear - as written in the conventions, it seems to imply that _all_ 
relevant conditions should always be added, even if their status has to be set to
`unknown`.
Gateway API resources do not currently require this, and the practice seems to be
uncommon.

## Proposed changes

### Proposed changes summary

* All the current Conditions that indicate that the resource is okay and ready
for processing will be replaced with `Accepted`.
* In general, resources should be considered `Accepted` if their config is valid
enough to generate some config in the underlying data plane. Examples are provided
below.
* There will be a limited set of positive polarity summary conditions, and a number
of other specific negative-polarity error conditions.
* All relevant positive-polarity summary Conditions for a resource must be added
when it's observed.
For example, HTTPRoutes must always have `Accepted` and `ResolvedRefs`, regardless
of their state.
* Negative polarity error conditions must only be added when the error is True.
* The `Ready` condition will be moved to Extended conformance, and we'll re-evaluate
if it's used by any implementations after some time has passed. If not, it may be
removed.
* To capture the behavior that `Ready` currently captures, `Programmed` will be
introduced. This means that the implementation has seen the config, has everything
it needs, parsed it, and sent configuration off to the data plane. The configuration
should be available "soon". We'll leave "soon" undefined for now.
* Resolving a comment that came up, documentation will be added to clarify that
it's okay to add your own Conditions, and that implementations should namespace
their custom Conditions with a domain prefix (so `implementation.io/CustomType`
rather than just `CustomType`), or run the risk of using a word that's reserved later.
* It's recommended that implementations publish both new and old conditions to
provide a smoother transition, but conformance tests will only require the new
conditions.

The exact list of changes is detailed below. The next few sections detail
the reasons for these large-scale changes.

### Conceptual and language changes

Gateway API resources are, conceptually, all about breaking up the configuration for a
data plane into separate resources that are _expressive_ and _extensible_, while being
split up along _role-oriented_ boundaries.

So, when we talk about Gateway API, it's _always_ about a _system of related resources_.

We already acknowledge this when we talk about Routes "attaching" to Gateways, or Gateways
referencing Services, or Gateways requiring a GatewayClass in their spec.

However, this GEP is proposing that we move all our discussion into using
"accepted" to indicate that a resource has attached correctly enough to be
_accepted_ for processing.

So resources are `Accepted` for processing when their attachment succeeds enough
to generate some configuration. This allows us to make calls about when partially
valid objects should be accepted and when they shouldn't.

Of course, because we're using all of this configuration to describe some sort of data
path from "outside"/lacking cluster context to "inside"/enriched with cluster context,
we also need a way to describe when that data path is configured and working.

We already have a word in the Kubernetes API, but it comes with some expectations
that implementations are not currently able to meet. That word is `Ready`, but it
implies that the data path is Ready _when you read the status_, rather than that
it _will be ready soon_ (which is what most implementations can guarantee currently.)

So we have an unresolved question as to what to do with the `Ready` condition.
This is addressed further below.

### Condition polarity

In terms of the polarity of conditions, we have three options, of which only two are
really viable:
* All conditions must be negative polarity
* All conditions must be positive polarity
* Some conditions can be positive polarity, but most should be negative.

The fact that the user experience of `Ready` or conditions like `Accepted` being `true`
in the healthy case is much better rules out the first option, so we are left to
decide between enforcing that all conditions are positive, or that we have a mix.

Having an arbitrary mix will make doing machine-based extraction of information
much harder, so here I'm going to talk about the distinction between having all
conditions positive or some, summary conditions positive, and the rest negative.

#### All Conditions Positive

In this case, all Condition types are written in such a way that they're positive
polarity, and are `true` in the healthy case.

As already discussed, `Ready`, and `Accepted` are current examples, but another
one that's a little more important here is `ResolvedRefs` which is set to `true`
when all references to other resources have been successfully resolved. This is
not a _blocking_ Condition that affects the `Ready` condition, since having _some_
references valid is enough to produce some configuration in the underlying data
plane.

So, All Conditions Positive pros:

* We're close already. Most conditions in the API are currently positive polarity.
* Easier to understand - there are no double negatives. "Good: true" is less
cognitive overhead than "NotGood: false".

Cons:

* Reduces flexibility - it can surprisingly difficult to avoid double negatives for
conditions that describe error states, as in general programmers are more used
to reporting "something went wrong" than they are "everything's okay".

Not sure if pro or con:

* Leans the design towards favoring conditions always being present, since you
can't be sure if everything is good unless you see `AllGood: true`. The absence
of a positive-polarity condition implies that the condition could be false. This
puts this option more in line with the API guidelines on this point.

#### Some Conditions Positive

In this case, only a limited set of summary conditions are positive, and the rest
are negative.

Pros:

* Error states can be described with `Error: true` instead of `NoError: false`.
* Negative polarity error conditions are more friendly to not being present (since
absence of `Error: true` implies everything's okay).

Cons:

* Any code handling conditions will need a list of the positive ones, and will
need to assume that any others are negative.

#### Decision

Gateway API conditions will be positive for conditions that describe the happy
state of the object, which is currently `Accepted` and `ResolvedRefs`, and will 
also include the new `Programmed` condition, and the newly-Extended condition
`Ready`. A separate set of negative-polarity Error conditions will be set on an
object when they are true.


### Should conditions always be added?

Not all of them.

Positive polarity Conditions that describe the desirable state of the object must
always be set. These are currently `Accepted`, `ResolvedRefs`, and `Programmed`.
Implementations that use `Ready` must also add it before programming the Route.

### Partial validity and Conditions

One of the trickiest parts of Gateway API objects is that it's very possible to
end up with an object that has some parts with valid configuration and some that
don't. We refer to this as _partial validity_, and communicating this via status
conditions is difficult.

The intent with the `Accepted` condition is that it serves as an indicator that
_something_ is working, that _some traffic_ from what the config specifies will
be routed as configured. 

At this time, we haven't added a "no errors at all present" Condition, choosing
to have a "some config is working" condition, with specific errors to aid in
finding the exact problem with the objects. We could conceivably add this later
if users find `Accepted` insufficient, but we're erring on the side of having
less positive Conditions for now.

### New and Updated Conditions

#### `Accepted`

This GEP proposes replacing all conditions that indicate syntactic and semantic
validity with one, `Accepted` condition type.

That is, the proposal is to replace:

* `Scheduled` on Gateway
* `Detached` on Listener

with `Accepted` in all these locations.

GatewayClass and Route will maintain the `Accepted` condition.

All of these conditions share the following meanings:

* The resource has been accepted for processing by the controller
* The resource is syntactically and semantically valid, and internally consistent
* The resource fits into a larger system of Gateway API resources, and there is
is no missing information, including but not limited to:
  * Any mandatory references resolve to existing resources (examples here are the
  Gateway's gatewayClass field, or the `parentRefs` field in Route resources)
  * Any specified TLS secrets exist
* The resource is supported by the controller by ensuring things like:
  * Any Kinds being referred to by the resource are supported
  * Features being used by the resource are supported

All of these rules can be summarized into:

* The resource is valid enough to produce some configuration in the underlying
data plane.

For Gateway, `Accepted` also subsumes the functions of `Scheduled`: `Accepted`
set to `true` means that sufficient capacity exists on underlying infrastructure
for the Gateway to be provisioned. If that capacity does not exist, then the
Gateway cannot be reconciled successfully, and so fails to attach to the
owning GatewayClass, and cannot be accepted.

Note that some classes of inter-resource reference failure do _not_ cause a resource
to become unattached and stop being accepted (that is, to have the `Accepted`
condition set to `status: false`).

* Non-existent Service backends - if the backend does not exist on a HTTPRoute that
is otherwise okay, then the data plane must generate 500s for traffic that matches
that HTTPRoute. In this case, the `Accepted` Condition must be true, and the
`ResolvedRefs` Condition must be false, with reasons and messages indicating that
the backend services do not exist.
* HTTPRoutes with *all* backends in other namespaces, but not permitted by ReferenceGrants.
In this case, the "non-existent service backends" rules apply, and 500s must be
generated. In this case, again, the `Accepted` condition is true, and the
`ResolvedRefs` Condition is false, with reasons and messages indicating that the
backend services are not reachable.

For ReferenceGrant or not-designed-yet Policy resources, `Accepted` means that:

* the resource has a correctly-defined set of resources that it applies to
* the resource has a syntactically and semantically valid `spec`

Note that having a correctly-defined set of resources that is empty does not make
these resources unattached, as long as it's possible to create some config in the
underlying data plane. By "empty" here we mean that there are no backends,
not that the config is incomplete or missing references. So you can have a
GatewayClass, Gateway, HTTPRoute and Service all present and referred to correctly
when there are no endpoints in the Service, and the resource will not stop being
accepted, because HTTPRoute contains rules about what to program in the data plane
if there are no endpoints (that is, it should return 500 for any matching request).

Note that for other Route types that don't have a clear mechanism like HTTP does
for indicating a server failure (like the HTTP code 500 does), not having existing
backends may not produce any configuration in the data plane, and so may cause
the resource to fail to attach. (An example here could be a TCP Route with
no backends, we need to decide if that means that a port should be opened that
actively closes connections, or if no port should be opened.)

Examples of Conditions:

* HTTPRoute with one match with one backend that is valid. `Accepted` is true,
`ResolvedRefs` is true.
* HTTPRoute with one match with one backend that is a non-existent Service backend.
The `Accepted` Condition is true, the `ResolvedRefs` condition is false, with
a reason of `BackendNotFound`. `Accepted` is true in this case because the data
path must respond to requests that would be sent to that backend with a 500 response.
* HTTPRoute with one match with two backends, one of which is a non-existent Service
backend. The `Accepted` Condition is true, the `ResolvedRefs` condition is false.
`Accepted` is true in this case because the data path must respond to a percentage
of the requests matching the rule corresponding to the weighting of the non-existent
backend (which would be fifty percent unless weights are applied).
* HTTPRoute with one match with one backend that is in a different namespace, and
does _not_ have a ReferenceGrant permitting that access. The `Accepted` condition
is true, and the `ResolvedRefs` Condition is false, with a reason of `RefNotPermitted`.
As before, `Accepted` is true because in this case, the data path must be
programmed with 500s for the match.
* TCPRoute with one match with a backend that is a non-existent Service. `Accepted`
is false, and `ResolvedRefs` is false. `Accepted` is false in this case because
there is not enough information to program any rules to handle the traffic in the
underlying data plane - TCP doesn't have a way to say "this is a valid destination
that has something wrong with it".
* HTTPRoute with one Custom supported filter added that is not supported by the
implementation. Our spec is currently unclear on what happens in this case, but
custom HTTP Filters require the use of the `ExtensionRef` filter type, and the
setting of the ExtensionRef field to the name, group, version, and kind of a
custom resource that describes the filter. If that custom resource is not supported,
it seems reasonable to say that this should be a reference failure, and be treated
like other reference failures (`Accepted` will be set to true, `ResolvedRefs` to
false with a `InvalidKind` Reason, and traffic that would have matched the filter
should receive a 500 error.)
* A HTTPRoute with one rule that specifies a HTTPRequestRedirect filter _and_ a
HTTPURLRewrite filter. `Accepted` must be false, because there's only one rule,
and this configuration for the rule is invalid (see [reference][httpreqredirect])
The error condition in this case is undefined currently - we should define it,
thanks @sunjayBhatia.
* A HTTPRoute with two rules, one valid and one which specifies a HTTPRequestRedirect
filter _and a HTTPURLRewrite filter. `Accepted` is true, because the valid rule
can produce some config in the data plane. We'll need to raise the more specific
error condition for an incompatible filter combination as well to make the partial
validity clear.


#### Ready

Currently, the `Ready` condition text for Gateway says:
```go
	// This condition is true when the Gateway is expected to be able
	// to serve traffic. Note that this does not indicate that the
	// Gateway configuration is current or even complete (e.g. the
	// controller may still not have reconciled the latest version,
	// or some parts of the configuration could be missing).
```

This is pretty unclear - how can the Gateway serve traffic if config is missing?
In the past, we've been asked to have a Condition that only flips to `true` when
*all* required configuration is present.

For many implementations (certainly for Envoy-based ones), getting this information
correctly and avoiding races on applying it is surprisingly difficult. 

For this reason, this GEP proposes that we exclude the `Ready` condition from Core
conformance, and make it a feature that implementations may opt in to - making it
an Extended condition.

It will have the following behavior:

* `Ready` is an optional Condition that has Extended support, with conformance
tests to verify the behavior.
* When it's set, the condition indicates that traffic is ready to flow through
the data plane _immediately_, not at some eventual point in the future.

We'll need to add conformance testing for this.

#### Programmed

The `Programmed` condition is being added to replicate the functionality that the
`Ready` condition currently indicates, namely that all the resources in the set
are valid enough to produce some data plane configuration, and that configuration
has been sent to the data plane, and should be ready soon.

It is a positive-polarity summary condition, and so should always be present on
the resource. It should be set to `Unknown` if the implementation performs updates
to the status before it has all the information it needs to be able to determine
if the condition is true.


## Alternatives

(Most alternatives have been discussed inline. Please comment here if this section
needs updating.)

## References
[kep-status]: https://github.com/kubernetes/enhancements/blob/master/keps/NNNN-kep-template/kep.yaml#L9

[1111]: https://github.com/kubernetes-sigs/gateway-api/issues/1111
[1110]: https://github.com/kubernetes-sigs/gateway-api/issues/1110
[1362]: https://github.com/kubernetes-sigs/gateway-api/issues/1362

[typstatus]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties
[httpreqredirect]: https://gateway-api.sigs.k8s.io/reference/spec/#gateway.networking.k8s.io%2fv1beta1.HTTPRequestRedirectFilter