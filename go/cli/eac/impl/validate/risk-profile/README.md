# risk-profile

Implements the `validate risk-profile` command that validates OSCAL profile documents against the OSCAL 1.1.2 schema using go-oscal types.

## Key Types

- **`Config`** -- Command configuration holding file path and workspace root

## Key Functions

- **`ValidateRiskProfile()`** -- Entry point for the `validate risk-profile` command
- **`parseConfig()`** -- Parse command-line arguments and validate file existence
- **`validateProfile()`** -- Validate profile content using the core OSCAL validator
- **`reportValidationResults()`** -- Print validation results with error/warning separation

## Patterns

- `init()` registration: registers command function with the global registry
- Core validator delegation: uses `core/validation/formats/oscal` for schema-based validation
- Error/warning separation: categorizes validation issues by severity for display
- Warning-only pass: exit code 0 if only warnings exist (no errors)
- Positional argument: takes file path as positional argument (not a flag)

## Internal Structure

| File | Responsibility |
| --- | --- |
| validate-profile.go | OSCAL profile validation command with config parsing, validation, and result reporting |

## Dependencies

- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration and workspace root
- `core/logging` -- structured logging
- `core/validation` -- validation error types and severity levels
- `core/validation/formats/oscal` -- OSCAL-specific validator (profile type)

## Role in System

The `risk-profile` sub-package validates OSCAL profile documents to ensure they conform to the OSCAL schema. Profiles define which controls from a catalog are selected for a specific system, and validating them ensures the risk assessment pipeline has correct control selections.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
