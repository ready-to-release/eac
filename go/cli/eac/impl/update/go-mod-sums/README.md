# go-mod-sums

Implements the `update go-mod-sums` command that updates `go.sum` files across all Go workspace modules by running `go mod download` in each module directory.

## Key Types

None (command-only package).

## Key Functions

- **`UpdateGoModSums()`** -- Entry point for the `update go-mod-sums` command (registered via `init()`)

## Patterns

- `init()` registration: registers command function with the global registry
- Workspace module iteration: uses `gowork.ParseModules()` to discover all Go workspace modules and processes each

## Internal Structure

| File | Responsibility |
| --- | --- |
| sums.go | Go module sum file update via `go mod download` across all workspace modules |

## Dependencies

- `cli/eac/impl/update/internal/gowork` -- Go workspace module discovery
- `clibase/registry` -- command registration
- `core/logging` -- structured logging

## Role in System

The `go-mod-sums` sub-package ensures `go.sum` files are up-to-date across all workspace modules. It complements `update go-tidy` by specifically targeting dependency hash verification without modifying `go.mod` files.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
