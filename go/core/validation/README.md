# validation

Provides structured validation types, error codes, and formatting utilities
used across all format-specific validators.

## Key Types

- **`ValidationError`** -- Structured error with code, message, and line number
- **`ErrorCode`** -- Categorized error code with severity and retriability
- **`Validator`** -- Interface for format-specific output validation
- **`ErrorFormatter`** -- Builds AI-friendly error messages with context
- **`AIExecutor`** -- Abstraction for AI execution in generation pipelines
- **`NoOpValidator`** -- Pass-through validator for unvalidated formats

## Patterns

- Error code registry: centralized lookup of all error codes by string key
- Category/severity classification: errors grouped by structure, semantic, format, constraint
- Retriability metadata: each error code declares if AI retry may fix it

## Internal Structure

| File/Sub-package     | Responsibility                                                   |
| -------------------- | ---------------------------------------------------------------- |
| validation.go        | `ValidationError`, `Validator` interface, `AIExecutor` interface |
| error_codes.go       | `ErrorCode` type and full registry of all error codes            |
| error_formatter.go   | `ErrorFormatter` for AI-friendly error messages                  |
| formats/gherkin/     | Gherkin feature file validation                                  |
| formats/json/        | JSON schema validation                                           |
| formats/oscal/       | OSCAL catalog and profile validation                             |
| formats/structurizr/ | Structurizr DSL validation                                       |

## Dependencies

_No internal repository imports in the root package (leaf package)._

## Role in System

The `validation` package defines the shared error vocabulary and validator
contract used by all AI-generation and content-validation flows in `core`.
Format-specific validators under `formats/` implement the `Validator` interface
and reference error codes from this package.

## Code Health

### Tech Debt
- `error_codes.go` is 1000 lines, dominated by static error code declarations
- `formats/gherkin/validator.go` is 405 lines
- `formats/gherkin/validator_tags.go` is 306 lines
- `formats/structurizr/quick.go` is 321 lines
- `formats/structurizr/docker.go` is 289 lines with global mutable `defaultContainerProvider` set via `SetContainerProvider()`; not concurrency-safe
- `formats/oscal/profile.go` is 299 lines
- `formats/json/validator.go` is 266 lines
- `error_formatter.go` is 224 lines

### Pain Points
- Four separate interfaces (`Validator`, `AIExecutor`, `AIExecutorWithProviderInfo`, structurizr `Validator`) may confuse consumers

### Optimization Opportunities
- Extract error code registry into a generated file to reduce manual maintenance
- Consider splitting large validator files by validation category
