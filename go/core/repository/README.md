# repository

Repository-level operations including file enumeration, module ownership
resolution, dependency graph construction, Git context extraction, and
cached file access with GitHub API fallback in CI.

## Key Types

| Type                       | Purpose                                                                  |
| -------------------------- | ------------------------------------------------------------------------ |
| `Repository`               | Wraps a Git repository root with file listing operations                 |
| `FileInfo`                 | File metadata with path, tracked status, and ignored status              |
| `RepositoryFileWithModule` | File path with its owning module monikers                                |
| `FileCache`                | Thread-safe, lazily-populated cache of tracked files with filtered views |
| `GitContext`               | Git repository URL, base commit, and current branch for link generation  |
| `ModuleDependencyGraph`    | Full module dependency graph with forward, reverse edges, and statistics |
| `ModuleDependency`         | Single from/to dependency edge for visualization                         |
| `DependencyGraphStats`     | Root/leaf counts, max fan-in/fan-out for the graph                       |
| `RepositoryError`          | Structured error with operation, path, and underlying cause              |

## Patterns

- Module ownership: files matched against module glob patterns with repository-root fallback
- Lazy file cache: `FileCache` populates from git or GitHub Trees API on first access
- CI optimization: `FileCache` uses GitHub Trees API via `GITHUB_SHA` in CI environments
- Dependency graph: forward and reverse maps with transitive rebuild detection
- Diagram generation: `GetPlantUMLDiagram` and `GetMermaidDiagram` from dependency graph
- Convenience wrappers: `GetRepositoryFilesWithModules` combines enumeration and enrichment

## Internal Structure

| File              | Purpose                                                                   |
| ----------------- | ------------------------------------------------------------------------- |
| `repository.go`   | `Repository`, `FileInfo`, `GetRepositoryRoot`, `GetRepositoryFiles`       |
| `modules.go`      | `EnrichFilesWithModules`, `GetFilesByModule`, `GetOrphanFiles`            |
| `file_cache.go`   | `FileCache` with git and GitHub Trees API backends                        |
| `git_context.go`  | `GitContext` with GitHub URL normalization                                |
| `dependencies.go` | `ModuleDependencyGraph`, transitive rebuild detection, diagram generation |
| `definitions/`    | `DefinitionFile`, YAML enumeration and merge across directory tree        |
| `gomod/`          | `GoModInfo`, `GraphBuilder`, `Mapper`, `Validator` for go.mod analysis    |
| `reports/`        | `FilesModulesReport` for file-module ownership statistics                 |

## Dependencies

| Package               | Purpose                                                                         |
| --------------------- | ------------------------------------------------------------------------------- |
| `core/domain/modules` | `ModuleContract`, `Registry` for module loading and matching                    |
| `core/environments`   | CI detection and container root environment variables                           |
| `core/git`            | `GitRepository`, `Repository`, `RepositoryManager` for git operations           |
| `core/github`         | `API` interface for GitHub Trees API in CI (injected via `FileCache.githubAPI`) |
| `core/paths`          | `EACConfigPath` for repository config directory                                 |
| `core/workspace`      | `DetectWithOptions` for workspace root detection                                |

## Role in System

This package bridges the filesystem and module domain, providing the foundation
for commands that need to know which files belong to which modules, which modules
depend on each other, and which modules require rebuilding after changes. The
`FileCache` accelerates repeated file queries during a single CLI invocation,
while the `gomod/` sub-package validates that Go module dependencies align with
the declared module contracts.

## Code Health

### Tech Debt

- None identified

### Pain Points

- Sub-packages (definitions/, gomod/, reports/) make this the widest package in core with multiple source files across 4 directories
- Large files: repository_test.go (500 lines), file_cache.go (311 lines), repository.go (306 lines)

### Optimization Opportunities

- None identified
