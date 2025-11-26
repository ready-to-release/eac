// Godog BDD step definitions for specs commands - Bridge to specs module tests
//
// This file bridges the specs step definitions from the specs module
// to the cross-module test runner in src/commands/tests.
//
// The actual step implementations live in:
//   src/commands/impl/specs/tests/
//
// State Synchronization Strategy:
// We sync context at the start of every step to ensure specs subpackages
// have access to the latest state, and sync back after every step to
// capture any changes they made.
package tests

import (
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
	specstests "github.com/ready-to-release/eac/src/commands/impl/specs/tests"
)

// syncToSpecsCtx copies state from main ctx to specstests.Ctx
func syncToSpecsCtx() {
	if specstests.Ctx == nil {
		specstests.Ctx = &specstests.TestContext{}
	}
	if ctx != nil {
		specstests.Ctx.CommandOutput = ctx.commandOutput
		specstests.Ctx.ExitCode = ctx.exitCode
		specstests.Ctx.CommandError = ctx.commandError
		specstests.Ctx.ValidationErrors = ctx.validationErrors
	}
}

// syncFromSpecsCtx copies state back from specstests.Ctx to main ctx
// Only syncs command results if specs tests actually ran a command
func syncFromSpecsCtx() {
	if specstests.Ctx != nil && ctx != nil {
		// Only sync command output if specs tests actually ran a command
		// (indicated by non-empty output or non-zero exit code)
		// Don't sync just because files were created - that doesn't mean
		// a command was run through specs context
		hasCommandOutput := specstests.Ctx.CommandOutput != "" ||
			specstests.Ctx.ExitCode != 0

		if hasCommandOutput {
			ctx.commandOutput = specstests.Ctx.CommandOutput
			ctx.exitCode = specstests.Ctx.ExitCode
			ctx.commandError = specstests.Ctx.CommandError
			ctx.validationErrors = specstests.Ctx.ValidationErrors
		}
	}
}

// setupSpecsMocks bridges to the specs module's mock setup
// Called automatically when @env:isolated-test-project tag is present
func setupSpecsMocks() error {
	// Initialize shared state
	syncToSpecsCtx()
	specstests.OriginalRepoRoot = originalRepoRoot
	specstests.IsolatedTestProjectDir = isolatedTestProjectDir
	specstests.RunCommand = iRunTheCommand

	// Set up test directory for file creation
	if specstests.Ctx != nil {
		specstests.Ctx.TestDir = isolatedTestProjectDir
		if specstests.Ctx.TestDir == "" {
			// Fallback to original repo root
			specstests.Ctx.TestDir = originalRepoRoot
		}
	}

	return specstests.SetupSpecsMocks()
}

// cleanupSpecsMocks bridges to the specs module's mock cleanup
func cleanupSpecsMocks() {
	syncFromSpecsCtx()
	specstests.CleanupSpecsMocks()

	// Clean up any test files created
	if specstests.Ctx != nil {
		for _, file := range specstests.Ctx.CreatedFiles {
			os.Remove(file)
			dir := filepath.Dir(file)
			os.Remove(dir) // Remove empty directories
		}
		specstests.Ctx.CreatedFiles = nil
	}
}

// InitializeSpecsScenario registers specs-specific step definitions
func InitializeSpecsScenario(sc *godog.ScenarioContext) {
	// Set up shared state
	specstests.OriginalRepoRoot = originalRepoRoot
	specstests.IsolatedTestProjectDir = isolatedTestProjectDir
	specstests.RunCommand = iRunTheCommand

	// Sync context before and after every step
	sc.BeforeStep(func(st *godog.Step) {
		syncToSpecsCtx()
	})
	sc.AfterStep(func(st *godog.Step, err error) {
		syncFromSpecsCtx()
	})

	// Register the specs module's steps
	specstests.InitializeSpecsScenario(sc)
}
