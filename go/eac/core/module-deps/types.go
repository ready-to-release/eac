// Package moduledeps provides module dependency verification
package moduledeps

// Checker interface for verifying module availability
type Checker interface {
	GetName() string
	IsAvailable() bool
	GetVersion() (string, error)
}

// Result represents the verification result for a module dependency
type Result struct {
	Dependency string
	Available  bool
	Version    string
	Error      error
}
