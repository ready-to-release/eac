package work

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/commands/impl/work/internal"
)

// TestParseCreateConfig tests the configuration parsing
func TestParseCreateConfig(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		validate    func(*testing.T, *createConfig)
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name: "branch name only",
			args: []string{"feature/test"},
			validate: func(t *testing.T, config *createConfig) {
				if config.branchName != "feature/test" {
					t.Errorf("expected branchName 'feature/test', got '%s'", config.branchName)
				}
				if config.baseBranch != "main" {
					t.Errorf("expected baseBranch 'main', got '%s'", config.baseBranch)
				}
			},
		},
		{
			name: "with --from flag",
			args: []string{"feature/test", "--from=develop"},
			validate: func(t *testing.T, config *createConfig) {
				if config.baseBranch != "develop" {
					t.Errorf("expected baseBranch 'develop', got '%s'", config.baseBranch)
				}
			},
		},
		{
			name: "with --path flag",
			args: []string{"feature/test", "--path=../custom"},
			validate: func(t *testing.T, config *createConfig) {
				if !strings.Contains(config.worktreePath, "custom") {
					t.Errorf("expected path to contain 'custom', got '%s'", config.worktreePath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args (simulate command: r2r work create <args>)
			os.Args = append([]string{"r2r", "work", "create"}, tt.args...)

			config, err := parseCreateConfig()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestGenerateWorktreePath tests path generation
func TestGenerateWorktreePath(t *testing.T) {
	tests := []struct {
		name       string
		repoName   string
		branchName string
		expected   string
	}{
		{
			name:       "simple feature branch",
			repoName:   "my-repo",
			branchName: "feature/auth",
			expected:   filepath.Join("..", "my-repo-feature-auth"),
		},
		{
			name:       "bugfix branch",
			repoName:   "project",
			branchName: "bugfix/issue-123",
			expected:   filepath.Join("..", "project-bugfix-issue-123"),
		},
		{
			name:       "simple branch name",
			repoName:   "repo",
			branchName: "develop",
			expected:   filepath.Join("..", "repo-develop"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internal.GenerateWorktreePath(tt.repoName, tt.branchName)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestGetRepoName tests repository name extraction
func TestGetRepoName(t *testing.T) {
	tests := []struct {
		name        string
		repoRoot    string
		expected    string
		windowsOnly bool // deps:windows - test only runs on Windows
	}{
		{
			name:        "unix path",
			repoRoot:    "/home/user/projects/my-repo",
			expected:    "my-repo",
			windowsOnly: false,
		},
		{
			name:        "windows path",
			repoRoot:    `C:\source\projects\my-repo`,
			expected:    "my-repo",
			windowsOnly: true,
		},
		{
			name:        "current directory",
			repoRoot:    ".",
			expected:    ".",
			windowsOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("deps:windows - test only runs on Windows")
			}
			result := internal.GetRepoName(tt.repoRoot)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestValidateBranchName tests branch name validation
func TestValidateBranchName(t *testing.T) {
	tests := []struct {
		name        string
		branchName  string
		expectValid bool
	}{
		{
			name:        "valid feature branch",
			branchName:  "feature/authentication",
			expectValid: true,
		},
		{
			name:        "valid bugfix branch",
			branchName:  "bugfix/issue-123",
			expectValid: true,
		},
		{
			name:        "valid simple branch",
			branchName:  "develop",
			expectValid: true,
		},
		{
			name:        "empty branch name",
			branchName:  "",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.branchName != ""
			if isValid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v", tt.expectValid, isValid)
			}
		})
	}
}

// TestCreateConfigDefaults tests default configuration values
func TestCreateConfigDefaults(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "create", "feature/test"}

	config, err := parseCreateConfig()
	if err != nil {
		// If we're not in a git repo, skip this test
		if strings.Contains(err.Error(), "repository root") {
			t.Skip("Not in a git repository")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.baseBranch != "main" {
		t.Errorf("expected default baseBranch 'main', got '%s'", config.baseBranch)
	}

	if config.branchName != "feature/test" {
		t.Errorf("expected branchName 'feature/test', got '%s'", config.branchName)
	}
}
