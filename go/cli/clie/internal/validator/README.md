# validator

Command validation against business rules and configuration validation against embedded JSON Schema.

## Key Types

- **`CommandValidator`** -- Validates parsed commands against EBNF schema and business rules
- **`CommandValidationResult`** -- Validation result with valid flag, errors, warnings, and parsed command
- **`EmbeddedValidator`** -- JSON Schema validator using gojsonschema with embedded contract
- **`ValidationResult`** -- Schema validation result with typed errors and warnings
- **`ValidationError`** -- Single validation error with field, rule, message, value, expected
- **`Severity`** -- Error severity enum: Error, Warning, Info
- **`PatternValidator`** -- Utility type exposing compiled regex pattern checks

## Key Functions

- **`ValidateCommand`** -- Full command validation: binary name, flags, subcommand, extension name
- **`ValidateForRun`** -- Run-specific validation requiring extension name
- **`ValidateForMetadata`** -- Metadata-specific validation requiring extension name
- **`NewEmbeddedValidator`** -- Creates JSON Schema validator from embedded contract
- **`ValidateJSON` / `ValidateInterface`** -- Validates config data against the schema
- **`GetEmbeddedSchemaVersion`** -- Returns embedded schema version string

## Patterns

- Embedded contract: JSON Schema loaded from `contracts/clie/0.1.0` at init via `clie.FS.ReadFile`
- Two validation layers: CommandValidator for CLI argument structure, EmbeddedValidator for YAML config content
- Compiled patterns: Regex patterns for extension names, env vars, versions, resources pre-compiled as package vars
- Enum validation: `IsInEnum` helper for checking values against valid option lists

## Internal Structure

| File                  | Responsibility                                                       |
| --------------------- | -------------------------------------------------------------------- |
| command-validator.go  | CommandValidator, ValidateCommand, ValidateForRun/Metadata           |
| embedded_validator.go | EmbeddedValidator, JSON Schema validation, schema version extraction |
| errors.go             | ValidationError, ValidationResult, rule constants, Severity enum     |
| patterns.go           | Compiled regex patterns, enum value lists, PatternValidator utility  |

## Dependencies

- `internal/command-parser` -- Parser for command argument structure
- `contracts/clie/0.1.0` -- Embedded JSON Schema for config validation

## Role in System

The validator package serves two distinct validation needs. The CommandValidator validates CLI invocations (used in PersistentPreRunE to catch malformed commands early). The EmbeddedValidator validates `.clie/clie.yml` configuration files against the JSON Schema contract (used by the `clie validate` command). The patterns and enums are also referenced by other packages for consistent validation.

## Code Health

### Tech Debt

- None identified.

### Pain Points

- None identified.

### Optimization Opportunities

- None identified.
