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

| File/Sub-package | Responsibility |
| --- | --- |
| validation.go | `ValidationError`, `Validator` interface, `AIExecutor` interface |
| error_codes.go | `ErrorCode` type and full registry of all error codes |
| error_formatter.go | `ErrorFormatter` for AI-friendly error messages |
| formats/gherkin/ | Gherkin feature file validation |
| formats/json/ | JSON schema validation |
| formats/oscal/ | OSCAL catalog and profile validation |
| formats/structurizr/ | Structurizr DSL validation |

## Dependencies

_No internal repository imports in the root package (leaf package)._

## Role in System

The `validation` package defines the shared error vocabulary and validator
contract used by all AI-generation and content-validation flows in `core`.
Format-specific validators under `formats/` implement the `Validator` interface
and reference error codes from this package.

## Code Health

### Tech Debt
- `formats/gherkin/validator.go`: `Validate()` is ~200 lines (line 57-260+); consider splitting structural vs semantic checks
- `error_codes.go`: 1000-line file dominated by static declarations; consider code-generating the registry
- `formats/structurizr/docker.go:29`: global mutable `defaultContainerProvider` set via `SetContainerProvider()`; not concurrency-safe

### Pain Points
- No unit tests in the root `validation/` package (error_codes.go, error_formatter.go, validation.go are untested directly)
- Four separate interfaces (`Validator`, `AIExecutor`, `AIExecutorWithProviderInfo`, structurizr `Validator`) may confuse consumers

### Optimization Opportunities
- Extract error code registry into a generated file to reduce manual maintenance (medium effort, high value)
- Add table-driven tests for `ErrorFormatter` formatting functions (low effort, high value)
