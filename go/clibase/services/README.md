# services

Service initialization layer that bridges CLI commands to core domain adapters.
Implements the `SimpleServicesPort` contract by assembling config, module, and tool adapters.

## Key Types

- `Services` -- aggregates all service adapters needed by CLI commands; implements `SimpleServicesPort` from the core contracts

## Patterns

- **Port implementation**: `Services` fulfills the `SimpleServicesPort` contract, providing a single initialization point for all domain adapters
- **Adapter composition**: internally creates config, repository, module, path, and tool adapters, exposing them through the port interface
- **Lazy initialization**: adapters are created on demand and cached for reuse within a command execution

## Internal Structure

| File | Purpose |
|---|---|
| `services.go` | `Services` struct with adapter creation and `SimpleServicesPort` implementation |

## Dependencies

- `contracts/core` -- `SimpleServicesPort` interface definition
- `core/adapters` -- concrete adapter implementations
- `core/config` -- EAC and repository configuration loading
- `core/domain/modules` -- module registry access
- `core/paths` -- workspace path resolution
- `core/tool` -- tool resolution and management
- `core/workspace` -- workspace root detection

## Role in System

Provides the dependency injection entry point for CLI commands. Commands obtain a `Services` instance to access config, modules, tools, and other domain services without directly depending on their concrete implementations.

## Code Health

### Tech Debt
- `services.go` (559 lines) contains ~15 private adapter structs in a single file; splitting adapters into separate files would improve navigability
- Several adapter methods return hardcoded zero values (e.g., `GetComponentTypesDisplay` returns `""`, `GetContentHash` returns `""`); these stubs may silently mask missing functionality

### Pain Points
- The large number of adapter types implementing port interfaces makes the file dense; new port methods require updates across many adapters

### Optimization Opportunities
- Extract adapter structs into `adapters.go` or per-domain files (e.g., `config_adapter.go`, `module_adapter.go`) (low effort)
- Audit stub methods returning zero values and either implement or document them as intentionally unsupported (low effort)
