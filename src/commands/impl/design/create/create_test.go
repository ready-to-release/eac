// File: src/commands/impl/design/create/create_test.go
package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Intent: Test design create command core functionality
//
// Design (Three Rules of Vibe Coding):
//
// Easy to understand:
//   - Table-driven tests with clear test case names
//   - Each test focuses on a single behavior
//
// Easy to change:
//   - Test data is clearly separated from test logic
//   - Helper functions reduce duplication
//
// Hard to break:
//   - Tests cover happy path and error cases
//   - File system operations use t.TempDir() for isolation

func TestParseConfig(t *testing.T) {
	// Save original args and restore after test
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	tests := []struct {
		name      string
		args      []string
		wantErr   bool
		errMsg    string
		checkFunc func(*testing.T, *DesignConfig)
	}{
		{
			name:    "valid module name",
			args:    []string{"cmd", "design", "create", "src-cli"},
			wantErr: false,
			checkFunc: func(t *testing.T, config *DesignConfig) {
				if config.Module != "src-cli" {
					t.Errorf("Module = %q, want %q", config.Module, "src-cli")
				}
				if config.Debug {
					t.Error("Debug should be false by default")
				}
				if config.Force {
					t.Error("Force should be false by default")
				}
			},
		},
		{
			name:    "with debug flag",
			args:    []string{"cmd", "design", "create", "src-cli", "--debug"},
			wantErr: false,
			checkFunc: func(t *testing.T, config *DesignConfig) {
				if !config.Debug {
					t.Error("Debug should be true")
				}
			},
		},
		{
			name:    "with force flag",
			args:    []string{"cmd", "design", "create", "src-cli", "--force"},
			wantErr: false,
			checkFunc: func(t *testing.T, config *DesignConfig) {
				if !config.Force {
					t.Error("Force should be true")
				}
			},
		},
		{
			name:    "with output path",
			args:    []string{"cmd", "design", "create", "src-cli", "--output", "custom/path.dsl"},
			wantErr: false,
			checkFunc: func(t *testing.T, config *DesignConfig) {
				if config.OutputPath != "custom/path.dsl" {
					t.Errorf("OutputPath = %q, want %q", config.OutputPath, "custom/path.dsl")
				}
			},
		},
		{
			name:    "with all flags",
			args:    []string{"cmd", "design", "create", "src-cli", "-d", "-f", "-o", "out.dsl"},
			wantErr: false,
			checkFunc: func(t *testing.T, config *DesignConfig) {
				if !config.Debug {
					t.Error("Debug should be true")
				}
				if !config.Force {
					t.Error("Force should be true")
				}
				if config.OutputPath != "out.dsl" {
					t.Errorf("OutputPath = %q, want %q", config.OutputPath, "out.dsl")
				}
			},
		},
		{
			name:    "missing module name",
			args:    []string{"cmd", "design", "create"},
			wantErr: true,
			errMsg:  "module name required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for repository root
			tmpDir := t.TempDir()

			// Setup .git directory to make it look like a repo
			gitDir := filepath.Join(tmpDir, ".git")
			if err := os.MkdirAll(gitDir, 0755); err != nil {
				t.Fatal(err)
			}

			// Change to temp directory
			origDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			defer os.Chdir(origDir)
			if err := os.Chdir(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Set args
			os.Args = tt.args

			// Run parseConfig
			config, err := parseConfig()

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
				return
			}

			// Run check function
			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestDetermineSourcePath(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		module     string
		modulePath string
		setupFunc  func(string) error
		want       string
	}{
		{
			name:       "standard location exists",
			module:     "src-cli",
			modulePath: "src-cli",
			setupFunc: func(root string) error {
				return os.MkdirAll(filepath.Join(root, "src", "src-cli"), 0755)
			},
			want: "src/src-cli",
		},
		{
			name:       "custom location exists",
			module:     "custom-module",
			modulePath: "custom/path",
			setupFunc: func(root string) error {
				return os.MkdirAll(filepath.Join(root, "custom", "path"), 0755)
			},
			want: "custom/path",
		},
		{
			name:       "neither exists - returns standard",
			module:     "nonexistent",
			modulePath: "nonexistent",
			setupFunc:  func(root string) error { return nil },
			want:       "src/nonexistent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup directories
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Run function
			got := determineSourcePath(tmpDir, tt.module, tt.modulePath)

			// Normalize paths for comparison
			want := filepath.Join(tmpDir, tt.want)
			if got != want {
				t.Errorf("determineSourcePath() = %q, want %q", got, want)
			}
		})
	}
}

func TestValidateModuleExists(t *testing.T) {
	tests := []struct {
		name      string
		module    string
		setupFunc func(string) error
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "source code exists",
			module: "src-cli",
			setupFunc: func(root string) error {
				sourcePath := filepath.Join(root, "src", "src-cli")
				return os.MkdirAll(sourcePath, 0755)
			},
			wantErr: false,
		},
		{
			name:      "source code does not exist",
			module:    "nonexistent",
			setupFunc: func(root string) error { return nil },
			wantErr:   true,
			errMsg:    "source code not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Create config
			config := &DesignConfig{
				Module:       tt.module,
				SourcePath:   filepath.Join(tmpDir, "src", tt.module),
				TemplateRoot: tmpDir,
			}

			// Run validation
			err := validateModuleExists(config)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("validateModuleExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error message = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestBuildContractBasedPrompt(t *testing.T) {
	tests := []struct {
		name      string
		config    *DesignConfig
		setupFunc func(string) error
		wantErr   bool
		checkFunc func(*testing.T, string)
	}{
		{
			name: "builds prompt with module context",
			config: &DesignConfig{
				Module:     "src-cli",
				SourcePath: "/path/to/src/src-cli",
			},
			setupFunc: func(root string) error {
				// Create contract files
				contractDir := filepath.Join(root, "contracts", "ai", "design", "0.1.0")
				if err := os.MkdirAll(contractDir, 0755); err != nil {
					return err
				}

				// Create design.md
				designMd := filepath.Join(contractDir, "design.md")
				designContent := `# Structurizr DSL Architecture Generation

Generate a complete Structurizr DSL workspace.

## Output Requirements

1. Valid DSL
2. Clean Format
3. Complete Structure
`
				if err := os.WriteFile(designMd, []byte(designContent), 0644); err != nil {
					return err
				}

				// Create contract.yml
				contractYml := filepath.Join(contractDir, "contract.yml")
				contractContent := `version: "0.1.0"
name: "Structurizr DSL Architecture"
description: "Formal contract"

workspace_name_pattern: "^[A-Z][A-Za-z0-9 -]+$"
identifier_pattern: "^[a-z][a-z0-9_]*$"

required_elements:
  - "workspace"
  - "model"
  - "views"
`
				if err := os.WriteFile(contractYml, []byte(contractContent), 0644); err != nil {
					return err
				}

				// Create anti-corruption.yml
				antiCorruptionYml := filepath.Join(contractDir, "anti-corruption.yml")
				antiCorruptionContent := `version: "0.1.0"
name: "Structurizr DSL Anti-Corruption Rules"

forbidden_prefixes:
  - "Here is"
  - "I'll"

forbidden_contains:
  - "generating"

forbidden_emojis:
  - "🚀"

agent_signatures: []
`
				return os.WriteFile(antiCorruptionYml, []byte(antiCorruptionContent), 0644)
			},
			wantErr: false,
			checkFunc: func(t *testing.T, prompt string) {
				// Check that prompt contains expected sections
				expectedSections := []string{
					"Structurizr DSL",
					"Target Module",
					"src-cli",
					"Analysis Instructions",
					"Generation Requirements",
					">>>>>>>>>>INPUT STARTS NOW<<<<<<<<<<<",
				}

				for _, section := range expectedSections {
					if !strings.Contains(prompt, section) {
						t.Errorf("prompt missing expected section: %q", section)
					}
				}
			},
		},
		{
			name: "fails when contract files missing",
			config: &DesignConfig{
				Module:     "src-cli",
				SourcePath: "/path/to/src/src-cli",
			},
			setupFunc: func(root string) error {
				// Don't create contract files
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup
			if err := tt.setupFunc(tmpDir); err != nil {
				t.Fatal(err)
			}

			// Set template root
			tt.config.TemplateRoot = tmpDir

			// Run function
			prompt, err := buildContractBasedPrompt(tt.config)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("buildContractBasedPrompt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Run check function
			if tt.checkFunc != nil {
				tt.checkFunc(t, prompt)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func(string) (string, error)
		want      bool
	}{
		{
			name: "file exists",
			setupFunc: func(root string) (string, error) {
				path := filepath.Join(root, "test.txt")
				err := os.WriteFile(path, []byte("test"), 0644)
				return path, err
			},
			want: true,
		},
		{
			name: "file does not exist",
			setupFunc: func(root string) (string, error) {
				path := filepath.Join(root, "nonexistent.txt")
				return path, nil
			},
			want: false,
		},
		{
			name: "directory exists (not a file)",
			setupFunc: func(root string) (string, error) {
				path := filepath.Join(root, "testdir")
				err := os.MkdirAll(path, 0755)
				return path, err
			},
			want: true, // os.Stat returns true for directories too
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup
			path, err := tt.setupFunc(tmpDir)
			if err != nil {
				t.Fatal(err)
			}

			// Run function
			got := fileExists(path)

			if got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCreateCommandArgs(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantMod string
		wantDbg bool
		wantFrc bool
		wantOut string
		wantErr bool
	}{
		{
			name:    "simple module",
			args:    []string{"cmd", "design", "create", "my-module"},
			wantMod: "my-module",
			wantDbg: false,
			wantFrc: false,
			wantOut: "",
			wantErr: false,
		},
		{
			name:    "with debug flag",
			args:    []string{"cmd", "design", "create", "my-module", "--debug"},
			wantMod: "my-module",
			wantDbg: true,
			wantFrc: false,
			wantOut: "",
			wantErr: false,
		},
		{
			name:    "with short debug flag",
			args:    []string{"cmd", "design", "create", "my-module", "-d"},
			wantMod: "my-module",
			wantDbg: true,
			wantFrc: false,
			wantOut: "",
			wantErr: false,
		},
		{
			name:    "with force flag",
			args:    []string{"cmd", "design", "create", "my-module", "--force"},
			wantMod: "my-module",
			wantDbg: false,
			wantFrc: true,
			wantOut: "",
			wantErr: false,
		},
		{
			name:    "with output flag",
			args:    []string{"cmd", "design", "create", "my-module", "--output", "custom.dsl"},
			wantMod: "my-module",
			wantDbg: false,
			wantFrc: false,
			wantOut: "custom.dsl",
			wantErr: false,
		},
		{
			name:    "with all flags",
			args:    []string{"cmd", "design", "create", "my-module", "-d", "-f", "-o", "out.dsl"},
			wantMod: "my-module",
			wantDbg: true,
			wantFrc: true,
			wantOut: "out.dsl",
			wantErr: false,
		},
		{
			name:    "no module specified",
			args:    []string{"cmd", "design", "create"},
			wantErr: true,
		},
		{
			name:    "no create keyword",
			args:    []string{"cmd", "design"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modulePath, flags, err := parseCreateCommandArgs(tt.args)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCreateCommandArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check module
			if modulePath != tt.wantMod {
				t.Errorf("module = %q, want %q", modulePath, tt.wantMod)
			}

			// Check flags
			if flags.debug != tt.wantDbg {
				t.Errorf("debug = %v, want %v", flags.debug, tt.wantDbg)
			}
			if flags.force != tt.wantFrc {
				t.Errorf("force = %v, want %v", flags.force, tt.wantFrc)
			}
			if flags.outputPath != tt.wantOut {
				t.Errorf("outputPath = %q, want %q", flags.outputPath, tt.wantOut)
			}
		})
	}
}
