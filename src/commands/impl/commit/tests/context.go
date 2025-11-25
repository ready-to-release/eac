// Package tests provides BDD step definitions for the commit command.
//
// This file contains shared state and context types used across step definitions.
package tests

import "github.com/ready-to-release/eac/src/core/git"

// TestContext holds state between steps - shared with main tests package
type TestContext struct {
	CommandOutput     string
	ExitCode          int
	CommandError      error
	TestCommitMessage string   // For commit validation testing
	AffectedModules   []string // Modules affected by current changes
	ValidationErrors  []string // Validation error codes for assertions
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
