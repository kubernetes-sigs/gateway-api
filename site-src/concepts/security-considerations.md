# Gateway API security considerations

Gateway controllers can be deployed in a multi-tenant environment, where different
namespaces are used by different users and customers.

Some caution should be taken by the cluster administrators and Gateway owners to
provide a safer environment.

## Avoiding hostname/domain hijacking

Gateway controllers work to disambiguate and detect conflicts caused by sharing
different ports, protocols, etc. between various listeners. (?)

Generally this conflict resolution works on a first-come, first-served basis, where
the first created resource wins in the  conflict management.

The [hostname definition](../reference/spec.md#httproutespec) is a list, so given the
following scenario:

* `Gateway` accepts routes from a set of namespaces
* `HTTPRoute` with name `route1` is created on namespace `ns1`, with `creationTimestamp: 00:00:01`
and defines hostname `something.tld`.
* `HTTPRoute` with name `route2` is created on namespace `ns2`, with `creationTimestamp: 00:00:30`
and defines hostname `otherthing.tld`.

If the owner of `route1` adds later `otherthing.tld` to the list of hostnames, the 
route will be hijacked from `route2` because `route1` is older.

To avoid this situation, the following actions should be taken:

* On Gateways, admins SHOULD ensure that hostnames are clearly delegated to a specific namespace or set of namespaces:

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gateway1
spec:
  listeners:
  - hostname: "something.tld"
    port: 80
    protocol: HTTP
    allowedRoutes:
      namespaces:
        from: Selector
        selector:
            kubernetes.io/metadata.name: ns1
```

### More than 64 listeners

Gateway resource has a limitation of 64 listener entries. If you need more than 64
listeners, you should consider allowing your users to set their hostnames directly
on the routes of each namespaces but limiting what namespace can use what hostname
by relying on a mechanism like `ValidatingAdmissionPolicy`.

In case you opt to use (the still experimental) `ListenerSet`, a similar mechanism
should also be considered to limit what hostnames a `ListenerSet` can claim.

### Example of a ValidatingAdmissionPolicy

A [ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
can be used to add rules that limits what namespaces can use what domains. 

!!! warning
    The validation policy shown here **IS AN EXAMPLE** and the cluster-admin should do
    adjustments to their own environment! Do not copy/paste this example with proper
    adjustments

The policy exemplified here will:

* Read the allowed domains from a comma-separated value of the `annotation` "domains" present on the namespace.
* Validate if all of the hostnames within `.spec.hostnames` are contained on this annotation.
* In case any of the entries are not authorized, the policy denies its admission.

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: httproute-hostname-policy
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups: ["gateway.networking.k8s.io"]
      apiVersions: ["v1", "v1beta1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["httproutes"]
  variables:
  - name: allowed_hostnames_str
    expression: |
      has(namespaceObject.metadata.annotations) && 
      has(namespaceObject.metadata.annotations.domains) ?
      namespaceObject.metadata.annotations['domains'] : '' 
  - name: allowed_hostnames_list
    expression: |
      variables.allowed_hostnames_str.split(',').
      map(h, h.trim()).filter(h, size(h) > 0)
  validations:
  - expression: |
      !has(object.spec.hostnames) || 
      size(object.spec.hostnames) == 0 || 
      object.spec.hostnames.all(hostname, hostname in variables.allowed_hostnames_list)
    message: "HTTPRoute validation failed. It contains unauthorized hostnames"
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: httproute-hostname-binding
spec:
  policyName: httproute-hostname-policy
  validationActions: ["Deny"]
  matchResources:
    namespaceSelector: {}
```

Once the policy is created, the cluster-admin should explicitly allow the usage of
domains with a command like `kubectl annotate ns default domains=www.dom1.tld,www.dom2.tld`

Additionally, when dealing with environments that provide DNS record creations,
admins should be aware and limit the DNS creation based on the same constraints above.

## Limiting Cross-Namespace References

Owners of resources should be aware of the usage of [ReferenceGrants](../api-types/referencegrant.md).

ReferenceGrant allows resource owners to make their resources available to
Gateway API resources in other namespaces. It may be beneficial to restrict where
this can be done.

A `ValidatingAdmissionPolicy` can be used to limit what kind of `resource` and which `namespace`
can create a `ReferenceGrant`.

Below is an **EXAMPLE** that can be used to limit the usage of a `ReferenceGrant` just 
on namespaces labeled with `referencegrants=allow` and only allowing objects of kind `HTTPRoute`
to reference an object of kind `Service` that MUST have a name:

```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: reference-grant-limit
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
    - apiGroups: ["gateway.networking.k8s.io"]
      apiVersions: ["v1beta1"]
      operations: ["CREATE", "UPDATE"]
      resources: ["referencegrants"]
  variables:
  - name: allowed_grant_ns
    expression: |
      has(namespaceObject.metadata.labels) && 
      has(namespaceObject.metadata.labels.referencegrants) &&
      namespaceObject.metadata.labels['referencegrants'] == 'allow' 
  - name: allowed_from_kind
    expression: |
      object.spec.from.all(f, f.kind=='HTTPRoute')
  - name: allowed_to_kind
    expression: |
      object.spec.to.all(t, t.kind == 'Service' && has(t.name) && t.name != '')
  validations:
  - expression: |
      variables.allowed_grant_ns && variables.allowed_from_kind && variables.allowed_to_kind
    message: "ReferenceGrant must be explicitly allowed on the namespace, from an HTTPRoute to a named service"
---
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: reference-grant-limit
spec:
  policyName: reference-grant-limit
  validationActions: ["Deny"]
  matchResources:
    namespaceSelector: {}
```

ReferenceGrant owners should ensure that the reference permissions being granted are as minimal as possible.

* Specify `to` targets in all possible ways (`group`, `kind`, **and** `name`)  
  * In particular, DO NOT leave `name` unspecified, even though it is optional, without a *very* good reason, as that is granting a blanket 


## Proper definition of Roles and RoleBinding

The creation of a new Gateway should be considered as a privileged permission.
The unguarded creation of Gateways may increase costs, infrastructure modifications
(like a LoadBalancer and a DNS record creation) and as in such case, admins should
be aware of it and create [Roles](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#role-and-clusterrole) 
and [RoleBindings](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding) 
that reflect proper user permissions.

Additionally, it is highly recommended that the strict permissions are given, not
allowing regular users to modify a Gateway API status.

For more information about the security model of Gateway API, check [Security Model](security-model.md)

## Usage and limit of GatewayClass

A cluster may have different GatewayClasses, with different purposes. As an example,
one GatewayClass may enforce that Gateways attached to it can only use internal load balancers.

Cluster admins should be aware of these requirements, and define validation
policies that limit the improper attachment of a Gateway to a GatewayClass by unauthorized users.

A [ValidatingAdmissionPolicy](https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/)
can be used to limit what namespaces can use a `GatewayClass`.
