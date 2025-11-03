# Gateway API security considerations

Gateway controllers can be deployed in a multi-tenant environment, where different
namespaces are used by different users and customers.

Some caution should be taken by the cluster administrators and Gateway owners to
provide a safer environment.

## Avoiding hostname/domain hijacking

Gateway controllers work to disambiguate and detect conflicts caused by sharing
different ports, protocols, etc. between various listeners. (?)

Generally this conflict detection works on a first-come, first-served basis, where
the first created resource wins in the  conflict management.

The [hostname definition](../reference/spec.md#httproutespec) is a list, so given the
following scenario:

* `Gateway` accepts routes from a set of namespaces
* `HTTPRoute` with name `route1` is created on namespace `ns1`, with `creationTimestamp: 00:00:01`
and defines hostname `something.tld`.
* `HTTPRoute` with name `route2` is created on namespace `ns2`, with `creationTimestamp: 00:00:30`
and defines hostname `otherthing.tld`.

The owner of `route1` can hijack the domain `otherthing.tld` from `route2` because
`route1` is an older resource.

To avoid this situation, the following actions should be taken:

* On shared gateways, admin SHOULD specify on the listener definitions different
domains, and specific namespaces allowed to use each domain, as the example below:

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

* Because the number of listeners is limited on a Gateway, the administrator should
instead rely on some validation (like ValidationAdmissionPolicy or some other mechanism)
that limits what hostnames can be used on routes of each namespaces.
* In case of `ListenerSet` (still experimental) a validation policy should also be applied.

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

## Usage of ReferenceGrant

Owners of resources should be aware of the usage of [ReferenceGrants](../api-types/referencegrant.md).
This should be audited and limited by admins (needs some better writing here).

The intended use of ReferenceGrants is to allow for the *owner* of an object where
Gateway API cannot change the spec (such as a Secret or Service) to have a way to
allow cross-namespace use of that object for Gateway API purposes.

It’s intended to require that the use of an object in a Gateway API object requires
both the ***referrer*** (generally the Gateway or Route) and the ***referent***
(a Secret or Service) to *agree* that the reference is allowed. This means that,
when the referrer and referent are owned by different people, then those two people
must also agree that the reference is allowed.

The design of ReferenceGrant is intended to be as secure as possible by default:

* Without a ReferenceGrant, cross-namespace references to a Secret or Gateway
(or any other resource that Gateway API does not control the spec of) MUST fail.  
* The ReferenceGrant MUST be created in the same namespace as the object that
reference permissions are being granted to. This makes it easier to ensure that
the same person owns both the referent object and the ReferenceGrant.  
* Most fields in the ReferenceGrant object are **required**, so that the
ReferenceGrant cannot be overly broad.

Because of this design, ReferenceGrant owners should ensure that the reference
permissions being granted are as minimal as possible.

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
