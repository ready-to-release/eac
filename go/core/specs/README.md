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
| steps_cache_context_test.go | Cache types, context, and step registration |
| steps_cache_setup_test.go | Repository setup and file manipulation |
| steps_cache_assertions_test.go | YAML and output assertions |
| steps_cache_ci_mock_test.go | CI mocking functions |
| steps_cache_state_test.go | Build/lint state management |
| steps_config_test.go | Config-defaults feature steps |
| steps_logging_test.go | Logging feature steps |
| steps_tool_test.go | Tool-system feature steps |

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
- Package-level vars for test state (`toolState`, `logCtx`, `cfgState`, `cacheCtx`) are reset per-scenario via Before/After hooks and documented as safe for sequential Godog execution
- `steps_config_test.go` is 595 lines, largest test file in package

### Pain Points
- None identified

### Optimization Opportunities
- None identified
