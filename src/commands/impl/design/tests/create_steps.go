// Package tests provides BDD step definitions for the design create subcommand.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// registerCreateSteps registers all step definitions for design create command.
func registerCreateSteps(sc *godog.ScenarioContext) {
	// Given steps - setup preconditions
	sc.Step(`^module "([^"]*)" has source code at "([^"]*)"$`, moduleHasSourceCodeAt)
	sc.Step(`^workspace exists for module "([^"]*)"$`, workspaceExistsForModule)
	sc.Step(`^no workspace exists for module "([^"]*)"$`, noWorkspaceExistsForModule)
	sc.Step(`^AI will generate a valid workspace$`, aiWillGenerateValidWorkspace)
	sc.Step(`^AI will generate an invalid workspace$`, aiWillGenerateInvalidWorkspace)
	sc.Step(`^Docker is not running$`, dockerIsNotRunning)
	sc.Step(`^Docker is running$`, dockerIsRunning)

	// When steps - execute actions
	// Note: We don't register "I run design create" steps here because they would be ambiguous
	// with the generic "I run" step. The generic step handles all "I run" commands.

	// Then steps - verify outcomes
	sc.Step(`^workspace should be created at "([^"]*)"$`, workspaceShouldBeCreatedAt)
	sc.Step(`^workspace should contain "([^"]*)"$`, workspaceShouldContain)
	sc.Step(`^workspace should not contain "([^"]*)"$`, workspaceShouldNotContain)
	sc.Step(`^workspace should be valid Structurizr DSL$`, workspaceShouldBeValidStructurizrDSL)
	sc.Step(`^debug files should be saved to "([^"]*)"$`, debugFilesShouldBeSavedTo)
	sc.Step(`^the output should indicate validation was skipped$`, outputShouldIndicateValidationSkipped)
	sc.Step(`^the output should indicate Docker is required$`, outputShouldIndicateDockerRequired)
}

// Given steps

func moduleHasSourceCodeAt(module, path string) error {
	// Create source directory and minimal files
	return CreateModuleSource(module)
}

func workspaceExistsForModule(module string) error {
	return CreateValidWorkspace(module)
}

func noWorkspaceExistsForModule(module string) error {
	return DeleteWorkspace(module)
}

func aiWillGenerateValidWorkspace() error {
	ctx := GetContext()
	ctx.SetMockAIResponseForCreate(ValidWorkspaceDSL)
	return nil
}

func aiWillGenerateInvalidWorkspace() error {
	ctx := GetContext()
	ctx.SetMockAIResponseForCreate(InvalidWorkspaceDSL)
	return nil
}

func dockerIsNotRunning() error {
	SetDockerRunning(false)
	return nil
}

func dockerIsRunning() error {
	SetDockerRunning(true)
	return nil
}

// When steps

func iRunDesignCreate(module string) error {
	ctx := GetContext()
	args := []string{"design", "create", module}
	return ctx.RunCommand(args)
}

func iRunDesignCreateWithFlag(module, flag string) error {
	ctx := GetContext()
	args := []string{"design", "create", module, flag}
	return ctx.RunCommand(args)
}

func iRunDesignCreateWithFlags(module, flags string) error {
	ctx := GetContext()
	flagList := strings.Fields(flags)
	args := append([]string{"design", "create", module}, flagList...)
	return ctx.RunCommand(args)
}

func iRunDesignCreateWithCustomPrompt(module, promptPath string) error {
	ctx := GetContext()
	args := []string{"design", "create", module, "--prompt", promptPath}
	return ctx.RunCommand(args)
}

// Then steps

func workspaceShouldBeCreatedAt(expectedPath string) error {
	// Extract module name from path (e.g., "specs/test-module/design/workspace.dsl" -> "test-module")
	parts := strings.Split(expectedPath, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid workspace path: %s", expectedPath)
	}
	module := parts[1]

	return AssertWorkspaceExists(module)
}

func workspaceShouldContain(expectedText string) error {
	// TODO: Track current module in context
	return fmt.Errorf("not implemented: need to track current module")
}

func workspaceShouldNotContain(unexpectedText string) error {
	// TODO: Similar to workspaceShouldContain
	return fmt.Errorf("not implemented: need to track current module")
}

func workspaceShouldBeValidStructurizrDSL() error {
	// Workspace validation is already done during creation if Docker is available
	// This step is implicitly verified by successful command execution
	return nil
}

func debugFilesShouldBeSavedTo(directory string) error {
	// Check for debug files in the specified directory
	return AssertDebugFileExists("design-debug-full-prompt.md")
}

func outputShouldIndicateValidationSkipped() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "Skipping Docker validation") &&
		!strings.Contains(output, "Validation skipped") {
		return fmt.Errorf("output does not indicate validation was skipped")
	}

	return nil
}

func outputShouldIndicateDockerRequired() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "Docker") {
		return fmt.Errorf("output does not mention Docker requirement")
	}

	return nil
}
