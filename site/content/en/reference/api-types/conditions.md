---
title: "Conditions and Reasons"
weight: 11
---

Conditions provide a standardized way for controllers to communicate the status of resources to users. Each condition has a `type`, `status` (True, False, or Unknown), `reason`, and `message`.

For an introduction to conditions and troubleshooting guidance, see [Troubleshooting and Status](/docs/concepts/troubleshooting/).

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

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| InvalidParameters |  | ✓ |  |
| Pending |  |  | ✓ |
| Unsupported |  | ✓ |  |
| UnsupportedVersion |  | ✓ |  |

</div>


### SupportedVersion

> **Experimental:** This condition indicates whether the GatewayClass supports the version(s) of Gateway API CRDs present in the cluster. This condition MUST be set by a controller when it marks a GatewayClass "Accepted"

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| SupportedVersion | ✓ |  |  |
| UnsupportedVersion |  | ✓ |  |

</div>


---

## Gateway

### Programmed

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| AddressNotAssigned |  | ✓ |  |
| Invalid |  | ✓ |  |
| NoResources |  | ✓ |  |
| Pending |  | ✓ | ✓ |
| Programmed | ✓ |  |  |

</div>


### InsecureFrontendValidationMode

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| ConfigurationChanged | ✓ |  |  |

</div>


### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| Invalid |  | ✓ |  |
| InvalidParameters |  | ✓ |  |
| ListenersNotValid | ✓ | ✓ |  |
| NotReconciled |  | ✓ |  |
| Pending |  |  | ✓ |
| UnsupportedAddress |  | ✓ |  |

</div>


### Scheduled

> **Deprecated:** Use Accepted instead.


### ResolvedRefs

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| InvalidClientCertificateRef |  | ✓ |  |
| ListenersNotResolved |  | ✓ |  |
| RefNotPermitted |  | ✓ |  |
| ResolvedRefs | ✓ |  |  |

</div>


### Ready

> **Reserved for future use:** Not used by implementations. If used in the future, will represent the final state where all configuration is confirmed good and has completely propagated to the data plane.


---

## Listener (Gateway status)

Listeners are defined in `Gateway.spec.listeners`. Their status appears in `Gateway.status.listeners[].conditions`.

### Conflicted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| HostnameConflict | ✓ |  |  |
| NoConflicts |  | ✓ |  |
| ProtocolConflict | ✓ |  |  |

</div>


### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| NoValidCACertificate |  | ✓ |  |
| Pending |  |  | ✓ |
| PortUnavailable |  | ✓ |  |
| UnsupportedProtocol |  | ✓ |  |
| UnsupportedValue |  | ✓ |  |

</div>


### Detached

> **Deprecated:** Use Accepted instead.


### ResolvedRefs

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| InvalidCACertificateKind |  | ✓ |  |
| InvalidCACertificateRef |  | ✓ |  |
| InvalidCertificateRef |  | ✓ |  |
| InvalidRouteKinds |  | ✓ |  |
| RefNotPermitted |  | ✓ |  |
| ResolvedRefs | ✓ |  |  |

</div>


### Programmed

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Invalid |  | ✓ |  |
| Pending |  | ✓ | ✓ |
| Programmed | ✓ |  |  |

</div>


### OverlappingTLSConfig

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| OverlappingCertificates | ✓ |  |  |
| OverlappingHostnames | ✓ |  |  |

</div>


### Ready

> **Reserved for future use:** Not used by implementations. If used in the future, will represent the final state where all configuration is confirmed good and has completely propagated to the data plane.


---

## Policy resources (BackendTLSPolicy, BackendTrafficPolicy)

### ResolvedRefs

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| InvalidCACertificateRef |  | ✓ |  |
| InvalidKind |  | ✓ |  |
| ResolvedRefs | ✓ |  |  |

</div>


### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| Conflicted |  | ✓ |  |
| Invalid |  | ✓ |  |
| TargetNotFound |  | ✓ |  |

</div>


---

## ListenerSet

### Programmed

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Invalid |  | ✓ |  |
| ListenersNotValid |  | ✓ |  |
| ParentNotProgrammed |  | ✓ |  |
| Pending |  |  | ✓ |
| PortUnavailable |  | ✓ |  |
| Programmed | ✓ |  |  |

</div>


### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| Invalid |  | ✓ |  |
| ListenersNotValid |  | ✓ |  |
| NotAllowed |  | ✓ |  |
| ParentNotAccepted |  | ✓ |  |
| Pending |  |  | ✓ |

</div>


---

## ListenerEntry (ListenerSet status)

ListenerEntries are defined in `ListenerSet.spec.listeners`. Their status appears in `ListenerSet.status.listeners[].conditions`. ListenerEntries represent listeners from both the Gateway and attached ListenerSets.

### Conflicted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| HostnameConflict | ✓ |  |  |
| ListenerConflict | ✓ |  |  |
| NoConflicts |  | ✓ |  |
| ProtocolConflict | ✓ |  |  |

</div>


### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| Pending |  |  | ✓ |
| PortUnavailable |  | ✓ |  |
| TooManyListeners |  | ✓ |  |
| UnsupportedProtocol |  | ✓ |  |

</div>


### ResolvedRefs

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| InvalidCertificateRef |  | ✓ |  |
| InvalidRouteKinds |  | ✓ |  |
| RefNotPermitted |  | ✓ |  |
| ResolvedRefs | ✓ |  |  |

</div>


### Programmed

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Invalid |  | ✓ |  |
| Pending |  |  | ✓ |
| PortUnavailable |  | ✓ |  |
| Programmed | ✓ |  |  |

</div>


### Ready

> **Reserved for future use:** Not used by implementations. If used in the future, will represent the final state where all configuration is confirmed good and has completely propagated to the data plane.


---

## Routes (HTTPRoute, GRPCRoute, TLSRoute, TCPRoute, UDPRoute)

Routes share the same condition types. Status appears in `Route.status.parents[].conditions` (per parent) and `Route.status.conditions` (route-level).

### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| NoMatchingListenerHostname |  | ✓ |  |
| NoMatchingParent |  | ✓ |  |
| NotAllowedByListeners |  | ✓ |  |
| Pending |  |  | ✓ |
| UnsupportedValue |  | ✓ |  |

</div>


### ResolvedRefs

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| BackendNotFound |  | ✓ |  |
| InvalidKind |  | ✓ |  |
| RefNotPermitted |  | ✓ |  |
| ResolvedRefs | ✓ |  |  |
| UnsupportedProtocol |  | ✓ |  |

</div>


### PartiallyInvalid

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| UnsupportedValue | ✓ |  |  |

</div>


---

## Mesh

> **Experimental:** See [GEP-3949](/geps/gep-3949/).

### Accepted

<div class="conditions-compact-table">

| Reason | True | False | Unknown |
| --- | --- | --- | --- |
| Accepted | ✓ |  |  |
| InvalidParameters |  | ✓ |  |
| Pending |  |  | ✓ |

</div>


---
