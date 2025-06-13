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
  Gateway for every Route, and without mutating Routes.

- Give Ana an easy way to determine which Gateway is the default, and which of
  her Routes are bound to it.

- Support multiple Gateways in a cluster, with exactly one of them being the
  default Gateway.

- Allow [Chihiro] to retain control over which Gateway is the default, so that
  they can ensure that it meets their requirements for security, performance,
  and other operational concerns.

- Allow Chihiro to rename, reconfigure, or replace the default Gateway without
  breaking existing Routes.

- Allow Chihiro to control which Routes may bind to the default Gateway, and
  to enumerate them.

- Support easy interoperation with common CI/CD and GitOps workflows.

## Non-Goals

- Support multiple "default" Gateways in a single cluster. If Ana has to make
  a choice about which Gateway she wants to use, she'll need to be explicit
  about that.

- Allow Ana to override Chihiro's choice for the default Gateway for a given
  Route without explicitly specifying the Gateway.

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

We can also find prior art where the lack of explicitness _is_ a problem: for
example, the North American telephone system. North American telephone numbers
are always ten digits long, split into three parts: the three-digit area code,
the three-digit exchange number, and the four-digit line number. Many years
ago, you could place a phone call by dialing just the exchange and line
numbers, if the area code was the same as your own. People found this
convenient, but over the years it led to massive confusion as area codes
became full and had to be split. Today, you always have to specify the area
code when using a phone in North America: it's more digits, but it's less
operationally complex, and it removes actual problematic ambiguity.

It's rare to find systems that are purely one or the other: in general, the
trick is to find a usable balance between explicitness and simplicity, while
managing ambiguity.

[Chihiro]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://https//gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian

### The Problems

## API

Most of the API work for this GEP is TBD at this point. The challenge is to
find a way to allow Ana to use Routes without requiring her to specify the
Gateway explicitly, while still allowing Chihiro and Ian to retain control
over the Gateway and its configuration.

An additional concern is CD tools and GitOps workflows: any solution that
mutates the `spec` of a Kubernetes resource tends to cause problems for these
tools. While it's possible to tell tools like Argo CD and Flux to ignore
certain elements of the `spec` that are known to mutate, the `parentRef` is
definitely _not_ a field where changes should be ignored! This strongly
implies that the ideal API will not mutate the `spec` of the Route.

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

## References
