# Gateway API Release Manager

## Overview

The Gateway API release manager role is responsible for coordinating and managing the release, taking ultimate accountability for the success of the release, and ensuring that a retrospective happens.

This role is based on the Release Lead role in Kubernetes SIG-Release, and is intended to be similar - a servant leader of the release, not the primary decision maker.

As a Release Manager, you should expect to spend a lot of time talking to people about the status of changes, tracking changes yourself, and updating the community about what changes are in and out of scope for the release.

## Authority and Responsibility

The release manager has the authority to:

* Bring together the project’s leadership (including maintainers and other leads) to coordinate decisions about what is in scope for the release  
* Bring together the project’s leadership to coordinate decisions about granting extensions to deadlines in specific cases  
* Perform the actual release  
* Call the community (or some parts of it) together for status and retrospective meetings outside of the normal meeting cadence  
* Remind reviewers and approvers of outstanding work that needs to be done  
* Delegate these authorities to others (with the agreement of the maintainer team)

The release manager has the responsibility to:

* Ensure that the release has agreed-on dates for the final release, and for as many phases as possible. The community has indicated a clear preference for dates communicated up front as much as possible.  
* Ensure the release happens on the agreed upon date, or as close as possible, including assisting the community in meeting agreed-upon deadlines  
* Ensure that the community is informed about the progress of the release during the burndown process, particularly what is in and out of scope of the release. This can be by the maintenance of Github Boards or Milestones, or any other method that the Release Manager agrees with the leadership team and the community  
* Notify feature owners if their feature is at risk of or has fallen out of scope for the release  
* Refine the design of and improve documentation for the release process itself  
* Prepare the change log for the release  
* Coordinate the blog post for the release  
* Perform the implementation page review process after the release

There are some practical requirements that fall out of these lists.

The release manager must

* Update the community no less than weekly on work that is currently slated for the release  
* Try to minimize the amount of reviews required in short time frames (that is, try to remind people to spread out the work across the release as much as possible, rather than happening just before or on the deadlines)

## Time commitments

The release manager is expected to keep both themselves and the community up to date about what is happening, which will require a reasonably large amount of time each week, and that time should be expected to increase when deadlines are approaching. You should plan on, *at a minimum*, **1 to 2 hours per working day** to begin with, with probably **half your time** spent on the release at busy times, and an **entire day or two** for the actual release cutting process.

## Release dates and deadlines

The Release Manager is responsible for:

* Ensuring that a final release date is picked as early as possible and communicated to the community for planning purposes. This is the most important date, and should be designed so that any timing mistakes have a minimal impact on it, as the community will depend on the release being on or *very* close to this date.  
* Ensuring that intermediate deadlines are chosen and communicated as early as possible, and that any changes to those deadlines are both discussed by the leadership team and communicated to the community well in advance of them being relevant.  
* Ensuring that deadline extensions are kept to a minimum, and are on a case-by-case basis only. Feature leads MUST ask for extensions before they can be granted, and there MUST be enough time for the entire leadership team to see and agree with the extension before it is granted. Practically, this means that there must be a *minimum* of 24-48 hours (not counting weekends) between an extension being asked for and the deadline that is being extended.

## Work examples in the current release phases

### Scoping

The scoping phase is about determining what will get focus in the next release.

The way that the community determines what work is in scope for the release is ultimately up to the release manager, although they are expected to build some consensus amongst both the maintainers and the community about the method.

In the interest of ensuring that the changes we include are what the community is interested in, and that we are doing necessary but boring work like tech debt reduction, scoping MUST involve:

* Community feedback  
* Maintainer input (to ensure we try to solve underlying issues as well as new features)  
* Reviewer input (to ensure that we keep changes deliverable within a release cycle)  
* Ensuring that features are moving towards Standard, and are not stuck in Experimental

In all cases, the Release Manager *must* ensure that all the factors involved in the scoping decisions are publicly available.

For example, for v1.4, we had scoping performed using the votes on the Github discussions that proposed inclusion (which are publicly visible, and attributable to individuals), and for v1.3, we used community votes plus some additional factors like “difficulty of change” and “future importance”. In the v1.3 case, those weights were agreed upon by the maintainers and made available in a [world-readable Google sheet](https://docs.google.com/spreadsheets/d/1tLVmYHCyVuRLwnvMJuYMhiEtVX0fUF1aLHBLIc7mBKE/edit?gid=0#gid=0).

### GEP Refinement

The GEP refinement phase is about getting the agreed-upon GEPs to the phase they need to be at so that required API changes can be made in the next phase.

Generally, this involves the feature owners pushing PRs with their proposed changes, and their reviewer sponsor coordinating getting the changes reviewed.

The release manager’s most important role here is reminding everyone what changes are on the table, and poking feature owners to ensure they are actually working on the required changes. This is *much* more work than it sounds, particularly because contentious changes can often run to hundreds of comments that need to be tracked and followed up.

Generally, tracking requested changes should be the role of the feature owner and/or the reviewer sponsor, but the Release Manager should feel confident to chase things that look like they may have been missed.

The Release Manager is encouraged to set deadlines within this phase for gates like “GEP PR is open” “GEP PR is merged at Provisional”, “GEP PR is merged at Implementable”. Note that once a GEP is Implementable, this phase is complete for that GEP.

The Release Manager should also feel free to perform reviews themselves if they have the context and bandwidth to do so.

The Release Manager is responsible for notifying feature owners if their feature has fallen out of scope due to missing deadlines.

The Release Manager is the final arbiter for approving extensions for deadlines in this phase, and should do so *only* if the extension is requested in public, and on a case-by-case basis. No extensions for everything.

### API implementation phase

In this phase, we expect to merge any required changes to the API itself into the repo, preferably with conformance tests to ensure that implementations do things correctly when implementing the new changes.

The Release Manager’s role here is to:

* Keep track of changes as they happen  
* Remind feature owners of upcoming deadlines  
* Communicate what features are ongoing, what at risk, and what are out of scope  
* Ensure that deadlines are met  
* Grant extensions for deadlines, if required, on a case-by-case basis. 

Generally, in this phase, most of the changes are relatively straightforward, as most argument happens during the design phase in GEP Refinement. However, sometimes actually making the changes to the API, and in particular writing conformance tests, can show problems with the APIs as designed that may need further API changes to fix. Part of the Release Manager’s role is to be the arbiter of judgement calls about if further API changes are deliverable in the required timeframe, or if the change should be pushed back to the next version.

### API Review and Release Candidate phase

In this phase, we MUST get sign-off from SIG-Network API reviewers for *all* API changes. This is a requirement of having an official [`k8s.io`](http://k8s.io) API, and is not negotiable. Practically, API reviewers are usually very busy, and it can be difficult to get time with them, so it’s important to ensure that all the pending API changes are completed before getting the API reviewers to look.

The API reviewers *may* review an RC build, which would coincide with making the RC available to the community. That does run the risk that early implementers may need to change things as a result of API review feedback (this has definitely happened before).

The Release Candidate builds are intended to allow implementations to get started on actually implementing the API changes from the release early, so that they can provide feedback about the implementation experience, and source early user feedback if possible.

We generally try to do at least two candidate releases, with the requirement being that there MUST NOT be any changes between the final RC and the actual release.

## Measuring success

Getting a release out at all is a success!

But to be able to quantify releases a bit, and evaluate if changes to the release process are helping or not, we track some metrics about each release, which can be summarized into two main buckets:

* Release size (How many features, and how large is each feature? How many new conformance tests added? Any process changes during the release timeframe?)  
* Release timeliness (Was the release released on the targeted date? Did we meet agreed upon deadlines within the release schedule? How many extensions did we need to grant during the release?)
