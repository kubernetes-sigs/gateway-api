# Policy Attachment

While features like timeouts, retries, and custom health checks are present in
most implementations, their details vary since there are no standards (RFCs)
around them. This makes these features less portable. So instead of pulling
these into the API, we offer a middle ground: a standard way to plug these
features in the API and offer a uniform UX across implementations. This standard
approach for policy attachment allows implementations to create their own custom
policy resources that can essentially extend Gateway API.

Policies attached to Gateway API resources and implementations must use the
following approach to ensure consistency across implementations of the API.
There are three primary components of this pattern:

* A standardized means of attaching policy to resources.
* Support for configuring both default and override values within policy
  resources.
* A hierarchy to illustrate how default and override values should interact.

This kind of standardization not only enables consistent patterns, it allows
future tooling such as kubectl plugins to be able to visualize all policies that
have been applied to a given resource.

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
The following structure MUST be used as for any Policy resource using this API
pattern. Within the spec, policy resources may omit `Override` or `Default`
fields, but at least one of them MUST be present.

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
Each policy MAY include default or override values. Overrides enable admins to
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

If using defaults and overrides, each policy resource MUST include 2 structs
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

![Hierarchical Policy Example](images/policy-hierarchy.png)

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
can choose to support a refernce to a virtual resource type:

```yaml
apiVersion: networking.example.net/v1alpha1
kind: RetryPolicy
metadata:
  name: foo
spec:
  default:
    maxRetries: 5
  targetRef:
    group: networking.example.net
    kind: ExternalService
    name: foo.com
```

### Conflict Resolution
It is possible for multiple policies to target the same resource. When this
happens, merging is the preferred outcome. If multiple policy resources target
the same resource _and_ have an identical field specified with different values,
precedence MUST be determined in order of the following criteria, continuing on
ties:

* The oldest Policy based on creation timestamp. For example, a Policy with a
  creation timestamp of "2021-07-15 01:02:03" is given precedence over a Policy
  with a creation timestamp of "2021-07-15 01:02:04".
* The Policy appearing first in alphabetical order by "{namespace}/{name}". For
  example, foo/bar is given precedence over foo/baz.

### Kubectl Plugin
To help improve UX and standardization, a kubectl plugin will be developed that
will be capable of describing the computed sum of policy that applies to a given
resource, including policies applied to parent resources.

Each Policy CRD that wants to be supported by this plugin will need to follow
the API structure defined above and add a
`gateway.networking.k8s.io/policy-attachment: ""` label to the CRD.

### Status
In the future, we may consider adding a new `Policies` field to status on
Gateways and Routes. This would be a list of `PolicyTargetReference` structs
with the fields instead used to refer to the Policy resource that has been
applied.

Unfortunately, this may create more confusion than it is worth, here are some of
the key concerns:

* When multiple controllers are implementing the same Route and recognize a
  policy, it would be difficult to determine which controller should be
  responsible for adding that policy reference to status.
* For this to be somewhat scalable, we'd need to limit the status entries to
  policies that had been directly applied to the resource. This could get
  confusing as it would not provide any insight into policies attached above or
  below.
* Since we only control some of the resources a policy might be attached to,
  adding policies to status would only be possible on Gateway API resources, not
  Services or other kinds of backends.

Although these concerns are not unsolvable, they lead to the conclusion that
a Kubectl plugin should be our primary approach to providing visibility here,
with a possibility of adding policies to status at a later point.

### Interaction with Custom Route Filters
Both Policy attachment and custom Route filters provide ways to extend Gateway
API. Although similar in nature, they have slightly different purposes.

Custom Route filters provide a way to configure request/response modifiers or
middleware embedded inside Route rules or backend references.

Policy attachment is more broad in scope. In contrast with filters, policies can
be attached to a wide variety of Gateway API resources, and include a concept of
hierarchical defaulting and overrides. Although Policy attachment can be used to
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

#### 2. Custom filters and policies should not overlap
Although it's possible that arbitrary fields could be supported by custom
policy, custom route filters, and core/extended fields concurrently, it is
strongly recommended that implementations not use multiple mechanisms for
representing the same fields. A given field should only be supported through a
single extension method. An example of potential conflict is policy precedence
and structured hierarchy, which only applies to custom policies. Allowing a
field to exist in custom policies and also other areas of the API, which are not
part of the structured hierarchy, breaks the precedence model. Note that this
guidance may change in the future as we gain a better understanding of how
extension mechanisms of the Gateway API can interoperate.

### Conformance Level
This policy attachment pattern is associated with an "EXTENDED" conformance
level. The implementations that support this policy attachment model will have
the same behavior and semantics, although they may not be able to support
attachment of all types of policy at all potential attachment points. When that
is the case, implementations MUST clearly document which resources a policy may
be attached to.
