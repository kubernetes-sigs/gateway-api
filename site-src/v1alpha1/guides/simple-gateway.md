# Deploying a simple Gateway

!!! warning "v1alpha1 has been deprecated"

    Please upgrade to v1alpha2, v1alpha1 will be removed from Gateway API
    in an upcoming release.

The simplest possible deployment is a Gateway and Route resource which are
deployed together by the same owner. This represents a similar kind of model
used for Ingress. In this guide, a Gateway and HTTPRoute are deployed which
match all HTTP traffic and directs it to a single Service named `foo-svc`. 

![Simple Gateway](/v1alpha1/images/single-service-gateway.png)

```yaml  
{% include 'v1alpha1/simple-gateway/gateway.yaml' %} 
```

The Gateway represents the instantation of a logical load balancer. It's
templated from a hypothetical `acme-lb` GatewayClass. The Gateway listens for
HTTP traffic on port 80. This particular GatewayClass automatically assigns an
IP address which will be shown in the `Gateway.status` after it has been
deployed. 

Gateways bind Routes to themselves via label selection (similar to how Services
label select across Pod labels). In this example, the `prod-web` Gateway will
bind any HTTPRoute resources which have the `gateway: prod-web-gw` label. The
label can be any arbitrary label, but using one that identifies the name or
capabilities of the Gateway is useful to Route owners and makes the relationship
more explicit. More complex bi-directional matching and permissions are possible
and explained in other guides.

The following HTTPRoute defines how traffic from the Gateway listener is routed
to backends. Because there are no host routes or paths specified, this HTTPRoute
will match all HTTP traffic that arrives at port 80 of the load balancer and
send it to the `foo-svc` Pods. 

```yaml  
{% include 'v1alpha1/simple-gateway/httproute.yaml' %} 
```

While Route resources are often used to filter traffic to many different
backends (potentially with different owners), this demonstrates the simplest
possible route with a single Service backend. This example shows how a service
owner can deploy both the Gateway and the HTTPRoute for their usage alone,
giving them more control and autonomy for how the service is exposed.
