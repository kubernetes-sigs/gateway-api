# Security Model

!!! danger
    This page is very of date and needs to be updated.
## Introduction
The Gateway API have been designed to enable granular authorization for each
role in a typical organization.

## Resources
The Gateway API have 3 primary API resources:

* **GatewayClass** defines a set of gateways with a common configuration and
  behavior.
* **Gateway** requests a point where traffic can be translated to Services
  within the cluster.
* **Routes** describe how traffic coming via the Gateway maps to the Services.

### Additional Configuration
There are two additional pieces of configuration that are important in this
security model:

* Which namespaces can contain Gateways of the specified GatewayClass.
* Which namespaces Routes can be targeted in by Gateways of the specified
  GatewayClass.

## Roles and personas

In the original design of Kubernetes, Ingress and Service resources were
based on a self-service model of usage; developers who create Services and
Ingresses control all aspects of defining and exposing their applications to
their users.

We have found that the self-service model does not fully capture some of the
more complex deployment and team structures that our users are seeing. The
Gateway API are designed to target the following personas:

* **Infrastructure provider**: The infrastructure provider (infra) is
  responsible for the overall environment that the cluster(s) are operating in.
  Examples include: the cloud provider (AWS, Azure, GCP, ...), the PaaS provider
  in a company.
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

Depending on the environment, multiple roles can map to the same user.
For example, giving the user all the above roles replicates the self-service
model.

## The Security Model
There are two primary components to the Gateway API security model: RBAC and
namespace restrictions.

## RBAC
RBAC (role-based access control) is the standard used for Kubernetes
authorization. This allows users to configure who can perform actions on
resources in specific scopes. RBAC can be used to enable each of the roles
defined above. In most cases, it will be desirable to have all resources be
readable by most roles, so instead we'll focus on write access for this model.

### Write Permissions for Simple 3 Tier Model
| | GatewayClass | Gateway | Route |
|-|-|-|-|
| Infrastructure Provider | Yes | Yes | Yes |
| Cluster Operators | No | Yes | Yes |
| Application Developers | No | No | Yes |

### Write Permissions for Advanced 4 Tier Model
| | GatewayClass | Gateway | Route |
|-|-|-|-|
| Infrastructure Provider | Yes | Yes | Yes |
| Cluster Operators | Sometimes | Yes | Yes |
| Application Admins | No | In Specified Namespaces | In Specified Namespaces |
| Application Developers | No | No | In Specified Namespaces |

## Limiting Namespaces Where a GatewayClass Can Be Used
Some infrastructure providers or cluster operators may wish to limit the
namespaces where a GatewayClass can be used. At this point, we do not have a
solution for this built into the API. We continue to [explore
options](https://github.com/kubernetes-sigs/gateway-api/issues/375) to improve
support for this. Until then, we recommend using a policy agent such as Open
Policy Agent and [Gatekeeper](https://github.com/open-policy-agent/gatekeeper)
to enforce these kinds of policies. For reference, we've created an [example of
configuration](https://github.com/open-policy-agent/gatekeeper-library/pull/24)
that could be used for this.

## Route Namespaces
Gateway API allow Gateways to select Routes across multiple Namespaces.
Although this can be remarkably powerful, this capability needs to be used
carefully. Gateways include a `RouteNamespaces` field that allows selecting
multiple namespaces with a label selector. By default, this is limited to Routes
in the same namespace as the Gateway. Additionally, Routes include a `Gateways`
field that allows them to restrict which Gateways use them. If the Gateways
field is not specified (i.e. its empty), then the Route will default to allowing
selection by Gateways in the same namespace.

## Controller Requirements
To be considered conformant with the Gateway API spec, controllers need to:

* Populate status fields on Gateways and Resources to indicate if they are
  compatible with the corresponding GatewayClass configuration.
* Ensure that all Routes added to a Gateway:
    * Have been selected by the Gateway.
    * Have a Gateways field that allows the Gateway use of the route.

## Alternative Approaches Considered
### New API Resources
We considered introducing new API resources to cover these use cases. These
resources might be look something like:

* **ClusterGateway**: A ClusterGateway could reference routes in any namespace.
* **ClusterRoute**: A ClusterRoute could be referenced by any Gateway or
  ClusterGateway.

**Benefits**

* Easy to model with RBAC.
* API validation tied directly to each resource.

**Downsides**

* New resources to deal with - more informers, clients, documentation, etc.
* Harder to expand with additional options in the future - may just end up with
  tons of API resources to cover all use cases.

### Boolean Multi Namespace Route Indicator on GatewayClass
Instead of having the `routeNamespaceSelector` field on GatewayClass, we would
use a boolean `multiNamespaceRoutes` field to indicate if Gateways of this class
can target routes in multiple namespaces. This would default to false. A false
value here would indicate that routes could only be targeted in the current
namespace.

**Benefits**

* Helpful for multi-tenant use cases with many isolated Gateways.
* Simple configuration with an easy to understand default value.

**Downsides**

* GatewayClass admins are unable to partially limit namespaces that can be
  targeted by Gateways. Admins would have to choose between allowing access to
  Routes in all namespaces or only the local one.

### Validating Webhook
A validating webhook could potentially handle some of the cross-resource
validation necessary for this security model and provide more immediate feedback
to end users.

**Benefits**

* Immediate validation feedback.
* More validation logic stays in core Gateway API codebase.

**Downsides**

* Imperfect solution for cross-resource validation. For example, a change to a
  GatewayClass could affect the validity of corresponding Gateway.
* Additional complexity involved in installing Gateway API in a cluster.
