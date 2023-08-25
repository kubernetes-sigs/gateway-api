# Contributor Ladder

Within the Kubernetes community, the concept of a contributor ladder has been
developed to define how individuals can earn formal roles within the project.
The Gateway API contributor ladder largely follows the [roles defined by the
broader Kubernetes
community](https://github.com/kubernetes/community/blob/master/community-membership.md),
though there are some aspects that are unique to this community.

## Goals

We hope that this doc will provide an initial step towards the following goals:

* Ensure the long term health of the Gateway API community
* Encourage new contributors to work towards formal roles and responsibilities
  in the project
* Clearly define the path towards leadership roles
* Develop a strong leadership pipeline so we have great candidates to fill
  project leadership roles


## Scope

The following repositories are covered by this doc:

* [kubernetes-sigs/gateway-api](https://github.com/kubernetes-sigs/gateway-api)
* [kubernetes-sigs/ingress2gateway](https://github.com/kubernetes-sigs/ingress2gateway)
* kubernetes-sigs/blixt (once migration is complete)

Within each of these projects, there are opportunities to become an approver or
reviewer for either the entire project, or a subset of that project. For
example, you could become a reviewer or approver focused on just docs, GEPs, API
changes, or conformance tests.

## General Guidelines

### 1. Everyone is welcome

We appreciate all contributions. You don’t need to have a formal role in the project to make or review pull requests, and help with issues or discussions. Accepting a formal role within the project is entirely optional.

### 2. These roles require continued contributions

Applying for one of the roles defined above should only be done if you intend to
continue to contribute at a level that would merit that role. If for any reason
you are unable to continue in one of the roles above, please resign. Members
with an extended period away from the project with no activity will be removed
from the Kubernetes GitHub Organizations and will be required to go through the
org membership process again after re-familiarizing themselves with the current
state.

### 3. Don’t merge without consensus

If you have reason to believe that a change may be contentious, please wait for
additional perspectives from others before merging any PRs. Even if you have
access to merge a PR, it doesn’t mean you should. Although we can’t have PRs
blocked indefinitely, we need to make sure everyone has had a chance to present
their perspective.

### 4. Start a discussion

If you’re interested in working towards one of these roles, please reach out to
a Gateway API maintainer on Slack.

## Contributor Ladder

The Gateway API contributor ladder has the following steps:

1. Member
2. Reviewer
3. Approver
4. Maintainer

This is also a GAMMA-specific leadership role that does not fit as cleanly on
this ladder. All of these roles will be defined in more detail below.

## Member, Reviewer, and Approver

The first steps on the contributor ladder are already [clearly defined in the
upstream Kubernetes
Community](https://github.com/kubernetes/community/blob/master/community-membership.md#community-membership).
Gateway API follows those guidelines along with the rest of the Kubernetes
community. Within Gateway API, there are a variety of areas one can become a
reviewer or approver, this includes:

* Conformance
* Documentation
* GEPs
* Webhook Validation

## Maintainers and GAMMA Leads

The final steps on the contributor ladder represent large overall leadership
roles within the project as a whole. The spaces available for these roles are
limited (generally 3-4 people in each role is ideal). Wherever possible, we try
to ensure that different companies are represented in these roles.

### Maintainers

Gateway API Maintainers are known as [Subproject
Owners](https://github.com/kubernetes/community/blob/master/community-membership.md#subproject-owner)
within the Kubernetes community. To become a Gateway API Maintainer, the most
important things we expect are:

* Long term, sustained contributions to Gateway API for at least 6 months
* Deep understanding of technical goals and direction of the project
* Successfully authored and led significant enhancement proposals
* Approver for at least 3 months
* Ability to lead community meetings

In addition to all of the expectations described above, we expect maintainers to
set the technical direction and goals for the project. This role is critical to
the health of the project, maintainers should mentor new approvers and
reviewers, and ensure that there are healthy processes in place for discussion
and decision making. Finally, maintainers are ultimately responsible for
releasing new versions of the API.

## GAMMA Leads

The concept of GAMMA Leads does not have a perfect parallel on the upstream
Kubernetes community ladder. They are essentially Subproject Owners, but for the
GAMMA initiative, which is a major initiative within Gateway API.

To become a GAMMA lead, the most important thing we expect are:

* Significant experience with Service Mesh implementation(s)
* Deep understanding of technical goals and direction of the project
* Long term, sustained contributions to the GAMMA initiative for at least 6 months
* Ability to lead community meetings

In addition to all of the expectations described above, we expect GAMMA Leads to
set the technical direction and goals for the GAMMA initiative. They should
ensure that there are healthy processes in place for discussion and decision
making and that the release goals and milestones are clearly defined.
