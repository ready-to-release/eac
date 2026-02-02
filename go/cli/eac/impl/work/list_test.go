//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/go/cli/eac/impl/work/internal"
	eactesting "github.com/ready-to-release/eac/go/core/testing"
)

// TestParseListConfig tests the configuration parsing
func TestParseListConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *listConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *listConfig) {
				if config.verbose {
					t.Error("expected verbose=false, got true")
				}
			},
		},
		{
			name: "with --verbose flag",
			args: []string{"--verbose"},
			validate: func(t *testing.T, config *listConfig) {
				if !config.verbose {
					t.Error("expected verbose=true, got false")
				}
			},
		},
		{
			name: "with -v shorthand",
			args: []string{"-v"},
			validate: func(t *testing.T, config *listConfig) {
				if !config.verbose {
					t.Error("expected verbose=true, got false")
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

			// Set test args (simulate command: r2r work list <args>)
			os.Args = append([]string{"r2r", "work", "list"}, tt.args...)

			config, err := parseListConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestFormatWorktreeStatus tests status formatting
func TestFormatWorktreeStatus(t *testing.T) {
	tests := []struct {
		name     string
		clean    bool
		expected string
	}{
		{
			name:     "clean worktree",
			clean:    true,
			expected: "clean",
		},
		{
			name:     "dirty worktree",
			clean:    false,
			expected: "dirty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatWorktreeStatus(tt.clean)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestShortenSHA tests SHA shortening
func TestShortenSHA(t *testing.T) {
	tests := []struct {
		name     string
		sha      string
		expected string
	}{
		{
			name:     "full SHA",
			sha:      "abc123def456789",
			expected: "abc123d",
		},
		{
			name:     "short SHA",
			sha:      "abc123",
			expected: "abc123",
		},
		{
			name:     "7 char SHA",
			sha:      "abc1234",
			expected: "abc1234",
		},
		{
			name:     "empty SHA",
			sha:      "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShortenSHA(tt.sha)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestFormatBranch tests branch name formatting
func TestFormatBranch(t *testing.T) {
	tests := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "normal branch",
			branch:   "main",
			expected: "main",
		},
		{
			name:     "feature branch",
			branch:   "feature/authentication",
			expected: "feature/authentication",
		},
		{
			name:     "empty branch (detached HEAD)",
			branch:   "",
			expected: "(detached)",
		},
		{
			name:     "whitespace branch",
			branch:   "   ",
			expected: "(detached)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBranch(tt.branch)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestBuildWorktreeTable tests table building
func TestBuildWorktreeTable(t *testing.T) {
	tests := []struct {
		name      string
		worktrees []internal.Worktree
		verbose   bool
		contains  []string
	}{
		{
			name: "single worktree non-verbose",
			worktrees: []internal.Worktree{
				{
					Path:   "/home/user/repo",
					Branch: "main",
					SHA:    "abc123def",
					Clean:  true,
				},
			},
			verbose:  false,
			contains: []string{"Path", "Branch", "Status", "main", "clean"},
		},
		{
			name: "single worktree verbose",
			worktrees: []internal.Worktree{
				{
					Path:   "/home/user/repo",
					Branch: "main",
					SHA:    "abc123def",
					Clean:  true,
				},
			},
			verbose:  true,
			contains: []string{"Path", "Branch", "SHA", "Status", "main", "abc123d", "clean"},
		},
		{
			name: "multiple worktrees",
			worktrees: []internal.Worktree{
				{
					Path:   "/home/user/repo",
					Branch: "main",
					SHA:    "abc123",
					Clean:  true,
				},
				{
					Path:   "/home/user/repo-feature",
					Branch: "feature/auth",
					SHA:    "def456",
					Clean:  false,
				},
			},
			verbose:  false,
			contains: []string{"main", "feature/auth", "clean", "dirty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildWorktreeTable(tt.worktrees, tt.verbose)

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected table to contain '%s', but it didn't\nTable:\n%s", expected, result)
				}
			}
		})
	}
}

