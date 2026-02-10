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

| File | Purpose |
|---|---|
| `services.go` | `Services` struct, `New()`, all private adapter structs implementing port interfaces |

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
- `services.go` (571 lines) contains ~15 private adapter structs in a single file; splitting adapters into separate files would improve navigability
- Several adapter methods return hardcoded zero values (e.g., `GetComponentTypesDisplay` returns `""`, `GetContentHash` returns `""`); these stubs may silently mask missing functionality
- `GetComponentAmp` always returns `1.0` regardless of input

### Pain Points
- The large number of adapter types implementing port interfaces makes the file dense; new port methods require updates across many adapters

### Optimization Opportunities
- Extract adapter structs into `adapters.go` or per-domain files (e.g., `config_adapter.go`, `module_adapter.go`) (low effort)
- Audit stub methods returning zero values and either implement or document them as intentionally unsupported (low effort)
