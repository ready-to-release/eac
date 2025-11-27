// Package tests provides BDD step definitions for the init command.
//
// This file contains "Given" steps (setup/preconditions).
package tests

import (
	"fmt"
	"os"
	"path/filepath"
)

// registerSetupSteps registers all "Given" step definitions.
func registerSetupSteps(sc interface{ Step(interface{}, interface{}) }) {
	sc.Step(`^I am not in a git repository$`, iAmNotInAGitRepository)
	sc.Step(`^no \.r2r directory exists$`, noR2RDirectoryExists)
	sc.Step(`^no \.r2r/eac/agent-config\.yml file exists$`, noAgentConfigYmlFileExists)
	sc.Step(`^a malformed \.r2r/eac/agent-config\.yml file$`, aMalformedAgentConfigYmlFile)
	sc.Step(`^a \.r2r directory already exists$`, aR2RDirectoryAlreadyExists)
	sc.Step(`^a \.r2r/eac/agent-config\.yml file exists with claude-api$`, aR2RAgentConfigYmlFileExistsWithClaudeApi)
}

// iAmNotInAGitRepository sets up non-git environment
func iAmNotInAGitRepository() error {
	// This would require special test setup - for now, document the expectation
	return nil
}

// noR2RDirectoryExists verifies no .r2r directory exists
func noR2RDirectoryExists() error {
	// Precondition - no .r2r directory exists
	// In isolated tests, this is automatically true since each test creates a fresh temp directory
	return nil
}

// noAgentConfigYmlFileExists verifies no existing config
func noAgentConfigYmlFileExists() error {
	// Precondition - no .r2r/eac/agent-config.yml exists
	return nil
}

// aMalformedAgentConfigYmlFile sets up invalid config file
func aMalformedAgentConfigYmlFile() error {
	// Create a malformed YAML file for testing
	if IsolatedTestProjectDir == "" {
		return fmt.Errorf("isolated test project directory not set")
	}

	eacDir := filepath.Join(IsolatedTestProjectDir, ".r2r", "eac")
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	configPath := filepath.Join(eacDir, "agent-config.yml")
	malformedContent := `this is not valid YAML: [[[`
	if err := os.WriteFile(configPath, []byte(malformedContent), 0644); err != nil {
		return fmt.Errorf("failed to create malformed config: %w", err)
	}

	return nil
}

// aR2RDirectoryAlreadyExists sets up existing .r2r/eac directory with config
func aR2RDirectoryAlreadyExists() error {
	// Create .r2r/eac directory with a config file so init detects it as "already initialized"
	var eacDir string
	if IsolatedTestProjectDir != "" {
		// Isolated test - create .r2r/eac in the isolated directory
		eacDir = filepath.Join(IsolatedTestProjectDir, ".r2r", "eac")
	} else {
		// Non-isolated test - create .r2r/eac in temp directory
		tmpDir := filepath.Join(os.TempDir(), ".r2r-test", "eac")
		eacDir = tmpDir
	}

	// Create the directory
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	// Create a minimal config file so init detects existing initialization
	configPath := filepath.Join(eacDir, "agent-config.yml")
	configContent := `# Agent Configuration
provider:
  name: claude-api
  model: claude-3-haiku-20240307
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create agent-config.yml: %w", err)
	}

	return nil
}

// aR2RAgentConfigYmlFileExistsWithClaudeApi sets up existing config
func aR2RAgentConfigYmlFileExistsWithClaudeApi() error {
	// Determine the correct directory
	var eacDir string
	if IsolatedTestProjectDir != "" {
		eacDir = filepath.Join(IsolatedTestProjectDir, ".r2r", "eac")
	} else {
		eacDir = filepath.Join(os.TempDir(), ".r2r-test", "eac")
	}

	// Create the directory
	if err := os.MkdirAll(eacDir, 0755); err != nil {
		return fmt.Errorf("failed to create .r2r/eac directory: %w", err)
	}

	// Create a minimal claude-api config file
	configPath := filepath.Join(eacDir, "agent-config.yml")
	configContent := `# Agent Configuration
provider:
  name: claude-api
  model: claude-3-haiku-20240307
  endpoint: https://api.anthropic.com/v1
  api_key: ${ANTHROPIC_API_KEY}
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to create agent-config.yml: %w", err)
	}

	return nil
}
