// Package update provides BDD step definitions for the design update subcommand.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// RegisterUpdateSteps registers all step definitions for design update command.
func registerUpdateSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^module "([^"]*)" has an existing workspace$`, moduleHasExistingWorkspace)
	sc.Step(`^workspace for "([^"]*)" contains "([^"]*)"$`, workspaceContains)
	sc.Step(`^source code for "([^"]*)" has changed$`, sourceCodeHasChanged)

	// When steps
	sc.Step(`^I run "design update ([^"]*)"$`, iRunDesignUpdate)
	sc.Step(`^I run "design update ([^"]*)" with flag "([^"]*)"$`, iRunDesignUpdateWithFlag)
	sc.Step(`^I run "design update ([^"]*)" with custom prompt$`, iRunDesignUpdateWithCustomPrompt)

	// Then steps
	sc.Step(`^workspace should be updated$`, workspaceShouldBeUpdated)
	sc.Step(`^workspace should still contain "([^"]*)"$`, workspaceShouldStillContain)
	sc.Step(`^workspace should now contain "([^"]*)"$`, workspaceShouldNowContain)
	sc.Step(`^the output should show "workspace updated"$`, outputShouldShowWorkspaceUpdated)
}

// Given steps

func moduleHasExistingWorkspace(module string) error {
	return CreateValidWorkspace(module)
}

func workspaceContains(module, content string) error {
	return AssertWorkspaceContains(module, content)
}

func sourceCodeHasChanged(module string) error {
	// Simulate source code changes by creating/updating source files
	return CreateModuleSource(module)
}

// When steps

func iRunDesignUpdate(module string) error {
	ctx := GetContext()
	args := []string{"design", "update", module}
	return ctx.RunCommand(args)
}

func iRunDesignUpdateWithFlag(module, flag string) error {
	ctx := GetContext()
	args := []string{"design", "update", module, flag}
	return ctx.RunCommand(args)
}

func iRunDesignUpdateWithCustomPrompt(module string) error {
	ctx := GetContext()
	// Assume custom prompt file exists in test fixtures
	args := []string{"design", "update", module, "--prompt", "fixtures/custom-prompt.md"}
	return ctx.RunCommand(args)
}

// Then steps

func workspaceShouldBeUpdated() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "updated") && !strings.Contains(output, "✅") {
		return fmt.Errorf("output does not indicate workspace was updated")
	}

	return nil
}

func workspaceShouldStillContain(expectedText string) error {
	// TODO: Track current module in context
	return fmt.Errorf("not implemented: need to track current module")
}

func workspaceShouldNowContain(expectedText string) error {
	// TODO: Track current module in context
	return fmt.Errorf("not implemented: need to track current module")
}

func outputShouldShowWorkspaceUpdated() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "updated") {
		return fmt.Errorf("output does not show 'workspace updated'")
	}

	return nil
}
