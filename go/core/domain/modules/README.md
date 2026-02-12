# modules

Module type system that wraps base domain contracts with workspace context.
Provides a registry for indexed lookup, dependency traversal, and file
ownership matching across all declared modules.

## Key Types

- `ModuleContract` embeds `BaseContract` and adds workspace root for absolute path resolution
- `Registry` stores modules indexed by moniker and by root path for fast lookup
- `NewModuleContract` constructs a module contract with its workspace context
- `NewRegistry` creates an empty registry for a given workspace root

## Patterns

- `ModuleContract` embeds `domain.BaseContract` via struct embedding, adding only `workspaceRoot`
- `Registry` maintains parallel indexes: by-moniker map and by-root map for different query paths
- File ownership uses doublestar glob matching with fallback handling for `**` patterns
- `GetGlobPatterns` collects patterns from all component types for GitHub Actions path filters
- `GetContentHash` delegates to `core/hash` for content-addressed caching of module files
- Path helpers (`GetSpecsRoot`, `GetDesignPath`, `GetCIWorkflowPath`) follow convention-over-configuration defaults
- `LoadFromWorkspace` converts `config.Module` values to domain `ModuleContract` instances through the config loader
- `ValidateRegistry` checks referential integrity (dependency targets must exist)

## Internal Structure

| Path | Purpose |
|------|---------|
| `types.go` | `ModuleContract` with glob matching, path derivation, and content hashing |
| `registry.go` | `Registry` with indexed lookup, filtering, and dependency graph |
| `loader.go` | `LoadFromWorkspace` entry point, config-to-domain conversion |
| `test_yaml.go` | Debug utility for YAML round-trip testing |

## Dependencies

- `core/domain` for `BaseContract` and `ContractError`
- `core/config` for `EACConfig`, `LoadOptions`, and config loading orchestration
- `core/hash` for content-addressed hashing via glob patterns

## Role in System

This is the primary domain model that commands and adapters interact with.
CLI commands load a `Registry` through `LoadFromWorkspace`, then query it
by moniker, root path, or component type. The `ModuleContract` methods
provide all path conventions (specs, design, workflows, changelogs) that
downstream tools need without re-implementing path logic.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- None identified
