# release

Manages the full release lifecycle including changelog generation, version calculation, CI verification, dependency gating, and layered workflow execution. Supports both SemVer and CalVer versioning schemes with configurable constraints.

## Key Types

- **`ReleaseResult`** -- Outcome of a single module release operation
- **`PendingRelease`** -- Release decision data for CI/CD pipelines
- **`PendingResult`** -- Aggregated pending status for multiple modules
- **`ChangeSummary`** -- Breakdown of changes by type (added, fixed, etc.)
- **`LayerModule`** -- Module entry in a dependency-ordered release layer
- **`LayerRun`** -- Tracks a dispatched workflow run
- **`DepCIStatus`** -- CI verification status for a dependency

## Patterns

- Registry-based subcommand dispatch: each subcommand file calls `registry.Register` in `init()`
- Idempotent release: detects already-pending versions to avoid double bumps
- Layered execution: processes releases in dependency order, awaiting each layer
- GitHub CLI integration: dispatches workflows and polls run status via `gh`
- File-ownership release decision: module file changes determine releases, not commit format

## Internal Structure

| File | Responsibility |
| --- | --- |
| this.go | Finalize changelog, calculate version, prepare release |
| pending.go | Check if module has unreleased changes for CI/CD |
| execute-layers.go | Execute releases layer by layer in dependency order |
| await-deps.go | Wait for dependency CI to pass before release |
| check-ci.go | Verify CI status for a module release |
| check-exists.go | Check if a release tag already exists |
| check-pending.go | Determine pending release layers |
| validate.go | Validate release configuration |
| changelog.go | Changelog operations |
| calver.go | CalVer version generation |
| get-version.go | Retrieve current module version |
| extract-version.go | Extract version from tag or changelog |
| tag-pending.go | Create git tags for pending releases |
| cleanup.go | Release cleanup operations |
| prune.go | Prune old release artifacts |
| prune_packages.go | Prune old container packages |
| eac-ext.go | EAC extension release handling |
| clie-cli.go | CLI-specific release handling |

## Dependencies

- `clibase/registry` -- command registration and workspace root
- `clibase/flags` -- flag validation from registry metadata
- `clibase/ghexec` -- GitHub CLI execution for workflow dispatch
- `clibase/gitexec` -- git command execution
- `core/changelog` -- changelog parsing, writing, and version calculation
- `core/config` -- configuration and versioning constraints
- `core/domain/modules` -- module contract registry
- `core/git` -- git repository operations (tags, commits)
- `core/releasenotes` -- release notes template generation and parsing
- `core/repository` -- repository root discovery

## Role in System

The `release` package orchestrates the entire release pipeline in `eac-cli`, from detecting pending changes through version calculation to dispatching GitHub Actions workflows in dependency order. It is the primary interface for both interactive developer-initiated releases (`release this`) and automated CI/CD release flows (`release pending`, `release execute-layers`).

## Code Health

### Tech Debt
- 18 files each declare `func init()` with near-identical registry-registration boilerplate
- Several oversized functions: `ReleaseCheckCI` (~213 lines in check-ci.go), `ReleaseAwaitDeps` (~214 lines in await-deps.go), `performRelease` (~292 lines in this.go), `ReleaseChangelog` (~286 lines in changelog.go)
- No tests for execute-layers.go, cleanup.go, tag-pending.go, or clie-cli.go

### Pain Points
- Duplication across command files: each subcommand repeats flag-parsing, usage-printing, and error-handling scaffolding that could be extracted into a shared harness
- CI-polling logic in check-ci.go and await-deps.go overlaps significantly (both query workflow runs, check ancestors, inherit previous CI)

### Optimization Opportunities
- Extract a generic command-scaffold helper to eliminate per-file init/flag/usage boilerplate (high feasibility, mechanical refactor)
- Split check-ci.go and await-deps.go into smaller focused functions and share CI-query utilities (moderate feasibility, needs careful integration testing)
