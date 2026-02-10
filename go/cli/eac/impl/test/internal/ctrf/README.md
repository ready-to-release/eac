# ctrf

Implements the Common Test Report Format (CTRF) for universal test result reporting. Provides types for constructing, parsing, merging, and serializing CTRF reports.

## Key Types

- **`Report`** -- Top-level CTRF report containing results, tool info, and summary
- **`Results`** -- Container for tool metadata, summary statistics, and individual test entries
- **`Tool`** -- Metadata about the test tool that generated the report (name, version)
- **`Summary`** -- Aggregate test counts (total, passed, failed, skipped, pending, other)
- **`Test`** -- Individual test result with name, status, duration, suite, file path, and optional message/trace
- **`Environment`** -- Optional test environment metadata

## Key Functions

- **`NewReport()`** -- Create a new CTRF report with tool name and version
- **`NewEmptyReport()`** -- Create an empty report for aggregation
- **`AddTest()`** -- Add a test result to the report
- **`Finalize()`** -- Recompute summary statistics from individual test entries
- **`ToJSON()`** -- Serialize report to JSON bytes
- **`Merge()`** -- Merge another report's tests into this report
- **`ParseFile()`** -- Load and parse a CTRF report from a JSON file
- **`Parse()`** -- Parse CTRF report from raw JSON bytes

## Patterns

- Builder pattern: `NewReport()` then `AddTest()` then `Finalize()` for constructing reports
- Merge-based aggregation: `Merge()` combines multiple reports into a single unified report
- Status constants: `StatusPassed`, `StatusFailed`, `StatusSkipped`, `StatusPending`, `StatusOther`

## Internal Structure

| File | Responsibility |
| --- | --- |
| ctrf.go | All CTRF types, constructors, serialization, parsing, merging, and status constants |

## Dependencies

None (standard library only).

## Role in System

The `ctrf` package provides the universal test result format used across the eac test reporting pipeline. Go test results, godog BDD results, and other test outputs are converted to CTRF format for uniform aggregation and display by show/test commands.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- All types and functions are in a single file (`ctrf.go`), which may grow as more test formats are supported

### Optimization Opportunities
- None identified.
