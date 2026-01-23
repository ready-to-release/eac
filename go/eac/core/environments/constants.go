// Package environments provides shared environment variable constants
// used across commands, specs, and core infrastructure.
package environments

// Environment variable names used in test and execution contexts.
const (
	// Test infrastructure environment variables
	EnvR2RTestLoggingActive = "R2R_TEST_LOGGING_ACTIVE"
	EnvR2RTestRunID         = "R2R_TEST_RUN_ID"
	EnvGodogFormat          = "GODOG_FORMAT"

	// Repository path environment variables
	EnvR2RPWD           = "R2R_PWD"
	EnvR2RRepoRoot      = "R2R_REPO_ROOT"
	EnvR2RContainerRoot = "R2R_CONTAINER_ROOT"

	// Mock environment variables for testing
	EnvR2RMockAIDir       = "R2R_MOCK_AI_DIR"
	EnvR2RMockSecurity    = "R2R_MOCK_SECURITY"
	EnvR2RMockStructurizr = "R2R_MOCK_STRUCTURIZR"
)
