# go-sums

Implements the `update go-sums` command that refreshes `go.sum` files across all Go workspace modules. Similar to `go-mod-sums` but uses a different update strategy.

## Key Types

None (command-only package).

## Key Functions

- **`UpdateGoSums()`** -- Entry point for the `update go-sums` command (registered via `init()`)

## Patterns

- `init()` registration: registers command function with the global registry
- Workspace module iteration: uses `gowork.ParseModules()` to discover all Go workspace modules and processes each

## Internal Structure

| File | Responsibility |
| --- | --- |
| tidy.go | Go sum file refresh across all workspace modules |

## Dependencies

- `cli/eac/impl/update/internal/gowork` -- Go workspace module discovery
- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `go-sums` sub-package provides an alternative approach to updating Go dependency hash files across workspace modules, complementing the `go-mod-sums` and `go-tidy` commands.

## Code Health

### Tech Debt
- None identified

### Pain Points
- tidy.go is 230 lines (under 300-line threshold but approaching it)

### Optimization Opportunities
- None identified
