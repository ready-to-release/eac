# ai

Provides consolidated access to AI configuration, prompt templating,
structured generation with retry, and mock test doubles.

## Key Types

- **`AIConfig`** -- AI provider and model configuration
- **`StructuredGenerator`** -- Format-aware generation with validation
- **`RetryConfig`** -- Retry parameters, strategy, executor, and validator
- **`RetryResult`** -- Generation result with attempt history
- **`RetryStrategy`** -- Interface for retry behavior (standard, focused, escalating)
- **`PromptData`** -- Template data for prompt rendering
- **`MockAIExecutor`** -- Test double for AI execution

## Patterns

- Facade re-export: root `ai.go` re-exports from sub-packages for convenience
- Retry with validation: generate, validate, retry on retriable errors
- Strategy pattern: pluggable retry strategies per generation type
- Factory function: `BuildRetryConfig` assembles config from command inputs

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| ai.go | Facade re-exporting types and functions from sub-packages |
| config/ | `AIConfig` loading, contract loading, type extraction helpers |
| generation/ | `StructuredGenerator`, retry loop, format constants, strategies |
| templates/ | Prompt template rendering with `PromptData` |
| mock/ | File-based mock executor and validator for acceptance tests |

## Dependencies

- `core/validation` -- `Validator`, `AIExecutor` interfaces
- `core/ai/config` -- AI configuration loading
- `core/ai/generation` -- retry and structured generation
- `core/ai/templates` -- prompt template building
- `core/ai/mock` -- mock AI responses for testing

## Role in System

The `ai` package is the primary integration point for AI-powered code
generation in `core`. Commands such as `create-commit-message`, `create-design`,
and `create-spec` use its retry-with-validation pipeline to produce validated
outputs in Gherkin, JSON, OSCAL, or Structurizr formats.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
