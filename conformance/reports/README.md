# Conformance Reports

## How this folder is structured

This folder stores all the conformance reports, for all the Gateway API Versions
and is structured as follows:

```text
|-- conformance/reports
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
```

## Rules

As represented above, each Gateway API version contains a set of folders, one for
each conformant implementation. The implementation is the owner of its folder and
can upload as many reports as it want and a mandatory README.md structured as follows:

```md
# Acme operator

General information about the Acme/operator project

## Table of contents

| API channel | Implementation version | Mode | Report |
|-------------|------------------------|------|--------|
|             |                        |      |        |
|             |                        |      |        |
|             |                        |      |        |

## Reproduce

Instructions on how to reproduce the claimed report.
```

### Table of Contents

The table of contents is supposed to contain one row for each submitted report and
is structured as follows:

- **API channel**: the channel of the Gateway API (standard or experimental). It
  MUST correspond to the `gatewayAPIChannel` field of the related report.
- **Implementation version**: the link to the GitHub/website page related to the
  release. The release value MUST always be a semver and correspond to the
  `implementation.version` field of the report.
- **Mode**: the operating mode of the implementation (the default is `standard`).
  It MUST correspond to the `mode` field specified in the related report.
  In case a mode different from `default` is declared in a report, the implementation
  needs to add information in the "reproduce" section (see below) on how to configure
  such a mode.
- **Report**: the link to the related report. It MUST be in the form of
  `[link](./report.yaml)`. The reports must be named according to the following
  pattern: `<API Channel>-<Implementation version>-<mode>-report.yaml`.

### Reproduce

The "Reproduce" sections MUST exist and contain the manual or automatic steps
to reproduce the results claimed by the uploaded conformance reports. In case
different implementation versions have different reproduction steps, this section
can have multiple sub-sections, each related to a specific or a subset of implementation
versions.

## Rules exceptions

This structure was introduced after the Gateway API release v1.0.0, while we started
gathering and storing conformance reports since Gateway API v0.7.1. Even if this
structure has been applied to all the Gateway API versions, it has been done in
a best-effort manner, without forcing the rules described above retroactively.
This is inevitable also because the `mode` and `gatewayAPIChannel` fields of the
reports have been introduced after v1.0.0. For these reasons, all the folders related
to Gateway API v1.0.0 or prior benefits from the following rules' exceptions:

- The fields `API channel` and `Mode` of the README.md's Table of Contents are replaced
  by an "x".
- The reports are named according to the pattern `<Implementation version>-report.yaml`.
- The implementation version can be different from a semver (for example, they can
  be commit hashes or pull request links).
- The "Reproduce" section is optional; implementations can decide whether to provide
  reproduction steps for those versions.

## Reports

The reports MUST be uploaded exactly as they have been created by the conformance
suite, without any modifications. The "Reproduce" section allows checking
any diff between the claimed report and the actual one.

### Reports rules

To be accepted, the reports uploaded **after** the Gateway API v1.1.0 release must
comply with the following rules:

- All the `implementation`-related fields must have a meaningful value assigned,
  in detail:
  - `organization`: can be an open source organization, an individual, a company,
    etc. Organizations can have multiple projects.
  - `project`: the name of the project. It must be unique within the same organization.
  - `contact`: it indicates the GitHub usernames of those who are responsible for
    maintaining this file, so they can be easily contacted when needed (e.g. for
    relevant release announcements regarding conformance, etc.). Optionally, it
    can be an email address or a support URL (e.g. Github new issue page).
  - `url`: it must be a valid url for a GitHub repository, or any other website with
    information related to the project.
  - `version`: it must be a snapshot of the project, which means it can be a commit,
    a tag, a release, or any other means of specifying a version. It can't be a
    branch name.
- `gatewayAPIVersion`: it must contain the version of the Gateway API used to run
  the conformance tests.
- `gatewayAPIChannel`: it must contain the channel of the Gateway API used to run
  the conformance tests (standard/experimental).
- `mode`: it is the mode of the implementation used to run the tests. If different
  from `default`, it must correspond to a specific setup of the implementation.
- Test result: in order to have a report valid to be accepted, all the profiles
  need to have the `result` fields (core and extended) set to `success` or `partial`.
  `partial` reports will be displayed separately from the `success` ones in the
  [implementations table][implementations-table].

  A `success` report implies **all** the following:
  - all the core conformance tests have been successfully run
  - all the tests related to the supported extended features have been successfully
    run.
  
  A `partial` report implies **at least one** of the following:
  - some core conformance tests have been skipped
  - some tests related to the supported extended features have been skipped
  - the conformance test run required some steps unexpected by the suite.

[implementations-table]: https://gateway-api.sigs.k8s.io/implementations/v1.1/
