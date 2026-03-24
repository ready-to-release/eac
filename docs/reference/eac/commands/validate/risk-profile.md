# validate risk-profile

<!-- book:cmd validate risk-profile -->

## What It Checks

- JSON parsing with go-oscal types.
- Required field presence (UUID, metadata, imports).
- Control ID format validation.
- Import href validation.

## Common Errors

- **file not found** -- The specified path does not exist.
- **missing required field** -- A required OSCAL field (UUID, metadata, imports) is absent.
- **invalid control ID format** -- A control ID does not match the expected pattern (e.g., `ac-2`).

## See Also

- [validate](./validate.md)
- [create risk-profile](../create/risk-profile.md)
- [validate risk-catalog](./risk-catalog.md)
- [validate Commands](../validate/index.md)
