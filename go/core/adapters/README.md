# adapters

Adapter layer that wraps concrete domain types to satisfy port interfaces defined
in the contracts package, enabling dependency inversion across module boundaries.

See doc.go for API details and usage examples.

## Key Types

- **`ModuleContractAdapter`** -- Wraps `*modules.ModuleContract` to `core.ModuleContractPort`
- **`ModuleRegistryAdapter`** -- Wraps `*modules.Registry` to `core.ModuleRegistryPort`
- **`UnitIDAdapter`** -- Wraps `workunit.UnitID` to `core.UnitIDPort`
- **`UnitSpecAdapter`** -- Wraps `workunit.UnitSpec` to `core.UnitSpecPort`
- **`UnitResultAdapter`** -- Wraps `workunit.UnitResult` to `core.UnitResultPort`
- **`PoolAllocationAdapter`** -- Wraps pool allocation to `core.PoolAllocationPort`

## Patterns

- Compile-time checks: `var _ Port = (*Adapter)(nil)` ensures interface compliance
- Unwrap escape hatch: Each adapter exposes `Unwrap()` for access to concrete types when needed
- Convenience functions: `AdaptModule`, `AdaptRegistry`, `AdaptUnitID` for single-call wrapping

## Internal Structure

| File | Responsibility |
| --- | --- |
| doc.go | Package documentation listing all adapters |
| module_adapter.go | `ModuleContractAdapter` with `AdaptModule`/`AdaptModules`/`UnwrapModule` |
| registry_adapter.go | `ModuleRegistryAdapter` with `AdaptRegistry` |
| unit_adapter.go | Unit ID, spec, result, and pool allocation adapters |

## Dependencies

- `core/domain/modules` -- concrete module and registry types
- `core/workunit` -- concrete UnitID, UnitSpec, UnitResult types
- `contracts/core` -- port interfaces (ModuleContractPort, UnitIDPort, etc.)

## Role in System

`adapters` is the dependency-inversion bridge in the `core` module. It allows
command handlers and orchestrators to depend on abstract port interfaces from the
contracts package rather than concrete domain types, enabling testability and
decoupling between the domain layer and consuming code.

## Code Health

### Tech Debt
- None identified

### Pain Points
- None identified

### Optimization Opportunities
- All adapter files are concise and focused (doc.go, module_adapter.go, registry_adapter.go, unit_adapter.go all under 150 lines)
- adapters_test.go present providing test coverage
