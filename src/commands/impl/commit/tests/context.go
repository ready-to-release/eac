// Package tests provides shared test context for commit subcommands.
//
// This file contains shared state and context types used across step definitions
// for all commit subcommands (message, reset, etc.).
package tests

import "github.com/ready-to-release/eac/src/core/git"

// TestContext holds shared state between steps across all commit subcommands.
type TestContext struct {
	// Shared fields
	CommandOutput string
	ExitCode      int
	CommandError  error

	// Message subcommand fields
	TestCommitMessage string   // For commit validation testing
	AffectedModules   []string // Modules affected by current changes
	ValidationErrors  []string // Validation error codes for assertions

	// Reset subcommand fields
	ResetTestFile string // File tracked for reset verification
}

// Ctx is the shared test context - set by the main test runner
var Ctx *TestContext

// OriginalRepoRoot stores the actual eac repository root - set by the main test runner
var OriginalRepoRoot string

// IsolatedTestProjectDir stores the temp directory for isolated test projects - set by the main test runner
var IsolatedTestProjectDir string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error

// TestMockRepo holds the mock git repository for isolated tests
var TestMockRepo *git.MockRepository

// TestAIOutput stores AI output for noise filtering tests
var TestAIOutput string
