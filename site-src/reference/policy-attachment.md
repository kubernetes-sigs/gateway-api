# Metaresources and Policy Attachment

Gateway API defines a Kubernetes object that _augments_ the behavior of an object
in a standard way as a _Metaresource_. ReferenceGrant
is an example of this general type of metaresource, but it is far from the only
one.

Gateway API also defines a pattern called _Policy Attachment_, which augments
the behavior of an object to add additional settings that can't be described
within the spec of that object.

A "Policy Attachment" is a specific type of _metaresource_ that can affect specific
settings across either one object (this is "Direct Policy Attachment"), or objects
in a hierarchy (this is "Inherited Policy Attachment").

This pattern is EXPERIMENTAL, and is described in [GEP-713](/geps/gep-713/).
Please see that document for technical details.
