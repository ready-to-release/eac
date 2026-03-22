# Validate risk-profile

<!-- book:cmd validate risk-profile -->

Validates an OSCAL profile document against the OSCAL 1.1.2 schema using go-oscal types.

## Usage

```bash
eac validate risk-profile <file>
```

## Arguments

| Argument | Description |
|----------|-------------|
| `file` | Path to the OSCAL profile JSON file (required) |

## What It Checks

- JSON parsing with go-oscal types.
- Required field presence (UUID, metadata, imports).
- Control ID format validation.
- Import href validation.

## Examples

```bash
eac validate risk-profile templates/specs/risk-catalog/profile.json
```

## Common Errors

- **file not found** -- The specified path does not exist.
- **missing required field** -- A required OSCAL field (UUID, metadata, imports) is absent.
- **invalid control ID format** -- A control ID does not match the expected pattern (e.g., `ac-2`).

## See Also

- [validate](./validate.md)
- [create risk-profile](../create/risk-profile.md)
- [validate risk-catalog](./risk-catalog.md)
- [validate Commands](../../categories/validate.md)
