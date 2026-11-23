# Prerelease Reports Directory

This directory contains _prerelease_ conformance reports for a given Gateway API version.

That is, these reports are from prerelease software versions, and are intended to show that
particular implementations _will_ support a particular Gateway API version when they release
their next actual release.

This allows the actual release to be tracked in the [reports](https://github.com/kubernetes-sigs/gateway-api/tree/main/conformance/reports/)
directory, where they can be consumed by automated processes such as the implementation list
builder, the matrix comparison builder, and the feature wizard, while still preserving a record
that a particular implementation passed tests for a Gateway API version before it was released.


## How this folder is structured

This folder stores all the conformance reports, for all the Gateway API Versions
and is structured as follows:

```text
|-- conformance/prerelease-reports
|   |-- v1.0
|   |   |-- acme-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v2.13-default-report.yaml
|   |-- v1.1
|   |   |-- acme-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v2.14-default-report.yaml
|   |   |   |-- standard-v2.14-with-the-lot-report.yaml
|   |   |   |-- experimental-v2.14-with-the-lot-report.yaml
|   |   |-- umbrella-operator
|   |   |   |-- README.md
|   |   |   |-- standard-v1.8-default-report.yaml
