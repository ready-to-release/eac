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
- No test files exist; adapter correctness relies solely on compile-time interface checks (`var _ Port = (*Adapter)(nil)`) with no behavioral tests
- unit_adapter.go:90-96 -- `PoolAllocationAdapter` declares an anonymous interface instead of importing the concrete type; this duplicates the method set and may drift

### Pain Points
- `ModuleContractAdapter` delegates 16 methods, making the underlying `ModuleContractPort` interface unusually wide; consider splitting into smaller focused interfaces (e.g., identity, components, versioning)
- `AdaptModules`, `AdaptUnitSpecs`, and `AdaptUnitResults` repeat the same slice-map-adapt loop; a generic `adaptSlice[T, P]` helper (Go 1.18+) would eliminate the boilerplate

### Optimization Opportunities
- Adapter allocation could be avoided for read-only callers by embedding the port interface directly in the domain type (requires contract package change, larger effort)
- Adding a table-driven adapter test for each type would be low-cost and catch method signature drift early (straightforward, high value)
