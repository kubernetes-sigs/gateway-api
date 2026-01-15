# Troubleshooting and Status

One of the biggest problems when using any Kubernetes object is how to know if the state requested by that object (generally encoded in its `spec` stanza) has been accepted, and when the state has been achieved.
The current status of most Kubernetes objects is stored in the `status` subresource and stanza, but in Gateway API, we've needed to lean hard into emphasizing the use of `status`.

One of the driving rules for Gateway API object design has been to try to ensure that, when a user creates an object, they can see as much as possible of the state of the system in the status of that object.
And, if that is not possible, that there is a way to find out what other objects are relevant.

## Status and Conditions

In particular, Gateway API has leant hard on the convention of Conditions, a portable respresentation of states of any given object.

Conditions have:

* a `type` (a CamelCase, single-word name for the state),
* a `status` (a boolean that indicates if that state is active or not),
* a `reason` (a CamelCase, single-word reason why the Condition is or is not in the state),
* and a `message` (a string representation of the reason, that is intended for human consumption).

One optional field that Gateway API _requires_ is the `observedGeneration` field, which indicates the value of the autoincremented `metadata.generation` field on the object at the time the status was written.
This functions as a staleness detection checksum - for any Gateway API status, you should check that the `observedGeneration` on its `conditions` matches the `metadata.generation` field.

If it does not, then that status is out of date, and for some reason your Gateway API implementation is not updating status correctly.
(This could be a controller fault, or the object may have fallen out of the implementation's scope.)

Additionally, part of the purpose of Gateway API is to fix some of the problems of earlier approaches, and we wanted to avoid the requirement to be able to look at the logs of an implementation to see what is happening.

We want the state of your object to be, as far as possible, visible _on_ your object.

This leads us to the first, most important rule of using Gateway API:

!!! info
    **_When troubleshooting Gateway API objects, always check the `status.conditions` of the object first._**

Every Gateway object has a `conditions` array in its `status` somewhere, and most have it at `status.conditions`.

We've also tried to re-use the same Condition `type`s as far as possible, and have a few commonly-used Conditions across multiple objects:

* `Accepted`: True when the object is semantically and syntactically valid, will produce some configuration in any underlying data plane, and has been accepted by a controller.
* `Programmed`: True when an object's config has been fully parsed, and has been successfully sent to a data plane for configuration. It will be ready "soon", where soon can have different definitions depending on the exact implementation.
* `ResolvedRefs`: True when all references to other objects inside an object are valid, in that the objects referred to exist, and each is a valid reference for the field where it is used.

In some cases, such as the Gateway object, there are additional Conditions arrays - on the Gateway object, there is also a Condition per `listener` field, as that status is also complex enough to need further clarification.

Conditions are complex enough to be difficult to summarize in a single line, so most `kubectl get` commands cannot summarize them correctly.

To check the status, you have a few options:

* `kubectl get -o yaml` - this will get you the full object, in YAML format, which includes the `status`.
* `gwctl` is a command-line tool created by the Gateway API subproject, which is designed to make managing Gateway API resources easier. It's available on the [Github repo](https://github.com/kubernetes-sigs/gwctl)
* `kubectl describe` - this will get you a more readable version of the full output, which can usually parse Conditions arrays correctly and show them. However, it often struggles to decode CustomResourceDefinitions correctly, especially when rendering lists.

### Scope and Status

One other peculiarity of Gateway API is that it is designed to allow for multiple implementations to run in the same cluster. 
In order to do this, there are strict requirements about what objects an implementation can update the status for. 
This is referred to as an object **being in scope** for a particular implementation.

!!! info
    **If an implementation cannot establish a chain of ownership from any
    object to a GatewayClass it owns, then the object is not in scope for that implementation, and MUST NOT have its status updated by it.**

This is so that multiple implementations do not end up fighting over status, repeatedly attempting to update the status, only to have one implementation's changes overwritten by another, and so on.

One important effect of this is that if a Route has a `parentRef` that does not point to a valid parent, then **there will be no status update to indicate that**.
The implementation cannot tell you that you made a mistake in pointing to a Gateway it cares about, because it has no way of knowing if that parentRef is its responsibility or not.

To put this another way, **a Route with an invalid parentRef will have no status to indicate that**. You should _always_ expect to see a status update for _any_ change in an in-scope object, even if it's just updating the `observedGeneration`.

