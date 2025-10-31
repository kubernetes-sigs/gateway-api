# GEP-4012: API documentation and readability

* Issue: [#4012](https://github.com/kubernetes-sigs/gateway-api/issues/4012)
* Status: Memorandum

**What**: Guide how to write API documentation and comments on Go structures / godoc
**Who**: Gateway API developers
**Why**: To provide meaningful information to the different Gateway API users without
leaking implementation details

## About this GEP

This GEP aims to guide Gateway API developers on how to write `godoc` for API
fields and structures in a meaningful and concise way, where information are
provided for the different Gateway API personas (Ian, Chihiro and Ana) without
leaking implementation details.

The implementation details are still important for a Gateway API implementation
developer, and they should still be provided but without being exposed on the
CRD generation, that can end leaking to users on a diverse set of ways, like
on Gateway API documentation website, or via `kubectl explain`.

Additionally, it is worth noticing that API documentation reflects on the CRD generation
size, which impacts directly on resource consumption like a maximum Kubernetes resource size 
(which is limited by etcd maximum value size) and avoiding problems with `last-applied-configuration` 
annotation, when doing a client-side apply.

This proposal defines two kinds of documentation:

* User facing - MUST define how a user should be consuming an API and its field, on a concise way.
* Developer facing - MUST define how a controller should implement an API and its fields. 

## Goals

* Define what should be considered when writing user facing API comments.
* Define what should be considered when writing developer facing API comments.
* Define how a Gateway API developer should separate the user facing comments
from an implementation specific comment.

## Non goals

TBD

## User facing Documentation

The API documentation, when meaningful, helps users of it on doing proper configuration
in a way that Gateway API controllers react and configure the proxies the right way.

A good API documentation should cover:
* What is the main feature of the API and Field - Eg.: "`Foo` allows configuring how a
a header should be forwarded to backends"
* What is the support level of the field - Eg.: "Support: Core/Extended/Implementation Specific"
* Caveats of that field - Eg.: "Setting `Foo` field can have conflicts with `Bar` field, and in this
case it will be shown as a Condition". (we don't need to cover all the conditions).

In a simple way, a user reading the field documentation should understand, on one or two 
phrases what happens when the field is configured, what can be configured and what are 
the impacts of that field

When adding a documentation, it is very important to remove your "Developer hat" 
and put yourself on a user that is trying to solve a problem: Does setting a field
solves my needs? How can I use it?

On an implementation, a user facing documentation belongs to the field documentation. Taking
`Listeners`, one of the most complex fields as an example:

```golang
// Listeners define logical endpoints that are bound on this Gateway's addresses.
// At least one Listener MUST be specified. When setting a Listener, conflicts can
// happen depending on its configuration like protocol, hostname and port, and in 
// this case a status condition will be added representing what was the conflict.
// 
// The definition of a Listener protocol implies what kind of Route can be attached 
// to it
Listeners []Listener `json:"listeners"`
```

We don't specify what are the Protocol types (saving this to the `Protocol` field),
what a hostname means, when a TLS configuration is required. All of these information
belongs to each field, so when a user does something like `kubectl explain gateway.spec.listeners`
they will also get the information of each field.

## Developer facing documentation

Developer facing documentation helps during implementations to define the expected
behavior of it, and should answer questions like:

* How that field should be reconciled?
* What conditions should be set during the reconciliation? 
* What should be validated during the reconciliation of that field?

In this case, as the API documentation serves as a guide for implementors on how 
their implementations should behave, it is very important to be as much verbose as
required to avoid any ambiguity. These information are used also to define expected
conformance behavior, and can even point to existing GEPs so a developer looking 
at it can know where to look for more references on what and why are those the expected
behavior of this field.

Still taking the `Listeners` field as an example, it does good definitions of situations 
like:

* Two listeners have different protocols but the same hostname. Should this be a conflict?
* A listener of type `XXX` sets the field `TLS`. Is this a problem? How to expose this to 
the user?

Because these information don't matter for a user, they should be hidden from the CRD/OpenAPI
generation and also from the website API Reference.

This can be achieved putting these information between the tags 
`<gateway:util:excludeFromCRD></gateway:util:excludeFromCRD>` and preferably 
contain a callout that those are a Note for implementors:

```golang
// Mode defines the TLS behavior for the TLS session initiated by the client.
// There are two possible modes:
//
// - Terminate: The TLS session between the downstream client and the
//   Gateway is terminated at the Gateway. This mode requires certificates
//   to be specified in some way, such as populating the certificateRefs
//   field.
// - Passthrough: The TLS session is NOT terminated by the Gateway. This
//   implies that the Gateway can't decipher the TLS stream except for
//   the ClientHello message of the TLS protocol. The certificateRefs field
//   is ignored in this mode.
//
// Support: Core
//
// <gateway:util:excludeFromCRD>
// Notes for implementors:
//
// Setting TLSModeType to Passthrough is only supported on Listeners that are of 
// type HTTP, HTTPS and TLS. In case a user sets a different type, the implementation
// MUST set a condition XXX with value XXX and a message specifying why the condition 
// happened.
// </gateway:util:excludeFromCRD>
Mode *TLSModeType `json:"mode,omitempty"`
```

## Advices when writing the documentation
As an advice, the person writing the documentation should always being questioning:

**As a user**:
* Does the documentation provide meaningful information and removes any doubt 
about what will happen when setting this field?
* Does the documentation provide information about where should I look if something
goes wrong?
* If I do `kubectl explain` or look into the API Reference, do I have enough
information to achieve my goals without being buried with information I don't care?

**As a developer/implementor**:
* Does the documentation provide enough information for another developer on 
how they should implement their controller?
* Does the documentation provide enough information on what other fields/resources
should be verified to provide the right behavior?
* Does the documentation provide enough information on how I should signal to the 
users what went right/wrong and how to fix it?

It is important to exercise changing the personas for which you are writing the 
documentation.

## References

TBD
