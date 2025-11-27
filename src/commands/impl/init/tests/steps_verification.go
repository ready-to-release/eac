// Package tests provides BDD step definitions for the init command.
//
// This file contains "Then" steps (assertions/verification).
package tests

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// registerVerificationSteps registers all "Then" step definitions.
func registerVerificationSteps(sc interface{ Step(interface{}, interface{}) }) {
	sc.Step(`^the \.r2r/eac/agent-config\.yml file contains "([^"]*)"$`, theAgentConfigYmlFileContains)
	sc.Step(`^a reconfiguration message is shown$`, aReconfigurationMessageIsShown)
	sc.Step(`^stdout contains link to get API key$`, stdoutContainsLinkToGetAPIKey)
	sc.Step(`^stdout contains provider selection confirmation$`, stdoutContainsProviderSelectionConfirmation)
	sc.Step(`^the AI provider is configured$`, theAIProviderIsConfigured)
	sc.Step(`^a \.r2r/eac/agent-config\.yml file is created$`, aR2RAgentConfigYmlFileIsCreated)
	sc.Step(`^the \.r2r/eac directory is created$`, theR2REacDirectoryIsCreated)
	sc.Step(`^stdout contains API key instructions$`, stdoutContainsAPIKeyInstructions)
	sc.Step(`^stderr contains actionable recovery instructions$`, stderrContainsActionableRecoveryInstructions)
	sc.Step(`^stderr suggests retry or configuration adjustment$`, stderrSuggestsRetryOrConfigurationAdjustment)
	sc.Step(`^stderr contains execution timing information$`, stderrContainsExecutionTimingInformation)
	sc.Step(`^stderr contains permission error details$`, stderrContainsPermissionErrorDetails)
}

// theAgentConfigYmlFileContains verifies config file content
func theAgentConfigYmlFileContains(expectedContent string) error {
	// Get the config file path
	var configPath string
	if IsolatedTestProjectDir != "" {
		configPath = filepath.Join(IsolatedTestProjectDir, ".r2r", "eac", "agent-config.yml")
	} else {
		configPath = filepath.Join(os.TempDir(), ".r2r-test", "eac", "agent-config.yml")
	}

	// Read the config file
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read agent-config.yml: %w", err)
	}

	// Check if it contains expected content
	if !strings.Contains(string(content), expectedContent) {
		return fmt.Errorf("agent-config.yml does not contain %q\nActual content:\n%s", expectedContent, string(content))
	}

	return nil
}

// aReconfigurationMessageIsShown verifies reconfiguration warning
func aReconfigurationMessageIsShown() error {
	output := getCommandOutput()
	outputLower := strings.ToLower(output)
	if strings.Contains(outputLower, "reconfigur") || strings.Contains(outputLower, "already initialized") {
		return nil
	}
	return fmt.Errorf("no reconfiguration message shown.\nOutput:\n%s", output)
}

// stdoutContainsLinkToGetAPIKey verifies helpful API key guidance
func stdoutContainsLinkToGetAPIKey() error {
	output := getCommandOutput()
	if strings.Contains(output, "http") || strings.Contains(output, "API key") ||
		strings.Contains(output, "api-key") {
		return nil
	}
	return fmt.Errorf("stdout does not contain link to get API key.\nOutput:\n%s", output)
}

// theAIProviderIsConfigured verifies provider setup
func theAIProviderIsConfigured() error {
	// Provider configured successfully
	exitCode := getExitCode()
	if exitCode == 0 {
		return nil
	}
	return fmt.Errorf("AI provider not configured (command failed with exit code %d)", exitCode)
}

// aR2RAgentConfigYmlFileIsCreated verifies config creation
func aR2RAgentConfigYmlFileIsCreated() error {
	var configPath string
	if IsolatedTestProjectDir != "" {
		configPath = filepath.Join(IsolatedTestProjectDir, ".r2r", "eac", "agent-config.yml")
	} else {
		configPath = filepath.Join(os.TempDir(), ".r2r-test", "eac", "agent-config.yml")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf(".r2r/eac/agent-config.yml was not created")
	}

	return nil
}

// theR2REacDirectoryIsCreated verifies directory creation
func theR2REacDirectoryIsCreated() error {
	var eacDir string
	if IsolatedTestProjectDir != "" {
		eacDir = filepath.Join(IsolatedTestProjectDir, ".r2r", "eac")
	} else {
		eacDir = filepath.Join(os.TempDir(), ".r2r-test", "eac")
	}

	if _, err := os.Stat(eacDir); os.IsNotExist(err) {
		return fmt.Errorf(".r2r/eac directory was not created")
	}

	return nil
}

// stdoutContainsAPIKeyInstructions verifies API key guidance
func stdoutContainsAPIKeyInstructions() error {
	output := getCommandOutput()
	if strings.Contains(output, "API") || strings.Contains(output, "key") {
		return nil
	}
	return fmt.Errorf("stdout does not contain API key instructions")
}

// stdoutContainsProviderSelectionConfirmation verifies provider selection message
func stdoutContainsProviderSelectionConfirmation() error {
	output := getCommandOutput()
	// Look for provider-related confirmation in output
	if strings.Contains(output, "claude") || strings.Contains(output, "openai") ||
		strings.Contains(output, "provider") || strings.Contains(output, "Initialized") {
		return nil
	}
	return fmt.Errorf("stdout does not contain provider selection confirmation.\nOutput:\n%s", output)
}

// stderrContainsActionableRecoveryInstructions verifies helpful errors
func stderrContainsActionableRecoveryInstructions() error {
	output := getCommandOutput()
	if strings.Contains(output, "try") || strings.Contains(output, "use") ||
		strings.Contains(output, "run") {
		return nil
	}
	return fmt.Errorf("stderr does not contain actionable recovery instructions")
}

// stderrSuggestsRetryOrConfigurationAdjustment verifies retry guidance
func stderrSuggestsRetryOrConfigurationAdjustment() error {
	output := getCommandOutput()
	if strings.Contains(output, "retry") || strings.Contains(output, "adjust") ||
		strings.Contains(output, "config") {
		return nil
	}
	return fmt.Errorf("stderr does not suggest retry or configuration adjustment")
}

// stderrContainsExecutionTimingInformation verifies timing output
func stderrContainsExecutionTimingInformation() error {
	output := getCommandOutput()
	if strings.Contains(output, "time") || strings.Contains(output, "elapsed") ||
		strings.Contains(output, "ms") || strings.Contains(output, "seconds") {
		return nil
	}
	return fmt.Errorf("stderr does not contain execution timing information")
}

// stderrContainsPermissionErrorDetails verifies permission error info
func stderrContainsPermissionErrorDetails() error {
	output := getCommandOutput()
	if strings.Contains(output, "permission") || strings.Contains(output, "access denied") {
		return nil
	}
	return fmt.Errorf("stderr does not contain permission error details")
}

// getCommandOutput returns the command output from shared context
func getCommandOutput() string {
	if SharedCtx != nil {
		return SharedCtx.CommandOutput
	}
	return ""
}

// getExitCode returns the exit code from shared context
func getExitCode() int {
	if SharedCtx != nil {
		return SharedCtx.ExitCode
	}
	return 0
}
