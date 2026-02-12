# initsummary

Data structures and formatters for command initialization summaries.
Captures the outcome of init-phase checks (dependencies, incremental detection, flags)
for display before execution begins.

## Key Types

- `Summary` -- top-level summary containing all init-phase results: flags, deps status, incremental info, parallelism, test metadata
- `Flags` -- records which CLI flags were active (dry-run, force-rebuild, turbo, sequential)
- `DepsStatus` -- system dependency verification results: tool availability, version checks
- `DepmStatus` -- module dependency (build artifact) validation results
- `IncrementalInfo` -- incremental change detection results: changed modules, skipped modules, cache hit ratio
- `TestInfo` -- test-specific summary: suite count, tag filter, coverage mode
- `ArtifactValidationInfo` -- build artifact validation results per module
- `ParallelismInfo` -- resolved concurrency settings: max workers, turbo multiplier, pool sizes

## Key Functions

- `FormatCompact` -- renders the summary as a compact single-line or few-line string for console mode
- `FormatDetailed` -- renders the summary as a detailed multi-line report for TUI init panel

## Patterns

- **Progressive population**: the `Summary` struct is populated incrementally as each init phase completes, then formatted once before execution
- **Dual formatting**: compact format for console mode, detailed format for TUI init panel, both driven by the same data

## Internal Structure

| File | Purpose |
|---|---|
| `summary.go` | `Summary` and all supporting data types (`Flags`, `DepsStatus`, `IncrementalInfo`, etc.) |
| `formatter.go` | `FormatCompact` and `FormatDetailed` rendering functions |
| `formatter_flags.go` | Flag-specific formatting helpers extracted from formatter.go |
| `formatter_helpers.go` | Shared section-building helpers used by both compact and detailed formatters |

## Dependencies

None (leaf package within clibase).

## Role in System

Provides the structured data layer between init-phase verification and user-facing summary display. The command framework populates a `Summary` during init, then passes it to the formatter for rendering in either console or TUI mode.

## Code Health

### Tech Debt

- None identified

### Pain Points

- `summary.go` (396 lines) and `formatter.go` (359 lines) exceed 300 lines
- `summary_test.go` (427 lines) is a large test file

### Optimization Opportunities

- None identified
