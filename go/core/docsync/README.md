# docsync

Scans CLI command documentation for missing or orphaned doc files and
generates stub content for undocumented commands.

## Key Types

- **`CommandDocSyncResult`** -- Scan result with counts, missing, and orphaned docs
- **`CommandDocStatus`** -- Sync status of a single command's documentation
- **`CommandInfo`** -- Command name and description from CLI introspection
- **`CommandSource`** -- Function type that returns valid commands

## Patterns

- Dependency injection: `CommandSource` function provides commands for testability
- Marker-based orphan detection: scans `<!-- book:cmd X -->` markers in markdown
- Config-driven paths: uses `CommandsConfig.GetDocPath()` for doc file location

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | `ScanCommandDocs`, `GenerateDocStub`, orphan detection |

## Dependencies

- `core/config` -- `CommandsConfig` for doc path mapping and skip rules

## Role in System

The `docsync` package provides the logic for the `validate docs` and
`update docs` commands in `core`. It ensures every CLI command has a
corresponding documentation file and identifies stale doc files for
removed commands.

## Code Health

### Tech Debt
- None identified. (`cmdMarkerPattern` is a package-level compiled regex with doc comment; this is idiomatic Go.)

### Pain Points
- Single-file package (201 lines) with all logic in one file; manageable now but may need splitting if more doc formats are added

### Optimization Opportunities
- Well-tested (555-line test file with 2.7x test-to-code ratio); no major issues
- Could cache `findOrphanedDocs` directory walk results if called repeatedly (low effort, low priority)
