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
//
// Generic helper:
//   - AdaptSlice: converts []T to []P via a supplied adapt function
package adapters

// AdaptSlice converts a slice of concrete values to a slice of port values
// using the provided adapt function. Returns nil for nil input.
func AdaptSlice[T any, P any](items []T, adapt func(T) P) []P {
	if items == nil {
		return nil
	}
	result := make([]P, len(items))
	for i, item := range items {
		result[i] = adapt(item)
	}
	return result
}
