# Migrating from Ingress

The Gateway API project is the successor to the [Ingress API][ing]. However, it
does not include the Ingress resource (the closest parallel is the HTTPRoute).
As a result, a one-time conversion from your existing Ingress resources to the
relevant Gateway API resources is necessary.

[ing]:https://kubernetes.io/docs/concepts/services-networking/ingress/

This guide will help you with the conversion. It will:

* Explain why you may want to switch to the Gateway API.
* Describe the key differences between the Ingress API and the Gateway API.
* Map Ingress features to Gateway API features.
* Show an example of an Ingress resource converted to Gateway API resources.
* Mention [ingress2gateway](https://github.com/kubernetes-sigs/ingress2gateway)
  for automatic conversion.

At the same time, it will not prepare you for a live migration or explain how to
convert some implementation-specific features of your Ingress controller.
Additionally, since the Ingress API only covers HTTP/HTTPS traffic, this guide
does not cover the Gateway API support for other protocols.

## Reasons to Switch to Gateway API

The [Ingress API](https://kubernetes.io/docs/concepts/services-networking/ingress/)
is the standard Kubernetes way to configure external HTTP/HTTPS load balancing
for Services. It is widely adopted by Kubernetes users and well-supported by
vendors with many implementations ([Ingress controllers](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/))
available. Additionally, several cloud-native projects integrate with the
Ingress API, such as [cert-manager](https://cert-manager.io/)
and [ExternalDNS](https://github.com/kubernetes-sigs/external-dns).

However, the Ingress API has several limitations:

- *Limited features*. The Ingress API only supports TLS termination and
  simple content-based request routing of HTTP traffic.
- *Reliance on annotations for extensibility*. The annotations approach to
  extensibility leads to limited portability as every implementation has its own
  supported extensions that may not translate to any other implementation.
- *Insufficient permission model*. The Ingress API is not well-suited for
  multi-team clusters with shared load-balancing infrastructure.

The Gateway API addresses those limitations, as the next section will show.

> Read more about the [design goals](https://gateway-api.sigs.k8s.io/#gateway-api-concepts)
> of the Gateway API.

## Key Differences Between Ingress API and Gateway API

There are three major differences between the Ingress API and the Gateway API:

* Personas
* Available features
* Approach to extensibility (implementation-specific features)

### Personas

At first, the Ingress API had only a single resource kind Ingress. As a result,
it had only one persona -- the user -- the owner of Ingress resources. The
Ingress features give the user a lot of control over how applications are
exposed to their external clients, including TLS termination configuration and
provisioning of the load balancing infrastructure (supported by some Ingress
controllers). Such a level of control is called the self-service model.

At the same time, the Ingress API also included two implicit personas to
describe somebody responsible for provisioning and managing an Ingress
controller: the infrastructure provider for provider-managed Ingress controllers
and the cluster operator (or admin) for self-hosted Ingress controllers. With
the late addition of
the [IngressClass](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
resource, the infrastructure provider and cluster operator became the owners of
that resource, and thus, explicit personas of the Ingress API.

The Gateway API
includes [four explicit personas](/concepts/security-model/#roles-and-personas):
the application developer, the application admin, the cluster operator, and the
infrastructure providers. This allows you to break away from the self-service
model by splitting the responsibilities of the user persona across those
personas (all except the infrastructure provider):

* The cluster operator/application admin defines entry points for the external
  client traffic including TLS termination configuration.
* The application developer defines routing rules for their applications that
  attach to those entry points.

Such a split adheres to a common organizational structure where multiple teams
share the same load-balancing infrastructure. At the same time, it is not
mandatory to give up the self-service model -- it is still possible to configure
a single RBAC Role that will fulfill the application developer, application
admin, and cluster operator responsibilities.

The table below summarizes the mapping between the Ingress API and the Gateway
API personas:

| Ingress API Persona | Gateway API Persona |
|-|-|
| User | Application developer, Application admin, Cluster operator |
| Cluster operator | Cluster operator |
| Infrastructure provider | Infrastructure provider |

### Available Features

The Ingress API comes with basic features only: TLS termination and
content-based routing of HTTP traffic based on the host header and the URI of a
request. To offer more features, Ingress controllers support them through
[annotations][anns] on the Ingress resource which are implementation-specific
extensions to the Ingress API.

[anns]:https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/
The annotations approach to extensibility has two negative consequences for
users of the Ingress API:

* *Limited portability*. Because so many features are available through
  annotations, switching between Ingress controllers becomes difficult or even
  impossible, as it is necessary to convert the annotations of one
  implementation to another (the other implementation might not even support
  some features of the first one). This limits the portability of the Ingress
  API.
* *Awkwardness of the API*. Because annotations are key-value strings (as
  opposed to a structured scheme like the spec of the Ingress resource) and
  applied at the top of a resource (rather than in the relevant parts of the
  spec), the Ingress API can become awkward to use, especially when a large
  number of annotations are added to an Ingress resource.

The Gateway API supports all the features of the Ingress resources and many
features that are only available through annotations. As a result, the Gateway
API is more portable than the Ingress API. Additionally, as the next section
will show, you will not need to use any annotations at all, which addresses the
awkwardness problem.

### Approach to Extensibility

The Ingress API has two extensions points:

* Annotations on the Ingress resource (described in the previous section)
* [Resource backends](https://kubernetes.io/docs/concepts/services-networking/ingress/#resource-backend),
   which is the ability to specify a backend other than a Service

The Gateway API is feature-rich compared with the Ingress API. However, to
configure some advanced features like authentication or common but non-portable
across data planes features like connection timeouts and health checks, you will
need to rely on the extensions of the Gateway API.

The Gateway API has the following primary extension points:

* *External references.* A feature (field) of a Gateway API resource can
  reference a custom resource specific to the Gateway implementation that
  configures that feature. For example:
    * [HTTPRouteFilter](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteFilter)
      can reference an external resource via the `extensionRef` field, thus
      configuring an implementation-specific filter.
    * [BackendObjectReference](/references/spec/#gateway.networking.k8s.io/v1beta1.BackendObjectReference)
      supports resources other than Services.
    * [SecretObjectReference](/references/spec/#gateway.networking.k8s.io/v1beta1.SecretObjectReference)
      supports resources other than Secrets.
* *Custom implementations*. For some features, it is left up to an
  implementation to define how to support them. Those features correspond to the
  implementation-specific
  (custom)  [conformance level](/concepts/conformance/#2-support-levels). For
  example:
    * The `RegularExpression` type of
      the [HTTPPathMatch](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPPathMatch).
* *Policies.* A Gateway implementation can define custom resources called
  Policies for exposing data plane features like authentication. The Gateway API
  does not prescribe the details of those resources. However, it prescribes a
  standard UX. See the [Policy attachment guide](/references/policy-attachment/)
  for more details. In contrast with the *external references* above, a Gateway
  API resource does not reference a Policy. Instead, a Policy must reference a
  Gateway API resource.

The extension points do not include annotations on the Gateway API resources.
This approach is strongly discouraged for implementations of the API.

## Mapping Ingress API features to Gateway API Features

This section will map the Ingress API features to the corresponding Gateway API
features, covering three major areas:

* Entry points
* TLS termination
* Routing rules

### Entry Points

Roughly speaking, an entry point is a combination of an IP address and port
through which the data plane is accessible to external clients.

Every Ingress resource has two implicit entry points -- one for HTTP and the
other for HTTPS traffic. An Ingress controller provides those entry points.
Typically, they will either be shared by all Ingress resources, or every Ingress
resource will get dedicated entry points.

In the Gateway API, entry points must be explicitly defined in
a [Gateway](/api-types/gateway/) resource. For example, if you want the data
plane to handle HTTP traffic on port 80, you need to define
a [listener](/references/spec/#gateway.networking.k8s.io/v1beta1.Listener) for
that traffic. Typically, a Gateway implementation provides a dedicated data
plane for each Gateway resource.

Gateway resources are owned by the cluster operator and the application admin.

### TLS Termination

The Ingress resource supports TLS termination via
the [TLS section](https://kubernetes.io/docs/concepts/services-networking/ingress/#tls),
where the TLS certificate and key are stored in a Secret.

In the Gateway API, TLS termination is a property of
the [Gateway listener](/references/spec/#gateway.networking.k8s.io/v1beta1.Listener),
and similarly to the Ingress, a TLS certificate and key are also stored in a
Secret.

Because the listener is part of the Gateway resource, the cluster operator and
application admin own TLS termination.

### Routing Rules

The [path-based routing rules](https://kubernetes.io/docs/concepts/services-networking/ingress/#path-types)
of the Ingress resource map directly to
the [routing rules](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteRule)
of the [HTTPRoute](/api-types/httproute/).

The [host-header-based routing rules](https://kubernetes.io/docs/concepts/services-networking/ingress/#name-based-virtual-hosting)
map to
the [hostnames](/references/spec/#gateway.networking.k8s.io/v1beta1.Hostname) of
the HTTPRoute. However, note that in the Ingress, each hostname has separate
routing rules, while in the HTTPRoute the routing rules apply to all hostnames.

> The Ingress API uses the term host while the Gateway API uses the hostname.
> This guide will use the Gateway API term to refer to the Ingress host.

> The `hostnames` of an HTTPRoute must match the `hostname` of the [Gateway listener](/references/spec/#gateway.networking.k8s.io/v1beta1.Listener).
> Otherwise, the listener will ignore the routing rules for the unmatched
> hostnames. See the [HTTPRoute documentation](/references/spec/#gateway.networking.k8s.io/v1beta1.HTTPRouteSpec).

HTTPRoutes are owned by the application developer.

The next three sections map additional features of the Ingress routing rules.

#### Rules Merging and Conflict Resolution

Typically, Ingress controllers merge routing rules from all Ingress resources
(unless they provision a data plane per each Ingress resource) and resolve
potential conflicts among the rules. However, both merging and conflict
resolution are not prescribed by the Ingress API, so Ingress controllers might
implement them differently.

In contrast, the Gateway API specifies how to merge rules and resolve conflicts:

* A Gateway implementation must merge the routing rules from all HTTPRoutes
  attached to a listener.
* Conflicts must be handled as
  prescribed [here](/concepts/guidelines/#conflicts). For example, more specific
  matches in a routing rule win over the less specific ones.

#### Default Backend

The
Ingress [default backend](https://kubernetes.io/docs/concepts/services-networking/ingress/#default-backend)
configures a backend that will respond to all unmatched HTTP requests related to
that Ingress resource. The Gateway API does not have a direct equivalent: it is
necessary to define such a routing rule explicitly. For example, define a rule
to route requests with the path prefix `/` to a Service that corresponds to the
default backend.

#### Selecting Data Plane to Attach to

An Ingress resource must specify
a [class](https://kubernetes.io/docs/concepts/services-networking/ingress/#ingress-class)
to select which Ingress controller to use. An HTTPRoute must specify which
Gateway (or Gateways) to attach to via
a [parentRef](/references/spec/#gateway.networking.k8s.io/v1beta1.ParentRef).

### Implementation-Specific Ingress Features (Annotations)

Ingress annotations configure implementation-specific features. Thus, converting
them to the Gateway API depends both on the Ingress controller and Gateway
implementations.

Luckily, some of the features supported through annotations are now part of the
Gateway API (HTTPRoute), primarily:

* Request redirects (including a TLS redirect)
* Request/response manipulation
* Traffic splitting
* Header, query param, or method-based routing

However, the remaining features remain largely implementation-specific. To
convert them, consult the Gateway implementation documentation to see
which [extension point](#approach-to-extensibility) to use.

## Example

This section shows an example of how to convert an Ingress resource to Gateway
API resources.

### Assumptions

The example includes the following assumptions:

* All resources belong to the same namespace.
* The Ingress controller:
    * Has the corresponding IngressClass resource  `prod` in the cluster.
    * Supports the TLS redirect feature via
      the `example-ingress-controller.example.org/tls-redirect` annotation.
* The Gateway implementation has the corresponding GatewayClass resource `prod`
  in the cluster.

Additionally, the content of the referenced Secret and Services as well as
IngressClass and GatewayClass are omitted for brevity.

### Ingress Resource

The Ingress below defines the following configuration:

* Configure a TLS redirect for any HTTP request for  `foo.example.com`
  and `bar.example.com` hostnames using
  the  `example-ingress-controller.example.org/tls-redirect` annotation.
* Terminate TLS for the `foo.example.com` and `bar.example.com` hostnames using
  the TLS certificate and key from the Secret `example-com`.
* Route HTTPS requests for the `foo.example.com` hostname with the URI
  prefix `/orders` to the `foo-orders-app` Service.
* Route HTTPS requests for the  `foo.example.com` hostname with any other prefix
  to the `foo-app` Service.
* Route HTTPS requests for the `bar.example.com` hostname with any URI to
  the `bar-app` Service.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-ingress
  annotations:
    some-ingress-controller.example.org/tls-redirect: "True"
spec:
  ingressClassName: prod
  tls:
  - hosts:
    - foo.example.com
    - bar.example.com
    secretName: example-com
  rules:
  - host: foo.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: foo-app
            port:
              number: 80
      - path: /orders
        pathType: Prefix
        backend:
          service:
            name: foo-orders-app
            port:
              number: 80
  - host: bar.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: bar-app
            port:
              number: 80
```

The next three sections convert that Ingress into Gateway API resources.

### Conversion Step 1 - Define Gateway

The following Gateway resource:

* Belongs to our GatewayClass `prod`.
* Provisions load balancing infrastructure (this depends on the Gateway
  implementation).
* Configures HTTP and HTTPS listeners (entry points), which the Ingress resource
  included implicitly:
    * The HTTP listener `http` on port `80`
    * The HTTPS listener `https` on port `443` with TLS termination with the
      cert and key stored in the `example-com` Secret, which is the same Secret
      used in the Ingress

Also, note that both listeners allow all HTTPRoutes from the same namespace
(which is the default setting) and restrict HTTPRoute hostnames to
the `example.com` subdomain (allow hostnames like `foo.example.com` but
not `foo.kubernetes.io`).

```yaml
{% include 'standard/simple-http-https/gateway.yaml' %}
```

### Conversion Step 2 - Define HTTPRoutes

The Ingress is split into two HTTPRoutes -- one for `foo.example.com` and one
for `bar.example.com` hostnames.

```yaml
{% include 'standard/simple-http-https/foo-route.yaml' %}
```

```yaml
{% include 'standard/simple-http-https/bar-route.yaml' %}
```

Both HTTPRoutes:

* Attach to the `https` listener of the Gateway resource from Step 1.
* Define the same routing rules as in the Ingress rules for the corresponding
  hostname.

### Step 3 - Configure TLS Redirect

The following HTTPRoute configures a TLS redirect, which the Ingress resource
configured via an annotation. The HTTPRoute below:

* Attaches to the `http` listener of our Gateway.
* Issues a TLS redirect for any HTTP request for the `foo.example.com`
  or `bar.example.com` hostnames.

```yaml
{% include 'standard/simple-http-https/tls-redirect-route.yaml' %}
```

## Automatic Conversion of Ingresses

The [Ingress to Gateway](https://github.com/kubernetes-sigs/ingress2gateway)
project helps translate Ingress resources to Gateway API resources, specifically
HTTPRoutes. The conversion results should always be tested and verified.
