# ai/mock

Provides mock AI responses for testing and development, plus mock implementations of `AIExecutor` and `Validator` interfaces for unit tests. File-based mocking uses environment variables to point at directories of pre-recorded AI responses, allowing deterministic testing without real AI calls.

## Key Types

| Type | Purpose |
|------|---------|
| `AIExecutor` | Mock `validation.AIExecutor` that returns responses from a pre-configured queue |
| `Validator` | Mock `validation.Validator` that returns validation results from a pre-configured queue |

## Key Functions

| Function | Purpose |
|----------|---------|
| `GetMockResponse` | Returns file-based mock AI response for a command (checks env var override, then default patterns) |
| `GetMockResponseWithSubcommand` | Returns mock response for a specific subcommand with fallback to command-level mock |
| `IsMockEnabled` | Returns true if `CLIE_MOCK_AI_DIR` environment variable is set |

## Patterns

- **Environment-driven mocking**: `CLIE_MOCK_AI_DIR` sets the base directory; `CLIE_MOCK_AI_<COMMAND>` overrides specific mock file names
- **Convention over configuration**: Default mock files follow `mock-response.{txt,md,json}` naming in `<dir>/<command>/`
- **Queue-based test doubles**: `AIExecutor` and `Validator` use indexed response queues for predictable multi-call test scenarios
- **Subcommand granularity**: `GetMockResponseWithSubcommand` enables per-subcommand mocks (e.g., `risks/mock-assessment-response.md`)

## Internal Structure

| File | Purpose |
|------|---------|
| `file_mock.go` | File-based mock response loading (`GetMockResponse`, `GetMockResponseWithSubcommand`, `IsMockEnabled`) |
| `testing.go` | Test double implementations (`AIExecutor`, `Validator`) with response queues |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/validation` | `ValidationError` type and `AIExecutor`/`Validator` interfaces |

## Role in System

Enables two testing scenarios: (1) integration tests and CI pipelines use file-based mocks via `CLIE_MOCK_AI_DIR` to avoid real AI API calls, and (2) unit tests use the queue-based `AIExecutor` and `Validator` mocks to verify generation/retry logic in isolation.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.
