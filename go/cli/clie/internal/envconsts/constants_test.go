package envconsts

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConstantValues verifies that each constant has the expected string value.
func TestConstantValues(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{name: "CLIE_CONFIG", constant: EnvCLIEConfig, expected: "CLIE_CONFIG"},
		{name: "CLIE_CONFIG_PATH", constant: EnvCLIEConfigPath, expected: "CLIE_CONFIG_PATH"},
		{name: "CLIE_FIXED_REDIRECT", constant: EnvCLIEFixedRedirect, expected: "CLIE_FIXED_REDIRECT"},
		{name: "CLIE_SKIP_PIN_WARNING", constant: EnvCLIESkipPinWarning, expected: "CLIE_SKIP_PIN_WARNING"},
		{name: "CLIE_NO_UPDATE_CHECK", constant: EnvCLIENoUpdateCheck, expected: "CLIE_NO_UPDATE_CHECK"},
		{name: "CLIE_DEBUG", constant: EnvCLIEDebug, expected: "CLIE_DEBUG"},
		{name: "CLIE_LOG_LEVEL", constant: EnvCLIELogLevel, expected: "CLIE_LOG_LEVEL"},
		{name: "CLIE_VERBOSE_LOG", constant: EnvCLIEVerboseLog, expected: "CLIE_VERBOSE_LOG"},
		{name: "CLIE_TESTING", constant: EnvCLIETesting, expected: "CLIE_TESTING"},
		{name: "CLIE_CHECK_TAGS", constant: EnvCLIECheckTags, expected: "CLIE_CHECK_TAGS"},
		{name: "CLIE_ORIGINAL_ARGS", constant: EnvCLIEOriginalArgs, expected: "CLIE_ORIGINAL_ARGS"},
		{name: "CLIE_FILTERED_ARGS", constant: EnvCLIEFilteredArgs, expected: "CLIE_FILTERED_ARGS"},
		{name: "CLIE_HOST_REPOROOT", constant: EnvCLIEHostRepoRoot, expected: "CLIE_HOST_REPOROOT"},
		{name: "CLIE_CONTAINER_REPOROOT", constant: EnvCLIEContainerRepoRoot, expected: "CLIE_CONTAINER_REPOROOT"},
		{name: "CLIE_DOCKER_MODE", constant: EnvCLIEDockerMode, expected: "CLIE_DOCKER_MODE"},
		{name: "CLIE_DOCKER_HOST", constant: EnvCLIEDockerHost, expected: "CLIE_DOCKER_HOST"},
		{name: "CLIE_HOST_GOOS", constant: EnvCLIEHostGOOS, expected: "CLIE_HOST_GOOS"},
		{name: "CLIE_HOST_GOARCH", constant: EnvCLIEHostGOARCH, expected: "CLIE_HOST_GOARCH"},
		{name: "CLIE_TERMINAL_DETECTION", constant: EnvCLIETerminalDetection, expected: "CLIE_TERMINAL_DETECTION"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.constant,
				"constant %s should equal %q", tt.name, tt.expected)
		})
	}
}

// TestConstantUniqueness ensures no two constants share the same value.
func TestConstantUniqueness(t *testing.T) {
	constants := map[string]string{
		"EnvCLIEConfig":            EnvCLIEConfig,
		"EnvCLIEConfigPath":        EnvCLIEConfigPath,
		"EnvCLIEFixedRedirect":     EnvCLIEFixedRedirect,
		"EnvCLIESkipPinWarning":    EnvCLIESkipPinWarning,
		"EnvCLIENoUpdateCheck":     EnvCLIENoUpdateCheck,
		"EnvCLIEDebug":             EnvCLIEDebug,
		"EnvCLIELogLevel":          EnvCLIELogLevel,
		"EnvCLIEVerboseLog":        EnvCLIEVerboseLog,
		"EnvCLIETesting":           EnvCLIETesting,
		"EnvCLIECheckTags":         EnvCLIECheckTags,
		"EnvCLIEOriginalArgs":      EnvCLIEOriginalArgs,
		"EnvCLIEFilteredArgs":      EnvCLIEFilteredArgs,
		"EnvCLIEHostRepoRoot":      EnvCLIEHostRepoRoot,
		"EnvCLIEContainerRepoRoot": EnvCLIEContainerRepoRoot,
		"EnvCLIEDockerMode":        EnvCLIEDockerMode,
		"EnvCLIEDockerHost":        EnvCLIEDockerHost,
		"EnvCLIEHostGOOS":          EnvCLIEHostGOOS,
		"EnvCLIEHostGOARCH":        EnvCLIEHostGOARCH,
		"EnvCLIETerminalDetection": EnvCLIETerminalDetection,
	}

	seen := make(map[string]string)
	for name, value := range constants {
		if existingName, exists := seen[value]; exists {
			t.Errorf("duplicate constant value %q: both %s and %s", value, existingName, name)
		}
		seen[value] = name
	}
}

// TestConstantNamingConvention verifies all constant values follow the CLIE_* naming convention.
func TestConstantNamingConvention(t *testing.T) {
	constants := []string{
		EnvCLIEConfig,
		EnvCLIEConfigPath,
		EnvCLIEFixedRedirect,
		EnvCLIESkipPinWarning,
		EnvCLIENoUpdateCheck,
		EnvCLIEDebug,
		EnvCLIELogLevel,
		EnvCLIEVerboseLog,
		EnvCLIETesting,
		EnvCLIECheckTags,
		EnvCLIEOriginalArgs,
		EnvCLIEFilteredArgs,
		EnvCLIEHostRepoRoot,
		EnvCLIEContainerRepoRoot,
		EnvCLIEDockerMode,
		EnvCLIEDockerHost,
		EnvCLIEHostGOOS,
		EnvCLIEHostGOARCH,
		EnvCLIETerminalDetection,
	}

	for _, value := range constants {
		assert.Contains(t, value, "CLIE_",
			"constant value %q should contain 'CLIE_' prefix", value)
	}
}
