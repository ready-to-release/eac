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
	sc.Step(`^I run the command "([^"]*)" without arguments$`, func(cmdLine string) error {
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
	sc.Step(`^the command should fail$`, func() error {
		return CommandFailed(ctx)
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
	sc.Step(`^the standard output contains "([^"]*)"$`, func(text string) error {
		return OutputContains(ctx, text)
	})
	sc.Step(`^the standard output does not contain "([^"]*)"$`, func(text string) error {
		return OutputDoesNotContain(ctx, text)
	})
	sc.Step(`^the standard output matches the pattern "([^"]*)"$`, func(pattern string) error {
		return OutputMatches(ctx, pattern)
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

	// NOTE: "the custom prompt is used" is NOT registered here.
	// Each verb that uses custom prompts should register its own step
	// with a verb-specific pattern, calling the CustomPromptWasUsed helper.
	// This avoids step conflicts between specs and risks.
}
