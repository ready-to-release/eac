# tool

Unified, pluggable tool composition system that enables any tool (system
binary or Docker container) to be assigned to any component type for any
operation (build, test, lint, scan, serve).

## Key Types

- **`ToolDefinition`** -- Describes a tool: binary, image, mounts, env, resources
- **`ToolConfig`** -- Complete YAML-driven tool configuration file
- **`ToolAssignment`** -- Maps operations to tool IDs for a component type
- **`Registry`** -- Interface for storing and retrieving tool definitions
- **`DefaultRegistry`** -- Thread-safe in-memory registry with binding resolution
- **`Executor`** -- Interface for executing tools (system or container)
- **`DefaultExecutor`** -- Handles system binary and container execution
- **`Resolver`** -- Layered config resolution (CLI > env > project > defaults)
- **`DefaultResolver`** -- Concrete resolver with environment detection
- **`HandlerToolBridge`** -- Adapts tool definitions for handler execution

## Patterns

- Dual-key registry: canonical name + type suffix for flexible lookup
- Binding resolution: auto/system/container modes drive tool selection
- Layered configuration: CLI overrides > environment > project > defaults
- Lazy container init: Docker runtime initialized on first container use
- Category resolution: scanner/server categories map to tool IDs

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `ToolDefinition`, `ToolAssignment`, `ExecutionContext`, `ExecutionResult` |
| registry.go | `DefaultRegistry` with dual-key storage and verification |
| executor.go | `DefaultExecutor` for system and container execution |
| command.go | `BuildCommand` for pipe-based streaming execution |
| resolver.go | `DefaultResolver` with layered configuration |
| handler_bridge.go | `HandlerToolBridge` for handler-to-tool adaptation |
| config.go | `LoadToolConfig`, YAML parsing, config merging and validation |
| categories.go | `CategoryResolver` for scanner/server category mapping |
| ids.go | Tool ID constants for all known tools |
| global.go | Global singletons for config, test type mapping |

## Dependencies

- `contracts/core` -- `ActionType` constants
- `contracts/container-runtime` -- `ContainerPort` interface
- `core/environments` -- debug environment variable names
- `core/paths` -- container repo root constant
- `core/testing` -- test type component registry provider

## Role in System

This package is the execution engine of the `core` module. Every build,
test, lint, scan, and serve command resolves a `ToolDefinition` through
the registry and resolver, then executes it via the executor. The YAML
configuration allows projects to override tool assignments per component
type and environment without changing code.

## Code Health

### Tech Debt
- None identified

### Pain Points
- `types.go` is 1031 lines, largest non-test file in package
- `executor.go` is 689 lines
- `config.go` is 449 lines
- `registry.go` is 536 lines
- `handler_bridge.go` is 382 lines
- `image.go` is 384 lines
- `resolver.go` is 365 lines
- `build_bridge.go` is 345 lines
- `categories.go` is 193 lines, but well-tested (342-line test file)

### Optimization Opportunities
- Consider splitting `types.go` into focused files (e.g., tool definitions, execution context, results)
- Consider splitting `executor.go` by execution mode (system vs container)
