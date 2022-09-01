# Deploying a simple Gateway


The simplest possible deployment is a Gateway and Route resource which are
deployed together by the same owner. This represents a similar kind of model
used for Ingress. In this guide, a Gateway and HTTPRoute are deployed which
match all HTTP traffic and directs it to a single Service named `foo-svc`. 

![Simple Gateway](/images/single-service-gateway.png)

```yaml  
{% include 'standard/simple-gateway/gateway.yaml' %}
```

The Gateway represents the instantation of a logical load balancer. It's
templated from a hypothetical `acme-lb` GatewayClass. The Gateway listens for
HTTP traffic on port 80. This particular GatewayClass automatically assigns an
IP address which will be shown in the `Gateway.status` after it has been
deployed. 

Route resources specify the Gateways they want to attach to using `ParentRefs`. As long as 
the Gateway allows this attachment (by default Routes from the same namespace are trusted),
this will allow the Route to receive traffic from the parent Gateway. 
`BackendRefs` define the backends that traffic will be sent to. More complex 
bi-directional matching and permissions are possible and explained in other guides.

The following HTTPRoute defines how traffic from the Gateway listener is routed
to backends. Because there are no host routes or paths specified, this HTTPRoute
will match all HTTP traffic that arrives at port 80 of the load balancer and
send it to the `foo-svc` Pods. 

```yaml  
{% include 'standard/simple-gateway/httproute.yaml' %}
```

While Route resources are often used to filter traffic to many different
backends (potentially with different owners), this demonstrates the simplest
possible route with a single Service backend. This example shows how a service
owner can deploy both the Gateway and the HTTPRoute for their usage alone,
giving them more control and autonomy for how the service is exposed.
