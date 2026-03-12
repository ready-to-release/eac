# adapters

Dependency-inversion bridge between concrete domain types and the abstract port
interfaces defined in `contracts/core/0.1.0`. Consuming code (commands,
orchestrators, MCP handlers) depends on the port interfaces; this package
provides the adapters that satisfy them.

## Why

The `contracts/core` module defines lightweight port interfaces
(`ModuleContractPort`, `UnitIDPort`, etc.) so that command implementations never
import heavy domain packages directly. This package closes the loop: it wraps
the real domain structs from `core/domain/modules` and `core/workunit` behind
those ports, keeping the dependency arrow pointing inward.

## Adapters

| Adapter | Wraps | Satisfies |
|---|---|---|
| `ModuleContractAdapter` | `*modules.ModuleContract` | `core.ModuleContractPort` |
| `ModuleRegistryAdapter` | `*modules.Registry` | `core.ModuleRegistryPort` |
| `UnitIDAdapter` | `workunit.UnitID` | `core.UnitIDPort` |
| `UnitSpecAdapter` | `workunit.UnitSpec` | `core.UnitSpecPort` |
| `UnitResultAdapter` | `workunit.UnitResult` | `core.UnitResultPort` |
| `PoolAllocationAdapter` | `core.PoolAllocationPort` | `core.PoolAllocationPort` |

## Convenience Functions

Single-call wrapping and unwrapping:

- `AdaptModule(m)` / `AdaptModules(ms)` / `UnwrapModule(port)` -- module contracts
- `AdaptRegistry(r)` -- module registry
- `AdaptUnitID(id)` -- unit IDs
- `AdaptUnitSpec(s)` / `AdaptUnitSpecs(ss)` -- unit specs
- `AdaptUnitResult(r)` / `AdaptUnitResults(rs)` -- unit results

## Generic Helper

`AdaptSlice[T, P](items, adaptFn)` converts `[]T` to `[]P` using a supplied
function. Returns nil for nil input. Used internally by `AdaptModules`,
`AdaptUnitSpecs`, and `AdaptUnitResults`.

## Patterns

- **Compile-time checks** -- every adapter file contains
  `var _ Port = (*Adapter)(nil)` to guarantee interface compliance at build time.
- **Unwrap escape hatch** -- each adapter exposes `Unwrap()` to recover the
  concrete type when native handler code needs fields or methods not on the port.
  `UnwrapModule(port)` is a free function that does a safe type assertion.
- **Nil passthrough** -- all `Adapt*` convenience functions return nil for nil
  input, avoiding nil-pointer panics in optional chains.

## File Layout

| File | Contents |
|---|---|
| `doc.go` | Package doc + `AdaptSlice` generic helper |
| `module_adapter.go` | `ModuleContractAdapter`, `AdaptModule`, `AdaptModules`, `UnwrapModule` |
| `registry_adapter.go` | `ModuleRegistryAdapter`, `AdaptRegistry` |
| `unit_adapter.go` | `UnitIDAdapter`, `UnitSpecAdapter`, `UnitResultAdapter`, `PoolAllocationAdapter` and their convenience functions |
| `adapters_test.go` | Tests for all adapters (build tags: `L0,ov`) |

## Dependencies

- `contracts/core/0.1.0` -- port interfaces
- `core/domain/modules` -- `ModuleContract`, `Registry`
- `core/workunit` -- `UnitID`, `UnitSpec`, `UnitResult`
