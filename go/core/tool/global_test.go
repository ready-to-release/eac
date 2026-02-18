package tool

import (
	"testing"
)

func TestGetTestTypeComponentType(t *testing.T) {
	tests := []struct {
		name     string
		config   *ToolConfig
		testType string
		expected string
	}{
		{
			name:     "unknown returns itself (no hardcoded fallback)",
			config:   nil,
			testType: "unknown",
			expected: "unknown",
		},
		{
			name:     "unregistered type returns itself",
			config:   nil,
			testType: "gotest",
			expected: "gotest",
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
			name: "config mapping does not cover unknown",
			config: &ToolConfig{
				TestTypeMapping: map[string]string{
					"gotest": "go",
				},
			},
			testType: "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewToolSystemForTesting()
			if tt.config != nil {
				ts.Config = tt.config
			}

			result := ts.GetTestTypeComponentType(tt.testType)
			if result != tt.expected {
				t.Errorf("GetTestTypeComponentType(%q) = %q, want %q", tt.testType, result, tt.expected)
			}
		})
	}
}
