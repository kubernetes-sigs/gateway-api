# Release Cycle

In Gateway API 1.2+, we will be following a more structured and predictable
release cycle that is inspired by the [upstream Kubernetes release
cycle](https://kubernetes.io/releases/release/).

## Goals

* Ensure a predictable release schedule that enables 2-3 releases a year
* Minimize the amount of time required from upstream API approvers and make it
  more predictable
* Avoid last minute additions to the scope of a release
* Prevent experimental channel from growing by requiring GEPs to leave or
  graduate before new ones can be added
* Ensure that SIG-Network TLs are in the loop throughout the process, and have a
  meaningful opportunity to review changes before a release
* Provide more advance notice to everyone (SIG-Network TLs, Docs Reviewers,
  Implementations, etc)

## Phases

### 1. Scoping

**Timeline:** 4-6 weeks

In this phase, the Gateway API maintainers and community will be responsible
for determining the set of features we want to include in the release. Although
we can always lessen scope after this point, we will avoid expanding the scope
of the release at a later point unless it is absolutely necessary (critical flaw
in design, security issue, etc).

A key guideline in this phase is that we want to avoid expanding the size of the
Experimental release channel. That means that each new experimental feature
should be accompanied by the graduation or removal of an enhancement that is
already in the Experimental channel.

Note that in many cases, this scoping work will require some initial work on
GEPs to determine their viability before committing to including them in a
release.

### 2. GEP Iteration and Review

**Timeline:** 5-7 weeks

In this phase, the Gateway API community will work to update GEPs and meet
graduation criteria for each feature that has been deemed in scope for the
release cycle. As we’re working on new features, we will bring these discussions
to the broader SIG-Network meetings for feedback throughout our development
process. If a GEP has not merged with the target status by the end of this
phase, it will be pulled from the scope of the release.

### 3. API Refinement and Documentation

**Timeline:** 3-5 weeks

This phase is entirely focused on translating the concepts defined in the GEP
(previous phase) into both API specification and documentation. This offers one
final chance for the Gateway API community to refine the details that have
already been agreed to in the GEP, but any modifications at this stage should be
minor. If either documentation or API Spec have not merged by the end of this
phase, this enhancement will be pulled from the scope of the release.

### 4. SIG-Network Review and Release Candidates

**Timeline:** 2-4 weeks

This phase officially begins with the review session scheduled with SIG-Network
TLs several weeks earlier in phase 3. In that review session, Gateway API
maintainers and SIG-Network TLs should reach an agreement on the following:

1. Any blockers for an initial release candidate
1. How much time, if any, SIG-Network TLs want to review any changes in this
   release
1. A time after which we can assume lazy consensus and move on with the final
   release of the API

In general, we expect each minor release to be preceded by two release
candidates. These release candidates will enable implementations to test against
our release, work out any bugs, and gather early feedback on the viability of
the release.

## What happens when GEPs are unable to meet a timeline?

There will be situations where the above phases and timelines can't be met by
a GEP for one reason or another. In special circumstances (particularly, when it
is is anticipated that only a little bit more time is needed to meet the goal) a
time extension _might_ be granted if requested by the GEP authors, at the
discretion of the maintainers.

In the normal case however, when a GEP misses the timeline for a phase it will
be pulled out of the release to maximize bandwidth and reduce disruptions to the
overall release timeline. In such a case the Gateway API maintainers will stop
progression and set the status of the GEP to a halted status (such as `Deferred`
or `Declined`) with a note on the GEP explaining why it reached this status and
what it would take to get it re-approved for work in a later iteration.

## Contributions Welcome in Each Phase

The following table illustrates when different kinds of contributions will be
welcome. There will be some exceptions to this, but it should be useful as an
overall guideline:

| | 1. Scope | 2. GEP | 3. API | 4. Review |
| - | :-: | :-: | :-:| :-: |
| New GEPs | ✅ | ❌ | ❌ | ❌ |
| Major GEP Updates | ✅ | ✅ | ❌ | ❌ |
| GEP Refinement | ✅ | ✅ | ✅ | ❌ |
| API Spec Additions | ❌ | ❌ | ✅ | ❌ |
| New Conformance Tests | ✅ | ✅ | ✅ | ❌ |
| Bug Fixes | ✅ | ✅ | ✅ | ✅ |
| Documentation | ✅ | ✅ | ✅ | ✅ |
| Review | ✅ | ✅ | ✅ | ✅ |

## Timeline

Given the above, we expect each release to take 14-22 weeks (4-5 months). At
least initially, Gateway API maintainers will set end dates for each phase as we
are beginning the phase. In future releases, we may choose to set all dates for
the release in advance.
