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

## Key Functions

- **`ReleaseThis()`** -- Finalize changelog, calculate version, and prepare release
- **`ReleasePending()`** -- Check if module has unreleased changes for CI/CD decision making
- **`ReleaseExecuteLayers()`** -- Execute releases layer by layer in dependency order
- **`ReleaseAwaitDeps()`** -- Wait for dependency CI to pass before release
- **`ReleaseCheckCI()`** -- Verify CI status for a module release
- **`ReleaseCheckExists()`** -- Check if a release tag already exists
- **`ReleaseCheckPending()`** -- Determine pending release layers
- **`ReleaseValidate()`** -- Validate release configuration
- **`ReleaseChangelog()`** -- Generate and format changelog entries
- **`ReleaseCalver()`** -- Generate CalVer version string
- **`ReleaseGetVersion()`** -- Retrieve current module version
- **`ExtractVersion()`** -- Extract version from tag or changelog
- **`ReleaseTagPending()`** -- Create git tags for pending releases
- **`ReleaseCleanup()`** -- Release cleanup operations
- **`ReleasePrune()`** -- Prune old release artifacts
- **`ReleasePrunePackages()`** -- Prune old container packages
- **`ReleaseExtEac()`** -- EAC extension release handling
- **`ReleaseSrcCli()`** -- CLI-specific release handling

## Patterns

- Table-driven command registration: `commands.go` registers all 18 subcommands via `RegisterAll()`
- Idempotent release: detects already-pending versions to avoid double bumps
- Layered execution: processes releases in dependency order, awaiting each layer
- GitHub CLI integration: dispatches workflows and polls run status via `gh`
- File-ownership release decision: module file changes determine releases, not commit format

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | Table-driven registration of all 18 release subcommands via `RegisterAll()` |
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
| clie.go | CLI-specific release handling |
| scaffold.go | Generic command-scaffold helper for flag/usage boilerplate |

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

The `release` package orchestrates the entire release pipeline in `eac`, from detecting pending changes through version calculation to dispatching GitHub Actions workflows in dependency order. It is the primary interface for both interactive developer-initiated releases (`release this`) and automated CI/CD release flows (`release pending`, `release execute-layers`).

## Code Health

### Tech Debt
- No files over 300 lines; largest files are under 200 lines
- No unit tests for execute-layers.go, cleanup.go, tag-pending.go, clie.go, eac-ext.go, changelog.go, check-exists.go, pending.go, prune.go, prune_packages.go, scaffold.go, or this.go

### Pain Points
- CI-polling logic in check-ci.go and await-deps.go overlaps (both query workflow runs, check ancestors, inherit previous CI)

### Optimization Opportunities
- None identified
