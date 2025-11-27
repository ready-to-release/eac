//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"testing"
)

// TestParseRemoveConfig tests the configuration parsing
func TestParseRemoveConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *removeConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *removeConfig) {
				if config.keepBranch {
					t.Error("expected keepBranch=false, got true")
				}
				if config.deleteRemote {
					t.Error("expected deleteRemote=false, got true")
				}
				if config.force {
					t.Error("expected force=false, got true")
				}
			},
		},
		{
			name: "with --keep-branch flag",
			args: []string{"--keep-branch"},
			validate: func(t *testing.T, config *removeConfig) {
				if !config.keepBranch {
					t.Error("expected keepBranch=true, got false")
				}
			},
		},
		{
			name: "with --delete-remote flag",
			args: []string{"--delete-remote"},
			validate: func(t *testing.T, config *removeConfig) {
				if !config.deleteRemote {
					t.Error("expected deleteRemote=true, got false")
				}
			},
		},
		{
			name: "with --force flag",
			args: []string{"--force"},
			validate: func(t *testing.T, config *removeConfig) {
				if !config.force {
					t.Error("expected force=true, got false")
				}
			},
		},
		{
			name: "with -f shorthand",
			args: []string{"-f"},
			validate: func(t *testing.T, config *removeConfig) {
				if !config.force {
					t.Error("expected force=true, got false")
				}
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--keep-branch", "--delete-remote", "--force"},
			validate: func(t *testing.T, config *removeConfig) {
				if !config.keepBranch {
					t.Error("expected keepBranch=true, got false")
				}
				if !config.deleteRemote {
					t.Error("expected deleteRemote=true, got false")
				}
				if !config.force {
					t.Error("expected force=true, got false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args (simulate command: r2r work remove <args>)
			os.Args = append([]string{"r2r", "work", "remove"}, tt.args...)

			config, err := parseRemoveConfig()
			if err != nil {
				// Skip test if not in a git repository or other expected errors
				if containsStr(err.Error(), "repository root") ||
					containsStr(err.Error(), "workspace not found") {
					t.Skip("Not in a git repository or workspace")
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestRemoveConfigDefaults tests default configuration values
func TestRemoveConfigDefaults(t *testing.T) {
	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"r2r", "work", "remove"}

	config, err := parseRemoveConfig()
	if err != nil {
		// Skip test if not in a git repository
		if containsStr(err.Error(), "repository root") {
			t.Skip("Not in a git repository")
		}
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.keepBranch {
		t.Error("expected default keepBranch=false, got true")
	}

	if config.deleteRemote {
		t.Error("expected default deleteRemote=false, got true")
	}

	if config.force {
		t.Error("expected default force=false, got true")
	}

	if config.repoRoot == "" {
		t.Error("expected repoRoot to be set")
	}

	if config.branchName == "" {
		t.Error("expected branchName to be set")
	}
}

// TestValidateRemoveEnvironment tests validation logic
func TestValidateRemoveEnvironment(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		expectError   bool
		errorContains string
	}{
		{
			name:          "fail on main branch",
			branchName:    "main",
			expectError:   true,
			errorContains: "cannot remove main workspace",
		},
		{
			name:          "fail on master branch",
			branchName:    "master",
			expectError:   true,
			errorContains: "cannot remove main workspace",
		},
		{
			name:        "allow feature branch",
			branchName:  "feature/test",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &removeConfig{
				branchName: tt.branchName,
			}

			err := validateRemoveEnvironment(config)

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

// TestIsInWorkspace tests workspace detection
func TestIsInWorkspace(t *testing.T) {
	// Get current directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}

	tests := []struct {
		name         string
		worktreePath string
		repoRoot     string
		want         bool
	}{
		{
			name:         "same as cwd",
			worktreePath: cwd,
			repoRoot:     cwd,
			want:         true,
		},
		{
			name:         "different path",
			worktreePath: "/some/other/path",
			repoRoot:     cwd,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isInWorkspace(tt.worktreePath, tt.repoRoot)
			if got != tt.want {
				t.Errorf("isInWorkspace() = %v, want %v", got, tt.want)
			}
		})
	}
}
