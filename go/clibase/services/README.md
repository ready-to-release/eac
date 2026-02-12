# services

Service initialization layer that bridges CLI commands to core domain adapters.
Implements the `SimpleServicesPort` contract by assembling config, module, and tool adapters.

## Key Types

- `Services` -- aggregates all service adapters needed by CLI commands; implements `SimpleServicesPort` from the core contracts; provides `WorkspaceRoot()`, `ConfigRoot()`, `Config()`, `Modules()`, `Tools()`, `RawConfig()`, `AddCleanup()`, and `Close()`

## Key Functions

- `New` -- creates a `Services` instance: detects workspace, loads EAC config, builds module registry, optionally initializes tool registry, and tracks cleanup functions

## Patterns

- **Port implementation**: `Services` fulfills the `SimpleServicesPort` contract, providing a single initialization point for all domain adapters
- **Adapter composition**: internally creates ~15 private adapter structs (configAdapter, repositoryAdapter, moduleAdapter, toolRegistryAdapter, etc.) that wrap concrete types to satisfy port interfaces
- **Cleanup lifecycle**: `AddCleanup` registers deferred functions; `Close()` runs them in reverse order and is idempotent
- **Compile-time interface checks**: `var _ core.SimpleServicesPort = (*Services)(nil)` and similar assertions ensure adapter compliance

## Internal Structure

| File                       | Purpose                                                          |
| -------------------------- | ---------------------------------------------------------------- |
| `services.go`              | `Services` struct, `New()`, and core service initialization      |
| `services_config.go`       | Config-domain adapter structs (configAdapter, repositoryAdapter) |
| `services_components.go`   | Component-domain adapter structs (componentAdapter, etc.)        |
| `services_environments.go` | Environment-domain adapter structs                               |
| `services_tools.go`        | Tool-domain adapter structs (toolRegistryAdapter, etc.)          |
| `services_testing.go`      | Test-domain adapter structs                                      |

## Dependencies

- `contracts/core` -- `SimpleServicesPort`, `ConfigPort`, `ModuleRegistryPort`, `ToolRegistryPort`, and related port interfaces
- `core/adapters` -- concrete adapter for module registry
- `core/config` -- EAC and repository configuration loading
- `core/domain/modules` -- module registry access via `LoadFromWorkspace`
- `core/paths` -- workspace path resolution (build, test, scan, lint output paths)
- `core/tool` -- tool resolution and management via `GlobalRegistry()`
- `core/workspace` -- workspace root detection via `Detect()`

## Role in System

Provides the dependency injection entry point for CLI commands. Commands obtain a `Services` instance to access config, modules, tools, and other domain services without directly depending on their concrete implementations.

## Code Health

### Tech Debt

- `GetComponentAmp` in `services_config.go` line 175 always returns `1.0` regardless of input
- Multiple adapter methods return hardcoded empty values: `GetComponentTypesDisplay()`, `GetContentHash()`, `GetRoot()` in bookAdapter
- Several methods in `services_config.go` return `nil` or empty strings without implementation

### Pain Points

- `services_test.go` (599 lines) exceeds 300 lines and is significantly larger than any implementation file
- `services_config.go` (224 lines) contains multiple stub adapter methods
- New port interface methods require updates across multiple adapter files (`services_config.go`, `services_components.go`, `services_tools.go`, etc.)

### Optimization Opportunities

- None identified
