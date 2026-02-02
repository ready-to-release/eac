//go:build L1 && ov
// +build L1,ov

// File: go/cli/eac/impl/create/spec/create_test.go
package spec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			name:              "error when contracts not found (no embedded fallback)",
			createLocalPrompt: false,
			localContent:      "",
			wantContainsLocal: false,
			wantErr:           true, // No embedded fallback anymore
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create local prompt if needed (in new contract structure)
			if tt.createLocalPrompt {
				localPath := filepath.Join(tmpDir, ".r2r", "contracts", "ai", "specifications", "0.1.0", "specification.md")
				if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
					t.Fatalf("failed to create local contract prompt directory: %v", err)
				}
				if err := os.WriteFile(localPath, []byte(tt.localContent), 0644); err != nil {
					t.Fatalf("failed to write local contract prompt: %v", err)
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
			}
			// No else clause - if not local, test expects an error (wantErr: true)
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

// ==============================================================================
// parseConfig Tests
// ==============================================================================

// withArgs temporarily sets os.Args for testing and restores it after
func withArgs(args []string, fn func()) {
	oldArgs := os.Args
	os.Args = args
	defer func() { os.Args = oldArgs }()
	fn()
}

func TestParseConfig_Flags(t *testing.T) {
	// Create a temp directory to act as a git repository
	tmpDir := t.TempDir()

	// Create .git directory to simulate a git repo
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	// Change to temp directory so repository.GetRepositoryRoot works
	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	tests := []struct {
		name        string
		args        []string
		wantDebug   bool
		wantForce   bool
		wantModule  string
		wantOutput  string
		wantPrompt  string
		wantDesc    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "simple description",
			args:     []string{"r2r", "specs", "create", "Add user authentication"},
			wantDesc: "Add user authentication",
			wantErr:  false,
		},
		{
			name:      "debug flag short",
			args:      []string{"r2r", "specs", "create", "-d", "Test description"},
			wantDebug: true,
			wantDesc:  "Test description",
			wantErr:   false,
		},
		{
			name:      "debug flag long",
			args:      []string{"r2r", "specs", "create", "--debug", "Test description"},
			wantDebug: true,
			wantDesc:  "Test description",
			wantErr:   false,
		},
		{
			name:      "force flag short",
			args:      []string{"r2r", "specs", "create", "-f", "Test description"},
			wantForce: true,
			wantDesc:  "Test description",
			wantErr:   false,
		},
		{
			name:      "force flag long",
			args:      []string{"r2r", "specs", "create", "--force", "Test description"},
			wantForce: true,
			wantDesc:  "Test description",
			wantErr:   false,
		},
		{
			name:       "module flag short",
			args:       []string{"r2r", "specs", "create", "-m", "eac-cli", "Test description"},
			wantModule: "eac-cli",
			wantDesc:   "Test description",
			wantErr:    false,
		},
		{
			name:       "module flag long",
			args:       []string{"r2r", "specs", "create", "--module", "core", "Test description"},
			wantModule: "core",
			wantDesc:   "Test description",
			wantErr:    false,
		},
		{
			name:       "output flag short",
			args:       []string{"r2r", "specs", "create", "-o", "custom/path.feature", "Test description"},
			wantOutput: "custom/path.feature",
			wantDesc:   "Test description",
			wantErr:    false,
		},
		{
			name:       "output flag long",
			args:       []string{"r2r", "specs", "create", "--output", "specs/out.feature", "Test description"},
			wantOutput: "specs/out.feature",
			wantDesc:   "Test description",
			wantErr:    false,
		},
		{
			name:       "prompt flag",
			args:       []string{"r2r", "specs", "create", "--prompt", "custom-prompt.md", "Test description"},
			wantPrompt: "custom-prompt.md",
			wantDesc:   "Test description",
			wantErr:    false,
		},
		{
			name:       "multiple flags combined",
			args:       []string{"r2r", "specs", "create", "-d", "-f", "-m", "r2r-cli", "Complex feature description"},
			wantDebug:  true,
			wantForce:  true,
			wantModule: "r2r-cli",
			wantDesc:   "Complex feature description",
			wantErr:    false,
		},
		{
			name:     "multi-word description",
			args:     []string{"r2r", "specs", "create", "Add", "user", "authentication", "with", "email"},
			wantDesc: "Add user authentication with email",
			wantErr:  false,
		},
		{
			name:        "missing description",
			args:        []string{"r2r", "specs", "create"},
			wantErr:     true,
			errContains: "description is required",
		},
		{
			name:        "empty description after flags",
			args:        []string{"r2r", "specs", "create", "-d"},
			wantErr:     true,
			errContains: "description is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config *SpecsConfig
			var err error

			withArgs(tt.args, func() {
				config, err = parseConfig()
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("parseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !strings.Contains(err.Error(), tt.errContains) {
						t.Errorf("parseConfig() error = %v, want error containing %q", err, tt.errContains)
					}
				}
				return
			}

			if config.Debug != tt.wantDebug {
				t.Errorf("parseConfig() Debug = %v, want %v", config.Debug, tt.wantDebug)
			}
			if config.Force != tt.wantForce {
				t.Errorf("parseConfig() Force = %v, want %v", config.Force, tt.wantForce)
			}
			if config.Module != tt.wantModule {
				t.Errorf("parseConfig() Module = %q, want %q", config.Module, tt.wantModule)
			}
			if config.OutputPath != tt.wantOutput {
				t.Errorf("parseConfig() OutputPath = %q, want %q", config.OutputPath, tt.wantOutput)
			}
			if config.PromptPath != tt.wantPrompt {
				t.Errorf("parseConfig() PromptPath = %q, want %q", config.PromptPath, tt.wantPrompt)
			}
			if config.Description != tt.wantDesc {
				t.Errorf("parseConfig() Description = %q, want %q", config.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseConfig_DescriptionTruncation(t *testing.T) {
	// Create a temp directory to act as a git repository
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	// Create a very long description (> 1000 chars)
	longDesc := strings.Repeat("a", 1500)
	args := []string{"r2r", "specs", "create", longDesc}

	var config *SpecsConfig
	var err error

	withArgs(args, func() {
		config, err = parseConfig()
	})

	if err != nil {
		t.Fatalf("parseConfig() unexpected error: %v", err)
	}

	// Description should be truncated to 1000 chars
	if len(config.Description) != 1000 {
		t.Errorf("parseConfig() Description length = %d, want 1000", len(config.Description))
	}
}

func TestTruncateForLog(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "short string unchanged",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "exact length unchanged",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "long string truncated",
			input:  "hello world",
			maxLen: 5,
			want:   "hello...",
		},
		{
			name:   "empty string",
			input:  "",
			maxLen: 10,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateForLog(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateForLog() = %q, want %q", got, tt.want)
			}
		})
	}
}
