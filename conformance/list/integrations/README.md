# ImplementationsDetails files

This directory contains ImplementationsDetails for Gateway API integrations.

These are YAML files that meet the ImplementationDetails spec as defined in
`conformance/apis/v1/implistdetails.go` and other files in that directory.
A sample is included below.

As there is no conformance for Integrations, this file is the only requirement for
being included, aside from the requirements documented in the implementations list
page.

The `description` field will be included verbatim (with escaping) in the generated
implementations list file.

```yaml
apiVersion: gateway.networking.kubernetes.io/v1
kind: ImplementationListDetails
organization: organizationName # Usually the Github organization for the project
project: projectName # Usually the Github repo name
fullName: Project Full Name # How you want your project represented in the implementations list
description: |
  A description to be included in the generated implementations list page.

  Note that this needs to be valid with any version of Gateway API, and must abide by the
  overall Kubernetes third-party linking policy, and so MUST NOT include:

  * Gateway API specific version information support
  * Information about commercial arrangements. Links to the project site are okay, links to
    "buy commmercial support" or similar are not.
```
