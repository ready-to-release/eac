package internal

import (
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
	sc.Step(`^the exit code is (\d+) or (\d+)$`, func(code1, code2 int) error {
		return ExitCodeIsOneOf(ctx, code1, code2)
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

	// File verification steps - using helpers
	sc.Step(`^the file "([^"]*)" should exist$`, func(path string) error {
		return FileExists(ctx, path)
	})
	sc.Step(`^the file "([^"]*)" should contain "([^"]*)"$`, func(path, content string) error {
		return FileContains(ctx, path, content)
	})
	sc.Step(`^a file exists at "([^"]*)"$`, func(path string) error {
		return FileExists(ctx, path)
	})
	sc.Step(`^debug files should exist in "([^"]*)"$`, func(dir string) error {
		return DirectoryHasFiles(ctx, dir)
	})

	// Git repository setup steps
	sc.Step(`^I am in a git repository$`, func() error {
		return IsGitRepository(ctx)
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
	sc.Step(`^a \.r2r directory already exists$`, func() error {
		return CreateDirectory(ctx, ".r2r/eac")
	})

	// NOTE: Feature-specific steps (work, risks, specs, etc.) are registered
	// in their respective step files in impl/src-commands/
}
