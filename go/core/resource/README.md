# resource

Domain types and port interfaces for resource pool management, including
capacity detection, calculation, and allocation across host and Docker pools.

## Key Types

- **`PoolType`** -- Identifies host vs Docker resource pool (alias)
- **`PoolAllocation`** -- Describes which pools a work unit needs (alias)
- **`PoolCapacity`** -- Current capacity state (total, used, waiting)
- **`SystemResources`** -- Detected system RAM, Docker memory, CPU count
- **`CapacityConfig`** -- Configuration for capacity calculation
- **`DefaultCalculator`** -- Standard capacity formulas from resources
- **`CapacityDetector`** -- Port for detecting system resources
- **`PoolAcquirer`** -- Port for semaphore-based capacity allocation

## Patterns

- Port interfaces: `CapacityDetector` and `PoolAcquirer` are adapter boundaries
- Type aliasing: `PoolType`, `PoolAllocation` alias from contracts
- Pure domain logic: `DefaultCalculator` has no external dependencies
- Dual-pool model: host and Docker pools allocated independently

## Internal Structure

| File | Responsibility |
| --- | --- |
| pool.go | `PoolType`, `PoolAllocation` aliases and factory functions |
| port.go | `CapacityDetector`, `CapacityCalculator`, `PoolAcquirer` interfaces |
| capacity.go | `PoolCapacity`, `SystemResources`, `CapacityConfig` types |
| calculator.go | `DefaultCalculator` with RAM/CPU-based capacity formulas |

## Dependencies

- `contracts/core` -- canonical `PoolType`, `PoolAllocation` definitions

## Role in System

This package defines the resource model that the orchestrator uses to
gate concurrent work unit execution. The `DefaultCalculator` computes
pool sizes from detected system resources, and the `PoolAcquirer` port
is implemented by the orchestrator layer to enforce capacity limits
with semaphores during parallel builds.

## Code Health

### Tech Debt
- None identified

### Pain Points
- `pool.go:24,28`: `var HostOnlyAllocation` and `var ContainerAllocation` are mutable package-level aliases from contracts; if overwritten they could silently break allocation logic

### Optimization Opportunities
- Consider making `HostOnlyAllocation` and `ContainerAllocation` constants or unexported to prevent accidental mutation (trivial effort)
- No TODO/FIXME markers found -- codebase is clean of deferred work items
