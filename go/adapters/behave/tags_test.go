package behave

import (
	"testing"

	core "github.com/ready-to-release/eac/contracts/core/0.1.0"
)

func TestBehaveTagTranslator_TranslateTagFilter(t *testing.T) {
	translator := &BehaveTagTranslator{}

	tests := []struct {
		name     string
		filter   core.TagFilter
		wantArgs []string
	}{
		{
			name: "single require tag produces --tags @tag",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L1"}},
				},
			},
			wantArgs: []string{"--tags", "@L1"},
		},
		{
			name: "multiple require tags produce separate --tags args",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L1", "@smoke"}},
				},
			},
			wantArgs: []string{"--tags", "@L1", "--tags", "@smoke"},
		},
		{
			name: "single exclude tag with @ prefix produces --tags ~@tag",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip"}},
				},
			},
			wantArgs: []string{"--tags", "~@skip:wip"},
		},
		{
			name: "single exclude tag without @ prefix gets @ prepended",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"slow"}},
				},
			},
			wantArgs: []string{"--tags", "~@slow"},
		},
		{
			name: "multiple exclude tags produce separate --tags args",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{ExcludeTags: []string{"@skip:wip", "@skip:broken"}},
				},
			},
			wantArgs: []string{"--tags", "~@skip:wip", "--tags", "~@skip:broken"},
		},
		{
			name: "any-of tags joined with comma as single --tags arg",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{AnyOfTags: []string{"@L0", "@L1"}},
				},
			},
			wantArgs: []string{"--tags", "@L0,@L1"},
		},
		{
			name: "single any-of tag",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{AnyOfTags: []string{"@smoke"}},
				},
			},
			wantArgs: []string{"--tags", "@smoke"},
		},
		{
			name: "combined require, exclude, and any-of tags",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{
						RequireTags: []string{"@L1"},
						ExcludeTags: []string{"@skip:wip"},
						AnyOfTags:   []string{"@smoke", "@regression"},
					},
				},
			},
			wantArgs: []string{
				"--tags", "@L1",
				"--tags", "~@skip:wip",
				"--tags", "@smoke,@regression",
			},
		},
		{
			name: "multiple selectors produce combined args",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{RequireTags: []string{"@L1"}},
					{ExcludeTags: []string{"@skip:wip"}},
				},
			},
			wantArgs: []string{
				"--tags", "@L1",
				"--tags", "~@skip:wip",
			},
		},
		{
			name:     "empty filter returns empty slice",
			filter:   core.TagFilter{},
			wantArgs: nil,
		},
		{
			name: "empty selectors list returns empty slice",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{},
			},
			wantArgs: nil,
		},
		{
			name: "selector with all empty tag slices returns empty slice",
			filter: core.TagFilter{
				Selectors: []core.TagFilterSelector{
					{
						RequireTags: []string{},
						ExcludeTags: []string{},
						AnyOfTags:   []string{},
					},
				},
			},
			wantArgs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translator.TranslateTagFilter(tt.filter)

			// Compare lengths
			if len(got) != len(tt.wantArgs) {
				t.Fatalf("TranslateTagFilter() returned %d args %v, want %d args %v",
					len(got), got, len(tt.wantArgs), tt.wantArgs)
			}

			// Compare each element
			for i := range got {
				if got[i] != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %q, want %q (full result: %v)", i, got[i], tt.wantArgs[i], got)
				}
			}
		})
	}
}
