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
			name:     "CLIE_TEST_LOGGING_ACTIVE",
			constant: environments.EnvCLIETestLoggingActive,
			expected: "CLIE_TEST_LOGGING_ACTIVE",
		},
		{
			name:     "CLIE_TEST_RUN_ID",
			constant: environments.EnvCLIETestRunID,
			expected: "CLIE_TEST_RUN_ID",
		},
		{
			name:     "CLIE_TEST_SCOPE",
			constant: environments.EnvCLIETestScope,
			expected: "CLIE_TEST_SCOPE",
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
			name:     "CLIE_PWD",
			constant: environments.EnvCLIEPWD,
			expected: "CLIE_PWD",
		},
		{
			name:     "CLIE_REPO_ROOT",
			constant: environments.EnvCLIERepoRoot,
			expected: "CLIE_REPO_ROOT",
		},
		{
			name:     "CLIE_CONTAINER_ROOT",
			constant: environments.EnvCLIEContainerRoot,
			expected: "CLIE_CONTAINER_ROOT",
		},
		{
			name:     "CLIE_DOCKER_MODE",
			constant: environments.EnvCLIEDockerMode,
			expected: "CLIE_DOCKER_MODE",
		},

		// Docker and container runtime
		{
			name:     "CLIE_HOST_REPOROOT",
			constant: environments.EnvCLIEHostRepoRoot,
			expected: "CLIE_HOST_REPOROOT",
		},
		{
			name:     "CLIE_CONTAINER_REPOROOT",
			constant: environments.EnvCLIEContainerRepoRoot,
			expected: "CLIE_CONTAINER_REPOROOT",
		},
		{
			name:     "CLIE_DOCKER_HOST",
			constant: environments.EnvCLIEDockerHost,
			expected: "CLIE_DOCKER_HOST",
		},
		{
			name:     "CLIE_HOST_GOOS",
			constant: environments.EnvCLIEHostGOOS,
			expected: "CLIE_HOST_GOOS",
		},
		{
			name:     "CLIE_HOST_GOARCH",
			constant: environments.EnvCLIEHostGOARCH,
			expected: "CLIE_HOST_GOARCH",
		},
		{
			name:     "CLIE_TERMINAL_DETECTION",
			constant: environments.EnvCLIETerminalDetection,
			expected: "CLIE_TERMINAL_DETECTION",
		},

		// Application configuration
		{
			name:     "CLIE_CONFIG",
			constant: environments.EnvCLIEConfig,
			expected: "CLIE_CONFIG",
		},
		{
			name:     "CLIE_CONFIG_PATH",
			constant: environments.EnvCLIEConfigPath,
			expected: "CLIE_CONFIG_PATH",
		},
		{
			name:     "CLIE_CONTEXT",
			constant: environments.EnvCLIEContext,
			expected: "CLIE_CONTEXT",
		},
		{
			name:     "CLIE_NO_BROWSER",
			constant: environments.EnvCLIENoBrowser,
			expected: "CLIE_NO_BROWSER",
		},
		{
			name:     "CLIE_NO_UPDATE_CHECK",
			constant: environments.EnvCLIENoUpdateCheck,
			expected: "CLIE_NO_UPDATE_CHECK",
		},
		{
			name:     "CLIE_SKIP_PIN_WARNING",
			constant: environments.EnvCLIESkipPinWarning,
			expected: "CLIE_SKIP_PIN_WARNING",
		},
		{
			name:     "CLIE_FIXED_REDIRECT",
			constant: environments.EnvCLIEFixedRedirect,
			expected: "CLIE_FIXED_REDIRECT",
		},

		// Debug and logging
		{
			name:     "CLIE_DEBUG",
			constant: environments.EnvCLIEDebug,
			expected: "CLIE_DEBUG",
		},
		{
			name:     "EAC_DEBUG",
			constant: environments.EnvEACDebug,
			expected: "EAC_DEBUG",
		},
		{
			name:     "CLIE_LOG_LEVEL",
			constant: environments.EnvCLIELogLevel,
			expected: "CLIE_LOG_LEVEL",
		},
		{
			name:     "CLIE_VERBOSE_LOG",
			constant: environments.EnvCLIEVerboseLog,
			expected: "CLIE_VERBOSE_LOG",
		},

		// Testing
		{
			name:     "CLIE_TESTING",
			constant: environments.EnvCLIETesting,
			expected: "CLIE_TESTING",
		},
		{
			name:     "CLIE_CHECK_TAGS",
			constant: environments.EnvCLIECheckTags,
			expected: "CLIE_CHECK_TAGS",
		},
		{
			name:     "CLIE_ORIGINAL_ARGS",
			constant: environments.EnvCLIEOriginalArgs,
			expected: "CLIE_ORIGINAL_ARGS",
		},
		{
			name:     "CLIE_FILTERED_ARGS",
			constant: environments.EnvCLIEFilteredArgs,
			expected: "CLIE_FILTERED_ARGS",
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
		{name: "CLIE_MOCK_AI", constant: environments.EnvCLIEMockAI, expected: "CLIE_MOCK_AI"},
		{name: "CLIE_MOCK_SECURITY_TOOLS", constant: environments.EnvCLIEMockSecurityTools, expected: "CLIE_MOCK_SECURITY_TOOLS"},
		{name: "CLIE_MOCK_GITHUB", constant: environments.EnvCLIEMockGitHub, expected: "CLIE_MOCK_GITHUB"},
		{name: "CLIE_MOCK_GITHUB_NO_WORKFLOWS", constant: environments.EnvCLIEMockGitHubNoWorkflows, expected: "CLIE_MOCK_GITHUB_NO_WORKFLOWS"},
		{name: "__CLIE_TEST_MOCK", constant: environments.EnvCLIETestMock, expected: "__CLIE_TEST_MOCK"},
		{name: "DEBUG_TOOL_RESOLVE", constant: environments.EnvDebugToolResolve, expected: "DEBUG_TOOL_RESOLVE"},
		{name: "DEBUG_CHANGEDETECT", constant: environments.EnvDebugChangeDetect, expected: "DEBUG_CHANGEDETECT"},
		{name: "DEBUG_CACHE_CMD", constant: environments.EnvDebugCacheCmd, expected: "DEBUG_CACHE_CMD"},
		{name: "CLIE_MOCK_CI_STATUS", constant: environments.EnvCLIEMockCIStatus, expected: "CLIE_MOCK_CI_STATUS"},
		{name: "CLIE_MOCK_HEAD_SHA", constant: environments.EnvCLIEMockHeadSHA, expected: "CLIE_MOCK_HEAD_SHA"},
		{name: "CLIE_MOCK_CHANGED_FILES", constant: environments.EnvCLIEMockChangedFiles, expected: "CLIE_MOCK_CHANGED_FILES"},
		{name: "CLIE_TEST_AI_RESPONSE", constant: environments.EnvCLIETestAIResponse, expected: "CLIE_TEST_AI_RESPONSE"},
		{name: "CLIE_MOCK_TIMEOUT", constant: environments.EnvCLIEMockTimeout, expected: "CLIE_MOCK_TIMEOUT"},
		{name: "CLIE_MOCK_INVALID_REF", constant: environments.EnvCLIEMockInvalidRef, expected: "CLIE_MOCK_INVALID_REF"},
		{name: "CLIE_MOCK_FAILING_WORKFLOW", constant: environments.EnvCLIEMockFailingWorkflow, expected: "CLIE_MOCK_FAILING_WORKFLOW"},
		{name: "EAC_MOCK_CONTAINER_REGISTRY", constant: environments.EnvEACMockContainerRegistry, expected: "EAC_MOCK_CONTAINER_REGISTRY"},
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
			name:     "CLIE_MOCK_AI_DIR",
			constant: environments.EnvCLIEMockAIDir,
			expected: "CLIE_MOCK_AI_DIR",
		},
		{
			name:     "CLIE_MOCK_SECURITY",
			constant: environments.EnvCLIEMockSecurity,
			expected: "CLIE_MOCK_SECURITY",
		},
		{
			name:     "CLIE_MOCK_DOCKER",
			constant: environments.EnvCLIEMockDocker,
			expected: "CLIE_MOCK_DOCKER",
		},
		{
			name:     "CLIE_MOCK_GITHUB_CLI",
			constant: environments.EnvCLIEMockGitHubCLI,
			expected: "CLIE_MOCK_GITHUB_CLI",
		},
		{
			name:     "CLIE_MOCK_NO_WORKFLOWS",
			constant: environments.EnvCLIEMockNoWorkflows,
			expected: "CLIE_MOCK_NO_WORKFLOWS",
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
		"EnvCLIETestLoggingActive":  environments.EnvCLIETestLoggingActive,
		"EnvCLIETestRunID":          environments.EnvCLIETestRunID,
		"EnvCLIETestScope":          environments.EnvCLIETestScope,
		"EnvGodogFormat":           environments.EnvGodogFormat,
		"EnvGodogSuiteTags":        environments.EnvGodogSuiteTags,
		"EnvGodogOutputDir":        environments.EnvGodogOutputDir,
		"EnvGodogPaths":            environments.EnvGodogPaths,
		"EnvGodogReportName":       environments.EnvGodogReportName,
		"EnvGodogDebugInit":        environments.EnvGodogDebugInit,
		"EnvCLIEPWD":                environments.EnvCLIEPWD,
		"EnvCLIERepoRoot":           environments.EnvCLIERepoRoot,
		"EnvCLIEContainerRoot":      environments.EnvCLIEContainerRoot,
		"EnvCLIEDockerMode":         environments.EnvCLIEDockerMode,
		"EnvCLIEHostRepoRoot":       environments.EnvCLIEHostRepoRoot,
		"EnvCLIEContainerRepoRoot":  environments.EnvCLIEContainerRepoRoot,
		"EnvCLIEDockerHost":         environments.EnvCLIEDockerHost,
		"EnvCLIEHostGOOS":           environments.EnvCLIEHostGOOS,
		"EnvCLIEHostGOARCH":         environments.EnvCLIEHostGOARCH,
		"EnvCLIEConfig":             environments.EnvCLIEConfig,
		"EnvCLIEConfigPath":         environments.EnvCLIEConfigPath,
		"EnvCLIEContext":            environments.EnvCLIEContext,
		"EnvCLIENoBrowser":          environments.EnvCLIENoBrowser,
		"EnvCLIENoUpdateCheck":      environments.EnvCLIENoUpdateCheck,
		"EnvCLIESkipPinWarning":     environments.EnvCLIESkipPinWarning,
		"EnvCLIEFixedRedirect":      environments.EnvCLIEFixedRedirect,
		"EnvCLIEDebug":              environments.EnvCLIEDebug,
		"EnvEACDebug":              environments.EnvEACDebug,
		"EnvCLIELogLevel":           environments.EnvCLIELogLevel,
		"EnvCLIEVerboseLog":         environments.EnvCLIEVerboseLog,
		"EnvCLIETesting":            environments.EnvCLIETesting,
		"EnvCLIECheckTags":          environments.EnvCLIECheckTags,
		"EnvCLIEOriginalArgs":       environments.EnvCLIEOriginalArgs,
		"EnvCLIEFilteredArgs":       environments.EnvCLIEFilteredArgs,
		"EnvEACUseDirectBinary":       environments.EnvEACUseDirectBinary,
		"EnvCommandsPath":             environments.EnvCommandsPath,
		"EnvGOOS":                     environments.EnvGOOS,
		"EnvEACPortRangeStart":        environments.EnvEACPortRangeStart,
		"EnvEACPortRangeEnd":          environments.EnvEACPortRangeEnd,
		"EnvGitAuthorName":            environments.EnvGitAuthorName,
		"EnvGitAuthorEmail":           environments.EnvGitAuthorEmail,
		"EnvCLIEMockAI":                environments.EnvCLIEMockAI,
		"EnvCLIEMockSecurityTools":     environments.EnvCLIEMockSecurityTools,
		"EnvCLIEMockGitHub":            environments.EnvCLIEMockGitHub,
		"EnvCLIEMockGitHubNoWorkflows": environments.EnvCLIEMockGitHubNoWorkflows,
		"EnvCLIETestMock":              environments.EnvCLIETestMock,
		"EnvDebugToolResolve":         environments.EnvDebugToolResolve,
		"EnvDebugChangeDetect":        environments.EnvDebugChangeDetect,
		"EnvDebugCacheCmd":            environments.EnvDebugCacheCmd,
		"EnvCLIEMockCIStatus":          environments.EnvCLIEMockCIStatus,
		"EnvCLIEMockHeadSHA":           environments.EnvCLIEMockHeadSHA,
		"EnvCLIEMockChangedFiles":      environments.EnvCLIEMockChangedFiles,
		"EnvCLIETestAIResponse":        environments.EnvCLIETestAIResponse,
		"EnvCLIEMockTimeout":           environments.EnvCLIEMockTimeout,
		"EnvCLIEMockInvalidRef":        environments.EnvCLIEMockInvalidRef,
		"EnvCLIEMockFailingWorkflow":   environments.EnvCLIEMockFailingWorkflow,
		"EnvEACMockContainerRegistry": environments.EnvEACMockContainerRegistry,
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
		"EnvCLIEMockAIDir":             environments.EnvCLIEMockAIDir,
		"EnvCLIEMockSecurity":          environments.EnvCLIEMockSecurity,
		"EnvCLIEMockDocker":            environments.EnvCLIEMockDocker,
		"EnvCLIEMockGitHubCLI":         environments.EnvCLIEMockGitHubCLI,
		"EnvCLIEMockNoWorkflows":       environments.EnvCLIEMockNoWorkflows,
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
		"EnvCLIETestLoggingActive":     environments.EnvCLIETestLoggingActive,
		"EnvCLIETestRunID":             environments.EnvCLIETestRunID,
		"EnvCLIETestScope":             environments.EnvCLIETestScope,
		"EnvGodogFormat":              environments.EnvGodogFormat,
		"EnvGodogSuiteTags":           environments.EnvGodogSuiteTags,
		"EnvGodogOutputDir":           environments.EnvGodogOutputDir,
		"EnvGodogPaths":               environments.EnvGodogPaths,
		"EnvGodogReportName":          environments.EnvGodogReportName,
		"EnvGodogDebugInit":           environments.EnvGodogDebugInit,
		"EnvCLIEPWD":                   environments.EnvCLIEPWD,
		"EnvCLIERepoRoot":              environments.EnvCLIERepoRoot,
		"EnvCLIEContainerRoot":         environments.EnvCLIEContainerRoot,
		"EnvCLIEDockerMode":            environments.EnvCLIEDockerMode,
		"EnvCLIEHostRepoRoot":          environments.EnvCLIEHostRepoRoot,
		"EnvCLIEContainerRepoRoot":     environments.EnvCLIEContainerRepoRoot,
		"EnvCLIEDockerHost":            environments.EnvCLIEDockerHost,
		"EnvCLIEHostGOOS":              environments.EnvCLIEHostGOOS,
		"EnvCLIEHostGOARCH":            environments.EnvCLIEHostGOARCH,
		"EnvCLIETerminalDetection":     environments.EnvCLIETerminalDetection,
		"EnvCLIEConfig":                environments.EnvCLIEConfig,
		"EnvCLIEConfigPath":            environments.EnvCLIEConfigPath,
		"EnvCLIEContext":               environments.EnvCLIEContext,
		"EnvCLIENoBrowser":             environments.EnvCLIENoBrowser,
		"EnvCLIENoUpdateCheck":         environments.EnvCLIENoUpdateCheck,
		"EnvCLIESkipPinWarning":        environments.EnvCLIESkipPinWarning,
		"EnvCLIEFixedRedirect":         environments.EnvCLIEFixedRedirect,
		"EnvCLIEDebug":                 environments.EnvCLIEDebug,
		"EnvEACDebug":                 environments.EnvEACDebug,
		"EnvCLIELogLevel":              environments.EnvCLIELogLevel,
		"EnvCLIEVerboseLog":            environments.EnvCLIEVerboseLog,
		"EnvCLIETesting":               environments.EnvCLIETesting,
		"EnvCLIECheckTags":             environments.EnvCLIECheckTags,
		"EnvCLIEOriginalArgs":          environments.EnvCLIEOriginalArgs,
		"EnvCLIEFilteredArgs":          environments.EnvCLIEFilteredArgs,
		"EnvEACUseDirectBinary":       environments.EnvEACUseDirectBinary,
		"EnvCommandsPath":             environments.EnvCommandsPath,
		"EnvGOOS":                     environments.EnvGOOS,
		"EnvEACPortRangeStart":        environments.EnvEACPortRangeStart,
		"EnvEACPortRangeEnd":          environments.EnvEACPortRangeEnd,
		"EnvGitAuthorName":            environments.EnvGitAuthorName,
		"EnvGitAuthorEmail":           environments.EnvGitAuthorEmail,
		"EnvCLIEMockAI":                environments.EnvCLIEMockAI,
		"EnvCLIEMockSecurityTools":     environments.EnvCLIEMockSecurityTools,
		"EnvCLIEMockGitHub":            environments.EnvCLIEMockGitHub,
		"EnvCLIEMockGitHubNoWorkflows": environments.EnvCLIEMockGitHubNoWorkflows,
		"EnvCLIETestMock":              environments.EnvCLIETestMock,
		"EnvDebugToolResolve":         environments.EnvDebugToolResolve,
		"EnvDebugChangeDetect":        environments.EnvDebugChangeDetect,
		"EnvDebugCacheCmd":            environments.EnvDebugCacheCmd,
		"EnvCLIEMockCIStatus":          environments.EnvCLIEMockCIStatus,
		"EnvCLIEMockHeadSHA":           environments.EnvCLIEMockHeadSHA,
		"EnvCLIEMockChangedFiles":      environments.EnvCLIEMockChangedFiles,
		"EnvCLIETestAIResponse":        environments.EnvCLIETestAIResponse,
		"EnvCLIEMockTimeout":           environments.EnvCLIEMockTimeout,
		"EnvCLIEMockInvalidRef":        environments.EnvCLIEMockInvalidRef,
		"EnvCLIEMockFailingWorkflow":   environments.EnvCLIEMockFailingWorkflow,
		"EnvEACMockContainerRegistry": environments.EnvEACMockContainerRegistry,
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
		"EnvCLIEMockAIDir":             environments.EnvCLIEMockAIDir,
		"EnvCLIEMockSecurity":          environments.EnvCLIEMockSecurity,
		"EnvCLIEMockDocker":            environments.EnvCLIEMockDocker,
		"EnvCLIEMockGitHubCLI":         environments.EnvCLIEMockGitHubCLI,
		"EnvCLIEMockNoWorkflows":       environments.EnvCLIEMockNoWorkflows,
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
