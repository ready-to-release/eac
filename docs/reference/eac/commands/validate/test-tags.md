# Validate test-tags

<!-- book:cmd validate test-tags -->

Validates that all tags used in Gherkin feature files are defined in the tag contract (`.eac/testing-tags.yml`). Prevents undefined tags that would be silently ignored by godog.

## Usage

```bash
eac validate test-tags
```

## What It Checks

- Discovers all `.feature` files under the specs root.
- Extracts all `@` tags from features, scenarios, and example tables.
- Checks each tag against the tag contract definitions.
- Validates pattern tags with additional rules:
    - `@skip:<reason>` -- reason must be a defined skip reason.
    - `@deps:<name>` -- name must be a registered tool or OS platform (`linux`, `macos`, `windows`).
    - `@env:<moniker>` -- moniker must be a defined environment.
    - `@depm:<module>` -- module must be a defined module.

## Examples

```bash
eac validate test-tags
```

## Common Errors

- **Undefined tag** -- A tag is used but not defined in `.eac/testing-tags.yml`. Add the tag definition to the contract.
- **Invalid skip reason** -- A `@skip:` tag uses a reason not listed in the contract's `skip_reasons`.
- **Invalid deps name** -- A `@deps:` tag references a tool not in the tool registry.

## See Also

- [validate](./validate.md)
- [validate Commands](../../categories/validate.md)
