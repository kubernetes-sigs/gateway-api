# Gateway Enhancement Proposal (GEP)

Gateway Enhancement Proposals (GEPs) will serve a similar purpose to
the [KEP][kep] process for the main Kubernetes project:

1. Ensure that changes to the API follow a known process and discussion
  in the OSS community.
1. Make changes and proposals discoverable (current and future).
1. Document design ideas, tradeoffs, decisions that were made for
  historical reference.

As this is still a relatively new project, we don't wish to encumber
proposals with a large amount of boilerplate process. The most
important part of the GEP are the last two items: making sure that we
can easily find previous discussions and alternatives considered.

## Process

### New GEP

* File an issue that will be used to track the GEP progress. Use the existing
  issue templates. Use this number for naming the GEP (i.e. issue #123 would be
  named `gep-123.md`).
* Discussion and comments occur on the issue and shared docs. Any
  relevant shared documents should be linked off of the Github issue.
  * Create a GEP document in `site-srcs/gep/gep-<issue
    #>-short-description.md` with status `proposed`; the sections that
    MUST be kept up to date are `TLDR` and `References`. Other
    sections can refer to the issue or shared doc temporarily.
  * We should try to resolve most of the discussion in this phase as doing the
    discussion on a GEP is not very easy (at the moment) from a workflow
    perspective.
* When there is consensus on the approach, consolidate any content,
  links into the document (see template format below). Submit the
  PR for merging.
* Propagate iterations on the GEP proposal from the issue/shared docs
  to the GEP.
* When there are no further comments, update GEP with status
  "accepted". This means that we are committing to implementing/have
  implemented the feature.
  * Make sure any shared docs are going to be in a format that is accessible in
    the future. If possible, copy the contents of the document into the GEP
    along with comments.
* Set the GEP status (accepted, implemented etc).
* Tracking issue should be marked as closed.

## Format

GEPs should match the format of the template found in [GEP-696](/geps/gep-696).

## Out of scope

What is out of scope: see [text from KEP][kep-when-to-use]. Examples:

* Bug fixes
* Small changes (API validation, documentation, fixups). It always
  possible that the reviewers will determine a "small" change ends up
  requiring a GEP.

## How much additional work is this?

* Need to copy out material into a `.md` document for future
  reference and keep it up to date.
* Need to keep GEP up to date with current references to shared doc.
* Some amount of overhead merging GEPs into the repo.

## FAQ

* Q: Why is it named GEP?
  * A: To avoid potential confusion if people start following the cross
    references to the full KEP process.
* Q: Why have a different process than mainline?
  * A: We would like to keep the machinery to an absolute minimum for now --
    this will likely change as we move to v1.
* Q: Is it ok to discuss using shared docs, scratch docs etc?
  * A: Yes! We view GEPs as primarily historical record to ensure that any
    artifacts are preserved for future discussions.

[kep]: https://github.com/kubernetes/enhancements
[kep-when-to-use]: https://github.com/kubernetes/enhancements/tree/master/keps#do-i-have-to-use-the-kep-process
