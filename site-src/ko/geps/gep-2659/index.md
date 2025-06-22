# GEP-2659: Document and improve the GEP process

* Issue: [#2659](https://github.com/kubernetes-sigs/gateway-api/issues/2659)
* Type: Memorandum
* Status: Accepted

(See [status definitions](../overview.md#gep-states).)

## TLDR

This GEP clarifies some details about GEPs, and adds relationships and a new
status.


## Goals

- Enumerate how we should use RFC2119 language
- Add relationships between GEPs
- Add a metadata YAML schema for GEPs, including the new relationships
- Add a new status to cover some only recently noticed cases
- Update existing documentation outside this GEP to support the new material

## Introduction

As part of preparing for work to split up GEP-713, we (the Gateway API and GAMMA
maintainers) have noticed a few shortcomings in our GEP process.

In particular, we have some GEPs that are very different to other GEPs.
[GEP-1324](../gep-1324/index.md) is a good example,
as it lays out the general agreement and use cases for the GAMMA initiative, but
has no firm deliverables of its own. In fact, its main purpose was to ensure
that the community was in agreement on the language around and scope of the
problem of representing Mesh config using Gateway API primitives. Essentially,
it lays out a shared understanding of the problem space, as a basis for further
work (in the form of subsequent GEPs). However, our current GEP system of checklists
for graduating levels fits this type of GEP poorly.

Additionally, we've had two GEPs moved to Declined ([GEP-735: TCP and UDP address
matching](../gep-735/index.md) and
[GEP-1282: Describing Backend Properties](../gep-1282/index.md)).
In the case of GEP-1282, we now have a replacement GEP,
[GEP-1897: BackendTLSPolicy](../gep-1897/index.md)
which obsoletes the older GEP. But we have no way of representing this
relationship (or any other, in fact) between GEPs at the moment.

With these previous two changes, the addition of a metadata YAML file, similar
to what the KEP process uses, seems like it is increasingly necessary. This GEP
introduces a schema, which will be detailed in `.go` files rather than completely
in this document.

Lastly, we should clarify our use of [RFC2119](https://www.rfc-editor.org/rfc/rfc2119.txt)
language - we use MUST, SHOULD, MAY and so on as per that RFC in general,
but there is an extension in [RFC8174](https://www.rfc-editor.org/rfc/rfc8174.txt)
that adds that these words are only to be interpreted per the RFC when they are
in ALL CAPS, and "must" "should", "may" are to be interpreted in their usual
English meaning (which is not as strong as the RFC2119 one). This seems like
a _very_ good idea to adopt to me.

## Proposals

### Adopt the RFC8174 modification to RFC2119 language

RFC8174 clarifies that the reserved words MUST be in all-caps to have their
assigned meaning. This should make the spec clearer over time as we migrate.

### Addition of a new GEP status

We will introduce a new GEP status, `Memorandum`, that marks a GEP as recording
an agreement on either the definition of a problem, its scope
and a common language, or further process changes to the GEP process itself.

The defining characteristic here is that the GEP MUST NOT result in any changes
to the Gateway API spec, and MAY result in further GEPs to further clarify.
For GEPs that _do_ make changes to the API, but also require further GEPs to
clarify, they SHOULD use the new "Extended By" relationship instead.

Memorandum GEPs should be used sparingly, and should form the umbrella for a
significant amount of work, particularly work that may have parts that can
move through the GEP phases at different speeds.

The status is reached when a Memorandum GEP is merged, although as we will document
in the "Addition of GEP relationships" section, it can still be Extended
or Obsoleted.

Existing GEPs that meet this criteria will be gradually moved to be proper
Memorandum GEPs after this GEP is merged.

### Addition of YAML metadata file

The core Gateway API maintainers were hoping not to need metadata YAMLs for a
while, but the addition of relationships has turbocharged the need for machine
parseable GEP metadata.

This should also help with building display for GEPs; theoretically we can build
tooling that will let people slice the list of GEPs by whatever dimensions
we (or they) wish.

Similarly to the `ConformanceReport` object, I'm proposing to make CRD definitions
for this `GEPDetails` object, which will not be included in the usual CRD
definitions for Gateway API, nor will it be part of the regular API spec.

The use of a CRD is just to make the schema more similar to the other schemas
in this repository.

With that said, here's a rough sample of what a GEPMetadata object will look like:

```yaml
apiVersion: internal.gateway.networking.k8s.io/v1alpha1
kind: GEPDetails
number: 2659
name: Document and improve the GEP process
status: Memorandum
authors:
  - youngnick
relationships:
  # obsoletes indicates that a GEP makes the linked GEP obsolete, and completely
  # replaces that GEP. The obsoleted GEP MUST have its obsoletedBy field
  # set back to this GEP, and MUST be moved to Declined.
  obsoletes: {}
  obsoletedBy: {}
  # extends indicates that a GEP extends the linked GEP, adding more detail
  # or additional implementation. The extended GEP MUST have its extendedBy
  # field set back to this GEP.
  extends: {}
  extendedBy: {}
  # seeAlso indicates other GEPs that are relevant in some way without being
  # covered by an existing relationship.
  seeAlso: {}
# references is a list of hyperlinks to relevant external references.
# It's intended to be used for storing GitHub discussions, Google docs, etc.
references: {}
# featureNames is a list of the feature names introduced by the GEP, if there
# are any. This will allow us to track which feature was introduced by which GEP.
featureNames: {}
# changelog is a list of hyperlinks to PRs that make changes to the GEP, in
# ascending date order.
changelog:
  - "https://github.com/kubernetes-sigs/gateway-api/pull/2689"
```

### Addition of GEP relationships

As you can see in the previous section, this GEP adds three relationships between
GEPs:
- `Obsoletes` and its backreference `ObsoletedBy` - when a GEP is made obsolete
  by another GEP, and has its functionality completely replaced.
- `Extends` and its backreference `ExtendedBy` - when a GEP has additional details
  or implementation added in another GEP.
- `SeeAlso` - when a GEP is relevant to another GEP, but is not affected in any
  other defined way.

Each stanza in `relationships` includes a `number`, `name`, and a optional free-form
`description` field, which can be used to better describe the relationship.

At this time it's the updater's responsibility to ensure that both directions
are created for bidirectional relationships.

Further relationships may be added at a later date (at which time that GEP will
have an `Extends` relationship to this one).

Because of the addition of structured definitions for these relationships, the
relationships will _not_ be recorded in the main GEP file (it's anticipated
that the metadata will eventually be rendered in table form in the canonical
display on the website for general consumption, and PRs will need to create or
update a YAML file for each GEP change).
