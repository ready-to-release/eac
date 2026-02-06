package environments_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ready-to-release/eac/go/core/environments"
)

// TestConstantValues verifies that all environment variable constants
// match their expected string values. This ensures we don't accidentally
// change constant values during refactoring, which would break compatibility.
func TestConstantValues(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		// Test infrastructure environment variables
		{
			name:     "R2R_TEST_LOGGING_ACTIVE",
			constant: environments.EnvR2RTestLoggingActive,
			expected: "R2R_TEST_LOGGING_ACTIVE",
		},
		{
			name:     "R2R_TEST_RUN_ID",
			constant: environments.EnvR2RTestRunID,
			expected: "R2R_TEST_RUN_ID",
		},
		{
			name:     "R2R_TEST_SCOPE",
			constant: environments.EnvR2RTestScope,
			expected: "R2R_TEST_SCOPE",
		},
		{
			name:     "GODOG_FORMAT",
			constant: environments.EnvGodogFormat,
			expected: "GODOG_FORMAT",
		},

		// Godog framework
		{name: "GODOG_SUITE_TAGS", constant: environments.EnvGodogSuiteTags, expected: "GODOG_SUITE_TAGS"},
		{name: "GODOG_OUTPUT_DIR", constant: environments.EnvGodogOutputDir, expected: "GODOG_OUTPUT_DIR"},
		{name: "GODOG_PATHS", constant: environments.EnvGodogPaths, expected: "GODOG_PATHS"},
		{name: "GODOG_REPORT_NAME", constant: environments.EnvGodogReportName, expected: "GODOG_REPORT_NAME"},
		{name: "GODOG_DEBUG_INIT", constant: environments.EnvGodogDebugInit, expected: "GODOG_DEBUG_INIT"},

		// Repository path environment variables
		{
			name:     "R2R_PWD",
			constant: environments.EnvR2RPWD,
			expected: "R2R_PWD",
		},
		{
			name:     "R2R_REPO_ROOT",
			constant: environments.EnvR2RRepoRoot,
			expected: "R2R_REPO_ROOT",
		},
		{
			name:     "R2R_CONTAINER_ROOT",
			constant: environments.EnvR2RContainerRoot,
			expected: "R2R_CONTAINER_ROOT",
		},
		{
			name:     "R2R_DOCKER_MODE",
			constant: environments.EnvR2RDockerMode,
			expected: "R2R_DOCKER_MODE",
		},

		// Docker and container runtime
		{
			name:     "R2R_HOST_REPOROOT",
			constant: environments.EnvR2RHostRepoRoot,
			expected: "R2R_HOST_REPOROOT",
		},
		{
			name:     "R2R_CONTAINER_REPOROOT",
			constant: environments.EnvR2RContainerRepoRoot,
			expected: "R2R_CONTAINER_REPOROOT",
		},
		{
			name:     "R2R_DOCKER_HOST",
			constant: environments.EnvR2RDockerHost,
			expected: "R2R_DOCKER_HOST",
		},
		{
			name:     "R2R_HOST_GOOS",
			constant: environments.EnvR2RHostGOOS,
			expected: "R2R_HOST_GOOS",
		},
		{
			name:     "R2R_HOST_GOARCH",
			constant: environments.EnvR2RHostGOARCH,
			expected: "R2R_HOST_GOARCH",
		},
		{
			name:     "R2R_TERMINAL_DETECTION",
			constant: environments.EnvR2RTerminalDetection,
			expected: "R2R_TERMINAL_DETECTION",
		},

		// Application configuration
		{
			name:     "R2R_CONFIG",
			constant: environments.EnvR2RConfig,
			expected: "R2R_CONFIG",
		},
		{
			name:     "R2R_CONFIG_PATH",
			constant: environments.EnvR2RConfigPath,
			expected: "R2R_CONFIG_PATH",
		},
		{
			name:     "R2R_CONTEXT",
			constant: environments.EnvR2RContext,
			expected: "R2R_CONTEXT",
		},
		{
			name:     "R2R_NO_BROWSER",
			constant: environments.EnvR2RNoBrowser,
			expected: "R2R_NO_BROWSER",
		},
		{
			name:     "R2R_NO_UPDATE_CHECK",
			constant: environments.EnvR2RNoUpdateCheck,
			expected: "R2R_NO_UPDATE_CHECK",
		},
		{
			name:     "R2R_SKIP_PIN_WARNING",
			constant: environments.EnvR2RSkipPinWarning,
			expected: "R2R_SKIP_PIN_WARNING",
		},
		{
			name:     "R2R_FIXED_REDIRECT",
			constant: environments.EnvR2RFixedRedirect,
			expected: "R2R_FIXED_REDIRECT",
		},

		// Debug and logging
		{
			name:     "R2R_DEBUG",
			constant: environments.EnvR2RDebug,
			expected: "R2R_DEBUG",
		},
		{
			name:     "EAC_DEBUG",
			constant: environments.EnvEACDebug,
			expected: "EAC_DEBUG",
		},
		{
			name:     "R2R_LOG_LEVEL",
			constant: environments.EnvR2RLogLevel,
			expected: "R2R_LOG_LEVEL",
		},
		{
			name:     "R2R_VERBOSE_LOG",
			constant: environments.EnvR2RVerboseLog,
			expected: "R2R_VERBOSE_LOG",
		},

		// Testing
		{
			name:     "R2R_TESTING",
			constant: environments.EnvR2RTesting,
			expected: "R2R_TESTING",
		},
		{
			name:     "R2R_CHECK_TAGS",
			constant: environments.EnvR2RCheckTags,
			expected: "R2R_CHECK_TAGS",
		},
		{
			name:     "R2R_ORIGINAL_ARGS",
			constant: environments.EnvR2ROriginalArgs,
			expected: "R2R_ORIGINAL_ARGS",
		},
		{
			name:     "R2R_FILTERED_ARGS",
			constant: environments.EnvR2RFilteredArgs,
			expected: "R2R_FILTERED_ARGS",
		},

		// Build
		{
			name:     "EAC_USE_DIRECT_BINARY",
			constant: environments.EnvEACUseDirectBinary,
			expected: "EAC_USE_DIRECT_BINARY",
		},

		// Build and system
		{name: "COMMANDS_PATH", constant: environments.EnvCommandsPath, expected: "COMMANDS_PATH"},
		{name: "GOOS", constant: environments.EnvGOOS, expected: "GOOS"},
		{name: "EAC_PORT_RANGE_START", constant: environments.EnvEACPortRangeStart, expected: "EAC_PORT_RANGE_START"},
		{name: "EAC_PORT_RANGE_END", constant: environments.EnvEACPortRangeEnd, expected: "EAC_PORT_RANGE_END"},

		// Git configuration
		{name: "GIT_AUTHOR_NAME", constant: environments.EnvGitAuthorName, expected: "GIT_AUTHOR_NAME"},
		{name: "GIT_AUTHOR_EMAIL", constant: environments.EnvGitAuthorEmail, expected: "GIT_AUTHOR_EMAIL"},

		// Additional mock and test configuration
		{name: "R2R_MOCK_AI", constant: environments.EnvR2RMockAI, expected: "R2R_MOCK_AI"},
		{name: "R2R_MOCK_SECURITY_TOOLS", constant: environments.EnvR2RMockSecurityTools, expected: "R2R_MOCK_SECURITY_TOOLS"},
		{name: "R2R_MOCK_GITHUB", constant: environments.EnvR2RMockGitHub, expected: "R2R_MOCK_GITHUB"},
		{name: "R2R_MOCK_GITHUB_NO_WORKFLOWS", constant: environments.EnvR2RMockGitHubNoWorkflows, expected: "R2R_MOCK_GITHUB_NO_WORKFLOWS"},
		{name: "__R2R_TEST_MOCK", constant: environments.EnvR2RTestMock, expected: "__R2R_TEST_MOCK"},
		{name: "DEBUG_TOOL_RESOLVE", constant: environments.EnvDebugToolResolve, expected: "DEBUG_TOOL_RESOLVE"},
		{name: "DEBUG_CHANGEDETECT", constant: environments.EnvDebugChangeDetect, expected: "DEBUG_CHANGEDETECT"},
		{name: "DEBUG_CACHE_CMD", constant: environments.EnvDebugCacheCmd, expected: "DEBUG_CACHE_CMD"},
		{name: "R2R_MOCK_CI_STATUS", constant: environments.EnvR2RMockCIStatus, expected: "R2R_MOCK_CI_STATUS"},
		{name: "R2R_MOCK_HEAD_SHA", constant: environments.EnvR2RMockHeadSHA, expected: "R2R_MOCK_HEAD_SHA"},
		{name: "R2R_MOCK_CHANGED_FILES", constant: environments.EnvR2RMockChangedFiles, expected: "R2R_MOCK_CHANGED_FILES"},
		{name: "R2R_TEST_AI_RESPONSE", constant: environments.EnvR2RTestAIResponse, expected: "R2R_TEST_AI_RESPONSE"},
		{name: "R2R_MOCK_TIMEOUT", constant: environments.EnvR2RMockTimeout, expected: "R2R_MOCK_TIMEOUT"},
		{name: "R2R_MOCK_INVALID_REF", constant: environments.EnvR2RMockInvalidRef, expected: "R2R_MOCK_INVALID_REF"},
		{name: "R2R_MOCK_FAILING_WORKFLOW", constant: environments.EnvR2RMockFailingWorkflow, expected: "R2R_MOCK_FAILING_WORKFLOW"},
		{name: "SKIP_DOCKER_TESTS", constant: environments.EnvSkipDockerTests, expected: "SKIP_DOCKER_TESTS"},
		{name: "FILES_BY_MODULE", constant: environments.EnvFilesByModule, expected: "FILES_BY_MODULE"},
		{name: "MODULE_STATUS", constant: environments.EnvModuleStatus, expected: "MODULE_STATUS"},

		// CI environment detection
		{name: "CI", constant: environments.EnvCI, expected: "CI"},
		{name: "GITHUB_ACTIONS", constant: environments.EnvGitHubActions, expected: "GITHUB_ACTIONS"},
		{name: "GITLAB_CI", constant: environments.EnvGitLabCI, expected: "GITLAB_CI"},

		// GitHub CI metadata
		{name: "GITHUB_SHA", constant: environments.EnvGitHubSHA, expected: "GITHUB_SHA"},
		{name: "GITHUB_TOKEN", constant: environments.EnvGitHubToken, expected: "GITHUB_TOKEN"},
		{name: "GITHUB_USERNAME", constant: environments.EnvGitHubUsername, expected: "GITHUB_USERNAME"},
		{name: "GITHUB_ACTOR", constant: environments.EnvGitHubActor, expected: "GITHUB_ACTOR"},
		{name: "GITHUB_RUN_ID", constant: environments.EnvGitHubRunID, expected: "GITHUB_RUN_ID"},
		{name: "GITHUB_REPOSITORY", constant: environments.EnvGitHubRepository, expected: "GITHUB_REPOSITORY"},
		{name: "GITHUB_ENV", constant: environments.EnvGitHubEnv, expected: "GITHUB_ENV"},
		{name: "GH_REPO", constant: environments.EnvGHRepo, expected: "GH_REPO"},

		// Mock environment variables for testing
		{
			name:     "R2R_MOCK_AI_DIR",
			constant: environments.EnvR2RMockAIDir,
			expected: "R2R_MOCK_AI_DIR",
		},
		{
			name:     "R2R_MOCK_SECURITY",
			constant: environments.EnvR2RMockSecurity,
			expected: "R2R_MOCK_SECURITY",
		},
		{
			name:     "R2R_MOCK_STRUCTURIZR",
			constant: environments.EnvR2RMockStructurizr,
			expected: "R2R_MOCK_STRUCTURIZR",
		},
		{
			name:     "R2R_MOCK_DOCKER",
			constant: environments.EnvR2RMockDocker,
			expected: "R2R_MOCK_DOCKER",
		},
		{
			name:     "R2R_MOCK_GITHUB_CLI",
			constant: environments.EnvR2RMockGitHubCLI,
			expected: "R2R_MOCK_GITHUB_CLI",
		},
		{
			name:     "R2R_MOCK_NO_WORKFLOWS",
			constant: environments.EnvR2RMockNoWorkflows,
			expected: "R2R_MOCK_NO_WORKFLOWS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant,
				"constant value mismatch: expected %q but got %q",
				tt.expected, tt.constant)
		})
	}
}

// TestConstantUniqueness verifies that all constants have unique values.
// This prevents accidental duplication during refactoring.
func TestConstantUniqueness(t *testing.T) {
	constants := map[string]string{
		"EnvR2RTestLoggingActive":  environments.EnvR2RTestLoggingActive,
		"EnvR2RTestRunID":          environments.EnvR2RTestRunID,
		"EnvR2RTestScope":          environments.EnvR2RTestScope,
		"EnvGodogFormat":           environments.EnvGodogFormat,
		"EnvGodogSuiteTags":        environments.EnvGodogSuiteTags,
		"EnvGodogOutputDir":        environments.EnvGodogOutputDir,
		"EnvGodogPaths":            environments.EnvGodogPaths,
		"EnvGodogReportName":       environments.EnvGodogReportName,
		"EnvGodogDebugInit":        environments.EnvGodogDebugInit,
		"EnvR2RPWD":                environments.EnvR2RPWD,
		"EnvR2RRepoRoot":           environments.EnvR2RRepoRoot,
		"EnvR2RContainerRoot":      environments.EnvR2RContainerRoot,
		"EnvR2RDockerMode":         environments.EnvR2RDockerMode,
		"EnvR2RHostRepoRoot":       environments.EnvR2RHostRepoRoot,
		"EnvR2RContainerRepoRoot":  environments.EnvR2RContainerRepoRoot,
		"EnvR2RDockerHost":         environments.EnvR2RDockerHost,
		"EnvR2RHostGOOS":           environments.EnvR2RHostGOOS,
		"EnvR2RHostGOARCH":         environments.EnvR2RHostGOARCH,
		"EnvR2RConfig":             environments.EnvR2RConfig,
		"EnvR2RConfigPath":         environments.EnvR2RConfigPath,
		"EnvR2RContext":            environments.EnvR2RContext,
		"EnvR2RNoBrowser":          environments.EnvR2RNoBrowser,
		"EnvR2RNoUpdateCheck":      environments.EnvR2RNoUpdateCheck,
		"EnvR2RSkipPinWarning":     environments.EnvR2RSkipPinWarning,
		"EnvR2RFixedRedirect":      environments.EnvR2RFixedRedirect,
		"EnvR2RDebug":              environments.EnvR2RDebug,
		"EnvEACDebug":              environments.EnvEACDebug,
		"EnvR2RLogLevel":           environments.EnvR2RLogLevel,
		"EnvR2RVerboseLog":         environments.EnvR2RVerboseLog,
		"EnvR2RTesting":            environments.EnvR2RTesting,
		"EnvR2RCheckTags":          environments.EnvR2RCheckTags,
		"EnvR2ROriginalArgs":       environments.EnvR2ROriginalArgs,
		"EnvR2RFilteredArgs":       environments.EnvR2RFilteredArgs,
		"EnvEACUseDirectBinary":       environments.EnvEACUseDirectBinary,
		"EnvCommandsPath":             environments.EnvCommandsPath,
		"EnvGOOS":                     environments.EnvGOOS,
		"EnvEACPortRangeStart":        environments.EnvEACPortRangeStart,
		"EnvEACPortRangeEnd":          environments.EnvEACPortRangeEnd,
		"EnvGitAuthorName":            environments.EnvGitAuthorName,
		"EnvGitAuthorEmail":           environments.EnvGitAuthorEmail,
		"EnvR2RMockAI":                environments.EnvR2RMockAI,
		"EnvR2RMockSecurityTools":     environments.EnvR2RMockSecurityTools,
		"EnvR2RMockGitHub":            environments.EnvR2RMockGitHub,
		"EnvR2RMockGitHubNoWorkflows": environments.EnvR2RMockGitHubNoWorkflows,
		"EnvR2RTestMock":              environments.EnvR2RTestMock,
		"EnvDebugToolResolve":         environments.EnvDebugToolResolve,
		"EnvDebugChangeDetect":        environments.EnvDebugChangeDetect,
		"EnvDebugCacheCmd":            environments.EnvDebugCacheCmd,
		"EnvR2RMockCIStatus":          environments.EnvR2RMockCIStatus,
		"EnvR2RMockHeadSHA":           environments.EnvR2RMockHeadSHA,
		"EnvR2RMockChangedFiles":      environments.EnvR2RMockChangedFiles,
		"EnvR2RTestAIResponse":        environments.EnvR2RTestAIResponse,
		"EnvR2RMockTimeout":           environments.EnvR2RMockTimeout,
		"EnvR2RMockInvalidRef":        environments.EnvR2RMockInvalidRef,
		"EnvR2RMockFailingWorkflow":   environments.EnvR2RMockFailingWorkflow,
		"EnvSkipDockerTests":          environments.EnvSkipDockerTests,
		"EnvFilesByModule":            environments.EnvFilesByModule,
		"EnvModuleStatus":             environments.EnvModuleStatus,
		"EnvCI":                       environments.EnvCI,
		"EnvGitHubActions":            environments.EnvGitHubActions,
		"EnvGitLabCI":                 environments.EnvGitLabCI,
		"EnvGitHubSHA":                environments.EnvGitHubSHA,
		"EnvGitHubToken":              environments.EnvGitHubToken,
		"EnvGitHubUsername":           environments.EnvGitHubUsername,
		"EnvGitHubActor":              environments.EnvGitHubActor,
		"EnvGitHubRunID":              environments.EnvGitHubRunID,
		"EnvGitHubRepository":         environments.EnvGitHubRepository,
		"EnvGitHubEnv":                environments.EnvGitHubEnv,
		"EnvGHRepo":                   environments.EnvGHRepo,
		"EnvR2RMockAIDir":             environments.EnvR2RMockAIDir,
		"EnvR2RMockSecurity":          environments.EnvR2RMockSecurity,
		"EnvR2RMockStructurizr":       environments.EnvR2RMockStructurizr,
		"EnvR2RMockDocker":            environments.EnvR2RMockDocker,
		"EnvR2RMockGitHubCLI":         environments.EnvR2RMockGitHubCLI,
		"EnvR2RMockNoWorkflows":       environments.EnvR2RMockNoWorkflows,
	}

	// Check for duplicate values
	seen := make(map[string]string)
	for name, value := range constants {
		if existingName, exists := seen[value]; exists {
			t.Errorf("duplicate constant value %q found in %s and %s",
				value, existingName, name)
		}
		seen[value] = name
	}
}

// TestConstantNaming verifies that all constant values follow the expected format.
// All values should be non-empty and contain only uppercase letters, numbers, and underscores.
func TestConstantNaming(t *testing.T) {
	constants := map[string]string{
		"EnvR2RTestLoggingActive":     environments.EnvR2RTestLoggingActive,
		"EnvR2RTestRunID":             environments.EnvR2RTestRunID,
		"EnvR2RTestScope":             environments.EnvR2RTestScope,
		"EnvGodogFormat":              environments.EnvGodogFormat,
		"EnvGodogSuiteTags":           environments.EnvGodogSuiteTags,
		"EnvGodogOutputDir":           environments.EnvGodogOutputDir,
		"EnvGodogPaths":               environments.EnvGodogPaths,
		"EnvGodogReportName":          environments.EnvGodogReportName,
		"EnvGodogDebugInit":           environments.EnvGodogDebugInit,
		"EnvR2RPWD":                   environments.EnvR2RPWD,
		"EnvR2RRepoRoot":              environments.EnvR2RRepoRoot,
		"EnvR2RContainerRoot":         environments.EnvR2RContainerRoot,
		"EnvR2RDockerMode":            environments.EnvR2RDockerMode,
		"EnvR2RHostRepoRoot":          environments.EnvR2RHostRepoRoot,
		"EnvR2RContainerRepoRoot":     environments.EnvR2RContainerRepoRoot,
		"EnvR2RDockerHost":            environments.EnvR2RDockerHost,
		"EnvR2RHostGOOS":              environments.EnvR2RHostGOOS,
		"EnvR2RHostGOARCH":            environments.EnvR2RHostGOARCH,
		"EnvR2RTerminalDetection":     environments.EnvR2RTerminalDetection,
		"EnvR2RConfig":                environments.EnvR2RConfig,
		"EnvR2RConfigPath":            environments.EnvR2RConfigPath,
		"EnvR2RContext":               environments.EnvR2RContext,
		"EnvR2RNoBrowser":             environments.EnvR2RNoBrowser,
		"EnvR2RNoUpdateCheck":         environments.EnvR2RNoUpdateCheck,
		"EnvR2RSkipPinWarning":        environments.EnvR2RSkipPinWarning,
		"EnvR2RFixedRedirect":         environments.EnvR2RFixedRedirect,
		"EnvR2RDebug":                 environments.EnvR2RDebug,
		"EnvEACDebug":                 environments.EnvEACDebug,
		"EnvR2RLogLevel":              environments.EnvR2RLogLevel,
		"EnvR2RVerboseLog":            environments.EnvR2RVerboseLog,
		"EnvR2RTesting":               environments.EnvR2RTesting,
		"EnvR2RCheckTags":             environments.EnvR2RCheckTags,
		"EnvR2ROriginalArgs":          environments.EnvR2ROriginalArgs,
		"EnvR2RFilteredArgs":          environments.EnvR2RFilteredArgs,
		"EnvEACUseDirectBinary":       environments.EnvEACUseDirectBinary,
		"EnvCommandsPath":             environments.EnvCommandsPath,
		"EnvGOOS":                     environments.EnvGOOS,
		"EnvEACPortRangeStart":        environments.EnvEACPortRangeStart,
		"EnvEACPortRangeEnd":          environments.EnvEACPortRangeEnd,
		"EnvGitAuthorName":            environments.EnvGitAuthorName,
		"EnvGitAuthorEmail":           environments.EnvGitAuthorEmail,
		"EnvR2RMockAI":                environments.EnvR2RMockAI,
		"EnvR2RMockSecurityTools":     environments.EnvR2RMockSecurityTools,
		"EnvR2RMockGitHub":            environments.EnvR2RMockGitHub,
		"EnvR2RMockGitHubNoWorkflows": environments.EnvR2RMockGitHubNoWorkflows,
		"EnvR2RTestMock":              environments.EnvR2RTestMock,
		"EnvDebugToolResolve":         environments.EnvDebugToolResolve,
		"EnvDebugChangeDetect":        environments.EnvDebugChangeDetect,
		"EnvDebugCacheCmd":            environments.EnvDebugCacheCmd,
		"EnvR2RMockCIStatus":          environments.EnvR2RMockCIStatus,
		"EnvR2RMockHeadSHA":           environments.EnvR2RMockHeadSHA,
		"EnvR2RMockChangedFiles":      environments.EnvR2RMockChangedFiles,
		"EnvR2RTestAIResponse":        environments.EnvR2RTestAIResponse,
		"EnvR2RMockTimeout":           environments.EnvR2RMockTimeout,
		"EnvR2RMockInvalidRef":        environments.EnvR2RMockInvalidRef,
		"EnvR2RMockFailingWorkflow":   environments.EnvR2RMockFailingWorkflow,
		"EnvSkipDockerTests":          environments.EnvSkipDockerTests,
		"EnvFilesByModule":            environments.EnvFilesByModule,
		"EnvModuleStatus":             environments.EnvModuleStatus,
		"EnvCI":                       environments.EnvCI,
		"EnvGitHubActions":            environments.EnvGitHubActions,
		"EnvGitLabCI":                 environments.EnvGitLabCI,
		"EnvGitHubSHA":                environments.EnvGitHubSHA,
		"EnvGitHubToken":              environments.EnvGitHubToken,
		"EnvGitHubUsername":           environments.EnvGitHubUsername,
		"EnvGitHubActor":              environments.EnvGitHubActor,
		"EnvGitHubRunID":              environments.EnvGitHubRunID,
		"EnvGitHubRepository":         environments.EnvGitHubRepository,
		"EnvGitHubEnv":                environments.EnvGitHubEnv,
		"EnvGHRepo":                   environments.EnvGHRepo,
		"EnvR2RMockAIDir":             environments.EnvR2RMockAIDir,
		"EnvR2RMockSecurity":          environments.EnvR2RMockSecurity,
		"EnvR2RMockStructurizr":       environments.EnvR2RMockStructurizr,
		"EnvR2RMockDocker":            environments.EnvR2RMockDocker,
		"EnvR2RMockGitHubCLI":         environments.EnvR2RMockGitHubCLI,
		"EnvR2RMockNoWorkflows":       environments.EnvR2RMockNoWorkflows,
	}

	for name, value := range constants {
		t.Run(name, func(t *testing.T) {
			assert.NotEmpty(t, value, "constant %s should not be empty", name)

			for _, ch := range value {
				valid := (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
				assert.True(t, valid,
					"constant %s contains invalid character %q in value %q", name, ch, value)
			}
		})
	}
}
