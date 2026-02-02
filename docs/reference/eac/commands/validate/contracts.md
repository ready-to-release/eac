# Validate contracts

<!-- book:cmd validate contracts -->

## How It Works

Validates repository contract files against JSON schemas:

- **Schema Validation**: Validates `.eac/repository.yml` against defined schema
- **Module Contracts**: Verifies all module definitions are valid
- **Required Fields**: Ensures all mandatory fields are present
- **Type Checking**: Validates field types and values
- **Reference Integrity**: Checks cross-references between contracts

Ensures repository configuration is valid before builds or releases.

## See Also

- [validate](./validate.md)
- [validate Commands](../categories/validate.md)
