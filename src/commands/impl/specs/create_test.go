// File: src/commands/impl/specs/create_test.go
package specs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Intent: Test specification creation command core functionality
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Each test focuses on a single behavior
//
// Easy to change:
//   - Test data is clearly separated from test logic
//
// Hard to break:
//   - Tests cover happy path and error cases
//   - File system operations use t.TempDir() for isolation

func TestLoadPromptWithFallback(t *testing.T) {
	tests := []struct {
		name              string
		createLocalPrompt bool
		localContent      string
		wantContainsLocal bool
		wantErr           bool
	}{
		{
			name:              "load local prompt when available",
			createLocalPrompt: true,
			localContent:      "# Local custom prompt\nGenerate specification",
			wantContainsLocal: true,
			wantErr:           false,
		},
		{
			name:              "fallback to embedded when local not found",
			createLocalPrompt: false,
			localContent:      "",
			wantContainsLocal: false,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create local prompt if needed
			if tt.createLocalPrompt {
				localPath := filepath.Join(tmpDir, ".r2r", "prompts", "specs", "specification.md")
				if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
					t.Fatalf("failed to create local prompt directory: %v", err)
				}
				if err := os.WriteFile(localPath, []byte(tt.localContent), 0644); err != nil {
					t.Fatalf("failed to write local prompt: %v", err)
				}
			}

			got, err := loadPromptWithFallback(tmpDir, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("loadPromptWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if tt.createLocalPrompt {
				if !strings.Contains(got, "Local custom prompt") {
					t.Errorf("expected local prompt content, got: %s", got)
				}
			} else {
				// Should get embedded prompt with Go template syntax
				if !strings.Contains(got, "# Generate Gherkin Specification") {
					t.Errorf("expected embedded prompt content, got: %s", got)
				}
			}
		})
	}
}

func TestLoadPromptWithFallback_CustomPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create custom prompt file
	customPath := filepath.Join(tmpDir, "custom", "my-prompt.md")
	if err := os.MkdirAll(filepath.Dir(customPath), 0755); err != nil {
		t.Fatal(err)
	}
	customContent := "# Custom Prompt\nGenerate specifications using this custom prompt"
	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		customPath  string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "load custom prompt from absolute path",
			customPath:  customPath,
			wantErr:     false,
			wantContain: "Custom Prompt",
		},
		{
			name:        "load custom prompt from relative path",
			customPath:  "custom/my-prompt.md",
			wantErr:     false,
			wantContain: "Custom Prompt",
		},
		{
			name:        "custom prompt not found",
			customPath:  "nonexistent/prompt.md",
			wantErr:     true,
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadPromptWithFallback(tmpDir, tt.customPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("loadPromptWithFallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantContain != "" {
				if !strings.Contains(got, tt.wantContain) {
					t.Errorf("loadPromptWithFallback() = %q, want to contain %q", got, tt.wantContain)
				}
			}
		})
	}
}

func TestStripAgentNoise(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "remove markdown code fences",
			raw:  "```gherkin\nFeature: test\n```",
			want: "Feature: test",
		},
		{
			name: "remove initialization message",
			raw:  "**Initialized and ready**\n\nFeature: test",
			want: "Feature: test",
		},
		{
			name: "remove multiple noise patterns",
			raw:  "I'll help you create a specification.\n\nFeature: test",
			want: "Feature: test",
		},
		{
			name: "preserve clean content",
			raw:  "Feature: test\n  Rule: test",
			want: "Feature: test\n  Rule: test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAgentNoise(tt.raw)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("stripAgentNoise()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestStripAgentNoise_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "clean gherkin no noise",
			raw:  "Feature: test\n  Rule: test",
			want: "Feature: test\n  Rule: test",
		},
		{
			name: "code fences with language",
			raw:  "```gherkin\nFeature: test\n```",
			want: "Feature: test",
		},
		{
			name: "code fences without language",
			raw:  "```\nFeature: test\n```",
			want: "Feature: test",
		},
		{
			name: "multiple noise patterns",
			raw:  "**Initialized**\nI'll help you\n\nFeature: test",
			want: "Feature: test",
		},
		{
			name: "noise after feature (should keep)",
			raw:  "Feature: test\n  Some explanation",
			want: "Feature: test\n  Some explanation",
		},
		{
			name: "empty input",
			raw:  "",
			want: "",
		},
		{
			name: "only noise",
			raw:  "I'll help you\nHere is",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripAgentNoise(tt.raw)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("stripAgentNoise()\ngot:  %q\nwant: %q", got, tt.want)
			}
		})
	}
}
