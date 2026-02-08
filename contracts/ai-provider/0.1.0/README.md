# ai

AI provider port interfaces for integrating language model providers into
the eac system.

## Key Types

- **`Provider`** -- Single AI provider (execute prompts, return responses)
- **`Executor`** -- Orchestrates provider selection and execution
- **`ConfigLoader`** -- Loads AI and git configuration from YAML
- **`Config`** -- AI and git provider configuration value type
- **`Option`** -- Functional option for execution parameters

## Patterns

- Hexagonal ports: `Provider` and `Executor` are consumed by AI commands
- Functional options: `Option` modifies `ExecuteOptions` at call site
- Factory pattern: `ProviderFactory` creates providers from config

## Internal Structure

| File     | Responsibility                                       |
| -------- | ---------------------------------------------------- |
| ports.go | All interfaces, config types, and functional options |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `ai` package (moniker: contracts-ai) defines the provider-agnostic
interface for AI operations such as commit message generation and code
summarization. Adapter implementations (Claude, OpenAI, Gemini) satisfy
`Provider`; commands interact only through `Executor`.

## Code Health

### Tech Debt

- No test files exist for this contract package; add compile-time interface satisfaction checks
- `AIConfig.APIKey` in ports.go is stored as a plain string -- consider a `SecretRef` type to prevent accidental logging

### Pain Points

- `ExecuteOptions` has no validation (e.g., Temperature range 0-1, MaxTokens > 0); adapters must each guard independently
- `ConfigLoader.LoadWithOverrides` takes three string paths -- a struct parameter would be clearer and extensible

### Optimization Opportunities

- Add a `_test.go` with `var _ Provider = ...` compile-time checks -- trivial effort, prevents drift
- Introduce a `Validate()` method on `ExecuteOptions` to centralize bounds checking -- low effort, high reuse
