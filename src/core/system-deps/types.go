// Package systemdeps provides system dependency verification
package systemdeps

// Checker verifies if a system dependency is available
type Checker interface {
	IsAvailable() bool
	GetVersion() (string, error)
	GetName() string
}

// Result contains the result of a dependency check
type Result struct {
	Dependency string // e.g., "@dep:docker"
	Available  bool
	Version    string
	Error      error
}
