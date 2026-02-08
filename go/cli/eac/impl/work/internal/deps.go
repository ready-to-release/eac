// Package internal provides shared utilities for work commands
package internal

// Deps holds injectable dependencies for work commands.
type Deps struct {
	GitOps WorkGitOperations
}

// DefaultDeps returns a Deps populated with production defaults.
func DefaultDeps(repoRoot string) *Deps {
	return &Deps{
		GitOps: NewDefaultGitOps(repoRoot),
	}
}
