# specs

Provides BDD specification parsing, scenario export, and the Godog test
runner for the `core` module's own acceptance tests.

## Key Types

- **`ScenarioDetail`** -- Scenario with tags, steps, feature name, and line number
- **`ExportFormatter`** -- Interface for writing manual test exports (JSON, CSV, Markdown)
- **`ManualTestExport`** -- Root export structure with metadata and scenarios
- **`ExportedScenario`** -- Single scenario in an export with ID, tags, and steps

## Patterns

- Official Gherkin parser: uses `cucumber/gherkin-go` for spec-compliant parsing
- Pluggable export formats: `ExportFormatter` interface with factory selection
- Scenario ID generation: deterministic slug-based IDs with collision detection

## Internal Structure

| File/Sub-package | Responsibility |
| --- | --- |
| gherkin/parser.go | `ParseFile` using official Gherkin library |
| gherkin/types.go | `ScenarioDetail` type definition |
| gherkin/scenarios.go | Tag filtering, ID generation, collision detection |
| export/formats/formatter.go | `ExportFormatter` interface and export types |
| export/formats/json.go | JSON export formatter |
| export/formats/csv.go | CSV export formatter |
| export/formats/markdown.go | Markdown export formatter |
| export/formats/factory.go | Format selection factory |
| godog_test.go | Godog test runner for `core` module specs |
| steps_*.go | Step definitions for cache, config, tool, logging tests |

## Dependencies

- `cucumber/gherkin-go` -- official Gherkin parser (external)
- `cucumber/messages` -- Gherkin message types (external)

## Role in System

The `specs` package serves two purposes in `core`. The `gherkin/` sub-package
provides the parsing backend for `get specs` and `test-export-manual` commands.
The `export/` sub-package formats manual test scenarios for regulatory handoff.
The root-level test files run BDD acceptance tests for the `core` module itself.

## Code Health

### Tech Debt
- `steps_cache_test.go` is 1104 lines; consider splitting into focused step-definition files by cache scenario group
- Multiple package-level mutable vars for test state: `toolState` (steps_tool_test.go:33), `logCtx` (steps_logging_test.go:34), `cfgState` (steps_config_test.go:30), `cacheCtx` (steps_cache_test.go:45)

### Pain Points
- Test state globals (`cacheCtx`, `cfgState`, etc.) make parallel test execution unsafe; moving to struct-based contexts would improve isolation
- `steps_logging_test.go:20`: `loggingTestCounter` is a global `int64` used for unique naming; an atomic or per-test approach would be safer

### Optimization Opportunities
- Split `steps_cache_test.go` into smaller files by feature area (e.g., CI status, file changes, YAML parsing) to improve navigability (medium effort)
- Library sub-packages (`gherkin/`, `export/formats/`) are clean and well-tested; no changes needed there
