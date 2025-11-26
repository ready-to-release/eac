// Package reset provides BDD step definitions for the commit reset subcommand.
//
// This file contains context variables for reset test state.
package reset

// Context holds state between steps for reset tests.
// These are set by the main test runner to synchronize with the shared context.
type Context struct {
	CommandOutput string
	ExitCode      int
	CommandError  error
	ResetTestFile string // File tracked for reset verification
	WasUsed       bool   // Track if reset context was actually used in this scenario
}

// Ctx is the reset test context - synchronized with main tests.Ctx
var Ctx *Context

// IsolatedTestProjectDir stores the temp directory for isolated test projects
var IsolatedTestProjectDir string

// OriginalRepoRoot stores the actual repository root
var OriginalRepoRoot string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error
