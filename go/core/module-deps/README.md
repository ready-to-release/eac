# module-deps

Verifies availability of internal module dependencies by checking build
artifacts or source roots.

## Key Types

| Type | Purpose |
|------|---------|
| `Checker` | Interface for verifying module availability and version |
| `Result` | Verification result with availability, version, and error |
| `ModuleChecker` | Checks if a module has been built or has source |

## Patterns

- Cross-platform artifact check: accepts any platform's build output for deps
- Fallback to source: when no build artifacts are defined, checks source root existence
- `@depm:` tag format: dependencies specified as `@depm:<moniker>` strings

## Internal Structure

| File | Purpose |
|------|---------|
| `types.go` | `Checker` interface, `Result` type |
| `verify.go` | `Verify`, `VerifyAll`, `ModuleChecker` with artifact resolution |

## Dependencies

| Package | Purpose |
|---------|---------|
| `core/config` | EAC config and artifact resolver |
| `core/domain/modules` | Module registry and contracts |
| `core/paths` | Build output path resolution |
| `core/repository` | Repository root detection |

## Role in System

The `module-deps` package is used by the test infrastructure and dependency
verification commands in `core`. It resolves `@depm:` tags to actual module
availability, enabling the system to skip tests whose module dependencies
have not been built.

## Code Health

- **Tech Debt**: `verify.go:229-230`: hardcoded platform/arch lists (`"linux", "windows", "darwin"`, `"amd64"`, `"arm64"`) should come from config.
- **Pain Points**: `GetVersion()` (line 192, ~70 lines) mixes version lookup with artifact resolution and platform iteration; splitting would improve clarity. `loadModuleContract()` calls `repository.GetRepositoryRoot("")` internally, creating hidden filesystem coupling.
- **Optimization Opportunities**: Extract platform/arch constants to a shared location to avoid drift between this package and build logic (low effort).
