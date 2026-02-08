# core

Core domain contracts defining port interfaces, embedded JSON schemas, and
YAML defaults for the entire eac system.

## Key Types

- **`ConfigPort`** -- Read access to repository configuration
- **`LoggerPort`** -- Structured logging with levels and child loggers
- **`ExecutionObserver`** -- Receives execution events (observer pattern)
- **`UnitSpecPort`** -- Input specification for a schedulable work unit
- **`UnitIDPort`** -- Globally unique identity for a work unit
- **`ToolRegistryPort`** -- Lookup tool definitions by name or component
- **`ToolConfigPort`** -- Namespace-based tool configuration access
- **`ModuleRegistryPort`** -- Registry of module contracts by moniker
- **`ModuleContractPort`** -- Single module identity, components, and deps
- **`SimpleServicesPort`** -- Minimal service bundle for any command
- **`TestConfigPort`** -- Test suite and tag configuration access
- **`TUIHooks`** -- Interaction hooks for TUI lifecycle phases

## Patterns

- Hexagonal ports: all interfaces are `*Port` suffixes consumed by adapters
- Observer pattern: `ExecutionObserver` receives immutable event value types
- Null object: `NullTUIHooks` provides no-op defaults for console mode
- Embedded filesystem: `FS` bundles schemas and defaults via `//go:embed`
- Value types: `UnitID`, `UnitSpec`, `PoolAllocation`, `TagSummary` are concrete structs

## Internal Structure

| File / Sub-directory | Responsibility |
| --- | --- |
| doc.go | Package documentation |
| version.go | Module version constant |
| config.go | `ConfigPort` and related config port interfaces |
| logging.go | `LoggerPort` and `LoggerFactoryPort` |
| repository.go | `WorkspacePort`, `RepositoryPort`, `GitRepositoryPort` |
| services.go | `SimpleServicesPort` aggregate |
| modules.go | `ModuleRegistryPort`, `ModuleContractPort`, `ComponentTypePort` |
| observer.go | `ExecutionObserver`, event types, `WriterFactory` |
| units.go | `UnitIDPort`, `UnitSpecPort`, `UnitRegistryPort`, `UnitResolverPort` |
| workunit.go | `UnitID` and `UnitSpec` concrete value types, `TagSummary` |
| tools.go | `ToolRegistryPort`, `ToolConfigPort`, `ToolDefPort`, namespace types |
| tool_types.go | `ToolDefinition`, `ToolAssignment` concrete config types |
| output.go | `OutputReaderPort`, `UoWManifestPort`, validation ports |
| output_buffer.go | `OutputBufferPort` for TUI stdout/stderr capture |
| action_type.go | `ActionType` enum and `ActionDescriptor` |
| tui_hooks.go | TUI hook interfaces and null implementations |
| testing.go | `TestCachePort`, `TestIsolationPort`, `TagFilter` |
| testing_types.go | `TestConfigPort`, `SuitePort`, `TagPort`, concrete definitions |
| embed.go | `FS` embedded filesystem with helper path functions |
| schemas/ | JSON Schema files for config validation |
| schemas/defaults/ | YAML default configurations |

## Dependencies

None -- this is a leaf contract module with no internal dependencies.

## Role in System

The `core` package is the foundational contract module (moniker: contracts-core)
that every other module in the repository depends on. It defines all port
interfaces that adapters implement and commands consume, plus the embedded
schemas and defaults that drive configuration loading and validation.

## Code Health

### Tech Debt
- `ToolDefPort` in tools.go has 24 methods -- consider splitting into identity, execution, and config facets
- `ToolDefinitionPort` (12 methods) and `ToolConfigAssignmentPort` (8 methods) are also large; role interfaces would reduce adapter burden
- `UnitSpec.Container` field is marked DEPRECATED in workunit.go:381 but still present alongside `PoolAllocation`

### Pain Points
- `actionRegistry` in action_type.go is package-level mutable state; make it a frozen init-time constant or func
- No test coverage for workunit.go value types (`UnitID`, `TagSummary`, `DisplayNameResolver`) despite complex formatting logic
- `TestCachePort` (8 methods in testing.go) mixes caching with file querying -- split into cache lifecycle and file query ports

### Optimization Opportunities
- Extract `ToolDefPort` into 3-4 role interfaces (identity, container, system, serve) -- high impact, moderate effort
- Add unit tests for `UnitID.DisplayName()` and `Longname()` edge cases -- low effort, high confidence gain
- Replace `actionRegistry` var with a `const`-friendly lookup function -- trivial change
