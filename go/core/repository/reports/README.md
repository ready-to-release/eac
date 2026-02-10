# repository/reports

Generates reports about file-to-module ownership across the repository. Calculates coverage statistics including single-owner, multi-owner, and orphan files.

## Key Types

| Type | Purpose |
|------|---------|
| `FilesModulesReport` | Contains all file-module ownership data: counts, per-module file lists, multi-ownership files, orphans |

## Key Functions

| Function | Purpose |
|----------|---------|
| `GetFilesModulesReport` | Generates a complete file-module ownership report with configurable git filtering (tracked, ignored, staged) |

## Patterns

- **Git-aware reporting**: Supports filtering by tracked-only, include-ignored, and staged-only git states
- **Ownership classification**: Files are categorized as single-owner, multi-owner, or orphan
- **Query methods**: `GetCoveragePercentage`, `GetFileModules`, `GetModuleStats` provide convenient access to report data

## Internal Structure

| File | Purpose |
|------|---------|
| `files_modules.go` | `FilesModulesReport` type, `GetFilesModulesReport` function, query methods |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/git` | `NewManager` for opening git repositories |
| `core/repository` | `GetRepositoryFilesWithModules`, `GetFilesByModule`, `GetMultiOwnershipFiles`, `GetOrphanFiles` |

## Role in System

Backs the `show-files` command and module ownership analysis. Provides visibility into which files belong to which modules, helping identify orphaned files that no module claims and multi-owned files that may indicate overlapping module boundaries.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.
