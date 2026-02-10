# testing/mocks

Provides mock implementations of all core port interfaces from the contracts layer. Every mock uses a fluent builder pattern for ergonomic test setup and includes compile-time interface compliance checks.

## Key Types

| Type | Purpose |
|------|---------|
| `MockLogger` | Implements `core.LoggerPort`; captures all log messages by level for test verification |
| `MockModule` | Implements `core.ModuleContractPort`; configurable module with components, dependencies, versioning, metadata |
| `MockModuleRegistry` | Implements `core.ModuleRegistryPort`; in-memory module registry with filtering and lookup |
| `MockModuleReport` | Implements `core.ModuleReportPort`; wraps a registry with optional loading errors |
| `MockWorkspace` | Implements `core.WorkspacePort`; configurable workspace root, source, container flag, dist root |
| `MockConfig` | Implements `core.ConfigPort`; configurable config with repository, environments, testing tags, test suites, component kinds |
| `MockRepositoryConfig` | Implements `core.RepositoryConfigPort`; module lookup with output path generation |
| `MockUnitID` | Implements `core.UnitIDPort`; configurable unit identity with context, module, component, tool, spec, extras |
| `MockUnitSpec` | Implements `core.UnitSpecPort`; configurable unit specification with dependencies and pool allocation |
| `MockUnitResult` | Implements `core.UnitResultPort`; configurable execution result with exit code, duration, log path |
| `MockUnitRegistry` | Implements `core.UnitRegistryPort`; in-memory unit registry with module/component/ID filtering |
| `MockPoolAllocation` | Implements `core.PoolAllocationPort`; configurable host and docker weights |

## Patterns

- **Fluent builder pattern**: All mocks use `With*` methods for chained configuration (e.g., `NewMockModule("m").WithGoComponent("go/m").WithDependsOn("dep")`)
- **Compile-time interface checks**: Each file includes `var _ core.Interface = (*Mock)(nil)` assertions
- **Sensible defaults**: Mocks initialize with reasonable defaults (e.g., `MockWorkspace` defaults to `/mock/workspace`)
- **Test verification**: `MockLogger` captures messages by level; `MockUnitResult` derives `Success()`/`Failed()`/`Cached()` from exit code

## Internal Structure

| File | Purpose |
|------|---------|
| `doc.go` | Package documentation with usage examples |
| `logger.go` | `MockLogger` implementing `core.LoggerPort` |
| `module.go` | `MockModule` implementing `core.ModuleContractPort` with full builder API |
| `registry.go` | `MockModuleRegistry` and `MockModuleReport` implementing registry/report ports |
| `unit.go` | `MockUnitID`, `MockUnitSpec`, `MockUnitResult`, `MockUnitRegistry`, `MockPoolAllocation` |
| `workspace.go` | `MockWorkspace` implementing `core.WorkspacePort` |
| `config.go` | `MockConfig` and `MockRepositoryConfig` implementing config ports |

## Dependencies

| Package | Purpose |
|---------|---------|
| `contracts/core` | Port interfaces (`LoggerPort`, `ModuleContractPort`, `ModuleRegistryPort`, `WorkspacePort`, `ConfigPort`, `UnitIDPort`, etc.) |

## Role in System

The standard test doubles for all packages that depend on core port interfaces. Used across the entire test suite to decouple tests from concrete implementations, enabling fast and focused unit testing without filesystem, git, or network dependencies.

## Code Health

- **Tech Debt**: None identified.
- **Pain Points**: None identified.
- **Optimization Opportunities**: None identified.
