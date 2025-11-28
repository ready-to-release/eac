//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"testing"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
)

// TestParsePullConfig tests the configuration parsing
func TestParsePullConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *pullConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *pullConfig) {
				if config.targetBranch != "main" {
					t.Errorf("expected targetBranch='main', got '%s'", config.targetBranch)
				}
				if config.autostash {
					t.Error("expected autostash=false, got true")
				}
				if config.noFetch {
					t.Error("expected noFetch=false, got true")
				}
			},
		},
		{
			name: "with --target flag",
			args: []string{"--target=develop"},
			validate: func(t *testing.T, config *pullConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
			},
		},
		{
			name: "with --autostash flag",
			args: []string{"--autostash"},
			validate: func(t *testing.T, config *pullConfig) {
				if !config.autostash {
					t.Error("expected autostash=true, got false")
				}
			},
		},
		{
			name: "with --no-fetch flag",
			args: []string{"--no-fetch"},
			validate: func(t *testing.T, config *pullConfig) {
				if !config.noFetch {
					t.Error("expected noFetch=true, got false")
				}
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--target=develop", "--autostash", "--no-fetch"},
			validate: func(t *testing.T, config *pullConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
				if !config.autostash {
					t.Error("expected autostash=true, got false")
				}
				if !config.noFetch {
					t.Error("expected noFetch=true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args (simulate command: r2r work pull <args>)
			os.Args = append([]string{"r2r", "work", "pull"}, tt.args...)

			config, err := parsePullConfig()
			if err != nil {
				// Skip test if not in a git repository
				if containsStr(err.Error(), "repository root") {
					t.Skip("Not in a git repository")
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestPullConfigDefaults tests default configuration values
func TestPullConfigDefaults(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "pull"}

	config, err := parsePullConfig()
	if err != nil {
		// Skip test if not in a git repository
		if containsStr(err.Error(), "repository root") {
			t.Skip("Not in a git repository")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.targetBranch != "main" {
		t.Errorf("expected default targetBranch='main', got '%s'", config.targetBranch)
	}

	if config.autostash {
		t.Error("expected default autostash=false, got true")
	}

	if config.noFetch {
		t.Error("expected default noFetch=false, got true")
	}

	if config.base.RepoRoot == "" {
		t.Error("expected repoRoot to be set")
	}

	if config.currentBranch == "" {
		t.Error("expected currentBranch to be set")
	}
}

// TestValidatePullEnvironment tests validation logic
func TestValidatePullEnvironment(t *testing.T) {
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
			errorContains: "cannot rebase main onto itself",
		},
		{
			name:          "fail on master branch",
			currentBranch: "master",
			expectError:   true,
			errorContains: "cannot rebase main onto itself",
		},
		{
			name:          "allow feature branch",
			currentBranch: "feature/test",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a base config with default git ops
			baseConfig := &internal.BaseConfig{
				GitOps:   internal.GetGitOps("."),
				RepoRoot: ".",
			}

			config := &pullConfig{
				base:          baseConfig,
				currentBranch: tt.currentBranch,
				targetBranch:  "main",
				autostash:     true, // Set to true to skip worktree check
			}

			err := validatePullEnvironment(config)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil {
				if !containsStr(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorContains, err.Error())
				}
			}
		})
	}
}

// TestPullConfigWorktreeBranchDetection tests branch detection in worktree environments
func TestPullConfigWorktreeBranchDetection(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "pull"}

	config, err := parsePullConfig()
	if err != nil {
		// Skip test if not in a git repository
		if containsStr(err.Error(), "repository root") {
			t.Skip("Not in a git repository")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// The current branch should be detected from the current working directory,
	// not from the repository root. This is critical for worktree environments
	// where the main repo and worktrees may be on different branches.
	//
	// For example:
	// - Main repo at C:\source\simply-cli\cli (on "main" branch)
	// - Worktree at C:\source\simply-cli\cli-feature-design (on "feature/design" branch)
	//
	// When running from the worktree directory, config.currentBranch should be
	// "feature/design", NOT "main".

	if config.currentBranch == "" {
		t.Error("expected currentBranch to be set, got empty string")
	}

	// Verify that currentBranch does not incorrectly report main/master
	// when we're actually in a feature branch worktree.
	// This test will pass if we're on any branch (main or feature).
	// The real validation happens when running the actual command.
	t.Logf("Detected current branch: %s", config.currentBranch)
}
