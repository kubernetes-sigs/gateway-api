# Traffic Matching


## Listener selection

When Routes attach, the process can be thought of as a negotiation between the Listeners designated by the Route in its `parentRef`, and the settings on those Listeners that restrict what Routes can attach.

When considering the `parentRef` of a Route, the set of Listeners that the Route _may_ attach to is called the set of **relevant** Listeners.

A Route that specifies a `parentRef` of a Gateway that contains multiple Listeners is effectively attempting to attach to _all_ Listeners on that Gateway, and _all_ Listeners are **relevant**.

If a Route that specifies both a `parentRef` of a Gateway, and a `sectionName` in that `parentRef`, the only **relevant** Listener is the one has a `name` field that matches that `sectionName`.
If there are no Listeners that have a matching `name` field, then the set of **relevant** Listeners is empty, and that `parentRef` will be ignored.

Note that there can be other ways that Routes will fail to attach as well - for example, HTTPRoutes can also match Listeners based on Hostname Intersection, which is explained in a separate page.

When it comes to Listeners, however, A Listener has two main ways to restrict what Routes may attach:

* Restrict Route Groups or Kinds (e.g. specify "HTTPRoutes" in `AllowedRoutes.Kinds` to allow only HTTPRoutes to attach.  Or you can make your own Route and only allow that).
* Restrict what Namespaces Routes can attach from (e.g. specify "All", "Same", or "Selector" in `AllowedRoutes.Namespaces` fields).

The Route **relevant** Listeners need to match these restrictions for the Route attachment to succeed.

When the attachment _does_ succeed, that Route is counted in the `attachedRoutes` field in the Listener status.
That field records the total number of successful Route attachments to that specific Listener.

## Traffic matching

The requirement that traffic flowing through a Gateway must _only_ match a single Listener also applies by extension to Route-based configuration.

In most uses, the way that traffic is matched against the various objects is as follows:

* Traffic flows into an IP address, selecting a Gateway (since only Gateways have addresses).
* On that IP address, the traffic is destined for a port, which selects one or more listeners. If it's more than one, then further information can also be used:
    * `hostname` is used for HTTP, HTTPS, and TLS and their associated Routes
* A single candidate Route is chosen for that traffic to flow through. (In the case of Route _match_ conflicts between Routes, then the oldest Route's match is chosen).

An important corollary of these requirements is that **if traffic does not match any traffic specified in some Routes, it cannot be allowed to choose another Listener that also matches for rerouting**.

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

Other requests, like one to `http://specific.example.com/otherpath` will return a 404, even though they _could_ be understood to match the wildcard listener (since `*.wildcard.com` also matches `specific.example.com`, and `/otherpath` matches the `/` prefix path on the `wildcard` HTTPRoute).
However, because of the Listener single-matching property, traffic cannot _also_ match HTTPRoutes attached to _other Listeners_.
