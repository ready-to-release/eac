package ci

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractModuleFromWorkflowName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid CI prefix", "CI: core", "core"},
		{"valid CI prefix with hyphen", "CI: mkdocs-render-oci", "mkdocs-render-oci"},
		{"no prefix", "core", ""},
		{"wrong prefix", "Build: core", ""},
		{"empty string", "", ""},
		{"just prefix", "CI: ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractModuleFromWorkflowName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
