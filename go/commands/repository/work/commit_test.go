//go:build L1 && ov
// +build L1,ov

package work

import (
	"os"
	"strings"
	"testing"

	eactesting "github.com/ready-to-release/eac/go/core/testing"
)

// TestParseCommitConfig tests the configuration parsing
func TestParseCommitConfig(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		validate func(*testing.T, *commitConfig)
	}{
		{
			name: "no flags",
			args: []string{},
			validate: func(t *testing.T, config *commitConfig) {
				if config.stageAll {
					t.Error("expected stageAll=false, got true")
				}
				if config.base.Debug {
					t.Error("expected debug=false, got true")
				}
				if config.customMessage != "" {
					t.Errorf("expected empty customMessage, got '%s'", config.customMessage)
				}
			},
		},
		{
			name: "with --all flag",
			args: []string{"--all"},
			validate: func(t *testing.T, config *commitConfig) {
				if !config.stageAll {
					t.Error("expected stageAll=true, got false")
				}
			},
		},
		{
			name: "with -a shorthand",
			args: []string{"-a"},
			validate: func(t *testing.T, config *commitConfig) {
				if !config.stageAll {
					t.Error("expected stageAll=true, got false")
				}
			},
		},
		{
			name: "with --debug flag",
			args: []string{"--debug"},
			validate: func(t *testing.T, config *commitConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with -d shorthand",
			args: []string{"-d"},
			validate: func(t *testing.T, config *commitConfig) {
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
			},
		},
		{
			name: "with --message flag",
			args: []string{"--message", "fix: test"},
			validate: func(t *testing.T, config *commitConfig) {
				if config.customMessage != "fix: test" {
					t.Errorf("expected customMessage='fix: test', got '%s'", config.customMessage)
				}
			},
		},
		{
			name: "with -m shorthand",
			args: []string{"-m", "feat: new feature"},
			validate: func(t *testing.T, config *commitConfig) {
				if config.customMessage != "feat: new feature" {
					t.Errorf("expected customMessage='feat: new feature', got '%s'", config.customMessage)
				}
			},
		},
		{
			name: "with multiple flags",
			args: []string{"--all", "--debug", "--message", "test commit"},
			validate: func(t *testing.T, config *commitConfig) {
				if !config.stageAll {
					t.Error("expected stageAll=true, got false")
				}
				if !config.base.Debug {
					t.Error("expected debug=true, got false")
				}
				if config.customMessage != "test commit" {
					t.Errorf("expected customMessage='test commit', got '%s'", config.customMessage)
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

			// Set test args (simulate command: eac work commit <args>)
			os.Args = append([]string{"eac", "work", "commit"}, tt.args...)

			config, err := parseCommitConfig()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

// TestParseCommitConfig_MissingMessageValue tests error handling
func TestParseCommitConfig_MissingMessageValue(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "--message without value",
			args:        []string{"--message"},
			expectError: true,
		},
		{
			name:        "-m without value",
			args:        []string{"-m"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original os.Args
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			// Set test args
			os.Args = append([]string{"clie", "work", "commit"}, tt.args...)

			_, err := parseCommitConfig()

			if tt.expectError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.expectError && err != nil {
				if !strings.Contains(err.Error(), "repository root") {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestCommitConfigDefaults tests default configuration values
func TestCommitConfigDefaults(t *testing.T) {
	eactesting.RequireGitRepository(t)

	// Save original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set test args
	os.Args = []string{"clie", "work", "commit"}

	config, err := parseCommitConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify defaults
	if config.stageAll {
		t.Error("expected default stageAll=false, got true")
	}

	if config.base.Debug {
		t.Error("expected default debug=false, got true")
	}

	if config.customMessage != "" {
		t.Errorf("expected default customMessage='', got '%s'", config.customMessage)
	}

	if config.base.RepoRoot == "" {
		t.Error("expected repoRoot to be set")
	}
}

