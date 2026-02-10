# reqnroll

Test runner adapter for .NET Reqnroll BDD (Behaviour-Driven Development) tests.

## Key Types

- **`ReqnrollRunner`** -- implements `testrunners.TestTypeRunner` to discover, classify, and execute Reqnroll BDD tests that are backed by .NET xUnit and driven by Gherkin `.feature` files

## Key Functions

- None beyond the `ReqnrollRunner` methods.

## Patterns

- Self-registration via `init()`: the runner and its `TestTypeDescriptor` (with `IsBDD: true`, `ComponentType: "gherkin"`, `MonikerStyle: "feature"`) are registered on import, enabling automatic BDD test discovery
- Feature-to-test-type resolution: a `FeatureTestTypeResolver` callback checks `info.HasDotnet` to claim ownership of `.feature` files only when a .NET project is present, avoiding false matches in non-.NET repositories
- Reuses the `dotnet` adapter's TRX parser (`dotnetadapter.ConvertTRXToCTRF`) rather than duplicating TRX parsing logic
- NuGet isolation via the `nuget` package: restores are serialized through `nuget.NuGetRestoreMu`, and tests run inside an isolated work directory
- Feature filtering: when a specific `.feature` file is provided, the runner appends `--filter FeatureTitle~<name>` to scope `dotnet test` to that feature only
- Compound package keys: uses a `featureFolder:testRoot:featurePath` triple as the package key to group and display BDD results by feature
- Compile-time interface check ensures `ReqnrollRunner` satisfies `testrunners.TestTypeRunner`

## Internal Structure

| File | Responsibility |
| --- | --- |
| runner.go | Complete `ReqnrollRunner` implementation: init registration with BDD descriptor, `GetTestInfo` for mapping feature files to modules, `FindTestRoot` for locating the .NET test project, `BuildPackagePath` for composite key construction, and `Execute` for orchestrating restore/test/TRX-parse/CTRF-emit |

## Dependencies

- `github.com/ready-to-release/eac/go/adapters/dotnet` -- `ConvertTRXToCTRF` for reusing TRX-to-CTRF conversion
- `github.com/ready-to-release/eac/go/adapters/nuget` -- `NewNuGetIsolation`, `NuGetRestoreMu` for isolated .NET environments and serialized restores
- `github.com/ready-to-release/eac/go/clibase/testrunners` -- runner registration, `TestTypeRunner` interface, `TestTypeDescriptor`, `RunConfig`, `RunResult`
- `github.com/ready-to-release/eac/go/core/config` -- `EACConfig` for module/component lookup and specs-root resolution
- `github.com/ready-to-release/eac/go/core/testing` -- `TestReference` type
- `github.com/ready-to-release/eac/go/core/tool` -- global tool registry and executor for running `dotnet` commands

## Role in System

This package is the BDD counterpart to the `dotnet` adapter. Where `dotnet` handles plain xUnit unit tests, `reqnroll` handles Gherkin-driven BDD acceptance tests that use the Reqnroll framework (the .NET successor to SpecFlow). It is imported for side-effects so that the `"reqnroll"` test type becomes available in the test orchestrator, allowing eac to map `.feature` files to their backing .NET test projects, execute them, and produce standardized CTRF results.

## Code Health

### Tech Debt
- runner.go:112-117 `FindTestRoot` returns the first component root it finds regardless of component type, which could be incorrect if a module has multiple components (e.g., both a `dotnet` and a `gherkin` component).

### Pain Points
- None identified

### Optimization Opportunities
- None identified
