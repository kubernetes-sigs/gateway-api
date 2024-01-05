# Roles and Personas

## Background

In the original design of Kubernetes, Ingress and Service resources were based
on a usage model in which the developers who create Services and Ingresses
controlled all aspects of defining and exposing their applications to their
users.

In practice, though, clusters and their infrastructure tend to be shared,
which the original Ingress model doesn't capture very well. A critical factor
is that when infrastructure is shared, not everyone using the infrastructure
has the same concerns, and to be successful, an infrastructure project needs
to address the needs of all the users.

This raises a fundamental challenge: how do you provide the flexibility needed
by the users of the infrastructure, while also maintaining control by the
owners of the infrastructure?

Gateway API defines several distinct roles, each with an associated
_persona_, as a tool for surfacing and discussing the differing needs of
different users in order to balance usability, flexibility, and control.
Design work within Gateway API is deliberately cast in terms of these
personas.

Note that, depending on the environment, a single human may end up taking on
multiple roles, as discussed below.

## Key Roles and Personas

Gateway API defines three roles and personas:

* **Ian**<a name="ian"></a> (he/him) is an _infrastructure provider_,
  responsible for the care and feeding of a set of infrastructure that permits
  multiple isolated clusters to serve multiple tenants. He is not beholden to
  any single tenant; rather, he worries about all of them collectively. Ian
  will often work for a cloud provider (AWS, Azure, GCP, ...) or for a PaaS
  provider.

* **Chihiro**<a name="Chihiro"></a> (they/them) is a _cluster operator_,
  responsible for managing clusters to ensure that they meet the needs of
  their several users. Chihiro will typically be concerned with policies,
  network access, application permissions, etc. Again, they are beholden to no
  single user of any cluster; rather, they need to make sure that the clusters
  serve all users as needed.

* **Ana**<a name="ana"></a> (she/her) is an _application developer_,
  responsible for creating and managing an application running in a cluster.
  From Gateway API's point of view, Ana will need to manage configuration
  (e.g. timeouts, request matching/filter) and Service composition (e.g. path
  routing to backends). She is in a unique position among Gateway API
  personas, since her focus is on the business needs her application is meant
  to serve, _not_ Kubernetes or Gateway API. In fact, Ana is likely to
  view Gateway API and Kubernetes as pure friction getting in her way to
  get things done.

Depending on the environment, multiple roles can map to the same user:

- Giving a single user all the above roles replicates the self-service model,
  and may actually happen in a small startup running Kubernetes on bare metal.

- A more typical small startup would use clusters from a cloud provider. In
  this situation, Ana and Chihiro may be embodied in the same person, with Ian
  being an employee (or automated process!) within the cloud provider.

- In a much larger organization, we would expect each persona above to be
  embodied by a distinct person (most likely working in different groups,
  perhaps with little direct contact).

## RBAC

RBAC (role-based access control) is the standard used for Kubernetes
authorization. This allows users to configure who can perform actions on
resources in specific scopes. We anticipate that each persona will map
approximately to a `Role` in the Kubernetes Role-Based Authentication (RBAC)
system and will define resource model responsibility and separation.

RBAC is discussed further in the [Security Model] description.

[Security Model]: /concepts/security-model#rbac
