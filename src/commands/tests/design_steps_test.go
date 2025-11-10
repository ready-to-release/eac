// Feature: src-commands_design-command
// Godog step implementations for design command BDD scenarios
//
// This file implements steps for the specification at:
// specs/src-commands/design-command/specification.feature
package tests

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// designTestContext holds state specific to design command tests
type designTestContext struct {
	dockerClient     *client.Client
	dockerAvailable  bool
	containerStarted bool
	containerID      string
	containerURL     string
}

var designCtx *designTestContext

// ============================================================================
// Given Steps
// ============================================================================

func dockerIsRunning() error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		designCtx.dockerAvailable = false
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	_, err = cli.Ping(context.Background())
	if err != nil {
		designCtx.dockerAvailable = false
		cli.Close()
		return fmt.Errorf("docker is not running: %w", err)
	}

	designCtx.dockerClient = cli
	designCtx.dockerAvailable = true
	return nil
}

func moduleHasWorkspaceDslFile(module string) error {
	// Tests run from src/commands/tests, so we need to go up three directories
	// Validation uses specs/<module>/design/workspace.dsl
	workspacePath := filepath.Join("..", "..", "..", "specs", module, "design", "workspace.dsl")
	if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
		return fmt.Errorf("workspace.dsl not found at %s", workspacePath)
	}
	return nil
}

func multipleModulesHaveWorkspaceDslFiles() error {
	// Check for at least 2 modules with workspace.dsl files
	// Tests run from src/commands/tests
	specsDir := filepath.Join("..", "..", "..", "specs")

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return fmt.Errorf("failed to read specs directory: %w", err)
	}

	moduleCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			workspacePath := filepath.Join(specsDir, entry.Name(), "design", "workspace.dsl")
			if _, err := os.Stat(workspacePath); err == nil {
				moduleCount++
			}
		}
	}

	if moduleCount < 2 {
		return fmt.Errorf("expected at least 2 modules with workspace.dsl files, found %d", moduleCount)
	}

	return nil
}

// ============================================================================
// Then Steps
// ============================================================================

func structurizrContainerShouldStartSuccessfully() error {
	if !designCtx.dockerAvailable {
		return fmt.Errorf("docker is not available")
	}

	// Check if container was created
	containerName := "structurizr-cli"
	containers, err := designCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	found := false
	for _, c := range containers {
		for _, name := range c.Names {
			if strings.TrimPrefix(name, "/") == containerName {
				designCtx.containerID = c.ID
				designCtx.containerStarted = true
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("structurizr container was not created")
	}

	return nil
}

func iShouldSeeSuccessMessageWithURL() error {
	// Check if output contains a URL
	if !strings.Contains(ctx.commandOutput, "http://localhost:") {
		return fmt.Errorf("output does not contain URL, got:\n%s", ctx.commandOutput)
	}

	// Extract URL for later verification
	lines := strings.Split(ctx.commandOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "http://localhost:") {
			start := strings.Index(line, "http://")
			if start >= 0 {
				url := strings.TrimSpace(line[start:])
				// Remove any trailing characters
				if idx := strings.IndexAny(url, " \t\n"); idx > 0 {
					url = url[:idx]
				}
				designCtx.containerURL = url
				break
			}
		}
	}

	if designCtx.containerURL == "" {
		return fmt.Errorf("could not extract URL from output")
	}

	return nil
}

func documentationShouldBeAccessibleAtTheURL() error {
	if designCtx.containerURL == "" {
		return fmt.Errorf("no URL found to check")
	}

	// Wait a bit for Structurizr to fully start
	time.Sleep(3 * time.Second)

	// Try to access the URL
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(designCtx.containerURL)
	if err != nil {
		return fmt.Errorf("failed to access %s: %w", designCtx.containerURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("expected status 200 or 302, got %d", resp.StatusCode)
	}

	return nil
}

func iShouldSeeAListOfAvailableModules() error {
	// Check if output contains table headers
	if !strings.Contains(ctx.commandOutput, "MODULE") {
		return fmt.Errorf("output does not contain module table, got:\n%s", ctx.commandOutput)
	}
	return nil
}

func moduleShouldBeInTheList(module string) error {
	if !strings.Contains(ctx.commandOutput, module) {
		return fmt.Errorf("module '%s' not found in output:\n%s", module, ctx.commandOutput)
	}
	return nil
}

// Validation-specific Then steps

func theWorkspaceShouldBeValidatedUsingStructurizrCLI() error {
	// Check if command executed successfully (exit code 0 or 1)
	// Exit code 1 means validation found errors, but still ran successfully
	if ctx.exitCode != 0 && ctx.exitCode != 1 {
		return fmt.Errorf("expected exit code 0 or 1, got %d", ctx.exitCode)
	}

	// Check for validation indicators in output
	if !strings.Contains(ctx.commandOutput, "Validating module:") {
		return fmt.Errorf("output does not indicate validation occurred:\n%s", ctx.commandOutput)
	}

	return nil
}

func allWorkspacesShouldBeValidatedUsingStructurizrCLI() error {
	// Similar to single validation but checks for multiple modules
	if ctx.exitCode != 0 && ctx.exitCode != 1 {
		return fmt.Errorf("expected exit code 0 or 1, got %d", ctx.exitCode)
	}

	// Check for validation indicators
	if !strings.Contains(ctx.commandOutput, "Validating") {
		return fmt.Errorf("output does not indicate validation occurred:\n%s", ctx.commandOutput)
	}

	return nil
}

func validationResultsShouldBeDisplayedInConsole() error {
	// Check for key validation output elements
	expectedElements := []string{
		"Validating module:",
		"Workspace:",
		"Summary:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(ctx.commandOutput, element) {
			return fmt.Errorf("expected output to contain '%s', got:\n%s", element, ctx.commandOutput)
		}
	}

	return nil
}

func validationResultsForEachModuleShouldBeDisplayedInConsole() error {
	// Check for summary section that aggregates results
	if !strings.Contains(ctx.commandOutput, "Summary:") {
		return fmt.Errorf("output does not contain validation summary:\n%s", ctx.commandOutput)
	}

	// Should show module-level results
	if !strings.Contains(ctx.commandOutput, "Module:") {
		return fmt.Errorf("output does not show per-module results:\n%s", ctx.commandOutput)
	}

	return nil
}

func validationResultsShouldBeWrittenToJSONFile() error {
	// Tests run from src/commands/tests, need to go up to project root
	jsonPath := filepath.Join("..", "..", "..", "out", "design-validation-results.json")

	// Check if JSON file exists
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		return fmt.Errorf("validation JSON file not found at %s", jsonPath)
	}

	// Read and verify it's valid JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}

	if len(data) == 0 {
		return fmt.Errorf("JSON file is empty")
	}

	// Basic JSON validation - should start with { or [
	trimmed := strings.TrimSpace(string(data))
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return fmt.Errorf("JSON file does not contain valid JSON")
	}

	return nil
}

func aggregatedValidationResultsShouldBeWrittenToJSONFile() error {
	// Same as single validation - both write to same file
	return validationResultsShouldBeWrittenToJSONFile()
}

func iShouldSeeValidationSummaryWithErrorsAndWarnings() error {
	// Check for summary section with counts
	if !strings.Contains(ctx.commandOutput, "Summary:") {
		return fmt.Errorf("output does not contain summary section:\n%s", ctx.commandOutput)
	}

	// Should show errors and warnings counts
	hasErrors := strings.Contains(ctx.commandOutput, "Errors:")
	hasWarnings := strings.Contains(ctx.commandOutput, "Warnings:")

	if !hasErrors || !hasWarnings {
		return fmt.Errorf("summary does not show errors and warnings counts:\n%s", ctx.commandOutput)
	}

	return nil
}

func iShouldSeeOverallSummaryWithTotalErrorsAndWarnings() error {
	// Check for overall summary with aggregated counts
	if !strings.Contains(ctx.commandOutput, "Summary:") {
		return fmt.Errorf("output does not contain summary section:\n%s", ctx.commandOutput)
	}

	// Should show total counts
	requiredFields := []string{
		"Total modules:",
		"Total errors:",
		"Total warnings:",
	}

	for _, field := range requiredFields {
		if !strings.Contains(ctx.commandOutput, field) {
			return fmt.Errorf("summary does not show '%s':\n%s", field, ctx.commandOutput)
		}
	}

	return nil
}

// ============================================================================
// Scenario Initialization
// ============================================================================

func InitializeDesignScenario(sc *godog.ScenarioContext) {
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		designCtx = &designTestContext{
			dockerAvailable:  false,
			containerStarted: false,
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		// Cleanup: stop and remove test containers
		if designCtx.dockerClient != nil {
			// Try to stop and remove structurizr-cli container
			containerName := "structurizr-cli"
			containers, listErr := designCtx.dockerClient.ContainerList(context.Background(), container.ListOptions{All: true})
			if listErr == nil {
				for _, c := range containers {
					for _, name := range c.Names {
						if strings.TrimPrefix(name, "/") == containerName {
							timeout := 5
							designCtx.dockerClient.ContainerStop(context.Background(), c.ID, container.StopOptions{Timeout: &timeout})
							designCtx.dockerClient.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{Force: true})
							break
						}
					}
				}
			}
			designCtx.dockerClient.Close()
		}
		return ctx, nil
	})

	// Given steps
	sc.Step(`^Docker is running$`, dockerIsRunning)
	sc.Step(`^module "([^"]*)" has workspace\.dsl file$`, moduleHasWorkspaceDslFile)
	sc.Step(`^multiple modules have workspace\.dsl files$`, multipleModulesHaveWorkspaceDslFiles)

	// Then steps - Structurizr serve
	sc.Step(`^Structurizr container should start successfully$`, structurizrContainerShouldStartSuccessfully)
	sc.Step(`^I should see success message with URL$`, iShouldSeeSuccessMessageWithURL)
	sc.Step(`^documentation should be accessible at the URL$`, documentationShouldBeAccessibleAtTheURL)
	sc.Step(`^I should see a list of available modules$`, iShouldSeeAListOfAvailableModules)
	sc.Step(`^"([^"]*)" module should be in the list$`, moduleShouldBeInTheList)

	// Then steps - Validation
	sc.Step(`^the workspace should be validated using Structurizr CLI$`, theWorkspaceShouldBeValidatedUsingStructurizrCLI)
	sc.Step(`^all workspaces should be validated using Structurizr CLI$`, allWorkspacesShouldBeValidatedUsingStructurizrCLI)
	sc.Step(`^validation results should be displayed in console$`, validationResultsShouldBeDisplayedInConsole)
	sc.Step(`^validation results for each module should be displayed in console$`, validationResultsForEachModuleShouldBeDisplayedInConsole)
	sc.Step(`^validation results should be written to JSON file$`, validationResultsShouldBeWrittenToJSONFile)
	sc.Step(`^aggregated validation results should be written to JSON file$`, aggregatedValidationResultsShouldBeWrittenToJSONFile)
	sc.Step(`^I should see validation summary with errors and warnings$`, iShouldSeeValidationSummaryWithErrorsAndWarnings)
	sc.Step(`^I should see overall summary with total errors and warnings$`, iShouldSeeOverallSummaryWithTotalErrorsAndWarnings)
}
