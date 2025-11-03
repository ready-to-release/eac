//go:build L2
// +build L2

package cmd

import (
	"bytes"
	"os"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/ready-to-release/eac/src/cli/internal/version"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupVersion   func()
		expectedOutput []string
	}{
		{
			name: "version with all fields",
			setupVersion: func() {
				// Reset to ensure clean state
				version.ResetToDefaults()
				version.SetVersion("v1.2.3", "2024-01-15T10:30:00Z", "abc123def", "2024-01-15T11:00:00Z", "false")
			},
			expectedOutput: []string{
				"Version:   v1.2.3",
				"Time:      2024-01-15T10:30:00Z",
				"BuildTime: 2024-01-15T11:00:00Z",
				"Revision:  abc123def",
			},
		},
		{
			name: "version with undefined version",
			setupVersion: func() {
				version.ResetToDefaults()
				version.SetVersion("", "2024-01-15T10:30:00Z", "abc123def", "2024-01-15T11:00:00Z", "false")
			},
			expectedOutput: []string{
				"Version:   undefined",
				"Time:      2024-01-15T10:30:00Z",
				"BuildTime: 2024-01-15T11:00:00Z",
				"Revision:  abc123def",
			},
		},
		{
			name: "version with modified flag",
			setupVersion: func() {
				version.ResetToDefaults()
				version.SetVersion("v1.0.0", "2024-01-15T10:30:00Z", "abc123def", "2024-01-15T11:00:00Z", "true")
			},
			expectedOutput: []string{
				"Version:   v1.0.0",
				"Time:      2024-01-15T10:30:00Z",
				"BuildTime: 2024-01-15T11:00:00Z",
				"Revision:  abc123def (modified)",
			},
		},
		{
			name: "version with empty fields",
			setupVersion: func() {
				version.ResetToDefaults()
				// Don't set anything - use defaults
			},
			expectedOutput: []string{
				"Version:   undefined",
				"BuildTime: no build time",
				"Revision: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original stdout
			originalStdout := os.Stdout
			defer func() { os.Stdout = originalStdout }()

			// Create pipe to capture stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Setup version state - need to set all fields to override defaults
			tt.setupVersion()

			// Execute command directly by calling the Run function
			versionCmd.Run(versionCmd, []string{})

			// Close writer and restore stdout
			w.Close()
			os.Stdout = originalStdout

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			// Check output
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestVersionCommandProperties(t *testing.T) {
	t.Run("command properties", func(t *testing.T) {
		if versionCmd.Use != "version" {
			t.Errorf("Expected Use to be 'version', got %q", versionCmd.Use)
		}

		if versionCmd.Short != "Print the version number of r2r CLI" {
			t.Errorf("Expected Short description, got %q", versionCmd.Short)
		}

		if !strings.Contains(versionCmd.Long, "r2r CLI's") {
			t.Errorf("Expected Long description to contain 'r2r CLI's', got %q", versionCmd.Long)
		}

		if versionCmd.Run == nil {
			t.Error("Expected Run function to be defined")
		}
	})
}

func TestGetSettingValue(t *testing.T) {
	tests := []struct {
		name     string
		info     *debug.BuildInfo
		key      string
		expected string
	}{
		{
			name: "key exists",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
					{Key: "vcs.time", Value: "2024-01-15T10:30:00Z"},
				},
			},
			key:      "vcs.revision",
			expected: "abc123",
		},
		{
			name: "key does not exist",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abc123"},
				},
			},
			key:      "vcs.modified",
			expected: "",
		},
		{
			name: "empty settings",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{},
			},
			key:      "vcs.revision",
			expected: "",
		},
		{
			name:     "nil info",
			info:     &debug.BuildInfo{},
			key:      "vcs.revision",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSettingValue(tt.info, tt.key)
			if result != tt.expected {
				t.Errorf("getSettingValue() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestVersionCommandInitialization(t *testing.T) {
	t.Run("command is registered with root", func(t *testing.T) {
		// Check if version command is in root's subcommands
		found := false
		for _, cmd := range RootCmd.Commands() {
			if cmd.Name() == "version" {
				found = true
				break
			}
		}
		if !found {
			t.Error("version command not registered with root command")
		}
	})
}

func TestVersionOutputFormat(t *testing.T) {
	t.Run("output format consistency", func(t *testing.T) {
		// Save original stdout
		originalStdout := os.Stdout
		defer func() { os.Stdout = originalStdout }()

		// Create pipe to capture stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Setup known version state
		version.ResetToDefaults()
		version.SetVersion("v1.0.0", "2024-01-15T10:30:00Z", "abc123def", "2024-01-15T11:00:00Z", "false")

		// Execute command directly by calling the Run function
		versionCmd.Run(versionCmd, []string{})

		// Close writer and restore stdout
		w.Close()
		os.Stdout = originalStdout

		// Read captured output
		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Verify we have exactly 4 lines
		if len(lines) != 4 {
			t.Errorf("Expected 4 output lines, got %d: %v", len(lines), lines)
		}

		// Verify line prefixes
		expectedPrefixes := []string{"Version:", "Time:", "BuildTime:", "Revision:"}
		for i, expectedPrefix := range expectedPrefixes {
			if i < len(lines) && !strings.HasPrefix(lines[i], expectedPrefix) {
				t.Errorf("Line %d should start with %q, got %q", i+1, expectedPrefix, lines[i])
			}
		}
	})
}
