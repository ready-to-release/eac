# testers

Provides test execution handler registration, shared test utilities, and static module test implementations. Acts as the bridge between the test framework and concrete test handlers.

## Key Types

- **`TestFunc`** -- Type alias for test handler functions (delegated from tool bridge)

## Key Functions

- **`GetTestFunc()`** -- Retrieve a registered test handler by name from the tool bridge
- **`HasHandler()`** -- Check if a test handler is registered for a given name
- **`Writeln()`** -- Write a line to stdout (utility for test output)
- **`RunTestCommand()`** -- Execute a test command as a subprocess with standard output
- **`RunTestCommandWithCapture()`** -- Execute a test command capturing stdout and stderr
- **`RunTestCommandWithEnv()`** -- Execute a test command with additional environment variables
- **`GenerateGherkinSummaryMarkdown()`** -- Generate a markdown summary from Gherkin test results
- **`FindModulesWithResults()`** -- Discover modules that have test result output files
- **`FormatDuration()`** -- Format a duration for human-readable display
- **`TestStaticModule()`** -- Test handler for static modules (non-Go, no test execution)
- **`TestMkDocsModule()`** -- Test handler for MkDocs documentation modules

## Patterns

- Tool bridge delegation: `GetTestFunc()` and `HasHandler()` delegate to `tool.GlobalTestBridge()`
- Subprocess execution: test commands run as child processes with configurable output capture
- Handler-per-module-type: different module types get different test handlers (static, mkdocs, go, etc.)

## Internal Structure

| File | Responsibility |
| --- | --- |
| registry.go | Test handler lookup delegating to `tool.GlobalTestBridge()` |
| helpers.go | Shared test utilities (command execution, markdown generation, module discovery, formatting) |
| static.go | Test handlers for static and MkDocs module types |

## Dependencies

- `adapters/tool` -- tool bridge for test handler registration
- `core/logging` -- structured logging

## Role in System

The `testers` package provides the test execution layer between the test framework and actual test runners. It offers handler lookup, shared utilities for subprocess execution and result formatting, and built-in handlers for non-Go module types.

## Code Health

### Tech Debt
- None identified.

### Pain Points
- None identified.

### Optimization Opportunities
- None identified.
