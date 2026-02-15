# risk-catalog

Implements the `validate risk-catalog` command that validates OSCAL catalog documents against the OSCAL 1.1.3 JSON schema from NIST.

## Key Types

- **`Config`** -- Command configuration holding file path and workspace root

## Key Functions

- **`ValidateRiskCatalog()`** -- Entry point for the `validate risk-catalog` command
- **`parseConfig()`** -- Parse command-line arguments and validate file existence
- **`validateCatalog()`** -- Validate catalog content using the core OSCAL validator
- **`reportValidationResults()`** -- Print validation results with error/warning separation

## Patterns

- `init()` registration: registers command function with the global registry
- Core validator delegation: uses `core/validation/formats/oscal` for schema-based validation
- Error/warning separation: categorizes validation issues by severity for display
- Positional argument: takes file path as positional argument (not a flag)

## Internal Structure

| File | Responsibility |
| --- | --- |
| validate.go | OSCAL catalog validation command with config parsing, validation, and result reporting |

## Dependencies

- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration and workspace root
- `core/logging` -- structured logging
- `core/validation` -- validation error types and severity levels
- `core/validation/formats/oscal` -- OSCAL-specific validator (catalog type)

## Role in System

The `risk-catalog` sub-package validates OSCAL catalog documents (e.g., NIST 800-53) to ensure they conform to the OSCAL 1.1.3 schema. This is used in compliance workflows to verify that security control catalogs are structurally correct before they are referenced by profiles or used in risk assessments.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
