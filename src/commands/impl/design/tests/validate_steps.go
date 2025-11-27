// Package validate provides BDD step definitions for the design validate subcommand.
package tests

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// RegisterValidateSteps registers all step definitions for design validate command.
func registerValidateSteps(sc *godog.ScenarioContext) {
	// Given steps
	sc.Step(`^module "([^"]*)" has a valid workspace$`, moduleHasValidWorkspace)
	sc.Step(`^module "([^"]*)" has an invalid workspace$`, moduleHasInvalidWorkspace)
	sc.Step(`^module "([^"]*)" has workspace with syntax errors$`, moduleHasWorkspaceWithSyntaxErrors)
	sc.Step(`^multiple modules have workspaces$`, multipleModulesHaveWorkspaces)

	// When steps
	sc.Step(`^I run "design validate ([^"]*)"$`, iRunDesignValidate)
	sc.Step(`^I run "design validate ([^"]*)" with --verbose$`, iRunDesignValidateWithVerbose)
	sc.Step(`^I run "design validate ([^"]*)" with --debug$`, iRunDesignValidateWithDebug)
	sc.Step(`^I run "design validate --all"$`, iRunDesignValidateAll)

	// Then steps
	sc.Step(`^validation should pass$`, validationShouldPass)
	sc.Step(`^validation should fail$`, validationShouldFail)
	sc.Step(`^the output should show validation errors$`, outputShouldShowValidationErrors)
	sc.Step(`^the output should show syntax error details$`, outputShouldShowSyntaxErrorDetails)
	sc.Step(`^the output should show Docker command when verbose$`, outputShouldShowDockerCommand)
	sc.Step(`^validation results should be written to "([^"]*)"$`, validationResultsShouldBeWrittenTo)
}

// Given steps

func moduleHasValidWorkspace(module string) error {
	return CreateValidWorkspace(module)
}

func moduleHasInvalidWorkspace(module string) error {
	return CreateInvalidWorkspace(module)
}

func moduleHasWorkspaceWithSyntaxErrors(module string) error {
	return CreateInvalidWorkspace(module)
}

func multipleModulesHaveWorkspaces() error {
	// Create valid workspaces for multiple test modules
	if err := CreateValidWorkspace("module-a"); err != nil {
		return err
	}
	if err := CreateValidWorkspace("module-b"); err != nil {
		return err
	}
	return CreateValidWorkspace("module-c")
}

// When steps

func iRunDesignValidate(module string) error {
	ctx := GetContext()
	args := []string{"design", "validate", module}
	return ctx.RunCommand(args)
}

func iRunDesignValidateWithVerbose(module string) error {
	ctx := GetContext()
	args := []string{"design", "validate", module, "--verbose"}
	return ctx.RunCommand(args)
}

func iRunDesignValidateWithDebug(module string) error {
	ctx := GetContext()
	args := []string{"design", "validate", module, "--debug"}
	return ctx.RunCommand(args)
}

func iRunDesignValidateAll() error {
	ctx := GetContext()
	args := []string{"design", "validate", "--all"}
	return ctx.RunCommand(args)
}

// Then steps

func validationShouldPass() error {
	ctx := GetContext()

	if ctx.LastExitCode != 0 {
		return fmt.Errorf("validation failed with exit code %d", ctx.LastExitCode)
	}

	output := ctx.LastOutput
	if !strings.Contains(output, "✅") && !strings.Contains(output, "Valid") {
		return fmt.Errorf("output does not indicate validation passed")
	}

	return nil
}

func validationShouldFail() error {
	ctx := GetContext()

	if ctx.LastExitCode == 0 {
		return fmt.Errorf("validation passed when it should have failed")
	}

	return nil
}

func outputShouldShowValidationErrors() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "error") && !strings.Contains(output, "Error") &&
		!strings.Contains(output, "❌") {
		return fmt.Errorf("output does not show validation errors")
	}

	return nil
}

func outputShouldShowSyntaxErrorDetails() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "syntax") && !strings.Contains(output, "Syntax") {
		return fmt.Errorf("output does not show syntax error details")
	}

	return nil
}

func outputShouldShowDockerCommand() error {
	ctx := GetContext()
	output := ctx.LastOutput

	if !strings.Contains(output, "docker") {
		return fmt.Errorf("output does not show Docker command (verbose mode)")
	}

	return nil
}

func validationResultsShouldBeWrittenTo(expectedPath string) error {
	return AssertDebugFileExists("validation-results.json")
}
