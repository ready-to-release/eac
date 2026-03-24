# validate test-tags

<!-- book:cmd validate test-tags -->

## What It Checks

- Discovers all `.feature` files under the specs root.
- Extracts all `@` tags from features, scenarios, and example tables.
- Checks each tag against the tag contract definitions.
- Validates pattern tags with additional rules:
  - `@skip:<reason>` -- reason must be a defined skip reason.
  - `@deps:<name>` -- name must be a registered tool or OS platform (`linux`, `macos`, `windows`).
  - `@env:<moniker>` -- moniker must be a defined environment.
  - `@depm:<module>` -- module must be a defined module.

## Common Errors

- **Undefined tag** -- A tag is used but not defined in `.eac/testing-tags.yml`. Add the tag definition to the contract.
- **Invalid skip reason** -- A `@skip:` tag uses a reason not listed in the contract's `skip_reasons`.
- **Invalid deps name** -- A `@deps:` tag references a tool not in the tool registry.

## See Also

- [validate](./validate.md)
- [validate Commands](../validate/index.md)
