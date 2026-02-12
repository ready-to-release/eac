//go:build L1 && ov
// +build L1,ov

package custom

import (
	"strings"
	"testing"
)

func TestRenderSummary(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSubstr []string
		wantErr    bool
	}{
		{
			name: "single module with type and moniker",
			input: `- moniker: cli
  name: CLI Tool
  type: application
`,
			wantSubstr: []string{
				"Module Contracts Summary",
				"Total Modules: 1",
				"Total Dependencies: 0",
				"application",
				"cli",
			},
		},
		{
			name: "multiple modules with dependencies",
			input: `- moniker: cli
  name: CLI Tool
  type: application
  depends_on:
    - core
    - utils
- moniker: core
  name: Core Library
  type: library
- moniker: utils
  name: Utilities
  type: library
  depends_on:
    - core
`,
			wantSubstr: []string{
				"Total Modules: 3",
				"Total Dependencies: 3",
				"application",
				"library",
				"cli",
				"core",
				"utils",
			},
		},
		{
			name:  "empty list",
			input: "[]",
			wantSubstr: []string{
				"Total Modules: 0",
				"Total Dependencies: 0",
			},
		},
		{
			name:    "invalid yaml",
			input:   "{{invalid",
			wantErr: true,
		},
		{
			name:    "non-list yaml",
			input:   "key: value",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderSummary([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSummary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for _, substr := range tt.wantSubstr {
				if !strings.Contains(got, substr) {
					t.Errorf("RenderSummary() output missing %q.\nGot:\n%s", substr, got)
				}
			}
		})
	}
}

func TestRenderSummary_MonikerAndNameDisplay(t *testing.T) {
	input := `- moniker: my-module
  name: My Module Display Name
  type: service
`
	got, err := RenderSummary([]byte(input))
	if err != nil {
		t.Fatalf("RenderSummary() returned error: %v", err)
	}

	// When name differs from moniker, both should be shown
	if !strings.Contains(got, "my-module") {
		t.Errorf("expected moniker 'my-module' in output, got:\n%s", got)
	}
	if !strings.Contains(got, "My Module Display Name") {
		t.Errorf("expected name 'My Module Display Name' in output, got:\n%s", got)
	}
}

func TestRenderSummary_MonikerSameAsName(t *testing.T) {
	input := `- moniker: cli
  name: cli
  type: application
`
	got, err := RenderSummary([]byte(input))
	if err != nil {
		t.Fatalf("RenderSummary() returned error: %v", err)
	}

	// When name equals moniker, the name should NOT be shown in parens
	// The moniker section should have "cli" but not "cli (cli)"
	lines := strings.Split(got, "\n")
	for _, line := range lines {
		if strings.Contains(line, "cli") && strings.Contains(line, "(cli)") {
			t.Errorf("when name equals moniker, should not show redundant name in parens, got line: %s", line)
		}
	}
}

func TestRenderSummary_ModuleWithoutType(t *testing.T) {
	input := `- moniker: untyped-module
  name: Untyped
`
	got, err := RenderSummary([]byte(input))
	if err != nil {
		t.Fatalf("RenderSummary() returned error: %v", err)
	}

	// Module without type should still be counted
	if !strings.Contains(got, "Total Modules: 1") {
		t.Errorf("expected total modules 1, got:\n%s", got)
	}

	// Moniker should still be listed
	if !strings.Contains(got, "untyped-module") {
		t.Errorf("expected moniker in output, got:\n%s", got)
	}
}

func TestRenderSummary_ModuleWithoutMoniker(t *testing.T) {
	input := `- name: NoMoniker
  type: library
`
	got, err := RenderSummary([]byte(input))
	if err != nil {
		t.Fatalf("RenderSummary() returned error: %v", err)
	}

	// Module without moniker should still be counted in totals
	if !strings.Contains(got, "Total Modules: 1") {
		t.Errorf("expected total modules 1, got:\n%s", got)
	}

	// But should not appear in monikers list (no moniker to display)
	// The module entry should be skipped in the moniker list
	if strings.Contains(got, "NoMoniker") {
		// This is acceptable -- the code skips modules without monikers in the list
		// so NoMoniker should NOT appear
		t.Errorf("module without moniker should not appear in moniker list, got:\n%s", got)
	}
}

func TestRenderSummary_SectionsPresent(t *testing.T) {
	input := `- moniker: test
  type: service
`
	got, err := RenderSummary([]byte(input))
	if err != nil {
		t.Fatalf("RenderSummary() returned error: %v", err)
	}

	expectedSections := []string{
		"=== Module Contracts Summary ===",
		"--- Modules by Type ---",
		"--- Module Monikers ---",
	}

	for _, section := range expectedSections {
		if !strings.Contains(got, section) {
			t.Errorf("expected section %q in output, got:\n%s", section, got)
		}
	}
}

func TestRenderSummary_ErrorMessage(t *testing.T) {
	_, err := RenderSummary([]byte("not: a list"))
	if err == nil {
		t.Fatal("expected error for non-list YAML, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse YAML") {
		t.Errorf("error should mention parsing failure, got: %v", err)
	}
}
