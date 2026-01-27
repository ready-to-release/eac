//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/eac/commands/impl/work/internal"
	"github.com/ready-to-release/eac/go/eac/core/logging"
	eactesting "github.com/ready-to-release/eac/go/eac/core/testing"
)

// TestParsePRConfig tests the configuration parsing
func TestParsePRConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *prConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *prConfig) {
				if config.targetBranch != "main" {
					t.Errorf("expected targetBranch='main', got '%s'", config.targetBranch)
				}
				if config.customTitle != "" {
					t.Errorf("expected customTitle='', got '%s'", config.customTitle)
				}
				if config.base.Debug {
					t.Error("expected debug=false, got true")
				}
			},
		},
		{
			name: "with --target flag",
			args: []string{"--target=develop"},
			validate: func(t *testing.T, config *prConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
			},
		},
		{
			name: "with --title flag (equals syntax)",
			args: []string{"--title=Custom PR Title"},
			validate: func(t *testing.T, config *prConfig) {
				if config.customTitle != "Custom PR Title" {
					t.Errorf("expected customTitle='Custom PR Title', got '%s'", config.customTitle)
				}
			},
		},
		{
			name: "with --title flag (space syntax)",
			args: []string{"--title", "Another Title"},
			validate: func(t *testing.T, config *prConfig) {
				if config.customTitle != "Another Title" {
					t.Errorf("expected customTitle='Another Title', got '%s'", config.customTitle)
				}
			},
		},
		{
			name: "with --debug flag",
			args: []string{"--debug"},
			validate: func(t *testing.T, config *prConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with -d shorthand",
			args: []string{"-d"},
			validate: func(t *testing.T, config *prConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--target=develop", "--title", "My PR", "--debug"},
			validate: func(t *testing.T, config *prConfig) {
				if config.targetBranch != "develop" {
					t.Errorf("expected targetBranch='develop', got '%s'", config.targetBranch)
				}
				if config.customTitle != "My PR" {
					t.Errorf("expected customTitle='My PR', got '%s'", config.customTitle)
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

			// Set test args (simulate command: r2r work pr <args>)
			os.Args = append([]string{"r2r", "work", "pr"}, tt.args...)

			config, err := parsePRConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestParsePRConfig_MissingTitleValue tests error handling
func TestParsePRConfig_MissingTitleValue(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "pr", "--title"}

	_, err := parsePRConfig()

	if err == nil {
		t.Error("expected error for --title without value, got nil")
	}

	if err != nil && !strings.Contains(err.Error(), "--title requires a value") {
		t.Errorf("expected error about missing title value, got: %v", err)
	}
}

// TestPRConfigDefaults tests default configuration values
func TestPRConfigDefaults(t *testing.T) {
	eactesting.RequireGitRepository(t)

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "pr"}

	config, err := parsePRConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.targetBranch != "main" {
		t.Errorf("expected default targetBranch='main', got '%s'", config.targetBranch)
	}

	if config.customTitle != "" {
		t.Errorf("expected default customTitle='', got '%s'", config.customTitle)
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

// TestValidatePREnvironment tests validation logic
func TestValidatePREnvironment(t *testing.T) {
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
			errorContains: "cannot create PR from main",
		},
		{
			name:          "fail on master branch",
			currentBranch: "master",
			expectError:   true,
			errorContains: "cannot create PR from main",
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
				GitOps:   internal.GetGitOps("."),
				RepoRoot: ".",
				Logger:   logger,
			}

			config := &prConfig{
				base:          baseConfig,
				currentBranch: tt.currentBranch,
				targetBranch:  "main",
			}

			err = validatePREnvironment(config)

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				// Skip if worktree not clean or gh not available
				if strings.Contains(err.Error(), "uncommitted changes") ||
					strings.Contains(err.Error(), "gh CLI not found") {
					t.Skip("Worktree not clean or gh CLI not available")
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

// TestGeneratePRTitle tests PR title generation
func TestGeneratePRTitle(t *testing.T) {
	tests := []struct {
		name       string
		commits    []string
		branchName string
		want       string
	}{
		{
			name:       "use first commit message",
			commits:    []string{"feat: add authentication", "fix: resolve bug"},
			branchName: "feature/authentication",
			want:       "feat: add authentication",
		},
		{
			name:       "format branch name when no commits",
			commits:    []string{},
			branchName: "feature/user-profile",
			want:       "Add User Profile",
		},
		{
			name:       "handle branch without slash",
			commits:    []string{},
			branchName: "hotfix",
			want:       "Changes from hotfix",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generatePRTitle(tt.commits, tt.branchName)
			if got != tt.want {
				t.Errorf("generatePRTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGeneratePRDescription tests PR description generation
func TestGeneratePRDescription(t *testing.T) {
	commits := []string{"feat: add feature", "fix: bug fix"}
	diffStat := "file1.go | 10 +++++++\nfile2.go | 5 +++++\n2 files changed, 15 insertions(+)"

	description := generatePRDescription(commits, diffStat)

	// Check that description contains expected sections
	if !strings.Contains(description, "## Summary") {
		t.Error("description should contain Summary section")
	}

	if !strings.Contains(description, "## Changes") {
		t.Error("description should contain Changes section")
	}

	if !strings.Contains(description, "## Test Plan") {
		t.Error("description should contain Test Plan section")
	}

	if !strings.Contains(description, "feat: add feature") {
		t.Error("description should contain commit messages")
	}

	if !strings.Contains(description, diffStat) {
		t.Error("description should contain diff stat")
	}
}
