// Package internal provides shared utilities for work commands
package internal

import (
	"github.com/ready-to-release/eac/go/core/tool"
)

// Deps holds injectable dependencies for work commands.
type Deps struct {
	GitOps WorkGitOperations
}

// DefaultDeps returns a Deps populated with production defaults.
func DefaultDeps(repoRoot string) *Deps {
	ts := tool.GlobalToolSystem()
	return &Deps{
		GitOps: NewDefaultGitOps(repoRoot, ts),
	}
}
