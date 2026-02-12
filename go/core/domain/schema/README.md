# schema

JSON Schema validation for contract YAML files. Compiles schemas from the
embedded contract filesystem at startup and validates configuration documents
against them.

## Key Types

- `Validator` compiles and caches JSON schemas, exposing `ValidateYAML` and `ValidateJSON` methods
- `SchemaType` is a string enum identifying which schema to validate against (repository, blueprints, etc.)
- `ValidationError` wraps schema violations with field path and description

## Patterns

- `FactoryValidator` provides a process-wide singleton compiled once via `sync.Once`
- Schemas are loaded from the embedded `contracts/core` filesystem at compilation time
- YAML documents are converted to JSON before validation since JSON Schema operates on JSON
- `SchemaType` constants map to file paths within the embedded contract filesystem
- Each schema is compiled and cached on first use within the `Validator` instance
- Validation returns a slice of `ValidationError` values rather than failing on the first error

## Internal Structure

| Path | Purpose |
|------|---------|
| `validator.go` | `Validator` struct, schema compilation, YAML and JSON validation |
| `factory_defaults.go` | `FactoryValidator` singleton with `sync.Once` initialization |

## Dependencies

- `contracts/core` for the embedded filesystem containing JSON schema definitions

## Role in System

This package guards the configuration boundary. When `config.LoadOptions` has
`ValidateSchemas` enabled, the config loader calls into this validator to
ensure that repository.yml, blueprints.yml, and other contract files conform
to their declared schemas before the rest of the system processes them.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
