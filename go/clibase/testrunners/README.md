# testrunners

Registry-based test type dispatch system. Each test framework (gotest, godog, tscucumber, mocha)
registers a runner implementation, and the registry resolves which runner handles each test type.

## Key Types

- `TestTypeRunner` -- interface (aliased from `test.TestRunnerPort`) for test type-specific runners: discovery, metadata extraction, and execution
- `TestTypeDescriptor` -- metadata about a test type: BDD flag, component type, moniker style, runner file conventions, and tag inference rules
- `RunConfig` -- configuration for test execution: workspace root, coverage, tag filters, parallelism, output directory
- `RunResult` -- results from a test package execution: pass/fail/skip counts, duration, log file path
- `TestInfo` -- structured metadata for a test reference: module, language, display name, test root
- `TestReference` -- reference to a discovered test with module and component context
- `StreamingRunner` -- executes `go test -json` with real-time output parsing and event collection
- `TestEvent` -- single event from `go test -json` output stream
- `TestResult` -- aggregated test results with event list and duration
- `Inference` -- tag inference rule contributed by a test type adapter
- `FeatureModuleInfo` -- module metadata for resolving which BDD type owns feature files

## Key Functions

- `Register` -- registers a test type runner for one or more test types (called from `init()`)
- `RegisterFallback` -- registers a fallback runner for unknown test types
- `RegisterDescriptor` -- registers test type descriptor metadata
- `Get` -- returns the runner for a specific test type
- `GetAll` -- returns all registered runners
- `SupportedTypes` -- returns a list of all registered test types
- `AllDescriptors` -- returns all registered descriptors (deduplicated)
- `ResolveFeatureTestType` -- determines which BDD runner owns feature files for a module
- `CollectInferences` -- returns all default tag inferences from all registered types
- `BDDComponentNames` -- returns component names used by BDD test types
- `NewStreamingRunner` -- creates a streaming test runner for real-time `go test -json` parsing

## Patterns

- **Plugin registry**: runners self-register via `init()`, and the registry dispatches by test type string
- **Type aliasing for migration**: types are aliased from `contracts/runner/0.1.0/test` for backward compatibility; new code should import directly from the contracts package
- **BDD type resolution**: `ResolveFeatureTestType` determines which BDD runner (godog vs tscucumber) owns a module's feature files based on module characteristics
- **Provider bridging**: `init()` wires registry functions as providers for `core/testing`, bridging the dependency gap without import cycles
- **Streaming JSON parsing**: `StreamingRunner` parses `go test -json` output line-by-line, sending human-readable output to TUI while collecting events for result calculation

## Internal Structure

| File | Purpose |
|---|---|
| `registry.go` | Type aliases, global registry with `Register`/`Get`/`AllDescriptors`, provider bridging via `init()` |
| `streaming.go` | `StreamingRunner` for real-time `go test -json` output parsing and result aggregation |

## Dependencies

- `contracts/runner` -- `TestRunnerPort`, `TestTypeDescriptor`, and related types (canonical definitions)
- `core/testing` -- provider interfaces that this package implements via `init()` bridging

## Role in System

Decouples the test command from specific test frameworks. The test command resolves test references to their types, looks up the appropriate runner from this registry, and delegates execution. This enables adding new test frameworks (e.g., a new BDD runner) without modifying the test command itself.

## Code Health

### Tech Debt
- `registry.go:29-32` four mutable package-level vars (`runners`, `descriptors`, `fallback`, `mu`); protected by mutex but still global state

### Pain Points
- `streaming.go` has no dedicated test file; the JSON-streaming parser is a critical path that should have unit tests

### Optimization Opportunities
- Add unit tests for `StreamingRunner` JSON parsing edge cases (malformed lines, partial output) (low effort)
