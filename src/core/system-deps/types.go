// Package systemdeps provides system dependency verification
package systemdeps

// Result contains the result of a dependency check
type Result struct {
	Dependency      string // Original input (e.g., "@deps:docker" or "docker")
	Moniker         string // Normalized moniker (e.g., "docker")
	Name            string // Human-readable name (e.g., "Docker")
	Available       bool
	Version         string // Detected version
	RequiredVersion string // Required version from config
	Error           error
}

// IsSuccess returns true if the dependency is available without errors
func (r Result) IsSuccess() bool {
	return r.Available && r.Error == nil
}
