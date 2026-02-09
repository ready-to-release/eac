# module-deps

Verifies availability of internal module dependencies by checking build
artifacts or source roots.

## Key Types

- **`Checker`** -- Interface for verifying module availability and version
- **`Result`** -- Verification result with availability, version, and error
- **`ModuleChecker`** -- Checks if a module has been built or has source

## Patterns

- Cross-platform artifact check: accepts any platform's build output for deps
- Fallback to source: when no build artifacts are defined, checks source root existence
- `@depm:` tag format: dependencies specified as `@depm:<moniker>` strings

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `Checker` interface, `Result` type |
| verify.go | `Verify`, `VerifyAll`, `ModuleChecker` with artifact resolution |

## Dependencies

- `core/config` -- EAC config and artifact resolver
- `core/domain/modules` -- module registry and contracts
- `core/paths` -- build output path resolution
- `core/repository` -- repository root detection

## Role in System

The `module-deps` package is used by the test infrastructure and dependency
verification commands in `core`. It resolves `@depm:` tags to actual module
availability, enabling the system to skip tests whose module dependencies
have not been built.

## Code Health

### Tech Debt
- ~~No unit tests exist for this package (0 test files); `Verify`, `VerifyAll`, and `ModuleChecker` are untested~~ (resolved: `verify_test.go` has 39 tests covering format validation, public API, `checkSourceRootExists`, `checkAnyPlatformExists` with platform/arch edge cases, and `init()` idempotency)
- `verify.go:229-230`: hardcoded platform/arch lists (`"linux", "windows", "darwin"`, `"amd64"`, `"arm64"`) should come from config

### Pain Points
- `GetVersion()` (line 192, ~70 lines) mixes version lookup with artifact resolution and platform iteration; splitting would improve clarity
- `loadModuleContract()` calls `repository.GetRepositoryRoot("")` internally, creating hidden filesystem coupling

### Optimization Opportunities
- ~~Add unit tests with a mock `ModuleRegistry` to cover artifact resolution paths~~ (resolved: covered in verify_test.go with filesystem-based happy-path and edge-case tests)
- Extract platform/arch constants to a shared location to avoid drift between this package and build logic (low effort)
