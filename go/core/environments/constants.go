// Package environments provides shared environment variable constants
// used across commands, specs, and core infrastructure.
//
// All application-specific environment variables should be defined here
// to provide a single source of truth and eliminate typo risks from
// hardcoded string literals.
//
// # Excluded Variables
//
// The following variables are intentionally NOT added as constants:
//   - Standard OS variables: COLUMNS, LINES, TERM, HOSTNAME, HOME, USER
//   - Third-party API keys: OPENAI_API_KEY, ANTHROPIC_API_KEY, etc.
//   - Go toolchain internals: GOPATH, GOROOT
//
// # Usage
//
//	import "github.com/ready-to-release/eac/go/core/environments"
//
//	if containerRoot := os.Getenv(environments.EnvCLIEContainerRoot); containerRoot != "" {
//	    // use container root
//	}
package environments

// Environment variable names used in test and execution contexts.
const (
	// Test infrastructure environment variables.
	EnvCLIETestLoggingActive = "CLIE_TEST_LOGGING_ACTIVE"
	EnvCLIETestRunID         = "CLIE_TEST_RUN_ID"
	EnvCLIETestScope         = "CLIE_TEST_SCOPE" // Set when running within spec test scope
	EnvGodogFormat          = "GODOG_FORMAT"

	// Repository path environment variables.
	EnvCLIEPWD           = "CLIE_PWD"
	EnvCLIERepoRoot      = "CLIE_REPO_ROOT"
	EnvCLIEContainerRoot = "CLIE_CONTAINER_ROOT"
	EnvCLIEDockerMode    = "CLIE_DOCKER_MODE"

	// Docker and container runtime configuration.
	EnvCLIEHostRepoRoot      = "CLIE_HOST_REPOROOT"
	EnvCLIEContainerRepoRoot = "CLIE_CONTAINER_REPOROOT"
	EnvCLIEDockerHost        = "CLIE_DOCKER_HOST"
	EnvCLIEHostGOOS          = "CLIE_HOST_GOOS"
	EnvCLIEHostGOARCH        = "CLIE_HOST_GOARCH"
	EnvCLIETerminalDetection = "CLIE_TERMINAL_DETECTION"

	// Application configuration and behavior.
	EnvCLIEConfig         = "CLIE_CONFIG"
	EnvCLIEConfigPath     = "CLIE_CONFIG_PATH"
	EnvCLIEContext        = "CLIE_CONTEXT"
	EnvCLIENoBrowser      = "CLIE_NO_BROWSER"
	EnvCLIENoUpdateCheck  = "CLIE_NO_UPDATE_CHECK"
	EnvCLIESkipPinWarning = "CLIE_SKIP_PIN_WARNING"
	EnvCLIEFixedRedirect  = "CLIE_FIXED_REDIRECT"

	// Debug and logging configuration.
	EnvCLIEDebug      = "CLIE_DEBUG"
	EnvEACDebug      = "EAC_DEBUG"
	EnvCLIELogLevel   = "CLIE_LOG_LEVEL"
	EnvCLIEVerboseLog = "CLIE_VERBOSE_LOG"

	// Testing infrastructure.
	EnvCLIETesting      = "CLIE_TESTING"
	EnvCLIECheckTags    = "CLIE_CHECK_TAGS"
	EnvCLIEOriginalArgs = "CLIE_ORIGINAL_ARGS"
	EnvCLIEFilteredArgs = "CLIE_FILTERED_ARGS"

	// Build and execution.
	EnvEACUseDirectBinary = "EAC_USE_DIRECT_BINARY"

	// Godog BDD test framework.
	EnvGodogSuiteTags  = "GODOG_SUITE_TAGS"
	EnvGodogOutputDir  = "GODOG_OUTPUT_DIR"
	EnvGodogPaths      = "GODOG_PATHS"
	EnvGodogReportName = "GODOG_REPORT_NAME"
	EnvGodogDebugInit  = "GODOG_DEBUG_INIT"

	// Mock and test configuration.
	EnvCLIEMockAIDir              = "CLIE_MOCK_AI_DIR"
	EnvCLIEMockSecurity           = "CLIE_MOCK_SECURITY"
	EnvCLIEMockDocker             = "CLIE_MOCK_DOCKER"
	EnvCLIEMockGitHubCLI          = "CLIE_MOCK_GITHUB_CLI"
	EnvCLIEMockNoWorkflows        = "CLIE_MOCK_NO_WORKFLOWS"
	EnvCLIEMockAI                 = "CLIE_MOCK_AI"
	EnvCLIEMockSecurityTools      = "CLIE_MOCK_SECURITY_TOOLS"
	EnvCLIEMockGitHub             = "CLIE_MOCK_GITHUB"
	EnvCLIEMockGitHubNoWorkflows  = "CLIE_MOCK_GITHUB_NO_WORKFLOWS"
	EnvCLIETestMock               = "__CLIE_TEST_MOCK"
	EnvDebugToolResolve          = "DEBUG_TOOL_RESOLVE"
	EnvDebugChangeDetect         = "DEBUG_CHANGEDETECT"
	EnvDebugCacheCmd             = "DEBUG_CACHE_CMD"
	EnvCLIEMockCIStatus           = "CLIE_MOCK_CI_STATUS"
	EnvCLIEMockHeadSHA            = "CLIE_MOCK_HEAD_SHA"
	EnvCLIEMockChangedFiles       = "CLIE_MOCK_CHANGED_FILES"
	EnvCLIETestAIResponse         = "CLIE_TEST_AI_RESPONSE"
	EnvCLIEMockTimeout            = "CLIE_MOCK_TIMEOUT"
	EnvCLIEMockInvalidRef         = "CLIE_MOCK_INVALID_REF"
	EnvCLIEMockFailingWorkflow    = "CLIE_MOCK_FAILING_WORKFLOW"
	EnvSkipDockerTests           = "SKIP_DOCKER_TESTS"
	EnvFilesByModule             = "FILES_BY_MODULE"
	EnvModuleStatus              = "MODULE_STATUS"

	// CI environment detection.
	EnvCI            = "CI"
	EnvGitHubActions = "GITHUB_ACTIONS"
	EnvGitLabCI      = "GITLAB_CI"

	// GitHub CI metadata.
	EnvGitHubSHA        = "GITHUB_SHA"
	EnvGitHubToken      = "GITHUB_TOKEN"
	EnvGitHubUsername   = "GITHUB_USERNAME"
	EnvGitHubActor      = "GITHUB_ACTOR"
	EnvGitHubRunID      = "GITHUB_RUN_ID"
	EnvGitHubRepository = "GITHUB_REPOSITORY"
	EnvGitHubEnv        = "GITHUB_ENV"
	EnvGHRepo           = "GH_REPO"

	// Build and system configuration.
	EnvCommandsPath      = "COMMANDS_PATH"
	EnvGOOS              = "GOOS"
	EnvEACPortRangeStart = "EAC_PORT_RANGE_START"
	EnvEACPortRangeEnd   = "EAC_PORT_RANGE_END"

	// Git configuration.
	EnvGitAuthorName  = "GIT_AUTHOR_NAME"
	EnvGitAuthorEmail = "GIT_AUTHOR_EMAIL"
)
