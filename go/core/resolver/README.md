# resolver

Unified component-to-tool resolution for build, lint, and scan commands,
mapping module components to executable work units with dependency ordering
and scheduling weights.

## Key Types

- **`ComponentResolver`** -- Resolves module components to `UnitSpec` slices for build, lint, and scan phases
- **`ResolverPort`** -- Adapts `ComponentResolver` to implement `core.UnitResolverPort` for dependency injection
- **`DependencyGraph`** -- Manages component build order within a module using `build_after` relationships
- **`Phase`** -- Enum for command phases (build, lint, test, scan)
- **`ScanCategory`** -- Enum for scanner categories (sbom, vuln, secrets, sast, iac, compliance, zap)
- **`ToolAssignment`** -- Resolved tool for a component and phase with handler and weight
- **`ExtendedComponentType`** -- Extends component type with phase-specific tool mappings
- **`PhaseConfig`** -- YAML-driven tool configuration for a single build/lint/test phase
- **`ScanPhaseConfig`** -- Scan-specific configuration with category-to-tool mappings
- **`ComponentTools`** -- Phase-to-tool mappings for a component type from tool-config.yml

## Patterns

- Phase-based resolution: separate methods for build, lint, and scan with phase-specific logic
- Dependency graph: `build_after` relationships resolved via DFS cycle detection and topological sort
- Tool chain expansion: multi-tool build sequences expanded into chained `UnitSpec` values
- Weight calculation: base weight from tool resources, amplified by component type and module factors
- Port adapter: `ResolverPort` wraps concrete resolver behind contract interface
- Global singleton: `GlobalUnitResolver` provides thread-safe shared instance

## Internal Structure

| File | Responsibility |
| --- | --- |
| types.go | `Phase`, `ScanCategory`, `ToolAssignment`, `ExtendedComponentType`, config types |
| component_resolver.go | `ComponentResolver` with `ResolveForBuild`, `ResolveForLint`, `ResolveForScan` |
| dependency.go | `DependencyGraph` with cycle detection and topological sort |
| tool_chain.go | `expandToolChain` for multi-tool build sequences |
| resolver_port.go | `ResolverPort` contract adapter and global singleton |

## Dependencies

- `contracts/core` -- `UnitResolverPort`, `ActionType`, `ModuleContractPort` interfaces
- `core/adapters` -- `AdaptUnitSpecs` for port conversion
- `core/config` -- `EACConfig`, `ComponentType` for tool and component configuration
- `core/domain/modules` -- `ModuleContract` concrete module type
- `core/resource` -- `PoolAllocation` for host vs container scheduling
- `core/tool` -- `BuildBridge`, `BuildHandler`, `ToolDefinition` for tool lookup
- `core/workunit` -- `UnitSpec`, `UnitID` for work unit creation

## Role in System

This package is the bridge between module definitions and the scheduling layer.
Commands call `ComponentResolver` methods to transform a module's component
list into weighted, dependency-ordered `UnitSpec` slices that the scheduler
can execute in parallel. The `ResolverPort` exposes this logic through the
contract interface for clean architecture boundaries.

## Code Health

### Tech Debt
- `component_resolver.go:41`: `ResolveForBuild` (~179 lines) handles dependency graph construction, tool-chain expansion, pool allocation, and weight calculation in a single method
- `component_resolver.go` (500 lines): high density of resolution concerns -- tool lookup, weight computation, scanner category mapping all in one file

### Pain Points
- Three resolve methods (`ResolveForBuild`, `ResolveForLint`, `ResolveForScan`) share overlapping logic for tool lookup and weight calculation but are not factored into a common helper
- Weight calculation is spread across three methods (`getToolWeight`, `getWeight`, `getScanWeight`) with subtly different heuristics that are hard to compare

### Optimization Opportunities
- Extract shared resolution logic (component iteration, tool lookup, weight assignment) into a template method to reduce duplication across build/lint/scan (medium effort, reduces ~100 lines)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
