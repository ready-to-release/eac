package commitmessage

import (
	"os/exec"

	"github.com/ready-to-release/eac/go/core/git"
)

// Deps holds injectable dependencies for testing.
// In production, use defaultDeps() which wires real implementations.
// Tests can construct a Deps with overrides (e.g. mock AI response, fake git repo).
type Deps struct {
	// AIResponse when non-empty bypasses AI and returns this response directly.
	AIResponse string
	// GetGitRepo returns a git repository for the given workspace root.
	GetGitRepo func(workspaceRoot string) (git.GitRepository, error)
	// ExecCmd creates an *exec.Cmd (mirrors exec.Command signature).
	ExecCmd func(name string, arg ...string) *exec.Cmd
}

// defaultDeps returns production defaults: real git, real exec, empty AI response.
func defaultDeps() *Deps {
	lazyRepo := &git.LazyRepo{}
	return &Deps{
		GetGitRepo: lazyRepo.Get,
		ExecCmd:    exec.Command,
	}
}
