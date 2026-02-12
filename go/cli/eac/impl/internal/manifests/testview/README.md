# testview

Provides UoW-based test data aggregation. Reads test UoW manifests and their referenced artifact files (Cucumber JSON, CTRF JSON, manual test files) to construct unified test result views with per-module, per-suite, and per-control breakdowns.

## Key Types

- **`TestModuleView`** -- Aggregated test data for a single module including timing, summary, suites, tests, and artifact references
- **`TestSummary`** -- Aggregate test counts (total, passed, failed, skipped)
- **`SuiteSummary`** -- Per-suite aggregate test counts
- **`TestEntry`** -- Single test result with name, module, package, type, suite, status, duration, and tags
- **`ArtifactRef`** -- Reference to a test output file located via UoW manifest
- **`CompleteTestData`** -- All aggregated test data for display/serialization (cross-module)
- **`TestResult`** -- Extended test entry with control tags, feature name, and feature path
- **`ModuleStats`** -- Per-module statistics including suite counts and control coverage
- **`SpecCoverage`** -- Test coverage for a feature file with scenario counts and control tags
- **`ControlSummary`** -- Aggregated test evidence for a security control across modules
- **`TypeSummary`** -- Test counts aggregated by test type
- **`SuiteSummaryEntry`** -- Test counts aggregated by suite

## Key Functions

- **`LoadModuleTestView()`** -- Read all test UoW manifests for a module and construct a `TestModuleView`
- **`LoadAllTestViews()`** -- Load test views for all modules that have test UoW manifests
- **`LoadTestViewsForModules()`** -- Load test views for specific modules including transitive dependencies
- **`BuildCompleteTestData()`** -- Aggregate all module views into cross-module `CompleteTestData`
- **`StripTagPrefixes()`** -- Remove `@` prefix from tags for display
- **`ExtractControlTags()`** -- Extract `@control:` tagged control IDs from tags array
- **`ExtractFeatureName()`** -- Extract feature name from a feature file path
- **`parseCucumberFile()`** -- Parse Cucumber JSON into `TestEntry` array
- **`parseCTRFFile()`** -- Parse CTRF JSON into `TestEntry` array
- **`parseManualTestFile()`** -- Parse manual test JSON into `TestEntry` array

## Patterns

- UoW manifest-driven loading: discovers test results through UoW manifest artifact references
- Multi-format parsing: handles Cucumber JSON, CTRF JSON, and manual test JSON formats
- Hierarchical aggregation: module-level views aggregate into cross-module complete data
- Suite derivation from tags: maps `@L0`/`@L1`/`@L2` tags to unit/integration/e2e suites
- Transitive dependency expansion: includes dependent module test results when loading for specific modules
- Control tag extraction: extracts `@control:` tags for security control evidence mapping

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | Core types (TestModuleView, TestSummary, TestEntry, ArtifactRef) and status/artifact constants |
| loader.go | UoW manifest loading, module view construction, and dependency expansion |
| aggregation.go | `BuildCompleteTestData` entry point for cross-module aggregation |
| aggregation_types.go | Cross-module aggregation types (`CompleteTestData`, `TestResult`, `ModuleStats`, `SpecCoverage`, `ControlSummary`, `TypeSummary`, `SuiteSummaryEntry`) |
| aggregation_builders.go | Builder functions for module stats, spec coverage, control summaries, type summaries, and suite summaries |
| aggregation_extractors.go | `extractTestResults` and `extractLastRun` from module views |
| helpers.go | Tag stripping, control tag extraction, and feature name extraction |
| ctrf_parser.go | CTRF JSON parsing into TestEntry array |
| cucumber_parser.go | Cucumber JSON parsing with scenario status derivation and suite mapping |
| manual_parser.go | Manual test JSON parsing into TestEntry array |

## Dependencies

- `contracts/core/0.1.0` -- action constants for UoW manifest reading
- `core/domain/modules` -- module registry for dependency expansion
- `core/output` -- UoW manifest reading
- `core/workunit` -- tag summary types

## Role in System

The `testview` package is the primary test data aggregation layer used by show and get commands for test results. It reads UoW manifests to discover test output files, parses multiple report formats, and produces unified views used by `show test-results`, `show test-summary`, `show ci-results`, `get test-results`, and evidence generation.

## Code Health

### Tech Debt
- aggregation_builders.go (233 lines) is the largest file; contains cross-module aggregation logic
- Most critical files now have dedicated _test.go coverage (loader, parsers, aggregation)

### Pain Points
- helpers.go (60 lines), aggregation_extractors.go (51 lines), and aggregation_types.go (88 lines) lack unit tests

### Optimization Opportunities
- None identified
