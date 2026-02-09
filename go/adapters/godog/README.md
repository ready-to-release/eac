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

| File/Sub-package | Responsibility |
| --- | --- |
| runner.go | `RunnerConfig`, `CreateScenarioInitializer`, `BuildOptions` |
| context.go | `TestContext` with isolation, command execution, mocking |
| steps.go | `RegisterCommonSteps` for shared step definitions |
| helpers.go | Composable helper functions for step implementations |
| fixtures.go | Test environment setup (modules, configs, git state) |
| templates.go | Named EAC configuration templates and substitution |
| cache.go | `TestCache` with git tracked files caching |
| dispatcher.go | In-process command dispatch for fast BDD execution |
| tag_translator.go | `GodogTagTranslator` for godog tag expression syntax |
| descriptor.go | Test type descriptor registration for godog |

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
- cache.go bundles ~200 lines of port-interface adapter wrappers (moduleReportAdapter, moduleRegistryAdapter, moduleContractAdapter) that could live in a dedicated `adapters` sub-file
- `buildMockingEnvironment` in context.go (~60 lines) mixes config loading, path resolution, and env-var construction in one method

### Pain Points
- `logBinaryNotFoundDiagnostics` in context.go writes directly to `os.Stderr` instead of using the structured logger, making output hard to capture in CI log aggregation
- Package-level `var log` in cache.go ties all test infrastructure to one logger instance at import time

### Optimization Opportunities
- `TestCache.FilesByExtension`, `FilesBySuffix`, and `FilesInDir` each iterate the full tracked-files list; pre-building an extension-keyed index during `EnsurePopulated` would turn O(n) lookups into O(1); feasible with low memory overhead
