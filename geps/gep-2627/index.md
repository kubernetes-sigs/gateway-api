# GEP-2627: DNS configuration within Gateway API

* Issue: [#2627](https://github.com/kubernetes-sigs/gateway-api/issues/2627)
* Status: Provisional

## TLDR

For gateway infrastructure to be valuable we need to be able to connect clients to these gateways. A common way to achieve this is to use domain names/hostnames and DNS. Gateways define listeners that can have assigned hostnames or wildcards.  The guidelines for DNS configuration are a critical piece of service networking, but this is currently not expressible as part of Gateway API.   Instead of leaving this as an exercise for the user to figure out, this proposal attempts to provide options to ease Gateway API operations.

## Goals
* Provide a way for Gateway API resource owners to mark their resources as relevant for external DNS provisioning
* Ensure that the above method has a way for multiple providers to be present in the cluster and be able to actuate external DNS provisioning requests
* Ensure that any method is based on structured fields and makes the most of `status` on whatever resources are relevant, whether they are existing Gateway API resources or new resources.
* Increase portability and supportability between Gateway API implementations and third party controllers offering DNS integration.
* Clarity on the scope of hostnames under management. (This should _not_ be able to be used to affect the standard in-cluster DNS configuration)

## Non-Goals

* Anything to do with configuring in-cluster DNS. This support is for configuration outside the cluster only.
* Ways to define if the Gateway API resources are allowed to request particular hostnames. These choices should be left to the implementations that actually actuate the requests for hostnames. However, `status` flows should be specified so as to make clear if a hostname provisioning request cannot be performed.
* Cover more complex DNS routing strategies that come into play for multi-cluster topologies such as round robin, failover, health checks, weighted and geo location with this first pass. Supporting these types of use cases for distributed gateways (e.g., in different regions or multiple gateways for resilience within a region) and offering a form of global load balancing leveraging DNS is a potential future goal.

## Use Cases

As a cluster administrator, I manage a set of domains and a set of gateways. I would like to declaratively define which gateways should be used for provisioning DNS records, and, if necessary, which DNS provider to use to configure connectivity for clients accessing these domains and my gateway so that I can see and configure which DNS provider is being used.

As a cluster administrator, I would like to have the DNS names automatically populated into my specified DNS zones as a set of records based on the assigned addresses of my gateways so that I do not have to undertake external automation or management of this essential task.

As a cluster administrator I would have the status of the DNS records reported back to me, so that I can leverage existing kube based monitoring tools to know the status of the integration.

As a cluster administrator, I would like the DNS records to be updated automatically if the `spec` of assigned gateways changes, whether those changes are for IP address or hostname. 
As a DNS administrator, I should be able to ensure that only approved External DNS controllers can make changes to DNS zone configuration. (This should in general be taken care of by DNS system <-> External DNS controller interactions like user credentials and operation status responses, but it is important to remember that it needs to happen).
## API

Initial draft will not offer an API yet until the use cases are agreed. Some thoughts worth thinking about: 
- I think it is important that we try to move away from APIs based on annotations which, while convenient, are not a full API and suffer from several limitations. An example: I want to configure a listener with a domain I own that is in a different provider than the domains of the other listeners. I want to add a new option to configure a particular weighting and so on. Soon you end up with a large set of connected annotations that often grow in complexity that really should be expressed as an API.

- It is also important that this API can be delegated to controllers other than the Gateway API provider/implementor. This is because there are existing solutions that may want to support whatever API decided upon. It should not **have** to be a gateway provider that has to integrate with many DNS providers. 

## Conformance Details

TBD

## Alternatives

it is possible to use `external-dns` to manage dns based on HTTPRoutes and Gateways https://github.com/kubernetes-sigs/external-dns/blob/7f3c10d65297ec1c4bcc8dd6f88c189b7f3e80d0/docs/tutorials/gateway-api.md. The aim of this GEP is not remove this as an option, but instead provide a common API that could then be leveraged by something like external-dns. 

## References

TBD