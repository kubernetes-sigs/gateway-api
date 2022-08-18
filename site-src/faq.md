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

*   **Q: Will there be a default controller implementation (in this repo)?<br>**
    A: There is no current plan to have an "official" or "default"
    implementation. You will see the controller code in this repo be
    used for testing the support libraries.

*   **Q: How can I expose custom capabilities through Gateway API?<br>**
    A: There is a lot of diversity in the networking and proxying
    ecosystem, and many products will have features that are not directly
    supported in the API.  However, there are a few mechanisms available
    for extending the API with implementation-specific capabilities:

    * Decorate Gateway API objects with implementation-specific objects. A
      policy or configuration object could match the Gateway API object either
      by name or by using an explicit object reference.

        For example, given a `Gateway` object with name "inbound",
        creating a `AccessPolicy` object that also has the name "inbound"
        could cause an implementation to attach a specified access
        control policy. This is an example of matching the object by name.

    * Use implementation-specific values for string fields. In many
      places, the fields of Gateway API resources have the type
      "string". This allows an implementation to support custom values
      for those fields in addition to any values specified in the API.

    * Use implementation-specific annotations. For some kinds of
      configuration, implementations may choose to support custom
      annotations on Gateway API objects. This approach continues
      a proud tradition of extending Ingress objects.

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

* **Q: Which Kubernetes versions are supported?<br>**
  A: Generally, we support Kubernetes 1.16+, but certain features like 
  AppProtocol depend on Kubernetes 1.18 (opt-in) or 1.19 (on by default). 
  There are not any other exceptions to the 1.16+ guideline right now.

* **Q: Is SSL Passthrough supported?**
  A: SSL Passthrough (wherein a Gateway routes traffic with the [Transport
  Layer Security (TLS)][tls] encryption _intact_ to a backend service instead of
  terminating it) is supported by [TLSRoutes][tlsroute]. See the
  [TLS Guide][tlsguide] for more details about passthrough and other TLS
  configurations.

[1]: https://github.com/kubernetes-sigs/gateway-api
[2]: https://github.com/kubernetes-sigs/gateway-api/releases
[tls]:https://en.wikipedia.org/wiki/Transport_Layer_Security
[tlsroute]:/concepts/api-overview#tlsroute
[tlsguide]:/v1alpha2/guides/tls
