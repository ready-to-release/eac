# ctrf

Common Test Report Format (CTRF) types and utilities.
Provides Go types for the CTRF JSON schema, a universal test reporting format.
See: https://ctrf.io/

## Key Types

- `Report` -- top-level CTRF report containing results, tool info, and summary
- `Results` -- test results container with tool, summary, tests, and optional environment
- `Summary` -- aggregate test counts: tests, passed, failed, pending, skipped, other, start/stop times
- `Test` -- single test result: name, status, duration, message, trace, suite, file path
- `Tool` -- identifies the test framework by name and version
- `Status` -- test result status enumeration: `StatusPassed`, `StatusFailed`, `StatusSkipped`, `StatusPending`, `StatusOther`
- `Environment` -- optional environment metadata: OS, architecture, language

## Key Functions

- `NewReport` -- creates a new report with current time as start
- `NewEmptyReport` -- creates a report for aggregation with sentinel start/stop values
- `Parse` -- parses CTRF JSON data into a `Report`
- `ParseFile` -- reads and parses a CTRF JSON file

## Patterns

- **Auto-counting**: `AddTest` automatically increments the appropriate summary counter based on test status
- **Merge support**: `Merge` combines two reports, aggregating test counts and adjusting start/stop times to the earliest/latest
- **Sentinel values**: `NewEmptyReport` uses max-start and zero-stop so the first merge correctly sets both boundaries

## Internal Structure

| File | Purpose |
|---|---|
| `ctrf.go` | All CTRF types, constructors, `AddTest`, `Merge`, `Parse`, and serialization |

## Dependencies

None (leaf package, pure stdlib).

## Role in System

Provides the standard test report format used across all test runners. Each runner produces CTRF reports that are aggregated into a workspace-level report, enabling consistent test result display in summaries, CI pipelines, and external tools.

## Code Health

### Tech Debt
- `ctrf.go` (167 lines) has no test file; `Parse`, `Merge`, and `AddTest` are critical aggregation paths that should have unit tests

### Pain Points
- All types, constructors, parsing, and merging logic live in a single file; splitting into `types.go` and `operations.go` would improve navigability as the format evolves

### Optimization Opportunities
- Add unit tests covering `Merge` edge cases (empty reports, overlapping time ranges, duplicate tests) (low effort)
- Add round-trip `Parse`/`Marshal` tests to verify JSON schema compliance (low effort)
