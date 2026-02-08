# show

Parent command that displays repository information in human-readable format. Houses over 30 subcommands producing markdown tables, CI summaries, and formatted reports for modules, builds, tests, releases, and more.

## Key Types

- **`SummaryFormatter`** -- Generates markdown headers, tables, and collapsible sections
- **`TestResults`** -- Parsed test results with package breakdown
- **`PackageTestResults`** -- Test results for a single package
- **`testStats`** -- Aggregated test statistics (total, passed, failed, skipped)
- **`artifactStats`** -- Summary statistics for a module's artifacts

## Patterns

- Registry-based subcommand dispatch: each subcommand file calls `registry.Register` in `init()`
- TUI selector fallback: interactive subcommand picker when run without arguments
- SummaryFormatter reuse: shared formatter drives consistent markdown output across summaries
- Status derivation from manifests: build/test status inferred from UoW manifest data
- Comment-driven metadata: command, flags, and grouping declared via structured comments

## Internal Structure

| File | Responsibility |
| --- | --- |
| show.go | Parent command entry point with TUI and subcommand dispatch |
| summary_formatter.go | Shared markdown formatting utilities (tables, sections, code blocks) |
| modules.go | Display module contracts in a table with optional artifact stats |
| build-summary.go | GitHub Actions build summary with metrics and diagnostics |
| test-summary.go | GitHub Actions test summary with per-module breakdown |
| ci-summary.go | CI workflow summary with job result table |
| release-summary.go | Release status summary |
| files.go | Repository files with module ownership display |
| dependencies.go | Module dependency display |
| specs.go | Specification file display |
| changelog.go | Changelog entry display |
| lint-summary.go | Lint results summary |
| scan-summary.go | Security scan results summary |

## Dependencies

- `cli/eac/help` -- help text rendering for parent command
- `cli/eac/impl/internal` -- shared artifact resolution types
- `cli/eac/impl/internal/manifests/testview` -- test view loading from UoW manifests
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

The `show` package is the human-readable counterpart to `get`, producing formatted markdown tables and CI summaries rather than machine-parseable output. It is heavily used in GitHub Actions workflows where its output is directed to `$GITHUB_STEP_SUMMARY`, providing consistent build, test, and release reporting across all modules in the `eac-cli` pipeline.

## Code Health

### Tech Debt
- `test-summary.go` is 655 lines with `generateCombinedSummary` spanning ~125 lines (155-279); it mixes table building, status derivation, and markdown generation
- Only 7 test files for ~35 source files; most summary generators (`build-summary.go`, `ci-summary.go`, `test-results.go`, etc.) have no unit tests
- `show.go:45`: `subcommands` is a package-level mutable var containing TUI subcommand metadata

### Pain Points
- Pattern duplication: each subcommand file repeats init/register and flag-parse boilerplate similar to `get/` commands, but without a shared `ExecuteShowCommand` helper
- `build-summary.go` (354 lines) and `test-summary.go` (655 lines) are significantly larger than typical subcommands, suggesting they could be decomposed

### Optimization Opportunities
- Introduce an `ExecuteShowCommand` helper analogous to `get`'s `ExecuteGetCommand` to standardize flag parsing and output rendering -- moderate effort
- Extract the table-building logic from `test-summary.go` (`buildTestTable`, `packageBreakdown`) into `summary_formatter.go` for reuse -- low effort
