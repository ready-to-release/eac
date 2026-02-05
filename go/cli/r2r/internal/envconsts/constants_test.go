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
		{name: "R2R_CONFIG", constant: EnvR2RConfig, expected: "R2R_CONFIG"},
		{name: "R2R_CONFIG_PATH", constant: EnvR2RConfigPath, expected: "R2R_CONFIG_PATH"},
		{name: "R2R_FIXED_REDIRECT", constant: EnvR2RFixedRedirect, expected: "R2R_FIXED_REDIRECT"},
		{name: "R2R_SKIP_PIN_WARNING", constant: EnvR2RSkipPinWarning, expected: "R2R_SKIP_PIN_WARNING"},
		{name: "R2R_NO_UPDATE_CHECK", constant: EnvR2RNoUpdateCheck, expected: "R2R_NO_UPDATE_CHECK"},
		{name: "R2R_DEBUG", constant: EnvR2RDebug, expected: "R2R_DEBUG"},
		{name: "R2R_LOG_LEVEL", constant: EnvR2RLogLevel, expected: "R2R_LOG_LEVEL"},
		{name: "R2R_VERBOSE_LOG", constant: EnvR2RVerboseLog, expected: "R2R_VERBOSE_LOG"},
		{name: "R2R_TESTING", constant: EnvR2RTesting, expected: "R2R_TESTING"},
		{name: "R2R_CHECK_TAGS", constant: EnvR2RCheckTags, expected: "R2R_CHECK_TAGS"},
		{name: "R2R_ORIGINAL_ARGS", constant: EnvR2ROriginalArgs, expected: "R2R_ORIGINAL_ARGS"},
		{name: "R2R_FILTERED_ARGS", constant: EnvR2RFilteredArgs, expected: "R2R_FILTERED_ARGS"},
		{name: "R2R_HOST_REPOROOT", constant: EnvR2RHostRepoRoot, expected: "R2R_HOST_REPOROOT"},
		{name: "R2R_CONTAINER_REPOROOT", constant: EnvR2RContainerRepoRoot, expected: "R2R_CONTAINER_REPOROOT"},
		{name: "R2R_DOCKER_MODE", constant: EnvR2RDockerMode, expected: "R2R_DOCKER_MODE"},
		{name: "R2R_DOCKER_HOST", constant: EnvR2RDockerHost, expected: "R2R_DOCKER_HOST"},
		{name: "R2R_HOST_GOOS", constant: EnvR2RHostGOOS, expected: "R2R_HOST_GOOS"},
		{name: "R2R_HOST_GOARCH", constant: EnvR2RHostGOARCH, expected: "R2R_HOST_GOARCH"},
		{name: "R2R_TERMINAL_DETECTION", constant: EnvR2RTerminalDetection, expected: "R2R_TERMINAL_DETECTION"},
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
		"EnvR2RConfig":            EnvR2RConfig,
		"EnvR2RConfigPath":        EnvR2RConfigPath,
		"EnvR2RFixedRedirect":     EnvR2RFixedRedirect,
		"EnvR2RSkipPinWarning":    EnvR2RSkipPinWarning,
		"EnvR2RNoUpdateCheck":     EnvR2RNoUpdateCheck,
		"EnvR2RDebug":             EnvR2RDebug,
		"EnvR2RLogLevel":          EnvR2RLogLevel,
		"EnvR2RVerboseLog":        EnvR2RVerboseLog,
		"EnvR2RTesting":           EnvR2RTesting,
		"EnvR2RCheckTags":         EnvR2RCheckTags,
		"EnvR2ROriginalArgs":      EnvR2ROriginalArgs,
		"EnvR2RFilteredArgs":      EnvR2RFilteredArgs,
		"EnvR2RHostRepoRoot":      EnvR2RHostRepoRoot,
		"EnvR2RContainerRepoRoot": EnvR2RContainerRepoRoot,
		"EnvR2RDockerMode":        EnvR2RDockerMode,
		"EnvR2RDockerHost":        EnvR2RDockerHost,
		"EnvR2RHostGOOS":          EnvR2RHostGOOS,
		"EnvR2RHostGOARCH":        EnvR2RHostGOARCH,
		"EnvR2RTerminalDetection": EnvR2RTerminalDetection,
	}

	seen := make(map[string]string)
	for name, value := range constants {
		if existingName, exists := seen[value]; exists {
			t.Errorf("duplicate constant value %q: both %s and %s", value, existingName, name)
		}
		seen[value] = name
	}
}

// TestConstantNamingConvention verifies all constant values follow the R2R_* naming convention.
func TestConstantNamingConvention(t *testing.T) {
	constants := []string{
		EnvR2RConfig,
		EnvR2RConfigPath,
		EnvR2RFixedRedirect,
		EnvR2RSkipPinWarning,
		EnvR2RNoUpdateCheck,
		EnvR2RDebug,
		EnvR2RLogLevel,
		EnvR2RVerboseLog,
		EnvR2RTesting,
		EnvR2RCheckTags,
		EnvR2ROriginalArgs,
		EnvR2RFilteredArgs,
		EnvR2RHostRepoRoot,
		EnvR2RContainerRepoRoot,
		EnvR2RDockerMode,
		EnvR2RDockerHost,
		EnvR2RHostGOOS,
		EnvR2RHostGOARCH,
		EnvR2RTerminalDetection,
	}

	for _, value := range constants {
		assert.Contains(t, value, "R2R_",
			"constant value %q should contain 'R2R_' prefix", value)
	}
}
