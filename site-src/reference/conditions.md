# Condition Types and Reasons Reference

Conditions provide a standardized way for controllers to communicate the status of resources to users. Each condition has a `type`, `status` (True, False, or Unknown), `reason`, and `message`.

For an introduction to conditions and troubleshooting guidance, see [Troubleshooting and Status](../concepts/troubleshooting.md).

## Common Conditions

The following condition types are used across multiple Gateway API resources:

| Condition | Description |
|-----------|-------------|
| **Accepted** | True when the object is semantically and syntactically valid, will produce some configuration in the underlying data plane, and has been accepted by a controller. |
| **Programmed** | True when an object's configuration has been fully parsed and successfully sent to the data plane. It will be ready "soon"—the exact definition depends on the implementation. |
| **ResolvedRefs** | True when all references to other objects (e.g., Secrets, Services) are valid—the objects exist and each reference is valid for the field where it is used. |

---

## GatewayClass

### Accepted

Indicates whether the GatewayClass has been accepted by the controller specified in `spec.controllerName`. Defaults to Unknown; the controller MUST set this when it sees a GatewayClass using its controller string.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | GatewayClass is accepted; controller will provision Gateways using this class. |
| InvalidParameters | | ✓ | | `parametersRef` refers to a namespaced resource without Namespace set, cluster-scoped with Namespace set, nonexistent object, unsupported resource/kind, or malformed data. |
| Pending | | | ✓ | Controller has not yet made a decision. Default on new GatewayClass. |
| Unsupported | | ✓ | | Implementation does not support user-defined GatewayClass. |
| UnsupportedVersion | | ✓ | | Gateway API CRD version in cluster is not supported. |
| Waiting | | ✓ | | *Deprecated:* Use Pending instead. |

### SupportedVersion

??? experimental "Experimental"

    Indicates whether the GatewayClass supports the Gateway API CRD version(s) present in the cluster. The version is defined by the `gateway.networking.k8s.io/bundle-version` annotation on the CRD.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| SupportedVersion | ✓ | | All Gateway API CRDs in cluster have supported versions. |
| UnsupportedVersion | | ✓ | One or more CRDs have unrecognized or unsupported versions. Message should include detected and supported versions. |

---

## Gateway

### Accepted

True when the controller finds the Gateway syntactically and semantically valid enough to produce configuration in the underlying data plane. Does not indicate whether configuration has been propagated.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | Gateway is accepted. |
| ListenersNotValid | ✓ | ✓ | | One or more Listeners have invalid/unsupported configuration. Can be True or False depending on impact. |
| Invalid | | ✓ | | Gateway is syntactically or semantically invalid (e.g., unspecified TLS, invalid values). |
| InvalidParameters | | ✓ | | `parametersRef` is invalid; see message for details. |
| NotReconciled | | ✓ | | *Deprecated:* Use Pending instead. |
| UnsupportedAddress | | ✓ | | Provided address type is not supported by the implementation. |
| Pending | | | ✓ | No controller has reconciled the Gateway yet. |

### Programmed

Indicates whether the Gateway has generated configuration assumed to be ready soon in the underlying data plane. A positive-polarity summary condition; should always be present with ObservedGeneration set.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Programmed | ✓ | | | Configuration is ready. |
| Invalid | | ✓ | | Gateway is syntactically or semantically invalid. |
| Pending | | ✓ | ✓ | Not yet reconciled or not yet ready. |
| NoResources | | ✓ | | Insufficient infrastructure resources available. |
| AddressNotAssigned | | ✓ | | Address not yet allocated (e.g., IPAM exhaustion). |
| AddressNotUsable | | ✓ | | Provided address cannot be used (e.g., not found, in use). |

### ResolvedRefs

??? experimental "Experimental"

    Indicates whether the controller resolved all object references for the Gateway (excluding Listener-specific refs). Also provides a summary of Listeners' ResolvedRefs. Does not directly impact Accepted or Programmed.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| ResolvedRefs | ✓ | | All references resolved. |
| RefNotPermitted | | ✓ | Backend TLS config references object in another namespace without ReferenceGrant. |
| InvalidClientCertificateRef | | ✓ | ClientCertificateRef is invalid (nonexistent, unsupported, or malformed). |
| ListenersNotResolved | | ✓ | One or more Listeners have ResolvedRefs set to False. |

### InsecureFrontendValidationMode

True when FrontendValidationModeType is configured to allow insecure fallback. Removed when changed back to `AllowValidOnly`.

| Reason | True | Description |
|--------|------|-------------|
| ConfigurationChanged | ✓ | FrontendValidationModeType changed from `AllowValidOnly` to `AllowInsecureFallback`. |

### Ready

!!! warning "Reserved for future use"

    Not used by implementations. If used in the future, will represent the final state where all configuration is confirmed good and has completely propagated to the data plane.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| Ready | ✓ | | *Reserved* |
| ListenersNotReady | | ✓ | *Reserved* |

### Scheduled

!!! warning "Deprecated"

    Use Accepted instead.

| Reason | True | Description |
|--------|------|-------------|
| Scheduled | ✓ | *Deprecated:* Use Accepted with reason Accepted. |

---

## Listener (Gateway status)

Listeners are defined in `Gateway.spec.listeners`. Their status appears in `Gateway.status.listeners[].conditions`.

### Accepted

Indicates the Listener is syntactically and semantically valid and all features are supported. Generally True when the configuration will generate at least some data plane configuration.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | Listener is accepted. |
| PortUnavailable | | ✓ | | Port cannot be used (already in use, not supported). |
| UnsupportedProtocol | | ✓ | | Protocol type is not supported. |
| NoValidCACertificate | | ✓ | | Could not resolve references to any CACertificate for client cert validation. |
| UnsupportedValue | | ✓ | | Field value not supported by implementation. |
| Pending | | | ✓ | Not yet reconciled or not yet ready. |

### Conflicted

Indicates the controller could not resolve conflicting specification requirements. If conflicted, the Listener's network port should not be configured.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| HostnameConflict | ✓ | | Conflicts with hostnames in other Listeners (e.g., same port, same hostname). |
| ProtocolConflict | ✓ | | Same port, conflicting protocol specifications. |
| NoConflicts | | ✓ | No conflicts. |

### ResolvedRefs

Indicates whether the controller resolved all object references for the Listener.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| ResolvedRefs | ✓ | | All references resolved. |
| InvalidCertificateRef | | ✓ | TLS CertificateRef is invalid (nonexistent, unsupported, or malformed). Use only when reference is allowed. |
| InvalidRouteKinds | | ✓ | Invalid or unsupported Route kind specified. |
| RefNotPermitted | | ✓ | TLS config references object in another namespace without ReferenceGrant. |
| InvalidCACertificateRef | | ✓ | CACertificate reference for client cert validation is invalid. |
| InvalidCACertificateKind | | ✓ | CACertificate reference has unknown or unsupported kind. |

### Programmed

Indicates whether the Listener has generated configuration that will soon be ready in the data plane.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Programmed | ✓ | | | Configuration ready. |
| Invalid | | ✓ | | Listener is syntactically or semantically invalid. |
| Pending | | ✓ | ✓ | Not yet reconciled or not yet ready. |

### OverlappingTLSConfig

Negative polarity condition (only set when True). Indicates TLS configuration conflicts with another Listener on the same port—overlapping hostnames (e.g., `*.example.com` vs `foo.example.com`) or overlapping certificate SANs.

| Reason | True | Description |
|--------|------|-------------|
| OverlappingHostnames | ✓ | Overlapping hostnames detected. |
| OverlappingCertificates | ✓ | Overlapping certificate SANs. Prefer this if both apply. |

### Detached

!!! warning "Deprecated"

    Use Accepted instead.

| Reason | False | Description |
|--------|-------|-------------|
| Attached | ✓ | *Deprecated:* Use Accepted with reason Accepted. |

### Ready

!!! warning "Reserved for future use"

    Not used by implementations.

---

## ListenerSet

### Accepted

True when the controller finds the ListenerSet syntactically and semantically valid enough to produce configuration. Does not indicate propagation to the data plane.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | ListenerSet is accepted. |
| Invalid | | ✓ | | ListenerSet is syntactically or semantically invalid. |
| NotAllowed | | ✓ | | ListenerSet is not allowed to attach to the Gateway. |
| ParentNotAccepted | | ✓ | | Parent Gateway is not accepted. |
| ListenersNotValid | | ✓ | | One or more Listeners have invalid/unsupported configuration. |
| Pending | | | ✓ | No controller has reconciled the ListenerSet yet. |

### Programmed

Indicates whether the ListenerSet has generated configuration assumed to be ready soon in the data plane.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Programmed | ✓ | | | Configuration ready. |
| Invalid | | ✓ | | ListenerSet is syntactically or semantically invalid. |
| ParentNotProgrammed | | ✓ | | Parent Gateway is not programmed. |
| ListenersNotValid | | ✓ | | One or more Listeners have invalid configuration. |
| PortUnavailable | | ✓ | | From child ListenerEntry conditions. |
| Pending | | | ✓ | No controller has reconciled yet. |

---

## ListenerEntry

ListenerEntries are defined in `ListenerSet.spec.listeners`. Their status appears in `ListenerSet.status.listeners[].conditions`. ListenerEntries represent listeners from both the Gateway and attached ListenerSets.

### Accepted

Indicates the Listener is syntactically and semantically valid. Generally True when the configuration will generate at least some data plane configuration.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | Listener is accepted. |
| PortUnavailable | | ✓ | | Port cannot be used (in use, not supported, or cannot be assigned). |
| UnsupportedProtocol | | ✓ | | Protocol type is not supported. |
| TooManyListeners | | ✓ | | Gateway has too many Listeners; implementation rejected this one. |
| Pending | | | ✓ | Not yet reconciled or not yet ready. |

### Conflicted

Indicates the controller could not resolve conflicting specification requirements.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| HostnameConflict | ✓ | | Conflicts with hostnames in other Listeners. |
| ProtocolConflict | ✓ | | Same port, conflicting protocols. |
| ListenerConflict | ✓ | | Generic conflict; use when multiple conflict types apply (prefer most specific). |
| NoConflicts | | ✓ | No conflicts. |

### ResolvedRefs

Indicates whether the controller resolved all object references for the Listener.

| Reason | True | False | Description |
|--------|------|-------|-------------|
| ResolvedRefs | ✓ | | All references resolved. |
| InvalidCertificateRef | | ✓ | TLS CertificateRef is invalid. Use only when reference is allowed. |
| InvalidRouteKinds | | ✓ | Invalid or unsupported Route kind specified. |
| RefNotPermitted | | ✓ | TLS config references object in another namespace without ReferenceGrant. |

### Programmed

Indicates whether the Listener has generated configuration that will soon be ready in the data plane.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Programmed | ✓ | | | Configuration ready. |
| Invalid | | ✓ | | Listener is syntactically or semantically invalid. |
| PortUnavailable | | ✓ | | Port cannot be used. |
| Pending | | ✓ | ✓ | Not yet reconciled or not yet ready. |

### Ready

!!! warning "Reserved for future use"

    Not used by implementations.

---

## Routes (HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute)

Routes share the same condition types. Status appears in `Route.status.parents[].conditions` (per parent) and `Route.status.conditions` (route-level).

### Accepted

Indicates whether the Route has been accepted or rejected by a Gateway, and why. A Route is Accepted if at least one rule is implemented by the Gateway.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | Route accepted by the Gateway. |
| NotAllowedByListeners | | ✓ | | No Listener's allowedRoutes criteria permit the Route. |
| NoMatchingListenerHostname | | ✓ | | No compatible Listener hostname matches the Route. |
| NoMatchingParent | | ✓ | | ParentRef Port/SectionName does not match any Listener. |
| UnsupportedValue | | ✓ | | Enum value not recognized. |
| Pending | | | ✓ | Controller has not yet reconciled the Route. |
| IncompatibleFilters | | ✓ | | Incompatible filters on a rule (e.g., URLRewrite and RequestRedirect on HTTPRoute). |

### ResolvedRefs

Indicates whether the controller resolved all object references for the Route (e.g., BackendRefs).

| Reason | True | False | Description |
|--------|------|-------|-------------|
| ResolvedRefs | ✓ | | All references resolved. |
| RefNotPermitted | | ✓ | BackendRef to object in another namespace without ReferenceGrant. |
| InvalidKind | | ✓ | Reference to unknown or unsupported Group/Kind. |
| BackendNotFound | | ✓ | Referenced resource does not exist. |
| UnsupportedProtocol | | ✓ | Referenced resource has app protocol not supported by implementation. |

### PartiallyInvalid

Indicates the Route contains both valid and invalid rules. **Only set when True.** When set, implementations either drop invalid rules (message prefix "Dropped Rule") or fall back to last known good state (message prefix "Fall Back").

| Reason | True | Description |
|--------|------|-------------|
| UnsupportedValue | ✓ | Some rules have unsupported values; valid rules may still be in effect. |

---

## Mesh

??? experimental "Experimental"

    Mesh is an experimental resource. See [GEP-3949](../geps/gep-3949/index.md).

### Accepted

Indicates whether the Mesh has been accepted by the controller specified in `spec.controllerName`. Defaults to Unknown.

| Reason | True | False | Unknown | Description |
|--------|------|-------|---------|-------------|
| Accepted | ✓ | | | Mesh is accepted. |
| InvalidParameters | | ✓ | | `parametersRef` refers to invalid resource (wrong Namespace, nonexistent, unsupported, or malformed). |
| Pending | | | ✓ | Controller has not yet made a decision. Default on new Mesh. |
