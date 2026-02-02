//go:build L1 && ov
// +build L1,ov

package util

import (
	"os"
	"testing"
)

func TestIsDinD(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     bool
	}{
		{
			name:     "not_dind_when_env_unset",
			envValue: "",
			want:     false,
		},
		{
			name:     "is_dind_when_env_set",
			envValue: "/some/path",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original value
			originalValue := os.Getenv(EnvHostRepoRoot)
			defer os.Setenv(EnvHostRepoRoot, originalValue)

			if tt.envValue == "" {
				os.Unsetenv(EnvHostRepoRoot)
			} else {
				os.Setenv(EnvHostRepoRoot, tt.envValue)
			}

			got := IsDinD()
			if got != tt.want {
				t.Errorf("IsDinD() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslatePathForMount(t *testing.T) {
	tests := []struct {
		name           string
		localPath      string
		hostRepoRoot   string
		containerRoot  string
		expectedOutput string
	}{
		{
			name:           "direct_host_mode_returns_unchanged",
			localPath:      "/home/user/project/docs",
			hostRepoRoot:   "",
			containerRoot:  "",
			expectedOutput: "/home/user/project/docs",
		},
		{
			name:           "dind_translates_var_task_path",
			localPath:      "/var/task/docs",
			hostRepoRoot:   "C:\\projects\\eac",
			containerRoot:  "/var/task",
			expectedOutput: "C:\\projects\\eac\\docs",
		},
		{
			name:           "dind_translates_nested_path",
			localPath:      "/var/task/specs/mymodule/.design",
			hostRepoRoot:   "C:\\projects\\eac",
			containerRoot:  "/var/task",
			expectedOutput: "C:\\projects\\eac\\specs\\mymodule\\.design",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original values
			originalHostRoot := os.Getenv(EnvHostRepoRoot)
			originalContainerRoot := os.Getenv(EnvContainerRepoRoot)
			defer func() {
				os.Setenv(EnvHostRepoRoot, originalHostRoot)
				os.Setenv(EnvContainerRepoRoot, originalContainerRoot)
			}()

			if tt.hostRepoRoot == "" {
				os.Unsetenv(EnvHostRepoRoot)
			} else {
				os.Setenv(EnvHostRepoRoot, tt.hostRepoRoot)
			}

			if tt.containerRoot == "" {
				os.Unsetenv(EnvContainerRepoRoot)
			} else {
				os.Setenv(EnvContainerRepoRoot, tt.containerRoot)
			}

			got, err := TranslatePathForMount(tt.localPath)
			if err != nil {
				t.Fatalf("TranslatePathForMount() error = %v", err)
			}

			if got != tt.expectedOutput {
				t.Errorf("TranslatePathForMount() = %q, want %q", got, tt.expectedOutput)
			}
		})
	}
}

func TestFormatDockerVolume(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "windows_path_converted",
			input:    "C:\\projects\\eac\\docs",
			expected: "/c/projects/eac/docs",
		},
		{
			name:     "unix_path_unchanged",
			input:    "/home/user/project/docs",
			expected: "/home/user/project/docs",
		},
		{
			name:     "relative_path_unchanged",
			input:    "docs/content",
			expected: "docs/content",
		},
		{
			name:     "windows_path_lowercase_drive",
			input:    "D:\\data\\files",
			expected: "/d/data/files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDockerVolume(tt.input)
			if got != tt.expected {
				t.Errorf("FormatDockerVolume(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMakeRelative(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		base        string
		expected    string
		expectError bool
	}{
		{
			name:     "simple_relative",
			target:   "/var/task/docs",
			base:     "/var/task",
			expected: "docs",
		},
		{
			name:     "nested_relative",
			target:   "/var/task/specs/module/.design",
			base:     "/var/task",
			expected: "specs/module/.design",
		},
		{
			name:     "same_path",
			target:   "/var/task",
			base:     "/var/task",
			expected: "",
		},
		{
			name:        "not_under_base",
			target:      "/home/user/docs",
			base:        "/var/task",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeRelative(tt.target, tt.base)

			if tt.expectError {
				if err == nil {
					t.Errorf("makeRelative() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("makeRelative() unexpected error: %v", err)
			}

			if got != tt.expected {
				t.Errorf("makeRelative(%q, %q) = %q, want %q", tt.target, tt.base, got, tt.expected)
			}
		})
	}
}
