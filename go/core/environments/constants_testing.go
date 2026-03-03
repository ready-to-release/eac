package environments

// Test infrastructure and BDD framework constants.
const (
	// Test infrastructure environment variables.
	EnvCLIETestLoggingActive = "CLIE_TEST_LOGGING_ACTIVE"
	EnvCLIETestRunID         = "CLIE_TEST_RUN_ID"
	EnvCLIETestScope         = "CLIE_TEST_SCOPE" // Set when running within spec test scope

	// Testing infrastructure.
	EnvCLIETesting      = "CLIE_TESTING"
	EnvCLIECheckTags    = "CLIE_CHECK_TAGS"
	EnvCLIEOriginalArgs = "CLIE_ORIGINAL_ARGS"
	EnvCLIEFilteredArgs = "CLIE_FILTERED_ARGS"

	// Godog BDD test framework.
	EnvGodogFormat     = "GODOG_FORMAT"
	EnvGodogSuiteTags  = "GODOG_SUITE_TAGS"
	EnvGodogOutputDir  = "GODOG_OUTPUT_DIR"
	EnvGodogPaths      = "GODOG_PATHS"
	EnvGodogReportName = "GODOG_REPORT_NAME"
	EnvGodogDebugInit  = "GODOG_DEBUG_INIT"

	// Non-mock test support.
	EnvSkipDockerTests = "SKIP_DOCKER_TESTS"
	EnvFilesByModule   = "FILES_BY_MODULE"
	EnvModuleStatus    = "MODULE_STATUS"
)

// Mock configuration constants for test doubles and fixture injection.
const (
	EnvCLIEMockAIDir             = "CLIE_MOCK_AI_DIR"
	EnvCLIEMockSecurity          = "CLIE_MOCK_SECURITY"
	EnvCLIEMockDocker            = "CLIE_MOCK_DOCKER"
	EnvCLIEMockGitHubCLI         = "CLIE_MOCK_GITHUB_CLI"
	EnvCLIEMockNoWorkflows       = "CLIE_MOCK_NO_WORKFLOWS"
	EnvCLIEMockAI                = "CLIE_MOCK_AI"
	EnvCLIEMockSecurityTools     = "CLIE_MOCK_SECURITY_TOOLS"
	EnvCLIEMockGitHub            = "CLIE_MOCK_GITHUB"
	EnvCLIEMockGitHubNoWorkflows = "CLIE_MOCK_GITHUB_NO_WORKFLOWS"
	EnvCLIETestMock              = "__CLIE_TEST_MOCK"
	EnvCLIEMockCIStatus          = "CLIE_MOCK_CI_STATUS"
	EnvCLIEMockHeadSHA           = "CLIE_MOCK_HEAD_SHA"
	EnvCLIEMockChangedFiles      = "CLIE_MOCK_CHANGED_FILES"
	EnvCLIETestAIResponse        = "CLIE_TEST_AI_RESPONSE"
	EnvCLIEMockTimeout           = "CLIE_MOCK_TIMEOUT"
	EnvCLIEMockInvalidRef        = "CLIE_MOCK_INVALID_REF"
	EnvCLIEMockFailingWorkflow   = "CLIE_MOCK_FAILING_WORKFLOW"

	// EAC mock environment variables (for eac commands, not clie)
	EnvEACMockContainerRegistry = "EAC_MOCK_CONTAINER_REGISTRY"
)
