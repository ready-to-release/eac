// Package environments provides shared environment variable constants
// used across commands, specs, and core infrastructure.
package environments

// Environment variable names used in test and execution contexts.
const (
	// Test infrastructure environment variables.
	EnvR2RTestLoggingActive = "R2R_TEST_LOGGING_ACTIVE"
	EnvR2RTestRunID         = "R2R_TEST_RUN_ID"
	EnvR2RTestScope         = "R2R_TEST_SCOPE" // Set when running within spec test scope
	EnvGodogFormat          = "GODOG_FORMAT"

	// Repository path environment variables.
	EnvR2RPWD           = "R2R_PWD"
	EnvR2RRepoRoot      = "R2R_REPO_ROOT"
	EnvR2RContainerRoot = "R2R_CONTAINER_ROOT"

	// Mock environment variables for testing.
	EnvR2RMockAIDir        = "R2R_MOCK_AI_DIR"
	EnvR2RMockSecurity     = "R2R_MOCK_SECURITY"
	EnvR2RMockStructurizr  = "R2R_MOCK_STRUCTURIZR" // Legacy: use EnvR2RMockDocker instead
	EnvR2RMockDocker       = "R2R_MOCK_DOCKER"       // Preferred for all Docker mocking
	EnvR2RMockGitHubCLI    = "R2R_MOCK_GITHUB_CLI"
	EnvR2RMockNoWorkflows  = "R2R_MOCK_NO_WORKFLOWS"
)
