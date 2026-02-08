# ownership

Resolves file ownership to modules and components using directory-root
specificity and extension-based tie-breaking.

## Key Types

- **`Resolver`** -- Resolves file paths to owning module and component
- **`Owner`** -- Identifies the module, component, and root for a file
- **`ModuleDefinition`** -- Minimal module data for ownership resolution
- **`ComponentOwnership`** -- Root, file, and extension data for a component
- **`ValidationError`** -- Describes unowned or multi-owned files

## Patterns

- Most-specific-root wins: deeper directory roots take precedence
- Extension tie-breaking: when components share a root, file extension selects
- File-level ownership: exact file matches override all root matches

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `Owner`, `ModuleDefinition`, `ComponentOwnership`, `ValidationError` |
| resolver.go | `Resolver` with root matching, extension tie-breaking, validation |

## Dependencies

_No internal repository imports (leaf package)._

## Role in System

The `ownership` package provides the file-to-module mapping used by
`get changed-modules`, `get files-by-module`, and commit message generation
in `core`. It operates on directory roots from module contracts rather than
glob patterns, keeping resolution deterministic and fast.

## Code Health

### Tech Debt
- None identified

### Pain Points
- `resolver.go:129`: `Validate()` iterates all files x all entries (O(n*m)); may become slow on very large repos

### Optimization Opportunities
- Pre-build a trie or path-prefix index for `findCandidates()` to reduce matching cost on large file sets (medium effort, high value for scale)
- Package is small and well-tested (547-line test file); good candidate for a reference example
