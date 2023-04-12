# Cross-Namespace routing

The Gateway API has core support for cross Namespace routing. This is useful
when more than one user or team is sharing the underlying networking
infrastructure, yet control and configuration must be segmented to minimize
access and fault domains.

Gateways and Routes can be deployed into different Namespaces and Routes  can
attach to Gateways across Namespace boundaries. This allows user access 
control to be applied differently across Namespaces for Routes and Gateways, 
effectively segmenting access and control to different parts of the
cluster-wide  routing configuration. The ability for Routes to attach to
Gateways across Namespace boundaries are governed by [_Route Attachment_](#cross-namespace-route-attachment). Route attachment is explored
in this guide and demonstrates how independent teams can safely share the same
Gateway.

In this guide there are two independent teams, _store_ and _site_, operating
in the same Kubernetes cluster in the `store-ns` and `site-ns` Namespaces. These
are their goals and how they use Gateway API resources to accomplish them:

- The site team has two applications, _home_ and _login_. The team wants to to
isolate access and  configuration across their apps as much as possible to
minimize access and failure domains. They use separate HTTPRoutes attached to
the same Gateway to isolate routing configurations, such as canary rollouts,
but still share the same IP address, port, DNS domain, and TLS certificate.
- The store team has a single Service called _store_ that they have deployed
in the `store-ns` Namespace which also needs to be exposed behind the same IP
address and domain.
- The Foobar Corporation operates behind the `foo.example.com` domain for all
apps. This is controlled by a central infrastructure team, operating in the
`infra-ns` Namespace.
- Lastly, the security team controls the certificate for `foo.example.com`.
By managing this certificate through the single shared Gateway they are able
to centrally control security without directly involving application teams.

The logical relationship between the Gateway API resources looks like this:

![Cross-Namespace routing](/images/cross-namespace-routing.svg)

## Cross-namespace Route Attachment

[Route attachment][attachment] is an important concept that dictates how Routes
attach to Gateways and program their routing rules. It is especially relevant
when there are Routes across Namespaces that share one or more Gateways.
Gateway and Route attachment is bidirectional - attachment can only succeed if
the Gateway owner and Route owner both agree to the relationship. This
bi-directional relationship exists for two reasons:

- Route owners don't want to overexpose their applications through paths they 
are not aware of.
- Gateway owners don't want certain apps or teams using Gateways without 
permission. For example, an internal service shouldn't be accessible 
through an internet Gateway.

Gateways support _attachment constraints_ which are fields on Gateway
listeners that restrict which Routes can be attached. Gateways support
Namespaces and Route types as attachment constraints. Any Routes that do not
meet the attachment constraints are not able to attach to that Gateway. 
Similarly, Routes explicitly reference Gateways that they want to attach to
through the Route's `parentRef` field. Together these create a handshake
between the infra owners and application owners that enables them to
independently define how applications are exposed through Gateways. This is
effectively a policy that reduces administrative overhead. App owners can
specify which Gateways their apps should use and infra owners can constrain
the Namespaces and types of Routes that a Gateway accepts.


## Shared Gateway

The infrastructure team deploys the `shared-gateway` Gateway into the `infra-ns`
Namespace:

```yaml
{% include 'standard/cross-namespace-routing/gateway.yaml' %}
```

The `https` listener in the above Gateway matches traffic for the
`foo.example.com` domain. This allows the infrastructure team to manage all 
aspects of the domain. The HTTPRoutes below do not need to specify domains
and will match all traffic by default if `hostname` is not set. This makes
it easier to manage HTTPRoutes because they can be domain agnostic, which is
helpful when application domains are not static.

This Gateway also configures HTTPS using the `foo-example-com` Secret
in the `infra-ns` Namespace. This allows the infrastructure team to centrally
manage TLS on behalf of app owners. The `foo-example-com` certificate will
terminate all traffic going to its attached Routes, without any TLS 
configuration on the HTTPRoutes themselves.

This Gateway uses a Namespace selector to define which HTTPRoutes are allowed 
to attach. This allows the infrastructure team to constrain who
or which apps can use this Gateway by allowlisting a set of Namespaces.


```yaml
apiVersion: gateway.networking.k8s.io/v1beta1
kind: Gateway
spec:
  listeners:
  - allowedRoutes:
      namespaces:
        from: Selector
        selector:
          matchLabels:
            shared-gateway-access: "true"
...
```

_Only_ Namespaces which are labelled `shared-gateway-access: "true"` will be
able to attach their Routes to `shared-gateway`. In the following set of
Namespaces, if an HTTPRoute existed in the `no-external-access` Namespace with
a `parentRef`  for `infra-ns/shared-gateway`, it would be ignored by the
Gateway because the  attachment constraint (Namespace label) was not met.

```yaml
{% include 'standard/cross-namespace-routing/0-namespaces.yaml' %}
```

Note that attachment constraints on the Gateway are not required, but they are
a best-practice if operating a cluster with many different teams and
Namespaces. In environments where all apps in a cluster have permission to
attach to a Gateway then the `listeners[].routes` field does not have to be
configured and all Routes can freely use the Gateway.


## Route Attachment 

The store team deploys their route for the `store` Service in the `store-ns`
Namespace:

```yaml
{% include 'standard/cross-namespace-routing/store-route.yaml' %}
```

This Route has straightforward routing logic as it just matches for
`/store` traffic which it sends to the `store` Service.

The site team now deploys Routes for their applications. They deploy two
HTTPRoutes into the `site-ns` Namespace:

- The `home` HTTPRoute acts as a default routing rule, matching for all traffic
to `foo.example.com/*` not matched by an existing routing rule and sending it to
the `home` Service.
- The `login` HTTPRoute  routes traffic for `foo.example.com/login` to
`service/login-v1` and `service/login-v2`. It uses weights to granularly
control traffic distribution between them.

Both of these Routes use the same Gateway attachment configuration which
specifies `gateway/shared-gateway` in the `infra-ns` Namespace as the only
Gateway that these Routes want to attach to.

```yaml
{% include 'standard/cross-namespace-routing/site-route.yaml' %}
```

After these three Routes are deployed, they will all be attached to the
`shared-gateway` Gateway. The Gateway merges these Routes into a single flat
list of routing rules. [Routing precedence](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteRule)
between these routing rules is determined by most specific match and
conflicts are handled according to [conflict
resolution](/concepts/guidelines#conflicts). This provides predictable and
deterministic merging of routing rules between independent users.

Thanks to cross-Namespace routing, the Foobar Corporation can distribute
ownership of their infrastructure more evenly, while still retaining centralized
control. This gives them the best of both worlds, all delivered through
declarative and open source APIs.

[attachment]:/concepts/api-overview/#attaching-routes-to-gateways
