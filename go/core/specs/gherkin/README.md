# specs/gherkin

Parses Gherkin `.feature` files using the official `cucumber/gherkin-go` library and provides utilities for scenario filtering, tagging, ID generation, and collision detection.

## Key Types

| Type | Purpose |
|------|---------|
| `ScenarioDetail` | Complete scenario metadata: name, tags (merged feature + scenario), steps, feature name, file path, line number |

## Key Functions

| Function | Purpose |
|----------|---------|
| `ParseFile` | Parses a Gherkin feature file returning all scenarios with full details |
| `IsManualScenario` | Checks if a scenario has the `@Manual` tag |
| `HasTag` | Checks if a specific tag exists in a tag list |
| `FilterScenariosByTag` | Filters scenarios to those with a specified tag |
| `GenerateScenarioID` | Generates a unique `feature-name:scenario-name` ID from slugified names |
| `DetectIDCollisions` | Finds scenarios with duplicate generated IDs |
| `CombineTags` | Merges feature and scenario tags with deduplication |

## Patterns

- **Official parser delegation**: Uses `cucumber/gherkin-go` v26 for parsing, wrapping results in the simpler `ScenarioDetail` type
- **Tag inheritance**: Feature-level tags are combined with scenario-level tags on each `ScenarioDetail`
- **Rule support**: Handles Gherkin `Rule` blocks by extracting nested scenarios
- **Slugified IDs**: `GenerateScenarioID` creates deterministic `feature:scenario` identifiers using URL-safe slugification

## Internal Structure

| File | Purpose |
|------|---------|
| `types.go` | `ScenarioDetail` type definition |
| `parser.go` | `ParseFile`, `CombineTags`, and internal extraction functions for scenarios, tags, steps, data tables, doc strings |
| `scenarios.go` | `IsManualScenario`, `HasTag`, `FilterScenariosByTag`, `GenerateScenarioID`, `DetectIDCollisions` |

## Dependencies

None (uses `cucumber/gherkin-go` external library, no internal eac dependencies).

## Role in System

Foundation for BDD specification processing. Used by the `test export-manual` command (via `specs/export/formats`) to extract manual test scenarios, and by the Gherkin validation layer (`validation/formats/gherkin`) indirectly through shared concepts. Provides the parsed scenario data that flows through the export pipeline.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
