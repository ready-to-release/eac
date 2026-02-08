# ghost

Discovers ghost-prefixed files and directories used for dark launching,
feature toggles, and monitoring probes.

## Key Types

- **`Ghost`** -- Discovered ghost entity with path, type, and module owner
- **`GhostReport`** -- All ghosts with summary statistics and config
- **`Scanner`** -- Discovers ghosts from a file source without filesystem walks
- **`FileSource`** -- Interface providing tracked file list for scanning
- **`FilterOptions`** -- Criteria for filtering ghosts by type, module, or ownership

## Patterns

- FileSource abstraction: scanner works with git-tracked file lists, not direct I/O
- Convention-based detection: matches `ghost-*`, `ghost.*`, or exact `ghost` names
- Module ownership: optional registry resolves each ghost to its owning module

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `Ghost`, `GhostReport`, `GhostSummary`, `FilterOptions`, `Filter` |
| scanner.go | `Scanner` with pattern matching and module resolution |
| report.go | `BuildReport` and `BuildSummary` from scan results |

## Dependencies

- `core/domain/modules` -- module registry for ownership resolution

## Role in System

The `ghost` package powers the `get ghosts` and `show ghosts` commands in
`core`. It scans git-tracked files for ghost-prefixed entities, resolves
module ownership, and produces reports that help teams track dark-launched
features across the codebase.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- Package is clean: 307 lines of production code, 343 lines of tests; well-covered
- `FileSource` interface is single-method, keeping the design minimal; no changes recommended
