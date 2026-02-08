package builders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCommandFragment_WithFrontmatter(t *testing.T) {
	cmd := CommandSource{
		Command:     "show modules",
		Target:      "reference/modules.md",
		Frontmatter: map[string]any{"title": "Modules"},
	}
	result := formatCommandFragment(cmd, "module1\nmodule2")
	assert.Contains(t, result, "---")
	assert.Contains(t, result, "title: Modules")
	assert.Contains(t, result, "module1\nmodule2")
}

func TestFormatCommandFragment_NoFrontmatter(t *testing.T) {
	cmd := CommandSource{Command: "show modules"}
	result := formatCommandFragment(cmd, "output here")
	assert.NotContains(t, result, "---")
	assert.Equal(t, "output here\n", result)
}

func TestFormatCommandFragment_TrimsOutput(t *testing.T) {
	cmd := CommandSource{Command: "test"}
	result := formatCommandFragment(cmd, "\n  output  \n\n")
	assert.Equal(t, "output\n", result)
}

func TestResolveBookNameForMarkdownCommands(t *testing.T) {
	tests := []struct {
		component string
		want      string
	}{
		{"", "site"},
		{"site-markdown-commands", "site"},
		{"report-mdcmds", "report"},
		{"tutorials-markdown-commands", "tutorials"},
		{"custom", "custom"},
	}
	for _, tt := range tests {
		t.Run(tt.component, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveBookNameForMarkdownCommands(nil, tt.component))
		})
	}
}

func TestFormatCommandFragment_MultipleFrontmatterKeys(t *testing.T) {
	cmd := CommandSource{
		Command: "show modules",
		Frontmatter: map[string]any{
			"title":       "Modules",
			"description": "All modules",
		},
	}
	result := formatCommandFragment(cmd, "content")
	assert.Contains(t, result, "---\n")
	assert.Contains(t, result, "title: Modules")
	assert.Contains(t, result, "description: All modules")
	assert.Contains(t, result, "content\n")
}
