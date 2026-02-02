// Package adapters provides adapters that wrap concrete implementations
// to satisfy port interfaces defined in github.com/ready-to-release/eac/contracts/core/0.1.0/interfaces.
//
// This package enables dependency inversion by allowing consuming code to depend
// on abstract port interfaces rather than concrete implementations.
//
// Key adapters:
//   - ModuleContractAdapter: wraps *modules.ModuleContract → interfaces.ModuleContractPort
//   - ModuleRegistryAdapter: wraps *modules.Registry → interfaces.ModuleRegistryPort
//   - UnitIDAdapter: wraps workunit.UnitID → interfaces.UnitIDPort
//   - UnitSpecAdapter: wraps workunit.UnitSpec → interfaces.UnitSpecPort
//   - UnitResultAdapter: wraps workunit.UnitResult → interfaces.UnitResultPort
package adapters
