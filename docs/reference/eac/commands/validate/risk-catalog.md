# validate risk-catalog

<!-- book:cmd validate risk-catalog -->

## What It Checks

- Valid JSON structure parseable as an OSCAL catalog.
- Required fields are present (UUID, metadata, groups/controls).
- Schema compliance against the official OSCAL 1.1.3 catalog schema.

## Common Errors

- **file not found** -- The specified path does not exist.
- **OSCAL schema violation** -- Missing required fields or invalid structure per the OSCAL 1.1.3 specification.

## See Also

- [validate](./validate.md)
- [validate risk-profile](./risk-profile.md)
- [validate Commands](../validate/index.md)
