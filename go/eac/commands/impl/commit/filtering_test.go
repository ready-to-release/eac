//go:build L1

package commit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStripWrappingCodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "removes wrapping code fences",
			input:    "```\nfeat(cli): add feature\n\nBody content.\n```",
			expected: "feat(cli): add feature\n\nBody content.",
		},
		{
			name:     "removes code fences with language hint",
			input:    "```text\nfeat(cli): add feature\n```",
			expected: "feat(cli): add feature",
		},
		{
			name:     "preserves content without code fences",
			input:    "feat(cli): add feature\n\nBody content.",
			expected: "feat(cli): add feature\n\nBody content.",
		},
		{
			name:     "handles empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "handles single line",
			input:    "feat(cli): add feature",
			expected: "feat(cli): add feature",
		},
		{
			name:     "handles only opening fence",
			input:    "```\nsome content",
			expected: "some content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripWrappingCodeFences(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommitPatternMatcher_IsValidTopLevelLine(t *testing.T) {
	matcher := &commitPatternMatcher{agentType: "top-level"}

	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "valid feat with scope",
			line:     "feat(cli): add feature",
			expected: true,
		},
		{
			name:     "valid fix with scope",
			line:     "fix(core): resolve bug",
			expected: true,
		},
		{
			name:     "valid multi-module",
			line:     "feat(multi-module): update modules",
			expected: true,
		},
		{
			name:     "invalid - no parentheses",
			line:     "feat: add feature",
			expected: false,
		},
		{
			name:     "invalid - plain text",
			line:     "This is not a commit message",
			expected: false,
		},
		{
			name:     "invalid - empty",
			line:     "",
			expected: false,
		},
		{
			name:     "invalid - markdown header",
			line:     "# feat(cli): add feature",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.isValidTopLevelLine(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommitPatternMatcher_IsValidModuleLine(t *testing.T) {
	matcher := &commitPatternMatcher{agentType: "module"}

	tests := []struct {
		name     string
		line     string
		expected bool
	}{
		{
			name:     "valid module name",
			line:     "eac-commands",
			expected: true,
		},
		{
			name:     "valid module with underscore",
			line:     "src_core",
			expected: true,
		},
		{
			name:     "valid module with numbers",
			line:     "module123",
			expected: true,
		},
		{
			name:     "invalid - contains colon",
			line:     "module: text",
			expected: false,
		},
		{
			name:     "invalid - markdown header",
			line:     "# module",
			expected: false,
		},
		{
			name:     "invalid - horizontal rule",
			line:     "---",
			expected: false,
		},
		{
			name:     "invalid - empty",
			line:     "",
			expected: false,
		},
		{
			name:     "invalid - uppercase",
			line:     "ModuleName",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matcher.isValidModuleLine(tt.line)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCommitPatternMatcher_FindFirstValidLine(t *testing.T) {
	tests := []struct {
		name      string
		agentType string
		lines     []string
		expected  int
	}{
		{
			name:      "finds commit line at start",
			agentType: "top-level",
			lines:     []string{"feat(cli): add feature", "body"},
			expected:  0,
		},
		{
			name:      "skips noise before commit line",
			agentType: "top-level",
			lines:     []string{"**Initialized**", "", "feat(cli): add feature"},
			expected:  2,
		},
		{
			name:      "finds module name",
			agentType: "module",
			lines:     []string{"eac-commands", "---", "content"},
			expected:  0,
		},
		{
			name:      "returns -1 when no valid line found",
			agentType: "top-level",
			lines:     []string{"invalid", "also invalid"},
			expected:  -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher := &commitPatternMatcher{agentType: tt.agentType}
			result := matcher.findFirstValidLine(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterWithHardcodedRules(t *testing.T) {
	tests := []struct {
		name     string
		lines    []string
		expected string
	}{
		{
			name:     "removes initialization message",
			lines:    []string{"**Initialized and ready to assist**", "feat(cli): add feature"},
			expected: "feat(cli): add feature",
		},
		{
			name:     "removes horizontal rules before content",
			lines:    []string{"---", "feat(cli): add feature"},
			expected: "feat(cli): add feature",
		},
		{
			name:     "preserves horizontal rules after content",
			lines:    []string{"feat(cli): add feature", "---", "more content"},
			expected: "feat(cli): add feature\n---\nmore content",
		},
		{
			name:     "removes leading empty lines",
			lines:    []string{"", "", "feat(cli): add feature"},
			expected: "feat(cli): add feature",
		},
		{
			name:     "removes multiple noise patterns",
			lines:    []string{"**INITIALIZED**", "Loading project context", "", "feat(cli): add feature"},
			expected: "feat(cli): add feature",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterWithHardcodedRules(tt.lines)
			assert.Equal(t, tt.expected, result)
		})
	}
}
