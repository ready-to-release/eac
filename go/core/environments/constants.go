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
//	if containerRoot := os.Getenv(environments.EnvR2RContainerRoot); containerRoot != "" {
//	    // use container root
//	}
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
	EnvR2RDockerMode    = "R2R_DOCKER_MODE"

	// Docker and container runtime configuration.
	EnvR2RHostRepoRoot      = "R2R_HOST_REPOROOT"
	EnvR2RContainerRepoRoot = "R2R_CONTAINER_REPOROOT"
	EnvR2RDockerHost        = "R2R_DOCKER_HOST"
	EnvR2RHostGOOS          = "R2R_HOST_GOOS"
	EnvR2RHostGOARCH        = "R2R_HOST_GOARCH"
	EnvR2RTerminalDetection = "R2R_TERMINAL_DETECTION"

	// Application configuration and behavior.
	EnvR2RConfig         = "R2R_CONFIG"
	EnvR2RConfigPath     = "R2R_CONFIG_PATH"
	EnvR2RContext        = "R2R_CONTEXT"
	EnvR2RNoBrowser      = "R2R_NO_BROWSER"
	EnvR2RNoUpdateCheck  = "R2R_NO_UPDATE_CHECK"
	EnvR2RSkipPinWarning = "R2R_SKIP_PIN_WARNING"
	EnvR2RFixedRedirect  = "R2R_FIXED_REDIRECT"

	// Debug and logging configuration.
	EnvR2RDebug      = "R2R_DEBUG"
	EnvEACDebug      = "EAC_DEBUG"
	EnvR2RLogLevel   = "R2R_LOG_LEVEL"
	EnvR2RVerboseLog = "R2R_VERBOSE_LOG"

	// Testing infrastructure.
	EnvR2RTesting      = "R2R_TESTING"
	EnvR2RCheckTags    = "R2R_CHECK_TAGS"
	EnvR2ROriginalArgs = "R2R_ORIGINAL_ARGS"
	EnvR2RFilteredArgs = "R2R_FILTERED_ARGS"

	// Build and execution.
	EnvEACUseDirectBinary = "EAC_USE_DIRECT_BINARY"

	// Godog BDD test framework.
	EnvGodogSuiteTags  = "GODOG_SUITE_TAGS"
	EnvGodogOutputDir  = "GODOG_OUTPUT_DIR"
	EnvGodogPaths      = "GODOG_PATHS"
	EnvGodogReportName = "GODOG_REPORT_NAME"
	EnvGodogDebugInit  = "GODOG_DEBUG_INIT"

	// Mock and test configuration.
	EnvR2RMockAIDir              = "R2R_MOCK_AI_DIR"
	EnvR2RMockSecurity           = "R2R_MOCK_SECURITY"
	EnvR2RMockDocker             = "R2R_MOCK_DOCKER"
	EnvR2RMockGitHubCLI          = "R2R_MOCK_GITHUB_CLI"
	EnvR2RMockNoWorkflows        = "R2R_MOCK_NO_WORKFLOWS"
	EnvR2RMockAI                 = "R2R_MOCK_AI"
	EnvR2RMockSecurityTools      = "R2R_MOCK_SECURITY_TOOLS"
	EnvR2RMockGitHub             = "R2R_MOCK_GITHUB"
	EnvR2RMockGitHubNoWorkflows  = "R2R_MOCK_GITHUB_NO_WORKFLOWS"
	EnvR2RTestMock               = "__R2R_TEST_MOCK"
	EnvDebugToolResolve          = "DEBUG_TOOL_RESOLVE"
	EnvDebugChangeDetect         = "DEBUG_CHANGEDETECT"
	EnvDebugCacheCmd             = "DEBUG_CACHE_CMD"
	EnvR2RMockCIStatus           = "R2R_MOCK_CI_STATUS"
	EnvR2RMockHeadSHA            = "R2R_MOCK_HEAD_SHA"
	EnvR2RMockChangedFiles       = "R2R_MOCK_CHANGED_FILES"
	EnvR2RTestAIResponse         = "R2R_TEST_AI_RESPONSE"
	EnvR2RMockTimeout            = "R2R_MOCK_TIMEOUT"
	EnvR2RMockInvalidRef         = "R2R_MOCK_INVALID_REF"
	EnvR2RMockFailingWorkflow    = "R2R_MOCK_FAILING_WORKFLOW"
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
