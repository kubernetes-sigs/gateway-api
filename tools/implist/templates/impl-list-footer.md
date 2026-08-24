## Adding new entries

This page is automatically generated; please do not make PRs to this page.

### For implementations

Implementations wanting to add themselves must:

* Add a conformance report that's at least partially conformant to `conformance/reports`.
* Add an ImplementationDetails YAML file to `conformance/list/implementations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make generate` in the top level of the repository, and
the implementations list generation code will update the page for you. Include
the updated page in your PR.

This process replaces an older, maintainer-performed process.

### For integrations

Integrations wanting to add themselves must:

* Add an ImplementationDetails YAML file to `conformance/list/integrations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make generate` in the top level of the repository, and
the implementations list generation code will update the page for you. Include
the updated page in your PR.
