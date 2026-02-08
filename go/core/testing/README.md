# testing

Provides test discovery, tag inference, suite selection, and BDD test isolation
infrastructure for the repository's multi-framework test system.

## Key Types

- **`TestReference`** -- Discovered test with tags, deps, and provenance
- **`TestSuite`** -- Suite definition with tag selectors and inferences
- **`SharedTestContext`** -- Shared state for BDD step definitions
- **`TestIsolation`** -- Isolated git repo environment for BDD tests
- **`FixturePool`** -- Reusable test environment templates for speed
- **`SuiteReport`** -- Complete suite report with validation results
- **`SuiteTestEntry`** -- Single test entry in a suite report
- **`DiscoveryOptions`** -- Configures enrichment for unified discovery

## Patterns

- Unified discovery pipeline: discover, infer tags, enrich deps in one call
- Provider registry: adapters register test-type metadata without circular imports
- Tag override semantics: scenario-level L-tags and verification tags override feature-level
- Fixture pooling: create template once, fast-copy per scenario (~50ms vs ~600ms)

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `TestReference`, `TestSuite`, `TagSelector`, `Inference` |
| discovery.go | Multi-framework test discovery (Go, Gherkin, TypeScript) |
| inference.go | Tag inference and system-dep enrichment rules |
| suite.go | Suite retrieval, composite suites, test selection |
| validation.go | Post-inference tag validation and GxP checks |
| context.go | `SharedTestContext` for BDD step definitions |
| isolation.go | `TestIsolation` for isolated git repos in tests |
| fixture.go | `FixturePool` for reusable test templates |
| providers.go | Provider registry for test-type metadata |
| moniker.go | Test moniker generation (BDD and Go styles) |
| reports.go | GxP and traceability matrix report generation |
| suite_report.go | `SuiteReport` and `GenerateSuiteReport` |
| conversion.go | `TestReference` to `SuiteTestEntry` conversion |
| tag_aggregation.go | Union tag summary across test sets |
| feature_parser.go | Gherkin feature file metadata extraction |
| helpers.go | Git-related test skip helpers |
| workspace_helpers.go | Workspace isolation helpers for unit tests |
| mock_registry.go | Mock module registry builders for testing |

## Dependencies

- `core/config` -- global and testing configuration
- `core/domain/modules` -- module registry and contracts
- `core/workspace` -- workspace detection for test isolation
- `core/domain` -- base contract types
- `core/paths` -- EAC config and mock file paths
- `core/logging` -- structured logging
- `core/workunit` -- `TagSummary` type for aggregation
- `contracts/core` -- `SuitePort`, `TagFilter` port interfaces

## Role in System

The `testing` package is the central test infrastructure in `core`. It powers
the `get suite`, `show suite`, `get tests`, and `validate test-tags` commands
by discovering tests across Go, Gherkin, and TypeScript, applying inference
rules, and selecting tests by suite. BDD acceptance tests use its isolation
and fixture-pooling facilities.

## Code Health

### Tech Debt
- `discovery.go:156`: TODO comment -- "Handle complex expressions if needed" for tag expression parsing
- `discovery.go:297`: `discoverModuleAllTests()` is ~110 lines; could split per-component discovery into a helper
- `discovery.go:20`: package-level `var log` global mutable logger

### Pain Points
- `discovery.go` is 951 lines covering Go, Gherkin, and TypeScript discovery in one file; splitting by language would improve navigability
- `validation.go` is 551 lines with deeply nested logic; could benefit from smaller validation step functions

### Optimization Opportunities
- Extract language-specific discovery (Go, Gherkin, TypeScript) into sub-files to reduce cognitive load (medium effort)
- Add benchmark tests for fixture pooling to guard against performance regressions (low effort, high value)
