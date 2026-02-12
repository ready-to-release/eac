# validation/formats/json

JSON schema validation utilities, migrated from `domain.JSONSchemaValidator`.

## Key Types

| Type        | Purpose                                                        |
| ----------- | -------------------------------------------------------------- |
| `Validator` | Validates JSON documents against a compiled JSON Schema        |

## Key Functions

| Function       | Purpose                                              |
| -------------- | ---------------------------------------------------- |
| `NewValidator` | Loads and compiles a JSON Schema, returns a Validator |

## Patterns

- Schema is compiled once in `NewValidator`; subsequent `Validate` calls are fast
- Array error pattern detection collapses repetitive per-item errors into a single contextual message
- Enhanced error hints for type mismatches, enum violations, and missing required fields
- Implements `validation.Validator` interface

## Internal Structure

| File               | Purpose                                                          |
| ------------------ | ---------------------------------------------------------------- |
| `validator.go`     | `Validator` type with `Validate`, `VerifyImplementation`, helpers |
| `validator_test.go`| Tests for schema loading, nil-schema guard, and URL generation    |

## Dependencies

| Package            | Purpose                                         |
| ------------------ | ----------------------------------------------- |
| `core/validation`  | `ValidationError` type and error codes           |
| `gojsonschema`     | JSON Schema compilation and validation engine    |

## Role in System

This is the canonical location for JSON schema validation within the
`validation/formats` hierarchy. It provides the same functionality previously
available only through `domain.JSONSchemaValidator`, consolidated here for
consistency with other format validators (Gherkin, etc.).

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
