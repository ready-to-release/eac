# godog

Godog BDD test infrastructure providing shared context, step definitions,
fixture management, and test runner configuration for Gherkin feature
specifications.

## Key Types

- **`TestContext`** -- Wraps `SharedTestContext` with godog-specific state
- **`TestCache`** -- Process-level repository data cache for test performance
- **`RunnerConfig`** -- Configuration for a spec runner instance
- **`GodogTagTranslator`** -- Translates `TagFilter` to godog tag syntax
- **`Template`** -- Named EAC configuration template for test fixtures
- **`TemplateParams`** -- Parameter map for template substitution

## Patterns

- Test isolation: Fixture pool with template-based fast-copy (~50ms)
- In-process dispatch: `MakeInProcessDispatcher` avoids subprocess overhead
- Global cache: Process-level `TestCache` shared across all test packages
- Template system: Named templates with `{{PLACEHOLDER}}` substitution
- Step composition: Helpers as pure functions, steps register in each verb

## Internal Structure

| File              | Responsibility                                              |
| ----------------- | ----------------------------------------------------------- |
| runner.go         | `RunnerConfig`, `CreateScenarioInitializer`, `BuildOptions` |
| context.go        | `TestContext` with isolation, command execution, mocking    |
| steps.go          | `RegisterCommonSteps` for shared step definitions           |
| helpers.go        | Composable helper functions for step implementations        |
| fixtures.go       | Test environment setup (modules, configs, git state)        |
| templates.go      | Named EAC configuration templates and substitution          |
| cache.go          | `TestCache` with git tracked files caching                  |
| cache_adapters.go | Port-interface adapter wrappers (moduleReportAdapter, moduleRegistryAdapter, moduleContractAdapter) |
| dispatcher.go     | In-process command dispatch for fast BDD execution          |
| tag_translator.go | `GodogTagTranslator` for godog tag expression syntax        |
| descriptor.go     | Test type descriptor registration for godog                 |
| doc.go            | Package documentation                                       |

## Dependencies

- `contracts/core/0.1.0` -- `TagFilter`, `ModuleReportPort` interfaces
- `core/config` -- EAC configuration loading
- `core/testing` -- `SharedTestContext`, `TestIsolation`, `FixturePool`
- `core/repository` -- repository root resolution
- `core/paths` -- binary path resolution
- `core/logging` -- structured logging
- `core/git` -- git operations for cache population
- `core/domain/modules` -- module registry adapters
- `clibase/testrunners` -- test type descriptor registration

## Role in System

The `godog-eac` module provides the shared BDD test infrastructure that
all module-specific spec implementations depend on. It sits between
`core/testing` (shared test primitives) and module specs, providing
godog-specific wiring including scenario initialization, test isolation,
fixture management, and common step definitions used across all feature
specifications.

## Code Health

### Tech Debt
- None identified

### Pain Points
- context.go is 543 lines; strong candidate for splitting (extract command execution, mocking, and assertion helpers into separate files)
- helpers.go is 373 lines; candidate for splitting by concern (command helpers, file helpers, assertion helpers)
- fixtures.go is 341 lines; candidate for splitting (extract template application and git state setup)
- cache.go is 292 lines; no immediate splitting needed but approaching threshold
- runner.go is 276 lines; no immediate splitting needed but approaching threshold
- templates.go is 251 lines; no immediate splitting needed but approaching threshold
- cache_adapters.go is 231 lines; no immediate splitting needed but approaching threshold
- Package-level var log in cache.go captures the component name at init time, though the underlying zap logger is resolved lazily on each call via logging.C()

### Optimization Opportunities
- None identified
