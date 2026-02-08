package cucumber

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
)

func TestCucumberTagTranslator_TranslateTagFilter(t *testing.T) {
	translator := &CucumberTagTranslator{}

	tests := []struct {
		name     string
		filter   core.TagFilter
		expected string
	}{
		{
			name:     "empty filter",
			filter:   core.TagFilter{},
			expected: "",
		},
		{
			name: "single require tag",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L2"}},
				},
			},
			expected: "@L2",
		},
		{
			name: "multiple require tags are ANDed",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L2", "@deps:docker"}},
				},
			},
			expected: "@L2 and @deps:docker",
		},
		{
			name: "single exclude tag uses 'not'",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip"}},
				},
			},
			expected: "not @skip:wip",
		},
		{
			name: "multiple exclude tags",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip", "@pending"}},
				},
			},
			expected: "not @skip:wip and not @pending",
		},
		{
			name: "any-of tags use 'or' in parentheses",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{AnyOfTags: []string{"@L0", "@L1"}},
				},
			},
			expected: "(@L0 or @L1)",
		},
		{
			name: "combined require, exclude, and any-of",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{
						RequireTags: []string{"@deps:docker"},
						ExcludeTags: []string{"@skip:wip"},
						AnyOfTags:   []string{"@L0", "@L1"},
					},
				},
			},
			expected: "@deps:docker and not @skip:wip and (@L0 or @L1)",
		},
		{
			name: "multiple selectors are ORed with parentheses",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L0"}},
					{RequireTags: []string{"@L1"}},
				},
			},
			expected: "(@L0) or (@L1)",
		},
		{
			name: "empty selector is skipped",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{},
					{RequireTags: []string{"@L2"}},
				},
			},
			expected: "@L2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := translator.TranslateTagFilter(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Verify CucumberTagTranslator implements TagFilterTranslator.
var _ core.TagFilterTranslator = (*CucumberTagTranslator)(nil)
