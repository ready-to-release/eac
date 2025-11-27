// Package message provides BDD step definitions for the commit message subcommand.
//
// This file contains context variables for message test state.
// Context is shared directly with the parent tests package to avoid complex synchronization.
package message

// Context holds state between steps for message tests.
// This is a pointer to the shared context in the parent tests package.
type Context struct {
	CommandOutput     string
	ExitCode          int
	CommandError      error
	TestCommitMessage string
	AffectedModules   []string
	ValidationErrors  []string
}

// Ctx is the message test context - this is set by the parent tests package
// and should be used as a direct reference (not copied).
var Ctx *Context

// TestAIOutput stores AI output for noise filtering tests
var TestAIOutput string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error
