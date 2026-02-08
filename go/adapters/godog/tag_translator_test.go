package godog

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
	"github.com/stretchr/testify/assert"
)

func TestGodogTagTranslator_TranslateTagFilter(t *testing.T) {
	translator := &GodogTagTranslator{}

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
			expected: "@L2 && @deps:docker",
		},
		{
			name: "single exclude tag",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip"}},
				},
			},
			expected: "~@skip:wip",
		},
		{
			name: "multiple exclude tags",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip", "@pending"}},
				},
			},
			expected: "~@skip:wip && ~@pending",
		},
		{
			name: "any-of tags are ORed with comma",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{AnyOfTags: []string{"@L0", "@L1"}},
				},
			},
			expected: "@L0,@L1",
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
			expected: "@deps:docker && ~@skip:wip && @L0,@L1",
		},
		{
			name: "multiple selectors are ORed with comma",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L0"}},
					{RequireTags: []string{"@L1"}},
				},
			},
			expected: "@L0,@L1",
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

// Verify GodogTagTranslator implements TagFilterTranslator.
var _ core.TagFilterTranslator = (*GodogTagTranslator)(nil)
