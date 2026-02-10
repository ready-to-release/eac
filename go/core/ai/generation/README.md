# ai/generation

Provides the AI generation pipeline with format-aware structured output, validation-driven retry logic, and configurable retry strategies. This is the core engine that all AI generation commands use to produce validated output in JSON, Gherkin, OSCAL, Structurizr DSL, or plain text formats.

## Key Types

| Type | Purpose |
|------|---------|
| `StructuredGenerator` | Orchestrates AI generation with format-specific instructions, validation, and retry loops |
| `GenerationResult` | Contains generated content, attempt count, validity status, and validation errors |
| `RetryConfig` | Configures retry behavior including output format, validator, executor, strategy, and max attempts |
| `RetryResult` | Result from `GenerateWithRetry` with output, validation errors, attempt count, and provider info |
| `RetryStrategy` | Interface for retry decision-making and prompt construction on validation failure |
| `StandardStrategy` | Retries on any retriable error with full error context |
| `FocusedStrategy` | Retries only when errors match specified focus categories |
| `EscalatingStrategy` | Wraps another strategy, adding urgency and detailed error grouping on later attempts |
| `StructuredFormat` | Enum of supported output formats (`json`, `gherkin`, `oscal-catalog`, `oscal-profile`, `structurizr`, `plaintext`) |
| `PhaseConfig` | Defines generation behavior for a format (format type, instruction, validator) |

## Key Functions

| Function | Purpose |
|----------|---------|
| `GenerateWithRetry` | Entry point for AI generation with retry; loads validator, runs generator, returns result |
| `BuildRetryConfig` | Factory function creating `RetryConfig` from command-level inputs with functional options |
| `GetRetryStrategy` | Creates a retry strategy from config name and focus categories |
| `GetPhaseInstruction` | Returns format-specific AI instruction text for a given `StructuredFormat` |
| `ContractSchemaPath` | Returns the relative path to contract schema files |

## Patterns

- **Strategy pattern**: `RetryStrategy` interface allows swappable retry behaviors (standard, focused, escalating, escalating-focused)
- **Functional options**: `BuildRetryConfig` uses `WithDebug`, `WithLogger`, `WithTagsConfig`, `WithDefaultMaxAttempts` options
- **Format-aware pipeline**: Each format gets specific instructions and validators loaded automatically
- **No component loggers**: Core library restriction -- uses `zap.NewNop()` by default to prevent writing to command log files
- **Decorator pattern**: `EscalatingStrategy` wraps a base strategy, adding urgency on later attempts

## Internal Structure

| File | Purpose |
|------|---------|
| `formats.go` | `StructuredFormat` enum, `PhaseConfig` type |
| `retry.go` | `RetryConfig`, `RetryResult`, `GenerateWithRetry`, `BuildRetryConfig` factory, functional options |
| `strategies.go` | `RetryStrategy` interface, `StandardStrategy`, `FocusedStrategy`, `EscalatingStrategy`, `GetRetryStrategy` |
| `structured_generator.go` | `StructuredGenerator` with generation loop, validation, output cleanup |
| `types.go` | AI type name constants, path constants, format-specific phase instruction strings |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/ai/config` | `ContractLoader` for loading validation contracts, strategy type constants |
| `core/config` | `TestingTagsConfig` for Gherkin validation tag configuration |
| `core/domain` | `Contract`, `NoOpValidator`, `FormatValidationErrors`, `NewJSONSchemaValidator` |
| `core/paths` | Path constants for schema and template resolution |
| `core/validation` | `Validator` interface, `AIExecutor` interface, `ValidationError`, error codes |
| `core/validation/formats/gherkin` | Gherkin validator for format-aware generation |
| `core/validation/formats/oscal` | OSCAL validator for catalog/profile generation |
| `core/validation/formats/structurizr` | `QuickValidator` for Structurizr DSL generation |

## Role in System

The generation engine behind all `create-*` and AI commands. Commands construct a `RetryConfig` via `BuildRetryConfig`, then call `GenerateWithRetry` which handles the full lifecycle: loading format-specific validators, appending format instructions to prompts, executing AI, validating output, and retrying with strategy-driven error feedback. Commands only handle deterministic formatting of the generated output.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: The `loadValidatorForFormat` function in `retry.go` (line 373) has a large switch statement that grows with each new format. Could benefit from a registry pattern.
- **Optimization Opportunities**: The `stripMarkdownFences` function in `structured_generator.go` (line 208) uses manual string scanning; could use `strings.TrimPrefix`/`strings.TrimSuffix` for clarity.
