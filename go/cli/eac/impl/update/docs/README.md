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
- None identified

### Pain Points
- `prune_test.go` is 925 lines (exceeds 300-line threshold)

### Optimization Opportunities
- None identified
