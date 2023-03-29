# Metaresources and Policy Attachment

The Gateway API defines a Kubernetes object that _augments_ the behavior of an object
in a standard way as a _Metaresource_. ReferenceGrant
is an example of this general type of metaresource, but it is far from the only
one.

This document also defines a concept called _Policy Attachment_, which augments
the behavior of an object to add additional settings that can't be described
within the spec of that object.

Why have this class of attachment? Well, while features like timeouts, retries,
and custom health checks are present in most implementations, their details vary
since there are no standards (RFCs) around them. This makes these features less
portable. So instead of pulling these into the API, we offer a middle ground:
a standard way to plug these features in the API and offer a uniform UX across
implementations. This standard approach for policy attachment allows
implementations to create their own custom policy resources that can essentially
extend Gateway API, and have those settings flow across multiple resources (like
attaching a Policy to a Gateway and having the settings affect all HTTPRoutes
attached to that Gateway, for example).

This document defines how we control the creation of configuration in the underlying
Gateway data plane using two types of Policy Attachment.

A "Policy Attachment" is a specific type of _metaresource_ that can affect specific
settings across either one object (this is "Direct Policy Attachment"), or objects
in a hierarchy (this is "Inherited Policy Attachment").

In either case, a Policy may either affect an object by controlling the value
of one of the existing _fields_ in the `spec` of an object, or it may add
additional fields that are _not_ in the `spec` of the object.

### Direct Policy Attachment

A Direct Policy Attachment is tightly bound to one instance of a particular
Kind within a single namespace (or to an instance of a single Kind at cluster scope),
and only modifies the behavior of the object that matches its binding.

As an example, one use case that Gateway API currently does not support is how
to configure details of the TLS required to connect to a backend (in other words,
if the process running inside the backend workload expects TLS, not that some
automated infrastructure layer is provisioning TLS as in the Mesh case).

A hypothetical TLSConnectionPolicy that targets a Service could be used for this,
using the functionality of the Service as describing a set of endpoints. (It
should also be noted this is not the only way to solve this problem, just an
example to illustrate Direct Policy Attachment.)

The TLSConnectionPolicy would look something like this:

```yaml
apiVersion: gateway.networking.k8s.io/v1alpha2
kind: TLSConnectionPolicy
metadata:
  name: tlsport8443
  namespace: foo
spec:
  targetRef: # This struct is defined as part of Gateway API
    group: "" # Empty string means core - this is a standard convention
    kind: Service
    name: fooService
  tls:
    certificateAuthorityRefs:
      - name: CAcert
    port: 8443

```

All this does is tell an implementation, that for connecting to port `8443` on the
Service `fooService`, it should assume that the connection is TLS, and expect the
service's certificate to be validated by the chain in the `CAcert` Secret.

Importantly, this would apply to _every_ usage of that Service across any HTTPRoutes
in that namespace, which could be useful for a Service that is reused in a lot of
HTTPRoutes.

With these two examples in mind, here are some guidelines for when to consider
using Direct Policy Attachment:

* The number or scope of objects to be modified is limited or singular. Direct
  Policy Attachments must target one specific object.
* The modifications to be made to the objects don’t have any transitive information -
  that is, the modifications only affect the single object that the targeted
  metaresource is bound to, and don’t have ramifications that flow beyond that
  object.
* In terms of status, it should be reasonably easy for a user to understand that
  everything is working - basically, as long as the targeted object exists, and
  the modifications are valid, the metaresource is valid, and this should be
  straightforward to communicate in one or two Conditions. Note that at the time
  of writing, this is *not* completed.
* Direct Policy Attachment _should_ only be used to target objects in the same
  namespace as the Policy object. Allowing cross-namespace references brings in
  significant security concerns, and/or difficulties about merging cross-namespace
  policy objects. Notably, Mesh use cases may need to do something like this for
  consumer policies, but in general, Policy objects that modify the behavior of
  things outside their own namespace should be avoided unless it uses a handshake
  of some sort, where the things outside the namespace can opt–out of the behavior.
  (Notably, this is the design that we used for ReferenceGrant).

### Inherited Policy Attachment: It's all about the defaults and overrides

Because an Inherited Policy is a metaresource, it targets some other resource
and _augments_ its behavior.

But why have this distinct from other types of metaresource? Because Inherited
Policy resources are designed to have a way for settings to flow down a hierarchy.

Defaults set the default value for something, and can be overridden by the
“lower” objects (like a connection timeout default policy on a Gateway being
overridable inside a HTTPRoute), and Overrides cannot be overridden by “lower”
objects (like setting a maximum client timeout to some non-infinite value at the
Gateway level to stop HTTPRoute owners from leaking connections over time).

Here are some guidelines for when to consider using an Inherited Policy object:

* The settings or configuration are bound to one containing object, but affect
  other objects attached to that one (for example, affecting HTTPRoutes attached
  to a single Gateway, or all HTTPRoutes in a GatewayClass).
* The settings need to able to be defaulted, but can be overridden on a per-object
  basis.
* The settings must be enforced by one persona, and not modifiable or removable
  by a lesser-privileged persona. (The owner of a GatewayClass may want to restrict
  something about all Gateways in a GatewayClass, regardless of who owns the Gateway,
  or a Gateway owner may want to enforce some setting across all attached HTTPRoutes).
* In terms of status, a good accounting for how to record that the Policy is
  attached is easy, but recording what resources the Policy is being applied to
  is not, and needs to be carefully designed to avoid fanout apiserver load.
  (This is not built at all in the current design either).

When multiple Inherited Policies are used, they can interact in various ways,
which are governed by the following rules, which will be expanded on later in
in this document.

* If a Policy does not affect an object's fields directly, then the resultant
  Policy should be the set of all distinct fields inside the relevant Policy objects,
  as set out by the rules below.
* For Policies that affect an object's existing fields, multiple instances of the
  same Policy Kind affecting an object's fields will be evaluated as
  though only a single Policy "wins" the right to affect each field. This operation
  is performed on a _per-distinct-field_ basis.
* Settings in `overrides` stanzas will win over the same setting in a `defaults`
  stanza.
* `overrides` settings operate in a "less specific beats more specific" fashion -
  Policies attached _higher_ up the hierarchy will beat the same type of Policy
  attached further down the hierarchy.
* `defaults` settings operate in a "more specific beats less specific" fashion -
  Policies attached _lower down_ the hierarchy will beat the same type of Policy
  attached further _up_ the hierarchy.
* For `defaults`, the _most specific_ value is the one _inside the object_ that
  the Policy applies to; that is, if a Policy specifies a `default`, and an object
  specifies a value, the _object's_ value will win.
* Policies interact with the fields they are controlling in a "replace value"
  fashion.
  * For fields where the `value` is a scalar, (like a string or a number)
    should have their value _replaced_ by the value in the Policy if it wins.
    Notably, this means that a `default` will only ever replace an empty or unset
    value in an object.
  * For fields where the value is an object, the Policy should include the fields
    in the object in its definition, so that the replacement can be on simple fields
    rather than complex ones.
  * For fields where the final value is non-scalar, but is not an _object_ with
    fields of its own, the value should be entirely replaced, _not_ merged. This
    means that lists of strings or lists of ints specified in a Policy will overwrite
    the empty list (in the case of a `default`) or any specified list (in the case
    of an `override`). The same applies to `map[string]string` fields. An example
    here would be a field that stores a map of annotations - specifying a Policy
    that overrides annotations will mean that a final object specifying those
    annotations will have its value _entirely replaced_ by an `override` setting.
* In the case that two Policies of the same type specify different fields, then
  _all_ of the specified fields should take effect on the affected object.

Examples to further illustrate these rules are given below.

## Policy Attachment for Ingress
Attaching policy to Gateway resources for ingress use cases is relatively
straightforward. A policy can reference the resource it wants to apply to.
Access is granted with RBAC - for example, anyone that has access to create a
RetryPolicy in a given namespace can attach it to any resource within that
namespace.

![Simple Ingress Example](/images/policy/ingress-simple.png)

To build on that example, it’s possible to attach policies to more resources.
Each policy applies to the referenced resource and everything below it in terms
of hierarchy. Although this example is likely more complex than many real world
use cases, it helps demonstrate how policy attachment can work across
namespaces.

![Complex Ingress Example](/images/policy/ingress-complex.png)

## Policy Attachment for Mesh
Although there is a great deal of overlap between ingress and mesh use cases,
mesh enables more complex policy attachment scenarios. For example, users may
want to apply policy to requests from a specific namespace to a backend in
another namespace.

![Simple Mesh Example](/images/policy/mesh-simple.png)

Policy attachment can be quite simple with mesh. Policy can be applied to any
resource in any namespace but it can only apply to requests from the same
namespace if the target is in a different namespace.

At the other extreme, policy can be used to apply to requests from a specific
workload to a backend in another namespace. A route can be used to intercept
these requests and split them between different backends (foo-a and foo-b in
this case).

![Complex Mesh Example](/images/policy/mesh-complex.png)

## Target Reference API

Each Policy resource MUST include a single `targetRef` field. It MUST not
target more than one resource at a time, but it can be used to target larger
resources such as Gateways or Namespaces that may apply to multiple child
resources.

The `targetRef` field MUST be an exact replica of the `PolicyTargetReference`
struct included in the Gateway API. Where possible, it is recommended to use
that struct directly instead of duplicating the type.

### Policy Boilerplate
The following (or something like it) SHOULD be used as for any Policy resource using this API
pattern. Within the spec, policy resources that omit both `Override` and `Default`
fields are defined as Direct Policy Attachment, and Inherited Policy Attachment must include
one or both.

```go
// ACMEServicePolicy provides a way to apply Service policy configuration with
// the ACME implementation of the Gateway API.
type ACMEServicePolicy struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`

    // Spec defines the desired state of ACMEServicePolicy.
    Spec ACMEServicePolicySpec `json:"spec"`

    // Status defines the current state of ACMEServicePolicy.
    Status ACMEServicePolicyStatus `json:"status,omitempty"`
}

// ACMEServicePolicySpec defines the desired state of ACMEServicePolicy.
type ACMEServicePolicySpec struct {
    // TargetRef identifies an API object to apply policy to.
    TargetRef gatewayv1a2.PolicyTargetReference `json:"targetRef"`

    // Override defines policy configuration that should override policy
    // configuration attached below the targeted resource in the hierarchy.
    // +optional
    Override *ACMEPolicyConfig `json:"override,omitempty"`

    // Default defines default policy configuration for the targeted resource.
    // +optional
    Default *ACMEPolicyConfig `json:"default,omitempty"`
}

// ACMEPolicyConfig contains ACME policy configuration.
type ACMEPolicyConfig struct {
    // Add configurable policy here
}

// ACMEServicePolicyStatus defines the observed state of ACMEServicePolicy.
type ACMEServicePolicyStatus struct {
    // Conditions describe the current conditions of the ACMEServicePolicy.
    //
    // +optional
    // +listType=map
    // +listMapKey=type
    // +kubebuilder:validation:MaxItems=8
    Conditions []metav1.Condition `json:"conditions,omitempty"`
}
```

### Hierarchy
Each Inherited policy MUST include default and/or override values. Overrides enable admins to
enforce policy from the top down. Defaults enable app owners to provide default
values from the bottom up for each individual application.

![Policy Hierarchy](/images/policy/hierarchy.png)

To illustrate this, consider 3 resources with the following hierarchy: A
(highest) > B > C. When attaching the concept of defaults and overrides to that,
the hierarchy would be expanded to this:

A override > B override > C override > C default > B default > A default.

Note that the hierarchy is reversed for defaults. The rationale here is that
overrides usually need to be enforced top down while defaults should apply to
the lowest resource first. For example, if an admin needs to attach required
policy, they can attach it as an override to a Gateway. That would have
precedence over Routes and Services below it. On the other hand, an app owner
may want to set a default timeout for their Service. That would have precedence
over defaults attached at higher levels such as Route or Gateway.

If using defaults _and_ overrides, each policy resource MUST include 2 structs
within the spec. One with override values and the other with default values.

In the following example, the policy attached to the Gateway requires cdn to
be enabled and provides some default configuration for that. The policy attached
to the Route changes the value for one of those fields (`includeQueryString`).

```yaml
kind: AcmeServicePolicy # Example of implementation specific policy name
spec:
  override:
    cdn:
      enabled: true
  default:
    cdn:
      cachePolicy:
        includeHost: true
        includeProtocol: true
        includeQueryString: true
  targetRef:
    kind: Gateway
    name: example
---
kind: AcmeServicePolicy
spec:
  default:
    cdn:
      cachePolicy:
        includeQueryString: false
  targetRef:
    kind: HTTPRoute
    name: example
```

In this final example, we can see how the override attached to the Gateway has
precedence over the default `drainTimeout` value attached to the Route. At the
same time, we can see that the default `connectionTimeout` attached to the Route
has precedence over the default attached to the Gateway.

Also note how the different resources interact - fields that are not common across
objects _may_ both end up affecting the final object.

![Inherited Policy Example](images/policy-hierarchy.png)

#### Supported Resources
It is important to note that not every implementation will be able to support
policy attachment to each resource described in the hierarchy above. When that
is the case, implementations MUST clearly document which resources a policy may
be attached to.

#### Attaching Policy to GatewayClass
GatewayClass may be the trickiest resource to attach policy to. Policy
attachment relies on the policy being defined within the same scope as the
target. This ensures that only users with write access to a policy resource in a
given scope will be able to modify policy at that level. Since GatewayClass is a
cluster scoped resource, this means that any policy attached to it must also be
cluster scoped.

GatewayClass parameters provide an alternative to policy attachment that may be
easier for some implementations to support. These parameters can similarly be
used to set defaults and requirements for an entire GatewayClass.

### Targeting External Services
In some cases (likely limited to mesh or egress) users may want to apply
policies to requests to external services. To accomplish this, implementations
can choose to support a reference to a virtual resource type:

```yaml
apiVersion: networking.example.net/v1alpha1
kind: RetryPolicy
metadata:
  name: foo
spec:
  maxRetries: 5
  targetRef:
    group: networking.example.net
    kind: ExternalService
    name: foo.com
```

Because this CRD does _not_ have a `defaults` or `overrides` section, it is
a Direct Attached Policy.

### Merging into existing `spec` fields

It's possible (even likely) that configuration in a Policy may need to be merged
into an existing object's fields somehow, particularly for Inherited policies.

When merging into an existing fields inside an object, Policy objects should
merge values at a scalar level, not at a struct or object level.

For example, in the `CDNCachingPolicy` example above, the `cdn` struct contains
a `cachePolicy` struct that contains fields. If an implementation was merging
this configuration into an existing object that contained the same fields, it
should merge the fields at a scalar level, with the `includeHost`,
`includeProtocol`, and `includeQueryString` values being defaulted if they were
not specified in the object being controlled. Similarly, for `overrides`, the
values of the innermost scalar fields should overwrite the scalar fields in the
affected object.

Implementations should not copy any structs from the Policy object directly into the
affected object, any fields that _are_ overridden should be overridden on a per-field
basis.

In the case that the field in the Policy affects a struct that is a member of a list,
each existing item in the list in the affected object should have each of its
fields compared to the corresponding fields in the Policy.

For non-scalar field _values_, like a list of strings, or a `map[string]string`
value, the _entire value_ must be overwritten by the value from the Policy. No
merging should take place. This mainly applies to `overrides`, since for
`defaults`, there should be no value present in a field on the final object.

This table shows how this works for various types:

|Type|Object config|Override Policy config|Result|
|----|-------------|----------------------|------|
|string| `key: "foo"` | `key: "bar"`  | `key: "bar"` |
|list| `key: ["a","b"]` | `key: ["c","d"]` | `key: ["c","d"]` |
|`map[string]string`| `key: {"foo": "a", "bar": "b"}` | `key: {"foo": "c", "bar": "d"}` | `key: {"foo": "c", "bar": "d"}` |


### Conflict Resolution
It is possible for multiple policies to target the same object _and_ the same
fields inside that object. If multiple policy resources target
the same resource _and_ have an identical field specified with different values,
precedence MUST be determined in order of the following criteria, continuing on
ties:

* Direct Policies should never overlap Inherited Policies. If preventing settings from
  being overwritten is important, implementations should only use Inherited
  Policies, and the `override` stanza that implies.
* Inside Inherited Policies, the same setting in `overrides` beats the one in
  `defaults`.
* The oldest Policy based on creation timestamp. For example, a Policy with a
  creation timestamp of "2021-07-15 01:02:03" is given precedence over a Policy
  with a creation timestamp of "2021-07-15 01:02:04".
* The Policy appearing first in alphabetical order by `{namespace}/{name}`. For
  example, foo/bar is given precedence over foo/baz.

For a better user experience, a validating webhook can be implemented to prevent
these kinds of conflicts all together.

### Status

In the current iteration of this design, metaresources and Policy objects don't
have any standard way to record what they're attaching to, or applying settings
to in the case of Policy Attachment. Previous experience in the Kubernetes API
has made it clear that having a single object that can cause status updates to
occur across many other objects can have a big performance impact, so the status
design must be very carefully done to avoid these kind of fanout problems.

However, the whole purpose of having a standardized Policy API structure and
patterns is intended to make this problem solvable both for human users and with
tooling.

This is currently a _very_ open question. A discussion is ongoing at
[#1531](https://github.com/kubernetes-sigs/gateway-api/discussions/1531), and this
GEP will be updated with any outcomes.

Some key concerns that we need to solve for status:

* When multiple controllers are implementing the same Route and recognize a
  policy, it must be possible to determine which controller was
  responsible for adding that policy reference to status.
* For this to be somewhat scalable, we must limit the number of status updates
  that can result from a metaresource update.
* Since we only control some of the resources a policy might be attached to,
  adding policies to status would only be possible on Gateway API resources, not
  Services or other kinds of backends.

### Interaction with Custom Route Filters
Both Policy attachment and custom Route filters provide ways to extend Gateway
API. Although similar in nature, they have slightly different purposes.

Custom Route filters provide a way to configure request/response modifiers or
middleware embedded inside Route rules or backend references.

Policy attachment is more broad in scope. In contrast with filters, policies can
be attached to a wide variety of Gateway API resources, and include a concept of
inherited defaulting and overrides. Although Policy attachment can be used to
target an entire Route or Backend, it cannot currently be used to target
specific Route rules or backend references. If there are sufficient use cases
for this, policy attachment may be expanded in the future to support this fine
grained targeting.

The following guidance should be considered when introducing a custom field into
any Gateway controller implementation:

#### 1. Use core or extended fields if available
For any given field that a Gateway controller implementation needs, the
possibility of using core or extended fields should always be considered
before using custom policy resources. This is encouraged to promote
standardization and, over time, to absorb capabilities into the API as first
class fields, which offer a more streamlined UX than custom policy
attachment.

#### 2. Custom filters and policies should only overlap if necessary
Although it's possible that arbitrary fields could be supported by custom
policy, custom route filters, and core/extended fields concurrently, it is
recommended that implementations only use multiple mechanisms for
representing the same fields when those fields really _need_ the defaulting
and/or overriding behavior that Policy Attachment provides. For example, a
custom filter that allowed the configuration of Authentication inside a
HTTPRoute object might also have an associated Policy resource that allowed
the filter's settings to be defaulted or overridden. It should be noted that
doing this in the absence of a solution to the status problem is likely to
be *very* difficult to troubleshoot.

### Conformance Level
This policy attachment pattern is associated with an "EXTENDED" conformance
level. The implementations that support this policy attachment model will have
the same behavior and semantics, although they may not be able to support
attachment of all types of policy at all potential attachment points. When that
is the case, implementations MUST clearly document which resources a policy may
be attached to.

## Examples

This section provides some examples of various types of Policy objects, and how
merging, `defaults`, `overrides`, and other interactions work.

### Direct Policy Attachment

The following Policy sets the minimum TLS version required on a Gateway Listener:
```yaml
apiVersion: networking.example.io/v1alpha1
kind: TLSMinimumVersionPolicy
metadata:
  name: minimum12
  namespace: appns
spec:
  minimumTLSVersion: 1.2
  targetRef:
    name: internet
    group: gateway.networking.k8s.io
    kind: Gateway
```

Note that because there is no version controlling the minimum TLS version in the
Gateway `spec`, this is an example of a Policy that affects fields that aren't
represented in the object.

### Inherited Policy Attachment

It also could be useful to be able to _default_ the `minimumTLSVersion` setting
across multiple Gateways.

This version of the above Policy allows this:
```yaml
apiVersion: networking.example.io/v1alpha1
kind: TLSMinimumVersionPolicy
metadata:
  name: minimum12
  namespace: appns
spec:
  defaults:
    minimumTLSVersion: 1.2
  targetRef:
    name: appns
    group: ""
    kind: namespace
```

This Inherited Policy is using the implicit hierarchy that all resources belong
to a namespace, so attaching a Policy to a namespace means affecting all possible
resources in a namespace. Multiple hierarchies are possible, even within Gateway
API, for example Gateway -> Route, Gateway -> Route -> Backend, Gateway -> Route
-> Service. GAMMA Policies could conceivably use a hierarchy of Service -> Route
as well.

Note that this will not be very discoverable for Gateway owners in the absence of
a solution to the Policy status problem.

Conceivably, a security or admin team may want to _force_ Gateways to have at least
a minimum TLS version of `1.2` - that would be a job for `overrides`, like so:

```yaml
apiVersion: networking.example.io/v1alpha1
kind: TLSMinimumVersionPolicy
metadata:
  name: minimum12
  namespace: appns
spec:
  overrides:
    minimumTLSVersion: 1.2
  targetRef:
    name: appns
    group: ""
    kind: namespace
```

This will make it so that _all Gateways_ in the `default` namespace _must_ use
a minimum TLS version of `1.2`, and this _cannot_ be changed by Gateway owners.
Only the Policy owner can change this Policy.

### Handling non-scalar values

In this example, we will assume that at some future point, HTTPRoute has grown
a Filter to configure retries (`RetryFilter`), including a field called `retryOn`
that reflects the HTTP status codes that should be retried. The _value_ of this
field is a list of strings, being the HTTP codes that must be retried. The `retryOn`
field has no defaults in the field definitions (which is probably a bad design,
but we need to show this interaction somehow!)

We also assume that an Inherited `RetryOnPolicy` exists that allows both
defaulting and overriding of the `retryOn` field in the `RetryFilter`.

A full `RetryOnPolicy` to default the field to the codes `501`, `502`, and `503`
would look like this:
```yaml
apiVersion: networking.example.io/v1alpha1
kind: RetryOnPolicy
metadata:
  name: retryon5xx
  namespace: appns
spec:
  defaults:
    retryOn:
      - "501"
      - "502"
      - "503"
  targetRef:
    kind: Gateway
    group: gateway.networking.k8s.io
    name: we-love-retries
```

This means that, for HTTPRoutes that use the `RetryFilter and do _NOT_ explicitly set this field to something
else, (in other words, they contain an empty list), then the field will be set to
a list containing `501`, `502`, and `503`. (Notably, because of Go zero values, this
would also occur if the user explicitly set the value to the empty list.)

However, if a HTTPRoute owner sets any value other than the empty list in the filter, then that
value will remain, and the Policy will have _no effect_. These values are _not_
merged.

If the Policy used `overrides` instead:
```yaml
apiVersion: networking.example.io/v1alpha1
kind: RetryOnPolicy
metadata:
  name: retryon5xx
  namespace: appns
spec:
  overrides:
    retryOn:
      - "501"
      - "503"
  targetRef:
    kind: Gateway
    group: gateway.networking.k8s.io
    name: you-must-retry
```

Then no matter what the value is in the filter, it will be set to `501`, `503`
by the Policy override.

### Interactions between defaults, overrides, and field values

All HTTPRoutes that attach to the `YouMustRetry` Gateway and use a `RetryFilter`
will have any value _overwritten_ by this policy. The empty list, or any number
of values, will all be replaced with `501`, `502`, and `503`.

Now, let's also assume that we use the Namespace -> Gateway hierarchy on top of
the Gateway -> HTTPRoute hierarchy, and allow attaching a `RetryOnPolicy` to a
_namespace_. The expectation here is that this will affect all Gateways in a namespace
and all HTTPRoutes that use the `RetryFilter` and attach to those Gateways.
(Note that the HTTPRoutes themselves may not necessarily be in the same namespace though.)

If we apply the default policy from earlier to the namespace:
```yaml
apiVersion: networking.example.io/v1alpha1
kind: RetryOnPolicy
metadata:
  name: retryon5xx
  namespace: appns
spec:
  defaults:
    retryOn:
      - "501"
      - "502"
      - "503"
  targetRef:
    kind: Namespace
    group: ""
    name: appns
```

Then this will have the same effect as applying that Policy to every Gateway in
the `default` namespace - namely that every HTTPRoute that attaches to every
Gateway will have its `retryOn` field in the `RetryFilter` set to `501`, `502`, `503`,
_if_ there is no other setting in the `RetryFilter` itself.

With two layers in the hierarchy, we have a more complicated set of interactions
possible.

Let's look at some tables for a particular HTTPRoute, assuming that it does _not_
configure the `retryOn` field, but _does_ configure a `RetryFilter`, for various
types of Policy at different levels.

#### Overrides interacting with defaults for RetryOnPolicy, empty list in RetryFilter

||None|Namespace override|Gateway override|HTTPRoute override|
|----|-----|-----|----|----|
|No default|Empty list|Namespace override| Gateway override Policy| HTTPRoute override|
|Namespace default| Namespace default| Namespace override | Gateway override | HTTPRoute override |
|Gateway default| Gateway default | Namespace override | Gateway override | HTTPRoute override |
|HTTPRoute default| HTTPRoute default | Namespace override | Gateway override | HTTPRoute override|

#### Overrides interacting with other overrides for RetryOnPolicy, empty list in RetryFilter
||No override|Namespace override A|Gateway override A|HTTPRoute override A|
|----|-----|-----|----|----|
|No override|Empty list|Namespace override| Gateway override| HTTPRoute override|
|Namespace override B| Namespace override B| Namespace override<br />first created wins<br />otherwise first alphabetically | Namespace override B | Namespace override B|
|Gateway override B| Gateway override B | Namespace override A| Gateway override<br />first created wins<br />otherwise first alphabetically | Gateway override B|
|HTTPRoute override B| HTTPRoute override B | Namespace override A| Gateway override A| HTTPRoute override<br />first created wins<br />otherwise first alphabetically|

#### Defaults interacting with other defaults for RetryOnPolicy, empty list in RetryFilter
||No default|Namespace default A|Gateway default A|HTTPRoute default A|
|----|-----|-----|----|----|
|No default|Empty list|Namespace default| Gateway default| HTTPRoute default A|
|Namespace default B| Namespace default B| Namespace default<br />first created wins<br />otherwise first alphabetically | Gateway default A | HTTPRoute default A|
|Gateway default B| Gateway default B| Gateway default B| Gateway default<br />first created wins<br />otherwise first alphabetically | HTTPRoute default A|
|HTTPRoute default B| HTTPRoute default B| HTTPRoute default B| HTTPRoute default B| HTTPRoute default<br />first created wins<br />otherwise first alphabetically|


Now, if the HTTPRoute _does_ specify a value in its `RetryFilter`,
it's a bit easier, because we can basically disregard all defaults:

#### Overrides interacting with defaults for RetryOnPolicy, value in RetryFilter

||None|Namespace override|Gateway override|HTTPRoute override|
|----|-----|-----|----|----|
|No default| Value in RetryFilter|Namespace override| Gateway override | HTTPRoute override|
|Namespace default|  Value in RetryFilter| Namespace override | Gateway override | HTTPRoute override |
|Gateway default|  Value in RetryFilter | Namespace override | Gateway override | HTTPRoute override |
|HTTPRoute default| Value in RetryFilter | Namespace override | Gateway override | HTTPRoute override|

#### Overrides interacting with other overrides for RetryOnPolicy, value in RetryFilter
||No override|Namespace override A|Gateway override A|HTTPRoute override A|
|----|-----|-----|----|----|
|No override|Value in RetryFilter|Namespace override A| Gateway override A| HTTPRoute override A|
|Namespace override B| Namespace override B| Namespace override<br />first created wins<br />otherwise first alphabetically | Namespace override B| Namespace override B|
|Gateway override B| Gateway override B| Namespace override A| Gateway override<br />first created wins<br />otherwise first alphabetically | Gateway override B|
|HTTPRoute override B| HTTPRoute override B | Namespace override A| Gateway override A| HTTPRoute override<br />first created wins<br />otherwise first alphabetically|

#### Defaults interacting with other defaults for RetryOnPolicy, value in RetryFilter
||No default|Namespace default A|Gateway default A|HTTPRoute default A|
|----|-----|-----|----|----|
|No default|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|
|Namespace default B|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|
|Gateway default B|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|
|HTTPRoute default B|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|Value in RetryFilter|
