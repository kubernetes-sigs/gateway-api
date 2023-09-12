# Security Model

## Introduction
Gateway API has been designed to enable granular authorization for each role in
a typical organization.

## Resources
Gateway API has 3 primary API resources:

* **GatewayClass** defines a set of gateways with a common configuration and
  behavior.
* **Gateway** requests a point where traffic can be translated to Services
  within the cluster.
* **Routes** describe how traffic coming via the Gateway maps to the Services.

## Roles and personas

There are 3 primary roles in Gateway API, as described in [roles and personas]:

- **Ian** (he/him): Infrastructure Provider
- **Chihiro** (they/them): Cluster Operator
- **Ana** (she/her): Application Developer

[roles and personas]:/concepts/roles-and-personas

### RBAC

RBAC (role-based access control) is the standard used for Kubernetes
authorization. This allows users to configure who can perform actions on
resources in specific scopes. RBAC can be used to enable each of the roles
defined above. In most cases, it will be desirable to have all resources be
readable by most roles, so instead we'll focus on write access for this model.

#### Write Permissions for Simple 3 Tier Model
| | GatewayClass | Gateway | Route |
|-|-|-|-|
| Infrastructure Provider | Yes | Yes | Yes |
| Cluster Operators | No | Yes | Yes |
| Application Developers | No | No | Yes |

#### Write Permissions for Advanced 4 Tier Model
| | GatewayClass | Gateway | Route |
|-|-|-|-|
| Infrastructure Provider | Yes | Yes | Yes |
| Cluster Operators | Sometimes | Yes | Yes |
| Application Admins | No | In Specified Namespaces | In Specified Namespaces |
| Application Developers | No | No | In Specified Namespaces |

## Crossing Namespace Boundaries
Gateway API provides new ways to cross namespace boundaries. These
cross-namespace capabilities are quite powerful but need to be used carefully to
avoid accidental exposure. As a rule, every time we allow a namespace boundary
to be crossed, we require a handshake between namespaces. There are 2 different
ways that can occur:

### 1. Route Binding
Routes can be connected to Gateways in different namespaces. To accomplish this,
The Gateway owner must explicitly allow Routes to bind from additional
namespaces. This is accomplished by configuring allowedRoutes within a Gateway
listener to look something like this:

```yaml
namespaces:
  from: Selector
  selector:
    matchExpressions:
    - key: kubernetes.io/metadata.name
      operator: In
      values:
      - foo
      - bar
```

This will allow routes from the "foo" and "bar" namespaces to attach to this
Gateway listener.

#### Risks of Other Labels
Although it's possible to use other labels with this selector, it is not quite
as safe. While the `kubernetes.io/metadata.name` label is consistently set on
namespaces to the name of the namespace, other labels do not have the same
guarantee. If you used a custom label such as `env`, anyone that is able to
label namespaces within your cluster would effectively be able to change the set
of namespaces your Gateway supported.

### 2. ReferenceGrant
There are some cases where we allow other object references to cross namespace
boundaries. This includes Gateways referencing Secrets and Routes referencing
Backends (usually Services). In these cases, the required handshake is
accomplished with a ReferenceGrant resource. This resource exists within a
target namespace and can be used to allow references from other namespaces.

For example, the following ReferenceGrant allows references from HTTPRoutes in
the "prod" namespace to Services that are deployed in the same namespace as
the ReferenceGrant.

```yaml
{% include 'standard/reference-grant.yaml' %}
```

For more information on ReferenceGrant, refer to our [detailed documentation
for this resource](/api-types/referencegrant).

## Advanced Concept: Limiting Namespaces Where a GatewayClass Can Be Used
Some infrastructure providers or cluster operators may wish to limit the
namespaces where a GatewayClass can be used. At this point, we do not have a
solution for this built into the API. In lieu of that, we recommend using a
policy agent such as Open Policy Agent and
[Gatekeeper](https://github.com/open-policy-agent/gatekeeper) to enforce these
kinds of policies. For reference, we've created an [example of
configuration](https://github.com/open-policy-agent/gatekeeper-library/pull/24)
that could be used for this.
