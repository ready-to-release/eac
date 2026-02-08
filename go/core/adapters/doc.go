// Package adapters provides adapters that wrap concrete implementations
// to satisfy port interfaces defined in github.com/ready-to-release/eac/contracts/core/0.1.0.
//
// This package enables dependency inversion by allowing consuming code to depend
// on abstract port interfaces rather than concrete implementations.
//
// Key adapters:
//   - ModuleContractAdapter: wraps *modules.ModuleContract → core.ModuleContractPort
//   - ModuleRegistryAdapter: wraps *modules.Registry → core.ModuleRegistryPort
//   - UnitIDAdapter: wraps workunit.UnitID → core.UnitIDPort
//   - UnitSpecAdapter: wraps workunit.UnitSpec → core.UnitSpecPort
//   - UnitResultAdapter: wraps workunit.UnitResult → core.UnitResultPort
package adapters
