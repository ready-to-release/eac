# results

Provides test result parsing and aggregation for Cucumber and CTRF report formats. Discovers result files and merges them into unified summaries.

## Key Types

- **`CucumberTestResult`** -- Wrapper around parsed Cucumber JSON report with source file path and module association

## Key Functions

- **`ParseCucumberResults()`** -- Discover and parse all Cucumber JSON result files from output directories
- **`isCucumberFile()`** -- Check if a file is a Cucumber JSON report by name pattern
- **`AggregateCucumberReports()`** -- Merge multiple Cucumber reports into a single aggregated report
- **`AggregateCTRFReports()`** -- Merge multiple CTRF reports into a single aggregated report

## Patterns

- File discovery: walks output directories to find test result files by naming convention
- Format-specific aggregation: separate aggregation paths for Cucumber and CTRF formats
- Module association: infers module ownership from directory structure

## Internal Structure

| File | Responsibility |
| --- | --- |
| results.go | Cucumber and CTRF result file discovery, parsing, and aggregation |

## Dependencies

- `cli/eac/impl/test/internal/ctrf` -- CTRF report parsing and merging
- `cli/eac/impl/test/internal/cucumber` -- Cucumber report parsing
- `core/logging` -- structured logging

## Role in System

The `results` package provides the discovery and aggregation layer for test result files. It finds test output files across the `out/test/` directory tree, parses them into typed structures, and aggregates multiple reports into unified summaries used by reporting and display commands.

## Code Health

### Tech Debt
- None identified

### Pain Points
- No test coverage for results.go (251 lines)

### Optimization Opportunities
- None identified
