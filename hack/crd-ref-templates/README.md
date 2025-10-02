The templates contained in this directory were copied from https://github.com/elastic/crd-ref-docs/tree/master/templates/markdown
and had the following modifications:

- `type.tpl` - Adds an "experimental" marker in case the API field is marked as "experimental"
- `type_members.tpl` - Removes the text between tags `<gateway:util:excludeFromCRD></gateway:util:excludeFromCRD>`