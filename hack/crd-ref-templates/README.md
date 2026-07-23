The templates contained in this directory were copied from https://github.com/elastic/crd-ref-docs/tree/master/templates/markdown
and had the following modifications:

- `type.tpl` - Adds an "experimental" marker in case the API field is marked as "experimental"
- `type.tpl` - Adds a "Union Feature" marker, linking to the related type(s), in case a type or field's doc comment
  contains one or more `<gateway:union:FeatureName>` tags (e.g. `<gateway:union:BackendTLSPolicy>`). This mirrors the
  "experimental" marker above and is used to cross-reference [union features](/guides/implementers-guide/#union-feature-conformance)
  with the resources/fields they are expected to interoperate with.
- `type_members.tpl` - Removes the text between tags `<gateway:util:excludeFromCRD></gateway:util:excludeFromCRD>`