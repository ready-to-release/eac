# reporter

Collects and aggregates test results from multiple report formats (Cucumber JSON, Go test JSON) into unified report data structures for display and evidence generation.

## Key Types

- **`TestSuiteReportData`** -- Top-level aggregated test report containing all modules
- **`ModuleReportData`** -- Per-module test results with feature data and unit test data
- **`FeatureReportData`** -- BDD feature test results from Cucumber reports
- **`UnitTestReportData`** -- Go unit test results parsed from JSON test output
- **`BuildReportData`** -- Combined build and test report data

## Key Functions

- **`BuildReportData()`** -- Construct a complete report from test suite data
- **`CollectModuleReports()`** -- Scan output directories and collect test reports for all modules
- **`processeCucumberReport()`** -- Process a single Cucumber JSON report into feature report data
- **`processGoTestLog()`** -- Process a Go test JSON log into unit test report data
- **`parseGoTestLog()`** -- Parse Go test JSON output into structured test events
- **`extractModuleFromFeaturePath()`** -- Extract module name from a feature file path

## Patterns

- Multi-format collection: discovers and processes both Cucumber JSON and Go test JSON formats
- Directory scanning: walks `out/test/{module}/` directories to find report files
- Aggregation: combines per-module results into a suite-level report

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Report data types and `BuildReportData()` constructor |
| collector.go | Directory scanning, report collection, and multi-format parsing |

## Dependencies

- `cli/eac/impl/test/internal/cucumber` -- Cucumber report parsing
- `core/logging` -- structured logging

## Role in System

The `reporter` package bridges raw test output files and the display/evidence layers. It collects test results from the `out/test/` directory structure, parses multiple formats, and produces unified data structures consumed by show commands and evidence generation.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
