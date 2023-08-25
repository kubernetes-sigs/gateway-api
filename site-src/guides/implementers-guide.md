# Gateway API Implementer's Guide

aka

Everything you wanted to know about building a Gateway API implementation
but were too afraid to ask.

This document is a place to collect tips and tricks for _writing a Gateway API
implementation_. It's intended to be a place to write down some guidelines to
help implementers of this API to skip making common mistakes. It may not be
very relevant if you are intending to _use_ this API as an end user as opposed
to _building_ something that uses it.

This is a living document, if you see something missing, PRs welcomed!

## Important things to remember about Gateway API

Hopefully most of these are not surprising, but they sometimes have non-obvious
implications that we'll try and lay out here.

### Gateway API is `kubernetes.io` API

Gateway API uses the `gateway.networking.k8s.io` API group. This means that,
like APIs delivered in the core Kubernetes binaries, each time a release happens,
the APIs have been reviewed by upstream Kubernetes reviewers, just like the APIs
delivered in the core binaries.

### Gateway API is delivered using CRDs

Gateway API is supplied as a set of CRDs, version controlled using our [versioning
policy](/concepts/versioning).

The most important part of that versioning policy is that what _appears to be_
the same object (that is, it has the same `group`,`version`, and `kind`) may have
a slightly different schema. We make changes in ways that are _compatible_, so
things should generally "just work", but there are some actions implementations
need to take to make "just work"ing more reliable; these are detailed below.

The CRD-based delivery also means that if an implementation tries to use (that is
get, list, watch, etc) Gateway API objects when the CRDs have _not_ been installed,
then it's likely that your Kubernetes client code will return serious errors.
Tips to deal with this are also detailed below.

## Implementation Guidelines

### CRD Management

### Version compatibility and conformance

### Resources to reconcile

### Listener traffic matching



