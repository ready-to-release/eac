# design

Implements the `update design` command that generates or updates Structurizr DSL architecture diagrams using AI analysis.

## Key Types

- **`Deps`** -- Injectable dependencies struct for testing (AI response override, git repo provider)

## Key Functions

- **`UpdateDesign()`** -- Entry point for the `update design` command
- **`defaultDeps()`** -- Returns production dependencies with LazyRepo-based git provider

## Patterns

- Constructor-based dependency injection: `Deps` struct threaded through `updateDesign(deps)` for test control
- Public entry point delegates to internal function: `UpdateDesign()` calls `updateDesign(defaultDeps())`

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | AI-based DSL update command entry point |
| deps.go | Injectable dependencies struct and production defaults |
| mocks.go | Reserved for test doubles (currently empty) |

## Dependencies

- `core/git` -- git repository access via LazyRepo
- `core/logging` -- structured logging

## Role in System

The `design` sub-package automates architecture diagram maintenance. It uses AI to analyze the codebase and generate or update Structurizr DSL workspace files, keeping architecture documentation in sync with the actual system structure.

## Code Health

### Tech Debt
- None identified

### Pain Points
- update.go is 507 lines (exceeds 300-line threshold)
- No test coverage for deps.go (16 lines), mocks.go (1 line), or update.go (507 lines)

### Optimization Opportunities
- None identified
