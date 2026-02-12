# testjson

Parses Go test JSON output (`go test -json`) into structured events for result aggregation and failure extraction.

## Key Types

- **`GoTestEvent`** -- Represents a single event from `go test -json` output (action, package, test name, output, elapsed time)

## Key Functions

- **`ParseJSONFile()`** -- Parse a file containing newline-delimited Go test JSON events
- **`CountTestResults()`** -- Count pass/fail/skip results from parsed test events
- **`ExtractFailedTests()`** -- Extract names and output of failed tests for display

## Patterns

- Line-by-line JSON parsing: processes newline-delimited JSON (one event per line)
- Event filtering: filters by action type (pass, fail, skip) to extract relevant results

## Internal Structure

| File | Responsibility |
| --- | --- |
| parser.go | Go test JSON event parsing, result counting, and failure extraction |

## Dependencies

None (standard library only).

## Role in System

The `testjson` package handles Go's native test output format. When tests are run with `go test -json`, the output is captured and this package parses it into structured events for aggregation into CTRF reports and display in test summary commands.

## Code Health

### Tech Debt
- None identified

### Pain Points
- No test coverage for parser.go (115 lines)

### Optimization Opportunities
- None identified
