// Package tests provides BDD step definitions for the templates commands.
package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	applytests "github.com/ready-to-release/eac/src/commands/impl/templates/apply/tests"
	installtests "github.com/ready-to-release/eac/src/commands/impl/templates/install/tests"
)

// InitializeListScenario registers all step definitions for all templates scenarios
func InitializeListScenario(sc *godog.ScenarioContext) {
	// Lifecycle hooks - initialize appropriate context based on scenario name
	sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		if !strings.Contains(scenario.Uri, "templates") {
			return ctx, nil
		}

		// Initialize based on scenario name
		if strings.Contains(scenario.Name, "Apply") {
			if err := applytests.InitializeContext(); err != nil {
				return ctx, fmt.Errorf("failed to initialize apply context: %w", err)
			}
		} else if strings.Contains(scenario.Name, "Install") {
			if err := installtests.InitializeContext(); err != nil {
				return ctx, fmt.Errorf("failed to initialize install context: %w", err)
			}
		} else if strings.Contains(scenario.Name, "List") {
			if err := initializeListContext(); err != nil {
				return ctx, fmt.Errorf("failed to initialize list context: %w", err)
			}
		} else {
			return ctx, nil
		}

		if err := SetupMocks(); err != nil {
			return ctx, fmt.Errorf("failed to setup mocks: %w", err)
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		if !strings.Contains(scenario.Uri, "templates") {
			return ctx, nil
		}

		CleanupMocks()

		// Cleanup based on scenario name
		if strings.Contains(scenario.Name, "Apply") {
			if cleanupErr := applytests.CleanupContext(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup apply test context: %v\n", cleanupErr)
			}
		} else if strings.Contains(scenario.Name, "Install") {
			if cleanupErr := installtests.CleanupContext(); cleanupErr != nil {
				fmt.Printf("Warning: failed to cleanup install test context: %v\n", cleanupErr)
			}
		} else if strings.Contains(scenario.Name, "List") {
			if cleanupErr := cleanupListContext(); cleanupErr != nil {
				// Log but don't fail the test on cleanup errors
			}
		}

		return ctx, nil
	})

	// Setup steps
	sc.Step(`^I have a template directory "([^"]*)"$`, iHaveATemplateDirectory)
	sc.Step(`^I have a template file "([^"]*)" with content:$`, iHaveATemplateFileWithContent)
	sc.Step(`^I have a values file "([^"]*)" with:$`, applytests.IHaveAValuesFileWith)

	// Execution steps
	sc.Step(`^I run the command "([^"]*)"$`, iRunCommand)
	sc.Step(`^I run the templates command "([^"]*)"$`, iRunCommand)

	// Verification steps
	sc.Step(`^the command should succeed$`, theCommandShouldSucceed)
	sc.Step(`^the command should fail$`, theCommandShouldFail)
	sc.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	sc.Step(`^the command should attempt to clone from "([^"]*)"$`, theCommandShouldAttemptToCloneFrom)
	sc.Step(`^debug files should exist in "([^"]*)"$`, debugFilesShouldExistIn)
	sc.Step(`^the file "([^"]*)" should exist$`, theFileShouldExist)
	sc.Step(`^the file "([^"]*)" should contain "([^"]*)"$`, theFileShouldContain)
}
