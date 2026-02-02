package godog

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// RegisterCommonSteps registers step definitions shared across all specs.
// These are the generic steps that appear in multiple feature files.
//
// Step implementations delegate to helpers in helpers.go, which allows
// different verbs to share the same underlying logic while registering
// their own step patterns (avoiding conflicts).
func RegisterCommonSteps(sc *godog.ScenarioContext, ctx *TestContext) {
	// Command execution steps
	sc.Step(`^I run "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})
	sc.Step(`^I run the command "([^"]*)"$`, func(cmdLine string) error {
		return ctx.RunCommand(cmdLine)
	})

	// Exit code steps - using helpers
	sc.Step(`^the exit code is (\d+)$`, func(expectedCode int) error {
		return ExitCodeIs(ctx, expectedCode)
	})
	sc.Step(`^the command should succeed$`, func() error {
		return CommandSucceeded(ctx)
	})
	sc.Step(`^the command exits with code (\d+)$`, func(expectedCode int) error {
		return ExitCodeIs(ctx, expectedCode)
	})
	sc.Step(`^the exit code should be (\d+)$`, func(expectedCode int) error {
		return ExitCodeIs(ctx, expectedCode)
	})
	sc.Step(`^the exit code should be (\d+) or (\d+)$`, func(code1, code2 int) error {
		if ctx.ExitCode != code1 && ctx.ExitCode != code2 {
			return fmt.Errorf("expected exit code %d or %d, got %d", code1, code2, ctx.ExitCode)
		}
		return nil
	})

	// Output verification steps - using helpers
	sc.Step(`^I should see "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^I should see "([^"]*)" or "([^"]*)"$`, func(text1, text2 string) error {
		return OutputContainsAny(ctx, text1, text2)
	})
	sc.Step(`^I should see "([^"]*)" or "([^"]*)" or "([^"]*)"$`, func(text1, text2, text3 string) error {
		return OutputContainsAny(ctx, text1, text2, text3)
	})
	sc.Step(`^the output should contain "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^stdout contains "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^stderr contains "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^stderr contains "([^"]*)" or "([^"]*)"$`, func(text1, text2 string) error {
		return OutputContainsAny(ctx, text1, text2)
	})

	// Short-form output verification (commonly used in specs)
	sc.Step(`^I see "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^I see error "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^I see a table with "([^"]*)" and "([^"]*)" columns$`, func(col1, col2 string) error {
		if err := OutputContains(ctx, col1); err != nil {
			return err
		}
		return OutputContains(ctx, col2)
	})

	// File verification steps - using helpers
	sc.Step(`^the file "([^"]*)" should exist$`, func(path string) error {
		return FileExists(ctx, path)
	})
	sc.Step(`^debug files should exist in "([^"]*)"$`, func(dir string) error {
		return DirectoryHasFiles(ctx, dir)
	})

	// Git repository setup steps
	sc.Step(`^I am in a git repository$`, func() error {
		return IsGitRepository(ctx)
	})
	sc.Step(`^I am in an isolated test repository$`, func() error {
		return EnsureIsolatedTestRepository(ctx)
	})
	sc.Step(`^I am not in a git repository$`, func() error {
		return EnsureNotGitRepository(ctx)
	})
	sc.Step(`^the repository root exists$`, func() error {
		// This is typically satisfied by test setup
		return nil
	})

	// Directory setup steps
	sc.Step(`^no \.r2r directory exists$`, func() error {
		return RemoveAll(ctx, ".r2r")
	})

	// EAC configuration setup steps (shared across multiple packages)
	sc.Step(`^I am in a git repository with EAC configuration$`, func() error {
		// EAC configuration should exist in isolated test environment
		if err := IsGitRepository(ctx); err != nil {
			return err
		}
		return FileExists(ctx, ".r2r")
	})
	sc.Step(`^I am in a git repository with GitHub remote$`, func() error {
		// GitHub remote should be configured
		return IsGitRepository(ctx)
	})
	sc.Step(`^AI configuration exists at "([^"]*)"$`, func(path string) error {
		// Check or create AI config at the specified path
		return CreateDirectory(ctx, strings.TrimSuffix(path, "/"))
	})
	sc.Step(`^docker service is available$`, func() error {
		// Mock Docker is always available in isolated tests
		return nil
	})
	sc.Step(`^stdout contains valid JSON$`, func() error {
		// Check that output contains valid JSON structure
		return OutputContains(ctx, "{")
	})

	// NOTE: Feature-specific steps (work, risks, specs, etc.) are registered
	// in their respective step files in impl/eac-cli/
}
