# Release Management

Major and minor releases for Gateway API are managed by a "Release Manager".
The responsibilities of the release manager are defined in the [Release Manager]
role definition doc.

This management process ultimately results in the manager of the release
shipping the release as per the [Release Process]. We will go into more detail
about this in the sections that follow.

> **Note**: Patch releases need someone assigned to them as well, however they
> just happen organically, and as soon as possible. We currently we don't bother
> sending patch releases through this entire process.

[Release Cycle]:https://gateway-api.sigs.k8s.io/contributing/release-cycle/
[Release Process]:/RELEASE.md
[Release Manager]:/roles/RELEASE_MANAGER.md

## Assigning a Release Manager

The [maintainers] will find and assign a release managers for upcoming
releases. A release manager can be a community member on the contributor ladder
who volunteers (at the discretion of the maintainers), or otherwise the
responsibility falls back to the maintainers.

[maintainers]:/OWNERS_ALIASES

## GitHub Milestone

Once a manager is assigned for the release, they must create a [GitHub
Milestone] for the release, where the name of the milestone is the eventual tag
that will be cut (e.g., `v1.0.0`).

The release manager must mark themselves as the manager for the release at the
top of the milestone's description, and add any description to the milestone
that communicates to the community any major themes in the release, and
important notes:

```console
release-manager: @<name>

This milestone introduces the following new experimental features:

* <experimental-feature-1>
* <experimental-feature-2>
* <experimental-feature-3>

This milestone graduates the following features:

* <GA-feature-1>
* <GA-feature-2>

feature-freeze: <mm-dd-yy>
```

> **Note**: The GitHub Milestone may change over time, as things are added or
> removed.

The release manager must assign a due date for the milestone, to communicate to
the community the intended date to ship the release.

> **Note**: The due date may change as well, but it's valuable to try and
> communicate dates in this way as many implementations and community members
> plan their releases on ours. We will make every effort to avoid changes to the
> release date.

[GitHub Milestone]:https://github.com/kubernetes-sigs/gateway-api/milestones

## GitHub Project Board

The release manager must create a [project board] for the release, which will
be used to help track the work in the [milestone](#github-milestone) over time
to refine and analyze progress as the release is ongoing. The columns for this
board should generally align with the [Release Cycle].

> **Note**: Changes to this board should be shared via updates on the community
> syncs as well.

[project board]:https://github.com/kubernetes-sigs/gateway-api/projects
[Release Cycle]:https://gateway-api.sigs.k8s.io/contributing/release-cycle/

## Release Cycle

The release manager is responsible for tracking and managing the release
through several phases, ultimately resulting in shipping the release. That
cycle is outlined in the [Release Cycle] documentation and should be followed
according to that.

The release manager is responsible for communicating with the community and
seeking volunteers for features to be included in the release.

Each enhancement included in the release must have at least two people
responsible for it:

* **Owner**: This is the person who asks for the GEP to be included in the
  release, and who is primarily responsible for doing the work to make the feature
  happen.
* **Shepherd**: This is the person who agrees to be the primary reviewer, and
  to assist the feature through the GEP review process. Ideally, this person has
  been a feature owner in the past and has experience with the GEP review process,
  and can help smooth out any rough patches.

The Release Manager is expected to communicate regularly with Owners and Shepherds
about the progress of their feature.

Additionally, the Release Manager MUST keep the broader community updated on the
release, including what is currently in scope for the release, and any changes
or updates. This MUST be done at least weekly.

Release candidates--and the eventual final release--must utilize the [Release
Process](/RELEASE.md) for delivery.

As the release nears completion, the release-manager should proactively reach
out to implementations to get them ready to send conformance reports for the
final release when it is cut.

[Release Cycle]:https://gateway-api.sigs.k8s.io/contributing/release-cycle/
[Milestone]:#github-milestone

## Time Extensions

Extensions to deadlines may be requested by contributors. Our guidelines for
this are based on the Kubernetes process:

* Extensions can be granted on a per-GEP basis
  * The owners of the GEP have to ask and provide a timeline (measured in
    days) as to when they believe the GEP will be ready for merge.
* The request and approval for a GEP extension needs to be in public.
* Extensions can only be granted with a majority agreement by maintainers
  / release-managers

For our purposes we use GitHub discussions as the public place for
requesting/approving extensions. Contributors should use an existing
discussion for the release when feasible, otherwise create a discussion.
