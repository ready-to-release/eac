# aiutil

Shared utilities for AI-powered generation commands (create commit-message, create squash-message).

## Key Types

- **`GenerationParams`** -- Configuration for a single AI generation call, including workspace root, prompt, type name, schema filename, optional model override, and an optional test executor

## Key Functions

- `ExecuteGeneration` -- Performs the shared AI executor pipeline: creates executor, registers providers, loads AI config, builds JSON schema validator, builds retry config, and generates with retry
- `LogDebugArtifact` -- Logs debug content with labeled start/end markers to the log file for troubleshooting intermediate outputs
- `ExtractFirstSentence` -- Extracts the first sentence from text, terminated by a period, question mark, or exclamation mark, with a fallback when no delimiter is found

## Patterns

- **Pipeline encapsulation**: The full AI generation pipeline (executor creation, provider registration, config loading, schema validation, retry) is encapsulated in a single `ExecuteGeneration` function, reducing duplication across commands
- **Dependency injection for testing**: `GenerationParams.TestExecutor` allows injecting a mock executor in tests, bypassing real AI provider setup

## Internal Structure

| File | Responsibility |
| --- | --- |
| aiutil.go | All package functionality: GenerationParams, ExecuteGeneration, LogDebugArtifact, ExtractFirstSentence, and ensurePeriod helper |

## Dependencies

- `go/adapters/ai` -- AI executor creation and adapter wrapping
- `go/adapters/ai/providers` -- Built-in AI provider registration
- `go/core/ai` -- Retry configuration, AI config loading, contract loader, GenerateWithRetry
- `go/core/domain` -- JSON schema validator
- `go/core/logging` -- Component logger for warnings and debug output
- `go/core/paths` -- Contract and schema file path resolution

## Role in System

This package serves as the shared foundation for AI-powered content generation commands in the CLI. It eliminates duplication by centralizing the AI executor pipeline -- from provider setup through schema validation and retry -- so that individual commands like `create commit-message` and `create squash-message` only need to supply their specific prompt and configuration.

## Code Health

### Tech Debt
- None

### Pain Points
- None identified

### Optimization Opportunities
- None identified
