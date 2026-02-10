# test

Implements the `test` command, which discovers, filters, and executes tests across modules using suite-based selection and the command framework for parallel execution with incremental caching.

## Key Types

- **`TestConfig`** -- Top-level test execution configuration from CLI flags
- **`TestFrameworkConfig`** -- Framework-level config with suite, selected tests, and cache state
- **`TestSpecificFlags`** -- Parsed test-only CLI flags (suite, coverage, list-only)
- **`TestCacheVerifier`** -- UoW-level cache verifier for incremental test detection
- **`ModuleMapper`** -- Maps file paths and package paths to module monikers

## Patterns

- Suite-based filtering: Tests are tagged with suites and filtered by `--suite` flag against config-defined suite definitions
- Discovery-then-execution: Tests are discovered and grouped by package, then dispatched as work units for parallel execution
- Component-level parallelism: Each test package becomes a work unit with its own log file and UoW manifest
- Incremental detection: Source hash comparison skips test packages whose inputs have not changed
- Result merging: `merge-results` aggregates per-module test outputs into unified reports

## Internal Structure

| File | Responsibility |
| --- | --- |
| test.go | Entry point, CLI flag parsing, test configuration, `Test()` function |
| framework.go | Framework hooks setup and test execution orchestration |
| framework_hooks.go | `AfterInit`, `AfterResolve`, `AfterExecute` hook implementations |
| framework_selection.go | Suite-based test selection and filtering logic |
| testflags.go | `ParseTestSpecificFlags` for test-only flags |
| unit_work.go | `ResolveTestUnitSpecs` converts test packages to work unit specs |
| unit_worker.go | `testUnitWorker` executes a single test package, `TestCacheVerifier` |
| discovery.go | `groupTestsByPackage` groups tests by package path using runner strategies |
| module_mapping.go | `ModuleMapper` for file-to-module ownership resolution |
| incremental.go | `buildModuleTestInfo` for UoW-level change detection |
| summary.go | Test summary markdown generation from cucumber results |
| merge-results.go | `MergeResults` aggregates per-module test outputs |

## Dependencies

- `contracts/core` -- action type constants and tag filter interface
- `clibase/cmdframework` -- command lifecycle framework and hook registration
- `clibase/testrunners` -- test runner registry for type-specific execution
- `core/testing` -- test reference, suite, and discovery types
- `core/hash` -- input hash computation for incremental detection
- `core/output` -- UoW manifest tracking and artifact collection
- `core/tool` -- test handler bridge and tool execution
- `core/workunit` -- `UnitSpec`, `UnitID`, and test module info types

## Role in System

The test package is the second-largest command implementation in `eac`, parallel to the build command in structure. It discovers tests across the repository, applies suite-based filtering from configuration, maps tests to modules, and dispatches them as component-level work units through the command framework. Sub-packages handle result format parsing (cucumber, CTRF, test-json) and report generation, while the testers sub-package delegates actual test execution to handlers registered in the tool system.

## Code Health

### Tech Debt
- `testAfterResolve` is ~231 lines -- the largest single function in the CLI; it handles suite filtering, test discovery, module mapping, hash pre-computation, and incremental detection all inline
- `test.go` is still 673 lines combining entry point, configuration, and usage display

### Pain Points
- No test file for `framework.go`, `unit_worker.go`, `discovery.go`, `incremental.go`, `module_mapping.go`, or `summary.go`

### Optimization Opportunities
- Break `testAfterResolve` into focused sub-functions (discover, filter-suite, map-modules, detect-incremental) -- high impact on maintainability
