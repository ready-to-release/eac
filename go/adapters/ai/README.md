# ai

AI adapter providing a pluggable provider abstraction for LLM integration,
with configuration loading, environment variable substitution, and schema
validation.

## Key Types

- **`Provider`** -- Interface all AI providers must implement
- **`Executor`** -- Orchestrates provider selection and execution
- **`ExecutorAdapter`** -- Adapts `Executor` to domain `AIExecutor` interface
- **`Config`** -- AI and Git configuration from `.eac/ai-provider.yml`
- **`AIConfig`** -- Provider, model, endpoint, and API key settings
- **`ProviderFactory`** -- Creates a provider from configuration
- **`Option`** -- Functional option for execution (model, temperature)
- **`MockProvider`** -- Test provider returning configured responses

## Patterns

- Provider pattern: Pluggable AI backends registered via `ProviderFactory`
- Functional options: `WithModel`, `WithTemperature`, `WithMaxTokens`
- Config layering: Defaults, team config, and personal overrides merged
- Env var substitution: `${VAR_NAME}` syntax in config values
- Schema validation: Configs validated against ai-provider schema

## Internal Structure

| File                | Responsibility                                                |
| ------------------- | ------------------------------------------------------------- |
| provider.go         | `Provider` interface and functional options                   |
| executor.go         | `Executor` orchestrating provider lifecycle                   |
| executor_adapter.go | `ExecutorAdapter` bridging to domain interface                |
| config.go           | `Config`, `AIConfig`, and `GitConfig` types                   |
| config_loader.go    | Config loading, merging, and env var substitution             |
| mock.go             | `MockProvider` for testing                                    |
| doc.go              | Package documentation                                         |
| providers/          | Concrete provider implementations (Anthropic, OpenAI, Gemini) |
| toolhandler/        | AI tool call handler integration                              |

## Dependencies

- `core/domain/schema` -- schema validation for config files
- `core/environments` -- environment variable constants
- `core/paths` -- config file path resolution

## Role in System

The `ai-eac` module implements the AI provider abstraction used by
commands that require LLM capabilities (spec generation, risk assessment,
commit messages). It loads provider configuration from the `.eac` directory,
supports multiple AI backends through a pluggable factory pattern, and
provides the `ExecutorAdapter` that satisfies the domain's `AIExecutor`
interface.

## Code Health

### Tech Debt

- Package-level `schemaValidator` and `schemaValidatorOnce` in config_loader.go are global mutable state coupled to the workspace root of the first caller

### Pain Points

- `Executor.loadConfig()` reads from disk on every `Execute()` call; caching the config with file-mtime invalidation would reduce I/O for repeated invocations

### Optimization Opportunities

- None identified
