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
- `loadModules` in `loader.go:24-173` manually maps `config.Module` fields to `domain.BaseContract` field-by-field (~100 lines of boilerplate); a shared conversion function or code generation would reduce drift risk
- `MatchesFile` in `types.go:109-151` hard-codes component names (`"static"`, `"book"`) as catch-all cases rather than using a component-type property

### Pain Points
- `GetUsedBy` in `registry.go:188` rebuilds the entire reverse dependency graph on every call; callers in hot loops pay O(n*m) per invocation
- `matchWithFallback` in `types.go:276-301` contains multiple overlapping `**` fallback branches that are difficult to reason about and extend

### Optimization Opportunities
- Pre-compute the reverse dependency graph in `Registry` at registration time and invalidate on `Add`; this is low effort and eliminates repeated O(n) scans
- `FindModulesForFile` in `registry.go:195-231` already uses a root-prefix index; no further optimization needed
