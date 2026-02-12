# reports

Report generation for CLI display commands. Each report type loads data from
the workspace, resolves cross-references, and produces structured output
that the TUI layer renders into tables and summaries.

## Key Types

- `ComponentReport` and `ComponentInfo` describe components with their phases (build, lint, test, scan)
- `ModuleContractReport` pairs module contracts with their resolved component details
- `UnitReport` and `UnitInfo` represent work units with cache status (up_to_date, stale, new)
- `SpecsReport` and `SpecFile` extract specification files with authorship from git history
- `ChangelogReport` wraps parsed changelog content for a module
- `ReleaseNotesReport` wraps release notes content for a module
- `ApprovalCommentsReport` and `ApprovalComment` fetch PR approval data from GitHub CLI
- `VersionInfo` resolves module versions across SemVer, CalVer, and Implicit schemes

## Patterns

- Process-level caching with `sync.Once` stores expensive results (module contracts, component types) for reuse
- Report functions accept a workspace root and return a typed report struct plus error
- `ResolveVersion` and `ResolveVersionWithValidation` handle the three versioning schemes uniformly
- `CacheStatus` enum tracks whether a work unit needs rebuilding based on content hash comparison
- Component phase resolution maps component types to their configured builders, linters, testers, and scanners
- GitHub CLI integration shells out to `gh` for approval comment retrieval

## Internal Structure

| Path | Purpose |
|------|---------|
| `components.go` | Component report with phase resolution and caching |
| `contracts-and-modules.go` | Module contract report with component enrichment |
| `units.go` | Work unit report with cache status computation |
| `specs.go` | Specification file listing with git blame integration |
| `changelog.go` | Changelog report for a single module |
| `release-notes.go` | Release notes report for a single module |
| `version_resolver.go` | Version resolution across SemVer, CalVer, Implicit |
| `approval_comments.go` | GitHub PR approval comment retrieval |

## Dependencies

- `core/config` for `EACConfig` and component type definitions
- `core/domain/modules` for `ModuleContract` and `Registry`
- `core/changelog` for changelog parsing
- `core/releasenotes` for release notes parsing
- `core/git` for git history queries
- `core/github` for GitHub CLI integration
- `core/hash` for content hash computation
- `core/paths` for workspace path resolution
- `core/workunit` for work unit state tracking
- `core/logging` for structured logging
- `contracts/core` for action type constants

## Role in System

This package sits between the domain model and the CLI presentation layer.
Commands like `show components`, `show units`, and `show specs` delegate to
report functions here, which load the module registry, enrich it with
runtime data (versions, cache state, git history), and return structured
results that the TUI renderers format for terminal output.

## Code Health

### Tech Debt
- `resolvePhases` in `components.go:307-313` hard-codes testable component types (go, gherkin, typescript, python) instead of reading from blueprints.yml component-kinds testers config
- `globalComponentCache` in `components.go:103` has no invalidation path beyond process restart

### Pain Points
- `components.go` (465 lines) is a candidate for splitting into separate files for cache management, dependency graph building, and report generation
- `approval_comments.go` (318 lines) and `specs.go` (209 lines) share nearly identical branch-resolution and commit-range logic; extracting a shared helper would reduce duplication
- Each report function independently calls config.Load() or config.Global(); callers cannot inject a pre-loaded config

### Optimization Opportunities
- GetSpecs and GetApprovalComments each open the git repository independently; sharing a single git.GitRepository instance across a report session would avoid repeated open calls
