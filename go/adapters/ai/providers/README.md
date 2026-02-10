# providers

Concrete AI provider implementations for multiple LLM backends, with a registry for bulk registration.

## Key Types

- **`ClaudeAPI`** -- Anthropic Claude API provider using API key authentication
- **`ClaudeCLI`** -- Claude CLI provider using subscription authentication
- **`OpenAI`** -- OpenAI GPT provider using API key authentication
- **`Gemini`** -- Google Gemini provider using API key authentication
- **`MockProvider`** -- Test provider returning a configured response
- **`TestProvider`** -- Acceptance test provider reading mock responses from files or env vars
- **`ExecutorRegistry`** -- Interface for registering providers with an executor

## Key Functions

- `RegisterBuiltIn` -- Registers all built-in providers (claude-cli, claude-api, openai, gemini, test) with an executor

## Patterns

- Factory registration: Each provider is registered as a `ProviderFactory` closure
- Fail-fast validation: Providers return errors immediately if API key or model is missing
- Subscription auth: `ClaudeCLI` removes `ANTHROPIC_API_KEY` from env to force subscription auth
- Model mapping: `ClaudeCLI` maps full model IDs to CLI short names (haiku, sonnet, opus)
- Mock resolution order: `TestProvider` checks file first, then env var, then errors

## Internal Structure

| File          | Responsibility                                                           |
| ------------- | ------------------------------------------------------------------------ |
| registry.go   | `RegisterBuiltIn` function and `ExecutorRegistry` interface              |
| anthropic.go  | `ClaudeAPI` provider using Anthropic SDK                                 |
| claude_cli.go | `ClaudeCLI` provider using Claude CLI tool with subscription auth        |
| openai.go     | `OpenAI` provider using OpenAI SDK                                       |
| gemini.go     | `Gemini` provider using Google Generative AI SDK                         |
| mock.go       | `MockProvider` for unit testing                                          |
| test.go       | `TestProvider` for acceptance testing with file/env-based mock responses |

## Dependencies

- `adapters/ai` -- `Provider` interface, `Config`, and `Option` types
- `core/environments` -- Environment variable constants for test provider
- `core/paths` -- AI test mock file path resolution
- `core/repository` -- Repository root resolution for test provider

## Role in System

The providers sub-package contains all concrete AI provider implementations used by the `ai` adapter. It is kept separate from the parent `ai` package to avoid import cycles, since provider implementations depend on external SDKs (Anthropic, OpenAI, Gemini) while the parent package defines the core interfaces.

## Code Health

### Tech Debt

- None

### Pain Points

- `Gemini.Execute` in gemini.go creates a new `genai.Client` on every call; caching the client would reduce connection overhead for repeated invocations

### Optimization Opportunities

- None identified
