package tool

import (
	"testing"
)

func TestGetTestTypeComponentType(t *testing.T) {
	// Save and restore global state using mutex-protected reset
	defer ResetGlobalToolConfigForTesting()

	tests := []struct {
		name     string
		config   *ToolConfig
		testType string
		expected string
	}{
		{
			name:     "gotest defaults to go",
			config:   nil,
			testType: "gotest",
			expected: "go",
		},
		{
			name:     "godog defaults to gherkin",
			config:   nil,
			testType: "godog",
			expected: "gherkin",
		},
		{
			name:     "mocha defaults to typescript",
			config:   nil,
			testType: "mocha",
			expected: "typescript",
		},
		{
			name:     "tscucumber defaults to gherkin",
			config:   nil,
			testType: "tscucumber",
			expected: "gherkin",
		},
		{
			name:     "unknown defaults to go",
			config:   nil,
			testType: "unknown",
			expected: "go",
		},
		{
			name: "config overrides gotest",
			config: &ToolConfig{
				TestTypeMapping: map[string]string{
					"gotest": "golang",
				},
			},
			testType: "gotest",
			expected: "golang",
		},
		{
			name: "config adds custom test type",
			config: &ToolConfig{
				TestTypeMapping: map[string]string{
					"jest": "javascript",
				},
			},
			testType: "jest",
			expected: "javascript",
		},
		{
			name: "config with mapping falls back for unknown",
			config: &ToolConfig{
				TestTypeMapping: map[string]string{
					"gotest": "go",
				},
			},
			testType: "unknown",
			expected: "go", // Falls back to hardcoded default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGlobalToolConfigForTesting(tt.config)

			result := GetTestTypeComponentType(tt.testType)
			if result != tt.expected {
				t.Errorf("GetTestTypeComponentType(%q) = %q, want %q", tt.testType, result, tt.expected)
			}
		})
	}
}
