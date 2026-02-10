# show

Parent command that displays repository information in human-readable format. Houses over 30 subcommands producing markdown tables, CI summaries, and formatted reports for modules, builds, tests, releases, and more.

## Key Types

- **`SummaryFormatter`** -- Generates markdown headers, tables, and collapsible sections
- **`TestResults`** -- Parsed test results with package breakdown
- **`PackageTestResults`** -- Test results for a single package
- **`testStats`** -- Aggregated test statistics (total, passed, failed, skipped)
- **`artifactStats`** -- Summary statistics for a module's artifacts

## Key Functions

- **`Show()`** -- Parent command entry point with TUI interactive selector fallback
- **`ShowModules()`** -- Display module contracts in a table with optional artifact stats
- **`ShowBuildSummary()`** -- GitHub Actions build summary with metrics and diagnostics
- **`ShowTestSummary()`** -- GitHub Actions test summary with per-module breakdown
- **`ShowTestResults()`** -- Detailed test result display from UoW manifests
- **`ShowCISummary()`** -- CI workflow summary with job result table
- **`ShowCIResults()`** -- CI run results display
- **`ShowConfig()`** -- Display full EAC configuration summary
- **`ShowComponents()`** -- Display all components with phase and dependency info
- **`ShowUnits()`** -- Display units of work for a framework action
- **`ShowDependencies()`** -- Module dependency display
- **`ShowSpecs()`** -- Specification file display
- **`ShowChangelog()`** -- Changelog entry display
- **`ShowReleaseSummary()`** -- Release status summary
- **`ShowReleaseNotes()`** -- Release notes display
- **`ShowArtifacts()`** -- Artifact listing with resolution status
- **`ShowFiles()`** -- Repository files with module ownership display
- **`ShowFilesChanged()`** -- Modified files with module ownership
- **`ShowFilesStaged()`** -- Staged files with module ownership
- **`ShowValidCommands()`** -- Available commands display

## Patterns

- Table-driven command registration: `commands.go` registers all 33 subcommands via `RegisterAll()` in `init()`
- TUI selector fallback: interactive subcommand picker when run without arguments
- SummaryFormatter reuse: shared formatter drives consistent markdown output across summaries
- Status derivation from manifests: build/test status inferred from UoW manifest data
- Comment-driven metadata: command, flags, and grouping declared via structured comments

## Internal Structure

| File | Responsibility |
| --- | --- |
| commands.go | Table-driven registration of all 33 show subcommands via `RegisterAll()` |
| show.go | Parent command entry point with TUI and subcommand dispatch |
| summary_formatter.go | Shared markdown formatting utilities (tables, sections, code blocks) |
| modules.go | Display module contracts in a table with optional artifact stats |
| components.go | Display all components with phase and dependency info |
| units.go | Display units of work for a framework action |
| config.go | Display full EAC configuration summary |
| build-summary.go | GitHub Actions build summary with metrics and diagnostics |
| build-times.go | Build duration analysis |
| test-summary.go | GitHub Actions test summary with per-module breakdown |
| test-results.go | Detailed test results from UoW manifests |
| test-timings.go | Test duration analysis |
| tests.go | Test listing display |
| suite.go | Test suite display |
| ci-summary.go | CI workflow summary with job result table |
| ci-results.go | CI run results display |
| release-summary.go | Release status summary |
| release-notes.go | Release notes display |
| artifacts.go | Artifact listing with resolution status |
| books.go | Configured books display |
| files.go | Repository files with module ownership display |
| files-changed.go | Modified files with module ownership |
| files-staged.go | Staged files with module ownership |
| dependencies.go | Module dependency display |
| environments.go | Environment configuration display |
| specs.go | Specification file display |
| changelog.go | Changelog entry display |
| ghosts.go | Ghost file detection |
| approval-comments.go | PR approval comments display |
| approve-summary.go | Approval status summary |
| component-types.go | Component type listing |
| lint-summary.go | Lint results summary |
| scan-summary.go | Security scan results summary |
| trigger-summary.go | CI trigger analysis |
| dependency-ci-summary.go | Dependency CI status summary |
| deps-setup-summary.go | Dependency setup summary |
| valid-commands.go | Available commands display |

## Dependencies

- `cli/eac/help` -- help text rendering for parent command
- `cli/eac/impl/internal` -- shared artifact resolution types
- `cli/eac/impl/internal/manifests/testview` -- test view loading from UoW manifests
- `cli/eac/impl/internal/testdata` -- shared test data preparation
- `cli/eac/impl/show/internal` -- artifact formatting utilities
- `clibase/registry` -- command registration and subcommand discovery
- `clibase/flags` -- flag validation and parsing
- `clibase/render` -- table builder for markdown output
- `adapters/tui` -- TUI detection and subcommand options
- `adapters/tui/selector` -- interactive command selector
- `core/config` -- configuration and module contract loading
- `core/repository` -- repository root discovery
- `core/output` -- UoW manifest reading for status derivation
- `core/domain/reports` -- module contract reports

## Role in System

The `show` package is the human-readable counterpart to `get`, producing formatted markdown tables and CI summaries rather than machine-parseable output. It is heavily used in GitHub Actions workflows where its output is directed to `$GITHUB_STEP_SUMMARY`, providing consistent build, test, and release reporting across all modules in the `eac` pipeline.

## Code Health

### Tech Debt
- `test-summary.go` is 655 lines with `generateCombinedSummary` spanning ~125 lines; it mixes table building, status derivation, and markdown generation
- Only 7 test files for ~35 source files; most summary generators (`build-summary.go`, `ci-summary.go`, `test-results.go`, etc.) have no unit tests

### Pain Points
- Pattern duplication: each subcommand file repeats flag-parse and error-handling boilerplate similar to `get/` commands, without a shared `ExecuteShowCommand` helper
- `build-summary.go` (354 lines) and `test-summary.go` (655 lines) are significantly larger than typical subcommands, suggesting they could be decomposed

### Optimization Opportunities
- Introduce an `ExecuteShowCommand` helper analogous to `get`'s `ExecuteGetCommand` to standardize flag parsing and output rendering (moderate effort)
- Extract the table-building logic from `test-summary.go` into `summary_formatter.go` for reuse (low effort)
