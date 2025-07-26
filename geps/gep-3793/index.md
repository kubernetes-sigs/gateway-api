# GEP-3793: Default Gateways

* Issue: [#3793](https://github.com/kubernetes-sigs/gateway-api/issues/3793)
* Status: Implementable

(See [status definitions](../overview.md#gep-states).)

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD",
"SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this
document are to be interpreted as described in BCP 14 ([RFC8174]) when, and
only when, they appear in all capitals, as shown here.

[RFC8174]: https://www.rfc-editor.org/rfc/rfc8174

## User Story

**[Ana] wants a concept of a default Gateway.**

Gateway API currently requires every north/south Route object to explicitly
specify its parent Gateway. This is helpful in that it removes ambiguity, but
it's less helpful in that [Ana] is stuck constantly explicitly configuring a
thing that she probably doesn't care much about: in a great many cases, Ana
just wants to create a Route that "works from the outside world" and she
really doesn't care what the Gateway is called.

Therefore, Ana would like a way to be able to rely on a default Gateway that
she doesn't have to explicitly name, and can simply trust to exist. Ana
recognizes that this will involve **giving up** a certain amount of control
over how requests reach her workloads. She's OK with that, and she understands
that it means that relying on a default Gateway is not always appropriate: for
example, if she needs to be sure that her Route is protected by specific
authorization policies, she should confer with Chihiro to make sure that she
explicitly specifies a Gateway that meets those requirements.

In the future, it may also be important to distinguish different kinds of
default Gateways -- for example, a default ingress Gateway or a default egress
Gateway. This GEP deliberately defines only a single _scope_ of default
Gateway (`All`) but recognizes the need to at least consider the possibility
of multiple scopes in the future.

## Definitions

- **defaulted Route**: a Route that Ana creates without explicitly specifying
  a Gateway

- **default Gateway**: a Gateway that Chihiro has configured to accept
  defaulted Routes

- **default Gateway scope**: the scope within which a default Gateway is
  applicable

## Goals

- Give Ana a way to use Gateway API without having to explicitly specify a
  Gateway for every Route, ideally without mutating Routes. (In other words,
  give Ana an easy way to create a defaulted Route.)

- Give Ana an easy way to define the scope for a defaulted Route.

- Give Ana an easy way to determine which default Gateways are present in the
  cluster, if any, and which of her Routes are currently bound to these
  Gateways.

- Continue supporting multiple Gateways in a cluster, while allowing zero or
  more of them to be configured as default Gateways.

- Allow [Chihiro] to retain control over which Gateways accept defaulted
  Routes, so that they can ensure that all Gateways meet their requirements
  for security, performance, and other operational concerns.

- Allow Chihiro to choose not to provide any default Gateways at all.

- Allow Chihiro to define the scope of a default Gateway.

- Allow Chihiro to rename, reconfigure, or replace any default Gateway at
  runtime.

  - While Kubernetes does not allow renaming a resource, Chihiro MUST be able
    to duplicate a default Gateway under a new name, then remove the old
    default Gateway, without disrupting routing. Ana MUST NOT need to go
    update all her Routes just because Chihiro is being indecisive about
    naming.

  - Determine how (or if) to signal changes in functionality if a default
    Gateway's implementation is changed. For example, suppose that Chihiro
    switches a default Gateway from an implementation that supports the
    `HTTPRoutePhaseOfTheMoon` filter to an implementation that does not.

    (Note that this problem is not unique to default Gateways; it affects
    explicitly-named Gateways as well.)

- Allow Chihiro to control which Routes may bind to a default Gateway, and to
  enumerate which Routes are currently bound to a default Gateway.

- Support easy interoperation with common CI/CD and GitOps workflows.

- Define how (or if) listener and Gateway merging applies to a default
  Gateway.

## Non-Goals

- Allow Ana to override Chihiro's choices for default Gateways for a given
  Route without explicitly specifying the Gateway: a Route can either be
  defaulted, or it MUST specify a Gateway explicitly.

- Require that every possible routing use case be met by a Route using a
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

In practice, the trick is to find a usable balance between explicitness and
simplicity, while managing ambiguity. A good example is the humble URL, where
the port number is not always explicit, but it _is_ always unambiguous.
Requiring everyone to type `:80` or `:443` at the end of the host portion of
every URL wouldn't actually help anyone, though allowing it to be specified
explicitly when needed definitely does help people.

### Prior Art

- **Ingress**

   The Ingress resource is the most obvious example of prior art: it permitted
   specifying a default IngressClass, allowing users to create Ingress
   resources that didn't specify the IngressClass explicitly. As with a great
   many things in the Ingress API, this caused problems:

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

- **Emissary Mapping**

  Emissary-ingress turns this idea on its head: it assumes that app developers
  will almost never care about which specific Emissary they're using, and will
  instead only care about the hostnames and ports involved.

  In Emissary:

  - a Listener resource defines which ports and protocols are in play;
  - a Host resource defines hostnames, TLS certificates, etc.;
  - a Mapping resource is roughly analogous to a Route.

  The Listener resource has selectors to control which Hosts it will claim;
  Mappings, though, are claimed by Hosts based on the hostname that the
  Mapping specifies. In other words, Mappings are not bound to a Listener
  explicitly, but rather are bound to a Listener implicitly based on the
  hostname that the Mapping specifies. There is no way to _explicitly_ specify
  which Listener a Mapping wants to be claimed by.

  This is obviously a very different model from Gateway API, shifting almost
  all the work of controlling route binding away from the application
  developer onto the cluster operator.

- **Service**

   We could also consider a Service of `type: LoadBalancer` as a kind of prior
   art: in many cases, Ana can directly create these Services and use them to
   provide direct, completely unmediated access to a workload, without
   worrying about the specifics of how her cluster provider implements them.

   Service's major disadvantages here are that it doesn't support Layer 7
   functionality, and that each Service of type `LoadBalancer` has direct
   costs in many cases. In other words, Service allows Ana to rely on the
   cluster provider to create the load balancer, while forcing Ana to shoulder
   the burden of basically everything else.

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

[Ana]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ana
[Chihiro]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#chihiro
[Ian]: https://gateway-api.sigs.k8s.io/concepts/roles-and-personas/#ian

## API

The main challenge in the API design is to find a way to allow Ana to use
Routes without requiring her to specify the Gateway explicitly, while still
allowing Chihiro and Ian to retain control over Gateways and their
configurations.

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
`status` stanza to carry this information. Ultimately, mutating the `spec` of
a Kubernetes resource introduces complexity which we should avoid if it's not
required. Since we can gracefully provide default-Gateway functionality
without mutating `parentRefs`, we will rely on `status` instead of mutating
`parentRefs`.

[discussion]: https://github.com/kubernetes-sigs/gateway-api/pull/3852#discussion_r2140117567

### Gateway for Ingress (North/South)

There are three main aspects to the API design for default Gateways:

1. Giving Ana a way to indicate that a Route should be defaulted.

2. Giving Chihiro a way to control which Gateways (if any) will accept
   defaulted Routes.

3. Give anyone with read access to Routes (Ana, Chihiro, or Ian) a way to
   enumerate which Routes are bound to the default Gateways.

We will describe each of these aspects in turn, laying out changes to Gateway
API behaviors and resources that are necessary to support them. **Any behavior
not explicitly discussed in this GEP is intended to remain unchanged;** the
GEP covers **all** intended changes to Gateway API behavior.

#### 1. Creating a Defaulted Route

Since Ana must be able to choose whether a Route is defaulted or not, marking
a Route as defaulted must be an active configuration step she takes, rather
than any kind of implicit behavior. To that end, the `CommonRouteSpec`
resource will gain a new field, `useDefaultGateway`, which defines the
_scope_ for the defaulted Route:

```go
// GatewayDefaultScope defines the set of default scopes that a Gateway
// can claim. At present the only supported scope is "All".
type GatewayDefaultScope string

const (
  // GatewayDefaultScopeAll indicates that a Gateway can claim absolutely
  // any Route asking for a default Gateway.
  GatewayDefaultScopeAll GatewayDefaultScope = "All"
)

type CommonRouteSpec struct {
    // ... other fields ...
    useDefaultGateway GatewayDefaultScope `json:"useDefaultGateway,omitempty"`
}
```

For Ana to indicate that a Route should use a default Gateway, she MUST set
the Route's `spec.useDefaultGateway` to the desired scope:

```yaml
...
spec:
  useDefaultGateway: All
```

A defaulted Route MUST be accepted only by Gateways that have been configured
with a matching `spec.useDefaultGateway` scope.

A Route MAY include explicit `parentRefs` in addition to setting
`spec.useDefaultGateway`. In this case, the Route will be a candidate for
being bound to default Gateways, but it will also be bound to its
explicitly-specified `parentRefs`. This allows Ana to create a single Route
that handles N/S traffic via the default Gateways and also handles E/W traffic
via a Service, for example.

All other characteristics of a defaulted Route MUST behave the same as if all
default Gateways were explicitly specified in `parentRefs`.

##### Examples

**Simple N/S Route**: The following HTTPRoute would route _all_ HTTP traffic
arriving at any default Gateway to `my-service` on port 80:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: my-route
spec:
  useDefaultGateway: All
  rules:
  - backendRefs:
    - name: my-service
      port: 80
```

**N/S and E/W Route**: The following HTTPRoute would be bound to both any
default Gateways and to a Service named `face` in the `faces` namespace,
permitting a single Route to handle both N/S traffic (via the default Gateway)
and E/W traffic (via the Service):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: ns-ew-route
spec:
  useDefaultGateway: All
  parentRefs:
  - kind: Service
    name: face
    namespace: faces
  rules:
  - backendRefs:
    - name: face
      port: 80
```

**Multiple Gateways**: A defaulted Route MAY both set `useDefaultGateway` and
name other Gateways in `parentRefs`, although this is not expected to be
common in practice:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: multi-gateway-route
spec:
  useDefaultGateway: All
  parentRefs:
  - kind: Gateway
    name: my-gateway
    namespace: default
  rules:
  - backendRefs:
    - name: my-service
      port: 80
```

##### `status` for a Defaulted Route

When a defaulted Route is claimed by a default Gateway, the Gateway MUST use
`status.parents` to announce that it has claimed the Route, for example:

```yaml
status:
  parents:
  - name: my-default-gateway
    namespace: default
    controllerName: gateway.networking.k8s.io/some-gateway-controller
    conditions:
    - type: Accepted
      status: "True"
      lastTransitionTime: "2025-10-01T12:00:00Z"
      message: "Route is bound to default Gateway"
```

##### Other Considerations

A default Gateway MUST NOT modify the `parentRefs` of a defaulted Route to
indicate that the Route has been claimed by a default Gateway. This becomes
important if the set of default Gateways changes, or (in some situations) if
GitOps tools are in play.

If there are no default Gateways in the cluster, `spec.useDefaultGateway` MUST
be treated as if it were set to `false` in all Routes, parallel to the
situation where a Route specifies a Gateway by name, but no Gateway of that
name exists in the cluster.

#### 2. Controlling which Gateways accept Defaulted Routes

Since Chihiro must be able to control which Gateways accept defaulted Routes,
configuring a Gateway to accept defaulted Routes must be an active
configuration step taken by Chihiro, rather than any kind of implicit
behavior. To that end, the Gateway resource will gain a new field,
`spec.defaultScope`:

```go
type GatewaySpec struct {
    // ... other fields ...
    DefaultScope GatewayDefaultScope `json:"defaultScope,omitempty"`
}
```

Again, the only currently-defined scope is `All`.

If `spec.defaultScope` is set, the Gateway MUST claim Routes that have set
`spec.useDefaultGateway` to a matching value (subject to the usual Gateway API
rules about which Routes may be bound to a Gateway), and it MUST update its
own `status` with a `condition` of type `DefaultGateway` and `status` true to
indicate that it is a default Gateway and what its scope is, for example:

```yaml
status:
  conditions:
  - type: DefaultGateway
    status: "True"
    lastTransitionTime: "2025-10-01T12:00:00Z"
    message: "Gateway has default scope All"
```

If `spec.defaultScope` is not present, the Gateway MUST NOT claim Routes that
do not name it specifically in `parentRefs`, and it MUST NOT set the
`DefaultGateway` condition in its `status`.

##### Access to a Default Gateway

The rules for which Routes may bind to a Gateway do not change for a default
Gateway. In particular, if a default Gateway should accept Routes from other
namespaces, then it MUST include the appropriate `AllowedRoutes` definition,
and without such an `AllowedRoutes`, a default Gateway MUST accept only Routes
from its own namespace.

##### Behavior with No Default Gateway

If no Gateway has `spec.defaultScope` set, then all Gateways MUST ignore
`spec.useDefaultGateway` in all Routes. A Route will be bound to only those Gateways that it specifically names in `parentRefs` entries.

##### Deleting a Default Gateway

Deleting a default Gateway MUST behave the same as deleting any other Gateway:
all Routes that were bound to that Gateway MUST be unbound, and the `Accepted`
conditions in the `status` of those Routes SHOULD be removed.

##### Multiple Default Gateways

Support for multiple default Gateways in a cluster was not one of the original
goals of this GEP. However, allowing Chihiro full control over which Gateways
accept defaulted Routes - including being able to change the set of default
Gateways at runtime, without requiring downtime - has always been a goal, and
this turns out to require support for multiple default Gateways.

Kubernetes itself will not prevent setting `spec.defaultScope` on multiple
Gateways in a cluster, and it also doesn't support any atomic swap mechanisms.
If we want to enforce only a single default Gateway, the Gateway controllers
will have to implement that enforcement logic. There are three possible
options here.

1. Don't bother with any enforcement logic.

    In this case, a Route that sets `spec.useDefaultGateway` would be bound to
    _all_ Gateways that have `spec.defaultScope` set a matching scope. Since
    Gateway API already allows a Route to be bound to multiple Gateways, and
    the Route `status` is already designed for it, this should function
    without difficulty.

2. Treat multiple Gateways with `spec.defaultScope` set as if no Gateway has
   `spec.defaultScope` set.

    If we assume that all Gateway controllers in a cluster can see all the
    Gateways in the cluster, then detecting that multiple Gateways have
    `spec.defaultScope` set is relatively straightforward.

    In this case, every Gateway with `spec.defaultScope` set would ignore it,
    with the final effect being the same as if no Gateway had
    `spec.defaultScope` set: all Gateways would ignore
    `spec.useDefaultGateway` in all Routes, and each Gateway would only accept
    Routes that explicitly named it in `parentRefs`.

    Each Gateway with `spec.defaultScope` set would also update its `status`
    with a `condition` of type `DefaultGateway` and `status` false to indicate
    that it is not the default Gateway, for example:

    ```yaml
    status:
      conditions:
      - type: DefaultGateway
        status: "False"
        lastTransitionTime: "2025-10-01T12:00:00Z"
        message: "Multiple Gateways are marked as default"
    ```

3. Perform conflict resolution as with Routes.

    In this case, the oldest Gateway with `spec.defaultScope` set would be
    considered the only default Gateway. That oldest Gateway would be the only
    one that honors `spec.useDefaultGateway` in Routes, and all other Gateways
    with `spec.defaultScope` set would ignore `spec.useDefaultGateway` in
    every Route.

    The oldest default Gateway would update its `status` to reflect that it
    the default Gateway; all other Gateways with `spec.defaultScope` set to
    `true` will update their `status` as in Option 2.

Unfortunately, option 2 will almost certainly cause downtime in any case where
Chihiro wants to change the implementation behind a default Gateway:

- If Chihiro deletes the old Gateway resource before creating the new one,
  then all routes using that Gateway will be unbound during the time between
  deletion and recreation, resulting in errors for any requests using those
  Routes.

- If Chihiro creates the new Gateway resource before deleting the old one,
  then all Routes using the old default Gateway will still be unbound during
  the time that both Gateways exist.

Option 3 gives Chihiro a way to change the default Gateway without downtime:
when they create the new default Gateway resource, it will not take effect
until the old default Gateway resource is deleted. However, it doesn't give
Chihiro any way to test the Routes through the new default Gateway before
deleting the old Gateway.

Reluctantly, we must therefore conclude that option 1 is the only viable
choice. Therefore: Gateways MUST NOT attempt to enforce a single default
Gateway, and MUST allow Routes that set `spec.useDefaultGateway` to bind to
_all_ Gateways that have `spec.defaultScope` set a matching scope. This is
simplest to implement, it permits zero-downtime changes to the default
Gateway, it allows for testing of the new default Gateway before the old one
is deleted, and it doesn't cause trouble with respect to security posture
(since Ana already accepts that she's giving up some control over how her
Routes are handled when she's using default Gateways).

##### Changes in Functionality

If Chihiro changes a default Gateway to a different implementation that does
not support all the functionality of the previous default Gateway, then the
Routes that were bound to the previous default Gateway will no longer function
as expected. This is not a new problem: it already exists when Ana changes a
Route's `parentRefs`, or when Chihiro changes the implementation of a Gateway
that is explicitly specified in a Route's `parentRefs`.

At present, we do not propose any solution to this problem, other than to note
that `gwctl` or similar tools SHOULD be able to show Ana not just the Gateways
to which a Route is bound, but also the features supported by those Gateways.
This will at least give Ana some visibility into whether she's trying to use
Gateways that don't support a feature that she needs. This is a definitely an
area for future work, and it is complicated by the fact that Ana may not have
access to read Gateway resources in the cluster at all.

##### Listeners, ListenerSets, and Merging

Setting `spec.defaultScope` on a Gateway affects which Routes will bind to the
Gateway, not where the Gateway listens for traffic. As such, setting
`spec.defaultScope` MUST NOT alter a Gateway's behavior with respect to
Listeners, ListenerSets, or merging.

In the future, we may want to consider allowing default ListenerSets rather
than only default Gateways, but that is not in scope for this GEP. Even if it
is considered later, the guiding principle SHOULD be that `spec.defaultScope`
SHOULD NOT affect where a Gateway listens for traffic or whether it can be
merged with other Gateways.

#### 4. Enumerating Routes Bound to Default Gateways

To enumerate Routes bound to the default Gateways, any of Ana, Chihiro, or Ian
can look for Routes that set `spec.useDefaultGateway` to `true`, and then
check the `status.parents` of those Routes to see if the Route has been
claimed. Since this will also show _which_ Gateways have claimed a given
defaulted Route, it neatly solves the problem of allowing Ana to determine
which default Gateway(s) her Route is using even if she doesn't have RBAC to
query Gateway resources directly.

While this is possible with `kubectl get -o yaml`, it's not exactly a friendly
user experience, so adding this functionality to a tool like `gwctl` would be
a dramatic improvement. In fact, looking at the `status` of a Route is very
much something that we should expect any user of Gateway API to do often,
whether or not default Gateways are in play; `gwctl` or something similar
SHOULD be able to show her which Routes are bound to which Gateways in every
case, not just with default Gateways.

### Gateway For Mesh (East/West)

Mesh traffic is defined by using a Service as a `parentRef` rather than a
Gateway. As such, there is no case where a default Gateway would be used for
mesh traffic.

As noted above, a Route MAY both set `spec.useDefaultGateway` _and_ include a
`Service` `parentRef` entry, allowing a single Route to handle both N/S and
E/W traffic. In this case, the Route will be bound to both the default Gateway
and the mesh, and the `status` will show both parents.

## Conformance Details

#### Feature Names

The default-gateway feature will be named `HTTPRouteDefaultGateway` and
`GRPCRouteDefaultGateway`. It is unlikely that an implementation would support
one of these Route types without the other, but `GatewayDefaultGateway` does
not seem like a good choice.

### Conformance tests

TBD.

## Alternatives

- A possible alternative API design is to modify the behavior of Listeners or
  ListenerSets; rather than having a "default Gateway", perhaps we would have
  "[default Listeners]". One challenge here is that the Route `status` doesn't
  currently expose information about which Listener is being used, though it
  does show which Gateway is being used.

[default Listeners]: https://github.com/kubernetes-sigs/gateway-api/pull/3852#discussion_r2149056246

- We could define the default Gateway as a Gateway with a magic name, e.g.
  "default". This doesn't actually make things that much simpler for Ana
  (she'd still have to specify `parentRefs`), and it raises questions about
  Chihiro's ability to control which Routes can bind to the default Gateway,
  as well as how namespacing would work -- it's especially unhelpful for Ana
  if she has to know the namespace of the default Gateway in order to use it.

  (Also, this is a breaking change if Chihiro has already created a
  non-default Gateway with whatever name we choose to use for the convention.)

- A default Gateway could overwrite a defaulted Route's `parentRefs` to point
  to the default Gateway. The main challenge with this approach is that once
  the `parentRefs` are overwritten, it's no longer possible to know what Ana
  originally intended. Using the `status` to indicate that the Route is bound
  to the default Gateway instead both preserves Ana's original intent and also
  makes it possible to change the default Gateway without requiring Ana to
  recreate all her Routes.

## References

TBD.
