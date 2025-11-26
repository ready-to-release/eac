// Package tests provides shared test context for specs subcommands (create and validate).
//
// This file contains shared state and context types used across step definitions
// for all specs subcommands.
package tests

import "github.com/ready-to-release/eac/src/core/git"

// TestContext holds shared state between steps across all specs subcommands.
type TestContext struct {
	// Shared fields
	CommandOutput string
	ExitCode      int
	CommandError  error

	// Create subcommand fields
	TestDescription  string   // For create command testing
	CreatedFiles     []string // Files created during test
	TestDir          string   // Temporary test directory

	// Validate subcommand fields
	ValidationErrors []string // Validation error codes for assertions
	RemovedContracts bool     // For contract missing tests
}

// Ctx is the shared test context - set by the main test runner
var Ctx *TestContext

// OriginalRepoRoot stores the actual repository root - set by the main test runner
var OriginalRepoRoot string

// IsolatedTestProjectDir stores the temp directory for isolated test projects - set by the main test runner
var IsolatedTestProjectDir string

// RunCommand is a function to run commands - set by the main test runner
var RunCommand func(cmdLine string) error

// TestMockRepo holds the mock git repository for isolated tests
var TestMockRepo *git.MockRepository

// TestAIOutput stores AI output for noise filtering tests
var TestAIOutput string
