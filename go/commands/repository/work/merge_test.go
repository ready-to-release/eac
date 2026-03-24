//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/commands/repository/work/internal"
	"github.com/ready-to-release/eac/go/core/logging"
	eactesting "github.com/ready-to-release/eac/go/core/testing"
)

// TestParseMergeConfig tests the configuration parsing
func TestParseMergeConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *mergeConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *mergeConfig) {
				if config.targetBranch != "main" {
					t.Errorf("expected targetBranch='main', got '%s'", config.targetBranch)
				}
				if config.noSquash {
					t.Error("expected noSquash=false, got true")
				}
				if config.keepWorktree {
					t.Error("expected keepWorktree=false, got true")
				}
				if config.base.Debug {
					t.Error("expected debug=false, got true")
				}
			},
		},
		{
			name: "with --target flag",
			args: []string{"--target=develop"},
			validate: func(t *testing.T, config *mergeConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
			},
		},
		{
			name: "with --no-squash flag",
			args: []string{"--no-squash"},
			validate: func(t *testing.T, config *mergeConfig) {
				if !config.noSquash {
					t.Error("expected noSquash=true, got false")
				}
			},
		},
		{
			name: "with --keep-worktree flag",
			args: []string{"--keep-worktree"},
			validate: func(t *testing.T, config *mergeConfig) {
				if !config.keepWorktree {
					t.Error("expected keepWorktree=true, got false")
				}
			},
		},
		{
			name: "with --debug flag",
			args: []string{"--debug"},
			validate: func(t *testing.T, config *mergeConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with -d shorthand",
			args: []string{"-d"},
			validate: func(t *testing.T, config *mergeConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--target=develop", "--no-squash", "--keep-worktree", "--debug"},
			validate: func(t *testing.T, config *mergeConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
				if !config.noSquash {
					t.Error("expected noSquash=true, got false")
				}
				if !config.keepWorktree {
					t.Error("expected keepWorktree=true, got false")
				}
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eactesting.RequireGitRepository(t)

			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args (simulate command: eac work merge <args>)
			os.Args = append([]string{"eac", "work", "merge"}, tt.args...)

			config, err := parseMergeConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestMergeConfigDefaults tests default configuration values
func TestMergeConfigDefaults(t *testing.T) {
	eactesting.RequireGitRepository(t)

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"clie", "work", "merge"}

	config, err := parseMergeConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.targetBranch != "main" {
		t.Errorf("expected default targetBranch='main', got '%s'", config.targetBranch)
	}

	if config.noSquash {
		t.Error("expected default noSquash=false, got true")
	}

	if config.keepWorktree {
		t.Error("expected default keepWorktree=false, got true")
	}

	if config.base.Debug {
		t.Error("expected default debug=false, got true")
	}

	if config.base.RepoRoot == "" {
		t.Error("expected repoRoot to be set")
	}

	if config.currentBranch == "" {
		t.Error("expected currentBranch to be set")
	}
}

// TestValidateMergeEnvironment tests validation logic
func TestValidateMergeEnvironment(t *testing.T) {
	tests := []struct {
		name          string
		currentBranch string
		expectError   bool
		errorContains string
	}{
		{
			name:          "fail on main branch",
			currentBranch: "main",
			expectError:   true,
			errorContains: "cannot merge main into itself",
		},
		{
			name:          "fail on master branch",
			currentBranch: "master",
			expectError:   true,
			errorContains: "cannot merge main into itself",
		},
		{
			name:          "allow feature branch",
			currentBranch: "feature/test",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a logger for the test
			logger, err := logging.NewDefault("test", ".")
			if err != nil {
				t.Fatalf("failed to create logger: %v", err)
			}

			// Create a base config with default git ops
			baseConfig := &internal.BaseConfig{
				GitOps:   internal.NewDefaultGitOps("."),
				RepoRoot: ".",
				Logger:   logger,
			}

			config := &mergeConfig{
				base:          baseConfig,
				currentBranch: tt.currentBranch,
				targetBranch:  "main",
			}

			err = validateMergeEnvironment(config)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				// Skip if worktree not clean (we can't control this in tests)
				if strings.Contains(err.Error(), "uncommitted changes") {
					t.Skip("Worktree not clean")
				}
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			}
		})
	}
}
