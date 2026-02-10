# dotnet

Test runner adapter for .NET xUnit tests, including TRX-to-CTRF result conversion.

## Key Types

- **`DotnetTestRunner`** -- implements `testrunners.TestTypeRunner` to discover, classify, and execute .NET xUnit tests via the `dotnet test` CLI
- **`TRXTestRun`** -- XML-deserialization root for Microsoft TRX test result files
- **`TRXResults`** -- container for the list of `TRXUnitTestResult` entries within a TRX file
- **`TRXUnitTestResult`** -- a single test outcome including name, outcome, duration, stdout/stderr, and error info
- **`TRXCounters`** -- aggregate pass/fail/skip/error counts from TRX
- **`TRXTimes`** -- creation, queuing, start, and finish timestamps from TRX

## Key Functions

- `ConvertTRXToCTRF` -- parses raw TRX XML bytes and returns a `*ctrf.Report` in the Common Test Report Format

## Patterns

- Self-registration via `init()`: the runner and its `TestTypeDescriptor` are registered with the global `testrunners` registry on import, so no manual wiring is needed
- Compile-time interface check (`var _ testrunners.TestTypeRunner = (*DotnetTestRunner)(nil)`) ensures the struct satisfies the contract
- Execution delegates to the `tool` package's global registry and executor for running `dotnet restore` and `dotnet test`, keeping the adapter decoupled from process management
- NuGet restore is serialized through `nuget.NuGetRestoreMu` to avoid Windows file-locking contention, while test execution itself remains parallel
- TRX parsing is a two-phase pipeline: XML unmarshal into typed structs, then field-by-field mapping to CTRF
- Fallback counting: if TRX parsing fails (file missing, corrupt), the runner synthesizes pass/fail counts from the input test list so results are never empty

## Internal Structure

| File | Responsibility |
| --- | --- |
| runner.go | `DotnetTestRunner` implementation: init registration, test-info extraction, module lookup, and `Execute` orchestration (restore, test, TRX parse, CTRF emit) |
| trx_types.go | XML struct definitions (`TRXTestRun`, `TRXUnitTestResult`, `TRXCounters`, `TRXTimes`) for deserializing Microsoft TRX files |
| trx_parser.go | `ConvertTRXToCTRF` and helpers (`mapTRXOutcomeToCTRF`, `parseTRXDuration`, `reformatTRXDuration`) that transform TRX data into a CTRF report |

## Dependencies

- `github.com/ready-to-release/eac/go/adapters/nuget` -- NuGet isolation and the shared restore mutex
- `github.com/ready-to-release/eac/go/clibase/testrunners` -- runner registration, `TestTypeRunner` interface, `TestTypeDescriptor`, `RunConfig`, `RunResult`
- `github.com/ready-to-release/eac/go/core/config` -- `EACConfig` for reading module/component definitions
- `github.com/ready-to-release/eac/go/core/testing` -- `TestReference` type used in discovery and execution
- `github.com/ready-to-release/eac/go/core/tool` -- global tool registry and executor for running `dotnet` CLI commands
- `github.com/ready-to-release/eac/go/core/ctrf` -- CTRF report model and builder (`NewReport`, `AddTest`, `SetTimes`, `Finalize`, `ToJSON`)

## Role in System

This package is one of the pluggable test-runner adapters in the eac test orchestration framework. It is imported for side-effects so that the `init()` function registers the `"dotnet"` test type, enabling eac to automatically discover .NET xUnit test projects, execute them through `dotnet test`, parse the native TRX output, and produce standardized CTRF JSON results.

## Code Health

### Tech Debt
- runner.go:180-202 manually pipes stdout/stderr and calls `cmd.Start`/`cmd.Wait`, while the restore step at line 144 uses the higher-level `tool.GlobalExecutor().Execute`. The test-execution step could use the same high-level API for consistency.
- trx_parser.go:77-83 `reformatTRXDuration` does naive string splitting without validating that each part is numeric; malformed durations silently produce "0s" via the downstream `time.ParseDuration` error path, but the intent is unclear.

### Pain Points
- None identified

### Optimization Opportunities
- None identified
