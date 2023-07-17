# Frequently Asked Questions (FAQ)

*   **Q: How can I get involved with Gateway API?<br>**
    A: The [community](/contributing/community) page keeps track of how to get
    involved with the project.

*   **Q: Will Gateway API replace the Ingress API?<br>**
    A: No. The Ingress API is GA since Kubernetes 1.19. There are no
    plans to deprecate this API and we expect most Ingress controllers
    to support it indefinitely.

*   **Q: What are the differences between Ingress and Gateway API?<br>**
    A: Ingress primarily targets exposing HTTP applications with a
    simple, declarative syntax. Gateway API exposes a more general API
    for proxying that can be used for more protocols than just HTTP,
    and models more infrastructure components to provide better
    deployment and management options for cluster operators.

    For more information, see the [Migrating from Ingress](https://gateway-api.sigs.k8s.io/guides/migrating-from-ingress/) guide.

*   **Q: Will there be a default controller implementation (in this repo)?<br>**
    A: There is no current plan to have an "official" or "default"
    implementation. You will see the controller code in this repo be
    used for testing the support libraries.

*   **Q: How can I expose custom capabilities through Gateway API?<br>**
    A: There are a few mechanisms available
    for extending the API with implementation-specific capabilities:

    * The [Policy Attachment](https://gateway-api.sigs.k8s.io/references/policy-attachment/)
      model allows you to decorate Gateway API objects with implementation-specific CRDs. A
      policy or configuration object could match the Gateway API object either
      by name or by using an explicit object reference.

    * Use implementation-specific values for string fields in Gateway API resources.

    * As a last resort, use implementation-specific annotations on Gateway API objects.

    * Use API-defined extension points. Some Gateway
      API objects have explicit [extension points](/concepts/api-overview#extension-points)
      for implementations to use.

*  **Q: Where can I find Gateway API releases?<br>**
   A: Gateway API releases are tags of the [Github repository][1].
   The [Github releases][2] page shows all the releases.

* **Q: How should I think about alpha API versions?<br>**
  A: Similar to upstream Kubernetes, alpha API versions indicate that resources
  are still experimental in nature and may either be removed or changed in
  breaking ways in future releases of Gateway API.

  See the [Versioning](https://gateway-api.sigs.k8s.io/concepts/versioning/) documentation for more info.

* **Q: Which Kubernetes versions are supported?<br>**
  A: See our policy on [Supported Version](https://gateway-api.sigs.k8s.io/concepts/versioning/#supported-versions)

* **Q: Is SSL Passthrough supported?<br>**
  A: SSL Passthrough (wherein a Gateway routes traffic with the [Transport
  Layer Security (TLS)][tls] encryption _intact_ to a backend service instead of
  terminating it) is supported by [TLSRoutes][tlsroute]. See the
  [TLS Guide][tlsguide] for more details about passthrough and other TLS
  configurations.

* **Q: What's the difference between Gateway API and an API Gateway?<br>**
  A: An API Gateway is a general concept that describes anything that exposes
  capabilities of a backend service, while providing extra capabilities for
  traffic routing and manipulation, such as load balancing, request and response
  transformation, and sometimes more advanced features like authentication and
  authorization, rate limiting, and circuit breaking.

  Gateway API is an interface, or set of resources, that model service networking
  in Kubernetes. One of the main resources is a `Gateway`, which declares the
  Gateway type (or class) to instantiate and its configuration. As a Gateway
  Provider, you can implement the Gateway API to model Kubernetes service
  networking in an expressive, extensible, and role-oriented way.

  Most Gateway API implementations are API Gateways to some extent, but not all
  API Gateways are Gateway API implementations.

* **Q: Is Gateway API a standard for API Management?<br>**
  A: No. API Management is a much broader concept than what Gateway API aims to
  be, or what an API Gateway is intended to provide. An API Gateway can be an
  essential part of an API Management solution. Gateway API can be seen as a
  way to standardize on that aspect of API Management.

[1]: https://github.com/kubernetes-sigs/gateway-api
[2]: https://github.com/kubernetes-sigs/gateway-api/releases
[tls]:https://en.wikipedia.org/wiki/Transport_Layer_Security
[tlsroute]:/concepts/api-overview#tlsroute
[tlsguide]:/guides/tls
