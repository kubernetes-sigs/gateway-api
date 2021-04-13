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
    and models more infrastucture components to provide better
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

        Gateway API uses this decorator pattern with the
        [`BackendPolicy`](/references/spec/#networking.x-k8s.io/v1alpha1.BackendPolicy)
        resource, which modifies how a `Gateway` should forward traffic
        to a backend target (commonly a Kubernetes `Service`). This is
        an example of using explicit object references.

    * Use implementation-specific values for string fields. In many
      places, the fields of Gateway API resources have the type
      "string". This allows an implementation to support custom values
      for those fields in addition to any values specified in the API.

    * Use implementation-specific annotations. For some kinds of
      configuration, implementations may choose to support custom
      annotations on Servce API objects. This approach continues
      a proud tradition of extending Ingress objects.

    * Use API-defined extension points. Some Service
      API objects have explicit [extension points](/concepts/api-overview#extension-points)
      for implementations to use.

*  **Q: Where can I find Gateway API releases?<br>**
   A: Gateway API releases are tags of the [Github repository][1].
   The [Github releases][2] page shows all the releases.

* **Q: How should I think about the alpha release?<br>**
  A: The `v1alpha1` release will be the first Gateway API release. As
  various projects begin implementing the API, and operators start using
  it, the working group will collect feedback and issues, which will
  guide what revisions are needed for the next release. It is possible
  that the next release will contain breaking changes.


[1]: https://github.com/kubernetes-sigs/gateway-api
[2]: https://github.com/kubernetes-sigs/gateway-api/releases
