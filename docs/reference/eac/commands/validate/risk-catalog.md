# Validate risk-catalog

<!-- book:cmd validate risk-catalog -->

Validates an OSCAL catalog document against the official OSCAL 1.1.3 JSON schema from NIST.

## Usage

```bash
eac validate risk-catalog <file>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `file` | Path to the OSCAL catalog JSON file (required) |

## What It Checks

- Valid JSON structure parseable as an OSCAL catalog.
- Required fields are present (UUID, metadata, groups/controls).
- Schema compliance against the official OSCAL 1.1.3 catalog schema.

## Examples

```bash
eac validate risk-catalog templates/specs/risk-catalog/controls.catalog.json
```

## Common Errors

- **file not found** -- The specified path does not exist.
- **OSCAL schema violation** -- Missing required fields or invalid structure per the OSCAL 1.1.3 specification.

## See Also

- [validate](./validate.md)
- [validate risk-profile](./risk-profile.md)
- [validate Commands](../categories/validate.md)
