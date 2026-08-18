## Adding new entries

Implementations are free to make a PR to add their entry to this page; however,
in order to meet the requirements for being Partially Conformant or Conformant,
the implementation must have had a conformance report submission PR merged.

Part of the review process for new additions to this page is that a maintainer
will check the conformance level and verify the state.

## Page Review Policy

This page is intended to showcase actively developed and conformant implementations
of Gateway API, and so is subject to regular reviews.

These reviews are performed at least one month after every Gateway API release
(starting with the Gateway API v1.3 release).

As part of the review, a maintainer will check:

* which implementations are **Conformant** - as defined above in this document.
* which implementations are **Partially Conformant**, as defined above in this
  document.

If the maintainer performing the review finds that there are implementations
that no longer satisfy the criteria for Partially Conformant or Conformant, or
finds implementations that are in the "Stale" state, then that maintainer will:

* Inform the other maintainers and get their agreement on the list of stale and
to-be-removed implementations
* Open a draft PR with the changes to this page.
* Post on the #sig-network-gateway-api channel informing the maintainers of
implementations that are no longer at least partially conformant should contact
the Gateway API maintainers to discuss the implementation's status. This period
is called the "**right-of-reply**" period, is at least two weeks long, and functions
as a lazy consensus period.
* Any implementations that do not respond within the right-of-reply period will be
downgraded in status, either by being moved to "Stale", or being removed
from this page if they are already "Stale".

Page review timeline, starting with the v1.4 Page Review:

* Gateway API v1.4 release Page Review (at least one month after the actual
  release): a maintainer will move anyone who hasn't submitted a conformance
  report using the rules above to "Stale". They will also contact anyone who
  moves to Stale to inform them about this rule change.
* Gateway API v1.5 release Page Review (at least one month after the actual
  release): A maintainer will perform the Page Review process again, removing
  any implementations that are still Stale (after a right-of-reply period).
* Gateway API v1.6 release Page Review (at least one month after the actual
  release): We will remove the Stale category, and implementation maintainers
  will need to be at least partially conformant on each review, or during the
  right-of-reply period, or be removed from the implementations page. **You are here**

This means that, after the Gateway API v1.6 release, implementations cannot be
added to this page unless they have submitted at least a Partially Conformant
conformance report.

