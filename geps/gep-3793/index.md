# GEP-3793: Default Gateways

* Issue: [#3793](https://github.com/kubernetes-sigs/gateway-api/issues/3793)
* Status: Provisional

(See [status definitions](../overview.md#gep-states).)

## User Story

**[Ana] wants a concept of a default Gateway.**

Gateway API currently requires every north/south Route object to explicitly
specify its parent Gateway. This is helpful in that it removes ambiguity, but
it's less helpful in that [Ana] is stuck constantly explicitly configuring a
thing that she probably doesn't care much about: in a great many cases, Ana
just wants to create a Route that "works from the outside world" and she
really doesn't care what the Gateway is called.

Therefore, Ana would like a way to be able to rely on a default Gateway that
she doesn't have to explicitly name, and can simply trust to exist.

[Ana]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana

## Goals

- Give Ana a way to use Gateway API without having to explicitly specify a
  Gateway for every Route, ideally without mutating Routes.

- Give Ana an easy way to determine which Gateway is the default, and which of
  her Routes are bound to it.

- Continue supporting multiple Gateways in a cluster, while allowing exactly
  one of them to be the default Gateway.

- Allow [Chihiro] to retain control over which Gateway is the default, so that
  they can ensure that it meets their requirements for security, performance,
  and other operational concerns.

- Allow Chihiro to choose not to provide a default Gateway.

- Allow Chihiro to rename, reconfigure, or replace the default Gateway at
  runtime.

  - If Chihiro renames the default Gateway, Routes using the default Gateway
    MUST remain bound to the new default Gateway. Ana shouldn't need to go
    recreate all her Routes just because Chihiro is being indecisive.

  - Determine how (or if) to signal changes in functionality if the default
    Gateway implementation is changed. For example, suppose that Chihiro
    switches the default Gateway from an implementation that supports the
    `HTTPRoutePhaseOfTheMoon` filter to an implementation that does not.

    (Note that this problem is not unique to default Gateways; it affects
    explicitly-named Gateways as well.)

- Allow Chihiro to control which Routes may bind to the default Gateway, and
  to enumerate which Routes are currently bound to the default Gateway.

- Support easy interoperation with common CI/CD and GitOps workflows.

- Define how (or if) listener and Gateway merging applies to a default
  Gateway.

## Non-Goals

- Support multiple "default" Gateways in a single cluster. If Ana has to make
  a choice about which Gateway she wants to use, she'll need to be explicit
  about that.

  Loosening this restriction later is a possibility. For example, we may later
  want to consider allowing a default Gateway per namespace, or a default
  Gateway per implementation running in a cluster. However, these examples are
  not in scope for this GEP, in order to have a fighting chance of getting
  functionality into Gateway API 1.4.

- Allow Ana to override Chihiro's choice for the default Gateway for a given
  Route without explicitly specifying the Gateway.

- Require that every possible routing use case be met by a Route using the
  default Gateway. There will be a great many situations that require Ana to
  explicitly choose a Gateway; the existence of a default Gateway is not a
  guarantee that it will be correct for any given use case.

- Allow for "default Gateway" functionality without a Gateway controller
  installed. Just as with any other Gateway, a default Gateway requires an
  implementation to be installed.

## Overview

Gateway API currently requires every north/south Route object to explicitly
specify its parent Gateway. This is a wonderful example of a fundamental
tension in Gateway API:

- [Chihiro] and [Ian] value _explicit definition_ of everything, because it
  makes it easier for them to reason about the system and ensure that it meets
  the standards they set for it.

- [Ana], on the other hand, values _simplicity_ and _ease of use_, because
  she just wants to get her job done without having to think about every little
  detail.

At present, Gateway API is heavily weighted towards the point of view of
Chihiro and Ian. This causes friction for Ana: for example, she can't write
examples or documentation for her colleagues (or her counterparts at other
companies) without telling them that they'll need to be sure to edit the
Gateway name in every Route. Nor can she write a Helm chart that includes a
Route without requiring the person using the chart to know the specific name
for the Gateway to use.

The root cause of this friction is a difference in perspective: to Chihiro and
Ian, the Gateway is a first-class thing that they think about regularly, while
to Ana, it's an implementation detail that she doesn't care about. Neither
point of view is wrong, but they are in tension with each other.

### Prior Art

This is very much not a new problem: there are many other systems out there
where being unambiguous is crucial, but where being completely explicit is a
burden. One of the simplest examples is the humble URL, where the port number
is not always explicit, but it _is_ always unambiguous. Requiring everyone to
type `:80` or `:443` at the end of the host portion of every URL wouldn't
actually help anyone, though allowing it to be specified explicitly when
needed definitely does help people.

The Ingress resource, of course, is another example of prior art: it permitted
specifying a default IngressClass, allowing users to create Ingress resources
that didn't specify the IngressClass explicitly. As with a great many things
in the Ingress API, this caused problems:

1. Ingress never defined how conflicts between multiple Ingress resources
   should be handled. Many (most?) implementations merged conflicting
   resources, which is arguably the worst possible choice.

2. Ingress also never defined a way to allow users to see which IngressClass
   was being used by a given Ingress resource, which made it difficult for
   users to understand what was going on if they were using the default
   IngressClass.

(Oddly enough, Ingress' general lack of attention to separation of concerns
wasn't really one of the problems here, since IngressClass was a separate
resource.)

It's rare to find systems that are completely explicit or completely implicit:
in practice, the trick is to find a usable balance between explicitness and
simplicity, while managing ambiguity.

### Debugging and Visibility

It's also critical to note that visibility is critical when debugging: if Ana
can't tell which Gateway is being used by a given Route, then her ability to
troubleshoot problems is _severely_ hampered. Of course, one of the major
strengths of Gateway API is that it _does_ provide visibility into what's
going on in the `status` stanzas of its resources: every Route already has a
`status` showing exactly which Gateways it is bound to. Making certain that
Ana has easy access to this information, and that it's clear enough for her to
understand, is clearly important for many more reasons than just default
Gateways.

[Chihiro]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian

## API

Most of the API work for this GEP is TBD at this point. The challenge is to
find a way to allow Ana to use Routes without requiring her to specify the
Gateway explicitly, while still allowing Chihiro and Ian to retain control
over the Gateway and its configuration.

An additional concern is CD tools and GitOps workflows. In very broad terms,
these tools function by applying manifests from a Git repository to a
Kubernetes cluster, and then monitoring the cluster for changes. If a tool
like Argo CD or Flux detects a change to a resource in the cluster, it will
attempt to reconcile that change with the manifest in the Git repository --
which means that changes to the `spec` of an HTTPRoute that are made by code
running in the cluster, rather than by a user with a Git commit, can
potentially trip up these tools.

These tools generally ignore strict additions: if a field in `spec` is not
present in the manifest in Git, but is added by code running in the cluster,
the tools know to ignore it. So, for example, if `spec.parentRefs` is not
present at all in the manifest in Git, CD tools can probably tolerate having a
Gateway controller write a new `parentRefs` stanza to the resource.

There has been (much!) [discussion] about whether the ideal API for this
feature will mutate the `parentRefs` of a Route using a default Gateway to
reflect the Gateway chosen, or whether it should not, relying instead on the
`status` stanza to carry this information. This is obviously a key point that
will need resolution before this GEP can graduate.

[discussion]: https://github.com/kubernetes-sigs/gateway-api/pull/3852#discussion_r2140117567

### Gateway for Ingress (North/South)

### Gateway For Mesh (East/West)

## Conformance Details

#### Feature Names

The default-gateway feature will be named `HTTPRouteDefaultGateway` and
`GRPCRouteDefaultGateway`. It is unlikely that an implementation would support
one of these Route types without the other, but `GatewayDefaultGateway` does
not seem like a good choice.

### Conformance tests

## Alternatives

A possible alternative API design is to modify the behavior of Listeners or
ListenerSets; rather than having a "default Gateway", perhaps we would have
"[default Listeners]". One challenge here is that the Route `status` doesn't
currently expose information about which Listener is being used, though it
does show which Gateway is being used.

[default Listeners]: https://github.com/kubernetes-sigs/gateway-api/pull/3852#discussion_r2149056246

## References
