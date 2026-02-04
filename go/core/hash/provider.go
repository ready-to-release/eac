package hash

import "github.com/ready-to-release/eac/go/core/workunit"

// ModuleInputHashProvider computes input hash for a UnitID based on module source files.
// This is a function type that matches output.InputHashProvider for use with DetectUoWChanges.
type ModuleInputHashProvider func(id workunit.UnitID) (string, error)

// NewModuleInputHashProvider creates a provider that computes input hashes from pre-collected module files.
// The moduleFiles map should contain module name -> list of source files (relative paths).
// All components within a module share the same input hash based on module source files.
//
// Returns empty string (not error) for unknown modules to treat them as changed.
// Returns error only for actual hash computation failures (e.g., missing files).
func NewModuleInputHashProvider(workspaceRoot string, moduleFiles map[string][]string) ModuleInputHashProvider {
	return func(id workunit.UnitID) (string, error) {
		files, ok := moduleFiles[id.Module]
		if !ok {
			return "", nil // Unknown module = treat as changed
		}
		return Files(workspaceRoot, files)
	}
}
