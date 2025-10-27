# Gateway Enhancement Proposal (GEP)

Gateway Enhancement Proposals (GEPs) serve a similar purpose to the [KEP][kep]
process for the main Kubernetes project:

1. Ensure that changes to the API follow a known process and discussion
  in the OSS community.
1. Make changes and proposals discoverable (current and future).
1. Document design ideas, tradeoffs, decisions that were made for
  historical reference.
1. Record the results of larger community discussions.
1. Record changes to the GEP process itself.

## Process

This diagram shows the state diagram of the GEP process at a high level, but the details are below.

<div align="center">

```mermaid
flowchart TD
    D([Discuss with<br />the community]) --> C
    C([Issue Created]) -------> Memorandum
    C([Issue Created]) -->|GEP PR with Who/What/Why merged| Provisional
    Provisional -->|GEP Doc PR with _proposed_ API details merged| Prototyping
    Provisional -->|GEP Doc PR with agreed API details merged| Implementable
    Prototyping -->|GEP Doc PR with agreed API details merged| Implementable
    Implementable -->|API Changes implemented in Go types and YAML| Experimental
    Experimental -->|Supported in<br />multiple implementations<br />+ Conformance tests| Standard
    Standard -->|Entire change is GA or implemented| Completed

```

</div>

## GEP Definitions

### GEP States

Each GEP has a state, which tracks where it is in the GEP process.

GEPs can move to some states from any other state:

  * **Deferred:** We do not currently have bandwidth to handle this GEP, it
    may be revisited in the future.
  * **Declined:** This proposal was considered by the community but ultimately
  rejected and further work will not occur.
  * **Withdrawn:** This proposal was considered by the community but ultimately
  withdrawn by the author.

There is a special state to cover Memorandum GEPs:

  * **Memorandum**: These GEPs either:
    * Document an agreement for further work, creating no spec changes themselves, or
    * Update the GEP process.

API GEPs flow through a number of states, which generally correspond to the level
of stability of the change described in the GEP:

  * **Provisional:** The goals described by this GEP have consensus but
    implementation details have not been agreed to yet.
  * **Prototyping:** An optional extension of `Provisional` in
    order to indicate to the community that there are some active practical tests
    and experiments going on which are intended to be a part of the development
    of this GEP. This may include APIs or code, but that content _must_ not be
    distributed with releases.
  * **Implementable:** The goals and implementation details described by this GEP
    have consensus but have not been fully implemented yet.
  * **Experimental:** This GEP has been implemented and is part of the
    "Experimental" release channel. Breaking changes are still possible, up to
    and including complete removal and moving to `Rejected`.
  * **Standard:** This GEP has been implemented and is part of the
    "Standard" release channel. It should be quite stable.
  * **Completed**: All implementation work on this API GEP has been completed.

### Relationships between GEPs

GEPs can have relationships between them. At this time, there are three possible
relationships:

* **Obsoletes** and its backreference **ObsoletedBy**: when a GEP is made obsolete
  by another GEP, and has its functionality completely replaced. The Obsoleted
  GEP is moved to the **Declined** state.
* **Extends** and its backreference **ExtendedBy**: when a GEP has additional details
  or implementation added in another GEP.
* **SeeAlso**: when a GEP is relevant to another GEP, but is not affected in any
  other defined way.

Relationships are tracked in the YAML metadata files accompanying each GEP.

### GEP metadata file

Each GEP has a YAML file containing metadata alongside it, please keep it up to
date as changes to the GEP occur.

In particular, note the `authors`, and `changelog` fields, please keep those up
to date.

## Process

### 1. Discuss with the community

Before creating a GEP, share your high level idea with the community. There are
several places this may be done:

- A [new GitHub Discussion](https://github.com/kubernetes-sigs/gateway-api/discussions/new)
- On our [Slack Channel](https://kubernetes.slack.com/archives/CR0H13KGA)
- On one of our [community meetings](../contributing/index.md?h=meetings#meetings)

Please default to GitHub discussions: they work a lot like GitHub issues which
makes them easy to search.

### 2. Create an Issue

[Create a GEP issue](https://github.com/kubernetes-sigs/gateway-api/issues/new?assignees=&labels=kind%2Ffeature&template=enhancement.md) in the repo describing your change.
At this point, you should copy the outcome of any other conversations or documents
into this document.

### 3. `Provisional` - Agree on the Goals

Although it can be tempting to start writing out all the details of your
proposal, it's important to first ensure we all agree on the goals.

For API GEPs, the first version of your GEP should aim for a "Provisional"
status and leave out any implementation details, focusing primarily on
"Goals" and "Non-Goals", and documenting "Who" the GEP is for, "What" the
GEP will do, and "Why" it is needed. For this reason, the `Provisional`
state is also sometimes called the "Who/What/Why" stage.

For Memorandum GEPs, the first version of your GEP will be the only one, as
Memorandums have only a single stage - `Accepted`.

The `Provisional` state is different to other states (aside from `Memorandum`),
in that iteration on it can occur outside of the usual Gateway API release process.
To put this another way, until we have agreement on the "Who/What/Why",
then the PR does not fall into the regular release process.

GEPs entering the `Provisional` phase need the following to have occurred:

* A GEP PR using the template in GEP-696 merged into the `geps/` directory,
  describing the "Who", "What", and "Why" of the proposal, along with Goals
  and Non-Goals.

### 3. `Implementable` - Document Implementation Details

Now that everyone agrees on the goals, it is time to start writing out your
proposed implementation details. These implementation details should be very
thorough, including the proposed API spec, and covering any relevant edge cases.
Note that it may be helpful to use a shared doc for part of this phase to enable
faster iteration on potential designs.

It is likely that throughout this process, you will discuss a variety of
alternatives. Be sure to document all of these in the GEP, and why we decided
against them. At this stage, the GEP should be targeting the "Implementable"
stage.

For a GEP to enter the `Implementable` phase, there are some additional
requirements:

* One or more Gateway API maintainers must agree that the GEP is in-scope
  for the project.
* At least two (2) implementations must agree that they are interested in
  implementing the feature within six (6) months of it reaching `Experimental`.
  This is to ensure that there's community interest outside of the GEP owner.
* At least one "shepherd" who will help with navigating the GEP through the
  rest of the process. This shepherd can be any community member, but someone
  with experience of the GEP process will be the most helpful. The shepherd
  for that GEP will be responsible for initial review, as well as to be
  available to answer questions for the GEP owner about the process. Being
  a GEP Shepherd is a reasonably significant time commitment, with the
  time required scaling up sharply as a GEP becomes more complex and/or
  controversial. This shepherd should be recorded on the GEP issue.
* A GEP PR that updates the existing `Provisional` documentation with details
  that will be required to actually make the changes. This must include any
  API changes, as well as an initial set of test scenarios for implementations
  and conformance tests to target. Note that, at this stage, _only_ changes to 
  GEP document must be included.

### 4. `Experimental` - Make the API changes

With the GEP marked as "Implementable", it is time to actually make those
proposed changes in our API. In some cases, these changes will be documentation
only, but in most cases, some API changes will also be required. It is important
that every new feature of the API is marked as "Experimental" when it is
introduced. Within the API, we use `<gateway:experimental>` tags to denote
experimental fields. Within Golang packages (conformance tests, CLIs, e.t.c.) we
use the `experimental` Golang build tag to denote experimental functionality.

Some other requirements must be met before marking a GEP `Experimental`:

* Any API changes must be made to the Go types.
* The graduation criteria to reach `Standard` MUST be filled out. Note that this
  must include sufficient conformance testing at a minimum, and should include
  any other relevant criteria for that GEP.
* The GEP must have at least one Feature Name for features described inside that
  will need to be tested by conformance tests.
* A proposed probationary period (see next section) must be included in the GEP
  and approved by maintainers.

Before changes are released they MUST be documented. GEPs that have not been
both implemented and documented before a release cut off will be excluded from
the release.

#### Probationary Period

Any GEP in the `Experimental` phase is automatically under a "probationary
period" where it will come up for re-assessment if its graduation criteria are
not met within a given time period. GEPs that wish to move into `Experimental`
status MUST document a proposed period (6 months is the suggested default) that
MUST be approved by maintainers. Maintainers MAY select an alternative time
duration for a probationary period if deemed appropriate, and will document
their reasoning.

> **Rationale**: This probationary period exists to avoid GEPs getting "stale"
> and to provide guidance to implementations about how relevant features should
> be used, given that they are not guaranteed to become supported.

At the end of a probationary period if the GEP has not been able to resolve
its graduation criteria it will move to "Rejected" status. In extenuating
circumstances an extension of that period may be accepted by approval from
maintainers. GEPs which are `Rejected` in this way are removed from the
experimental CRDs and more or less put on hold. GEPs may be allowed to move back
into `Experimental` status from `Rejected` for another probationary period if a
new strategy for achieving their graduation criteria can be established. Any
such plan to take a GEP "off the shelf" must be reviewed and accepted by the
maintainers.

> **Warning**: It is **extremely important** that projects which implement
> `Experimental` features clearly document that these features may be removed in
> future releases.

### 5. `Standard` - Graduate the GEP

Once this feature has met the [graduation criteria](../concepts/versioning.md#graduation-criteria), it is
time to graduate it to the "Standard" channel of the API. Depending on the feature, this will usually
include one or more of the following:

* Graduating the resource to `v1`, and ensuring it is included in the Standard channel API Group and
  YAML install files.
* Graduating fields to "standard" by removing `<gateway:experimental>` tags.
* Graduating a concept to "standard" by updating documentation.

### 6. Close out the GEP issue

The GEP issue should only be closed once the feature has been:

* Moved to the standard channel for distribution (if necessary).
* Moved to a "v1" `apiVersion` for CRDs.
* Completely implemented and has wide acceptance (for process changes).

In short, the GEP issue should only be closed when the work is "done" (whatever
that means for that GEP).

## Format

GEPs should match the format of the template found in [GEP-696](gep-696/index.md).

## Out of scope

What is out of scope: see [text from KEP][kep-when-to-use]. Examples:

* Bug fixes
* Small changes (API validation, documentation, fixups). It is always
  possible that the reviewers will determine a "small" change ends up
  requiring a GEP.

## FAQ

#### Why is it named GEP?
To avoid potential confusion if people start following the cross references to
the full KEP process.

#### Why have a different process than mainline?
Gateway API has some differences with most upstream KEPs. Notably Gateway API
intentionally avoids including any implementation with the project, so this
process is focused entirely on the substance of the API. As this project is
based on CRDs it also has an entirely separately release process, and has
developed concepts like "release channels" that do not exist in upstream.

#### Is it ok to discuss using shared docs, scratch docs etc?
Yes, this can be a helpful intermediate step when iterating on design details.
It is important that all major feedback, discussions, and alternatives
considered in that step are represented in the GEP though. A key goal of GEPs is
to show why we made a decision and which alternatives were considered. If
separate docs are used, it's important that we can still see all relevant
context and decisions in the final GEP.

#### When should I mark a GEP as `Prototyping` as opposed to `Provisional`?
The `Prototyping` status carries the same base meaning as `Provisional` in that
consensus is not complete between stakeholders and we're not ready to move
toward releasing content yet. You should use `Prototyping` to indicate to your
fellow community members that we're in a state of active practical tests and
experiments which are intended to help us learn and iterate on the GEP. These
can include distributing content, but not under any release channel.

#### Should I implement support for `Experimental` channel features?
Ultimately one of the main ways to get something into `Standard` is for it to
mature through the `Experimental` phase, so we really _need_ people to implement
these features and provide feedback in order to have progress. That said, the
graduation of a feature past `Experimental` is not a forgone conclusion. Before
implementing an experimental feature, you should:

* Clearly document that support for the feature is experimental and may
  disappear in the future.
* Have a plan in place for how you would handle the removal of this feature from
  the API.

[kep]: https://github.com/kubernetes/enhancements
[kep-when-to-use]: https://github.com/kubernetes/enhancements/tree/master/keps#do-i-have-to-use-the-kep-process
