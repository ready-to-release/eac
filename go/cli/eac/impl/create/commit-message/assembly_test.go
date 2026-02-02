//go:build L1 && ov
// +build L1,ov

package commitmessage

import (
	"strings"
	"testing"
)

func TestCombineCommitSections_Deduplication(t *testing.T) {
	tests := []struct {
		name           string
		topLevel       string
		moduleSections []string
		wantModules    []string
		wantNotContain []string
		wantSeparators int
	}{
		{
			name:     "deduplicates identical module names",
			topLevel: "feat(multi-module): add feature",
			moduleSections: []string{
				"core\n--------\ncore: feat: first change",
				"contracts\n---------\ncontracts: feat: update contracts",
				"core\n--------\ncore: feat: second change (duplicate)",
			},
			wantModules: []string{
				"core",
				"contracts",
			},
			wantNotContain: []string{
				"second change (duplicate)",
			},
			wantSeparators: 1, // Only one separator between two unique modules
		},
		{
			name:     "keeps first occurrence of duplicate module",
			topLevel: "feat(multi-module): refactor",
			moduleSections: []string{
				"docs\n----\ndocs: docs: update readme",
				"specs\n-----\nspecs: feat: add specs",
				"docs\n----\ndocs: docs: update changelog (duplicate)",
			},
			wantModules: []string{
				"docs",
				"specs",
			},
			wantNotContain: []string{
				"update changelog (duplicate)",
			},
			wantSeparators: 1,
		},
		{
			name:     "no duplicates - passes through unchanged",
			topLevel: "fix(multi-module): fix bugs",
			moduleSections: []string{
				"r2r-cli\n--------\nr2r-cli: fix: resolve issue",
				"core\n---------\ncore: fix: fix bug",
			},
			wantModules: []string{
				"r2r-cli",
				"core",
			},
			wantSeparators: 1,
		},
		{
			name:     "handles empty sections",
			topLevel: "chore(multi-module): cleanup",
			moduleSections: []string{
				"",
				"core\n--------\ncore: chore: cleanup",
				"",
			},
			wantModules: []string{
				"core",
			},
			wantSeparators: 0, // No separators for single module
		},
		{
			name:     "multiple duplicates of same module",
			topLevel: "feat(multi-module): add features",
			moduleSections: []string{
				"eac-cli\n------------\neac-cli: feat: add command 1",
				"eac-cli\n------------\neac-cli: feat: add command 2",
				"eac-cli\n------------\neac-cli: feat: add command 3",
			},
			wantModules: []string{
				"eac-cli",
			},
			wantNotContain: []string{
				"add command 2",
				"add command 3",
			},
			wantSeparators: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := combineCommitSections(tt.topLevel, tt.moduleSections)

			// Check that expected modules are present
			for _, module := range tt.wantModules {
				if !strings.Contains(got, module) {
					t.Errorf("combineCommitSections() missing module %q\nGot:\n%s", module, got)
				}
			}

			// Check that unwanted content is not present (duplicates removed)
			for _, unwanted := range tt.wantNotContain {
				if strings.Contains(got, unwanted) {
					t.Errorf("combineCommitSections() contains unwanted content %q (should be removed)\nGot:\n%s", unwanted, got)
				}
			}

			// Check number of separators (specifically "\n---\n" not just "---")
			separatorCount := strings.Count(got, "\n---\n")
			if separatorCount != tt.wantSeparators {
				t.Errorf("combineCommitSections() separator count = %d, want %d\nGot:\n%s", separatorCount, tt.wantSeparators, got)
			}
		})
	}
}

func TestExtractModuleName(t *testing.T) {
	tests := []struct {
		name    string
		section string
		want    string
	}{
		{
			name:    "extracts module name from valid section",
			section: "core\n--------\ncore: feat: add feature",
			want:    "core",
		},
		{
			name:    "extracts module with underscores",
			section: "my_module\n---------\nmy_module: fix: bug",
			want:    "my_module",
		},
		{
			name:    "extracts module with numbers",
			section: "module123\n---------\nmodule123: chore: update",
			want:    "module123",
		},
		{
			name:    "returns empty for invalid module name (uppercase)",
			section: "MyModule\n--------\nMyModule: feat: add",
			want:    "",
		},
		{
			name:    "returns empty for invalid module name (spaces)",
			section: "my module\n---------\nmy module: feat: add",
			want:    "",
		},
		{
			name:    "returns empty for empty section",
			section: "",
			want:    "",
		},
		{
			name:    "returns empty for header line",
			section: "## Module Header\n---------\nContent",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModuleName(tt.section)
			if got != tt.want {
				t.Errorf("extractModuleName() = %q, want %q", got, tt.want)
			}
		})
	}
}
