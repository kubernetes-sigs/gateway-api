# Release Management

Major and minor releases for Gateway API are managed by a "Release Manager".
The responsibilities of the release manager include:

* Creating and managing a GitHub Milestone for the release.
* Creating and managing a GitHub Project board for the release.
* Creating and managing a GitHub Discussion Boards announcement for the release.
  * This includes discussions to handle scoping for each release _channel_ as well.
* Working through the [Release Cycle](#release-phases) for the release.

This management process ultimately results in the manager of the release
shipping the release as per the [Release Process]. We will go into more detail
about this in the sections that follow.

> **Note**: Patch releases need someone assigned to them as well, however they
> just happen organically, and as soon as possible. We currently we don't bother
> sending patch releases through this entire process.

[Release Cycle]:https://gateway-api.sigs.k8s.io/contributing/release-cycle/
[Release Process]:/RELEASE.md

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
seeking volunteers for features to be included in the release, and thus the
release will be considered a **feature release**. If there are few or no
volunteers the release may simply end being a smaller **maintenance release**.

> **Note**: Updates on the community sync and discussion boards about the
> release process should be communicated regularly. We recommend bi-weekly
> unless there's a clear reason to do more, or less.

> **Note**: Communicating whether a release is expected to be a **feature
> release** or a **maintenance release** is largely done via the [Milestone].
> However, the final GitHub release should also make note of this in the
> release description.

Release candidates--and the eventual final release--must utilize the [Release
Process](/RELEASE.md) for delivery.

As the release nears completion, the release-manager should proactively reach
out to implementations to get them ready to send conformance reports for the
final release when it is cut.

[Release Cycle]:https://gateway-api.sigs.k8s.io/contributing/release-cycle/
[Milestone]:#github-milestone

## Time Extensions

Extensions to timelines may be requested by contributors. Our guidelines for
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
