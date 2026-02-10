# design

Implements the `update design` command that generates or updates Structurizr DSL architecture diagrams using AI analysis.

## Key Types

None (command-only package with mock support types).

## Key Functions

- **`UpdateDesign()`** -- Entry point for the `update design` command (registered via `init()`)
- **`SetMockAIResponse()`** -- Inject a mock AI response for testing
- **`ResetMockAIResponse()`** -- Clear the mock AI response
- **`SetGitRepo()`** -- Inject a mock git repository provider for testing
- **`ResetGitRepo()`** -- Clear the mock git repository provider

## Patterns

- `init()` registration: registers command function with the global registry
- Mock injection via package-level functions: `SetMockAIResponse()` and `SetGitRepo()` for test control
- Global mock state: `mockAIResponse` and `gitRepoProvider` package-level variables for test doubles

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | AI-based DSL update command entry point |
| mocks.go | Mock injection functions for AI responses and git repository providers |

## Dependencies

- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `design` sub-package automates architecture diagram maintenance. It uses AI to analyze the codebase and generate or update Structurizr DSL workspace files, keeping architecture documentation in sync with the actual system structure.

## Code Health

### Tech Debt
- `mockAIResponse` is a mutable global variable; test mock injection uses package-level state rather than dependency injection

### Pain Points
- Mock state must be explicitly reset between tests to avoid cross-test contamination

### Optimization Opportunities
- Replace global mock state with constructor-based dependency injection (moderate effort, improves test isolation)
