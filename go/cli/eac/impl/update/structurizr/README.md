# structurizr

Implements the `update structurizr` command that exports Structurizr diagram views from `workspace.dsl` files to SVG format and caches them for acceleration.

## Key Types

None (command-only package; uses inline `moduleStatus` struct for tracking export state).

## Key Functions

- **`UpdateStructurizr()`** -- Entry point for the `update structurizr` command
- **`findCachedSVGs()`** -- Find existing cached SVG files for a module matching a DSL hash

## Patterns

- `init()` registration: registers command function with the global registry
- Hash-based caching: DSL file hash determines cache validity, avoiding redundant exports
- Docker-based export: delegates to `design/helper` package for Structurizr container-based SVG export
- Module iteration: processes all modules with `workspace.dsl` files, with optional single-module filter
- Dry-run support: `--dry-run` flag shows what would be exported without executing

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | Structurizr SVG export with module discovery, cache validation, and Docker-based rendering |

## Dependencies

- `design/helper` -- Structurizr workspace discovery, DSL hashing, and Docker export
- `clibase/flags` -- flag validation from registry metadata
- `clibase/registry` -- command registration
- `core/logging` -- structured logging
- `core/paths` -- Structurizr cache directory path resolution
- `core/repository` -- repository root discovery

## Role in System

The `structurizr` sub-package maintains a cache of SVG renderings of architecture diagrams. It exports views from Structurizr `workspace.dsl` files into the `.cache/eac/structurizr/` directory, enabling fast access to rendered diagrams without running the Structurizr container interactively.

## Code Health

### Tech Debt
- None identified

### Pain Points
- update.go is 283 lines (approaching 300-line threshold)
- No test coverage (missing update_test.go)

### Optimization Opportunities
- None identified
