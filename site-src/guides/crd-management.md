# CRD Management

Gateway API is built with CRDs. That comes with a number of significant
benefits, notably that each release of Gateway API supports the 5 more recent
minor versions of Kubernetes. That means you likely won't need to upgrade your
Kubernetes cluster to get the latest version of this API.

Unfortunately, this extra flexibility also adds some room for confusion. This
guide aims to answer some of the most common questions related to Gateway API
CRD management.

## Who Should Manage CRDs?

Ultimately CRDs are a highly-privileged cluster-scoped resource. That means that
either a cluster admin or cluster provider should be responsible for managing
the CRDs in a cluster.

Practically that means that any of the following are reasonable approaches:

* Cluster admin installs CRDs
* Cluster provisioning tool or provider installs and manages CRDs

Some implementations may also want to bundle CRDs to simplify installation. This
is acceptable as long as they never:

1. Overwrite Gateway API CRDs that have unrecognized or newer versions.
1. Overwrite Gateway API CRDs that have a different release channel.
1. Remove Gateway API CRDs.

[Issue #2678](https://github.com/kubernetes-sigs/gateway-api/issues/2678)
explores one possible approach implementations could use to accomplish this.

## Upgrading to a new version

Gateway API releases CRDs in two [release
channels](../concepts/versioning.md#release-channels).
Sticking with standard channel CRDs will ensure CRD upgrades are both simpler
and safer.

### Overall Guidelines

1. Avoid moving backwards. New versions of CRDs can add new fields and features.
   Rolling back to a previous version of these CRDs could result in a loss of
   that configuration.
1. Read the release notes before upgrading. In some cases, they may contain some
   guidelines you need to follow before upgrading.
1. Understand the [Gateway API versioning policy](../concepts/versioning.md) so you
   know what can change.
1. Although it is usually safe to upgrade across multiple Gateway API minor
   versions at once, the safest and most widely tested path will involve
   upgrading one minor version at a time.

### Validating Webhook

A validating webhook was included with earlier versions of Gateway API. Starting
in v1.0, that webhook has formally been deprecated in favor of the CEL
validation included directly within CRDs. In Gateway API v1.1, the webhook will
be fully removed. That means that the validating webhook is no longer a
consideration when upgrading to newer Gateway API versions.

### API Version Removal

!!! note
    This is an advanced use case that is currently only applicable to users that
    have been using Gateway API since v0.5.0 within the same cluster.

It's possible that a Gateway API release will remove an alpha API version like
v1alpha2 in CRDs that have newer or more stable API versions. Within the
Standard Channel, the removal of an API version is spread into at least four
minor releases:

1. A newer API version is configured as the storage version.
1. Version is deprecated (will be noted in release notes and via deprecation
   warning when using deprecated API version).
1. Version is no longer served but is still included in the CRD for the sake
   of automatic translation between API versions.
1. Version is no longer included in the CRD.

If you were using a CRD that went through this process (including the storage
version migration), it's possible that some of your resources are stuck on the
older (deprecated) storage version. When a CRD storage version is updated, that
only takes effect when the individual resources using that CRD are saved again.

For example, if you created a "foo" GatewayClass using Gateway API v0.5.0 CRDs,
the storage version of that GatewayClass would be v1alpha2. If that "foo"
GatewayClass had never been modified or updated by the time you would not be
able to upgrade to Gateway API v1.0.0 CRDs because one of our resources was
still using v1alpha2 as a storage version and that was no longer included in the
CRD (step 4 above).

To be able to upgrade, you'd need to take some action that would update any
GatewayClasses that were using the old storage versions. For example, sending
an empty kubectl patch to each GatewayClass would have this effect. Fortunately
there's a tool that can automate this for us -
[kube-storage-version-migrator](https://github.com/kubernetes-sigs/kube-storage-version-migrator)
will automatically update resources to ensure they're using the latest storage
version.

### Experimental Channel

As the name implies, Experimental Channel does not provide the same stability
guarantees that Standard Channel does. When it comes to a minor release, any of
the following are possible for Experimental Channel CRDs:

* Breaking changes for existing API fields or resources
* Removing API fields or resources without prior deprecation

In practice this means that some upgrades to new Experimental versions may
require you to uninstall and reinstall the Experimental CRDs. If that is ever
the case, it will be clearly communicated in the release notes.
