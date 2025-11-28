package common

import (
	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
	"github.com/ready-to-release/eac/src/core/logging"
)

// SharedContext contains state that needs to be shared across test modules
type SharedContext struct {
	RepoRoot      string
	CurrentBranch string
	ExitCode      int
	Output        string
}

// TestContext holds state for work command tests
type TestContext struct {
	Args          []string
	ExitCode      int
	Output        string
	Error         error
	MockGitOps    *MockGitOps
	CurrentBranch string
	Worktrees     []internal.Worktree
	DebugFiles    []string

	// Shared context pointer for cross-module communication
	SharedCtx *SharedContext
}

// NewTestContext creates a new test context
func NewTestContext() *TestContext {
	return &TestContext{
		Args:       []string{},
		MockGitOps: NewMockGitOps(),
		SharedCtx:  &SharedContext{},
	}
}

// Reset clears the test context for a new scenario
func (ctx *TestContext) Reset() {
	ctx.Args = []string{}
	ctx.ExitCode = 0
	ctx.Output = ""
	ctx.Error = nil
	ctx.MockGitOps = NewMockGitOps()
	ctx.CurrentBranch = ""
	ctx.Worktrees = []internal.Worktree{}
	ctx.DebugFiles = []string{}

	// Keep shared context pointer but reset values
	if ctx.SharedCtx != nil {
		ctx.SharedCtx.RepoRoot = ""
		ctx.SharedCtx.CurrentBranch = ""
		ctx.SharedCtx.ExitCode = 0
		ctx.SharedCtx.Output = ""
	}
}

// SetupLogger initializes a test logger
func (ctx *TestContext) SetupLogger() *logging.Logger {
	logger, _ := logging.New(logging.Config{
		Module:        "work-test",
		WorkspaceRoot: ".",
		DebugMode:     true,
		Development:   true,
	})
	return logger
}
