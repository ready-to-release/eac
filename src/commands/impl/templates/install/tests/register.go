package tests

import (
	"context"
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// InitializeInstallScenario registers all step definitions for install command tests
func InitializeInstallScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		// Only initialize for templates install scenarios
		if !strings.Contains(scenario.Uri, "templates") || !strings.Contains(scenario.Name, "Install") {
			return ctx, nil
		}
		if err := InitializeContext(); err != nil {
			return ctx, fmt.Errorf("failed to initialize install context: %w", err)
		}
		if err := SetupMocks(); err != nil {
			return ctx, fmt.Errorf("failed to setup mocks: %w", err)
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		// Only cleanup for templates install scenarios
		if !strings.Contains(scenario.Uri, "templates") || !strings.Contains(scenario.Name, "Install") {
			return ctx, nil
		}
		CleanupMocks()
		if cleanupErr := CleanupContext(); cleanupErr != nil {
			fmt.Printf("Warning: failed to cleanup install test context: %v\n", cleanupErr)
		}
		return ctx, nil
	})

	// Setup steps
	sc.Step(`^I have a template directory "([^"]*)"$`, iHaveATemplateDirectory)
	sc.Step(`^I have a template file "([^"]*)" with content:$`, iHaveATemplateFileWithContent)

	// Execution steps
	sc.Step(`^I run the templates command "([^"]*)"$`, iRunCommand)

	// Verification steps
	sc.Step(`^the command should succeed$`, theCommandShouldSucceed)
	sc.Step(`^the command should fail$`, theCommandShouldFail)
	sc.Step(`^the file "([^"]*)" should exist$`, theFileShouldExist)
	sc.Step(`^the file "([^"]*)" should contain "([^"]*)"$`, theFileShouldContain)
	sc.Step(`^the output should contain "([^"]*)"$`, theOutputShouldContain)
	sc.Step(`^debug files should exist in "([^"]*)"$`, debugFilesShouldExistIn)
}
