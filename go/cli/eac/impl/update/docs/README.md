# docs

Synchronizes documentation assets by rendering mermaid diagrams, optimizing drawio images, and syncing command reference docs. Supports selective area updates, dry-run preview, and cache pruning for orphaned files.

## Key Types

- **`UpdateOptions`** -- Parsed flags (areas, dry-run, force, verbose, prune)
- **`Area`** -- Documentation update area (mermaid, drawio, command-refs, all)
- **`AreaSet`** -- Set of selected areas with inclusion check

## Patterns

- Area-based execution: selectively run mermaid, drawio, or command-ref updates
- Cache-based rendering: hash-addressed cache avoids re-rendering unchanged assets
- Dry-run support: preview changes without writing files
- Orphan pruning: detect and remove cache files no longer referenced by docs

## Internal Structure

| File | Responsibility |
| --- | --- |
| update.go | Command entry point, flag parsing, area dispatch |
| types.go | `UpdateOptions` definition |
| areas.go | Area enum, parsing, and `AreaSet` implementation |
| mermaid.go | Mermaid diagram rendering to cache |
| drawio.go | Drawio image optimization to cache |
| command_refs.go | Command reference doc sync (create missing, remove orphans) |
| prune.go | Cache orphan detection and deletion |

## Dependencies

- `clibase/registry` -- command registration
- `clibase/flags` -- flag validation from registry metadata
- `core/logging` -- structured logging
- `core/paths` -- docs cache path resolution
- `core/repository` -- repository root discovery

## Role in System

The `docs` update package keeps documentation assets current in `eac` by rendering diagrams and syncing command references into a content-addressed cache. It is typically invoked before commits that include documentation changes, ensuring rendered assets stay in sync with their source definitions.

## Code Health

### Tech Debt
- No unit tests for mermaid.go (304 lines), drawio.go (169 lines), command_refs.go (149 lines), or update.go (214 lines)
- prune.go has excellent test coverage (912-line test file) but rendering logic has none

### Pain Points
- mermaid.go and drawio.go both shell out to external tools (mermaid-cli, drawio) without testable abstractions, making unit testing difficult
- `runMermaidUpdate` (mermaid.go:32) orchestrates discovery, caching, and batch rendering in a single function

### Optimization Opportunities
- Introduce interface wrappers for external tool invocations to enable unit testing of mermaid.go and drawio.go without Docker/CLI dependencies (moderate feasibility, requires adapter layer)
- Split `runMermaidUpdate` into discovery, cache-check, and render phases for independent testability (high feasibility, clear pipeline stages)
