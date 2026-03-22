# Validate control-tags

<!-- book:cmd validate control-tags -->

Validates that all `@control:` and `@controls:` tags in Gherkin feature files reference valid control IDs from the OSCAL catalog. Ensures evidence collection can find controls during risk assessment.

## Usage

```bash
eac validate control-tags
```

## What It Checks

- Discovers all `.feature` files under the specs root directory.
- Extracts `@control:<id>` and `@controls:<id1>,<id2>` tags from every feature file.
- Loads the OSCAL catalog from `templates/specs/risk-catalog/`.
- Verifies each control ID exists in the catalog (case-insensitive).

## Examples

```bash
eac validate control-tags
```

## Common Errors

- **OSCAL catalog not found** -- The catalog file does not exist at the expected path. Run `eac templates install` to create it.
- **Control 'xx-99' not found in catalog** -- A feature file references a control ID that does not exist. Check the ID format (e.g., `ac-2`, `au-3`, `ia-5(1)`) and add missing controls to the catalog if needed.

## See Also

- [validate](./validate.md)
- [validate risk-profile](./risk-profile.md)
- [validate Commands](../../categories/validate.md)
