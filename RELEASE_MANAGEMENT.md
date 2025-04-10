# Release Management

Major and minor releases for Gateway API are managed by a "Release Manager".
The responsibilities of the release manager include:

* Creating and managing a GitHub Milestone for the release.
* Creating and managing a GitHub Project board for the release.
* Creating and managing a GitHub Discussion Boards announcement for the release.
* Working through the [Release Phases](#release-phases) for the release.

This management process ultimately results in the manager of the release
shipping the release as per the [Release Process](#release-process). We will
go into more detail about this in the sections that follow.

> **Note**: Patch releases need someone assigned to them as well, however they
> just happen organically, and as soon as possible. We currently we don't bother
> sending patch releases through this entire process.

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

> **Note**: It's normal for this description to change over time, as things are
> never completely certain at the onset.

The release manager must assign a due date for the milestone, to communicate to
the community the intended date to ship the release. They must also indicate an
expected [feature freeze](#freeze-phase) date in the description.

> **Note**: This can change over time as well, but it's valuable to try and
> communicate dates in this way as many implementations and community members
> plan their releases on ours.

[GitHub Milestone]:https://github.com/kubernetes-sigs/gateway-api/milestones

## GitHub Project Board

The release manager must create a [project board] for the release, which will
be used to help track the work in the [milestone](#github-milestone) over time
to refine and analyze progress as the release is ongoing. The board should
include `Next`, `In Progress`, `Review` and `Done` columns to indicate progress
(which can be shared via updates on the community syncs as well).

[project board]:https://github.com/kubernetes-sigs/gateway-api/projects

## Release Phases

The release manager is responsible for tracking and managing the release
through several phases, ultimately resulting in shipping the release.

> **Note**: Updates on the community sync and discussion boards about the
> release process should be communicated regularly. We recommend bi-weekly
> unless there's a clear reason to do more, or less.

### Compile Phase

> **Note**: This phase should generally start **3 months before the projected
> due date**, to provide ample notice for the community and to help avoid any
> rush, particularly since all engineering time for Gateway API is inherently
> volunteer time and not guaranteed.

The [milestone](#github-milestone) and [project board](#github-project-board)
are created and configured first. A [GitHub Discussion Board] announcement is
created to signal the coming release to ask for input from contributors on
any features they expect they can deliver by the due date. The release
milestone, project board and announcement post are then announced at the next
available Gateway API community sync.

**This phase is expected to last for 1 month**, during that month the release
manager is responsible for bringing the release up repeatedly as a topic at
community meetings and seeking feedback on potential features for inclusion.

High priority bugs and smaller issues should be added to the milestone in
addition to features, and the release manager should be consistently requesting
volunteers from the community to work on those.

Major themes and features are expected to be hardened and communicated via the
milestone for this period. Any development on a feature someone desires to be
included in the release should be underway _early_ during this phase.

[GitHub Discussion Board]:https://github.com/kubernetes-sigs/gateway-api/discussions

### Freeze Phase

This is a feature freeze. The release announcement on the discussion board
should be updated indicating the feature freeze a week before it starts. This
should also be communicated at the next available community sync.

During this freeze exceptions to bring in any additional features or other
issues can be made only if agreed upon by the maintainers, but generally should
not be made unless deemed critical.

This process should **generally be around a month**, giving time for development
to complete.

### Release Phase

Any development which is not complete at this point should result in the issue
or feature in question getting dropped from the release, for potential pickup
in the next release. Exceptions can be granted under extenuating circumstances
if the maintainers agree.

This process should **generally be between 2 weeks and a month**, and include
shipping at least one **release candidate** for evaluation by the community.

Release candidates--and the eventual final release--must utilize the [Release
Process](#release-process) for delivery.
