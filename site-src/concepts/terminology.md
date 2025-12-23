# Terminology and related concepts

Gateway API is a complex API, solving a complex problem. The designers of Gateway API have done our best to try to add the flexibility our users demand, while keeping the operability as high as possible.

In order to do that, we've needed to create a number of new ideas and concepts, and to emphasize existing ideas from the rest of Kubernetes. This page goes through the most important things to learn when you are starting out in Gateway API.

On this page, important words and concepts that are used in other places in the Gateway API docs are written in **bold text**, as are other points that are really important to remember.


## Status and Conditions

One of the biggest problems when using any Kubernetes object is how to know if the state requested by that object (generally encoded in its `spec` stanza) has been accepted, and when the state has been achieved. The current status of most Kubernetes objects is stored in the `status` subresource and stanza, but in Gateway API, we've needed to lean hard into emphasizing the use of `status`.

In particular, Gateway API has leant hard on the convention of Conditions, a portable respresentation of states of any given object.

Conditions have:

* a `type` (a CamelCase, single-word name for the state),
* a `status` (a boolean that indicates if that state is active or not),
* a `reason` (a CamelCase, single-word reason why the Condition is or is not in the state),
* and a `message` (a string representation of the reason, that is intended for human consumption).

One optional field that Gateway API _requires_ is the `observedGeneration` field, which indicates the value of the autoincremented `metadata.generation` field on the object at the time the status was written. This functions as a staleness detection checksum - for any Gateway API status, you should check that the `observedGeneration` on its `conditions` matches the `metadata.generation` field.

If it does not, then that status is out of date, and for some reason your Gateway API implementation is not updating status correctly. (This could be a controller fault, or the object may have fallen out of the implementation's scope.)

Additionally, part of the purpose of Gateway API is to fix some of the problems of earlier approaches, and we wanted to avoid the requirement to be able to look at the logs of an implementation to see what is happening.

We want the state of your object to be, as far as possible, visible _on_ your object.

This leads us to the first, most important rule of using Gateway API:

!!! info
    **_When troubleshooting Gateway API objects, always check the `status.conditions` of the object first._**

Every Gateway object has a `conditions` array in its `status` somewhere, and most have it at `status.conditions`.

We've also tried to re-use the same Condition `type`s as far as possible, and have a few commonly-used Conditions across multiple objects:

* `Accepted`: True when the object is semantically and syntactically valid, will produce some configuration in any underlying data plane, and has been accepted by a controller.
* `Programmed`: True when an object's config has been fully parsed, and has been successfully sent to a data plane for configuration. It will be ready "soon", where soon can have different definitions depending on the exact implementation.
* `ResolvedRefs`: True when all references to other objects inside an object are valid, in that the objects referred to exist, and each is a valid reference for the field where it is used.

In some cases, such as the Gateway object, there are additional Conditions arrays - on the Gateway object, there is also a Condition per `listener` field, as that status is also complex enough to need further clarification.

Conditions are complex enough to be difficult to summarize in a single line, so most `kubectl get` commands cannot summarize them correctly.

To check the status, you have a few options:

* `kubectl get -o yaml` - this will get you the full object, in YAML format, which includes the `status`.
* `kubectl describe` - this will get you a more readable version of the full output, which can usually parse Conditions arrays corectly and show them
* `gwctl` is a command-line tool created by the Gateway API subproject, which is designed to make managing Gateway API resources easier. It's available at (link)

### Scope and Status

One other peculiarity of Gateway API is that it is designed to allow for multiple implementations to run in the same cluster. In order to do this,
there are strict requirements about what objects an implementation can
update the status for. This is referred to as an object **being in scope** for a particular implementation.

!!! info
    **If an implementation cannot establish a chain of ownership from any
    object to a GatewayClass it owns, then the object is not in scope for that implementation, and MUST NOT have its status updated by it.**

This is so that multiple implementations do not end up fighting over status,
repeatedly attempting to update the status, only to have one implementation's changes overwritten by another, and so on.

One important effect of this is that if a Route has a `parentRef` that does not point to a valid parent, then **there will be no status update to indicate that**. The implementation cannot tell you that you made a mistake
in pointing to a Gateway it cares about, because it has no way of knowing if that parentRef is its responsibility or not.

To put this another way, **a Route with an invalid parentRef will have no status to indicate that**. You should _always_ expect to see a status update for _any_ change in an in-scope object, even if it's just updating the `observedGeneration`.

## Gateways, Listeners, and Routes

As explained in the [general Gateway API concepts](api-overview.md), Gateway API depends on a directed graph of resources, that pivots around the Gateway object.

<img src="/images/resource-model.png" alt="Gateway API Resource Model" class="center" />

One of the critical parts of this graph is the relationship between Gateways and Routes.

Gateway objects bind one or more **Addresses** to one or more **Listeners**.

**Addresses** are how the Gateway is reached, and are usually IP addresses, although some implementations (particularly ones that route traffic via one of Amazon's load balancers) use domain names instead.

**Listeners** describe how the Gateway should listen for traffic, and have a `port`, a `protocol`, and other protocol-specific details. Listeners that are not **distinct** are in conflict, and Gateway API includes instructions for what happens in various conflict cases. What makes Listeners distinct is a bit complicated and is discussed in the Distinctiveness section (link).

A critical reason for the requirement that Listeners are distinct is that traffic flowing through a Gateway **must only match a single Listener**.
Any particular traffic must only be able to be asssigned to a single Listener, and once that Listener is chosen, the traffic **must** be routable via an attached, protcol-specific Route, **or it must be dropped by the Gateway**.

The most important outcome here is that traffic can't fail to be routed by one Listener, then fall back to another for further processing.

However, for the Gateway -> Route relationship, the most important thing is that Routes **attach** to one or more Listeners on the the Gateway.

## Route attachment

For a Route to attach to a Listener, two things have to be true:

* The Route must reference a Gateway (or one of its Listeners) in its `parentRefs` stanza. All Routes **must** contain a `parentRefs` stanza to be a Gateway API compliant Route object.
* The Listener must _accept_ the Route's attachment. There are many ways for Listeners to describe the shape of what Routes should attach to a given Listener, and some attachments require agreement between the Listener and the Route about protocol-specific details (for example, matching hostnames between a Listener and a HTTPRoute).

The Route side of this is intended to ensure that, at all times, control over where an application is exposed is in the hands of Ana, the Application Developer. Chihiro, the Cluster Admin and Gateway owner, can control what _sorts_ of Routes can attach to their Gateway, but they can't force Ana to do anything. The final decision is always Ana's.

The Listener side is intended to ensure that Ana's Route produces a valid confguration, and also matches Chihiro's requirements (if any) about what traffic is allowed to be exposed on that Gateway.

## Distinctiveness

For the property that traffic must only match _single_ Listener to be true, then Listeners must be **distinct**.

Gateway API defines distinctiveness according to the protocol selected on the Listener. For any given protocol, some fields (particularly `port`) MAY be shared, but exactly what that means depends on the protocol.

Listeners that are _not_ distinct are **Conflicted**, and may not be present on the same Gateway. If they are, then the entire Gateway is not valid and will not be `Accepted` for processing.

(Feel free to skip these exact definitions below if you are just starting out, they are very important, but you can learn by doing as well).

??? example "Distinctiveness rules, in increasing order of complexity"

    * `TCP` and `UDP` Listeners are distinct only on the combination of `protocol` and `port`. So, two Listeners that listen on port `53`, but where one has `protocol` set to `TCP` and the other `UDP` are distinct, but two Listeners that both have `protocol` `TCP`, and `port` `22` are not distinct, and are thus Conflicted.
    * `TLS` Listeners are distinct based on the combination of the fields `protocol`, `port`, and `hostname`. In this case, the `hostname` describes the Server Name Indicator (SNI) used as part of the TLS handshake, which can be used as a routing discriminator. Note that, in this case the TLS config in the `tls` stanza is _not_ relevant, as that allows you to specify whether or not the connection is terminated, which is not relevant in this case - because only the SNI, which applies in either terminated or not terminated cases, is defined.
    * `HTTP` Listeners are distinct via the combination of the fields `protocol`, `port`, `hostname`. Two Listeners that both expose `HTTP` on port `80`, with different `hostname` fields, are distinct.
    * `HTTPS` Listeners are distinct via the combination of the fields `protocol`, `port`, `hostname`, and the `tls` stanza, particularly the `tls.mode` field. If the protocol is `HTTPS`, then you must have a Secret reference that points to a Kubernetes Secret of type `kubernetes.io/tls`. Listeners with different `hostnames` may point to different Secrets, but we don't mandate that (since a single certificate can support many hostnames). So, two Listeners that expose `HTTPS` on port `443` are distinct if they have different `hostname` fields.

## Listener selection

When Routes attach, the process can be thought of a negotiation between the Listeners selected by the Route in its `parentRef`, and the settings on those Listeners that restrict what Routes can attach.

When considering the `parentRef` of a Route, the set of Listeners that the Route _may_ attach to is called the set of **relevant** Listeners.

A Route that specifies a `parentRef` of a Gateway that contains multiple Listeners is effectively attempting to attach to _all_ Listeners on that Gateway, and _all_ Listeners are **relevant**.

A Route that specifies a `parentRef` of a Gateway, with a `sectionName` set, the only **relevant** Listener is the one has a `name` field that matches that `sectionName`. If there are no Listeners that have a matching
`name` field, then the set of **relevant** Listeners is empty, and that `parentRef` will be ignored.

Note that there can be other ways that Routes will fail to attach as well - for example, HTTPRoutes can also match Listeners based on Hostname Intersection, which is explained in a separate page.

When it comes to Listeners, however, A Listener has two main ways to restrict what Routes may attach:

* Restrict Route Groups or Kinds (so you can say "only HTTPRoutes", for example. Or you can make your own Route and only allow that).
* Restrict what Namespaces Routes can attach from, with valid ways being "All Namespaces", "Same namespace as the Gateway", or "Namespace Label Selector".

The Route **relevant** Listeners need to match these restrictions for the Route attachment to succeed.

When the attachement _does_ succeed, that Route is counted in the `attachedRoutes` field in the Listener status. That field records the total number of successful Route attachements to that specific Listener.


## Routing discriminators

One of the most important use cases for Gateway API is to allow Ana the Application Developer and her peers to be able to multiplex Routes onto the same Listener. Each Route type has different information that can be used to accomplish this, and these sets of information are called **Routing Discriminators**.

More precisely, **Routing discriminators** are the information that can be used to allow multiple Routes to share a single port on a Listener.

This table from the [API Overview](link) page summarizes the various Route types included in Gateway API, with their **Routing Discriminators**.

|Object|Protocol|OSI Layer|Routing Discriminator|TLS Support|Purpose|
|------|--------|---------|---------------------|-----------|-------|
|HTTPRoute|HTTP or HTTPS| Layer 7 | Anything in the HTTP Protocol | Terminated only | HTTP and HTTPS Routing|
|TLSRoute|TLS| Somewhere between layer 4 and 7| SNI or other TLS properties| Passthrough or Terminated | Routing of TLS protocols including HTTPS where inspection of the HTTP stream is not required.|
|GRPCRoute|HTTP or HTTPS| Layer 7 | Anything in the gRPC Protocol | Terminated only | gRPC Routing over HTTP/2 and HTTP/2 cleartext|
|TCPRoute|TCP| Layer 4| None | Passthrough or Terminated | Allows for forwarding of a TCP stream from the Listener to the Backends |
|UDPRoute|UDP| Layer 4| None | None | Allows for forwarding of a UDP stream from the Listener to the Backends. |

Notably, because of their lack of other **routing discriminators**, TCPRoute and UDPRoute can only have a _single_ Route attached to any particular Listener.

## Traffic matching

The requirement that traffic flowing through a Gateway must _only_ match a single Listener also applies by extension to Route-based configuration.

In most uses, the way that traffic is matched against the various objects is as follows:

* Traffic flows into an IP address, selecting a Gateway (since only Gateways have addresses).
* On that IP address, the traffic is destined for a port, which selects one or more listeners. If it's more than one, then further information can also be used:
    * `hostname` is used for HTTP, HTTPS, and TLS and their associated Routes
* A single candidate Route is chosen for that traffic to flow through. (In the case of Route _match_ conflicts between Routes, then the oldest Route's match is chosen).

An important corollary of these requirements is that **if traffic does not match any traffic specified in some Routes, it cannot go back to another Listener that also matches for rerouting**.

For example, if a Gateway has two `HTTP` Listeners, one for `specific.example.com`, and one for `*.example.com`, traffic for `specific.example.com` _must_ be captured by HTTPRoutes attached to the `specific.example.com` Listener, or it will receive a 404.

So, if the Gateway and Routes look like this:
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: example-com
  namespace: default
spec:
  listeners:
    - name: specific
      hostname: specific.example.com
      protocol: HTTP
      port: 80
    - name: wildcard
      hostname: *.example.com
      protocol: HTTP
      port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: specific
  namespace: default
spec:
  parentRefs:
    - name: example-com
  rules:
    - matches:
      - path:
          type: Exact
          value: /specific
      backendRefs:
      - name: specific
        port: 8080
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: wildcard
  namespace: default
spec:
  parentRefs:
    - name: example-com
  rules:
    - matches:
      - path:
          type: prefix
          value: /
      backendRefs:
      - name: prefix
        port: 8080


```

Then the `specific` Route will _only_ match traffic bound to the URL `http://specific.example.com/specific`.

Other requests, like one to `http://specific.example.com/otherpath` will return a 404, even though they _could_ be understood to match the wildcard listener (since `*.wildcard.com` also matches `specific.example.com`, and `/otherpath` matches the `/` prefix path on the `wildcard` HTTPRoute). However, because of the Listener single-matching property, traffic cannot _also_ match HTTPRoutes attached to _other Listeners_.
