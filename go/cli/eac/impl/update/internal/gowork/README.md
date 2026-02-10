# gowork

Provides Go workspace module discovery by parsing the `go.work` file. Used by update commands that need to iterate over all workspace modules.

## Key Types

None (utility functions only).

## Key Functions

- **`ParseModules()`** -- Read `go.work` at the repository root and return absolute paths to all workspace module directories
- **`IndentLines()`** -- Prefix each non-empty line in text with a given indent string

## Patterns

- Line-by-line parsing: reads `go.work` file and extracts module paths from `use ( ... )` blocks
- Absolute path resolution: converts relative module paths from `go.work` to absolute paths using the repo root

## Internal Structure

| File | Responsibility |
| --- | --- |
| gowork.go | `go.work` file parsing for module discovery and text indentation utility |

## Dependencies

None (standard library only).

## Role in System

The `gowork` package is a shared utility used by `update go-tidy`, `update go-sums`, and `update go-mod-sums` commands. It provides the module discovery mechanism that allows these commands to iterate over all Go modules in the workspace.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
