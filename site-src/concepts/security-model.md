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

In the original design of Kubernetes, Ingress and Service resources were based
on a self-service model of usage; developers who create Services and Ingresses
control all aspects of defining and exposing their applications to their users.

We have found that the self-service model does not fully capture some of the
more complex deployment and team structures that our users are seeing. Gateway
API is designed to target the following personas:

* **Infrastructure provider**: The infrastructure provider (infra) is
  responsible for the overall environment that the cluster(s) are operating in.
  Examples include: the cloud provider (AWS, Azure, GCP, ...) or the PaaS
  provider in a company.
* **Cluster operator**: The cluster operator (ops) is responsible for
  administration of entire clusters. They manage policies, network access,
  application permissions.
* **Application developer**: The application developer (dev) is responsible for
  defining their application configuration (e.g. timeouts, request
  matching/filter) and Service composition (e.g. path routing to backends).

Although these roles can cover a wide variety of use cases, some organizations
may be structured slightly differently. Many organizations may also have a
fourth role that sits between "cluster operator" and "application developer":

* **Application admin**: The application admin has administrative access to some
  namespaces within a cluster, but not the cluster as a whole.

We expect that each persona will map approximately to a `Role` in the Kubernetes
Role-Based Authentication (RBAC) system and will define resource model
responsibility and separation.

Depending on the environment, multiple roles can map to the same user. For
example, giving the user all the above roles replicates the self-service model.

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
