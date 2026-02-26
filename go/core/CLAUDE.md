# core Module

Module: `github.com/ready-to-release/eac/go/core`
Role: Shared domain libraries used by eac, mcp, adapters, and clibase.

## Dependency Position

core depends on: `contracts/*` modules only
core is depended on by: eac CLI, mcp server, adapters, clibase

Do NOT add imports from `go/cli/*` into core — that would create a cycle.

## Key Packages

- `environments/` — centralized environment variable constants (76+ vars)
- `workspace/` — workspace detection and path resolution
- `module-deps/` — module dependency graph
- `scheduling/` — execution order and scheduling
- `config/` — configuration loading

## Environment Constants Note

`go/core/environments/constants.go` is the source of truth for env vars used by
eac, mcp, and adapters. The `go/cli/clie` module maintains a SEPARATE copy in
its own `internal/envconsts/` — do NOT merge them. The separation is required
by the clie isolation constraint.
