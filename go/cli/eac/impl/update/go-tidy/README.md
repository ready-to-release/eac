# go-tidy

Implements the `update go-tidy` command that runs `go mod tidy` across all Go workspace modules to clean up `go.mod` and `go.sum` files.

## Key Types

None (command-only package).

## Key Functions

- **`UpdateGoTidy()`** -- Entry point for the `update go-tidy` command (registered via `init()`)

## Patterns

- `init()` registration: registers command function with the global registry
- Workspace module iteration: uses `gowork.ParseModules()` to discover all Go workspace modules and runs `go mod tidy` in each

## Internal Structure

| File | Responsibility |
| --- | --- |
| tidy.go | Run `go mod tidy` across all workspace modules |

## Dependencies

- `cli/eac/impl/update/internal/gowork` -- Go workspace module discovery
- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `go-tidy` sub-package ensures all Go workspace modules have clean `go.mod` and `go.sum` files by running `go mod tidy` in each module directory. This is essential before commits and CI runs to prevent dependency resolution failures.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
