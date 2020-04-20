# Design Decisions

Throughout the design of these APIs some significant design decisions have been
made. This provides some background on the factors considered when making these
decisions.

## GatewayClass Scope

### Problem
Moving to a namespace-scoped GatewayClass would provide more flexibility for
RBAC configuration and simplify some use cases. This would also contradict the
existing pattern for class resources to be cluster-scoped.

### Advantages of a Namespace-Scoped GatewayClass
* Namespaces could be used to separate different GatewayClasses, enabling admins
  to more closely restrict who can modify individual GatewayClasses.
* A full Gateway stack, including a GatewayClass, can be created by users with
  only namespace level access, similar to the initial implementation of Ingress.
* Enables a more restrictive (and potentially more secure) default for where a
  GatewayClass could be used. This could mean only allowing Gateways and Routes
  in the same namespace by default.
* Provides better model for using Service APIs on a [management cluster](mc)
  designed to provision and manage other clusters and infrastructure needed by
  other clusters.

[mc]: https://cluster-api.sigs.k8s.io/reference/glossary.html#management-cluster

### Advantages of a Cluster-Scoped GatewayClass
* Consistent with existing class resources, including StorageClass and
  IngressClass, that are all cluster-scoped.
* Discussions with sig-storage showed that StorageClass had worked well as a
  cluster-scoped resource and that they were continuing to follow that pattern
  for new Class resources.
* Follows general pattern that resources managed by infrastructure provider
  and/or cluster administrators are cluster scoped. This matches the proposed
  [security model](sm) for GatewayClass.
* Impossible for namespace level admins to grant RBAC permissions for
  GatewayClass resources. With a namespace-scoped resource, anyone that could
  configure RBAC within a namespace could grant GatewayClass access.
* Although both a namespace-scoped or cluster-scoped GatewayClass could be
  referenced from multiple namespaces, it's significantly more common for a
  cluster-scoped resource to be accessible from multiple namespaces.
* Works well when there are only a few total GatewayClasses per cluster, likely
  the majority of use cases.
* Simple globally unique names with no need to specify a namespace when
  referencing a GatewayClass.

[sm]: https://github.com/kubernetes-sigs/service-apis/blob/master/docs-src/security-model.md#rbac

### Decision
The advantages of the cluster-scoped GatewayClass ended up outweighing the
advantages of a namespace-scoped resource. This was primarily driven by the
desire to maintain consistency with previous class resources. There was a high
bar to move away from the established pattern here and we determined that these
reasons did not quite reach that.

For more context on this decision, refer to the [corresponding pull request](pr).

[pr]: https://github.com/kubernetes-sigs/service-apis/pull/156
