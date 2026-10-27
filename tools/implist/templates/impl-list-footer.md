## Adding new entries

This page is automatically generated; please do not make PRs to this page.

### For implementations

Implementations wanting to add themselves must:

* Add a conformance report to `conformance/reports`.
* Add an ImplementationDetails YAML file to `conformance/list/implementations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make update-implist` in the top level of the repository, and
the implementations list generation code will update the page for you. You can see what
the page will look like using the live docs preview with `make live-docs` command.

If your implementation does not appear, check the output of `make update-implist`, and it
will indicate what is wrong. Note that all conformance decisions are now handled by this
tool, and changes to conformance require updating the tool.

This process replaces an older, maintainer-performed process.

### For integrations

Integrations wanting to add themselves must:

* Add an ImplementationDetails YAML file to `conformance/list/integrations`. See the
  `README.md` file in that directory for more.

Once the PR is ready, run `make update-implist` in the top level of the repository, and
the implementations list generation code will update the page for you. You can see what
the page will look like using the live docs preview with `make live-docs` command.
