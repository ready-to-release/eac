# nuget

Isolated NuGet environments for safe, parallel .NET test and build execution.

## Key Types

- **`NuGetIsolation`** -- manages the lifecycle of isolated work directories and a shared NuGet package cache, rooted under `.cache/eac/nuget/`
- **`IsolatedEnv`** -- represents a fully prepared isolated environment: the work directory, environment variables (with `NUGET_PACKAGES` set), the original source root, and any discovered `.sln` file

## Key Functions

- `NewNuGetIsolation` -- constructs a `NuGetIsolation` for a given workspace root, resolving cache paths via the `paths` package

## Patterns

- Copy-on-write isolation: project files (`.csproj`, `.sln`, `.props`, `.targets`, `nuget.config`, `global.json`) and source directories (`src/`, `test/`, `tests/`) are copied into a per-unit-of-work subdirectory so that `dotnet restore` and `dotnet build` never mutate the original tree
- Incremental sync: `copyFileIfChanged` compares mtime and size before copying, and `syncDirectory` removes stale files from the destination, keeping isolation cheap on repeat runs
- Change-triggered reset: `projectFilesChanged` uses `filepath.Walk` to recursively detect when `.csproj`/`.fsproj`/`.sln`/`Directory.Build.props` have changed (including nested subdirectories) and triggers a full directory wipe before re-syncing, preventing stale build state
- Shared NuGet package cache: all isolated environments point `NUGET_PACKAGES` at a single `.cache/eac/nuget/packages` directory so packages are downloaded once
- Global restore mutex (`NuGetRestoreMu`): a package-level `sync.Mutex` that callers use to serialize `dotnet restore` invocations, preventing NuGet cache corruption from concurrent writes on Windows

## Internal Structure

| File | Responsibility |
| --- | --- |
| isolation.go | All package code: `NuGetIsolation` and `IsolatedEnv` types, `PrepareIsolatedEnv` orchestration, file-sync helpers (`syncDotnetProjectFiles`, `syncDirectory`, `copyFileIfChanged`, `projectFilesChanged`, `discoverSlnFile`), and the `NuGetRestoreMu` mutex |

## Dependencies

- `github.com/ready-to-release/eac/go/core/fileutil` -- `RemoveAllWithRetry` for reliable directory cleanup on Windows (handles EPERM retries)
- `github.com/ready-to-release/eac/go/core/paths` -- `NuGetWorkCachePath` and `NuGetPackageCachePath` for resolving cache directory locations

## Role in System

This package is a shared infrastructure adapter used by the `dotnet` and `reqnroll` test runners. It solves a Windows-specific problem where parallel `dotnet restore` and `dotnet build` operations corrupt NuGet caches and fail with EPERM errors on locked `bin/obj` files. By copying project files into isolated work directories and serializing restores, it enables the test orchestrator to run multiple .NET test suites in parallel safely.

## Code Health

### Tech Debt
- None identified

### Pain Points
- isolation.go is 289 lines; no immediate splitting needed but approaching threshold

### Optimization Opportunities
- None identified
