# lint

Lints one or more modules in parallel using provider-specific linters, with incremental caching at the unit-of-work level, TUI progress display, and auto-fix support.

## Key Types

- **`LintConfig`** -- Lint-specific flags: fix mode, config path override, force lint bypass
- **`LintModuleResult`** -- Per-module lint outcome with issue counts, providers, and duration
- **`LintSpecificFlags`** -- Parsed lint-only flags (fix, config path) separated from shared flags
- **`lintContext`** -- Mutable state during execution: results map, cached modules/UoWs, input hashes, tracker

## Patterns

- Framework delegation: delegates to `cmdframework.Run` with lint-registered unit provider and worker
- Hook-based lifecycle: `AfterInit`, `AfterResolve`, and `AfterExecute` hooks customize framework phases
- UoW-level caching: incremental detection at component:provider granularity, not just module level
- Component resolver: uses `resolver.NewComponentResolver` to map modules to lintable component:provider pairs
- Concurrent results: mutex-protected results map aggregated across parallel component workers
- Manifest tracking: writes UoW manifests atomically via `InMemoryTracker` for cache validation

## Internal Structure

| File | Responsibility |
| --- | --- |
| lint.go | Command entry point, flag parsing, environment detection, usage display |
| framework.go | Framework integration: hooks, unit provider/worker registration, caching, deps verification |
| lintflags.go | Lint-specific flag parsing (--fix, --config) separate from shared flags |
| unit_work.go | Component resolution: maps modules to lintable `UnitSpec` work items |

## Dependencies

- `contracts/core` -- action type constants (`ActionLint`)
- `clibase/cmdframework` -- parallel execution framework with TUI and hooks
- `clibase/caching` -- incremental change detection for UoW-level cache
- `clibase/environment` -- execution environment detection
- `clibase/flags` -- shared flag parsing with environment awareness
- `clibase/initsummary` -- init summary and dependency status reporting
- `clibase/locking` -- component-level file locks for concurrent lint safety
- `clibase/output` -- formatted log writing to worker streams
- `clibase/registry` -- command registration
- `adapters/tui` -- TUI default height constant
- `core/config` -- global config and lint provider lookup
- `core/domain` -- module linting overrides
- `core/hash` -- input file hashing for cache keys
- `core/logging` -- structured logging and execution context
- `core/output` -- UoW manifest tracker and reader for cache validation
- `core/paths` -- lint output path constants
- `core/resolver` -- component-to-tool resolution for lint specs
- `core/tool` -- lint handler bridge, platform filtering, tool registry
- `core/workunit` -- `UnitID` and `UnitSpec` types for work scheduling

## Role in System

The `lint` package provides the `lint` command for `eac-cli`, running language-appropriate linters across all modules with parallel execution, incremental caching, and TUI feedback. It integrates with the `cmdframework` to share execution infrastructure with build, test, and scan commands while maintaining lint-specific behavior like auto-fix support and provider-per-component granularity.

## Code Health

### Tech Debt
- `lintUnitWorker` in framework.go (~201 lines, starting at line 185) is the largest function, handling tool resolution, execution, result aggregation, and manifest writing
- framework.go (556 lines) concentrates hooks, worker logic, caching, hashing, and dependency verification in one file
- Only component_work_test.go exists (82 lines); no tests for framework.go, lint.go, or lintflags.go

### Pain Points
- `lintUnitWorker` mixes tool execution with result parsing and manifest writing, making it hard to test individual concerns
- The helper `containsString` (framework.go:444) is a generic utility that could live in a shared package

### Optimization Opportunities
- Split `lintUnitWorker` into tool-execution and result-collection phases (high feasibility, clear separation point at result aggregation)
- Add unit tests for `ParseLintSpecificFlags` and `computeLintInputHash` to cover flag parsing and cache-key correctness (high feasibility, pure functions)
