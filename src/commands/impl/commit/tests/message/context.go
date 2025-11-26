// Package message provides BDD step definitions for the commit message subcommand.
//
// This file contains context variables for message test state.
package message

// Context holds state between steps for message tests.
// These are set by the main test runner to synchronize with the shared context.
type Context struct {
	CommandOutput     string
	ExitCode          int
	CommandError      error
	TestCommitMessage string
	AffectedModules   []string
	ValidationErrors  []string
}

// Ctx is the message test context - synchronized with main tests.Ctx
var Ctx *Context

// TestAIOutput stores AI output for noise filtering tests
var TestAIOutput string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error
